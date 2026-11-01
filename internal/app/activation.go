package app

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"exist.io/internal/store"
)

// ActivationRecord is the once-only record of the first qualifying merge.
type ActivationRecord struct {
	ActivatedAt   string `json:"activated_at"`
	PullRequestID string `json:"pull_request_id"`
	ActorUserID   string `json:"actor_user_id"`
}

// recordActivation writes the activation record when the merge qualifies:
// an agent/* ref merged by someone other than the bootstrap administrator.
// It returns nil when the merge does not qualify or the record already exists.
func recordActivation(dataDir string, pr store.PullRequest, bootstrapAdminID string) (*ActivationRecord, error) {
	if dataDir == "" || dataDir == ":memory:" {
		return nil, nil
	}
	if bootstrapAdminID != "" && pr.AuthorUserID == bootstrapAdminID {
		return nil, nil
	}
	if !strings.HasPrefix(pr.HeadRef, "agent/") && !strings.HasPrefix(pr.HeadRef, "refs/heads/agent/") {
		return nil, nil
	}

	record := ActivationRecord{
		ActivatedAt:   time.Now().UTC().Format(time.RFC3339),
		PullRequestID: pr.ID,
		ActorUserID:   pr.AuthorUserID,
	}
	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(dataDir, "activation.json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600) // #nosec G304 -- path is derived from the operator-configured data directory
	if errors.Is(err, os.ErrExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	if _, err := file.Write(body); err != nil {
		return nil, err
	}
	return &record, nil
}

// readActivation returns the persisted activation record, if any.
func readActivation(dataDir string) (ActivationRecord, bool) {
	if dataDir == "" || dataDir == ":memory:" {
		return ActivationRecord{}, false
	}
	body, err := os.ReadFile(filepath.Join(dataDir, "activation.json")) // #nosec G304 -- path is derived from the operator-configured data directory
	if err != nil {
		return ActivationRecord{}, false
	}
	var record ActivationRecord
	if err := json.Unmarshal(body, &record); err != nil {
		return ActivationRecord{}, false
	}
	return record, true
}
