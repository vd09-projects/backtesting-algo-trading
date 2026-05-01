package cmdutil_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// ── ParseCommissionModel ──────────────────────────────────────────────────────

func TestParseCommissionModel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    model.CommissionModel
		wantErr bool
	}{
		{"zerodha accepted", "zerodha", model.CommissionZerodha, false},
		{"zerodha_full accepted", "zerodha_full", model.CommissionZerodhaFull, false},
		{"zerodha_full_mis accepted", "zerodha_full_mis", model.CommissionZerodhaFullMIS, false},
		{"flat accepted", "flat", model.CommissionFlat, false},
		{"percentage accepted", "percentage", model.CommissionPercentage, false},
		{"empty string rejected", "", "", true},
		{"unknown value rejected", "free", "", true},
		{"case-sensitive — Zerodha rejected", "Zerodha", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := cmdutil.ParseCommissionModel(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCommissionModel(%q): expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCommissionModel(%q): unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Errorf("ParseCommissionModel(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── LoadDotEnv ────────────────────────────────────────────────────────────────

func TestLoadDotEnv(t *testing.T) {
	tests := []struct {
		name    string
		content string
		preSet  map[string]string
		key     string
		want    string
	}{
		{
			name:    "sets key=value pair from file",
			content: "CMDUTIL_T_FOO=bar\n",
			key:     "CMDUTIL_T_FOO",
			want:    "bar",
		},
		{
			name:    "trims whitespace around key and value",
			content: " CMDUTIL_T_TRIM = trimmed \n",
			key:     "CMDUTIL_T_TRIM",
			want:    "trimmed",
		},
		{
			name:    "real env var takes precedence over file",
			content: "CMDUTIL_T_PREC=from_file\n",
			preSet:  map[string]string{"CMDUTIL_T_PREC": "from_env"},
			key:     "CMDUTIL_T_PREC",
			want:    "from_env",
		},
		{
			name:    "skips lines without = separator",
			content: "CMDUTIL_T_NOSEP\n",
			key:     "CMDUTIL_T_NOSEP",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.preSet {
				t.Setenv(k, v)
			}
			t.Cleanup(func() { os.Unsetenv(tt.key) }) //nolint:errcheck // Unsetenv failure in test cleanup is non-fatal

			dir := t.TempDir()
			f := filepath.Join(dir, ".env")
			if err := os.WriteFile(f, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}
			cmdutil.LoadDotEnv(f)
			if got := os.Getenv(tt.key); got != tt.want {
				t.Errorf("os.Getenv(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}

	t.Run("skips blank lines and comments", func(t *testing.T) {
		t.Cleanup(func() { os.Unsetenv("CMDUTIL_T_SKIP") }) //nolint:errcheck // Unsetenv failure in test cleanup is non-fatal
		dir := t.TempDir()
		f := filepath.Join(dir, ".env")
		if err := os.WriteFile(f, []byte("# comment\n\nCMDUTIL_T_SKIP=ok\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		cmdutil.LoadDotEnv(f)
		if got := os.Getenv("CMDUTIL_T_SKIP"); got != "ok" {
			t.Errorf("os.Getenv(CMDUTIL_T_SKIP) = %q, want %q", got, "ok")
		}
	})

	t.Run("missing file is silently ignored", func(t *testing.T) {
		cmdutil.LoadDotEnv("/nonexistent/path/.env") // must not panic or error
	})
}

// ── TokenFilePath ─────────────────────────────────────────────────────────────

func TestTokenFilePath(t *testing.T) {
	t.Run("returns BACKTEST_TOKEN_PATH override when set", func(t *testing.T) {
		t.Setenv("BACKTEST_TOKEN_PATH", "/custom/token.json")
		if got := cmdutil.TokenFilePath(); got != "/custom/token.json" {
			t.Errorf("TokenFilePath() = %q, want %q", got, "/custom/token.json")
		}
	})

	t.Run("returns default path under home dir when override is unset", func(t *testing.T) {
		t.Setenv("BACKTEST_TOKEN_PATH", "")
		got := cmdutil.TokenFilePath()
		if got == "" {
			t.Fatal("TokenFilePath() returned empty string")
		}
		if filepath.Base(got) != "token.json" {
			t.Errorf("TokenFilePath() base = %q, want token.json", filepath.Base(got))
		}
	})
}

// ── MustEnv ───────────────────────────────────────────────────────────────────

func TestMustEnv_set(t *testing.T) {
	t.Setenv("CMDUTIL_T_MUSTENV", "expected")
	if got := cmdutil.MustEnv("CMDUTIL_T_MUSTENV"); got != "expected" {
		t.Errorf("MustEnv() = %q, want %q", got, "expected")
	}
}

// ── Fatalf ────────────────────────────────────────────────────────────────────

// TestFatalf_exits verifies Fatalf exits with code 1.
// Uses the subprocess pattern because os.Exit cannot be tested in-process.
func TestFatalf_exits(t *testing.T) {
	if os.Getenv("CMDUTIL_RUN_FATALF") == "1" {
		cmdutil.Fatalf("subprocess test %s", "error")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestFatalf_exits$")
	cmd.Env = append(os.Environ(), "CMDUTIL_RUN_FATALF=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got: %v", err)
	}
}

// ── LoginFlow ─────────────────────────────────────────────────────────────────

// pipeStdin replaces os.Stdin with a pipe whose write end has the given content
// already written and closed. Restores the original stdin via t.Cleanup.
func pipeStdin(t *testing.T, content string) {
	t.Helper()
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(content); err != nil {
		t.Fatal(err)
	}
	w.Close() //nolint:errcheck // write-end closed after writing; failure is non-fatal in a test
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close() //nolint:errcheck // read-end cleanup; failure is non-fatal in a test
	})
}

// tokenServer returns an httptest.Server that responds to POST /session/token
// with a JSON body containing the given access token. Close it with t.Cleanup.
func tokenServer(t *testing.T, accessToken string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"success","data":{"access_token":"` + accessToken + `"}}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestLoginFlow_eofStdin verifies LoginFlow returns an error when stdin closes before input.
func TestLoginFlow_eofStdin(t *testing.T) {
	pipeStdin(t, "") // empty write + close → EOF on first Scan
	_, gotErr := cmdutil.LoginFlow(
		context.Background(), http.DefaultClient, "https://api.kite.trade",
		"key", "secret", filepath.Join(t.TempDir(), "token.json"),
	)
	if gotErr == nil {
		t.Error("LoginFlow() expected error on EOF stdin, got nil")
	}
}

func TestLoginFlow_emptyToken(t *testing.T) {
	pipeStdin(t, "\n") // scan succeeds but token is whitespace-only
	_, gotErr := cmdutil.LoginFlow(
		context.Background(), http.DefaultClient, "https://api.kite.trade",
		"key", "secret", filepath.Join(t.TempDir(), "token.json"),
	)
	if gotErr == nil {
		t.Fatal("LoginFlow() expected error for empty token, got nil")
	}
	if gotErr.Error() != "request_token cannot be empty" {
		t.Errorf("LoginFlow() error = %q, want %q", gotErr.Error(), "request_token cannot be empty")
	}
}

func TestLoginFlow_exchangeError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	t.Cleanup(srv.Close)

	pipeStdin(t, "myreqtoken\n")
	_, gotErr := cmdutil.LoginFlow(
		context.Background(), srv.Client(), srv.URL,
		"key", "secret", filepath.Join(t.TempDir(), "token.json"),
	)
	if gotErr == nil {
		t.Fatal("LoginFlow() expected error from ExchangeToken, got nil")
	}
}

func TestLoginFlow_success_tokenSaved(t *testing.T) {
	srv := tokenServer(t, "fresh-access-token")
	tokenPath := filepath.Join(t.TempDir(), "token.json")

	pipeStdin(t, "myreqtoken\n")
	got, err := cmdutil.LoginFlow(
		context.Background(), srv.Client(), srv.URL,
		"key", "secret", tokenPath,
	)
	if err != nil {
		t.Fatalf("LoginFlow() unexpected error: %v", err)
	}
	if got != "fresh-access-token" {
		t.Errorf("LoginFlow() = %q, want %q", got, "fresh-access-token")
	}
	if _, statErr := os.Stat(tokenPath); statErr != nil {
		t.Errorf("token file not created: %v", statErr)
	}
}

func TestLoginFlow_success_saveFails(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("root can write to read-only directories")
	}
	srv := tokenServer(t, "fresh-access-token")

	dir := t.TempDir()
	readOnly := filepath.Join(dir, "readonly")
	if err := os.MkdirAll(readOnly, 0o500); err != nil {
		t.Fatal(err)
	}

	pipeStdin(t, "myreqtoken\n")
	// Save failure must not cause LoginFlow to return an error — it prints a warning only.
	got, err := cmdutil.LoginFlow(
		context.Background(), srv.Client(), srv.URL,
		"key", "secret", filepath.Join(readOnly, "token.json"),
	)
	if err != nil {
		t.Fatalf("LoginFlow() unexpected error when save fails: %v", err)
	}
	if got != "fresh-access-token" {
		t.Errorf("LoginFlow() = %q, want %q", got, "fresh-access-token")
	}
}

// ── DefaultOutPath ────────────────────────────────────────────────────────────

func TestDefaultOutPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		strategy   string
		instrument string
		tf         string
		from       string
		to         string
		want       string
	}{
		{
			name:       "basic daily run",
			strategy:   "sma-crossover",
			instrument: "NSE:RELIANCE",
			tf:         "daily",
			from:       "2018-01-01",
			to:         "2024-01-01",
			want:       "sma-crossover-NSE_RELIANCE-daily-2018-01-01-2024-01-01.json",
		},
		{
			name:       "instrument with space",
			strategy:   "stub",
			instrument: "NSE:NIFTY 50",
			tf:         "daily",
			from:       "2020-01-01",
			to:         "2023-01-01",
			want:       "stub-NSE_NIFTY_50-daily-2020-01-01-2023-01-01.json",
		},
		{
			name:       "intraday timeframe",
			strategy:   "momentum",
			instrument: "NSE:INFY",
			tf:         "15min",
			from:       "2022-01-01",
			to:         "2023-01-01",
			want:       "momentum-NSE_INFY-15min-2022-01-01-2023-01-01.json",
		},
		{
			name:       "instrument already safe",
			strategy:   "donchian-breakout",
			instrument: "INSTRUMENT",
			tf:         "daily",
			from:       "2021-01-01",
			to:         "2022-01-01",
			want:       "donchian-breakout-INSTRUMENT-daily-2021-01-01-2022-01-01.json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := cmdutil.DefaultOutPath(tt.strategy, tt.instrument, tt.tf, tt.from, tt.to)
			if got != tt.want {
				t.Errorf("DefaultOutPath(%q, %q, %q, %q, %q) = %q, want %q",
					tt.strategy, tt.instrument, tt.tf, tt.from, tt.to, got, tt.want)
			}
		})
	}
}

// ── BuildProvider ─────────────────────────────────────────────────────────────

// TestBuildProvider_MissingAPIKey verifies BuildProvider returns an error via Fatalf
// when KITE_API_KEY is unset. Uses the subprocess pattern (same as TestFatalf_exits)
// because MustEnv calls Fatalf which calls os.Exit(1).
func TestBuildProvider_MissingAPIKey(t *testing.T) {
	if os.Getenv("CMDUTIL_RUN_BUILD_PROVIDER") == "1" {
		// In subprocess: clear env so MustEnv("KITE_API_KEY") fires Fatalf.
		os.Unsetenv("KITE_API_KEY")                        //nolint:errcheck // best-effort in subprocess test helper
		os.Unsetenv("KITE_API_SECRET")                     //nolint:errcheck // best-effort in subprocess test helper
		_, _ = cmdutil.BuildProvider(context.Background()) //nolint:errcheck // subprocess: expect Fatalf before return
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestBuildProvider_MissingAPIKey$")
	cmd.Env = append(os.Environ(), "CMDUTIL_RUN_BUILD_PROVIDER=1")
	// Remove any real credentials from the subprocess environment.
	filtered := cmd.Env[:0]
	for _, e := range cmd.Env {
		if !strings.HasPrefix(e, "KITE_API_KEY=") && !strings.HasPrefix(e, "KITE_API_SECRET=") {
			filtered = append(filtered, e)
		}
	}
	cmd.Env = filtered
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1 when KITE_API_KEY is unset, got: %v", err)
	}
}
