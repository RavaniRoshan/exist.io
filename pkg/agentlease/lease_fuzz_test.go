package agentlease

import (
	"testing"
	"time"
)

func FuzzTokenParser(f *testing.F) {
	secret := []byte("fuzz-test-secret-key-32-bytes!!")

	// Valid token seed
	validTok, _, err := Issue(secret, "fuzz-agent", "repo-123", []string{"refs/heads/agent/"}, nil, 1*time.Hour)
	if err == nil {
		f.Add(validTok)
	}

	// Corpus seeds
	f.Add("")
	f.Add("invalid")
	f.Add("a.b")
	f.Add("a.b.c")
	f.Add("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3j")
	f.Add("eyJpZCI6ImxlYXNlXzEifQ.invalidsig")
	f.Add(string(make([]byte, 1000)))

	f.Fuzz(func(t *testing.T, token string) {
		// Verify must never panic on arbitrary input
		_, _ = Verify(secret, token)
	})
}
