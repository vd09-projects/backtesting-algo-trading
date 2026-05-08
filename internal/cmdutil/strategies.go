package cmdutil

import (
	"fmt"
	"math"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/bollinger"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/ccimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/donchian"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/macd"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/momentum"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/rsimeanrev"
	"github.com/vikrantdhawan/backtesting-algo-trading/strategies/smacrossover"
	stubstrategy "github.com/vikrantdhawan/backtesting-algo-trading/strategies/stub"
)

// GlobalRegistry is the single authoritative source for all strategy names and
// their full wiring: default factory, CLI param definitions, instance builder, and
// sweep factories. It is populated once at package-init time and read-only after.
//
// To add a new strategy: add one StrategyEntry block to buildRegistry() below.
// No other file requires changes — cmd mains call registry methods and pick up the
// new strategy automatically.
var GlobalRegistry = buildRegistry()

// defaultTimeframe is used by DefaultFactory entries when no timeframe is specified.
const defaultTimeframe = model.TimeframeDaily

// buildRegistry constructs and returns GlobalRegistry with all strategies fully wired.
func buildRegistry() *StrategyRegistry {
	r := NewStrategyRegistry()

	r.Register("stub", StrategyEntry{
		DefaultFactory: func() strategy.Strategy { return stubstrategy.New(defaultTimeframe) },
		Params:         nil,
		Build: func(tf model.Timeframe, _ map[string]float64) (strategy.Strategy, error) {
			return stubstrategy.New(tf), nil
		},
		SweepParams: nil,
	})

	r.Register("bollinger-mean-reversion", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := bollinger.New(defaultTimeframe, 20, 2.0)
			if err != nil {
				panic(fmt.Sprintf("bollinger-mean-reversion default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "bb-period", Default: 20, Usage: "bollinger-mean-reversion: Bollinger Band period", IsInt: true},
			{Name: "bb-num-std-dev", Default: 2.0, Usage: "bollinger-mean-reversion: number of standard deviations"},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return bollinger.New(tf, intParam(p, "bb-period"), p["bb-num-std-dev"])
		},
		SweepParams: []SweepParamDef{
			{Name: "bb-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return bollinger.New(tf, int(math.Round(v)), fixed["bb-num-std-dev"])
			}},
			{Name: "bb-num-std-dev", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return bollinger.New(tf, intParam(fixed, "bb-period"), v)
			}},
		},
	})

	r.Register("cci-mean-reversion", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := ccimeanrev.New(defaultTimeframe, 20, -100, 0)
			if err != nil {
				panic(fmt.Sprintf("cci-mean-reversion default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "cci-period", Default: 20, Usage: "cci-mean-reversion: CCI period", IsInt: true},
			{Name: "cci-entry-threshold", Default: -100, Usage: "cci-mean-reversion: entry threshold (buy when CCI < this)", IsInt: true},
			{Name: "cci-exit-threshold", Default: 0, Usage: "cci-mean-reversion: exit threshold (sell when CCI crosses above this)", IsInt: true},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return ccimeanrev.New(tf, intParam(p, "cci-period"), intParam(p, "cci-entry-threshold"), intParam(p, "cci-exit-threshold"))
		},
		SweepParams: []SweepParamDef{
			{Name: "cci-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return ccimeanrev.New(tf, int(math.Round(v)), intParam(fixed, "cci-entry-threshold"), intParam(fixed, "cci-exit-threshold"))
			}},
			{Name: "cci-entry-threshold", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return ccimeanrev.New(tf, intParam(fixed, "cci-period"), int(math.Round(v)), intParam(fixed, "cci-exit-threshold"))
			}},
			{Name: "cci-exit-threshold", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return ccimeanrev.New(tf, intParam(fixed, "cci-period"), intParam(fixed, "cci-entry-threshold"), int(math.Round(v)))
			}},
		},
	})

	r.Register("donchian-breakout", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := donchian.New(defaultTimeframe, 20)
			if err != nil {
				panic(fmt.Sprintf("donchian-breakout default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "donchian-period", Default: 20, Usage: "donchian-breakout: channel lookback period", IsInt: true},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return donchian.New(tf, intParam(p, "donchian-period"))
		},
		SweepParams: []SweepParamDef{
			{Name: "donchian-period", Factory: func(v float64, tf model.Timeframe, _ map[string]float64) (strategy.Strategy, error) {
				return donchian.New(tf, int(math.Round(v)))
			}},
		},
	})

	r.Register("macd-crossover", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := macd.New(defaultTimeframe, 12, 26, 9)
			if err != nil {
				panic(fmt.Sprintf("macd-crossover default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "macd-fast-period", Default: 12, Usage: "macd-crossover: fast EMA period", IsInt: true},
			{Name: "macd-slow-period", Default: 26, Usage: "macd-crossover: slow EMA period", IsInt: true},
			{Name: "macd-signal-period", Default: 9, Usage: "macd-crossover: signal EMA period", IsInt: true},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return macd.New(tf, intParam(p, "macd-fast-period"), intParam(p, "macd-slow-period"), intParam(p, "macd-signal-period"))
		},
		SweepParams: []SweepParamDef{
			{Name: "macd-fast-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return macd.New(tf, int(math.Round(v)), intParam(fixed, "macd-slow-period"), intParam(fixed, "macd-signal-period"))
			}},
			{Name: "macd-slow-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return macd.New(tf, intParam(fixed, "macd-fast-period"), int(math.Round(v)), intParam(fixed, "macd-signal-period"))
			}},
		},
	})

	r.Register("momentum", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := momentum.New(defaultTimeframe, 231, 10.0)
			if err != nil {
				panic(fmt.Sprintf("momentum default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "momentum-lookback", Default: 231, Usage: "momentum: ROC lookback period (default 231 = 252-21, skip-last-month convention)", IsInt: true},
			{Name: "momentum-threshold", Default: 10.0, Usage: "momentum: ROC threshold in percent (buy above, sell below negative)"},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return momentum.New(tf, intParam(p, "momentum-lookback"), p["momentum-threshold"])
		},
		SweepParams: []SweepParamDef{
			{Name: "momentum-lookback", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return momentum.New(tf, int(math.Round(v)), fixed["momentum-threshold"])
			}},
			{Name: "momentum-threshold", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return momentum.New(tf, intParam(fixed, "momentum-lookback"), v)
			}},
		},
	})

	r.Register("rsi-mean-reversion", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := rsimeanrev.New(defaultTimeframe, 14, 30, 70)
			if err != nil {
				panic(fmt.Sprintf("rsi-mean-reversion default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "rsi-period", Default: 14, Usage: "rsi-mean-reversion: RSI period", IsInt: true},
			{Name: "oversold", Default: 30, Usage: "rsi-mean-reversion: oversold threshold (buy below)"},
			{Name: "overbought", Default: 70, Usage: "rsi-mean-reversion: overbought threshold (sell above)"},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return rsimeanrev.New(tf, intParam(p, "rsi-period"), p["oversold"], p["overbought"])
		},
		SweepParams: []SweepParamDef{
			{Name: "rsi-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return rsimeanrev.New(tf, int(math.Round(v)), fixed["oversold"], fixed["overbought"])
			}},
			{Name: "oversold", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return rsimeanrev.New(tf, intParam(fixed, "rsi-period"), v, 100-v)
			}},
		},
	})

	r.Register("sma-crossover", StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			s, err := smacrossover.New(defaultTimeframe, 10, 50)
			if err != nil {
				panic(fmt.Sprintf("sma-crossover default factory: %v", err))
			}
			return s
		},
		Params: []ParamDef{
			{Name: "fast-period", Default: 10, Usage: "sma-crossover: fast SMA period", IsInt: true},
			{Name: "slow-period", Default: 50, Usage: "sma-crossover: slow SMA period", IsInt: true},
		},
		Build: func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
			return smacrossover.New(tf, intParam(p, "fast-period"), intParam(p, "slow-period"))
		},
		SweepParams: []SweepParamDef{
			{Name: "fast-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return smacrossover.New(tf, int(math.Round(v)), intParam(fixed, "slow-period"))
			}},
			{Name: "slow-period", Factory: func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error) {
				return smacrossover.New(tf, intParam(fixed, "fast-period"), int(math.Round(v)))
			}},
		},
	})

	return r
}

// intParam reads a float64 from p by name and rounds to int. Returns zero if the
// key is absent (programming error — callers should populate params via
// RegisterFlags+BuildParamMap or DefaultParams before calling Build).
func intParam(p map[string]float64, name string) int {
	return int(math.Round(p[name]))
}
