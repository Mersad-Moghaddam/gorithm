package binarysearch

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps of binary search on the given array and target.
func Trace(initial []int, target int) []trace.Step {
	b := trace.NewBuilder()
	arr := cloneSlice(initial)
	n := len(arr)

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Binary Search for target %d in array of %d elements.", target, n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"target": target, "n": n})

	if n == 0 {
		b.AddStep("L2", trace.ActionSet, "Array is empty; search bounds low = 0, high = -1.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"low": 0, "high": -1, "target": target})

		b.AddStep("L3", trace.ActionCompare, "Check condition low (0) <= high (-1): false. Search space is empty.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"low": 0, "high": -1, "target": target})

		b.AddStep("L13", trace.ActionDone, fmt.Sprintf("Target %d not found in array. Return -1.", target), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"low": 0, "high": -1, "target": target, "found": false, "result": -1})

		return b.Steps()
	}

	low := 0
	high := n - 1

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Set search bounds: low = 0, high = %d.", high), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"low": low, "high": high, "target": target})

	for low <= high {
		b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("Check condition low (%d) <= high (%d): true. Search range [%d, %d] has %d element(s).", low, high, low, high, high-low+1), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"low": low, "high": high, "target": target})

		mid := low + (high-low)/2
		b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("Calculate middle index mid = %d + (%d - %d)/2 = %d (value = %d).", low, high, low, mid, arr[mid]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: mid,
		}, map[string]any{"low": low, "high": high, "mid": mid, "target": target, "midValue": arr[mid]})

		b.IncComparisons()
		b.AddStep("L5", trace.ActionCompare, fmt.Sprintf("Compare arr[%d] (%d) with target (%d).", mid, arr[mid], target), renderers.ArrayState{
			Values:    cloneSlice(arr),
			Comparing: []int{mid},
			Current:   mid,
		}, map[string]any{"low": low, "high": high, "mid": mid, "target": target, "midValue": arr[mid]})

		if arr[mid] == target {
			b.AddStep("L6", trace.ActionDone, fmt.Sprintf("Target %d found at index %d! Return %d.", target, mid, mid), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  []int{mid},
				Current: mid,
			}, map[string]any{"low": low, "high": high, "mid": mid, "target": target, "found": true, "result": mid})
			return b.Steps()
		}

		if arr[mid] < target {
			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("arr[%d] (%d) < target (%d): target is in the right subarray.", mid, arr[mid], target), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: mid,
			}, map[string]any{"low": low, "high": high, "mid": mid, "target": target})

			low = mid + 1
			b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Update low = mid + 1 = %d. Discard left half.", low), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: -1,
			}, map[string]any{"low": low, "high": high, "target": target})
		} else {
			b.AddStep("L9", trace.ActionHighlight, fmt.Sprintf("arr[%d] (%d) > target (%d): target is in the left subarray.", mid, arr[mid], target), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: mid,
			}, map[string]any{"low": low, "high": high, "mid": mid, "target": target})

			high = mid - 1
			b.AddStep("L10", trace.ActionSet, fmt.Sprintf("Update high = mid - 1 = %d. Discard right half.", high), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: -1,
			}, map[string]any{"low": low, "high": high, "target": target})
		}
	}

	b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("Check condition low (%d) <= high (%d): false. Search range exhausted.", low, high), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"low": low, "high": high, "target": target})

	b.AddStep("L13", trace.ActionDone, fmt.Sprintf("Target %d not found in array. Return -1.", target), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"low": low, "high": high, "target": target, "found": false, "result": -1})

	return b.Steps()
}

func cloneSlice(s []int) []int {
	if s == nil {
		return []int{}
	}
	c := make([]int, len(s))
	copy(c, s)
	return c
}
