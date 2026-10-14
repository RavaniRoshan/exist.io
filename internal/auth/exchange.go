package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// exchangeResponse is the JSON body Origin returns from the exchange endpoint.
type exchangeResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// ExchangeAPIKey exchanges a raw Origin API key ("key_…" or "crsr_…") for an
// access/refresh token pair. It POSTs "{}" to
// {endpoint}/auth/exchange_user_api_key with the key sent as the bearer
// credential.
func ExchangeAPIKey(ctx context.Context, endpoint, key string) (accessToken, refreshToken string, err error) {
	url := strings.TrimSuffix(endpoint, "/") + "/auth/exchange_user_api_key"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return "", "", fmt.Errorf("auth: exchange request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("auth: exchange: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", "", fmt.Errorf("auth: exchange response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", "", fmt.Errorf("auth: exchange: unexpected status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var out exchangeResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return "", "", fmt.Errorf("auth: exchange response: %w", err)
	}
	if out.AccessToken == "" {
		return "", "", errors.New("auth: exchange response missing accessToken")
	}
	return out.AccessToken, out.RefreshToken, nil
}
