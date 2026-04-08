// cmd/providertest exercises the real Provider and CachedProvider against the live Kite Connect API.
// It is NOT production code — manual smoke test for TASK-0008 and TASK-0009 verification.
//
// Usage:
//
//	go run ./cmd/providertest
//
// Reads KITE_API_KEY and KITE_API_SECRET from .env (same as cmd/authtest).
// If a saved token exists at ~/.config/backtest/token.json and is not expired,
// it is reused. Otherwise the full login flow is triggered.
//
// Cache files are written to .cache/zerodha/ (override with BACKTEST_CACHE_DIR).
package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha/cache"
)

func tokenFilePath() string {
	if p := os.Getenv("BACKTEST_TOKEN_PATH"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fatalf("UserHomeDir: %v", err)
	}
	return home + "/.config/backtest/token.json"
}

func main() {
	loadDotEnv(".env")

	apiKey := mustEnv("KITE_API_KEY")
	apiSecret := mustEnv("KITE_API_SECRET")
	path := tokenFilePath()

	// Try to load a saved token first.
	accessToken, err := zerodha.LoadToken(path)
	if err != nil {
		fmt.Println("No valid saved token found — starting login flow.")
		accessToken = loginFlow(apiKey, apiSecret, path)
	} else {
		fmt.Printf("✓ Loaded saved token from %s\n", path)
	}

	ctx := context.Background()

	fmt.Println("\nConstructing Provider (downloads instruments CSV)…")
	inner, err := zerodha.NewProvider(ctx, zerodha.Config{
		APIKey:      apiKey,
		AccessToken: accessToken,
	})
	if err != nil {
		fatalf("NewProvider: %v", err)
	}
	fmt.Printf("✓ Provider ready. Supported timeframes: %v\n", inner.SupportedTimeframes())

	cacheDir := os.Getenv("BACKTEST_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = ".cache/zerodha"
	}
	p := cache.NewCachedProvider(inner, cacheDir)
	fmt.Printf("✓ CachedProvider wrapping inner provider (cache dir: %s)\n", cacheDir)

	// Fetch last 10 days of NIFTY 50 daily candles — twice, to show miss then hit.
	instrument := "NSE:NIFTY 50"
	to := time.Now()
	from := to.AddDate(0, 0, -10)

	fmt.Printf("\n── Fetch 1 (expect cache miss) ──────────────────────────────────────\n")
	fmt.Printf("Fetching %s daily candles %s → %s…\n",
		instrument, from.Format("2006-01-02"), to.Format("2006-01-02"))

	t0 := time.Now()
	candles, err := p.FetchCandles(ctx, instrument, model.TimeframeDaily, from, to)
	if err != nil {
		fatalf("FetchCandles: %v", err)
	}
	elapsed1 := time.Since(t0)

	fmt.Printf("✓ %d candles in %v\n", len(candles), elapsed1.Round(time.Millisecond))
	printCandles(candles)

	fmt.Printf("\n── Fetch 2 (expect cache hit) ───────────────────────────────────────\n")
	t0 = time.Now()
	candles2, err := p.FetchCandles(ctx, instrument, model.TimeframeDaily, from, to)
	if err != nil {
		fatalf("FetchCandles (cached): %v", err)
	}
	elapsed2 := time.Since(t0)

	fmt.Printf("✓ %d candles in %v", len(candles2), elapsed2.Round(time.Millisecond))
	if elapsed2 < elapsed1/2 {
		fmt.Printf("  ← cache hit (%.0fx faster)\n", float64(elapsed1)/float64(elapsed2))
	} else {
		fmt.Printf("  (cache may not have been written — check %s)\n", cacheDir)
	}
}

func printCandles(candles []model.Candle) {
	fmt.Printf("  %-28s  %8s  %8s  %8s  %8s  %10s\n", "Timestamp", "Open", "High", "Low", "Close", "Volume")
	fmt.Printf("  %s\n", strings.Repeat("─", 78))
	for _, c := range candles {
		fmt.Printf("  %-28s  %8.2f  %8.2f  %8.2f  %8.2f  %10.0f\n",
			c.Timestamp.Format(time.RFC3339), c.Open, c.High, c.Low, c.Close, c.Volume)
	}
}

func loginFlow(apiKey, apiSecret, path string) string {
	fmt.Printf("\nOpen this URL in your browser:\n\n  %s\n\n", zerodha.LoginURL(apiKey))
	fmt.Println("After login, copy the request_token from the redirect URL.")

	fmt.Print("Paste request_token: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	requestToken := strings.TrimSpace(scanner.Text())
	if requestToken == "" {
		fatalf("request_token cannot be empty")
	}

	accessToken, err := zerodha.ExchangeToken(
		context.Background(), http.DefaultClient, "https://api.kite.trade",
		apiKey, requestToken, apiSecret,
	)
	if err != nil {
		fatalf("ExchangeToken: %v", err)
	}
	fmt.Printf("✓ access_token obtained\n")

	if err := zerodha.SaveToken(path, accessToken); err != nil {
		fmt.Printf("⚠  SaveToken: %v (continuing without saving)\n", err)
	} else {
		fmt.Printf("✓ Token saved to %s\n", path)
	}
	return accessToken
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fatalf("environment variable %s is not set", key)
	}
	return v
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close() //nolint:errcheck // close on a read-only file is non-actionable
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
		if os.Getenv(strings.TrimSpace(key)) == "" {
			os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value)) //nolint:errcheck // key is non-empty and contains no '=' (already split on it)
		}
	}
}
