package recommendation

import "strings"

type Registry struct {
	strategies map[string]Strategy
	fallback   Strategy
}

func NewRegistry(all []Strategy, fallbackName string) *Registry {
	m := make(map[string]Strategy, len(all))
	var fallback Strategy
	for _, s := range all {
		key := normalizeStrategyName(s.Name())
		m[key] = s
		if key == normalizeStrategyName(fallbackName) {
			fallback = s
		}
	}
	if fallback == nil {
		fallback = m[StrategyRule]
	}
	return &Registry{strategies: m, fallback: fallback}
}

func (r *Registry) Get(name string) Strategy {
	if r == nil {
		return nil
	}
	if s, ok := r.strategies[normalizeStrategyName(name)]; ok {
		return s
	}
	return r.fallback
}

func normalizeStrategyName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case StrategyRule, StrategyWeighted, StrategyCF:
		return name
	default:
		return StrategyRule
	}
}

func NormalizeStrategyName(name string) string {
	return normalizeStrategyName(name)
}
