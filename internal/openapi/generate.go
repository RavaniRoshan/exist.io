package openapi

import (
	_ "embed"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

//go:embed origin.openapi.yaml
var RawSpec []byte

var opIDRegex = regexp.MustCompile(`(?m)^\s*operationId:\s*([^\s#]+)`)

// Spec returns the raw embedded OpenAPI 3.1 YAML bytes.
func Spec() []byte {
	return RawSpec
}

// Handler returns an HTTP handler serving the OpenAPI 3.1 spec.
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(RawSpec)
	})
}

// ExtractOperationIDs parses the specification and returns all operation IDs found.
func ExtractOperationIDs() ([]string, error) {
	matches := opIDRegex.FindAllStringSubmatch(string(RawSpec), -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("openapi: no operations found in embedded spec")
	}

	var opIDs []string
	for _, m := range matches {
		opIDs = append(opIDs, strings.TrimSpace(m[1]))
	}
	return opIDs, nil
}

// ValidateCoverage checks that all operation IDs in the spec exist in the registered set.
func ValidateCoverage(registeredOps map[string]bool) ([]string, error) {
	opIDs, err := ExtractOperationIDs()
	if err != nil {
		return nil, err
	}

	var missing []string
	for _, id := range opIDs {
		if !registeredOps[id] {
			missing = append(missing, id)
		}
	}

	if len(missing) > 0 {
		return missing, fmt.Errorf("openapi: %d operations missing implementations: %s", len(missing), strings.Join(missing, ", "))
	}
	return nil, nil
}
