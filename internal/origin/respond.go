// Package origin implements the Origin REST surface (/v1/origin/*) as HTTP
// handlers over the metadata store and git object store. Handler files are
// grouped by resource (handlers_repo.go, handlers_grant.go, ...); this file
// holds the shared response/error plumbing every group uses.
//
// Wire-format rules (from the Origin OpenAPI 3.1 spec), enforced here so every
// handler is consistent:
//   - JSON field names are camelCase (struct json tags).
//   - Success bodies are the resource or a list envelope ({nextPageToken, items}).
//   - Errors use the google.rpc.Status envelope {code, message, details} with
//     the HTTP status derived from the code.
package origin

import (
	"encoding/json"
	"net/http"
)

// writeJSON writes v as a JSON body with the given HTTP status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// rpcCode mirrors the subset of google.rpc.Code values Origin returns, paired
// with their canonical HTTP status.
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details []any  `json:"details"`
}

// writeError writes an Origin Status error envelope. httpStatus is the HTTP
// status to send; code is the google.rpc.Code; message is developer-facing.
func writeError(w http.ResponseWriter, httpStatus, code int, message string) {
	writeJSON(w, httpStatus, rpcError{
		Code:    code,
		Message: message,
		Details: []any{},
	})
}

// Convenience writers for the codes handlers hit most. Origin deliberately does
// not distinguish "not found" from "no access" — both are 404 NOT_FOUND.
func notFound(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusNotFound, codeNotFound, msg)
}

func badRequest(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusBadRequest, codeInvalidArgument, msg)
}

func conflict(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusConflict, codeAlreadyExists, msg)
}

func permissionDenied(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusForbidden, codePermissionDenied, msg)
}

func internalError(w http.ResponseWriter, msg string) {
	writeError(w, http.StatusInternalServerError, codeInternal, msg)
}

// google.rpc.Code values used across handlers.
const (
	codeOK                 = 0
	codeInvalidArgument    = 3
	codeNotFound           = 5
	codeAlreadyExists      = 6
	codePermissionDenied   = 7
	codeFailedPrecondition = 9
	codeInternal           = 13
	codeUnauthenticated    = 16
)

// decodeJSON reads and JSON-decodes the request body into dst. A malformed body
// is a 400 invalid-argument.
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return nil
	}
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	return dec.Decode(dst)
}

// queryInt parses an optional integer query parameter, returning def when
// absent or invalid.
func pathParam(r *http.Request, key string) string {
	return r.PathValue(key)
}
