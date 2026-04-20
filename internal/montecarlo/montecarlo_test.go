package montecarlo_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vikrantdhawan/backtesting-algo-trading/internal/montecarlo"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

// makeTrade builds a Trade where ReturnOnNotional() == pnl/(entryPrice*qty).
func makeTrade(pnl, entryPrice, qty float64) model.Trade {
	return model.Trade{
		Instrument:  "TEST",
		Direction:   model.DirectionLong,
		EntryPrice:  entryPrice,
		Quantity:    qty,
		RealizedPnL: pnl,
		EntryTime:   time.Date(2024, 1, 1, 9, 15, 0, 0, time.UTC),
		ExitTime:    time.Date(2024, 1, 2, 9, 15, 0, 0, time.UTC),
	}
}

// positiveVariedTrades returns n trades cycling through returns in [+1%,+9%].
func positiveVariedTrades(n int) []model.Trade {
	rets := []float64{0.01, 0.03, 0.05, 0.02, 0.08, 0.04, 0.07, 0.02, 0.06, 0.09}
	trades := make([]model.Trade, n)
	for i := range trades {
		r := rets[i%len(rets)]
		trades[i] = makeTrade(r*100, 100, 1)
	}
	return trades
}

// negativeVariedTrades returns n trades cycling through returns in [-1%,-9%].
func negativeVariedTrades(n int) []model.Trade {
	rets := []float64{-0.01, -0.03, -0.05, -0.02, -0.08, -0.04, -0.07, -0.02, -0.06, -0.09}
	trades := make([]model.Trade, n)
	for i := range trades {
		r := rets[i%len(rets)]
		trades[i] = makeTrade(r*100, 100, 1)
	}
	return trades
}

// variedMixedTrades returns 20 trades with both positive and negative returns.
func variedMixedTrades() []model.Trade {
	rets := []float64{
		0.05, -0.02, 0.08, -0.01, 0.03, -0.04, 0.06, -0.03, 0.02, 0.07,
		-0.05, 0.04, -0.02, 0.09, -0.03, 0.01, 0.05, -0.06, 0.03, -0.01,
	}
	trades := make([]model.Trade, len(rets))
	for i, r := range rets {
		trades[i] = makeTrade(r*100, 100, 1)
	}
	return trades
}

func TestBootstrap_ZeroTrades(t *testing.T) {
	result := montecarlo.Bootstrap(nil, montecarlo.BootstrapConfig{NSimulations: 100, Seed: 42})
	assert.Equal(t, montecarlo.BootstrapResult{}, result)
}

func TestBootstrap_OneTrade(t *testing.T) {
	result := montecarlo.Bootstrap(
		[]model.Trade{makeTrade(5, 100, 1)},
		montecarlo.BootstrapConfig{NSimulations: 100, Seed: 42},
	)
	assert.Equal(t, montecarlo.BootstrapResult{}, result)
}

func TestBootstrap_Deterministic(t *testing.T) {
	trades := variedMixedTrades()
	cfg := montecarlo.BootstrapConfig{NSimulations: 1000, Seed: 99}
	assert.Equal(t, montecarlo.Bootstrap(trades, cfg), montecarlo.Bootstrap(trades, cfg))
}

func TestBootstrap_DifferentSeeds(t *testing.T) {
	trades := variedMixedTrades()
	r1 := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: 1000, Seed: 1})
	r2 := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: 1000, Seed: 2})
	assert.NotEqual(t, r1.SharpeP50, r2.SharpeP50)
}

func TestBootstrap_DefaultNSimulations(t *testing.T) {
	trades := variedMixedTrades()
	// NSimulations=0 must produce the same result as NSimulations=10_000 with the same seed.
	zeroSim := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{Seed: 7})
	explicit := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: 10_000, Seed: 7})
	assert.Equal(t, zeroSim, explicit)
}

func TestBootstrap_AllPositiveReturns(t *testing.T) {
	// 30 trades, all strictly positive returns (varied so sample variance is non-zero).
	trades := positiveVariedTrades(30)
	result := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: 1000, Seed: 42})

	// Percentile ordering invariants.
	assert.LessOrEqual(t, result.SharpeP5, result.SharpeP50)
	assert.LessOrEqual(t, result.SharpeP50, result.SharpeP95)
	assert.LessOrEqual(t, result.WorstDrawdownP5, result.WorstDrawdownP50)
	assert.LessOrEqual(t, result.WorstDrawdownP50, result.WorstDrawdownP95)

	// All returns positive → Sharpe > 0 in every simulation.
	assert.InDelta(t, 1.0, result.ProbPositiveSharpe, 1e-9)
	assert.Greater(t, result.MeanSharpe, 0.0)
	assert.Greater(t, result.SharpeP5, 0.0)

	// All returns positive → any resampled ordering produces monotonically increasing equity
	// → worst drawdown is always 0.
	assert.Equal(t, 0.0, result.WorstDrawdownP95)
}

func TestBootstrap_AllNegativeReturns(t *testing.T) {
	trades := negativeVariedTrades(30)
	result := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: 1000, Seed: 42})

	assert.LessOrEqual(t, result.SharpeP5, result.SharpeP50)
	assert.LessOrEqual(t, result.SharpeP50, result.SharpeP95)

	// All returns negative → Sharpe < 0 in every simulation.
	assert.InDelta(t, 0.0, result.ProbPositiveSharpe, 1e-9)
	assert.Less(t, result.MeanSharpe, 0.0)
	assert.Less(t, result.SharpeP95, 0.0)

	// All returns negative → equity falls monotonically → significant drawdown.
	assert.Greater(t, result.WorstDrawdownP5, 0.0)
}

func TestBootstrap_MixedReturns(t *testing.T) {
	trades := variedMixedTrades()
	result := montecarlo.Bootstrap(trades, montecarlo.BootstrapConfig{NSimulations: 1000, Seed: 42})

	assert.LessOrEqual(t, result.SharpeP5, result.SharpeP50)
	assert.LessOrEqual(t, result.SharpeP50, result.SharpeP95)
	assert.LessOrEqual(t, result.WorstDrawdownP5, result.WorstDrawdownP50)
	assert.LessOrEqual(t, result.WorstDrawdownP50, result.WorstDrawdownP95)

	// Mixed returns → some simulations positive, some negative.
	assert.Greater(t, result.ProbPositiveSharpe, 0.0)
	assert.Less(t, result.ProbPositiveSharpe, 1.0)

	// Mixed returns → drawdown present in every simulation.
	assert.Greater(t, result.WorstDrawdownP5, 0.0)
}
