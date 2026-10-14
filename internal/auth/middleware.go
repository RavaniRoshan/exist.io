package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"exist.io/pkg/scopes"
)

// errMissingToken is returned by a Resolver when the request carries no usable
// credential. It maps to HTTP 401 / UNAUTHENTICATED.
var errMissingToken = errors.New("auth: missing bearer token")

// errMustExchange is returned when the credential is a raw API key ("key_…" or
// "crsr_…") that must be exchanged for an access token before use.
var errMustExchange = errors.New("auth: raw API key must be exchanged via ExchangeAPIKey before use")

// Resolver extracts the acting credential from a request. It returns a Token on
// success, or an error (typically errMissingToken or errMustExchange) that the
// middleware maps to a 401. Production wires a resolver that also looks up the
// credential's granted scopes from the store; tests inject a fixed Token.
type Resolver func(*http.Request) (Token, error)

// OperationMeta pairs an Origin operationId with the scope requirements derived
// from its x-origin-scopes extension. The registry (below) maps operationIds to
// these, and is populated incrementally as handler groups are added.
type OperationMeta struct {
	// ID is the Origin operationId, e.g. "OriginService_ListRepos".
	ID string
	// Reqs is the operation's x-origin-scopes requirement.
	Reqs scopes.Requirements
}

// Registry maps Origin operationIds to their scope requirements. It is the
// single source the middleware consults when enforcing an operation.
type Registry struct {
	ops map[string]scopes.Requirements
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{ops: make(map[string]scopes.Requirements)}
}

// Register records the scope requirements for an operationId.
func (r *Registry) Register(operationID string, reqs scopes.Requirements) {
	r.ops[operationID] = reqs
}

// Requirements returns the requirements registered for operationID.
func (r *Registry) Requirements(operationID string) (scopes.Requirements, bool) {
	reqs, ok := r.ops[operationID]
	return reqs, ok
}

// Operations returns the list of all registered operation IDs.
func (r *Registry) Operations() []string {
	out := make([]string, 0, len(r.ops))
	for op := range r.ops {
		out = append(out, op)
	}
	return out
}

// authMiddleware enforces op against the credential resolved from each request.
// On success it calls next; on failure it writes Origin's Status envelope and
// does not call next.
//
//   - Resolver error (no/expired/unusable credential) → 401 UNAUTHENTICATED
//   - scopes.Require denial (wrong type or missing scope) → 403 PERMISSION_DENIED
type tokenContextKey struct{}

// ContextToken extracts the acting Token from a guarded request context.
func ContextToken(ctx context.Context) (Token, bool) {
	tok, ok := ctx.Value(tokenContextKey{}).(Token)
	return tok, ok
}

// authMiddleware enforces op against the credential resolved from each request.
// On success it calls next; on failure it writes Origin's Status envelope and
// does not call next.
//
//   - Resolver error (no/expired/unusable credential) → 401 UNAUTHENTICATED
//   - scopes.Require denial (wrong type or missing scope) → 403 PERMISSION_DENIED
func authMiddleware(resolve Resolver, next http.Handler, op OperationMeta) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tok, err := resolve(r)
		if err != nil {
			writeStatus(w, http.StatusUnauthorized, scopes.CodeUnauthenticated, err.Error())
			return
		}
		if err := scopes.Require(tok, op.Reqs); err != nil {
			var denied *scopes.Denied
			if errors.As(err, &denied) {
				writeStatus(w, http.StatusForbidden, denied.Code, denied.Reason)
				return
			}
			writeStatus(w, http.StatusForbidden, scopes.CodePermissionDenied, err.Error())
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), tokenContextKey{}, tok))
		next.ServeHTTP(w, r)
	})
}

// Require wraps next with scope enforcement for the operation registered under
// operationID. It is the ergonomic entry point handlers use: register the
// operation's x-origin-scopes once, then guard the route by operationId. If the
// operation is not registered, the wrapped handler fails closed with 403.
func Require(reg *Registry, operationID string, next http.Handler) http.Handler {
	return RequireWith(reg, operationID, BearerResolver, next)
}

// RequireWith is Require with an explicit credential resolver, so callers (and
// tests) can supply a fixed credential with known scopes.
func RequireWith(reg *Registry, operationID string, resolve Resolver, next http.Handler) http.Handler {
	reqs, ok := reg.Requirements(operationID)
	if !ok {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			writeStatus(w, http.StatusForbidden, scopes.CodePermissionDenied,
				fmt.Sprintf("auth: operation %q is not registered", operationID))
		})
	}
	return authMiddleware(resolve, next, OperationMeta{ID: operationID, Reqs: reqs})
}

// BearerResolver parses the Authorization: Bearer header into a Token by
// credential shape. It does not resolve granted scopes (that requires the
// store); callers that need scoped enforcement wrap the returned Token.
func BearerResolver(r *http.Request) (Token, error) {
	raw, ok := bearerCredential(r)
	if !ok {
		return nil, errMissingToken
	}
	switch DetectType(raw) {
	case TypeInstallation:
		return NewInstallToken(raw), nil
	case TypeExchangeable:
		return nil, errMustExchange
	}
	// DetectType maps any unprefixed credential — including an app JWT — to
	// "user". Refine that here: a compact JWT whose subject is an app id
	// ("app_…") is an app credential.
	if appTok, err := ParseAppToken(raw); err == nil && strings.HasPrefix(appTok.Subject(), "app_") {
		return appTok, nil
	}
	return NewUserToken(raw), nil
}

// bearerCredential extracts the token from an "Authorization: Bearer <token>"
// header.
func bearerCredential(r *http.Request) (string, bool) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", false
	}
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	tok := strings.TrimSpace(h[len(prefix):])
	if tok == "" {
		return "", false
	}
	return tok, true
}

// statusEnvelope is Origin's canonical error body, after google.rpc.Status.
type statusEnvelope struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Details []json.RawMessage `json:"details"`
}

// writeStatus writes an Origin Status envelope with the given HTTP status,
// google.rpc.Code, and developer-facing message.
func writeStatus(w http.ResponseWriter, httpStatus, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(statusEnvelope{
		Code:    code,
		Message: message,
		Details: []json.RawMessage{},
	})
}
