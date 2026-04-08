package zerodha

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	kiteBaseURL    = "https://kite.zerodha.com"
	kiteAPIBaseURL = "https://api.kite.trade"
	kiteVersion    = "3"
)

// TokenRecord holds a persisted access token and its expiry time.
type TokenRecord struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// LoginURL returns the Kite Connect browser login URL for the given API key.
func LoginURL(apiKey string) string {
	return fmt.Sprintf("%s/connect/login?v=%s&api_key=%s", kiteBaseURL, kiteVersion, apiKey)
}

// Checksum computes SHA-256(apiKey + requestToken + apiSecret) hex-encoded,
// as required by the Kite Connect session exchange endpoint.
func Checksum(apiKey, requestToken, apiSecret string) string {
	h := sha256.New()
	h.Write([]byte(apiKey + requestToken + apiSecret))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ExchangeToken POSTs to the Kite Connect session endpoint and returns the access token.
// client and baseURL are injectable for testing.
func ExchangeToken(ctx context.Context, client *http.Client, baseURL, apiKey, requestToken, apiSecret string) (string, error) {
	checksum := Checksum(apiKey, requestToken, apiSecret)

	form := url.Values{}
	form.Set("api_key", apiKey)
	form.Set("request_token", requestToken)
	form.Set("checksum", checksum)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		baseURL+"/session/token",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("build session exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Kite-Version", kiteVersion)

	body, err := doHTTP(client, req)
	if err != nil {
		return "", fmt.Errorf("session exchange: %w", err)
	}

	var envelope struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", fmt.Errorf("parse session response: %w", err)
	}
	if envelope.Data.AccessToken == "" {
		return "", fmt.Errorf("session exchange: empty access_token in response")
	}
	return envelope.Data.AccessToken, nil
}

// SaveToken writes a TokenRecord to path. The expiry is set to the next
// occurrence of 6:00 AM IST (00:30 UTC), which is when Kite tokens expire.
// Parent directories are created as needed with mode 0700.
// The file is written with mode 0600 (owner read/write only).
func SaveToken(path, accessToken string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}

	rec := TokenRecord{
		AccessToken: accessToken,
		ExpiresAt:   nextKiteExpiry(time.Now().UTC()),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("marshal token: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	return nil
}

// LoadToken reads the token record from path and returns the access token.
// Returns ErrAuthRequired (wrapped) if the file is missing, the token is
// empty, or the token has expired.
func LoadToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrAuthRequired
		}
		return "", fmt.Errorf("read token file: %w", err)
	}

	var rec TokenRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return "", fmt.Errorf("parse token file: %w", err)
	}

	if rec.AccessToken == "" || time.Now().UTC().After(rec.ExpiresAt) {
		return "", ErrAuthRequired
	}
	return rec.AccessToken, nil
}

// nextKiteExpiry returns the next 6:00 AM IST (00:30 UTC) after now.
// If now is before today's 00:30 UTC, returns today's 00:30 UTC.
// Otherwise returns tomorrow's 00:30 UTC.
func nextKiteExpiry(now time.Time) time.Time {
	todayExpiry := time.Date(now.Year(), now.Month(), now.Day(), 0, 30, 0, 0, time.UTC)
	if now.Before(todayExpiry) {
		return todayExpiry
	}
	return todayExpiry.AddDate(0, 0, 1)
}
