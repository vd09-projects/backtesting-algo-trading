package zerodha

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestChecksum(t *testing.T) {
	// Known SHA-256(apiKey + requestToken + apiSecret) value.
	// Verified: echo -n "mykeymytokenmysecret" | sha256sum
	got := Checksum("mykey", "mytoken", "mysecret")
	want := "4eea113ceed06c7a929529a2404bb08ba045af145da927f23fd7f30d6dcdbc45"
	if got != want {
		t.Errorf("Checksum: want %s, got %s", want, got)
	}
}

func TestLoginURL(t *testing.T) {
	got := LoginURL("testapikey")
	if got == "" {
		t.Fatal("LoginURL returned empty string")
	}
	// Must contain the api_key and point at the Kite login domain.
	for _, want := range []string{"testapikey", "kite.zerodha.com", "connect/login"} {
		if !strings.Contains(got, want) {
			t.Errorf("LoginURL %q does not contain %q", got, want)
		}
	}
}

func TestNextKiteExpiry(t *testing.T) {
	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before today 00:30 UTC — returns today 00:30",
			now:  time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
			want: time.Date(2024, 6, 15, 0, 30, 0, 0, time.UTC),
		},
		{
			name: "exactly at 00:30 UTC — returns tomorrow 00:30 (not expired yet but treat as boundary)",
			now:  time.Date(2024, 6, 15, 0, 30, 0, 0, time.UTC),
			want: time.Date(2024, 6, 16, 0, 30, 0, 0, time.UTC),
		},
		{
			name: "after 00:30 UTC — returns tomorrow 00:30",
			now:  time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC),
			want: time.Date(2024, 6, 16, 0, 30, 0, 0, time.UTC),
		},
		{
			name: "end of day — returns tomorrow 00:30",
			now:  time.Date(2024, 6, 15, 23, 59, 59, 0, time.UTC),
			want: time.Date(2024, 6, 16, 0, 30, 0, 0, time.UTC),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := nextKiteExpiry(tc.now)
			if !got.Equal(tc.want) {
				t.Errorf("nextKiteExpiry(%v) = %v, want %v", tc.now, got, tc.want)
			}
		})
	}
}

func TestSaveToken_LoadToken_roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	const token = "test-access-token-abc123"
	if err := SaveToken(path, token); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}

	got, err := LoadToken(path)
	if err != nil {
		t.Fatalf("LoadToken: %v", err)
	}
	if got != token {
		t.Errorf("LoadToken: want %q, got %q", token, got)
	}
}

func TestSaveToken_createsParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "token.json")

	if err := SaveToken(path, "tok"); err != nil {
		t.Fatalf("SaveToken: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("token file not created: %v", err)
	}
}

func TestLoadToken_missing_file(t *testing.T) {
	_, err := LoadToken(filepath.Join(t.TempDir(), "nonexistent.json"))
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired, got %v", err)
	}
}

func TestLoadToken_expired(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	rec := TokenRecord{
		AccessToken: "old-token",
		ExpiresAt:   time.Now().UTC().Add(-1 * time.Hour), // already expired
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err = LoadToken(path)
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired for expired token, got %v", err)
	}
}

func TestLoadToken_empty_access_token(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")

	rec := TokenRecord{
		AccessToken: "",
		ExpiresAt:   time.Now().UTC().Add(1 * time.Hour),
	}
	data, err := json.Marshal(rec)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if err = os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err = LoadToken(path)
	if !errors.Is(err, ErrAuthRequired) {
		t.Errorf("want ErrAuthRequired for empty token, got %v", err)
	}
}

func TestExchangeToken_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/session/token" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"access_token":"fresh-token-xyz"}}`))
	}))
	defer srv.Close()

	got, err := ExchangeToken(context.Background(), srv.Client(), srv.URL, "key", "reqtok", "secret")
	if err != nil {
		t.Fatalf("ExchangeToken: %v", err)
	}
	if got != "fresh-token-xyz" {
		t.Errorf("ExchangeToken: want %q, got %q", "fresh-token-xyz", got)
	}
}

func TestExchangeToken_http_error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := ExchangeToken(context.Background(), srv.Client(), srv.URL, "key", "reqtok", "secret")
	if err == nil {
		t.Fatal("want error, got nil")
	}
}

func TestExchangeToken_invalid_url(t *testing.T) {
	_, err := ExchangeToken(context.Background(), http.DefaultClient, "://bad-url", "key", "reqtok", "secret")
	if err == nil {
		t.Fatal("want error for invalid base URL, got nil")
	}
}

func TestExchangeToken_malformed_json(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	_, err := ExchangeToken(context.Background(), srv.Client(), srv.URL, "key", "reqtok", "secret")
	if err == nil {
		t.Fatal("want error for malformed JSON, got nil")
	}
}

func TestExchangeToken_empty_access_token_in_response(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"access_token":""}}`))
	}))
	defer srv.Close()

	_, err := ExchangeToken(context.Background(), srv.Client(), srv.URL, "key", "reqtok", "secret")
	if err == nil {
		t.Fatal("want error for empty access_token, got nil")
	}
}

func TestSaveToken_mkdir_failure(t *testing.T) {
	dir := t.TempDir()
	// Create a regular file where a directory needs to be so MkdirAll fails.
	blockingFile := filepath.Join(dir, "notadir")
	if err := os.WriteFile(blockingFile, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}
	err := SaveToken(filepath.Join(blockingFile, "token.json"), "tok")
	if err == nil {
		t.Fatal("want error when parent path is a file, got nil")
	}
}

func TestSaveToken_write_failure(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write to read-only directories")
	}
	dir := t.TempDir()
	readOnly := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(readOnly, 0o500); err != nil {
		t.Fatal(err)
	}
	err := SaveToken(filepath.Join(readOnly, "token.json"), "tok")
	if err == nil {
		t.Fatal("want error writing to read-only directory, got nil")
	}
}

func TestLoadToken_unreadable_file(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can read files with mode 0")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	if err := os.WriteFile(path, []byte(`{}`), 0o000); err != nil {
		t.Fatal(err)
	}
	_, err := LoadToken(path)
	if err == nil {
		t.Fatal("want error for unreadable file, got nil")
	}
	if errors.Is(err, ErrAuthRequired) {
		t.Errorf("want a read error, not ErrAuthRequired")
	}
}

func TestLoadToken_bad_json(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "token.json")
	if err := os.WriteFile(path, []byte(`not json`), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := LoadToken(path)
	if err == nil {
		t.Fatal("want error for bad JSON, got nil")
	}
}
