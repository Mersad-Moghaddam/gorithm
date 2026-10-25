package trace

import (
	"testing"
)

func TestTraceBuilder(t *testing.T) {
	b := NewBuilder()
	b.IncComparisons()
	b.AddStep("L1", ActionCompare, "Comparing elements", []int{1, 2}, map[string]any{"i": 0})
	b.IncSwaps()
	b.AddStep("L2", ActionSwap, "Swapping elements", []int{2, 1}, map[string]any{"i": 0, "j": 1})

	steps := b.Steps()
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	if steps[0].Metrics.Comparisons != 1 || steps[0].Metrics.Swaps != 0 {
		t.Fatalf("step 0 metrics incorrect: %+v", steps[0].Metrics)
	}
	if steps[1].Metrics.Swaps != 1 || steps[1].Metrics.Comparisons != 1 {
		t.Fatalf("step 1 metrics incorrect: %+v", steps[1].Metrics)
	}
}
