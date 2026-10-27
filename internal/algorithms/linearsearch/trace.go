package linearsearch

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for linear search on initial slice with given target.
func Trace(initial []int, target int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Linear Search for target %d in array of %d elements.", target, n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"target": target, "n": n})

	for i := 0; i < n; i++ {
		b.AddStep("L2", trace.ActionHighlight, fmt.Sprintf("Scanning index %d (value %d).", i, arr[i]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
		}, map[string]any{"i": i, "target": target, "val": arr[i]})

		b.IncComparisons()
		if arr[i] == target {
			b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) == %d: match found!", i, arr[i], target), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{i},
				Current:   i,
			}, map[string]any{"i": i, "target": target, "val": arr[i], "match": true})

			b.AddStep("L4", trace.ActionDone, fmt.Sprintf("Target %d found at index %d.", target, i), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
				Sorted:  []int{i},
			}, map[string]any{"i": i, "target": target, "result": i})
			return b.Steps()
		}

		b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) == %d: not equal.", i, arr[i], target), renderers.ArrayState{
			Values:    cloneSlice(arr),
			Comparing: []int{i},
			Current:   i,
		}, map[string]any{"i": i, "target": target, "val": arr[i], "match": false})
	}

	b.AddStep("L7", trace.ActionDone, fmt.Sprintf("Target %d not found in array. Return -1.", target), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"target": target, "result": -1})

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
