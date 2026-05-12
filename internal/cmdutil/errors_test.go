package cmdutil_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
)

// ---------------------------------------------------------------------------
// ExitCodeError
// ---------------------------------------------------------------------------

func TestExitCodeError_Error(t *testing.T) {
	t.Parallel()
	e := &cmdutil.ExitCodeError{Code: 2}
	if got := e.Error(); got != "exit code 2" {
		t.Errorf("ExitCodeError.Error() = %q, want %q", got, "exit code 2")
	}
}

func TestExitCodeError_ErrorsAs(t *testing.T) {
	t.Parallel()
	e := &cmdutil.ExitCodeError{Code: 2}
	wrapped := fmt.Errorf("outer: %w", e)
	var ee *cmdutil.ExitCodeError
	if !errors.As(wrapped, &ee) {
		t.Fatal("errors.As did not find *ExitCodeError in wrapped error")
	}
	if ee.Code != 2 {
		t.Errorf("ee.Code = %d, want 2", ee.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleIncompleteDataError — match path
// ---------------------------------------------------------------------------

func TestHandleIncompleteDataError_MatchesWrappedErrIncompleteData(t *testing.T) {
	t.Parallel()

	from := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ie := &zerodha.ErrIncompleteData{
		Instrument: "NSE:RELIANCE",
		From:       from,
		To:         to,
		Expected:   261,
		Got:        20,
	}
	wrapped := fmt.Errorf("engine: fetching candles: %w", ie)

	var stderr bytes.Buffer
	result := cmdutil.HandleIncompleteDataError(wrapped, &stderr)

	if result == nil {
		t.Fatal("HandleIncompleteDataError() = nil, want *ExitCodeError")
	}
	if result.Code != 2 {
		t.Errorf("result.Code = %d, want 2", result.Code)
	}

	out := stderr.String()
	if !strings.Contains(out, "incomplete data:") {
		t.Errorf("stderr missing 'incomplete data:' prefix; got: %q", out)
	}
	if !strings.Contains(out, "NSE:RELIANCE") {
		t.Errorf("stderr missing instrument name; got: %q", out)
	}
	if !strings.Contains(out, "2022-01-01") {
		t.Errorf("stderr missing from date; got: %q", out)
	}
	if !strings.Contains(out, "2023-01-01") {
		t.Errorf("stderr missing to date; got: %q", out)
	}
	if !strings.Contains(out, "261") {
		t.Errorf("stderr missing expected count; got: %q", out)
	}
	if !strings.Contains(out, "20") {
		t.Errorf("stderr missing got count; got: %q", out)
	}
}

func TestHandleIncompleteDataError_MatchesDirectErrIncompleteData(t *testing.T) {
	t.Parallel()

	ie := &zerodha.ErrIncompleteData{
		Instrument: "NSE:INFY",
		From:       time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		To:         time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
		Expected:   261,
		Got:        100,
	}

	var stderr bytes.Buffer
	result := cmdutil.HandleIncompleteDataError(ie, &stderr)

	if result == nil {
		t.Fatal("HandleIncompleteDataError() = nil for direct *ErrIncompleteData, want *ExitCodeError")
	}
	if result.Code != 2 {
		t.Errorf("result.Code = %d, want 2", result.Code)
	}
}

// ---------------------------------------------------------------------------
// HandleIncompleteDataError — no-match path (control tests)
// ---------------------------------------------------------------------------

func TestHandleIncompleteDataError_NoMatchForGenericError(t *testing.T) {
	t.Parallel()

	generic := fmt.Errorf("some generic provider error")
	var stderr bytes.Buffer
	result := cmdutil.HandleIncompleteDataError(generic, &stderr)

	if result != nil {
		t.Errorf("HandleIncompleteDataError() = %v, want nil for generic error", result)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr should be empty for non-matching error; got: %q", stderr.String())
	}
}

func TestHandleIncompleteDataError_NoMatchForNilError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	result := cmdutil.HandleIncompleteDataError(nil, &stderr)

	if result != nil {
		t.Errorf("HandleIncompleteDataError(nil) = %v, want nil", result)
	}
}

func TestHandleIncompleteDataError_NoMatchForOtherTypedError(t *testing.T) {
	t.Parallel()

	// A different typed error (not ErrIncompleteData) should return nil.
	other := fmt.Errorf("wrapped: %w", errors.New("auth required"))
	var stderr bytes.Buffer
	result := cmdutil.HandleIncompleteDataError(other, &stderr)

	if result != nil {
		t.Errorf("HandleIncompleteDataError() = %v, want nil for non-ErrIncompleteData error", result)
	}
}
