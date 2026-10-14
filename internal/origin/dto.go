package origin

import "exist.io/internal/store"

// This file maps store entities to Origin wire DTOs. JSON field names are
// camelCase, matching the Origin OpenAPI 3.1 spec exactly.

// repoDTO is the Origin Repo response shape.
type repoDTO struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	FullName            string     `json:"fullName"`
	Owner               ownerDTO   `json:"owner"`
	DefaultBranch       string     `json:"defaultBranch"`
	Visibility          string     `json:"visibility"`
	CloneURL            string     `json:"cloneUrl"`
	WebURL              string     `json:"webUrl,omitempty"`
	Mirror              *mirrorDTO `json:"mirror,omitempty"`
	AllowMergeCommit    bool       `json:"allowMergeCommit"`
	AllowSquashMerge    bool       `json:"allowSquashMerge"`
	DeleteBranchOnMerge bool       `json:"deleteBranchOnMerge"`
	CreatedAt           string     `json:"createdAt"`
	UpdatedAt           string     `json:"updatedAt"`
	PushedAt            string     `json:"pushedAt,omitempty"`
}

// ownerDTO is the Origin Owner (namespace reference) shape. Note: Origin's
// fullName derives from owner.login, and the Owner object uses slug.
type ownerDTO struct {
	ID   string `json:"id,omitempty"`
	Slug string `json:"slug"`
	Type string `json:"type,omitempty"` // team | user; output-only
}

// mirrorDTO is the Origin RepositoryMirror shape.
type mirrorDTO struct {
	Source   string `json:"source"` // github
	SourceID string `json:"sourceId,omitempty"`
	Status   string `json:"status,omitempty"`
}

// namespaceDTO is the Origin NamespaceListing shape.
type namespaceDTO struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// toRepoDTO converts a store.Repo (+ its namespace) to the wire DTO.
func (s *Server) toRepoDTO(r store.Repo, ns store.Namespace) repoDTO {
	dto := repoDTO{
		ID:                  r.ID,
		Name:                r.Name,
		FullName:            r.FullName,
		Owner:               ownerDTO{ID: ns.ID, Slug: ns.Slug, Type: ns.Type},
		DefaultBranch:       r.DefaultBranch,
		Visibility:          r.Visibility,
		CloneURL:            r.CloneURL,
		WebURL:              r.WebURL,
		AllowMergeCommit:    r.AllowMergeCommit,
		AllowSquashMerge:    r.AllowSquashMerge,
		DeleteBranchOnMerge: r.DeleteBranchOnMerge,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
		PushedAt:            r.PushedAt,
	}
	if r.MirrorSource != "" {
		dto.Mirror = &mirrorDTO{Source: r.MirrorSource, SourceID: r.MirrorSourceID, Status: r.MirrorStatus}
	}
	return dto
}

func toNamespaceDTO(ns store.Namespace) namespaceDTO {
	return namespaceDTO{ID: ns.ID, Slug: ns.Slug, Type: ns.Type, Name: ns.Name}
}
