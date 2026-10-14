package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"exist.io/pkg/idgen"
)

type Ruleset struct {
	ID               string
	RepoID           string
	Name             string
	Description      string
	Enforcement      string // active | evaluate | disabled
	Kind             string // merge_branch | push_branch | push_tag | push_repository
	IncludedRefNames []string
	ExcludedRefNames []string
	Rules            []RulesetRule
	BypassActors     []RulesetBypassActor
	CreatedAt        string
	UpdatedAt        string
}

type RulesetRule struct {
	ID         string         `json:"id"`
	RuleType   string         `json:"ruleType"`
	Parameters map[string]any `json:"parameters,omitempty"`
}

type RulesetBypassActor struct {
	ID         string      `json:"id"`
	BypassMode string      `json:"bypassMode"` // always | pull_request_only
	User       *BypassUser `json:"user,omitempty"`
	Team       *BypassTeam `json:"team,omitempty"`
	App        *BypassApp  `json:"app,omitempty"`
	OriginRole *BypassRole `json:"originRole,omitempty"`
}

type BypassUser struct {
	ID string `json:"id"`
}

type BypassTeam struct {
	ID string `json:"id"`
}

type BypassApp struct {
	ID string `json:"id"`
}

type BypassRole struct {
	Role string `json:"role"`
}

// CreateRuleset stores a new ruleset for a repository.
func (db *DB) CreateRuleset(ctx context.Context, rs Ruleset) (Ruleset, error) {
	if rs.ID == "" {
		rs.ID = idgen.NewRuleset()
	}
	now := time.Now().UTC().Format(time.RFC3339)
	if rs.CreatedAt == "" {
		rs.CreatedAt = now
	}
	rs.UpdatedAt = now

	// Ensure sub-entity IDs
	for i := range rs.Rules {
		if rs.Rules[i].ID == "" {
			rs.Rules[i].ID = idgen.NewRulesetRule()
		}
	}
	for i := range rs.BypassActors {
		if rs.BypassActors[i].ID == "" {
			rs.BypassActors[i].ID = idgen.NewRulesetBypassActor()
		}
	}

	incJSON, _ := json.Marshal(rs.IncludedRefNames)
	excJSON, _ := json.Marshal(rs.ExcludedRefNames)
	rulesJSON, _ := json.Marshal(rs.Rules)
	bypassJSON, _ := json.Marshal(rs.BypassActors)

	_, err := db.sql.ExecContext(ctx, `
		INSERT INTO rulesets (
			id, repo_id, name, description, enforcement, kind,
			included_ref_names, excluded_ref_names, rules, bypass_actors,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, rs.ID, rs.RepoID, rs.Name, rs.Description, rs.Enforcement, rs.Kind,
		string(incJSON), string(excJSON), string(rulesJSON), string(bypassJSON),
		rs.CreatedAt, rs.UpdatedAt)
	if err != nil {
		return Ruleset{}, fmt.Errorf("insert ruleset: %w", err)
	}
	return rs, nil
}

// GetRuleset retrieves a ruleset by its ID.
func (db *DB) GetRuleset(ctx context.Context, id string) (Ruleset, error) {
	var rs Ruleset
	var incRaw, excRaw, rulesRaw, bypassRaw sql.NullString
	err := db.sql.QueryRowContext(ctx, `
		SELECT id, repo_id, name, IFNULL(description, ''), enforcement, kind,
		       included_ref_names, excluded_ref_names, rules, bypass_actors,
		       created_at, updated_at
		FROM rulesets
		WHERE id = ?
	`, id).Scan(
		&rs.ID, &rs.RepoID, &rs.Name, &rs.Description, &rs.Enforcement, &rs.Kind,
		&incRaw, &excRaw, &rulesRaw, &bypassRaw,
		&rs.CreatedAt, &rs.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Ruleset{}, ErrNotFound
	}
	if err != nil {
		return Ruleset{}, fmt.Errorf("get ruleset: %w", err)
	}

	if incRaw.Valid && incRaw.String != "" {
		_ = json.Unmarshal([]byte(incRaw.String), &rs.IncludedRefNames)
	}
	if excRaw.Valid && excRaw.String != "" {
		_ = json.Unmarshal([]byte(excRaw.String), &rs.ExcludedRefNames)
	}
	if rulesRaw.Valid && rulesRaw.String != "" {
		_ = json.Unmarshal([]byte(rulesRaw.String), &rs.Rules)
	}
	if bypassRaw.Valid && bypassRaw.String != "" {
		_ = json.Unmarshal([]byte(bypassRaw.String), &rs.BypassActors)
	}

	return rs, nil
}

// ListRulesets retrieves all rulesets for a repository.
func (db *DB) ListRulesets(ctx context.Context, repoID string) ([]Ruleset, error) {
	rows, err := db.sql.QueryContext(ctx, `
		SELECT id, repo_id, name, IFNULL(description, ''), enforcement, kind,
		       included_ref_names, excluded_ref_names, rules, bypass_actors,
		       created_at, updated_at
		FROM rulesets
		WHERE repo_id = ?
		ORDER BY created_at ASC, id ASC
	`, repoID)
	if err != nil {
		return nil, fmt.Errorf("list rulesets: %w", err)
	}
	defer rows.Close()

	var list []Ruleset
	for rows.Next() {
		var rs Ruleset
		var incRaw, excRaw, rulesRaw, bypassRaw sql.NullString
		if err := rows.Scan(
			&rs.ID, &rs.RepoID, &rs.Name, &rs.Description, &rs.Enforcement, &rs.Kind,
			&incRaw, &excRaw, &rulesRaw, &bypassRaw,
			&rs.CreatedAt, &rs.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if incRaw.Valid && incRaw.String != "" {
			_ = json.Unmarshal([]byte(incRaw.String), &rs.IncludedRefNames)
		}
		if excRaw.Valid && excRaw.String != "" {
			_ = json.Unmarshal([]byte(excRaw.String), &rs.ExcludedRefNames)
		}
		if rulesRaw.Valid && rulesRaw.String != "" {
			_ = json.Unmarshal([]byte(rulesRaw.String), &rs.Rules)
		}
		if bypassRaw.Valid && bypassRaw.String != "" {
			_ = json.Unmarshal([]byte(bypassRaw.String), &rs.BypassActors)
		}
		list = append(list, rs)
	}
	return list, rows.Err()
}

// UpdateRuleset updates an existing ruleset.
func (db *DB) UpdateRuleset(ctx context.Context, rs Ruleset) (Ruleset, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	rs.UpdatedAt = now

	for i := range rs.Rules {
		if rs.Rules[i].ID == "" {
			rs.Rules[i].ID = idgen.NewRulesetRule()
		}
	}
	for i := range rs.BypassActors {
		if rs.BypassActors[i].ID == "" {
			rs.BypassActors[i].ID = idgen.NewRulesetBypassActor()
		}
	}

	incJSON, _ := json.Marshal(rs.IncludedRefNames)
	excJSON, _ := json.Marshal(rs.ExcludedRefNames)
	rulesJSON, _ := json.Marshal(rs.Rules)
	bypassJSON, _ := json.Marshal(rs.BypassActors)

	res, err := db.sql.ExecContext(ctx, `
		UPDATE rulesets
		SET name = ?, description = ?, enforcement = ?, kind = ?,
		    included_ref_names = ?, excluded_ref_names = ?, rules = ?, bypass_actors = ?,
		    updated_at = ?
		WHERE id = ? AND repo_id = ?
	`, rs.Name, rs.Description, rs.Enforcement, rs.Kind,
		string(incJSON), string(excJSON), string(rulesJSON), string(bypassJSON),
		rs.UpdatedAt, rs.ID, rs.RepoID)
	if err != nil {
		return Ruleset{}, fmt.Errorf("update ruleset: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return Ruleset{}, err
	}
	if affected == 0 {
		return Ruleset{}, ErrNotFound
	}
	return rs, nil
}

// DeleteRuleset deletes a ruleset by ID.
func (db *DB) DeleteRuleset(ctx context.Context, repoID, id string) error {
	res, err := db.sql.ExecContext(ctx, `
		DELETE FROM rulesets
		WHERE id = ? AND repo_id = ?
	`, id, repoID)
	if err != nil {
		return fmt.Errorf("delete ruleset: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
