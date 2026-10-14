package idgen

import (
	"strings"
	"sync"
	"testing"

	"github.com/oklog/ulid"
)

func checkID(t *testing.T, prefix, id string) {
	t.Helper()
	if !strings.HasPrefix(id, prefix+"_") {
		t.Fatalf("id %q missing %q prefix", id, prefix+"_")
	}
	suffix := strings.TrimPrefix(id, prefix+"_")
	if len(suffix) != 26 {
		t.Fatalf("%s: suffix = %d chars, want 26 (full ULID)", prefix, len(suffix))
	}
	if _, err := ulid.Parse(suffix); err != nil {
		t.Fatalf("%s: suffix %q is not a valid ULID: %v", prefix, suffix, err)
	}
}

func TestNew(t *testing.T) {
	checkID(t, "repo", New("repo"))
	checkID(t, "custom", New("custom"))
}

func TestTypedConstructors(t *testing.T) {
	cases := []struct {
		prefix string
		id     string
	}{
		{"app", NewApp()},
		{"inst", NewInstallation()},
		{"ns", NewNS()},
		{"repo", NewRepo()},
		{"pr", NewPR()},
		{"cr", NewCheckRun()},
		{"crg", NewCheckRunGroup()},
		{"cra", NewCheckRunAction()},
		{"nsca", NewNSCA()},
		{"nsip", NewNSIP()},
		{"whd", NewWebhookDelivery()},
		{"wha", NewWebhookAttempt()},
		{"evt", NewEvent()},
		{"grp", NewGroup()},
		{"user", NewUser()},
		{"stk", NewStack()},
		{"oit", NewInstallToken()},
	}
	for _, c := range cases {
		checkID(t, c.prefix, c.id)
	}
}

func TestNewUnique(t *testing.T) {
	const n = 10000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := New("repo")
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id %q at iteration %d", id, i)
		}
		seen[id] = struct{}{}
	}
}

func TestNewMonotonic(t *testing.T) {
	prev := parseSuffix(t, New("repo"))
	for i := 0; i < 1000; i++ {
		cur := parseSuffix(t, New("repo"))
		if cur.Compare(prev) <= 0 {
			t.Fatalf("id %d not strictly increasing: %s then %s", i, prev, cur)
		}
		prev = cur
	}
}

func TestNewConcurrent(t *testing.T) {
	const workers = 8
	const perWorker = 500
	ids := make(chan string, workers*perWorker)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perWorker; i++ {
				ids <- New("repo")
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := make(map[string]struct{}, workers*perWorker)
	for id := range ids {
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func parseSuffix(t *testing.T, id string) ulid.ULID {
	t.Helper()
	parsed, err := ulid.Parse(strings.TrimPrefix(id, "repo_"))
	if err != nil {
		t.Fatalf("parse %q: %v", id, err)
	}
	return parsed
}
