package slug

import "testing"

func TestValidateSlug(t *testing.T) {
	valid := []string{
		"repo-name",
		"exist.io",
		"cool_project",
		"a",
		"user123",
		"v1.0.0-release",
	}
	for _, s := range valid {
		if err := Validate(s); err != nil {
			t.Errorf("expected valid slug %q, got error: %v", s, err)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"../escape",
		"escape/path",
		"foo\\bar",
		"-start-dash",
		".start-dot",
		"my-repo.git",
		"has spaces",
		"has*wildcards",
		"has:colon",
		string(make([]byte, 101)),
	}
	for _, s := range invalid {
		if err := Validate(s); err == nil {
			t.Errorf("expected invalid slug %q to error, but passed", s)
		}
	}
}
