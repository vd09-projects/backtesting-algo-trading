package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
)

// writeValidTokenFile writes a token fixture that will pass zerodha.LoadToken validation.
// Uses the known TokenRecord JSON shape: access_token + expires_at (future).
func writeValidTokenFile(t *testing.T, path, token string) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"access_token": token,
		"expires_at":   time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("marshal token fixture: %v", err)
	}
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write token fixture: %v", err)
	}
}

// mockProvider is a test double for provider.DataProvider.
type mockProvider struct {
	candles   []model.Candle
	err       error
	callCount int
	calls     []fetchCall
}

type fetchCall struct {
	instrument string
	tf         model.Timeframe
	from, to   time.Time
}

func (m *mockProvider) FetchCandles(_ context.Context, instrument string, tf model.Timeframe, from, to time.Time) ([]model.Candle, error) {
	m.callCount++
	m.calls = append(m.calls, fetchCall{instrument: instrument, tf: tf, from: from, to: to})
	return m.candles, m.err
}

func (m *mockProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.Timeframe5Min, model.TimeframeDaily}
}

// compile-time check that mockProvider satisfies DataProvider.
var _ provider.DataProvider = (*mockProvider)(nil)

// writeUniverseYAML writes a minimal universe YAML file and returns its path.
func writeUniverseYAML(t *testing.T, dir string, instruments []string) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("instruments:\n")
	for _, inst := range instruments {
		sb.WriteString("  - ")
		sb.WriteString(inst)
		sb.WriteString("\n")
	}
	path := filepath.Join(dir, "universe.yaml")
	if err := os.WriteFile(path, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("writeUniverseYAML: %v", err)
	}
	return path
}

// mockFactory returns a providerFactory that always returns mock.
func mockFactory(mock *mockProvider) func(fetchFlags) (provider.DataProvider, error) {
	return func(_ fetchFlags) (provider.DataProvider, error) {
		return mock, nil
	}
}

// panicFactory is a providerFactory that panics if called — used to verify dry-run never constructs a provider.
func panicFactory() func(fetchFlags) (provider.DataProvider, error) {
	return func(_ fetchFlags) (provider.DataProvider, error) {
		panic("providerFactory must not be called during dry-run")
	}
}

// TestRun_MissingUniverseFlag verifies that --universe is required.
func TestRun_MissingUniverseFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := run([]string{"--from", "2024-01-01", "--timeframe", "5min"}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --universe, got nil")
	}
	if !strings.Contains(err.Error(), "--universe") {
		t.Errorf("error should mention --universe, got: %v", err)
	}
}

// TestRun_MissingFromFlag verifies that --from is required.
func TestRun_MissingFromFlag(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})
	var stdout, stderr bytes.Buffer
	err := run([]string{"--universe", uPath, "--timeframe", "5min"}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --from, got nil")
	}
	if !strings.Contains(err.Error(), "--from") {
		t.Errorf("error should mention --from, got: %v", err)
	}
}

// TestRun_MissingTimeframeFlag verifies that at least one --timeframe is required.
func TestRun_MissingTimeframeFlag(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})
	var stdout, stderr bytes.Buffer
	err := run([]string{"--universe", uPath, "--from", "2024-01-01"}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for missing --timeframe, got nil")
	}
	if !strings.Contains(err.Error(), "--timeframe") {
		t.Errorf("error should mention --timeframe, got: %v", err)
	}
}

// TestRun_InvalidFromDate verifies that an unparseable --from date returns an error.
func TestRun_InvalidFromDate(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})
	var stdout, stderr bytes.Buffer
	err := run([]string{"--universe", uPath, "--from", "not-a-date", "--timeframe", "5min"}, &stdout, &stderr, panicFactory())
	if err == nil {
		t.Fatal("expected error for invalid date, got nil")
	}
}

// TestRun_DryRun_PrintsWhatWouldBeFetched verifies dry-run output lists instruments × timeframes
// without calling the provider factory.
func TestRun_DryRun_PrintsWhatWouldBeFetched(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE", "NSE:TCS"})

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--timeframe", "daily",
		"--cache-dir", dir,
		"--dry-run",
	}, &stdout, &stderr, panicFactory()) // panicFactory: must not be called
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	// Must mention both instruments and both timeframes.
	for _, want := range []string{"NSE:RELIANCE", "NSE:TCS", "5min", "daily"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry-run output missing %q; got:\n%s", want, out)
		}
	}
	// Must mention dry-run label or similar qualifier.
	if !strings.Contains(strings.ToLower(out), "dry") {
		t.Errorf("dry-run output should mention 'dry'; got:\n%s", out)
	}
}

// TestRun_DryRun_NoManifestWritten verifies that dry-run does not write a progress manifest.
func TestRun_DryRun_NoManifestWritten(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--dry-run",
	}, &stdout, &stderr, panicFactory())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manifestPath := filepath.Join(dir, "fetch-progress.json")
	if _, statErr := os.Stat(manifestPath); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf("dry-run must not write fetch-progress.json; file found at %s", manifestPath)
	}
}

// TestRun_SuccessfulFetch_WritesProgressManifest verifies that a successful single-instrument
// fetch writes a progress manifest with the completed entry.
func TestRun_SuccessfulFetch_WritesProgressManifest(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})
	mock := &mockProvider{candles: []model.Candle{}}

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, mockFactory(mock))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	manifestPath := filepath.Join(dir, "fetch-progress.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected fetch-progress.json to be written: %v", err)
	}

	var manifest progressManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}

	if len(manifest.Completed) != 1 {
		t.Fatalf("expected 1 completed entry, got %d", len(manifest.Completed))
	}
	if manifest.Completed[0].Instrument != "NSE:RELIANCE" {
		t.Errorf("expected instrument NSE:RELIANCE, got %s", manifest.Completed[0].Instrument)
	}
	if manifest.Completed[0].Timeframe != "5min" {
		t.Errorf("expected timeframe 5min, got %s", manifest.Completed[0].Timeframe)
	}
}

// TestRun_PartialFailure_WritesManifestForCompletedOnly verifies that when the second
// instrument fetch fails, the manifest records the first (successful) but not the second.
func TestRun_PartialFailure_WritesManifestForCompletedOnly(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE", "NSE:TCS"})

	callCount := 0
	// First call succeeds, second fails.
	factory := func(_ fetchFlags) (provider.DataProvider, error) {
		return &callCountProvider{
			onCall: func(instrument string) ([]model.Candle, error) {
				callCount++
				if callCount == 2 {
					return nil, errors.New("network error")
				}
				return []model.Candle{}, nil
			},
		}, nil
	}

	var stdout, stderr bytes.Buffer
	// Partial failure is expected; ignore the error — we assert on manifest content instead.
	//nolint:errcheck // intentional: partial run returns error; manifest content is the assertion target
	_ = run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, factory)

	manifestPath := filepath.Join(dir, "fetch-progress.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("expected fetch-progress.json after partial failure: %v", err)
	}

	var manifest progressManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("invalid manifest JSON: %v", err)
	}

	if len(manifest.Completed) != 1 {
		t.Fatalf("expected 1 completed entry (first instrument only), got %d", len(manifest.Completed))
	}
	if manifest.Completed[0].Instrument != "NSE:RELIANCE" {
		t.Errorf("expected first completed instrument to be NSE:RELIANCE, got %s", manifest.Completed[0].Instrument)
	}
}

// TestRun_ResumeFromManifest_SkipsCompleted verifies that when a progress manifest
// already marks an instrument+timeframe as done, FetchCandles is not called for it.
func TestRun_ResumeFromManifest_SkipsCompleted(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE", "NSE:TCS"})

	// Pre-write a manifest marking NSE:RELIANCE/5min as done.
	existingManifest := progressManifest{
		Completed: []progressEntry{
			{Instrument: "NSE:RELIANCE", Timeframe: "5min"},
		},
	}
	data, err := json.Marshal(existingManifest)
	if err != nil {
		t.Fatalf("marshal existing manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fetch-progress.json"), data, 0o644); err != nil {
		t.Fatalf("write existing manifest: %v", err)
	}

	mock := &mockProvider{candles: []model.Candle{}}
	var stdout, stderr bytes.Buffer
	runErr := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, mockFactory(mock))

	if runErr != nil {
		t.Fatalf("unexpected error: %v", runErr)
	}

	// Only NSE:TCS should have been fetched (NSE:RELIANCE was already in manifest).
	if mock.callCount != 1 {
		t.Errorf("expected 1 FetchCandles call (skip NSE:RELIANCE), got %d", mock.callCount)
	}
	if len(mock.calls) > 0 && mock.calls[0].instrument != "NSE:TCS" {
		t.Errorf("expected fetch for NSE:TCS, got %s", mock.calls[0].instrument)
	}
}

// TestRun_EnvVarFallback_ApiKey verifies that KITE_API_KEY env var is used when --api-key is not set.
func TestRun_EnvVarFallback_ApiKey(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})
	mock := &mockProvider{candles: []model.Candle{}}

	t.Setenv("KITE_API_KEY", "env-api-key")
	t.Setenv("KITE_ACCESS_TOKEN", "env-access-token")

	var stdout, stderr bytes.Buffer
	// No --api-key or --access-token flags — must use env vars.
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
	}, &stdout, &stderr, mockFactory(mock))
	if err != nil {
		t.Fatalf("expected env var fallback to succeed, got: %v", err)
	}
}

// TestRun_ProgressLogging verifies that progress lines are printed to stdout after each fetch.
func TestRun_ProgressLogging(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})
	mock := &mockProvider{candles: []model.Candle{}}

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, mockFactory(mock))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, "NSE:RELIANCE") {
		t.Errorf("progress log missing instrument name; got:\n%s", out)
	}
	if !strings.Contains(out, "5min") {
		t.Errorf("progress log missing timeframe; got:\n%s", out)
	}
}

// TestRun_MultipleTimeframes verifies that FetchCandles is called once per instrument×timeframe.
func TestRun_MultipleTimeframes(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE", "NSE:TCS"})
	mock := &mockProvider{candles: []model.Candle{}}

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--timeframe", "daily",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, mockFactory(mock))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 2 instruments × 2 timeframes = 4 calls.
	if mock.callCount != 4 {
		t.Errorf("expected 4 FetchCandles calls (2 instruments × 2 timeframes), got %d", mock.callCount)
	}
}

// callCountProvider counts calls by instrument; lets partial failure tests inject custom error logic.
type callCountProvider struct {
	onCall func(instrument string) ([]model.Candle, error)
}

func (c *callCountProvider) FetchCandles(_ context.Context, instrument string, _ model.Timeframe, _, _ time.Time) ([]model.Candle, error) {
	return c.onCall(instrument)
}

func (c *callCountProvider) SupportedTimeframes() []model.Timeframe {
	return []model.Timeframe{model.Timeframe5Min}
}

// ── resolveBatchToken ─────────────────────────────────────────────────────────

func TestResolveBatchToken_usesFlagTokenWhenNonEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	// No file written — would error if read.

	got, err := resolveBatchToken("flag-token", tokenPath)
	if err != nil {
		t.Fatalf("resolveBatchToken: unexpected error: %v", err)
	}
	if got != "flag-token" {
		t.Errorf("resolveBatchToken() = %q, want %q", got, "flag-token")
	}
}

func TestResolveBatchToken_fallsBackToTokenFile_whenFlagEmpty(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	writeValidTokenFile(t, tokenPath, "file-token")

	got, err := resolveBatchToken("", tokenPath)
	if err != nil {
		t.Fatalf("resolveBatchToken: unexpected error: %v", err)
	}
	if got != "file-token" {
		t.Errorf("resolveBatchToken() = %q, want %q", got, "file-token")
	}
}

func TestResolveBatchToken_errorWhenFlagEmptyAndFileAbsent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "no-token.json")

	_, err := resolveBatchToken("", tokenPath)
	if err == nil {
		t.Fatal("resolveBatchToken: expected error when flag empty and file absent, got nil")
	}
	// Error must name the flag, env var, and file paths so users know all three options.
	for _, want := range []string{"--access-token", "KITE_ACCESS_TOKEN", tokenPath} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %q; got: %v", want, err)
		}
	}
}

func TestResolveBatchToken_errorWhenFlagEmptyAndFileCorrupt(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "token.json")
	if err := os.WriteFile(tokenPath, []byte("not-json"), 0o600); err != nil {
		t.Fatalf("write corrupt token: %v", err)
	}

	_, err := resolveBatchToken("", tokenPath)
	if err == nil {
		t.Fatal("resolveBatchToken: expected error for corrupt token file, got nil")
	}
}

// makeTestCandle builds a valid Candle for use in fetch-history tests.
func makeTestCandle(t *testing.T) model.Candle {
	t.Helper()
	ts := time.Date(2021, 1, 11, 3, 45, 0, 0, time.UTC) // 09:15 IST
	c, err := model.NewCandle("NSE:HDFCBANK", model.Timeframe5Min, ts, 841.0, 845.0, 839.0, 843.0, 4500)
	if err != nil {
		t.Fatalf("makeTestCandle: %v", err)
	}
	return c
}

// TestFetchOne_BadCandleSkip verifies the realistic OHLC bad-candle fix scenario:
// 1 bad candle skipped, remaining valid candles returned (the Zerodha artifact case).
// fetchOne must:
//   - return nil (treat as success)
//   - log a warning to stderr containing the specific skipped candle index
//   - write the instrument to completed in the manifest
//   - record skipped_candles with the correct index and non-empty reason
func TestFetchOne_BadCandleSkip(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:HDFCBANK"})

	skipped := []zerodha.SkippedCandle{
		{Index: 450, Reason: "candle: open (835.6000) must be within [low=837.4000, high=843.8000]"},
	}
	// Realistic scenario: 1 bad candle skipped, many valid candles returned.
	validCandles := []model.Candle{makeTestCandle(t)}

	factory := func(_ fetchFlags) (provider.DataProvider, error) {
		return &mockProvider{
			candles: validCandles,
			err:     &zerodha.ErrBadCandles{Instrument: "NSE:HDFCBANK", Skipped: skipped},
		}, nil
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2021-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, factory)
	// run() must succeed: valid candles returned alongside warning.
	if err != nil {
		t.Fatalf("run() with ErrBadCandles + valid candles: want nil error, got %v", err)
	}

	// stderr must contain a warning with the specific candle index "450".
	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "450") {
		t.Errorf("stderr must contain skipped candle index 450; got:\n%s", stderrOut)
	}

	// stdout must mention the fetch completed with skipped count.
	stdoutOut := stdout.String()
	if !strings.Contains(stdoutOut, "skipped") {
		t.Errorf("stdout must mention skipped candles; got:\n%s", stdoutOut)
	}

	// Manifest must record NSE:HDFCBANK as completed.
	manifestPath := filepath.Join(dir, "fetch-progress.json")
	data, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		t.Fatalf("expected fetch-progress.json to be written: %v", readErr)
	}
	var manifest progressManifest
	if unmarshalErr := json.Unmarshal(data, &manifest); unmarshalErr != nil {
		t.Fatalf("invalid manifest JSON: %v", unmarshalErr)
	}
	if len(manifest.Completed) != 1 {
		t.Fatalf("expected 1 completed entry, got %d", len(manifest.Completed))
	}
	if manifest.Completed[0].Instrument != "NSE:HDFCBANK" {
		t.Errorf("expected completed instrument NSE:HDFCBANK, got %s", manifest.Completed[0].Instrument)
	}

	// Manifest entry must record skipped_candles with correct index and non-empty reason.
	if len(manifest.Completed[0].SkippedCandles) != 1 {
		t.Fatalf("expected 1 skipped_candle in manifest, got %d", len(manifest.Completed[0].SkippedCandles))
	}
	sc := manifest.Completed[0].SkippedCandles[0]
	if sc.Index != 450 {
		t.Errorf("skipped_candle[0].Index = %d, want 450", sc.Index)
	}
	if sc.Reason == "" {
		t.Error("skipped_candle[0].Reason must not be empty")
	}
}

// TestFetchOne_BadCandleSkip_AllCandlesBad verifies the degenerate case where
// ALL candles in the response are bad (empty valid candle slice returned alongside
// *ErrBadCandles). fetchOne must:
//   - return a non-nil error (instrument not marked completed — will retry next run)
//   - log an error to stderr mentioning "all ... candles were invalid"
//   - NOT write the instrument to the completed list in the manifest
func TestFetchOne_BadCandleSkip_AllCandlesBad(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:HDFCBANK"})

	skipped := []zerodha.SkippedCandle{
		{Index: 0, Reason: "candle: open (835.6000) must be within [low=837.4000, high=843.8000]"},
	}
	// All candles bad — zero valid candles returned.
	factory := func(_ fetchFlags) (provider.DataProvider, error) {
		return &mockProvider{
			candles: []model.Candle{},
			err:     &zerodha.ErrBadCandles{Instrument: "NSE:HDFCBANK", Skipped: skipped},
		}, nil
	}

	var stdout, stderr bytes.Buffer
	//nolint:errcheck // intentional: all-bad-candles run returns error; we assert on side effects
	runErr := run([]string{
		"--universe", uPath,
		"--from", "2021-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, factory)

	// run() must return non-nil: all candles bad is treated as failure.
	if runErr == nil {
		t.Fatal("run() with all-bad ErrBadCandles: want non-nil error, got nil")
	}

	// stderr must mention the all-bad condition.
	stderrOut := stderr.String()
	if !strings.Contains(stderrOut, "all") || !strings.Contains(stderrOut, "invalid") {
		t.Errorf("stderr must mention all candles invalid; got:\n%s", stderrOut)
	}

	// Manifest must NOT record NSE:HDFCBANK as completed (nothing useful cached).
	manifestPath := filepath.Join(dir, "fetch-progress.json")
	data, readErr := os.ReadFile(manifestPath)
	if readErr != nil {
		// Manifest may or may not exist; if absent, that's also acceptable.
		if !errors.Is(readErr, os.ErrNotExist) {
			t.Fatalf("unexpected error reading manifest: %v", readErr)
		}
		return
	}
	var manifest progressManifest
	if unmarshalErr := json.Unmarshal(data, &manifest); unmarshalErr != nil {
		t.Fatalf("invalid manifest JSON: %v", unmarshalErr)
	}
	if len(manifest.Completed) != 0 {
		t.Errorf("expected 0 completed entries (all-bad instrument must not be marked done), got %d", len(manifest.Completed))
	}
}

// TestFetchOne_HardErrorStillFails verifies that a hard error (non-ErrBadCandles) from
// FetchCandles still propagates as a failure (fetchOne returns non-nil).
func TestFetchOne_HardErrorStillFails(t *testing.T) {
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})

	hardErr := errors.New("network timeout")
	factory := func(_ fetchFlags) (provider.DataProvider, error) {
		return &mockProvider{err: hardErr}, nil
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2021-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
		"--access-token", "testtoken",
	}, &stdout, &stderr, factory)

	// run() must return non-nil for a hard error.
	if err == nil {
		t.Fatal("run() with hard provider error: want non-nil error, got nil")
	}
}

// TestRun_PropagatesProviderFactoryError verifies that run() propagates errors returned
// by the provider factory. This tests that removing the early accessToken == "" validation
// from run() is safe — factory errors still surface.
func TestRun_PropagatesProviderFactoryError(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	uPath := writeUniverseYAML(t, dir, []string{"NSE:RELIANCE"})

	factoryErr := errors.New("no access token: set --access-token flag, KITE_ACCESS_TOKEN env var, or run cmd/backtest")
	errFactory := func(_ fetchFlags) (provider.DataProvider, error) {
		return nil, factoryErr
	}

	var stdout, stderr bytes.Buffer
	err := run([]string{
		"--universe", uPath,
		"--from", "2024-01-01",
		"--timeframe", "5min",
		"--cache-dir", dir,
		"--api-key", "testkey",
	}, &stdout, &stderr, errFactory)

	if err == nil {
		t.Fatal("expected error propagated from factory, got nil")
	}
	if !strings.Contains(err.Error(), "no access token") {
		t.Errorf("error should contain factory message; got: %v", err)
	}
}
