package app

import (
	"os"
	"path/filepath"
	"testing"

	"exist.io/internal/store"
)

func TestRecordActivationWritesOnce(t *testing.T) {
	dir := t.TempDir()
	pr := store.PullRequest{ID: "pr_1", HeadRef: "agent/fix", AuthorUserID: "usr_agent"}

	record, err := recordActivation(dir, pr, "usr_admin")
	if err != nil {
		t.Fatal(err)
	}
	if record == nil {
		t.Fatal("expected an activation record")
	}
	if record.PullRequestID != "pr_1" || record.ActorUserID != "usr_agent" {
		t.Fatalf("unexpected record: %+v", record)
	}

	info, err := os.Stat(filepath.Join(dir, "activation.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("activation.json mode = %o, want 600", info.Mode().Perm())
	}

	second, err := recordActivation(dir, pr, "usr_admin")
	if err != nil {
		t.Fatal(err)
	}
	if second != nil {
		t.Fatal("activation was recorded twice")
	}

	stored, ok := readActivation(dir)
	if !ok || stored.PullRequestID != "pr_1" {
		t.Fatalf("readActivation = %+v, %v", stored, ok)
	}
}

func TestRecordActivationIgnoresBootstrapAdmin(t *testing.T) {
	dir := t.TempDir()
	pr := store.PullRequest{ID: "pr_2", HeadRef: "agent/fix", AuthorUserID: "usr_admin"}

	record, err := recordActivation(dir, pr, "usr_admin")
	if err != nil {
		t.Fatal(err)
	}
	if record != nil {
		t.Fatal("bootstrap administrator must not activate the instance")
	}
}

func TestRecordActivationIgnoresNonAgentRef(t *testing.T) {
	dir := t.TempDir()
	pr := store.PullRequest{ID: "pr_3", HeadRef: "feature/regular", AuthorUserID: "usr_agent"}

	record, err := recordActivation(dir, pr, "usr_admin")
	if err != nil {
		t.Fatal(err)
	}
	if record != nil {
		t.Fatal("a non-agent ref must not activate the instance")
	}
}
