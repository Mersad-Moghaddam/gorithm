package binarysearch

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace preserves Search's midpoint semantics, including for unsorted direct input.
func Trace(initial []int, target int) []trace.Step {
	b := trace.NewBuilder()
	arr := cloneSlice(initial)
	s := renderers.BinarySearchLessonState{ArrayState: renderers.ArrayState{Current: -1}, Target: target, Low: 0, High: len(arr) - 1, Mid: -1, Result: -1, PreviousLow: 0, PreviousHigh: len(arr) - 1, DiscardLow: -1, DiscardHigh: -1, SortedInput: slices.IsSorted(arr)}
	add := func(line, phase string, action trace.ActionType, text string) {
		s.Phase = phase
		s.Values = cloneSlice(arr)
		s.Current = s.Mid
		s.Comparing = nil
		s.Sorted = nil
		if phase == "compare" && s.Mid >= 0 {
			s.Comparing = []int{s.Mid}
		}
		if s.ResultKnown && s.Result >= 0 {
			s.Sorted = []int{s.Result}
		}
		vars := map[string]any{"target": target, "n": len(arr), "low": s.Low, "high": s.High}
		if s.Mid >= 0 {
			vars["mid"] = s.Mid
			vars["midValue"] = arr[s.Mid]
		}
		if s.ResultKnown {
			vars["found"] = s.Result >= 0
			vars["result"] = s.Result
		}
		b.AddStep(line, action, text, s, vars)
	}
	finish := func() []trace.Step {
		steps := b.Steps()
		for i := range steps {
			if steps[i].LineID == "L1" || steps[i].LineID == "L3" {
				steps[i].HoldMS = 350
			}
			if i+1 < len(steps) && ((steps[i].LineID == "L7" && steps[i+1].LineID == "L8") || (steps[i].LineID == "L9" && steps[i+1].LineID == "L10") || (steps[i].LineID == "L5" && steps[i+1].LineID == "L6")) {
				steps[i].HoldMS = 600
				steps[i+1].HoldMS = 1200
			}
		}
		return steps
	}
	add("L1", "init", trace.ActionInit, fmt.Sprintf("Search for %d in %d values. Input must be sorted.", target, len(arr)))
	add("L2", "bounds", trace.ActionSet, fmt.Sprintf("Set inclusive bounds [%d, %d].", s.Low, s.High))
	for s.Low <= s.High {
		add("L3", "bounds", trace.ActionCompare, fmt.Sprintf("Bounds [%d, %d] are nonempty.", s.Low, s.High))
		s.Mid = s.Low + (s.High-s.Low)/2
		add("L4", "midpoint", trace.ActionHighlight, fmt.Sprintf("mid = %d + (%d - %d)/2 = %d; value %d.", s.Low, s.High, s.Low, s.Mid, arr[s.Mid]))
		b.IncComparisons()
		add("L5", "compare", trace.ActionCompare, fmt.Sprintf("Compare value %d with target %d.", arr[s.Mid], target))
		if arr[s.Mid] == target {
			s.ResultKnown = true
			s.Result = s.Mid
			add("L6", "done", trace.ActionDone, fmt.Sprintf("Target %d found at index %d.", target, s.Result))
			return finish()
		}
		s.PreviousLow = s.Low
		s.PreviousHigh = s.High
		if arr[s.Mid] < target {
			add("L7", "compare", trace.ActionHighlight, fmt.Sprintf("%d < %d. Move low past midpoint; exclusion requires sorted input.", arr[s.Mid], target))
			s.DiscardLow = s.Low
			s.DiscardHigh = s.Mid
			s.Low = s.Mid + 1
			s.Mid = -1
			add("L8", "discard", trace.ActionSet, fmt.Sprintf("low = %d; remove interval [%d, %d] from search.", s.Low, s.DiscardLow, s.DiscardHigh))
		} else {
			add("L9", "compare", trace.ActionHighlight, fmt.Sprintf("%d > %d. Move high before midpoint; exclusion requires sorted input.", arr[s.Mid], target))
			s.DiscardLow = s.Mid
			s.DiscardHigh = s.High
			s.High = s.Mid - 1
			s.Mid = -1
			add("L10", "discard", trace.ActionSet, fmt.Sprintf("high = %d; remove interval [%d, %d] from search.", s.High, s.DiscardLow, s.DiscardHigh))
		}
		s.DiscardLow = -1
		s.DiscardHigh = -1
	}
	add("L3", "bounds", trace.ActionCompare, "low > high: empty interval.")
	s.ResultKnown = true
	add("L13", "done", trace.ActionDone, fmt.Sprintf("Search returns -1 for target %d.", target))
	return finish()
}

func cloneSlice(s []int) []int { c := make([]int, len(s)); copy(c, s); return c }
