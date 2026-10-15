// Package gittransport implements authenticated Git Smart-HTTP transfer protocols
// (upload-pack, receive-pack) for exist.io repositories.
package gittransport

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"exist.io/internal/audit"
	"exist.io/internal/auth"
	"exist.io/internal/gitlfs"
	"exist.io/internal/gitstore"
	"exist.io/internal/metrics"
	"exist.io/internal/store"
)

// Server implements authenticated Git Smart HTTP transport for exist.io repositories.
// It maps /{owner}/{repo}.git and /{owner}/{repo} to bare git repositories in gitstore,
// authenticates callers against SQLite credentials (API keys and access tokens),
// verifies repository permissions (read vs write), and delegates packfile streaming
// to the official git-http-backend.
type Server struct {
	store       *store.DB
	git         *gitstore.Store
	lfs         *gitlfs.Server
	backendPath string
}

// NewServer returns an initialized Git Smart HTTP transport Server.
func NewServer(store *store.DB, git *gitstore.Store, lfs ...*gitlfs.Server) *Server {
	var lfsServer *gitlfs.Server
	if len(lfs) > 0 {
		lfsServer = lfs[0]
	}
	return &Server{
		store:       store,
		git:         git,
		lfs:         lfsServer,
		backendPath: findGitBackend(),
	}
}

// SetLFS sets the Git LFS server instance.
func (s *Server) SetLFS(lfs *gitlfs.Server) {
	s.lfs = lfs
}

func findGitBackend() string {
	out, err := exec.Command("git", "--exec-path").Output()
	if err == nil {
		p := filepath.Join(strings.TrimSpace(string(out)), "git-http-backend")
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	for _, candidate := range []string{
		"/usr/lib/git-core/git-http-backend",
		"/usr/libexec/git-core/git-http-backend",
		"/usr/local/libexec/git-core/git-http-backend",
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "git-http-backend"
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clean := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(clean, "/")
	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	ownerSlug := parts[0]
	repoPart := parts[1]
	repoName := strings.TrimSuffix(repoPart, ".git")

	subPath := ""
	if len(parts) > 2 {
		subPath = "/" + strings.Join(parts[2:], "/")
	}

	isUpload := strings.HasSuffix(r.URL.Path, "/git-upload-pack") || r.URL.Query().Get("service") == "git-upload-pack"
	isReceive := strings.HasSuffix(r.URL.Path, "/git-receive-pack") || r.URL.Query().Get("service") == "git-receive-pack"
	isLFS := strings.Contains(r.URL.Path, "/info/lfs")
	if !isUpload && !isReceive && !isLFS && !strings.HasSuffix(r.URL.Path, "/info/refs") && !strings.HasSuffix(r.URL.Path, "/HEAD") {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	repo, err := s.store.GetRepo(ctx, ownerSlug, repoName)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	userID, err := s.authenticate(ctx, r)
	if err != nil {
		w.Header().Set("WWW-Authenticate", `Basic realm="exist.io"`)
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	isLFSWrite := r.Method == http.MethodPut || r.Method == http.MethodDelete
	if isLFS && r.Method == http.MethodPost {
		if strings.HasSuffix(r.URL.Path, "/verify") || strings.Contains(r.URL.Path, "/locks") {
			isLFSWrite = true
		} else if strings.HasSuffix(r.URL.Path, "/objects/batch") {
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			var peek struct {
				Operation string `json:"operation"`
			}
			_ = json.Unmarshal(bodyBytes, &peek)
			if peek.Operation == "upload" {
				isLFSWrite = true
			}
		}
	}

	needsWrite := isReceive || isLFSWrite
	allowed, err := s.store.CheckRepoAccess(ctx, repo, userID, needsWrite)
	if err != nil || !allowed {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if isLFS {
		if s.lfs == nil {
			http.Error(w, "Git LFS is not enabled on this server", http.StatusNotImplemented)
			return
		}
		idx := strings.Index(r.URL.Path, "/info/lfs")
		lfsPath := strings.TrimPrefix(r.URL.Path[idx+len("/info/lfs"):], "/")
		s.lfs.HandleLFS(w, r, ownerSlug, repoName, lfsPath, userID)
		return
	}

	metrics.DefaultRegistry.IncGitTransfers()

	if isReceive {
		_ = gitstore.EnsureHooks(filepath.Join(s.git.Root(), ownerSlug, repoName+".git"))
		_ = audit.Log(audit.Event{
			Actor:     userID,
			ActorType: "user",
			Action:    "git.receive_pack",
			Resource:  ownerSlug + "/" + repoName,
			ClientIP:  r.RemoteAddr,
			Severity:  audit.SeverityInfo,
			Details:   "Incoming git push / receive-pack operation",
		})
	}

	s.execBackend(w, r, ownerSlug, repoName, subPath, userID)
}

func (s *Server) authenticate(ctx context.Context, r *http.Request) (string, error) {
	// 1. Basic Auth (standard Git credential helper)
	if user, pass, ok := r.BasicAuth(); ok {
		if id, err := s.checkCredential(ctx, pass); err == nil {
			return id, nil
		}
		if id, err := s.checkCredential(ctx, user); err == nil {
			return id, nil
		}
		u, err := s.store.GetUserByLogin(ctx, user)
		if err != nil {
			u, err = s.store.GetUserByEmail(ctx, user)
		}
		if err == nil && u.PasswordHash != "" && auth.CheckPassword(u.PasswordHash, pass) {
			return u.ID, nil
		}
	}

	// 2. Bearer Auth
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tok := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if id, err := s.checkCredential(ctx, tok); err == nil {
			return id, nil
		}
		if sess, _, err := s.store.GetSessionByHash(ctx, auth.HashCredential(tok)); err == nil {
			exp, err := time.Parse(time.RFC3339, sess.ExpiresAt)
			if err == nil && time.Now().Before(exp) {
				return sess.UserID, nil
			}
		}
	}

	// 3. Web session cookie
	if cookie, err := r.Cookie("exist_session"); err == nil && cookie.Value != "" {
		if sess, _, err := s.store.GetSessionByHash(ctx, auth.HashCredential(cookie.Value)); err == nil {
			exp, err := time.Parse(time.RFC3339, sess.ExpiresAt)
			if err == nil && time.Now().Before(exp) {
				return sess.UserID, nil
			}
		}
	}

	return "", errors.New("unauthenticated")
}

func (s *Server) checkCredential(ctx context.Context, cred string) (string, error) {
	if cred == "" {
		return "", errors.New("empty credential")
	}
	hash := auth.HashCredential(cred)

	// Check access tokens
	tok, err := s.store.GetAccessTokenByHash(ctx, hash)
	if err == nil {
		if tok.RevokedAt.Valid && tok.RevokedAt.String != "" {
			return "", errors.New("token revoked")
		}
		exp, err := time.Parse(time.RFC3339, tok.ExpiresAt)
		if err == nil && time.Now().After(exp) {
			return "", errors.New("token expired")
		}
		return tok.UserID, nil
	}

	// Check API keys
	key, err := s.store.GetAPIKeyByHash(ctx, hash)
	if err == nil {
		if key.RevokedAt.Valid && key.RevokedAt.String != "" {
			return "", errors.New("key revoked")
		}
		if key.ExpiresAt.Valid && key.ExpiresAt.String != "" {
			exp, err := time.Parse(time.RFC3339, key.ExpiresAt.String)
			if err == nil && time.Now().After(exp) {
				return "", errors.New("key expired")
			}
		}
		_ = s.store.TouchAPIKey(ctx, key.ID)
		return key.UserID, nil
	}

	return "", errors.New("invalid credential")
}

func (s *Server) execBackend(w http.ResponseWriter, r *http.Request, ownerSlug, repoName, subPath, userID string) {
	pathInfo := "/" + ownerSlug + "/" + repoName + ".git" + subPath

	cmd := exec.CommandContext(r.Context(), s.backendPath)
	cmd.Dir = s.git.Root()
	cmd.Env = append(os.Environ(),
		"GIT_PROJECT_ROOT="+s.git.Root(),
		"GIT_HTTP_EXPORT_ALL=1",
		"PATH_INFO="+pathInfo,
		"REMOTE_USER="+userID,
		"REMOTE_ADDR="+r.RemoteAddr,
		"CONTENT_TYPE="+r.Header.Get("Content-Type"),
		"QUERY_STRING="+r.URL.RawQuery,
		"REQUEST_METHOD="+r.Method,
	)
	cmd.Stdin = r.Body

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stdout.Close()

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reader := bufio.NewReader(stdout)
	statusCode := http.StatusOK

	// Parse and forward CGI headers
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		v := strings.TrimSpace(parts[1])

		if strings.EqualFold(k, "Status") {
			statusParts := strings.Fields(v)
			if len(statusParts) > 0 {
				if code, err := strconv.Atoi(statusParts[0]); err == nil {
					statusCode = code
				}
			}
		} else {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(statusCode)
	_, _ = io.Copy(w, reader)
	_ = cmd.Wait()
}
