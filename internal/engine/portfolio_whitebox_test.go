package engine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
)

func TestApplySignal_UnknownSignalReturnsError(t *testing.T) {
	p := newPortfolio(10_000, model.OrderConfig{})
	err := p.applySignal(model.Signal("BOGUS"), "TEST", 100.0, time.Now(), 1.0)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown signal")
	assert.Contains(t, err.Error(), "BOGUS")
}

func TestApplySignal_HoldIsNoOp(t *testing.T) {
	p := newPortfolio(10_000, model.OrderConfig{})
	err := p.applySignal(model.SignalHold, "TEST", 100.0, time.Now(), 1.0)
	require.NoError(t, err)
	assert.Empty(t, p.Trades)
	assert.Empty(t, p.Positions)
	assert.Equal(t, 10_000.0, p.Cash)
}
