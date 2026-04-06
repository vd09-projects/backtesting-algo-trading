package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/engine"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// makeDailyCandles generates n valid daily OHLCV candles starting from 2015-01-01.
// Price oscillates so the test exercises realistic signal patterns without drifting to zero.
func makeDailyCandles(n int) []model.Candle {
	candles := make([]model.Candle, n)
	base := time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range candles {
		open := 100.0 + float64(i%200) // oscillates 100–299
		candles[i] = model.Candle{
			Instrument: "NIFTY",
			Timeframe:  model.TimeframeDaily,
			Timestamp:  base.AddDate(0, 0, i),
			Open:       open,
			High:       open + 5,
			Low:        open - 4,
			Close:      open + 1,
			Volume:     100_000,
		}
	}
	return candles
}

// leanStrategy is a zero-allocation strategy that alternates Buy/Sell each bar.
// It never copies the candles slice, keeping the benchmark focused on engine overhead.
type leanStrategy struct{ i int }

func (s *leanStrategy) Name() string               { return "lean" }
func (s *leanStrategy) Timeframe() model.Timeframe { return model.TimeframeDaily }
func (s *leanStrategy) Lookback() int              { return 1 }
func (s *leanStrategy) Next(_ []model.Candle) model.Signal {
	s.i++
	if s.i%2 == 0 {
		return model.SignalSell
	}
	return model.SignalBuy
}

// BenchmarkEngineRun processes 10 years of daily candles (~3650 bars) with a realistic
// OrderConfig (Zerodha commission + 0.05% slippage). Budget: < 1ms per iteration.
func BenchmarkEngineRun(b *testing.B) {
	const tenYearDays = 365 * 10
	candles := makeDailyCandles(tenYearDays)

	cfg := engine.Config{
		Instrument:           "NIFTY",
		From:                 candles[0].Timestamp,
		To:                   candles[len(candles)-1].Timestamp.Add(24 * time.Hour),
		InitialCash:          1_000_000,
		PositionSizeFraction: 0.1,
		OrderConfig: model.OrderConfig{
			SlippagePct:     0.0005,
			CommissionModel: model.CommissionZerodha,
		},
	}
	prov := &stubProvider{candles: candles}

	b.ResetTimer()
	for range b.N {
		s := &leanStrategy{}
		e := engine.New(cfg)
		if err := e.Run(context.Background(), prov, s); err != nil {
			b.Fatal(err)
		}
	}

	// Budget: 10 years of bars must complete in < 1ms per iteration (≈ 270 ns/bar).
	nsPerOp := b.Elapsed().Nanoseconds() / int64(b.N)
	if nsPerOp > 1_000_000 {
		b.Errorf("too slow: %d ns/op (budget: 1ms = 1_000_000 ns/op)", nsPerOp)
	}
}
