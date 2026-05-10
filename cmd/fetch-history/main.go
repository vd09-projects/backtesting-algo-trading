// cmd/fetch-history is a bulk historical data fetch CLI.
//
// It reads a universe YAML file, iterates every instrument × timeframe combination,
// and calls FetchCandles for the range [--from, today). Results are written through
// the existing CachedProvider so the on-disk cache keys match what cmd/backtest and
// cmd/universe-sweep expect. Once the cache is populated, subsequent backtest runs
// work entirely from disk without a Zerodha token.
//
// # Usage
//
//	go run ./cmd/fetch-history \
//	    --universe universes/nifty50-large-cap.yaml \
//	    --timeframe 5min \
//	    --timeframe daily \
//	    --from 2015-01-01 \
//	    --cache-dir .cache/zerodha \
//	    --api-key $KITE_API_KEY \
//	    --access-token $KITE_ACCESS_TOKEN
//
// # Flags
//
//	--universe       Path to YAML universe file (required)
//	--timeframe      Candle timeframe, repeatable: 1min | 5min | 15min | daily | weekly (required, min 1)
//	--from           Start date YYYY-MM-DD (inclusive, required)
//	--cache-dir      Cache directory (default .cache/zerodha, same as other CLIs)
//	--api-key        Kite Connect API key (env: KITE_API_KEY)
//	--access-token   Kite Connect access token (env: KITE_ACCESS_TOKEN)
//	--dry-run        Print what would be fetched without hitting the API
//
// # Partial-failure recovery
//
// On each successful instrument×timeframe fetch, cmd/fetch-history writes a
// fetch-progress.json manifest in --cache-dir. On subsequent runs the completed
// pairs are skipped. Delete fetch-progress.json to force a full re-fetch from --from.
//
// # Incremental mode (pending TASK-0080)
//
// The incremental delta-fetch path (skip already-cached ranges via
// CachedProvider.LastCachedTime) is stubbed until TASK-0080 adds that API to
// CachedProvider. Until then every run fetches the full [--from, today) range
// (subject to progress manifest skip for completed pairs).
//
// # Auth
//
// --api-key and --access-token (or their env-var equivalents) are passed directly
// to zerodha.NewProvider. This differs from cmdutil.BuildProvider which runs an
// interactive OAuth login flow — fetch-history is a batch automation tool designed
// for scripted/CI use where a fresh token is injected externally.
//
// **Decision (fetch-history uses direct apiKey+accessToken, not OAuth login flow) — architecture: experimental**
// scope: cmd/fetch-history
// tags: auth, zerodha, access-token, batch-tooling, BuildProvider
// owner: priya
//
// cmdutil.BuildProvider is interactive (browser login). Batch fetch-history must
// accept a pre-obtained token. KITE_ACCESS_TOKEN env-var matches standard CI/CD
// token injection patterns. Users generating tokens via the login flow can set the
// env var from the token file before running fetch-history.
//
// **Decision (fetch-history progress manifest is cmd-layer concern, not CachedProvider) — architecture: experimental**
// scope: cmd/fetch-history
// tags: partial-failure, manifest, progress-tracking, cmd-layer
// owner: priya
//
// TASK-0080's CachedProvider manifest tracks per-instrument incremental timestamps
// for all callers. fetch-progress.json tracks which {instrument × timeframe} pairs
// in the current bulk run have been fully completed, enabling resume after failure.
// These are different concerns at different abstraction layers.
//
// **Decision (providerFactory receives parsed flags — no global state required) — convention: experimental**
// scope: cmd/fetch-history
// tags: testability, provider-factory, no-global-state, dry-run
// owner: priya
//
// run() accepts a providerFactory func(fetchFlags) (provider.DataProvider, error).
// The factory receives the fully-parsed fetchFlags so main() can close over nothing —
// it simply passes buildProductionProvider as the factory. Tests inject a factory
// that ignores flags and returns a mock. The factory is only called on non-dry-run
// paths; dry-run returns before the provider is constructed. This eliminates the
// prior productionMode package-level mutable var (CLAUDE.md: "No global state").
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/universesweep"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	zerodha "github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha/cache"
)

func main() {
	err := run(os.Args[1:], os.Stdout, os.Stderr, buildProductionProvider)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err) //nolint:errcheck // stderr write at process exit; non-fatal
		os.Exit(1)
	}
}

// manifestSkippedCandle mirrors zerodha.SkippedCandle for JSON serialization in the manifest.
//
// **Decision (SkippedCandle index and reason recorded in progress manifest skipped_candles field) — convention: experimental**
// scope: cmd/fetch-history.progressEntry.SkippedCandles
// tags: bad-candle, manifest, skip, OHLC-validation, TASK-0100
// owner: priya
//
// A local DTO mirrors zerodha.SkippedCandle rather than embedding the provider type in the
// manifest. The manifest is a cmd-layer concern; importing provider types into its JSON schema
// would couple the manifest format to the provider package. The fields are identical for now;
// if zerodha.SkippedCandle changes, the manifest format is unaffected.
type manifestSkippedCandle struct {
	Index  int    `json:"index"`
	Reason string `json:"reason"`
}

// progressEntry records a completed instrument+timeframe fetch.
// SkippedCandles is non-nil only when one or more candles were skipped due to
// OHLC validation failures (Zerodha tick-vs-aggregation artifact).
type progressEntry struct {
	Instrument     string                  `json:"instrument"`
	Timeframe      string                  `json:"timeframe"`
	SkippedCandles []manifestSkippedCandle `json:"skipped_candles,omitempty"`
}

// progressManifest is the on-disk partial-failure recovery state for a fetch-history run.
// It is written to {cache-dir}/fetch-progress.json after each successful fetch.
// Delete the file to force a full re-fetch.
type progressManifest struct {
	Completed []progressEntry `json:"completed"`
}

// progressKey returns a canonical key for deduplicating manifest entries.
func progressKey(instrument, tf string) string {
	return instrument + "|" + tf
}

// loadManifest reads a progress manifest from path.
// Returns an empty manifest if the file is absent; returns an error for corrupt files.
func loadManifest(path string) (progressManifest, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return progressManifest{}, nil
	}
	if err != nil {
		return progressManifest{}, fmt.Errorf("read progress manifest %q: %w", path, err)
	}
	var m progressManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return progressManifest{}, fmt.Errorf("parse progress manifest %q: %w", path, err)
	}
	return m, nil
}

// saveManifest writes a progress manifest atomically (write to .tmp, then rename).
func saveManifest(path string, m progressManifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write manifest tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp) //nolint:errcheck // best-effort cleanup of failed atomic write
		return fmt.Errorf("rename manifest: %w", err)
	}
	return nil
}

// timeframeFlag is a repeatable string flag.
type timeframeFlag []string

func (f *timeframeFlag) String() string {
	return fmt.Sprintf("%v", []string(*f))
}

func (f *timeframeFlag) Set(v string) error {
	*f = append(*f, v)
	return nil
}

// fetchFlags holds validated, parsed flag values.
type fetchFlags struct {
	universeFile string
	from         time.Time
	timeframes   []model.Timeframe
	cacheDir     string
	apiKey       string
	accessToken  string
	dryRun       bool
}

// parseFlags parses args into a fetchFlags, applying env-var fallbacks and validation.
func parseFlags(args []string, stderr io.Writer) (fetchFlags, error) {
	fs := flag.NewFlagSet("fetch-history", flag.ContinueOnError)
	fs.SetOutput(stderr)

	universeFile := fs.String("universe", "", "Path to YAML universe file (required)")
	fromStr := fs.String("from", "", "Start date YYYY-MM-DD (inclusive, required)")
	cacheDir := fs.String("cache-dir", ".cache/zerodha", "Cache directory (default .cache/zerodha)")
	apiKey := fs.String("api-key", "", "Kite Connect API key (env: KITE_API_KEY)")
	accessToken := fs.String("access-token", "", "Kite Connect access token (env: KITE_ACCESS_TOKEN)")
	dryRun := fs.Bool("dry-run", false, "Print what would be fetched without hitting the API")

	var rawTimeframes timeframeFlag
	fs.Var(&rawTimeframes, "timeframe", "Candle timeframe, repeatable: 1min | 5min | 15min | daily | weekly (required)")

	if err := fs.Parse(args); err != nil {
		return fetchFlags{}, err
	}

	// Apply env-var fallbacks.
	if *apiKey == "" {
		*apiKey = os.Getenv("KITE_API_KEY")
	}
	if *accessToken == "" {
		*accessToken = os.Getenv("KITE_ACCESS_TOKEN")
	}

	if *universeFile == "" {
		return fetchFlags{}, fmt.Errorf("--universe is required (e.g. universes/nifty50-large-cap.yaml)")
	}
	if *fromStr == "" {
		return fetchFlags{}, fmt.Errorf("--from is required (e.g. 2015-01-01)")
	}
	if len(rawTimeframes) == 0 {
		return fetchFlags{}, fmt.Errorf("--timeframe is required (e.g. --timeframe 5min --timeframe daily)")
	}

	from, err := time.Parse("2006-01-02", *fromStr)
	if err != nil {
		return fetchFlags{}, fmt.Errorf("--from %q: %w", *fromStr, err)
	}

	tfs := make([]model.Timeframe, 0, len(rawTimeframes))
	for _, tfStr := range rawTimeframes {
		tf := model.Timeframe(tfStr)
		switch tf {
		case model.Timeframe1Min, model.Timeframe5Min, model.Timeframe15Min,
			model.TimeframeDaily, model.TimeframeWeekly:
			tfs = append(tfs, tf)
		default:
			return fetchFlags{}, fmt.Errorf("--timeframe %q is not valid; choose one of: 1min, 5min, 15min, daily, weekly", tfStr)
		}
	}

	return fetchFlags{
		universeFile: *universeFile,
		from:         from,
		timeframes:   tfs,
		cacheDir:     *cacheDir,
		apiKey:       *apiKey,
		accessToken:  *accessToken,
		dryRun:       *dryRun,
	}, nil
}

// run is the testable entry point.
//
// providerFactory receives the fully-parsed fetchFlags and returns a DataProvider.
// main() passes buildProductionProvider directly. Tests inject a factory that
// ignores flags and returns a mock. The factory is only called on non-dry-run paths.
//
// **Decision (run() extraction for testability in cmd/fetch-history) — convention: experimental**
// scope: cmd/fetch-history
// tags: testability, flag-parse, coverage, run-function, TASK-0070
// owner: priya
//
// Follows the convention established in cmd/walk-forward and cmd/monitor.
func run(args []string, stdout, stderr io.Writer, providerFactory func(fetchFlags) (provider.DataProvider, error)) error {
	flags, err := parseFlags(args, stderr)
	if err != nil {
		return err
	}

	instruments, err := universesweep.ParseUniverseFile(flags.universeFile)
	if err != nil {
		return fmt.Errorf("universe file: %w", err)
	}

	if flags.dryRun {
		return printDryRunPlan(stdout, instruments, flags.timeframes, flags.from)
	}

	if flags.apiKey == "" {
		return fmt.Errorf("--api-key is required (or set KITE_API_KEY)")
	}

	p, err := providerFactory(flags)
	if err != nil {
		return fmt.Errorf("build provider: %w", err)
	}

	return fetchAll(stdout, stderr, p, instruments, flags.timeframes, flags.from, flags.cacheDir)
}

// printDryRunPlan writes the fetch plan to stdout without hitting the API.
func printDryRunPlan(stdout io.Writer, instruments []string, tfs []model.Timeframe, from time.Time) error {
	fmt.Fprintf(stdout, "[dry-run] Would fetch %d instruments × %d timeframes from %s to today:\n", //nolint:errcheck // progress output; non-fatal
		len(instruments), len(tfs), from.Format("2006-01-02"))
	for _, inst := range instruments {
		for _, tf := range tfs {
			fmt.Fprintf(stdout, "  [dry-run] %s × %s: %s → today\n", inst, tf, from.Format("2006-01-02")) //nolint:errcheck // progress output; non-fatal
		}
	}
	return nil
}

// fetchAll runs the instrument × timeframe fetch loop, updating the progress manifest on each success.
func fetchAll(stdout, stderr io.Writer, p provider.DataProvider, instruments []string, tfs []model.Timeframe, from time.Time, cacheDir string) error {
	manifestPath := filepath.Join(cacheDir, "fetch-progress.json")
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		fmt.Fprintf(stderr, "warning: could not read progress manifest (%v); starting from beginning\n", err) //nolint:errcheck // warning; non-fatal
		manifest = progressManifest{}
	}

	completed := buildCompletedSet(manifest)
	today := time.Now().UTC().Truncate(24 * time.Hour)
	ctx := context.Background()
	var fetchErr error

	for _, inst := range instruments {
		for _, tf := range tfs {
			if err := fetchOne(ctx, stdout, stderr, p, inst, tf, from, today, manifestPath, &manifest, completed); err != nil {
				fetchErr = err
			}
		}
	}

	if fetchErr != nil {
		return fmt.Errorf("one or more fetches failed; progress saved to %s", manifestPath)
	}
	return nil
}

// buildCompletedSet converts manifest entries to a set for O(1) skip lookups.
func buildCompletedSet(manifest progressManifest) map[string]bool {
	completed := make(map[string]bool, len(manifest.Completed))
	for _, e := range manifest.Completed {
		completed[progressKey(e.Instrument, e.Timeframe)] = true
	}
	return completed
}

// fetchOne fetches a single instrument×timeframe pair and updates the manifest.
// Returns nil on success, or the fetch error for this pair on failure.
// The caller (fetchAll) is responsible for accumulating errors across multiple pairs.
func fetchOne(
	ctx context.Context,
	stdout, stderr io.Writer,
	p provider.DataProvider,
	inst string,
	tf model.Timeframe,
	from, today time.Time,
	manifestPath string,
	manifest *progressManifest,
	completed map[string]bool,
) error {
	key := progressKey(inst, string(tf))
	if completed[key] {
		fmt.Fprintf(stdout, "%s × %s: already completed (skipping)\n", inst, tf) //nolint:errcheck // progress output; non-fatal
		return nil
	}

	// Incremental stub: check LastCachedTime when TASK-0080 adds it.
	// TODO(TASK-0080): if t, ok := p.(interface{ LastCachedTime(string, model.Timeframe) (time.Time, bool) }); ok { ... }
	effectiveFrom := from

	candles, fetchErr := p.FetchCandles(ctx, inst, tf, effectiveFrom, today)

	// Detect the non-fatal ErrBadCandles warning: some candles were skipped due to OHLC
	// validation failures (Zerodha tick-vs-aggregation artifact). When at least one valid
	// candle was returned, log per-candle warnings, record in the manifest, and treat the
	// fetch as successful. When ALL candles were bad (len(candles) == 0), log an error and
	// treat as failed — nothing useful was cached and a retry may succeed if the artifact
	// is intermittent.
	var badCandles *zerodha.ErrBadCandles
	if errors.As(fetchErr, &badCandles) {
		// Log a warning for each skipped candle regardless of outcome.
		for _, sc := range badCandles.Skipped {
			fmt.Fprintf(stderr, "warning: %s × %s: skipped candle[%d]: %s\n", //nolint:errcheck // warning; non-fatal
				inst, tf, sc.Index, sc.Reason)
		}

		if len(candles) == 0 {
			// All candles were bad — nothing cached, treat as failure so the instrument
			// is retried on the next run.
			fmt.Fprintf(stderr, "%s × %s: all %d candles were invalid, nothing cached — instrument will retry on next run\n", //nolint:errcheck // error log; non-fatal
				inst, tf, len(badCandles.Skipped))
			if saveErr := saveManifest(manifestPath, *manifest); saveErr != nil {
				fmt.Fprintf(stderr, "warning: could not save progress manifest: %v\n", saveErr) //nolint:errcheck // warning; non-fatal
			}
			return badCandles
		}

		// At least one valid candle — treat as success with skipped-candle annotation.
		manifestSkipped := make([]manifestSkippedCandle, len(badCandles.Skipped))
		for i, sc := range badCandles.Skipped {
			manifestSkipped[i] = manifestSkippedCandle{Index: sc.Index, Reason: sc.Reason}
		}
		fmt.Fprintf(stdout, "%s × %s: fetched %d candles [%s → %s] (%d skipped)\n", //nolint:errcheck // progress output; non-fatal
			inst, tf, len(candles),
			effectiveFrom.Format("2006-01-02"),
			today.Format("2006-01-02"),
			len(badCandles.Skipped),
		)
		manifest.Completed = append(manifest.Completed, progressEntry{
			Instrument:     inst,
			Timeframe:      string(tf),
			SkippedCandles: manifestSkipped,
		})
		completed[key] = true
		if saveErr := saveManifest(manifestPath, *manifest); saveErr != nil {
			fmt.Fprintf(stderr, "warning: could not save progress manifest: %v\n", saveErr) //nolint:errcheck // warning; non-fatal
		}
		return nil
	}

	if fetchErr != nil {
		fmt.Fprintf(stderr, "%s × %s: fetch error: %v\n", inst, tf, fetchErr) //nolint:errcheck // error log; non-fatal
		if saveErr := saveManifest(manifestPath, *manifest); saveErr != nil {
			fmt.Fprintf(stderr, "warning: could not save progress manifest: %v\n", saveErr) //nolint:errcheck // warning; non-fatal
		}
		return fetchErr
	}

	fmt.Fprintf(stdout, "%s × %s: fetched %d candles [%s → %s]\n", //nolint:errcheck // progress output; non-fatal
		inst, tf, len(candles),
		effectiveFrom.Format("2006-01-02"),
		today.Format("2006-01-02"),
	)

	manifest.Completed = append(manifest.Completed, progressEntry{
		Instrument: inst,
		Timeframe:  string(tf),
	})
	completed[key] = true

	if saveErr := saveManifest(manifestPath, *manifest); saveErr != nil {
		fmt.Fprintf(stderr, "warning: could not save progress manifest: %v\n", saveErr) //nolint:errcheck // warning; non-fatal
	}
	return nil
}

// resolveBatchToken resolves an access token for non-interactive (batch/CI) use.
// Priority: explicit flagToken → saved token file at tokenFilePath → error.
// Any error from the token file (absent, expired, corrupt) falls through to the
// descriptive error. No interactive login flow is triggered.
//
// **Decision (resolveBatchToken in fetch-history to name non-interactive token resolution) — convention: experimental**
// scope: cmd/fetch-history
// tags: auth, token, batch, ci, non-interactive
// owner: priya
//
// BuildProvider in cmdutil may trigger an interactive browser login as a last resort.
// resolveBatchToken is limited to non-interactive sources — safe for CI, scripts,
// and scheduled runs where stdin is absent. The name makes this explicit.
func resolveBatchToken(flagToken, tokenFilePath string) (string, error) {
	if flagToken != "" {
		return flagToken, nil
	}
	tok, err := zerodha.LoadToken(tokenFilePath)
	if err != nil {
		return "", fmt.Errorf(
			"no access token: set --access-token flag, KITE_ACCESS_TOKEN env var, or run cmd/backtest to generate a saved token at %s",
			tokenFilePath,
		)
	}
	return tok, nil
}

// buildProductionProvider constructs the real zerodha.Provider wrapped in CachedProvider.
// Passed as the providerFactory to run() by main(). Tests inject a mock factory instead.
//
// Token resolution priority: --access-token flag → KITE_ACCESS_TOKEN env var (applied by
// parseFlags before this is called) → saved token file at cmdutil.TokenFilePath().
// No interactive login flow — use cmd/backtest to generate a saved token if needed.
func buildProductionProvider(flags fetchFlags) (provider.DataProvider, error) { //nolint:gocritic // hugeParam: fetchFlags is a value at the cmd/main API boundary; pointer semantics not justified for a once-per-process call
	accessToken, err := resolveBatchToken(flags.accessToken, cmdutil.TokenFilePath())
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	p, err := zerodha.NewProvider(ctx, zerodha.Config{
		APIKey:              flags.apiKey,
		AccessToken:         accessToken,
		InstrumentsCacheDir: flags.cacheDir,
	})
	if err != nil {
		return nil, fmt.Errorf("zerodha.NewProvider: %w", err)
	}
	return cache.NewCachedProvider(p, flags.cacheDir), nil
}
