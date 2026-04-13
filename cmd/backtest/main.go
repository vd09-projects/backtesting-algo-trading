// cmd/backtest is the CLI entrypoint for the backtesting engine.
//
// Usage:
//
//	go run ./cmd/backtest \
//	    --instrument "NSE:NIFTY 50" \
//	    --from 2024-01-01 \
//	    --to   2024-12-31 \
//	    --timeframe daily \
//	    --cash 100000 \
//	    --strategy stub \
//	    --out results.json
//
// Credentials are read from KITE_API_KEY and KITE_API_SECRET environment
// variables (or a .env file in the working directory). A saved access token is
// reused when present at ~/.config/backtest/token.json and not expired;
// otherwise the Kite Connect login flow is triggered.
//
// Optional overrides:
//
//	BACKTEST_TOKEN_PATH — override the token file location
//	BACKTEST_CACHE_DIR  — override the candle cache directory (default: .cache/zerodha)
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/output"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha/cache"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
	stubstrategy "github.com/vikrantdhawan/backtesting-algo-trading/strategies/stub"
)

func main() {
	instrument := flag.String("instrument", "NSE:NIFTY 50", "Instrument to backtest (e.g. \"NSE:NIFTY 50\", \"NSE:INFY\")")
	fromStr := flag.String("from", "", "Start date in YYYY-MM-DD (inclusive)")
	toStr := flag.String("to", "", "End date in YYYY-MM-DD (exclusive)")
	tfStr := flag.String("timeframe", "daily", "Candle timeframe: 1min | 5min | 15min | daily | weekly")
	cash := flag.Float64("cash", 100000, "Starting cash in ₹")
	stratName := flag.String("strategy", "stub", "Strategy name: stub, sma-crossover")
	fastPeriod := flag.Int("fast-period", 10, "SMA crossover: fast period (default 10)")
	slowPeriod := flag.Int("slow-period", 50, "SMA crossover: slow period (default 50)")
	outPath := flag.String("out", "", "Path for JSON results export (omit to skip)")
	flag.Parse()

	// Validate required flags.
	if *fromStr == "" {
		cmdutil.Fatalf("--from is required (e.g. 2024-01-01)")
	}
	if *toStr == "" {
		cmdutil.Fatalf("--to is required (e.g. 2024-12-31)")
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
	case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
		model.TimeframeDaily, model.TimeframeWeekly:
	default:
		cmdutil.Fatalf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", *tfStr)
	}

	selectedStrategy, err := strategyRegistry(*stratName, tf, *fastPeriod, *slowPeriod)
	if err != nil {
		cmdutil.Fatalf("--strategy: %v", err)
	}

	ctx := context.Background()

	cmdutil.LoadDotEnv(".env")

	p, err := buildProvider(ctx)
	if err != nil {
		cmdutil.Fatalf("provider: %v", err)
	}

	eng := engine.New(engine.Config{
		Instrument:           *instrument,
		From:                 from,
		To:                   to,
		InitialCash:          *cash,
		PositionSizeFraction: 0.1,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: model.CommissionZerodha,
		},
	})

	fmt.Printf("Running strategy %q on %s  %s → %s\n",
		selectedStrategy.Name(), *instrument,
		from.Format("2006-01-02"), to.Format("2006-01-02"))

	if err := eng.Run(ctx, p, selectedStrategy); err != nil {
		cmdutil.Fatalf("engine: %v", err)
	}

	port := eng.Portfolio()
	report := analytics.Compute(port.ClosedTrades(), port.EquityCurve(), tf)

	if err := output.Write(report, output.Config{
		PrintToStdout: true,
		FilePath:      *outPath,
	}); err != nil {
		cmdutil.Fatalf("output: %v", err)
	}
}

func strategyRegistry(name string, tf model.Timeframe, fastPeriod, slowPeriod int) (strategy.Strategy, error) {
	switch name {
	case "stub":
		return stubstrategy.New(tf), nil
	case "sma-crossover":
		return smacrossover.New(tf, fastPeriod, slowPeriod)
	default:
		return nil, fmt.Errorf("unknown strategy %q; available: stub, sma-crossover", name)
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
