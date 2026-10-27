package quick

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace executes Quick Sort on initial slice, recording each step and metrics for visualization.
// ponytail: Lomuto partition trace recorded; upgrade to Hoare or randomized pivot if visual demonstration needs alternative schemes.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("quick-%d", i)
	}

	settled := make([]bool, n)
	sortedSlice := func() []int {
		res := make([]int, 0, n)
		for i, isSet := range settled {
			if isSet {
				res = append(res, i)
			}
		}
		return res
	}

	snapshot := func(caption string, pivotIdx int) renderers.ArrayState {
		return renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  sortedSlice(),
			Current: pivotIdx,
			Caption: caption,
		}
	}

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, "Initialize Quick Sort: array is empty.", snapshot("Empty array: nothing to sort", -1), map[string]any{"n": 0})
		b.AddStep("L7", trace.ActionDone, "Quick Sort completed! Array is empty.", snapshot("Sorted: empty array", -1), map[string]any{"status": "complete"})
		return b.Steps()
	}

	if n == 1 {
		settled[0] = true
		b.AddStep("L1", trace.ActionInit, "Initialize Quick Sort: single-element array.", snapshot("Single element: already sorted", -1), map[string]any{"n": 1})
		b.AddStep("L7", trace.ActionDone, "Quick Sort completed! Single element is sorted.", snapshot("Sorted: single element in final position", -1), map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Quick Sort: divide array around pivots recursively.", snapshot("Quick Sort: partition around pivot; smaller elements move left, larger move right", -1), map[string]any{"n": n})

	var sortHelper func(low, high int)
	var partitionHelper func(low, high int) int

	sortHelper = func(low, high int) {
		if low >= high {
			if low == high && low >= 0 && low < n {
				settled[low] = true
				caption := fmt.Sprintf("Range [%d..%d] has 1 element: arr[%d]=%d is settled", low, high, low, arr[low])
				b.AddStep("L2", trace.ActionHighlight, caption, snapshot(caption, -1), map[string]any{"low": low, "high": high, "index": low})
			}
			return
		}

		caption := fmt.Sprintf("Active partition range [%d..%d]: condition low < high holds", low, high)
		b.AddStep("L2", trace.ActionHighlight, caption, snapshot(caption, -1), map[string]any{"low": low, "high": high})

		caption = fmt.Sprintf("Partition range [%d..%d] around pivot arr[%d]=%d", low, high, high, arr[high])
		b.AddStep("L3", trace.ActionHighlight, caption, snapshot(caption, high), map[string]any{"low": low, "high": high, "pivot": arr[high], "pivotIndex": high})

		p := partitionHelper(low, high)

		caption = fmt.Sprintf("Recursively sort left partition [%d..%d] (elements < pivot %d)", low, p-1, arr[p])
		b.AddStep("L4", trace.ActionHighlight, caption, snapshot(caption, -1), map[string]any{"low": low, "high": p - 1, "p": p})
		sortHelper(low, p-1)

		caption = fmt.Sprintf("Recursively sort right partition [%d..%d] (elements >= pivot %d)", p+1, high, arr[p])
		b.AddStep("L5", trace.ActionHighlight, caption, snapshot(caption, -1), map[string]any{"low": p + 1, "high": high, "p": p})
		sortHelper(p+1, high)
	}

	partitionHelper = func(low, high int) int {
		caption := fmt.Sprintf("Begin partition on range [%d..%d]", low, high)
		b.AddStep("L8", trace.ActionHighlight, caption, snapshot(caption, high), map[string]any{"low": low, "high": high})

		pivot := arr[high]
		caption = fmt.Sprintf("Select pivot arr[%d]=%d (purple); smaller elements will move left of pivot", high, pivot)
		b.AddStep("L9", trace.ActionSet, caption, snapshot(caption, high), map[string]any{"low": low, "high": high, "pivot": pivot, "pivotIndex": high})

		i := low - 1
		caption = fmt.Sprintf("Initialize boundary i=%d; elements in [%d..i] will be < pivot %d", i, low, pivot)
		b.AddStep("L10", trace.ActionSet, caption, snapshot(caption, high), map[string]any{"low": low, "high": high, "i": i, "pivot": pivot, "pivotIndex": high})

		for j := low; j < high; j++ {
			caption = fmt.Sprintf("Partition [%d..%d]: scan j=%d (val=%d), boundary i=%d, pivot arr[%d]=%d", low, high, j, arr[j], i, high, pivot)
			state := snapshot(caption, high)
			state.Comparing = []int{j}
			b.AddStep("L11", trace.ActionHighlight, caption, state, map[string]any{"low": low, "high": high, "i": i, "j": j, "pivot": pivot, "pivotIndex": high})

			b.IncComparisons()
			if arr[j] < pivot {
				caption = fmt.Sprintf("Compare arr[j=%d]=%d < pivot %d: true; element belongs left of pivot", j, arr[j], pivot)
				state = snapshot(caption, high)
				state.Comparing = []int{j}
				b.AddStep("L12", trace.ActionCompare, caption, state, map[string]any{"low": low, "high": high, "i": i, "j": j, "pivot": pivot, "pivotIndex": high, "val": arr[j]})

				i++
				caption = fmt.Sprintf("Increment boundary pointer i to %d to expand smaller-than-pivot partition", i)
				b.AddStep("L13", trace.ActionSet, caption, snapshot(caption, high), map[string]any{"low": low, "high": high, "i": i, "j": j, "pivot": pivot, "pivotIndex": high})

				arr[i], arr[j] = arr[j], arr[i]
				ids[i], ids[j] = ids[j], ids[i]
				if i != j {
					b.IncSwaps()
					caption = fmt.Sprintf("Swap arr[%d]=%d and arr[%d]=%d: smaller element %d moves left of pivot into partition [%d..%d]", i, arr[i], j, arr[j], arr[i], low, i)
				} else {
					caption = fmt.Sprintf("arr[%d]=%d < pivot %d is already in left partition (i==j==%d); smaller element stays left of pivot", i, arr[i], pivot, i)
				}
				state = snapshot(caption, high)
				state.Swapped = []int{i, j}
				b.AddStep("L14", trace.ActionSwap, caption, state, map[string]any{"low": low, "high": high, "i": i, "j": j, "pivot": pivot, "pivotIndex": high})
			} else {
				caption = fmt.Sprintf("Compare arr[j=%d]=%d < pivot %d: false; %d >= pivot stays right of boundary i=%d", j, arr[j], pivot, arr[j], i)
				state = snapshot(caption, high)
				state.Comparing = []int{j}
				b.AddStep("L12", trace.ActionCompare, caption, state, map[string]any{"low": low, "high": high, "i": i, "j": j, "pivot": pivot, "pivotIndex": high, "val": arr[j]})
			}
		}

		p := i + 1
		arr[p], arr[high] = arr[high], arr[p]
		ids[p], ids[high] = ids[high], ids[p]
		if p != high {
			b.IncSwaps()
		}
		settled[p] = true

		if p != high {
			caption = fmt.Sprintf("Swap pivot %d into final position arr[%d] with arr[%d]=%d; pivot settled (left <= pivot <= right)", arr[p], p, high, arr[high])
		} else {
			caption = fmt.Sprintf("Pivot %d already at final position arr[%d]; pivot settled (left <= pivot <= right)", arr[p], p)
		}
		state := snapshot(caption, p)
		state.Swapped = []int{p, high}
		b.AddStep("L17", trace.ActionSwap, caption, state, map[string]any{"low": low, "high": high, "p": p, "pivot": arr[p], "pivotIndex": p})

		caption = fmt.Sprintf("Partition complete: pivot %d settled at [%d]; returns partition index %d", arr[p], p, p)
		b.AddStep("L18", trace.ActionSet, caption, snapshot(caption, p), map[string]any{"low": low, "high": high, "p": p, "pivotIndex": p})

		return p
	}

	sortHelper(0, n-1)

	for i := 0; i < n; i++ {
		settled[i] = true
	}

	b.AddStep("L7", trace.ActionDone, "Quick Sort completed! All elements sorted.", snapshot("Sorted: all elements in final positions", -1), map[string]any{"status": "complete"})

	return b.Steps()
}
