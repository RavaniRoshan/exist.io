package gitstore

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"golang.org/x/crypto/ssh"
)

func TestVerifyCommitSignature_Unsigned(t *testing.T) {
	c := &object.Commit{
		TreeHash: plumbing.NewHash("4b825dc642cb6eb9a060e54bf8d69288fbee4904"),
		Author: object.Signature{
			Name:  "Alice",
			Email: "alice@exist.io",
			When:  time.Now(),
		},
		Committer: object.Signature{
			Name:  "Alice",
			Email: "alice@exist.io",
			When:  time.Now(),
		},
		Message: "initial commit",
	}

	res := VerifyCommitSignature(c, nil)
	if res.Status != "unsigned" {
		t.Fatalf("expected status unsigned, got %s", res.Status)
	}
}

func TestVerifyCommitSignature_SSH(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey: %v", err)
	}
	sshPub, err := ssh.NewPublicKey(pub)
	if err != nil {
		t.Fatalf("NewPublicKey: %v", err)
	}
	fingerprint := ssh.FingerprintSHA256(sshPub)

	c := &object.Commit{
		TreeHash: plumbing.NewHash("4b825dc642cb6eb9a060e54bf8d69288fbee4904"),
		Author: object.Signature{
			Name:  "Bob",
			Email: "bob@exist.io",
			When:  time.Unix(1700000000, 0).UTC(),
		},
		Committer: object.Signature{
			Name:  "Bob",
			Email: "bob@exist.io",
			When:  time.Unix(1700000000, 0).UTC(),
		},
		Message: "feat: enterprise ssh signatures\n",
	}

	payload := commitSignedPayload(c)
	h := sha512.Sum512([]byte(payload))
	msg := buildSSHSIGMessage("git", "sha512", h[:])

	signer, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatalf("NewSignerFromKey: %v", err)
	}
	sig, err := signer.Sign(rand.Reader, msg)
	if err != nil {
		t.Fatalf("signer.Sign: %v", err)
	}

	// Pack into SSHSIG binary format
	var packet bytes.Buffer
	packet.WriteString("SSHSIG")
	packet.Write([]byte{0, 0, 0, 1})
	writeLengthPrefixed(&packet, sshPub.Marshal())
	writeLengthPrefixed(&packet, []byte("git"))
	writeLengthPrefixed(&packet, nil)
	writeLengthPrefixed(&packet, []byte("sha512"))
	writeLengthPrefixed(&packet, ssh.Marshal(sig))

	encodedSig := "-----BEGIN SSH SIGNATURE-----\n" +
		base64.StdEncoding.EncodeToString(packet.Bytes()) +
		"\n-----END SSH SIGNATURE-----"

	c.Signature = encodedSig

	// 1. Without enrollment: status should be "unverified"
	res := VerifyCommitSignature(c, nil)
	if res.Status != "unverified" {
		t.Fatalf("expected unverified without enrollment, got %s: %s", res.Status, res.Reason)
	}
	if res.KeyID != fingerprint {
		t.Fatalf("expected key id %s, got %s", fingerprint, res.KeyID)
	}

	// 2. With enrollment: status should be "verified"
	enrolled := map[string]string{
		fingerprint: "bob@exist.io",
	}
	resVerified := VerifyCommitSignature(c, enrolled)
	if resVerified.Status != "verified" {
		t.Fatalf("expected verified with enrollment, got %s: %s", resVerified.Status, resVerified.Reason)
	}

	// 3. Corrupt signature
	c.Signature = strings.Replace(encodedSig, "A", "B", 1)
	resCorrupt := VerifyCommitSignature(c, enrolled)
	if resCorrupt.Status != "invalid" {
		t.Fatalf("expected invalid for corrupt signature, got %s", resCorrupt.Status)
	}
}
