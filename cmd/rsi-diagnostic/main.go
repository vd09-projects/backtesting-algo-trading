// cmd/rsi-diagnostic counts how many bars RSI(period) spends below
// the oversold threshold and above the overbought threshold on a given
// instrument and date range, without requiring any closed trades.
//
// This is the pre-condition diagnostic for TASK-0031: if the total
// signal bar count is < 30, the fixed thresholds are miscalibrated
// for the instrument; if >= 30, investigate whether something in the
// engine is suppressing entries (e.g. vol-targeting zeroing position
// size during high-volatility bars).
//
// Usage:
//
//	go run ./cmd/rsi-diagnostic \
//	    --instrument "NSE:RELIANCE" \
//	    --from 2018-01-01 \
//	    --to   2025-01-01 \
//	    --timeframe daily \
//	    --rsi-period 14 \
//	    --oversold 30 \
//	    --overbought 70
package main

import (
	"context"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	talib "github.com/markcheno/go-talib"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha/cache"
)

func main() {
	instrument := flag.String("instrument", "NSE:RELIANCE", "Instrument (e.g. \"NSE:RELIANCE\")")
	fromStr := flag.String("from", "", "Start date YYYY-MM-DD (inclusive)")
	toStr := flag.String("to", "", "End date YYYY-MM-DD (exclusive)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily")
	rsiPeriod := flag.Int("rsi-period", 14, "RSI period")
	oversold := flag.Float64("oversold", 30.0, "Oversold threshold (signal when RSI < this)")
	overbought := flag.Float64("overbought", 70.0, "Overbought threshold (signal when RSI > this)")
	flag.Parse()

	if *fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2018-01-01)")
	}
	if *toStr == "" {
		cmdutil.Fatalf("--to is required (e.g. 2025-01-01)")
	}

	from, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		cmdutil.Fatalf("--from %q: %v", *fromStr, err)
	}
	to, err := time.Parse("2006-01-02", *toStr)
	if err != nil {
		cmdutil.Fatalf("--to %q: %v", *toStr, err)
	}
	if !to.After(from) {
		cmdutil.Fatalf("--to must be strictly after --from")
	}

	tf := model.Timeframe(*tfStr)
	switch tf {
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min, model.TimeframeDaily:
	default:
		cmdutil.Fatalf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily", *tfStr)
	}

	ctx := context.Background()
	cmdutil.LoadDotEnv(".env")

	p, err := buildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	fmt.Printf("Fetching %s %s candles %s → %s ...\n",
		*instrument, tf, from.Format("2006-01-02"), to.Format("2006-01-02"))

	candles, err := p.FetchCandles(ctx, *instrument, tf, from, to)
	if err != nil {
		cmdutil.Fatalf("FetchCandles: %v", err)
	}
	if len(candles) == 0 {
		cmdutil.Fatalf("no candles returned for %s", *instrument)
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	rsiVals := talib.Rsi(closes, *rsiPeriod)

	var oversoldBars, overboughtBars int
	// talib.Rsi fills the first rsiPeriod entries with NaN; start after them.
	for i := *rsiPeriod; i < len(rsiVals); i++ {
		v := rsiVals[i]
		if math.IsNaN(v) {
			continue
		}
		if v < *oversold {
			oversoldBars++
		} else if v > *overbought {
			overboughtBars++
		}
	}
	totalSignals := oversoldBars + overboughtBars
	validBars := len(rsiVals) - *rsiPeriod

	fmt.Println()
	fmt.Println("RSI Signal Frequency Diagnostic")
	fmt.Println("================================")
	fmt.Printf("Instrument:             %s\n", *instrument)
	fmt.Printf("Period:                 %s → %s\n", from.Format("2006-01-02"), to.Format("2006-01-02"))
	fmt.Printf("Timeframe:              %s\n", tf)
	fmt.Printf("Total bars fetched:     %d\n", len(candles))
	fmt.Printf("Valid RSI bars:         %d  (after %d-bar lookback)\n", validBars, *rsiPeriod)
	fmt.Println()
	fmt.Printf("RSI(%d) thresholds:     oversold < %.1f  |  overbought > %.1f\n",
		*rsiPeriod, *oversold, *overbought)
	fmt.Printf("Oversold bars  (buy):   %d\n", oversoldBars)
	fmt.Printf("Overbought bars (sell): %d\n", overboughtBars)
	fmt.Printf("Total signal bars:      %d\n", totalSignals)
	fmt.Println()

	const minSignals = 30
	if totalSignals < minSignals {
		fmt.Printf("VERDICT: MISCALIBRATED — only %d signal bars across %d valid bars.\n",
			totalSignals, validBars)
		fmt.Printf("Fixed thresholds (%.0f/%.0f) are rarely breached by %s on %s data.\n",
			*oversold, *overbought, *instrument, tf)
		fmt.Println("Next step: record decision and propose adaptive thresholds before any re-test.")
	} else {
		fmt.Printf("VERDICT: SIGNALS PRESENT — %d signal bars found (>= %d minimum).\n",
			totalSignals, minSignals)
		fmt.Println("Thresholds are breached sufficiently. Investigate entry suppression")
		fmt.Println("(e.g. vol-targeting zeroing position size during high-volatility events).")
	}
}

func buildProvider(ctx context.Context) (*cache.CachedProvider, error) {
	apiKey := cmdutil.MustEnv("KITE_API_KEY")
	apiSecret := cmdutil.MustEnv("KITE_API_SECRET")

	path := cmdutil.TokenFilePath()
	accessToken, err := zerodha.LoadToken(path)
	if err != nil {
		fmt.Println("No valid saved token — starting Kite Connect login flow.")
		accessToken, err = cmdutil.LoginFlow(ctx, http.DefaultClient, "https://api.kite.trade", apiKey, apiSecret, path)
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
