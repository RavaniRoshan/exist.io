package slug

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// slugRegex allows alphanumeric characters, underscores, dashes, and periods.
	// Must begin and end with an alphanumeric character, 1-100 characters in length.
	slugRegex = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9_\-\.]{0,98}[a-zA-Z0-9])?$`)

	// ErrInvalidSlug is returned when a slug contains forbidden characters or path separators.
	ErrInvalidSlug = errors.New("invalid slug: must be 1-100 characters containing only letters, numbers, hyphens, underscores, or periods")
)

// Validate checks whether a username, namespace, or repository slug is secure and conforms to standards.
// It explicitly rejects path traversal sequences (., .., /, \), control characters, and leading/trailing dashes/periods.
func Validate(s string) error {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 100 {
		return ErrInvalidSlug
	}
	if s == "." || s == ".." {
		return ErrInvalidSlug
	}
	if strings.ContainsAny(s, "/\\*?<>|:\"") {
		return ErrInvalidSlug
	}
	if strings.HasSuffix(strings.ToLower(s), ".git") {
		return errors.New("repository slug cannot end in .git")
	}
	if !slugRegex.MatchString(s) {
		return ErrInvalidSlug
	}
	return nil
}
