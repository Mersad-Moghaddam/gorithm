package binarysearch

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps of binary search on the given array and target.
func Trace(initial []int, target int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	if arr == nil {
		arr = []int{}
	}
	n := len(arr)

	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("binarysearch-%d", i)
	}

	snapshot := func(caption string, current int, comparing []int, sorted []int) renderers.ArrayState {
		return renderers.ArrayState{
			IDs:       slices.Clone(ids),
			Caption:   caption,
			Values:    slices.Clone(arr),
			Comparing: slices.Clone(comparing),
			Sorted:    slices.Clone(sorted),
			Current:   current,
		}
	}

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Binary Search for target %d in array of 0 elements.", target),
			snapshot(fmt.Sprintf("Target: %d | Empty array", target), -1, nil, nil),
			map[string]any{"target": target, "n": 0, "low": 0, "high": -1})

		b.AddStep("L2", trace.ActionSet, "Array is empty; search bounds low = 0, high = -1.",
			snapshot(fmt.Sprintf("Target: %d | Window: empty [0..-1]", target), -1, nil, nil),
			map[string]any{"target": target, "low": 0, "high": -1})

		b.AddStep("L3", trace.ActionCompare, "Check condition low (0) <= high (-1): false. Search space is empty.",
			snapshot(fmt.Sprintf("Target: %d | Window empty: low (0) > high (-1)", target), -1, nil, nil),
			map[string]any{"target": target, "low": 0, "high": -1})

		b.AddStep("L13", trace.ActionDone, fmt.Sprintf("Target %d not found in array. low (0) > high (-1), search space empty. Return -1.", target),
			snapshot(fmt.Sprintf("Target %d not found | Return -1", target), -1, nil, nil),
			map[string]any{"target": target, "low": 0, "high": -1, "found": false, "result": -1})

		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Binary Search for target %d in sorted array of %d elements.", target, n),
		snapshot(fmt.Sprintf("Target: %d | Sorted array of %d elements", target, n), -1, nil, nil),
		map[string]any{"target": target, "n": n, "low": 0, "high": n - 1})

	low := 0
	high := n - 1

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Set search bounds: low = 0, high = %d.", high),
		snapshot(fmt.Sprintf("Target: %d | Active window [%d..%d]", target, low, high), -1, nil, nil),
		map[string]any{"target": target, "low": low, "high": high})

	for low <= high {
		b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("Check condition low (%d) <= high (%d): true. Active window [%d..%d] has %d element(s).", low, high, low, high, high-low+1),
			snapshot(fmt.Sprintf("Target: %d | Active window [%d..%d]", target, low, high), -1, nil, nil),
			map[string]any{"target": target, "low": low, "high": high})

		mid := low + (high-low)/2
		b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("Calculate middle index: mid = %d + (%d - %d)/2 = %d. Middle element arr[%d] is %d.", low, high, low, mid, mid, arr[mid]),
			snapshot(fmt.Sprintf("Target: %d | Window [%d..%d] | mid = %d + (%d-%d)/2 = %d (arr[%d]=%d)", target, low, high, low, high, low, mid, mid, arr[mid]), mid, nil, nil),
			map[string]any{"target": target, "low": low, "high": high, "mid": mid, "midValue": arr[mid]})

		b.IncComparisons()
		b.AddStep("L5", trace.ActionCompare, fmt.Sprintf("Compare middle element arr[%d] (%d) with target (%d).", mid, arr[mid], target),
			snapshot(fmt.Sprintf("Target: %d | Compare arr[%d]=%d vs target=%d", target, mid, arr[mid], target), mid, []int{mid}, nil),
			map[string]any{"target": target, "low": low, "high": high, "mid": mid, "midValue": arr[mid]})

		if arr[mid] == target {
			b.AddStep("L6", trace.ActionDone, fmt.Sprintf("arr[%d] == target (%d == %d). Target %d found at index %d! Return %d.", mid, arr[mid], target, target, mid, mid),
				snapshot(fmt.Sprintf("Target %d found at index %d!", target, mid), mid, nil, []int{mid}),
				map[string]any{"target": target, "low": low, "high": high, "mid": mid, "midValue": arr[mid], "found": true, "result": mid})
			return b.Steps()
		}

		if arr[mid] < target {
			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("arr[%d]=%d < target=%d, so target cannot be in [%d..%d]; move low to mid+1", mid, arr[mid], target, low, mid),
				snapshot(fmt.Sprintf("arr[%d]=%d < target=%d: eliminate [%d..%d]; move low to %d", mid, arr[mid], target, low, mid, mid+1), mid, nil, nil),
				map[string]any{"target": target, "low": low, "high": high, "mid": mid, "midValue": arr[mid]})

			oldLow := low
			low = mid + 1
			b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Update low = mid + 1 = %d. Discard left half [%d..%d]. Active window is now [%d..%d].", low, oldLow, mid, low, high),
				snapshot(fmt.Sprintf("Target: %d | Active window updated to [%d..%d]", target, low, high), -1, nil, nil),
				map[string]any{"target": target, "low": low, "high": high})
		} else {
			b.AddStep("L9", trace.ActionHighlight, fmt.Sprintf("arr[%d]=%d > target=%d, so target cannot be in [%d..%d]; move high to mid-1", mid, arr[mid], target, mid, high),
				snapshot(fmt.Sprintf("arr[%d]=%d > target=%d: eliminate [%d..%d]; move high to %d", mid, arr[mid], target, mid, high, mid-1), mid, nil, nil),
				map[string]any{"target": target, "low": low, "high": high, "mid": mid, "midValue": arr[mid]})

			oldHigh := high
			high = mid - 1
			b.AddStep("L10", trace.ActionSet, fmt.Sprintf("Update high = mid - 1 = %d. Discard right half [%d..%d]. Active window is now [%d..%d].", high, mid, oldHigh, low, high),
				snapshot(fmt.Sprintf("Target: %d | Active window updated to [%d..%d]", target, low, high), -1, nil, nil),
				map[string]any{"target": target, "low": low, "high": high})
		}
	}

	b.AddStep("L3", trace.ActionCompare, fmt.Sprintf("Check condition low (%d) <= high (%d): false. Active window is empty (low > high).", low, high),
		snapshot(fmt.Sprintf("Target: %d | Search space empty (low=%d > high=%d)", target, low, high), -1, nil, nil),
		map[string]any{"target": target, "low": low, "high": high})

	b.AddStep("L13", trace.ActionDone, fmt.Sprintf("Target %d not found in array. low (%d) > high (%d), search space empty. Return -1.", target, low, high),
		snapshot(fmt.Sprintf("Target %d not found | Return -1", target), -1, nil, nil),
		map[string]any{"target": target, "low": low, "high": high, "found": false, "result": -1})

	return b.Steps()
}
