package quick

import (
	"fmt"
	"sort"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace executes Quick Sort on initial slice, recording each step and metrics for visualization.
// ponytail: Lomuto partition trace recorded; upgrade to Hoare or randomized pivot if visual demonstration needs alternative schemes.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)

	b.AddStep("L1", trace.ActionInit, "Initialize Quick Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n <= 1 {
		sorted := make([]int, n)
		for i := 0; i < n; i++ {
			sorted[i] = i
		}
		b.AddStep("L7", trace.ActionDone, "Quick Sort completed! All elements sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  sorted,
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	sortedSet := make(map[int]bool)

	getSortedSlice := func() []int {
		res := make([]int, 0, len(sortedSet))
		for idx := range sortedSet {
			res = append(res, idx)
		}
		sort.Ints(res)
		return res
	}

	var sortHelper func(low, high int)
	var partitionHelper func(low, high int) int

	sortHelper = func(low, high int) {
		b.AddStep("L2", trace.ActionHighlight, fmt.Sprintf("Checking base case for range [%d..%d].", low, high), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: -1,
		}, map[string]any{"low": low, "high": high})

		if low < high {
			b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Partitioning range [%d..%d].", low, high), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: high,
			}, map[string]any{"low": low, "high": high})

			p := partitionHelper(low, high)

			b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("Recursively sorting left partition [%d..%d].", low, p-1), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: -1,
			}, map[string]any{"low": low, "p": p})
			sortHelper(low, p-1)

			b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Recursively sorting right partition [%d..%d].", p+1, high), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: -1,
			}, map[string]any{"high": high, "p": p})
			sortHelper(p+1, high)
		} else if low == high {
			sortedSet[low] = true
			b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Single element at index %d is sorted.", low), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: low,
			}, map[string]any{"index": low})
		}
	}

	partitionHelper = func(low, high int) int {
		b.AddStep("L8", trace.ActionHighlight, fmt.Sprintf("Starting partition on range [%d..%d].", low, high), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: high,
		}, map[string]any{"low": low, "high": high})

		pivot := arr[high]
		b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Selected pivot arr[%d] = %d.", high, pivot), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: high,
		}, map[string]any{"pivot": pivot, "pivotIndex": high})

		i := low - 1
		b.AddStep("L10", trace.ActionSet, fmt.Sprintf("Initialized boundary i = %d.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: high,
		}, map[string]any{"i": i})

		for j := low; j < high; j++ {
			b.AddStep("L11", trace.ActionHighlight, fmt.Sprintf("Examining arr[%d] = %d.", j, arr[j]), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Sorted:    getSortedSlice(),
				Comparing: []int{j, high},
				Current:   j,
			}, map[string]any{"j": j, "pivot": pivot})

			b.IncComparisons()
			b.AddStep("L12", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) < pivot (%d).", j, arr[j], pivot), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Sorted:    getSortedSlice(),
				Comparing: []int{j, high},
				Current:   j,
			}, map[string]any{"val": arr[j], "pivot": pivot})

			if arr[j] < pivot {
				i++
				b.AddStep("L13", trace.ActionSet, fmt.Sprintf("Incremented boundary i = %d.", i), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  getSortedSlice(),
					Current: i,
				}, map[string]any{"i": i})

				arr[i], arr[j] = arr[j], arr[i]
				b.IncSwaps()
				b.AddStep("L14", trace.ActionSwap, fmt.Sprintf("Swapped arr[%d] and arr[%d].", i, j), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  getSortedSlice(),
					Swapped: []int{i, j},
					Current: i,
				}, map[string]any{"i": i, "j": j})
			}
		}

		arr[i+1], arr[high] = arr[high], arr[i+1]
		b.IncSwaps()
		p := i + 1
		sortedSet[p] = true
		b.AddStep("L17", trace.ActionSwap, fmt.Sprintf("Placed pivot into sorted position at index %d.", p), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Swapped: []int{p, high},
			Current: p,
		}, map[string]any{"pivotIndex": p})

		b.AddStep("L18", trace.ActionSet, fmt.Sprintf("Partition complete. Pivot index is %d.", p), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: p,
		}, map[string]any{"pivotIndex": p})

		return p
	}

	sortHelper(0, n-1)

	for i := 0; i < n; i++ {
		sortedSet[i] = true
	}

	b.AddStep("L7", trace.ActionDone, "Quick Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  getSortedSlice(),
		Current: -1,
	}, map[string]any{"status": "complete"})

	return b.Steps()
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}
