package stub_test

import (
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/stub"
)

func TestStub_Name(t *testing.T) {
	s := stub.New(model.TimeframeDaily)
	if got := s.Name(); got != "stub" {
		t.Errorf("Name() = %q, want %q", got, "stub")
	}
}

func TestStub_Timeframe(t *testing.T) {
	timeframes := []model.Timeframe{
		model.Timeframe1Min,
		model.Timeframe5Min,
		model.Timeframe15Min,
		model.TimeframeDaily,
		model.TimeframeWeekly,
	}
	for _, tf := range timeframes {
		s := stub.New(tf)
		if got := s.Timeframe(); got != tf {
			t.Errorf("Timeframe() = %q, want %q", got, tf)
		}
	}
}

func TestStub_Lookback(t *testing.T) {
	s := stub.New(model.TimeframeDaily)
	if got := s.Lookback(); got != 1 {
		t.Errorf("Lookback() = %d, want 1", got)
	}
}

func TestStub_Next_alwaysHold(t *testing.T) {
	s := stub.New(model.TimeframeDaily)

	if got := s.Next(nil); got != model.SignalHold {
		t.Errorf("Next(nil) = %q, want %q", got, model.SignalHold)
	}
	if got := s.Next([]model.Candle{}); got != model.SignalHold {
		t.Errorf("Next([]) = %q, want %q", got, model.SignalHold)
	}
	// Many candles — must still hold.
	many := make([]model.Candle, 100)
	if got := s.Next(many); got != model.SignalHold {
		t.Errorf("Next(many) = %q, want %q", got, model.SignalHold)
	}
}
