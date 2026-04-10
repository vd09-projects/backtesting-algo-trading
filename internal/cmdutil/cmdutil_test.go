package cmdutil_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
)

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

// TestLoginFlow_eofStdin verifies LoginFlow returns an error when stdin closes before input.
func TestLoginFlow_eofStdin(t *testing.T) {
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	w.Close() //nolint:errcheck // closing write-end to signal EOF; failure here is non-fatal in a test
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close() //nolint:errcheck // read-end cleanup; failure is non-fatal in a test
	})

	_, gotErr := cmdutil.LoginFlow(
		context.Background(), "key", "secret",
		filepath.Join(t.TempDir(), "token.json"),
	)
	if gotErr == nil {
		t.Error("LoginFlow() expected error on EOF stdin, got nil")
	}
}
