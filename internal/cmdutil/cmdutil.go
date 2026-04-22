// Package cmdutil provides shared utilities for cmd/ entrypoints:
// .env loading, environment variable validation, token path resolution,
// the interactive Kite Connect login flow, and the shared BuildProvider
// constructor used by cmd/backtest, cmd/sweep, and cmd/universe-sweep.
package cmdutil

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha/cache"
)

// LoadDotEnv reads key=value pairs from path and sets them as environment
// variables. Blank lines and lines starting with # are skipped. Real
// environment variables already set take precedence and are never overwritten.
// A missing file is silently ignored — .env is optional.
func LoadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close() //nolint:errcheck // read-only file; close error is non-fatal
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if os.Getenv(key) == "" {
			os.Setenv(key, value) //nolint:errcheck // best-effort; key is non-empty
		}
	}
	// .env loading is best-effort — treat a scan error the same as a missing file.
	_ = scanner.Err() //nolint:errcheck // explicitly acknowledged; no action taken
}

// Fatalf prints a formatted error message to stderr and exits with code 1.
func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

// MustEnv returns the value of the named environment variable.
// It calls Fatalf if the variable is unset or empty.
func MustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		Fatalf("environment variable %s is not set", key)
	}
	return v
}

// TokenFilePath returns the path to the saved Kite Connect access token.
// BACKTEST_TOKEN_PATH overrides the default ~/.config/backtest/token.json.
func TokenFilePath() string {
	if p := os.Getenv("BACKTEST_TOKEN_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		Fatalf("UserHomeDir: %v", err)
	}
	return filepath.Join(home, ".config", "backtest", "token.json")
}

// BuildProvider constructs the Zerodha cached provider used by all cmd/ entrypoints.
// It loads credentials from the environment, loads or exchanges an access token,
// and wraps the result in a disk cache. The cache directory is read from
// BACKTEST_CACHE_DIR (default: .cache/zerodha).
//
// **Decision (BuildProvider extracted to cmdutil) — architecture: experimental**
// scope: internal/cmdutil, cmd/backtest, cmd/sweep, cmd/universe-sweep
// tags: provider, DRY, cmd, zerodha
// owner: priya
//
// cmd/backtest and cmd/sweep each had an identical private buildProvider function.
// cmd/universe-sweep would have been a third copy. Extracting to cmdutil.BuildProvider
// eliminates the duplication. The function is a pure I/O constructor with no
// business logic, so it belongs in cmdutil alongside MustEnv, TokenFilePath,
// and LoginFlow — the other shared cmd-layer plumbing.
func BuildProvider(ctx context.Context) (*cache.CachedProvider, error) {
	apiKey := MustEnv("KITE_API_KEY")
	apiSecret := MustEnv("KITE_API_SECRET")

	path := TokenFilePath()
	accessToken, err := zerodha.LoadToken(path)
	if err != nil {
		fmt.Println("No valid saved token — starting Kite Connect login flow.")
		accessToken, err = LoginFlow(ctx, http.DefaultClient, "https://api.kite.trade", apiKey, apiSecret, path)
		if err != nil {
			return nil, fmt.Errorf("login: %w", err)
		}
	} else {
		fmt.Printf("Loaded saved token from %s\n", path)
	}

	inner, err := zerodha.NewProvider(ctx, zerodha.Config{
		APIKey:      apiKey,
		AccessToken: accessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("NewProvider: %w", err)
	}

	cacheDir := os.Getenv("BACKTEST_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = ".cache/zerodha"
	}
	return cache.NewCachedProvider(inner, cacheDir), nil
}

// LoginFlow runs the interactive Kite Connect browser login flow and returns
// an access token. It prompts the user to open a URL and paste the
// request_token from the redirect. The token is saved to tokenPath on success;
// a save failure prints a warning but does not abort.
// client and baseURL are injected so callers can substitute a test server.
func LoginFlow(ctx context.Context, client *http.Client, baseURL, apiKey, apiSecret, tokenPath string) (string, error) {
	fmt.Printf("\nOpen this URL in your browser:\n\n  %s\n\n", zerodha.LoginURL(apiKey))
	fmt.Println("After login, copy the request_token from the redirect URL and paste it here.")

	fmt.Print("request_token: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("read request_token: %w", err)
		}
		return "", fmt.Errorf("request_token: stdin closed unexpectedly")
	}
	requestToken := strings.TrimSpace(scanner.Text())
	if requestToken == "" {
		return "", fmt.Errorf("request_token cannot be empty")
	}

	accessToken, err := zerodha.ExchangeToken(ctx, client, baseURL, apiKey, requestToken, apiSecret)
	if err != nil {
		return "", fmt.Errorf("ExchangeToken: %w", err)
	}

	if err := zerodha.SaveToken(tokenPath, accessToken); err != nil {
		fmt.Printf("warning: could not save token to %s: %v\n", tokenPath, err)
	} else {
		fmt.Printf("Token saved to %s\n", tokenPath)
	}
	return accessToken, nil
}
