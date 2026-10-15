package sshserver

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"path/filepath"
	"testing"
	"time"

	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"golang.org/x/crypto/ssh"
)

func TestSSHServerAuthentication(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	git, err := gitstore.New(filepath.Join(dataDir, "git"))
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	// 1. Generate client keypair
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("NewPublicKey: %v", err)
	}
	fingerprint := ssh.FingerprintSHA256(sshPub)

	// 2. Enroll user and SSH key in database
	ctx := context.Background()
	user := store.User{
		ID:        "user_ssh_tester",
		Email:     "ssh@exist.io",
		Login:     "sshtester",
		Role:      "member",
		Status:    "active",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.CreateUserWithPassword(ctx, user); err != nil {
		t.Fatalf("CreateUserWithPassword: %v", err)
	}

	keyRecord := store.SSHKey{
		ID:          "sshkey_test",
		UserID:      user.ID,
		Name:        "Test Key",
		Fingerprint: fingerprint,
		PublicKey:   string(ssh.MarshalAuthorizedKey(sshPub)),
		KeyType:     sshPub.Type(),
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	if err := db.AddUserSSHKey(ctx, keyRecord); err != nil {
		t.Fatalf("AddUserSSHKey: %v", err)
	}

	// 3. Start SSH Server on dynamic port
	server, err := New("127.0.0.1:0", db, git, filepath.Join(dataDir, "host_key"))
	if err != nil {
		t.Fatalf("New SSH Server: %v", err)
	}
	if err := server.Start(); err != nil {
		t.Fatalf("server.Start: %v", err)
	}
	defer server.Close()

	// 4. Test client connection with enrolled key
	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("NewSignerFromKey: %v", err)
	}
	clientConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         3 * time.Second,
	}

	client, err := ssh.Dial("tcp", server.Addr().String(), clientConfig)
	if err != nil {
		t.Fatalf("ssh.Dial with enrolled key failed: %v", err)
	}
	defer client.Close()

	// 5. Test connection with an unenrolled key (should fail)
	_, badPriv, _ := ed25519.GenerateKey(rand.Reader)
	badSigner, _ := ssh.NewSignerFromKey(badPriv)
	badConfig := &ssh.ClientConfig{
		User:            "git",
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(badSigner)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         1 * time.Second,
	}

	_, err = ssh.Dial("tcp", server.Addr().String(), badConfig)
	if err == nil {
		t.Fatalf("expected dial with unenrolled key to fail, but succeeded")
	}
}
