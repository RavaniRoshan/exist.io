package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"exist.io/internal/audit"
	"exist.io/internal/store"
	"exist.io/pkg/idgen"
)

const (
	SCIMMediaType = "application/scim+json"
	UserSchema    = "urn:ietf:params:scim:schemas:core:2.0:User"
	ListSchema    = "urn:ietf:params:scim:api:messages:2.0:ListResponse"
	PatchSchema   = "urn:ietf:params:scim:api:messages:2.0:PatchOp"
)

// SCIMUser represents the SCIM 2.0 User resource (RFC 7643).
type SCIMUser struct {
	Schemas  []string    `json:"schemas"`
	ID       string      `json:"id"`
	UserName string      `json:"userName"`
	Name     SCIMName    `json:"name"`
	Emails   []SCIMEmail `json:"emails"`
	Active   bool        `json:"active"`
	Meta     SCIMMeta    `json:"meta"`
}

type SCIMName struct {
	Formatted  string `json:"formatted,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type SCIMMeta struct {
	ResourceType string `json:"resourceType"`
	Created      string `json:"created,omitempty"`
	Location     string `json:"location,omitempty"`
}

// SCIMListResponse represents a paginated search response (RFC 7644).
type SCIMListResponse struct {
	Schemas      []string   `json:"schemas"`
	TotalResults int        `json:"totalResults"`
	StartIndex   int        `json:"startIndex"`
	ItemsPerPage int        `json:"itemsPerPage"`
	Resources    []SCIMUser `json:"Resources"`
}

// SCIMPatchRequest represents an RFC 7644 PatchOp.
type SCIMPatchRequest struct {
	Schemas    []string        `json:"schemas"`
	Operations []SCIMOperation `json:"Operations"`
}

type SCIMOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path,omitempty"`
	Value any    `json:"value"`
}

// RegisterSCIMRoutes mounts all SCIM 2.0 endpoints onto the router.
func RegisterSCIMRoutes(mux *http.ServeMux, db *store.DB) {
	handler := &scimHandler{db: db}

	mux.HandleFunc("/scim/v2/ServiceProviderConfig", handler.serviceProviderConfig)
	mux.HandleFunc("GET /scim/v2/Users", handler.authenticate(handler.listUsers))
	mux.HandleFunc("POST /scim/v2/Users", handler.authenticate(handler.createUser))
	mux.HandleFunc("GET /scim/v2/Users/{id}", handler.authenticate(handler.getUser))
	mux.HandleFunc("PUT /scim/v2/Users/{id}", handler.authenticate(handler.replaceUser))
	mux.HandleFunc("PATCH /scim/v2/Users/{id}", handler.authenticate(handler.patchUser))
	mux.HandleFunc("DELETE /scim/v2/Users/{id}", handler.authenticate(handler.deleteUser))
}

type scimHandler struct {
	db *store.DB
}

func (h *scimHandler) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			scimError(w, http.StatusUnauthorized, "Bearer token required")
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		hash := HashCredential(token)

		// Check SCIM tokens table
		valid, err := h.db.ValidateSCIMToken(r.Context(), hash)
		if err == nil && valid {
			next(w, r)
			return
		}

		// Also permit active admin API keys
		key, err := h.db.GetAPIKeyByHash(r.Context(), hash)
		if err == nil && key.ID != "" {
			if key.RevokedAt.Valid && key.RevokedAt.String != "" {
				scimError(w, http.StatusUnauthorized, "SCIM API key has been revoked")
				return
			}
			if key.ExpiresAt.Valid && key.ExpiresAt.String != "" {
				exp, err := time.Parse(time.RFC3339, key.ExpiresAt.String)
				if err == nil && time.Now().After(exp) {
					scimError(w, http.StatusUnauthorized, "SCIM API key has expired")
					return
				}
			}
			user, err := h.db.GetUser(r.Context(), key.UserID)
			if err == nil && user.Role == "admin" && user.Status != "suspended" {
				next(w, r)
				return
			}
		}

		scimError(w, http.StatusUnauthorized, "Invalid SCIM bearer token")
	}
}

func (h *scimHandler) serviceProviderConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", SCIMMediaType)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"schemas": []string{"urn:ietf:params:scim:schemas:core:2.0:ServiceProviderConfig"},
		"patch":   map[string]bool{"supported": true},
		"bulk":    map[string]bool{"supported": false},
		"filter":  map[string]any{"supported": true, "maxResults": 1000},
		"changePassword": map[string]bool{"supported": false},
		"sort":           map[string]bool{"supported": false},
		"etag":           map[string]bool{"supported": false},
		"authenticationSchemes": []map[string]string{
			{
				"name":        "OAuth Bearer Token",
				"description": "Authentication scheme using the OAuth Bearer Standard",
				"specUri":     "http://www.rfc-editor.org/info/rfc6750",
				"type":        "oauthbearertoken",
			},
		},
	})
}

func (h *scimHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.db.ListUsers(ctx)
	if err != nil {
		scimError(w, http.StatusInternalServerError, err.Error())
		return
	}

	filter := r.URL.Query().Get("filter")
	var filtered []store.User

	if filter != "" {
		targetEmail := extractFilterValue(filter)
		for _, u := range users {
			if strings.EqualFold(u.Email, targetEmail) || strings.EqualFold(u.Login, targetEmail) {
				filtered = append(filtered, u)
			}
		}
	} else {
		filtered = users
	}

	startIndex := 1
	if s := r.URL.Query().Get("startIndex"); s != "" {
		if idx, err := strconv.Atoi(s); err == nil && idx > 0 {
			startIndex = idx
		}
	}

	count := 100
	if c := r.URL.Query().Get("count"); c != "" {
		if cnt, err := strconv.Atoi(c); err == nil && cnt > 0 {
			count = cnt
		}
	}

	var paged []store.User
	startZero := startIndex - 1
	if startZero < len(filtered) {
		end := startZero + count
		if end > len(filtered) {
			end = len(filtered)
		}
		paged = filtered[startZero:end]
	}

	resources := make([]SCIMUser, 0, len(paged))
	for _, u := range paged {
		resources = append(resources, toSCIMUser(u))
	}

	resp := SCIMListResponse{
		Schemas:      []string{ListSchema},
		TotalResults: len(filtered),
		StartIndex:   startIndex,
		ItemsPerPage: len(resources),
		Resources:    resources,
	}

	w.Header().Set("Content-Type", SCIMMediaType)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *scimHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var in SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		scimError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	email := in.UserName
	if len(in.Emails) > 0 && in.Emails[0].Value != "" {
		email = in.Emails[0].Value
	}
	if email == "" {
		scimError(w, http.StatusBadRequest, "userName or email is required")
		return
	}

	login := in.UserName
	if strings.Contains(login, "@") {
		login = strings.Split(login, "@")[0]
	}

	name := in.Name.Formatted
	if name == "" {
		name = strings.TrimSpace(in.Name.GivenName + " " + in.Name.FamilyName)
	}

	status := "active"
	if !in.Active {
		status = "suspended"
	}

	// Create user with locked initial password hash
	randomPass, _, _ := GenerateSessionToken()
	hash, _ := HashPassword(randomPass)

	user := store.User{
		ID:           idgen.NewUser(),
		Email:        email,
		Login:        login,
		Name:         name,
		PasswordHash: hash,
		Role:         "member",
		Status:       status,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	if err := h.db.CreateUserWithPassword(r.Context(), user); err != nil {
		scimError(w, http.StatusConflict, "User already exists or email conflict")
		return
	}

	_ = audit.Log(audit.Event{
		Actor:     "scim",
		ActorType: "service",
		Action:    "scim.user_provisioned",
		Resource:  user.Email,
		Severity:  audit.SeverityInfo,
		Details:   "User provisioned via SCIM 2.0",
	})

	w.Header().Set("Content-Type", SCIMMediaType)
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(toSCIMUser(user))
}

func (h *scimHandler) getUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.db.GetUser(r.Context(), id)
	if err != nil {
		scimError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", SCIMMediaType)
	_ = json.NewEncoder(w).Encode(toSCIMUser(user))
}

func (h *scimHandler) replaceUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	var in SCIMUser
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		scimError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	status := "active"
	if !in.Active {
		status = "suspended"
	}
	_ = h.db.UpdateUserStatus(ctx, id, status)

	user, err := h.db.GetUser(ctx, id)
	if err != nil {
		scimError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", SCIMMediaType)
	_ = json.NewEncoder(w).Encode(toSCIMUser(user))
}

func (h *scimHandler) patchUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ctx := r.Context()

	var req SCIMPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		scimError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	for _, op := range req.Operations {
		if strings.EqualFold(op.Op, "replace") {
			// Check if setting active status
			if strings.EqualFold(op.Path, "active") {
				if b, ok := op.Value.(bool); ok {
					newStatus := "active"
					if !b {
						newStatus = "suspended"
					}
					_ = h.db.UpdateUserStatus(ctx, id, newStatus)
				}
			} else if opMap, ok := op.Value.(map[string]any); ok {
				if activeVal, hasActive := opMap["active"]; hasActive {
					if b, ok := activeVal.(bool); ok {
						newStatus := "active"
						if !b {
							newStatus = "suspended"
						}
						_ = h.db.UpdateUserStatus(ctx, id, newStatus)
					}
				}
			}
		}
	}

	user, err := h.db.GetUser(ctx, id)
	if err != nil {
		scimError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", SCIMMediaType)
	_ = json.NewEncoder(w).Encode(toSCIMUser(user))
}

func (h *scimHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	// SCIM deprovision: set status to suspended
	err := h.db.UpdateUserStatus(r.Context(), id, "suspended")
	if err != nil && errors.Is(err, store.ErrNotFound) {
		scimError(w, http.StatusNotFound, "User not found")
		return
	}

	_ = audit.Log(audit.Event{
		Actor:     "scim",
		ActorType: "service",
		Action:    "scim.user_deprovisioned",
		Resource:  id,
		Severity:  audit.SeverityWarning,
		Details:   "User suspended/deprovisioned via SCIM 2.0",
	})

	w.WriteHeader(http.StatusNoContent)
}

func toSCIMUser(u store.User) SCIMUser {
	isActive := u.Status == "active"
	return SCIMUser{
		Schemas:  []string{UserSchema},
		ID:       u.ID,
		UserName: u.Email,
		Name: SCIMName{
			Formatted: u.Name,
		},
		Emails: []SCIMEmail{
			{Value: u.Email, Type: "work", Primary: true},
		},
		Active: isActive,
		Meta: SCIMMeta{
			ResourceType: "User",
			Created:      u.CreatedAt,
			Location:     "/scim/v2/Users/" + u.ID,
		},
	}
}

func extractFilterValue(filter string) string {
	// e.g. userName eq "alice@example.com" or emails[type eq "work" and value eq "alice@example.com"]
	parts := strings.Split(filter, "\"")
	if len(parts) >= 2 {
		return parts[1]
	}
	fields := strings.Fields(filter)
	if len(fields) >= 3 && strings.EqualFold(fields[1], "eq") {
		return strings.Trim(fields[2], "'\"")
	}
	return filter
}

func scimError(w http.ResponseWriter, code int, detail string) {
	w.Header().Set("Content-Type", SCIMMediaType)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"schemas": []string{"urn:ietf:params:scim:api:messages:2.0:Error"},
		"status":  fmt.Sprintf("%d", code),
		"detail":  detail,
	})
}
