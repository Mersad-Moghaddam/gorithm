package linearsearch

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace records the first match without changing input or earlier snapshots.
func Trace(initial []int, target int) []trace.Step {
	b := trace.NewBuilder()
	s := renderers.LinearSearchLessonState{
		ArrayState: renderers.ArrayState{Values: cloneSlice(initial), Current: -1},
		Phase:      "init", Target: target, Index: -1, CheckedThrough: -1, Result: -1,
	}
	add := func(line string, action trace.ActionType, explanation string, variables map[string]any) {
		snapshot := s
		snapshot.Values = cloneSlice(s.Values)
		snapshot.Comparing = cloneSlice(s.Comparing)
		snapshot.Sorted = cloneSlice(s.Sorted)
		b.AddStep(line, action, explanation, snapshot, variables)
	}
	add("L1", trace.ActionInit, fmt.Sprintf("Find the first %d. Check values from left to right.", target), map[string]any{"target": target, "n": len(initial)})
	for i, value := range initial {
		s.Phase, s.Index, s.Current, s.Compared, s.Equal = "scan", i, i, false, false
		s.Comparing = nil
		add("L2", trace.ActionHighlight, fmt.Sprintf("Move i to index %d. Compare %d with %d next.", i, value, target), map[string]any{"i": i, "target": target, "val": value})
		b.IncComparisons()
		s.Phase, s.Compared, s.Equal, s.Comparing = "compare", true, value == target, []int{i}
		if !s.Equal {
			s.CheckedThrough = i
		}
		add("L3", trace.ActionCompare, fmt.Sprintf("Compare %d with %d: equal = %t.", value, target, s.Equal), map[string]any{"i": i, "target": target, "val": value, "match": s.Equal})
		if s.Equal {
			s.Phase, s.Result, s.ResultKnown, s.Comparing, s.Sorted = "done", i, true, nil, []int{i}
			add("L4", trace.ActionDone, fmt.Sprintf("Found %d at index %d. Stop at the first match.", target, i), map[string]any{"i": i, "target": target, "result": i})
			steps := b.Steps()
			steps[len(steps)-2].HoldMS, steps[len(steps)-1].HoldMS = 600, 1200
			return steps
		}
	}
	s.Phase, s.Index, s.Current, s.Compared, s.Equal, s.ResultKnown, s.Comparing = "done", -1, -1, false, false, true, nil
	add("L7", trace.ActionDone, fmt.Sprintf("All %d values checked. Target %d is absent; return -1.", len(initial), target), map[string]any{"target": target, "result": -1})
	return b.Steps()
}

func cloneSlice(s []int) []int {
	if s == nil {
		return nil
	}
	c := make([]int, len(s))
	copy(c, s)
	return c
}
