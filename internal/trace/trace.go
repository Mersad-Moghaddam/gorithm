package trace

type Builder struct {
	steps   []Step
	metrics Metrics
}

func NewBuilder() *Builder {
	return &Builder{
		steps: make([]Step, 0, 64),
	}
}

func (b *Builder) CurrentMetrics() Metrics {
	return b.metrics
}

func (b *Builder) IncComparisons() {
	b.metrics.Comparisons++
	b.metrics.Operations++
}

func (b *Builder) IncSwaps() {
	b.metrics.Swaps++
	b.metrics.Operations++
}

func (b *Builder) IncOperations() {
	b.metrics.Operations++
}

func (b *Builder) AddStep(lineID string, action ActionType, explanation string, state any, vars map[string]any) {
	clonedVars := make(map[string]any, len(vars))
	for k, v := range vars {
		clonedVars[k] = v
	}
	step := Step{
		Index:       len(b.steps),
		LineID:      lineID,
		Action:      action,
		Explanation: explanation,
		State:       state,
		Variables:   clonedVars,
		Metrics:     b.metrics,
	}
	b.steps = append(b.steps, step)
}

func (b *Builder) Steps() []Step {
	return b.steps
}
