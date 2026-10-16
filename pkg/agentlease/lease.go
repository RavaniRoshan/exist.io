package agentlease

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"exist.io/pkg/idgen"
)

var (
	ErrExpiredLease     = errors.New("agent capability lease has expired")
	ErrRevokedLease     = errors.New("agent capability lease has been revoked")
	ErrInvalidSignature = errors.New("invalid lease signature")
	ErrRefNotAllowed    = errors.New("agent is not permitted to mutate this ref under active lease")
	ErrMalformedToken   = errors.New("malformed lease token")
)

// Lease represents a short-lived, cryptographically scoped capability granted to an AI coding agent.
type Lease struct {
	ID                    string    `json:"id"`
	AgentID               string    `json:"agent_id"`
	RepoID                string    `json:"repo_id"`
	AllowedRefPrefixes    []string  `json:"allowed_ref_prefixes"`
	DisallowedRefPrefixes []string  `json:"disallowed_ref_prefixes"`
	IssuedAt              time.Time `json:"issued_at"`
	ExpiresAt             time.Time `json:"expires_at"`
}

// Issue creates and cryptographically signs an ephemeral capability lease.
func Issue(secret []byte, agentID, repoID string, allowedRefs, disallowedRefs []string, ttl time.Duration) (string, *Lease, error) {
	if len(secret) == 0 {
		return "", nil, errors.New("lease signing secret must not be empty")
	}
	if agentID == "" || repoID == "" {
		return "", nil, errors.New("agentID and repoID are required")
	}
	if ttl == 0 {
		ttl = 15 * time.Minute
	}

	now := time.Now().UTC()
	lease := &Lease{
		ID:                    idgen.New("lease"),
		AgentID:               agentID,
		RepoID:                repoID,
		AllowedRefPrefixes:    allowedRefs,
		DisallowedRefPrefixes: disallowedRefs,
		IssuedAt:              now,
		ExpiresAt:             now.Add(ttl),
	}

	payloadBytes, err := json.Marshal(lease)
	if err != nil {
		return "", nil, fmt.Errorf("marshal lease: %w", err)
	}

	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadBytes)
	sig := sign(secret, payloadB64)
	token := fmt.Sprintf("%s.%s", payloadB64, sig)

	return token, lease, nil
}

// Verify decodes and validates the cryptographic integrity and expiration of a lease token.
func Verify(secret []byte, token string) (*Lease, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrMalformedToken
	}

	payloadB64, sig := parts[0], parts[1]
	expectedSig := sign(secret, payloadB64)
	if !hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return nil, ErrInvalidSignature
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return nil, ErrMalformedToken
	}

	var lease Lease
	if err := json.Unmarshal(payloadBytes, &lease); err != nil {
		return nil, ErrMalformedToken
	}

	if time.Now().UTC().After(lease.ExpiresAt) {
		return &lease, ErrExpiredLease
	}

	return &lease, nil
}

// CanPushRef checks whether the lease grants permission to modify the target ref.
func (l *Lease) CanPushRef(refName string) bool {
	// 1. Explicitly disallowed ref prefixes always take precedence (e.g. main, master, production)
	for _, dis := range l.DisallowedRefPrefixes {
		if dis != "" && strings.HasPrefix(refName, dis) {
			return false
		}
	}

	// 2. If allowed prefixes are specified, ref must match at least one
	if len(l.AllowedRefPrefixes) > 0 {
		for _, allow := range l.AllowedRefPrefixes {
			if allow != "" && strings.HasPrefix(refName, allow) {
				return true
			}
		}
		return false
	}

	return true
}

func sign(secret []byte, data string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
