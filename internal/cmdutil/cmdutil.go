// Package cmdutil provides shared utilities for cmd/ entrypoints:
// .env loading, environment variable validation, token path resolution,
// and the interactive Kite Connect login flow.
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

// LoginFlow runs the interactive Kite Connect browser login flow and returns
// an access token. It prompts the user to open a URL and paste the
// request_token from the redirect. The token is saved to tokenPath on success;
// a save failure prints a warning but does not abort.
func LoginFlow(ctx context.Context, apiKey, apiSecret, tokenPath string) (string, error) {
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

	accessToken, err := zerodha.ExchangeToken(
		ctx, http.DefaultClient, "https://api.kite.trade",
		apiKey, requestToken, apiSecret,
	)
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
