package store

import (
	"context"
	"testing"

	"exist.io/pkg/idgen"
)

func seedNamespace(t *testing.T, db *DB, slug string) Namespace {
	t.Helper()
	ns := Namespace{ID: idgen.NewNS(), Slug: slug, Type: "team", Name: slug}
	if err := db.CreateNamespace(context.Background(), ns); err != nil {
		t.Fatalf("CreateNamespace: %v", err)
	}
	return ns
}

func TestNamespaceCRUD(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()

	ns := seedNamespace(t, db, "acme")

	got, err := db.GetNamespaceBySlug(ctx, "acme")
	if err != nil {
		t.Fatalf("GetNamespaceBySlug: %v", err)
	}
	if got.ID != ns.ID || got.Type != "team" {
		t.Errorf("got %+v, want id %s", got, ns.ID)
	}

	if _, err := db.GetNamespaceBySlug(ctx, "nope"); err != ErrNotFound {
		t.Errorf("GetNamespaceBySlug(missing) = %v, want ErrNotFound", err)
	}

	// Duplicate slug conflicts.
	dup := Namespace{ID: idgen.NewNS(), Slug: "acme", Type: "team"}
	if err := db.CreateNamespace(ctx, dup); err == nil {
		t.Error("CreateNamespace(dup slug) = nil, want conflict")
	}
}

func TestRepoCRUD(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	ns := seedNamespace(t, db, "acme")

	repo := Repo{
		ID: idgen.NewRepo(), NamespaceID: ns.ID, Name: "widgets", FullName: "acme/widgets",
		DefaultBranch: "main", Visibility: "private",
		AllowMergeCommit: true, AllowSquashMerge: true,
	}
	if err := db.CreateRepo(ctx, repo); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	got, err := db.GetRepo(ctx, "acme", "widgets")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if got.ID != repo.ID || got.FullName != "acme/widgets" || got.DefaultBranch != "main" {
		t.Errorf("got %+v", got)
	}
	if !got.AllowMergeCommit || !got.AllowSquashMerge {
		t.Error("merge flags not persisted as true")
	}
	if got.DeleteBranchOnMerge {
		t.Error("deleteBranchOnMerge should be false")
	}

	byID, err := db.GetRepoByID(ctx, repo.ID)
	if err != nil || byID.Name != "widgets" {
		t.Fatalf("GetRepoByID = %+v, %v", byID, err)
	}

	if _, err := db.GetRepo(ctx, "acme", "ghost"); err != ErrNotFound {
		t.Errorf("GetRepo(missing) = %v, want ErrNotFound", err)
	}

	// Duplicate repo name in same namespace conflicts.
	if err := db.CreateRepo(ctx, Repo{ID: idgen.NewRepo(), NamespaceID: ns.ID, Name: "widgets", FullName: "acme/widgets", DefaultBranch: "main", Visibility: "private"}); err == nil {
		t.Error("CreateRepo(dup) = nil, want conflict")
	}
}

func TestRepoUpdate(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	seedNamespace(t, db, "acme")
	if err := db.CreateRepo(ctx, Repo{ID: idgen.NewRepo(), NamespaceID: mustNS(t, db, "acme"), Name: "w", FullName: "acme/w", DefaultBranch: "main", Visibility: "private"}); err != nil {
		t.Fatalf("CreateRepo: %v", err)
	}

	branch := "develop"
	if err := db.UpdateRepo(ctx, "acme", "w", RepoUpdate{DefaultBranch: &branch}); err != nil {
		t.Fatalf("UpdateRepo: %v", err)
	}
	got, _ := db.GetRepo(ctx, "acme", "w")
	if got.DefaultBranch != "develop" {
		t.Errorf("defaultBranch = %q, want develop", got.DefaultBranch)
	}

	// Update missing repo → ErrNotFound.
	if err := db.UpdateRepo(ctx, "acme", "ghost", RepoUpdate{DefaultBranch: &branch}); err != ErrNotFound {
		t.Errorf("UpdateRepo(missing) = %v, want ErrNotFound", err)
	}

	// Test description and archived status
	desc := "Enterprise quantum computing engine"
	arch := true
	if err := db.UpdateRepo(ctx, "acme", "w", RepoUpdate{Description: &desc, Archived: &arch}); err != nil {
		t.Fatalf("UpdateRepo(desc, arch): %v", err)
	}
	updated, err := db.GetRepo(ctx, "acme", "w")
	if err != nil {
		t.Fatalf("GetRepo: %v", err)
	}
	if updated.Description != desc {
		t.Errorf("got Description %q, want %q", updated.Description, desc)
	}
	if !updated.Archived {
		t.Errorf("got Archived %v, want true", updated.Archived)
	}

	// Test repository rename
	if err := db.RenameRepo(ctx, "acme", "w", "quantum"); err != nil {
		t.Fatalf("RenameRepo: %v", err)
	}
	renamed, err := db.GetRepo(ctx, "acme", "quantum")
	if err != nil {
		t.Fatalf("GetRepo(quantum): %v", err)
	}
	if renamed.Name != "quantum" || renamed.FullName != "acme/quantum" {
		t.Errorf("got name %q fullName %q, want quantum and acme/quantum", renamed.Name, renamed.FullName)
	}
	if _, err := db.GetRepo(ctx, "acme", "w"); err != ErrNotFound {
		t.Errorf("GetRepo(oldName) = %v, want ErrNotFound", err)
	}
}

func TestListReposPagination(t *testing.T) {
	db := openMem(t)
	ctx := context.Background()
	ns := seedNamespace(t, db, "acme")
	for i := 0; i < 5; i++ {
		name := "r" + string(rune('a'+i))
		if err := db.CreateRepo(ctx, Repo{ID: idgen.NewRepo(), NamespaceID: ns.ID, Name: name, FullName: "acme/" + name, DefaultBranch: "main", Visibility: "private"}); err != nil {
			t.Fatalf("CreateRepo: %v", err)
		}
	}

	// Page size 2 → 3 pages (2,2,1).
	page1, tok, err := db.ListRepos(ctx, "acme", 2, "")
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(page1) != 2 || tok == "" {
		t.Fatalf("page1 = %d repos, tok=%q; want 2 and a token", len(page1), tok)
	}
	page2, tok2, _ := db.ListRepos(ctx, "acme", 2, tok)
	if len(page2) != 2 || tok2 == "" {
		t.Fatalf("page2 = %d repos, tok=%q", len(page2), tok2)
	}
	page3, tok3, _ := db.ListRepos(ctx, "acme", 2, tok2)
	if len(page3) != 1 || tok3 != "" {
		t.Fatalf("page3 = %d repos, tok=%q; want 1 and empty token", len(page3), tok3)
	}

	// No overlap across pages.
	seen := map[string]bool{}
	for _, r := range append(append(page1, page2...), page3...) {
		if seen[r.ID] {
			t.Errorf("repo %s appeared twice", r.ID)
		}
		seen[r.ID] = true
	}
	if len(seen) != 5 {
		t.Errorf("saw %d distinct repos, want 5", len(seen))
	}
}

func mustNS(t *testing.T, db *DB, slug string) string {
	t.Helper()
	ns, err := db.GetNamespaceBySlug(context.Background(), slug)
	if err != nil {
		t.Fatalf("GetNamespaceBySlug: %v", err)
	}
	return ns.ID
}
