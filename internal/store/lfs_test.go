package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLFSStore(t *testing.T) {
	dataDir := t.TempDir()
	db, err := Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	oid := "4b97f339f408992c900609ec78465b53eb905b76fc87ab007ca5a8a7db32d321"

	// 1. Not found initially
	has, err := db.HasLFSObject(ctx, oid)
	if err != nil {
		t.Fatalf("HasLFSObject: %v", err)
	}
	if has {
		t.Fatalf("expected false, got true")
	}

	// 2. Put object
	obj := LFSObject{
		OID:         oid,
		Size:        1048576,
		StoragePath: "/tmp/lfs/4b/97/" + oid,
	}
	if err := db.PutLFSObject(ctx, obj); err != nil {
		t.Fatalf("PutLFSObject: %v", err)
	}

	// 3. Has and Get
	has, err = db.HasLFSObject(ctx, oid)
	if err != nil || !has {
		t.Fatalf("HasLFSObject after put: %v, has=%v", err, has)
	}

	fetched, err := db.GetLFSObject(ctx, oid)
	if err != nil {
		t.Fatalf("GetLFSObject: %v", err)
	}
	if fetched.Size != obj.Size || fetched.StoragePath != obj.StoragePath {
		t.Fatalf("unexpected fetched object: %+v", fetched)
	}

	// 4. Test Locks
	repoID := "repo_lfs_test"
	lock := LFSLock{
		ID:           "lock_123",
		RepositoryID: repoID,
		Path:         "assets/hero.psd",
		OwnerName:    "designer",
	}
	if err := db.CreateLFSLock(ctx, lock); err != nil {
		t.Fatalf("CreateLFSLock: %v", err)
	}

	// Duplicate lock on same repo and path must fail
	lockDup := LFSLock{
		ID:           "lock_456",
		RepositoryID: repoID,
		Path:         "assets/hero.psd",
		OwnerName:    "developer",
	}
	if err := db.CreateLFSLock(ctx, lockDup); err == nil {
		t.Fatalf("expected duplicate lock error, got nil")
	}

	// List locks
	locks, err := db.ListLFSLocks(ctx, repoID, "")
	if err != nil {
		t.Fatalf("ListLFSLocks: %v", err)
	}
	if len(locks) != 1 || locks[0].ID != "lock_123" {
		t.Fatalf("unexpected locks list: %+v", locks)
	}

	// Delete lock
	deleted, err := db.DeleteLFSLock(ctx, "lock_123")
	if err != nil {
		t.Fatalf("DeleteLFSLock: %v", err)
	}
	if deleted.Path != "assets/hero.psd" {
		t.Fatalf("unexpected deleted lock: %+v", deleted)
	}
}
