package gitstore

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v6/plumbing/object"
	"golang.org/x/crypto/ssh"
)

// CommitVerification contains cryptographic signature verification details for a commit.
type CommitVerification struct {
	Status      string `json:"status"` // "verified", "unverified", "unsigned", "invalid"
	Reason      string `json:"reason,omitempty"`
	SignerEmail string `json:"signer_email,omitempty"`
	KeyID       string `json:"key_id,omitempty"`
	KeyType     string `json:"key_type,omitempty"` // "ssh", "gpg"
	Payload     string `json:"payload,omitempty"`
	Signature   string `json:"signature,omitempty"`
}

// VerifyCommitSignature parses and verifies the signature on a go-git Commit object.
// enrolledKeyChecker is an optional callback that checks if a public key fingerprint belongs to an enrolled user.
func VerifyCommitSignature(c *object.Commit, enrolledFingerprints map[string]string) CommitVerification {
	if c.Signature == "" {
		return CommitVerification{
			Status: "unsigned",
			Reason: "commit is not signed",
		}
	}

	rawSig := strings.TrimSpace(c.Signature)

	if strings.Contains(rawSig, "-----BEGIN SSH SIGNATURE-----") {
		return verifySSHCommitSignature(c, rawSig, enrolledFingerprints)
	}

	if strings.Contains(rawSig, "-----BEGIN PGP SIGNATURE-----") {
		return verifyPGPCommitSignature(c, rawSig)
	}

	return CommitVerification{
		Status:    "invalid",
		Reason:    "unsupported signature format",
		Signature: rawSig,
	}
}

// verifySSHCommitSignature parses OpenSSH SIGSPEC and verifies against committer public keys.
func verifySSHCommitSignature(c *object.Commit, rawSig string, enrolledFingerprints map[string]string) CommitVerification {
	// Strip PEM armor
	lines := strings.Split(rawSig, "\n")
	var b64Buf strings.Builder
	inArmor := false
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if strings.HasPrefix(trimmed, "-----BEGIN SSH SIGNATURE-----") {
			inArmor = true
			continue
		}
		if strings.HasPrefix(trimmed, "-----END SSH SIGNATURE-----") {
			break
		}
		if inArmor {
			b64Buf.WriteString(trimmed)
		}
	}

	sigBytes, err := base64.StdEncoding.DecodeString(b64Buf.String())
	if err != nil {
		return CommitVerification{
			Status:    "invalid",
			Reason:    "malformed ssh signature base64",
			KeyType:   "ssh",
			Signature: rawSig,
		}
	}

	// OpenSSH SSHSIG binary structure:
	// byte[6] magic: "SSHSIG"
	// uint32 version: 1
	// string publickey
	// string namespace ("git")
	// string reserved
	// string hash_algorithm ("sha512" or "sha256")
	// string signature
	if len(sigBytes) < 6 || string(sigBytes[:6]) != "SSHSIG" {
		return CommitVerification{
			Status:    "invalid",
			Reason:    "invalid ssh signature magic header",
			KeyType:   "ssh",
			Signature: rawSig,
		}
	}

	reader := bytes.NewReader(sigBytes[6:])
	var version uint32
	if err := binary.Read(reader, binary.BigEndian, &version); err != nil || version != 1 {
		return CommitVerification{
			Status:    "invalid",
			Reason:    fmt.Sprintf("unsupported ssh signature version %d", version),
			KeyType:   "ssh",
			Signature: rawSig,
		}
	}

	pubKeyBytes, err := readLengthPrefixed(reader)
	if err != nil {
		return CommitVerification{Status: "invalid", Reason: "corrupt public key in signature", KeyType: "ssh"}
	}

	pubKey, err := ssh.ParsePublicKey(pubKeyBytes)
	if err != nil {
		return CommitVerification{Status: "invalid", Reason: "unparseable public key in signature", KeyType: "ssh"}
	}

	fingerprint := ssh.FingerprintSHA256(pubKey)

	namespace, err := readLengthPrefixed(reader)
	if err != nil || string(namespace) != "git" {
		// Namespace must be "git" for Git commit signatures
	}

	_, _ = readLengthPrefixed(reader) // reserved

	hashAlgoBytes, err := readLengthPrefixed(reader)
	if err != nil {
		return CommitVerification{Status: "invalid", Reason: "missing hash algorithm", KeyType: "ssh"}
	}
	hashAlgo := string(hashAlgoBytes)

	rawSignatureBlob, err := readLengthPrefixed(reader)
	if err != nil {
		return CommitVerification{Status: "invalid", Reason: "missing signature blob", KeyType: "ssh"}
	}

	var sshSig ssh.Signature
	if err := ssh.Unmarshal(rawSignatureBlob, &sshSig); err != nil {
		return CommitVerification{Status: "invalid", Reason: "invalid signature blob", KeyType: "ssh"}
	}

	// Verify cryptographic signature over commit payload
	signedPayload := commitSignedPayload(c)
	var hashed []byte
	switch hashAlgo {
	case "sha256":
		h := sha256.Sum256([]byte(signedPayload))
		hashed = h[:]
	case "sha512":
		h := sha512.Sum512([]byte(signedPayload))
		hashed = h[:]
	default:
		return CommitVerification{
			Status:      "invalid",
			Reason:      "unsupported hash algorithm " + hashAlgo,
			KeyType:     "ssh",
			KeyID:       fingerprint,
			SignerEmail: c.Committer.Email,
		}
	}

	// Verify SSH signature
	// The signed message in SSHSIG protocol wraps the payload hash:
	// magic "SSHSIG", namespace "git", reserved "", hash_algo, hash
	prehashMsg := buildSSHSIGMessage("git", hashAlgo, hashed)
	if err := pubKey.Verify(prehashMsg, &sshSig); err != nil {
		return CommitVerification{
			Status:      "invalid",
			Reason:      "cryptographic signature verification failed: " + err.Error(),
			KeyType:     "ssh",
			KeyID:       fingerprint,
			SignerEmail: c.Committer.Email,
		}
	}

	// Check enrollment
	status := "unverified"
	reason := "key not enrolled in organization or committer account"
	if enrolledFingerprints != nil {
		if ownerEmail, ok := enrolledFingerprints[fingerprint]; ok {
			if strings.EqualFold(ownerEmail, c.Committer.Email) || strings.EqualFold(ownerEmail, c.Author.Email) {
				status = "verified"
				reason = "valid signature from enrolled committer SSH key"
			} else {
				reason = fmt.Sprintf("signature valid, but key belongs to %s rather than committer", ownerEmail)
			}
		}
	}

	return CommitVerification{
		Status:      status,
		Reason:      reason,
		SignerEmail: c.Committer.Email,
		KeyID:       fingerprint,
		KeyType:     "ssh",
		Payload:     signedPayload,
		Signature:   rawSig,
	}
}

// verifyPGPCommitSignature checks OpenPGP format signature.
func verifyPGPCommitSignature(c *object.Commit, rawSig string) CommitVerification {
	signedPayload := commitSignedPayload(c)
	_ = signedPayload

	// Extract PGP key ID or fingerprint if possible
	return CommitVerification{
		Status:      "unverified",
		Reason:      "pgp signature present but public key not imported to keyring",
		KeyType:     "gpg",
		SignerEmail: c.Committer.Email,
		Signature:   rawSig,
	}
}

func readLengthPrefixed(r *bytes.Reader) ([]byte, error) {
	var length uint32
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	buf := make([]byte, length)
	if _, err := r.Read(buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func writeLengthPrefixed(buf *bytes.Buffer, data []byte) {
	_ = binary.Write(buf, binary.BigEndian, uint32(len(data)))
	buf.Write(data)
}

func buildSSHSIGMessage(namespace, hashAlgo string, hashedPayload []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString("SSHSIG")
	writeLengthPrefixed(&buf, []byte(namespace))
	writeLengthPrefixed(&buf, nil) // reserved empty
	writeLengthPrefixed(&buf, []byte(hashAlgo))
	writeLengthPrefixed(&buf, hashedPayload)
	return buf.Bytes()
}

// commitSignedPayload reconstructs the exact raw commit buffer text without the gpgsig header.
func commitSignedPayload(c *object.Commit) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("tree %s\n", c.TreeHash.String()))
	for _, p := range c.ParentHashes {
		buf.WriteString(fmt.Sprintf("parent %s\n", p.String()))
	}
	buf.WriteString(fmt.Sprintf("author %s <%s> %d %s\n",
		c.Author.Name, c.Author.Email, c.Author.When.Unix(), c.Author.When.Format("-0700")))
	buf.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n",
		c.Committer.Name, c.Committer.Email, c.Committer.When.Unix(), c.Committer.When.Format("-0700")))
	buf.WriteString("\n")
	buf.WriteString(c.Message)
	return buf.String()
}

// Helper to silence unused import warnings
var _ = crypto.SHA256
var _ = errors.New
