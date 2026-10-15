package gitlfs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"exist.io/internal/store"
)

func TestGitLFSFullLifecycle(t *testing.T) {
	dataDir := t.TempDir()
	db, err := store.Open(filepath.Join(dataDir, "meta.db"))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	defer db.Close()

	lfsStorage := filepath.Join(dataDir, "lfs")
	lfsServer := New(db, lfsStorage, "http://localhost:8080")

	owner := "acme"
	repo := "models"
	userID := "developer"

	// 1. Prepare sample binary object
	payload := []byte("Large machine learning tensor model weights binary content 0123456789")
	hash := sha256.Sum256(payload)
	oid := hex.EncodeToString(hash[:])
	size := int64(len(payload))

	// 2. Batch request for upload
	batchUploadReq := BatchRequest{
		Operation: "upload",
		Transfers: []string{"basic"},
		Objects: []BatchObj{
			{OID: oid, Size: size},
		},
	}
	bodyJSON, _ := json.Marshal(batchUploadReq)
	req := httptest.NewRequest(http.MethodPost, "/acme/models.git/info/lfs/objects/batch", bytes.NewReader(bodyJSON))
	rec := httptest.NewRecorder()
	lfsServer.HandleLFS(rec, req, owner, repo, "objects/batch", userID)

	if rec.Code != http.StatusOK {
		t.Fatalf("batch upload status = %d, want 200. body: %s", rec.Code, rec.Body.String())
	}

	var batchResp BatchResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &batchResp); err != nil {
		t.Fatalf("unmarshal batch response: %v", err)
	}
	if len(batchResp.Objects) != 1 {
		t.Fatalf("expected 1 object, got %d", len(batchResp.Objects))
	}
	uploadAction, ok := batchResp.Objects[0].Actions["upload"]
	if !ok || uploadAction.Href == "" {
		t.Fatalf("missing upload action: %+v", batchResp.Objects[0])
	}

	// 3. Upload object
	putReq := httptest.NewRequest(http.MethodPut, "/acme/models.git/info/lfs/objects/"+oid, bytes.NewReader(payload))
	putRec := httptest.NewRecorder()
	lfsServer.HandleLFS(putRec, putReq, owner, repo, "objects/"+oid, userID)
	if putRec.Code != http.StatusOK {
		t.Fatalf("upload object status = %d, want 200", putRec.Code)
	}

	// 4. Verify object
	verifyBody, _ := json.Marshal(VerifyRequest{OID: oid, Size: size})
	verifyReq := httptest.NewRequest(http.MethodPost, "/acme/models.git/info/lfs/objects/verify", bytes.NewReader(verifyBody))
	verifyRec := httptest.NewRecorder()
	lfsServer.HandleLFS(verifyRec, verifyReq, owner, repo, "objects/verify", userID)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify object status = %d, want 200", verifyRec.Code)
	}

	// 5. Batch request for download
	batchDownloadReq := BatchRequest{
		Operation: "download",
		Objects:   []BatchObj{{OID: oid, Size: size}},
	}
	dlJSON, _ := json.Marshal(batchDownloadReq)
	dlReq := httptest.NewRequest(http.MethodPost, "/acme/models.git/info/lfs/objects/batch", bytes.NewReader(dlJSON))
	dlRec := httptest.NewRecorder()
	lfsServer.HandleLFS(dlRec, dlReq, owner, repo, "objects/batch", userID)

	if dlRec.Code != http.StatusOK {
		t.Fatalf("batch download status = %d, want 200", dlRec.Code)
	}
	var dlResp BatchResponse
	_ = json.Unmarshal(dlRec.Body.Bytes(), &dlResp)
	if len(dlResp.Objects[0].Actions["download"].Href) == 0 {
		t.Fatalf("missing download action in response: %+v", dlResp.Objects[0])
	}

	// 6. Download object
	getReq := httptest.NewRequest(http.MethodGet, "/acme/models.git/info/lfs/objects/"+oid, nil)
	getRec := httptest.NewRecorder()
	lfsServer.HandleLFS(getRec, getReq, owner, repo, "objects/"+oid, userID)

	if getRec.Code != http.StatusOK {
		t.Fatalf("download object status = %d, want 200", getRec.Code)
	}
	if !bytes.Equal(getRec.Body.Bytes(), payload) {
		t.Fatalf("downloaded body does not match uploaded payload")
	}

	// 7. Test Locks
	lockBody := `{"path":"models/weights.bin"}`
	lockReq := httptest.NewRequest(http.MethodPost, "/acme/models.git/info/lfs/locks", strings.NewReader(lockBody))
	lockRec := httptest.NewRecorder()
	lfsServer.HandleLFS(lockRec, lockReq, owner, repo, "locks", userID)
	if lockRec.Code != http.StatusCreated {
		t.Fatalf("create lock status = %d, want 201", lockRec.Code)
	}

	var lockCreated struct {
		Lock struct {
			ID   string `json:"id"`
			Path string `json:"path"`
		} `json:"lock"`
	}
	_ = json.Unmarshal(lockRec.Body.Bytes(), &lockCreated)
	if lockCreated.Lock.Path != "models/weights.bin" {
		t.Fatalf("unexpected lock path: %s", lockCreated.Lock.Path)
	}

	// Unlock
	unlockReq := httptest.NewRequest(http.MethodPost, "/acme/models.git/info/lfs/locks/"+lockCreated.Lock.ID+"/unlock", strings.NewReader(`{}`))
	unlockRec := httptest.NewRecorder()
	lfsServer.HandleLFS(unlockRec, unlockReq, owner, repo, "locks/"+lockCreated.Lock.ID+"/unlock", userID)
	if unlockRec.Code != http.StatusOK {
		t.Fatalf("unlock status = %d, want 200", unlockRec.Code)
	}
}

func TestGitLFSHashMismatch(t *testing.T) {
	dataDir := t.TempDir()
	db, _ := store.Open(filepath.Join(dataDir, "meta.db"))
	defer db.Close()

	lfsStorage := filepath.Join(dataDir, "lfs")
	lfsServer := New(db, lfsStorage, "http://localhost:8080")

	// Provide a fake OID that doesn't match content
	fakeOID := strings.Repeat("a", 64)
	putReq := httptest.NewRequest(http.MethodPut, "/acme/models.git/info/lfs/objects/"+fakeOID, strings.NewReader("hello world"))
	putRec := httptest.NewRecorder()
	lfsServer.HandleLFS(putRec, putReq, "acme", "models", "objects/"+fakeOID, "user1")

	if putRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for hash mismatch, got %d", putRec.Code)
	}
}
