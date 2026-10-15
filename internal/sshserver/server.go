// Package sshserver implements an embedded pure-Go Git SSH server for exist.io.
// It authenticates Git clients against public keys enrolled in SQLite and trusted
// SSH Certificate Authorities, executing Git transport commands in bare repositories.
package sshserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"exist.io/internal/audit"
	"exist.io/internal/gitstore"
	"exist.io/internal/metrics"
	"exist.io/internal/store"
	"golang.org/x/crypto/ssh"
)

// Server is an embedded Git SSH server.
type Server struct {
	addr        string
	store       *store.DB
	git         *gitstore.Store
	hostKeyPath string
	listener    net.Listener
	sshConfig   *ssh.ServerConfig
	quit        chan struct{}
	wg          sync.WaitGroup
}

// New constructs an embedded SSH server.
func New(addr string, db *store.DB, git *gitstore.Store, hostKeyPath string) (*Server, error) {
	s := &Server{
		addr:        addr,
		store:       db,
		git:         git,
		hostKeyPath: hostKeyPath,
		quit:        make(chan struct{}),
	}

	signer, err := s.getOrCreateHostKey()
	if err != nil {
		return nil, fmt.Errorf("ssh: load host key: %w", err)
	}

	config := &ssh.ServerConfig{
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			fingerprint := ssh.FingerprintSHA256(key)
			ctx := context.Background()

			// 1. Look up user by public key fingerprint
			sshKey, user, err := s.store.GetSSHKeyByFingerprint(ctx, fingerprint)
			if err == nil {
				if user.Status == "suspended" {
					return nil, errors.New("ssh: user account suspended")
				}
				return &ssh.Permissions{
					Extensions: map[string]string{
						"user_id":     user.ID,
						"user_email":  user.Email,
						"key_id":      sshKey.ID,
						"fingerprint": fingerprint,
					},
				}, nil
			}

			// 1b. Look up deploy key by public key fingerprint
			deployKey, depRepo, err := s.store.GetDeployKeyByFingerprint(ctx, fingerprint)
			if err == nil {
				return &ssh.Permissions{
					Extensions: map[string]string{
						"deploy_key_id": deployKey.ID,
						"repo_id":       depRepo.ID,
						"read_only":     fmt.Sprintf("%t", deployKey.ReadOnly),
						"fingerprint":   fingerprint,
					},
				}, nil
			}

			// 2. Check trusted SSH Certificate Authorities
			// If key is an OpenSSH certificate, check against enrolled CAs
			cert, ok := key.(*ssh.Certificate)
			if ok && cert != nil {
				caFingerprint := ssh.FingerprintSHA256(cert.SignatureKey)
				caList, caErr := s.store.ListAllSSHCAs(ctx)
				if caErr == nil {
					for _, ca := range caList {
						if strings.TrimSpace(ca.Fingerprint) == caFingerprint {
							principal := conn.User()
							if len(cert.ValidPrincipals) > 0 {
								principal = cert.ValidPrincipals[0]
							}
							return &ssh.Permissions{
								Extensions: map[string]string{
									"user_id":     "ca_user_" + principal,
									"user_email":  principal,
									"fingerprint": caFingerprint,
								},
							}, nil
						}
					}
				}
			}

			return nil, fmt.Errorf("ssh: unauthorized key %s", fingerprint)
		},
	}
	config.AddHostKey(signer)
	s.sshConfig = config

	return s, nil
}

// Start begins accepting Git SSH connections in the background.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("ssh: listen %s: %w", s.addr, err)
	}
	s.listener = l

	s.wg.Add(1)
	go s.acceptLoop()
	return nil
}

// Addr returns the listener address.
func (s *Server) Addr() net.Addr {
	if s.listener != nil {
		return s.listener.Addr()
	}
	return nil
}

// Close stops the SSH listener and waits for active connections.
func (s *Server) Close() error {
	close(s.quit)
	if s.listener != nil {
		_ = s.listener.Close()
	}
	s.wg.Wait()
	return nil
}

func (s *Server) acceptLoop() {
	defer s.wg.Done()
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				continue
			}
		}

		s.wg.Add(1)
		go func(c net.Conn) {
			defer s.wg.Done()
			s.handleConn(c)
		}(conn)
	}
}

func (s *Server) handleConn(netConn net.Conn) {
	defer netConn.Close()

	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConfig)
	if err != nil {
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "only session channels supported")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			return
		}

		go s.handleSession(sshConn, channel, requests)
	}
}

func (s *Server) handleSession(conn *ssh.ServerConn, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		switch req.Type {
		case "exec":
			var payload struct {
				Command string
			}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				return
			}
			_ = req.Reply(true, nil)
			s.executeGitCommand(conn, channel, payload.Command)
			return

		default:
			_ = req.Reply(false, nil)
		}
	}
}

func (s *Server) executeGitCommand(conn *ssh.ServerConn, channel ssh.Channel, rawCmd string) {
	parts := strings.Fields(rawCmd)
	if len(parts) < 2 {
		fmt.Fprintf(channel.Stderr(), "exist.io: invalid git command %q\n", rawCmd)
		sendExitStatus(channel, 1)
		return
	}

	subCmd := parts[0]
	repoPath := strings.Trim(parts[1], "'\"")
	repoPath = strings.TrimPrefix(repoPath, "/")
	repoPath = strings.TrimSuffix(repoPath, ".git")

	pathParts := strings.Split(repoPath, "/")
	if len(pathParts) != 2 {
		fmt.Fprintf(channel.Stderr(), "exist.io: invalid repository path %q\n", repoPath)
		sendExitStatus(channel, 1)
		return
	}
	ownerSlug, repoName := pathParts[0], pathParts[1]

	isReceive := subCmd == "git-receive-pack"
	isUpload := subCmd == "git-upload-pack"
	if !isReceive && !isUpload {
		fmt.Fprintf(channel.Stderr(), "exist.io: unsupported command %q\n", subCmd)
		sendExitStatus(channel, 1)
		return
	}

	ctx := context.Background()
	repo, err := s.store.GetRepo(ctx, ownerSlug, repoName)
	if err != nil {
		fmt.Fprintf(channel.Stderr(), "exist.io: repository %s/%s not found\n", ownerSlug, repoName)
		sendExitStatus(channel, 1)
		return
	}

	userID := conn.Permissions.Extensions["user_id"]
	deployKeyID := conn.Permissions.Extensions["deploy_key_id"]
	if deployKeyID != "" {
		userID = "deploy_key:" + deployKeyID
		boundRepoID := conn.Permissions.Extensions["repo_id"]
		if boundRepoID != repo.ID {
			fmt.Fprintf(channel.Stderr(), "exist.io: deploy key is not authorized for repository %s/%s\n", ownerSlug, repoName)
			sendExitStatus(channel, 1)
			return
		}
		if isReceive && conn.Permissions.Extensions["read_only"] == "true" {
			fmt.Fprintf(channel.Stderr(), "exist.io: deploy key does not have write access to repository %s/%s\n", ownerSlug, repoName)
			sendExitStatus(channel, 1)
			return
		}
		_ = s.store.TouchDeployKey(ctx, deployKeyID)
	} else {
		allowed, err := s.store.CheckRepoAccess(ctx, repo, userID, isReceive)
		if err != nil || !allowed {
			fmt.Fprintf(channel.Stderr(), "exist.io: access denied to %s/%s\n", ownerSlug, repoName)
			sendExitStatus(channel, 1)
			return
		}
	}

	targetDir := filepath.Join(s.git.Root(), ownerSlug, repoName+".git")
	if isReceive {
		_ = gitstore.EnsureHooks(targetDir)
		actor := conn.Permissions.Extensions["user_email"]
		actorType := "user"
		if deployKeyID != "" {
			actor = "deploy_key:" + deployKeyID
			actorType = "deploy_key"
		}
		_ = audit.Log(audit.Event{
			Actor:     actor,
			ActorType: actorType,
			Action:    "git.ssh_push",
			Resource:  ownerSlug + "/" + repoName,
			ClientIP:  conn.RemoteAddr().String(),
			Severity:  audit.SeverityInfo,
			Details:   "Incoming SSH git push transfer",
		})
	}

	metrics.DefaultRegistry.IncGitTransfers()

	cmd := exec.Command(subCmd, targetDir)
	cmd.Dir = targetDir
	cmd.Env = append(os.Environ(),
		"GIT_PROJECT_ROOT="+s.git.Root(),
		"REMOTE_USER="+userID,
		"SSH_CONNECTION="+conn.RemoteAddr().String(),
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Fprintf(channel.Stderr(), "exist.io: stdin error: %v\n", err)
		sendExitStatus(channel, 1)
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Fprintf(channel.Stderr(), "exist.io: stdout error: %v\n", err)
		sendExitStatus(channel, 1)
		return
	}
	cmd.Stderr = channel.Stderr()

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(channel.Stderr(), "exist.io: failed to start %s: %v\n", subCmd, err)
		sendExitStatus(channel, 1)
		return
	}

	go func() {
		_, _ = io.Copy(stdin, channel)
		_ = stdin.Close()
	}()

	_, _ = io.Copy(channel, stdout)
	waitErr := cmd.Wait()

	exitCode := 0
	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}
	sendExitStatus(channel, exitCode)
}

func sendExitStatus(channel ssh.Channel, status int) {
	var payload struct {
		ExitStatus uint32
	}
	payload.ExitStatus = uint32(status)
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(payload))
}

func (s *Server) getOrCreateHostKey() (ssh.Signer, error) {
	if s.hostKeyPath != "" {
		if data, err := os.ReadFile(s.hostKeyPath); err == nil {
			return ssh.ParsePrivateKey(data)
		}
	}

	// Generate ephemeral/persisted Ed25519 host key
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	bytes, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		return nil, err
	}

	if s.hostKeyPath != "" {
		_ = os.MkdirAll(filepath.Dir(s.hostKeyPath), 0700)
		pemBlock := pem.EncodeToMemory(bytes)
		_ = os.WriteFile(s.hostKeyPath, pemBlock, 0600)
	}

	return ssh.NewSignerFromKey(priv)
}
