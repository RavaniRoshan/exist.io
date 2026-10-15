package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a typed API client for interacting with the exist.io Git forge.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient returns a configured Client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, in, out any) (int, error) {
	url := c.BaseURL + path
	var bodyReader io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return 0, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "terraform-provider-exist/1.0.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http execute %s %s: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return resp.StatusCode, fmt.Errorf("api error (%d): %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return resp.StatusCode, fmt.Errorf("unmarshal response: %w", err)
		}
	}

	return resp.StatusCode, nil
}

// Namespace models
type Namespace struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (c *Client) GetNamespace(ctx context.Context, slug string) (*Namespace, error) {
	var resp struct {
		Namespace Namespace `json:"namespace"`
	}
	_, err := c.do(ctx, http.MethodGet, "/v1/origin/namespaces/"+slug, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Namespace, nil
}

func (c *Client) CreateNamespace(ctx context.Context, slug, name, nsType string) (*Namespace, error) {
	body := map[string]string{
		"slug": slug,
		"name": name,
		"type": nsType,
	}
	var resp struct {
		Namespace Namespace `json:"namespace"`
	}
	_, err := c.do(ctx, http.MethodPost, "/v1/origin/namespaces", body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Namespace, nil
}

// Repository models
type Repository struct {
	ID            string `json:"id"`
	Owner         struct {
		ID   string `json:"id"`
		Slug string `json:"slug"`
	} `json:"owner"`
	Name          string `json:"name"`
	DefaultBranch string `json:"defaultBranch"`
	Visibility    string `json:"visibility"`
}

func (c *Client) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	var resp Repository
	_, err := c.do(ctx, http.MethodGet, fmt.Sprintf("/v1/origin/repos/%s/%s", owner, name), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CreateRepository(ctx context.Context, owner, name, defaultBranch, visibility string) (*Repository, error) {
	body := map[string]any{
		"ownerSlug": owner,
		"repo": map[string]string{
			"name":          name,
			"defaultBranch": defaultBranch,
			"visibility":    visibility,
		},
	}
	var resp Repository
	_, err := c.do(ctx, http.MethodPost, "/v1/origin/repos/"+owner, body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Ruleset models
type Ruleset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Enforcement string `json:"enforcement"`
}

func (c *Client) CreateRuleset(ctx context.Context, owner, repo string, ruleset any) (map[string]any, error) {
	var resp map[string]any
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/origin/repos/%s/%s/rulesets", owner, repo), ruleset, &resp)
	return resp, err
}

// IP Allowlist models
func (c *Client) AddIPAllowlistEntry(ctx context.Context, nsSlug, cidr, desc string) (map[string]any, error) {
	body := map[string]string{
		"cidr":        cidr,
		"description": desc,
	}
	var resp map[string]any
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/origin/namespaces/%s/inbound-ip-allowlist/entries", nsSlug), body, &resp)
	return resp, err
}

// SSH CA models
func (c *Client) AddSSHCA(ctx context.Context, nsSlug, name, pubKey string) (map[string]any, error) {
	body := map[string]string{
		"name":      name,
		"publicKey": pubKey,
	}
	var resp map[string]any
	_, err := c.do(ctx, http.MethodPost, fmt.Sprintf("/v1/origin/namespaces/%s/ssh-certificate-authorities", nsSlug), body, &resp)
	return resp, err
}
