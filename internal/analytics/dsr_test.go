package analytics_test

import (
	"math"
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/analytics"
)

func TestDSR_MoreTrialsLowerValue(t *testing.T) {
	t.Parallel()
	// Same observed Sharpe and nObservations; more trials must produce a lower DSR.
	low := analytics.DSR(1.0, 10, 252)
	high := analytics.DSR(1.0, 100, 252)
	if high >= low {
		t.Errorf("DSR(1.0, 100, 252)=%.6f should be < DSR(1.0, 10, 252)=%.6f", high, low)
	}
}

func TestDSR_HigherSharpeHigherValue(t *testing.T) {
	t.Parallel()
	low := analytics.DSR(1.0, 10, 252)
	high := analytics.DSR(2.0, 10, 252)
	if high <= low {
		t.Errorf("DSR(2.0, ...)=%.6f should be > DSR(1.0, ...)=%.6f", high, low)
	}
}

func TestDSR_SingleTrialNoDeflation(t *testing.T) {
	t.Parallel()
	// nTrials=1: no multiple testing, DSR equals observed Sharpe.
	got := analytics.DSR(1.5, 1, 252)
	if math.Abs(got-1.5) > 1e-12 {
		t.Errorf("DSR(1.5, 1, 252): got %.12f, want 1.5", got)
	}
}

func TestDSR_InsufficientObservations(t *testing.T) {
	t.Parallel()
	// nObservations=1: undefined SE, returns observed Sharpe unchanged.
	got := analytics.DSR(1.5, 10, 1)
	if math.Abs(got-1.5) > 1e-12 {
		t.Errorf("DSR(1.5, 10, 1): got %.12f, want 1.5", got)
	}
}

func TestDSR_MoreObservationsSmallerDeflation(t *testing.T) {
	t.Parallel()
	// Same observed Sharpe and nTrials; more observations → smaller SE → smaller deflation → higher DSR.
	few := analytics.DSR(1.0, 10, 50)
	many := analytics.DSR(1.0, 10, 500)
	if many <= few {
		t.Errorf("DSR with 500 obs (%.6f) should be > DSR with 50 obs (%.6f)", many, few)
	}
}
