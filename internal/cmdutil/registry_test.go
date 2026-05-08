package cmdutil_test

import (
	"flag"
	"slices"
	"strings"
	"testing"

	"github.com/vikrantdhawan/backtesting-algo-trading/internal/cmdutil"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// ---------------------------------------------------------------------------
// StrategyRegistry contract (unit tests on a hand-built registry)
// ---------------------------------------------------------------------------

func TestStrategyRegistry_KnownStrategyReturnsFactory(t *testing.T) {
	t.Parallel()

	known := cmdutil.GlobalRegistry.ListStrategies()
	if len(known) == 0 {
		t.Fatal("GlobalRegistry.ListStrategies() returned empty slice — no strategies registered")
	}

	for _, name := range known {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			factory := cmdutil.GlobalRegistry.MustGet(name)
			if factory == nil {
				t.Fatalf("GlobalRegistry.MustGet(%q) returned nil factory", name)
			}
			instance := factory()
			if instance == nil {
				t.Fatalf("factory for %q returned nil strategy", name)
			}
		})
	}
}

func TestStrategyRegistry_UnknownStrategyPanics(t *testing.T) {
	t.Parallel()

	const unknown = "no-such-strategy"
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("MustGet with unknown name did not panic")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("panic value is not a string: %T: %v", r, r)
		}
		if !strings.Contains(msg, unknown) {
			t.Errorf("panic message %q does not mention unknown strategy name %q", msg, unknown)
		}
		if !strings.Contains(msg, "available") && !strings.Contains(msg, "known") {
			t.Errorf("panic message %q does not list available strategies", msg)
		}
	}()

	cmdutil.GlobalRegistry.MustGet(unknown)
}

func TestListStrategies_ReturnsSortedNames(t *testing.T) {
	t.Parallel()

	names := cmdutil.GlobalRegistry.ListStrategies()
	if len(names) == 0 {
		t.Fatal("ListStrategies() returned empty slice")
	}

	if !slices.IsSorted(names) {
		t.Errorf("ListStrategies() = %v — not sorted", names)
	}
}

func TestStrategyRegistry_Register_MustGet(t *testing.T) {
	t.Parallel()

	reg := cmdutil.NewStrategyRegistry()

	called := false
	reg.Register("test-strategy", cmdutil.StrategyEntry{
		DefaultFactory: func() strategy.Strategy {
			called = true
			return &testStubStrategy{}
		},
	})

	factory := reg.MustGet("test-strategy")
	if factory == nil {
		t.Fatal("MustGet returned nil factory for registered strategy")
	}
	instance := factory()
	if instance == nil {
		t.Fatal("factory returned nil instance")
	}
	if !called {
		t.Error("factory was not called")
	}
}

func TestStrategyRegistry_Register_DuplicatePanics(t *testing.T) {
	t.Parallel()

	reg := cmdutil.NewStrategyRegistry()
	reg.Register("dup", cmdutil.StrategyEntry{
		DefaultFactory: func() strategy.Strategy { return &testStubStrategy{} },
	})

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("registering duplicate name did not panic")
		}
	}()
	reg.Register("dup", cmdutil.StrategyEntry{
		DefaultFactory: func() strategy.Strategy { return &testStubStrategy{} },
	})
}

// ---------------------------------------------------------------------------
// GlobalRegistry — Build
// ---------------------------------------------------------------------------

func TestGlobalRegistry_Build_AllStrategies(t *testing.T) {
	t.Parallel()

	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := cmdutil.GlobalRegistry.DefaultParams(name)
			s, err := cmdutil.GlobalRegistry.Build(name, model.TimeframeDaily, p)
			if err != nil {
				t.Fatalf("Build(%q) error: %v", name, err)
			}
			if s == nil {
				t.Fatalf("Build(%q) returned nil strategy", name)
			}
		})
	}
}

func TestGlobalRegistry_Build_UnknownStrategy(t *testing.T) {
	t.Parallel()
	_, err := cmdutil.GlobalRegistry.Build("no-such-strategy", model.TimeframeDaily, nil)
	if err == nil {
		t.Fatal("expected error for unknown strategy, got nil")
	}
}

// ---------------------------------------------------------------------------
// GlobalRegistry — WalkForwardFactory
// ---------------------------------------------------------------------------

func TestGlobalRegistry_WalkForwardFactory_AllStrategies(t *testing.T) {
	t.Parallel()

	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := cmdutil.GlobalRegistry.DefaultParams(name)
			factory, err := cmdutil.GlobalRegistry.WalkForwardFactory(name, model.TimeframeDaily, p)
			if err != nil {
				t.Fatalf("WalkForwardFactory(%q) error: %v", name, err)
			}
			if factory == nil {
				t.Fatalf("WalkForwardFactory(%q) returned nil", name)
			}
			s := factory()
			if s == nil {
				t.Fatalf("WalkForwardFactory(%q): factory() returned nil", name)
			}
		})
	}
}

func TestGlobalRegistry_WalkForwardFactory_InvalidParams(t *testing.T) {
	t.Parallel()
	// fast-period >= slow-period is invalid for sma-crossover.
	bad := map[string]float64{
		"fast-period": 50,
		"slow-period": 10,
	}
	_, err := cmdutil.GlobalRegistry.WalkForwardFactory("sma-crossover", model.TimeframeDaily, bad)
	if err == nil {
		t.Fatal("expected error for invalid sma-crossover params (fast >= slow), got nil")
	}
}

// ---------------------------------------------------------------------------
// GlobalRegistry — SweepFactory
// ---------------------------------------------------------------------------

func TestGlobalRegistry_SweepFactory_AllSweepParams(t *testing.T) {
	t.Parallel()

	// Every strategy except stub should have at least one sweep param.
	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		if name == "stub" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixed := cmdutil.GlobalRegistry.DefaultParams(name)
			entry := cmdutil.GlobalRegistry.SweepParamNames(name)
			if len(entry) == 0 {
				t.Fatalf("%s has no SweepParams registered", name)
			}
			for _, sweepParam := range entry {
				factory, err := cmdutil.GlobalRegistry.SweepFactory(name, sweepParam, model.TimeframeDaily, fixed)
				if err != nil {
					t.Fatalf("SweepFactory(%q, %q) error: %v", name, sweepParam, err)
				}
				// Call the factory with the default value for the swept param.
				defVal := fixed[sweepParam]
				s, err := factory(defVal)
				if err != nil {
					t.Fatalf("SweepFactory(%q, %q) factory(%v) error: %v", name, sweepParam, defVal, err)
				}
				if s == nil {
					t.Fatalf("SweepFactory(%q, %q) factory returned nil", name, sweepParam)
				}
			}
		})
	}
}

func TestGlobalRegistry_SweepFactory_UnsupportedParam(t *testing.T) {
	t.Parallel()
	fixed := cmdutil.GlobalRegistry.DefaultParams("sma-crossover")
	_, err := cmdutil.GlobalRegistry.SweepFactory("sma-crossover", "no-such-param", model.TimeframeDaily, fixed)
	if err == nil {
		t.Fatal("expected error for unsupported sweep-param, got nil")
	}
	if !strings.Contains(err.Error(), "no-such-param") {
		t.Errorf("error should mention the bad param name; got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// GlobalRegistry — RegisterFlags + BuildParamMap
// ---------------------------------------------------------------------------

func TestGlobalRegistry_RegisterFlags_AllParamsCovered(t *testing.T) {
	t.Parallel()

	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	ptrs := cmdutil.GlobalRegistry.RegisterFlags(fs)
	if err := fs.Parse([]string{}); err != nil {
		t.Fatalf("Parse empty args: %v", err)
	}
	params := cmdutil.BuildParamMap(ptrs)

	// Every strategy's params must be in the returned map with their defaults.
	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		defs := cmdutil.GlobalRegistry.DefaultParams(name)
		for k, want := range defs {
			got, ok := params[k]
			if !ok {
				t.Errorf("strategy %q param %q missing from RegisterFlags output", name, k)
				continue
			}
			if got != want {
				t.Errorf("strategy %q param %q: got default %v, want %v", name, k, got, want)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// GlobalRegistry — ParamsMap
// ---------------------------------------------------------------------------

func TestGlobalRegistry_ParamsMap_AllStrategies(t *testing.T) {
	t.Parallel()

	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := cmdutil.GlobalRegistry.DefaultParams(name)
			m := cmdutil.GlobalRegistry.ParamsMap(name, p)
			// stub has no params — nil map is valid.
			defs := cmdutil.GlobalRegistry.DefaultParams(name)
			if len(defs) > 0 && len(m) == 0 {
				t.Errorf("ParamsMap(%q) returned empty map but strategy has params", name)
			}
			for k := range defs {
				if _, ok := m[k]; !ok {
					t.Errorf("ParamsMap(%q): param %q missing from result", name, k)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// GlobalRegistry — DefaultParams
// ---------------------------------------------------------------------------

func TestGlobalRegistry_DefaultParams_AllStrategies(t *testing.T) {
	t.Parallel()

	for _, name := range cmdutil.GlobalRegistry.ListStrategies() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			p := cmdutil.GlobalRegistry.DefaultParams(name)
			// nil is valid for stub (no params); non-nil map for all others.
			if name != "stub" && p == nil {
				t.Fatalf("DefaultParams(%q) returned nil for non-stub strategy", name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

type testStubStrategy struct{}

func (s *testStubStrategy) Name() string                       { return "test-stub" }
func (s *testStubStrategy) Timeframe() model.Timeframe         { return model.TimeframeDaily }
func (s *testStubStrategy) Lookback() int                      { return 1 }
func (s *testStubStrategy) Next(_ []model.Candle) model.Signal { return model.SignalHold }
