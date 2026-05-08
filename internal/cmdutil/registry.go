package cmdutil

import (
	"flag"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/model"
	"github.com/vikrantdhawan/backtesting-algo-trading/pkg/strategy"
)

// ParamDef describes one named float64 parameter for a strategy.
// IsInt signals that the underlying value is integer-typed; callers round with
// int(math.Round(v)). Using float64 uniformly for all flag values lets
// RegisterFlags register a single float64 flag per param regardless of kind.
type ParamDef struct {
	Name    string
	Default float64
	Usage   string
	IsInt   bool
}

// SweepParamDef defines one sweepable dimension for cmd/sweep.
// Factory receives the swept value, timeframe, and the full fixed-param map and
// returns a fresh strategy instance (or a validation error).
type SweepParamDef struct {
	Name    string
	Factory func(v float64, tf model.Timeframe, fixed map[string]float64) (strategy.Strategy, error)
}

// StrategyEntry holds all cmdutil-layer wiring for one strategy.
//
// Adding a new strategy = add one StrategyEntry block to buildRegistry() in
// strategies.go. No other file changes required.
type StrategyEntry struct {
	// DefaultFactory returns a fresh instance with default params and defaultTimeframe.
	// Used by cmd/signal-audit and MustGet-based name validation.
	DefaultFactory func() strategy.Strategy

	// Params lists all CLI-level parameters. RegisterFlags iterates these to
	// register flag.Float64 entries on a FlagSet.
	Params []ParamDef

	// Build constructs a strategy instance from the given timeframe and param map.
	Build func(tf model.Timeframe, p map[string]float64) (strategy.Strategy, error)

	// SweepParams lists the dimensions cmd/sweep can vary.
	SweepParams []SweepParamDef
}

// StrategyRegistry maps strategy names to their full wiring.
type StrategyRegistry struct {
	m map[string]StrategyEntry
}

// NewStrategyRegistry returns an empty registry.
func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{m: make(map[string]StrategyEntry)}
}

// Register adds entry under name. Panics on duplicate registration — duplicate
// registration is a programming error that must surface at startup.
func (r *StrategyRegistry) Register(name string, entry StrategyEntry) {
	if _, exists := r.m[name]; exists {
		panic(fmt.Sprintf("cmdutil.StrategyRegistry: duplicate registration for %q", name))
	}
	r.m[name] = entry
}

// MustGet returns the default factory for name. Panics with a descriptive message
// on unknown names — fail-fast startup validation in cmd mains.
func (r *StrategyRegistry) MustGet(name string) func() strategy.Strategy {
	entry, ok := r.m[name]
	if !ok {
		available := r.ListStrategies()
		panic(fmt.Sprintf(
			"unknown strategy %q; available: %s",
			name, strings.Join(available, ", "),
		))
	}
	return entry.DefaultFactory
}

// ListStrategies returns registered strategy names in lexicographic order.
func (r *StrategyRegistry) ListStrategies() []string {
	names := make([]string, 0, len(r.m))
	for name := range r.m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Build constructs a single strategy instance from the given timeframe and param map.
func (r *StrategyRegistry) Build(name string, tf model.Timeframe, p map[string]float64) (strategy.Strategy, error) {
	entry, ok := r.m[name]
	if !ok {
		return nil, fmt.Errorf("unknown strategy %q", name)
	}
	if entry.Build == nil {
		return nil, fmt.Errorf("strategy %q has no Build function registered", name)
	}
	return entry.Build(tf, p)
}

// WalkForwardFactory validates params eagerly (returns an error on bad params), then
// returns a zero-arg factory that constructs a fresh instance per call. Each walk-forward
// fold calls the factory to get its own isolated state.
func (r *StrategyRegistry) WalkForwardFactory(name string, tf model.Timeframe, p map[string]float64) (func() strategy.Strategy, error) {
	entry, ok := r.m[name]
	if !ok {
		return nil, fmt.Errorf("unknown strategy %q", name)
	}
	if entry.Build == nil {
		return nil, fmt.Errorf("strategy %q has no Build function registered", name)
	}
	// Validate params once — fail at startup, not mid-fold.
	if _, err := entry.Build(tf, p); err != nil {
		return nil, fmt.Errorf("%s params: %w", name, err)
	}
	build := entry.Build
	return func() strategy.Strategy {
		s, err := build(tf, p)
		if err != nil {
			panic(fmt.Sprintf("%s: params validated at startup, unexpected error: %v", name, err))
		}
		return s
	}, nil
}

// SweepFactory returns a one-param factory that sweeps sweepParam while holding all
// other params in fixed at their flag-parsed values. Returns an error if sweepParam
// is not registered for the named strategy.
func (r *StrategyRegistry) SweepFactory(name, sweepParam string, tf model.Timeframe, fixed map[string]float64) (func(float64) (strategy.Strategy, error), error) {
	entry, ok := r.m[name]
	if !ok {
		return nil, fmt.Errorf("unknown strategy %q", name)
	}
	for _, sp := range entry.SweepParams {
		if sp.Name == sweepParam {
			factory := sp.Factory
			return func(v float64) (strategy.Strategy, error) {
				return factory(v, tf, fixed)
			}, nil
		}
	}
	names := make([]string, len(entry.SweepParams))
	for i, sp := range entry.SweepParams {
		names[i] = sp.Name
	}
	return nil, fmt.Errorf("%s does not support sweep-param %q; use: %s",
		name, sweepParam, strings.Join(names, ", "))
}

// SweepParamNames returns the names of all registered sweep params for name.
// Returns nil for unknown names or strategies with no sweep params.
func (r *StrategyRegistry) SweepParamNames(name string) []string {
	entry, ok := r.m[name]
	if !ok || len(entry.SweepParams) == 0 {
		return nil
	}
	names := make([]string, len(entry.SweepParams))
	for i, sp := range entry.SweepParams {
		names[i] = sp.Name
	}
	return names
}

// RegisterFlags registers all strategy-specific parameters as float64 flags on fs.
// Params shared across strategies are registered once (first occurrence wins; all
// strategies sharing a param name must agree on its default and usage).
// Returns a map from param name to flag pointer; call BuildParamMap after fs.Parse.
func (r *StrategyRegistry) RegisterFlags(fs *flag.FlagSet) map[string]*float64 {
	ptrs := make(map[string]*float64)
	// Iterate in sorted name order for deterministic flag registration.
	for _, name := range r.ListStrategies() {
		entry := r.m[name]
		for _, pd := range entry.Params {
			if _, exists := ptrs[pd.Name]; !exists {
				v := fs.Float64(pd.Name, pd.Default, pd.Usage)
				ptrs[pd.Name] = v
			}
		}
	}
	return ptrs
}

// BuildParamMap dereferences the flag pointers returned by RegisterFlags into a plain
// map. Call after flag parsing.
func BuildParamMap(ptrs map[string]*float64) map[string]float64 {
	m := make(map[string]float64, len(ptrs))
	for k, v := range ptrs {
		m[k] = *v
	}
	return m
}

// DefaultParams returns a param map for name populated with all registered defaults.
// Returns nil for unknown names. Primarily used in tests.
func (r *StrategyRegistry) DefaultParams(name string) map[string]float64 {
	entry, ok := r.m[name]
	if !ok {
		return nil
	}
	p := make(map[string]float64, len(entry.Params))
	for _, pd := range entry.Params {
		p[pd.Name] = pd.Default
	}
	return p
}

// ParamsMap returns strategy-specific params formatted as strings for run metadata.
// Only params registered for name are included. Unknown names return nil.
func (r *StrategyRegistry) ParamsMap(name string, p map[string]float64) map[string]string {
	entry, ok := r.m[name]
	if !ok {
		return nil
	}
	result := make(map[string]string, len(entry.Params))
	for _, pd := range entry.Params {
		v, exists := p[pd.Name]
		if !exists {
			v = pd.Default
		}
		if pd.IsInt {
			result[pd.Name] = fmt.Sprintf("%d", int(math.Round(v)))
		} else {
			result[pd.Name] = fmt.Sprintf("%.4g", v)
		}
	}
	return result
}
