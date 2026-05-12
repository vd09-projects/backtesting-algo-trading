package cmdutil

import (
	"errors"
	"fmt"
	"io"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/provider/zerodha"
)

// ExitCodeError is returned by run() when the process should exit with a
// specific non-zero code. main() does errors.As(err, &ee); os.Exit(ee.Code).
//
// **Decision (ExitCodeError promoted to internal/cmdutil, shared across cmd/backtest and cmd/walk-forward) — convention: experimental**
// scope: internal/cmdutil, cmd/backtest, cmd/walk-forward
// tags: exit-code, errors, DRY, ErrIncompleteData, TASK-0083
// owner: priya
//
// Previously only cmd/walk-forward had a local exitCodeError. cmd/backtest and
// cmd/universe-sweep lacked it. Promoting to cmdutil avoids three copies of the
// same three-line struct and keeps exit-code semantics consistent across all
// single-instrument cmd/ binaries. Walk-forward's local type is removed in favor
// of this one.
type ExitCodeError struct{ Code int }

func (e *ExitCodeError) Error() string {
	return fmt.Sprintf("exit code %d", e.Code)
}

// HandleIncompleteDataError inspects err for *zerodha.ErrIncompleteData using
// errors.As. If found, it prints the per-instrument diagnostic to stderr and
// returns *ExitCodeError{Code: 2}. If not found, it returns nil — the caller
// should handle the error normally.
//
// **Decision (HandleIncompleteDataError in internal/cmdutil, not duplicated in 3 package main files) — convention: experimental**
// scope: internal/cmdutil, cmd/backtest, cmd/walk-forward
// tags: ErrIncompleteData, exit-code, DRY, diagnostic, TASK-0083
// owner: priya
//
// Three cmd/ binaries (backtest, walk-forward, and potentially fetch-history)
// need identical typed-error inspection for *ErrIncompleteData. Extracting to
// cmdutil avoids duplicating the errors.As check and the diagnostic format string.
// The diagnostic format ("incomplete data: instrument=…") matches the task AC
// exactly and is tested at this layer.
func HandleIncompleteDataError(err error, stderr io.Writer) *ExitCodeError {
	var ie *zerodha.ErrIncompleteData
	if !errors.As(err, &ie) {
		return nil
	}
	fmt.Fprintf(stderr, "incomplete data: instrument=%s from=%s to=%s expected≈%d got=%d\n", //nolint:errcheck // diagnostic to stderr; non-fatal write
		ie.Instrument,
		ie.From.Format("2006-01-02"),
		ie.To.Format("2006-01-02"),
		ie.Expected,
		ie.Got,
	)
	return &ExitCodeError{Code: 2}
}
