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
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha/cache"
)

func main() {
	cmdutil.LoadDotEnv(".env")

	apiKey := cmdutil.MustEnv("KITE_API_KEY")
	apiSecret := cmdutil.MustEnv("KITE_API_SECRET")
	path := cmdutil.TokenFilePath()

	ctx := context.Background()

	// Try to load a saved token first.
	accessToken, err := zerodha.LoadToken(path)
	if err != nil {
		fmt.Println("No valid saved token found — starting login flow.")
		accessToken, err = cmdutil.LoginFlow(ctx, apiKey, apiSecret, path)
		if err != nil {
			cmdutil.Fatalf("login: %v", err)
		}
	} else {
		fmt.Printf("✓ Loaded saved token from %s\n", path)
	}

	fmt.Println("\nConstructing Provider (downloads instruments CSV)…")
	inner, err := zerodha.NewProvider(ctx, zerodha.Config{
		APIKey:      apiKey,
		AccessToken: accessToken,
	})
	if err != nil {
		cmdutil.Fatalf("NewProvider: %v", err)
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
		cmdutil.Fatalf("FetchCandles: %v", err)
	}
	elapsed1 := time.Since(t0)

	fmt.Printf("✓ %d candles in %v\n", len(candles), elapsed1.Round(time.Millisecond))
	printCandles(candles)

	fmt.Printf("\n── Fetch 2 (expect cache hit) ───────────────────────────────────────\n")
	t0 = time.Now()
	candles2, err := p.FetchCandles(ctx, instrument, model.TimeframeDaily, from, to)
	if err != nil {
		cmdutil.Fatalf("FetchCandles (cached): %v", err)
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
