package gittransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"exist.io/internal/auth"
	"exist.io/internal/gitstore"
	"exist.io/internal/store"
	"exist.io/pkg/agentlease"
	"exist.io/pkg/idgen"

	"github.com/go-git/go-git/v6/plumbing/object"
)

func TestGitSmartHTTPProtocolAndClonePush(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	gitDir := filepath.Join(t.TempDir(), "git")
	git, err := gitstore.New(gitDir)
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	// 1. Seed user & namespace
	userID := idgen.NewUser()
	if _, err := db.UpsertUserByEmail(ctx, "dev@exist.io", func() string { return userID }); err != nil {
		t.Fatal(err)
	}
	nsID := idgen.NewNS()
	if err := db.CreateNamespace(ctx, store.Namespace{
		ID:          nsID,
		Slug:        "myteam",
		Type:        "user",
		Name:        "My Team",
		OwnerUserID: userID,
	}); err != nil {
		t.Fatal(err)
	}

	// 2. Seed repo in store and initialize in gitstore
	repoID := idgen.NewRepo()
	if err := db.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "demo-repo",
		FullName:      "myteam/demo-repo",
		DefaultBranch: "main",
		Visibility:    "private",
	}); err != nil {
		t.Fatal(err)
	}

	_, err = git.Init("myteam", "demo-repo", "main")
	if err != nil {
		t.Fatalf("git.Init: %v", err)
	}
	// Create an initial commit in the bare repo so HEAD is valid
	sig := object.Signature{Name: "Dev", Email: "dev@exist.io", When: time.Now()}
	_, err = git.WriteFiles("myteam", "demo-repo", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Demo Repo\n"},
	}, sig, "Initial commit")
	if err != nil {
		t.Fatalf("WriteFiles: %v", err)
	}

	// 3. Create API key for user
	rawKey, prefix, hash, err := auth.GenerateAPIKey()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.CreateAPIKey(ctx, store.APIKey{
		ID:        idgen.New("key"),
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Name:      "dev-key",
		Scopes:    "*",
	}); err != nil {
		t.Fatal(err)
	}

	// 4. Start HTTP test server
	gitHandler := NewServer(db, git)
	srv := httptest.NewServer(gitHandler)
	defer srv.Close()

	// 5. Test unauthenticated request -> 401
	unauthReq, _ := http.NewRequest(http.MethodGet, srv.URL+"/myteam/demo-repo.git/info/refs?service=git-upload-pack", nil)
	unauthResp, err := http.DefaultClient.Do(unauthReq)
	if err != nil {
		t.Fatal(err)
	}
	defer unauthResp.Body.Close()
	if unauthResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthenticated status = %d, want 401", unauthResp.StatusCode)
	}
	if !strings.Contains(unauthResp.Header.Get("WWW-Authenticate"), "Basic") {
		t.Fatalf("missing Basic auth header: %s", unauthResp.Header.Get("WWW-Authenticate"))
	}

	// 6. Test nonexistent repo -> 404
	missingReq, _ := http.NewRequest(http.MethodGet, srv.URL+"/myteam/no-such-repo.git/info/refs?service=git-upload-pack", nil)
	missingReq.SetBasicAuth("dev", rawKey)
	missingResp, err := http.DefaultClient.Do(missingReq)
	if err != nil {
		t.Fatal(err)
	}
	defer missingResp.Body.Close()
	if missingResp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing repo status = %d, want 404", missingResp.StatusCode)
	}

	// 7. Test authenticated info/refs protocol response
	infoReq, _ := http.NewRequest(http.MethodGet, srv.URL+"/myteam/demo-repo.git/info/refs?service=git-upload-pack", nil)
	infoReq.SetBasicAuth("dev", rawKey)
	infoResp, err := http.DefaultClient.Do(infoReq)
	if err != nil {
		t.Fatal(err)
	}
	defer infoResp.Body.Close()
	if infoResp.StatusCode != http.StatusOK {
		t.Fatalf("authenticated info/refs status = %d, want 200", infoResp.StatusCode)
	}
	contentType := infoResp.Header.Get("Content-Type")
	if contentType != "application/x-git-upload-pack-advertisement" {
		t.Fatalf("unexpected content-type: %s", contentType)
	}

	// 8. Test real `git clone` using system git binary
	cloneDir := filepath.Join(t.TempDir(), "cloned")
	authCloneURL := strings.Replace(srv.URL, "http://", "http://dev:"+rawKey+"@", 1) + "/myteam/demo-repo.git"

	cmd := exec.Command("git", "clone", authCloneURL, cloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone failed: %v\nOutput: %s", err, string(out))
	}

	// Verify cloned file exists
	readmeContent, err := os.ReadFile(filepath.Join(cloneDir, "README.md"))
	if err != nil {
		t.Fatalf("read cloned README.md: %v", err)
	}
	if string(readmeContent) != "# Demo Repo\n" {
		t.Fatalf("unexpected README content: %s", string(readmeContent))
	}

	// 9. Test real `git commit` and `git push`
	newFilePath := filepath.Join(cloneDir, "feature.txt")
	if err := os.WriteFile(newFilePath, []byte("built with exist.io\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	gitExec := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = cloneDir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Dev",
			"GIT_AUTHOR_EMAIL=dev@exist.io",
			"GIT_COMMITTER_NAME=Dev",
			"GIT_COMMITTER_EMAIL=dev@exist.io",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %s failed: %v\nOutput: %s", args[0], err, string(out))
		}
	}

	gitExec("checkout", "-B", "main")
	gitExec("add", "feature.txt")
	gitExec("commit", "-m", "add feature.txt")
	gitExec("push", "origin", "main")

	// 10. Verify commit was written into exist.io bare repo on disk
	fileBytes, _, _, _, err := git.ReadFileAtPath("myteam", "demo-repo", "main", "feature.txt")
	if err != nil {
		t.Fatalf("ReadFileAtPath in bare repo: %v", err)
	}
	if string(fileBytes) != "built with exist.io\n" {
		t.Fatalf("unexpected file in bare repo: %s", string(fileBytes))
	}

	// 11. Test unauthorized write denial (user with only read access)
	readerUserID := idgen.NewUser()
	_ = db.CreateNamespace(ctx, store.Namespace{
		ID:        idgen.NewNS(),
		Slug:      "reader-ns",
		Type:      "user",
		Name:      "Reader",
		OwnerUserID: readerUserID,
	})
	readerRawKey, readerPrefix, readerHash, _ := auth.GenerateAPIKey()
	_ = db.CreateAPIKey(ctx, store.APIKey{
		ID:        idgen.New("key"),
		UserID:    readerUserID,
		KeyHash:   readerHash,
		KeyPrefix: readerPrefix,
		Name:      "reader-key",
		Scopes:    "*",
	})
	// Grant read permission only on repo
	_ = db.CreateAccessToken(ctx, store.AccessToken{
		ID:        idgen.New("tok"),
		TokenHash: "dummy",
		UserID:    readerUserID,
		TokenType: "user",
		Scopes:    "*",
		ExpiresAt: time.Now().Add(time.Hour).Format(time.RFC3339),
	})
	_, _ = db.UpsertGrant(ctx, store.Grant{
		ID:            idgen.New("grant"),
		Scope:         "repository",
		NamespaceID:   nsID,
		RepoID:        repoID,
		PrincipalKind: "user",
		UserID:        readerUserID,
		Permission:    "read",
	})

	// Try push with reader credentials
	unauthPushURL := strings.Replace(srv.URL, "http://", "http://reader:"+readerRawKey+"@", 1) + "/myteam/demo-repo.git"
	denyCmd := exec.Command("git", "push", unauthPushURL, "main")
	denyCmd.Dir = cloneDir
	denyOut, err := denyCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected push to be denied for read-only user, but succeeded: %s", string(denyOut))
	}
}

func TestGitSmartHTTP_AgentLeaseEnforcement(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	gitDir := filepath.Join(t.TempDir(), "git")
	git, err := gitstore.New(gitDir)
	if err != nil {
		t.Fatalf("gitstore.New: %v", err)
	}

	// Build existd for pre-receive hook execution
	existdBin := filepath.Join(t.TempDir(), "existd")
	buildCmd := exec.Command("go", "build", "-o", existdBin, "exist.io/cmd/existd")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("build existd: %v\nOutput: %s", err, string(out))
	}
	t.Setenv("EXISTD_PATH", existdBin)

	// 1. Seed user & namespace
	userID := idgen.NewUser()
	if _, err := db.UpsertUserByEmail(ctx, "admin@exist.io", func() string { return userID }); err != nil {
		t.Fatal(err)
	}
	nsID := idgen.NewNS()
	if err := db.CreateNamespace(ctx, store.Namespace{
		ID:          nsID,
		Slug:        "agentteam",
		Type:        "user",
		Name:        "Agent Team",
		OwnerUserID: userID,
	}); err != nil {
		t.Fatal(err)
	}

	// 2. Seed repo
	repoID := idgen.NewRepo()
	if err := db.CreateRepo(ctx, store.Repo{
		ID:            repoID,
		NamespaceID:   nsID,
		Name:          "agent-repo",
		FullName:      "agentteam/agent-repo",
		DefaultBranch: "main",
		Visibility:    "private",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := git.Init("agentteam", "agent-repo", "main"); err != nil {
		t.Fatal(err)
	}
	sig := object.Signature{Name: "Admin", Email: "admin@exist.io", When: time.Now().UTC()}
	if _, err := git.WriteFiles("agentteam", "agent-repo", "main", []gitstore.FileChange{
		{Path: "README.md", Content: "# Agent Repo\n"},
	}, sig, "Initial commit"); err != nil {
		t.Fatal(err)
	}

	leaseSecret := []byte("agent-signing-secret-key-12345678")
	gitHandler := NewServer(db, git)
	gitHandler.SetLeaseSecret(leaseSecret)
	srv := httptest.NewServer(gitHandler)
	defer srv.Close()

	// 3. Issue valid capability lease
	tok, lease, err := agentlease.Issue(leaseSecret, "agent-007", repoID, []string{"refs/heads/agent/"}, []string{"refs/heads/main"}, 15*time.Minute)
	if err != nil {
		t.Fatalf("issue lease: %v", err)
	}

	// 4. Test real git clone with lease token as Basic Auth password
	cloneDir := filepath.Join(t.TempDir(), "cloned-agent")
	authCloneURL := strings.Replace(srv.URL, "http://", "http://agent:"+tok+"@", 1) + "/agentteam/agent-repo.git"

	cmd := exec.Command("git", "clone", authCloneURL, cloneDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone under lease failed: %v\nOutput: %s", err, string(out))
	}

	gitExec := func(dir string, args ...string) (string, error) {
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Agent007",
			"GIT_AUTHOR_EMAIL=agent007@exist.io",
			"GIT_COMMITTER_NAME=Agent007",
			"GIT_COMMITTER_EMAIL=agent007@exist.io",
		)
		out, err := c.CombinedOutput()
		return string(out), err
	}

	// 5. Create feature on refs/heads/agent/patch-1 and push -> MUST SUCCEED
	_, _ = gitExec(cloneDir, "checkout", "-b", "agent/patch-1")
	if err := os.WriteFile(filepath.Join(cloneDir, "agent_work.txt"), []byte("agent work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _ = gitExec(cloneDir, "add", "agent_work.txt")
	_, _ = gitExec(cloneDir, "commit", "-m", "agent patch")
	out, err := gitExec(cloneDir, "push", "origin", "agent/patch-1:refs/heads/agent/patch-1")
	if err != nil {
		t.Fatalf("push to refs/heads/agent/patch-1 failed: %v\nOutput: %s", err, out)
	}

	// 6. Push to refs/heads/main -> MUST FAIL with GH007
	out, err = gitExec(cloneDir, "push", "origin", "agent/patch-1:refs/heads/main")
	if err == nil {
		t.Fatalf("expected push to refs/heads/main to be rejected under lease, but succeeded: %s", out)
	}
	if !strings.Contains(out, "GH007") {
		t.Fatalf("expected GH007 error in push output, got: %s", out)
	}

	// 7. Expired lease -> MUST FAIL with 401
	expiredTok, _, _ := agentlease.Issue(leaseSecret, "agent-007", repoID, []string{"refs/heads/agent/"}, nil, -1*time.Minute)
	expiredURL := strings.Replace(srv.URL, "http://", "http://agent:"+expiredTok+"@", 1) + "/agentteam/agent-repo.git"
	out, err = gitExec(cloneDir, "push", expiredURL, "agent/patch-1:refs/heads/agent/patch-2")
	if err == nil {
		t.Fatalf("expected push with expired lease to fail, but succeeded: %s", out)
	}

	// 8. Revoked lease -> MUST FAIL with 401
	if err := db.RevokeAgentLease(ctx, lease.ID, "agent-007", repoID, "task cancelled"); err != nil {
		t.Fatalf("revoke lease: %v", err)
	}
	out, err = gitExec(cloneDir, "push", authCloneURL, "agent/patch-1:refs/heads/agent/patch-3")
	if err == nil {
		t.Fatalf("expected push with revoked lease to fail, but succeeded: %s", out)
	}
}
