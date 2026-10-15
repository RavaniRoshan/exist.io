package gitlfs

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

const (
	LFSMediaType = "application/vnd.git-lfs+json"
)

// Server implements the Git Large File Storage (LFS) v1 HTTP API.
type Server struct {
	store     *store.DB
	storageDir string
	publicURL string
}

// New creates a new Git LFS Server.
func New(db *store.DB, storageDir, publicURL string) *Server {
	return &Server{
		store:      db,
		storageDir: storageDir,
		publicURL:  strings.TrimRight(publicURL, "/"),
	}
}

// BatchRequest represents a batch request from git-lfs.
type BatchRequest struct {
	Operation string     `json:"operation"`
	Transfers []string   `json:"transfers,omitempty"`
	Ref       *BatchRef  `json:"ref,omitempty"`
	Objects   []BatchObj `json:"objects"`
	HashAlgo  string     `json:"hash_algo,omitempty"`
}

type BatchRef struct {
	Name string `json:"name"`
}

type BatchObj struct {
	OID  string `json:"oid"`
	Size int64  `json:"size"`
}

type BatchResponse struct {
	Transfer string             `json:"transfer"`
	Objects  []BatchObjResponse `json:"objects"`
}

type BatchObjResponse struct {
	OID           string                 `json:"oid"`
	Size          int64                  `json:"size"`
	Authenticated bool                   `json:"authenticated,omitempty"`
	Actions       map[string]BatchAction `json:"actions,omitempty"`
	Error         *BatchError            `json:"error,omitempty"`
}

type BatchAction struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresAt string            `json:"expires_at,omitempty"`
	ExpiresIn int64             `json:"expires_in,omitempty"`
}

type BatchError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type VerifyRequest struct {
	OID  string `json:"oid"`
	Size int64  `json:"size"`
}

// HandleLFS processes all Git LFS requests routed under a repository.
// lfsPath is the suffix after /{owner}/{repo}.git/info/lfs (e.g. "/objects/batch" or "/objects/4b97f...").
func (s *Server) HandleLFS(w http.ResponseWriter, r *http.Request, ownerSlug, repoName, lfsPath, userID string) {
	lfsPath = strings.TrimPrefix(lfsPath, "/")

	switch {
	case lfsPath == "objects/batch":
		s.handleBatch(w, r, ownerSlug, repoName)
	case strings.HasPrefix(lfsPath, "objects/") && !strings.HasSuffix(lfsPath, "/verify"):
		oid := strings.TrimPrefix(lfsPath, "objects/")
		if r.Method == http.MethodPut {
			s.handleUpload(w, r, oid)
		} else if r.Method == http.MethodGet || r.Method == http.MethodHead {
			s.handleDownload(w, r, oid)
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	case lfsPath == "objects/verify":
		s.handleVerify(w, r)
	case lfsPath == "locks" || strings.HasPrefix(lfsPath, "locks/"):
		s.handleLocks(w, r, ownerSlug, repoName, lfsPath, userID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleBatch(w http.ResponseWriter, r *http.Request, ownerSlug, repoName string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp := BatchResponse{
		Transfer: "basic",
		Objects:  make([]BatchObjResponse, 0, len(req.Objects)),
	}

	baseURL := fmt.Sprintf("%s/%s/%s.git/info/lfs/objects", s.publicURL, ownerSlug, repoName)

	for _, obj := range req.Objects {
		objResp := BatchObjResponse{
			OID:           obj.OID,
			Size:          obj.Size,
			Authenticated: true,
		}

		exists, err := s.store.HasLFSObject(ctx, obj.OID)
		if err != nil {
			http.Error(w, "internal database error", http.StatusInternalServerError)
			return
		}

		if req.Operation == "upload" {
			if !exists {
				objResp.Actions = map[string]BatchAction{
					"upload": {
						Href:      fmt.Sprintf("%s/%s", baseURL, obj.OID),
						ExpiresIn: 86400,
					},
					"verify": {
						Href:      fmt.Sprintf("%s/verify", baseURL),
						ExpiresIn: 86400,
					},
				}
			}
		} else if req.Operation == "download" {
			if exists {
				objResp.Actions = map[string]BatchAction{
					"download": {
						Href:      fmt.Sprintf("%s/%s", baseURL, obj.OID),
						ExpiresIn: 86400,
					},
				}
			} else {
				objResp.Error = &BatchError{
					Code:    http.StatusNotFound,
					Message: "Object does not exist",
				}
			}
		}

		resp.Objects = append(resp.Objects, objResp)
	}

	w.Header().Set("Content-Type", LFSMediaType)
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request, oid string) {
	if len(oid) != 64 {
		http.Error(w, "invalid oid format", http.StatusBadRequest)
		return
	}

	tmpDir := filepath.Join(s.storageDir, "tmp")
	_ = os.MkdirAll(tmpDir, 0755)

	tmpFile, err := os.CreateTemp(tmpDir, "lfs-upload-*")
	if err != nil {
		http.Error(w, "failed to create temporary file", http.StatusInternalServerError)
		return
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)

	written, err := io.Copy(writer, r.Body)
	if err != nil {
		http.Error(w, "failed to write payload", http.StatusInternalServerError)
		return
	}
	_ = tmpFile.Close()

	computedOID := hex.EncodeToString(hasher.Sum(nil))
	if !strings.EqualFold(computedOID, oid) {
		http.Error(w, fmt.Sprintf("corrupt object: computed sha256 %s != expected %s", computedOID, oid), http.StatusBadRequest)
		return
	}

	destDir := filepath.Join(s.storageDir, oid[0:2], oid[2:4])
	if err := os.MkdirAll(destDir, 0755); err != nil {
		http.Error(w, "failed to create storage directory", http.StatusInternalServerError)
		return
	}

	destPath := filepath.Join(destDir, oid)
	if err := os.Rename(tmpFile.Name(), destPath); err != nil {
		// Fallback to copy if rename across filesystems fails
		if err := copyFile(tmpFile.Name(), destPath); err != nil {
			http.Error(w, "failed to persist object", http.StatusInternalServerError)
			return
		}
	}

	err = s.store.PutLFSObject(r.Context(), store.LFSObject{
		OID:         oid,
		Size:        written,
		StoragePath: destPath,
	})
	if err != nil {
		http.Error(w, "failed to record metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request, oid string) {
	ctx := r.Context()
	obj, err := s.store.GetLFSObject(ctx, oid)
	if err != nil {
		if errors.Is(err, store.ErrLFSNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	f, err := os.Open(obj.StoragePath)
	if err != nil {
		http.Error(w, "object file missing from disk", http.StatusNotFound)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		http.Error(w, "stat error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, oid, fi.ModTime(), f)
}

func (s *Server) handleVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req VerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	obj, err := s.store.GetLFSObject(r.Context(), req.OID)
	if err != nil || obj.Size != req.Size {
		http.Error(w, "object verification failed", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLocks(w http.ResponseWriter, r *http.Request, ownerSlug, repoName, lfsPath, userID string) {
	repoID := ownerSlug + "/" + repoName
	ctx := r.Context()

	if lfsPath == "locks" {
		if r.Method == http.MethodGet {
			path := r.URL.Query().Get("path")
			locks, err := s.store.ListLFSLocks(ctx, repoID, path)
			if err != nil {
				http.Error(w, "database error", http.StatusInternalServerError)
				return
			}
			var list []map[string]any
			for _, l := range locks {
				list = append(list, map[string]any{
					"id":        l.ID,
					"path":      l.Path,
					"locked_at": l.LockedAt,
					"owner":     map[string]string{"name": l.OwnerName},
				})
			}
			w.Header().Set("Content-Type", LFSMediaType)
			_ = json.NewEncoder(w).Encode(map[string]any{"locks": list})
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				Path string `json:"path"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Path == "" {
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			}

			lock := store.LFSLock{
				ID:           idgen.New("lock"),
				RepositoryID: repoID,
				Path:         req.Path,
				OwnerName:    userID,
				LockedAt:     time.Now().UTC().Format(time.RFC3339),
			}

			if err := s.store.CreateLFSLock(ctx, lock); err != nil {
				w.Header().Set("Content-Type", LFSMediaType)
				w.WriteHeader(http.StatusConflict)
				_ = json.NewEncoder(w).Encode(map[string]any{
					"message": "already locked",
				})
				return
			}

			w.Header().Set("Content-Type", LFSMediaType)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"lock": map[string]any{
					"id":        lock.ID,
					"path":      lock.Path,
					"locked_at": lock.LockedAt,
					"owner":     map[string]string{"name": lock.OwnerName},
				},
			})
			return
		}
	}

	if strings.HasSuffix(lfsPath, "/unlock") && r.Method == http.MethodPost {
		id := strings.TrimSuffix(strings.TrimPrefix(lfsPath, "locks/"), "/unlock")
		lock, err := s.store.DeleteLFSLock(ctx, id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", LFSMediaType)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"lock": map[string]any{
				"id":        lock.ID,
				"path":      lock.Path,
				"locked_at": lock.LockedAt,
				"owner":     map[string]string{"name": lock.OwnerName},
			},
		})
		return
	}

	http.NotFound(w, r)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
