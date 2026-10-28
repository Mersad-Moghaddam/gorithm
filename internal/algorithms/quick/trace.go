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
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i
	}
	context := renderers.QuickLessonState{Low: 0, High: n - 1, Boundary: -1, Scan: -1, PivotIndex: -1, PivotID: -1}
	add := func(line string, action trace.ActionType, explanation string, state renderers.ArrayState, variables map[string]any) {
		snapshot := context
		snapshot.ArrayState = state
		snapshot.IDs = cloneSlice(ids)
		snapshot.Stack = append([]renderers.QuickLessonRange(nil), context.Stack...)
		b.AddStep(line, action, explanation, snapshot, variables)
	}

	add("L1", trace.ActionInit, "Initialize Quick Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n <= 1 {
		sorted := make([]int, n)
		for i := 0; i < n; i++ {
			sorted[i] = i
		}
		add("L7", trace.ActionDone, "Quick Sort completed! All elements sorted.", renderers.ArrayState{
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
		parent := context
		context.Low, context.High = low, high
		context.Boundary, context.Scan, context.PivotIndex, context.PivotID = -1, -1, -1, -1
		context.PivotValue, context.PivotActive = 0, false
		context.Stack = append(append([]renderers.QuickLessonRange(nil), parent.Stack...), renderers.QuickLessonRange{Low: low, High: high})
		defer func() { context = parent }()
		add("L2", trace.ActionHighlight, fmt.Sprintf("Checking base case for range [%d..%d].", low, high), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: -1,
		}, map[string]any{"low": low, "high": high})

		if low < high {
			add("L3", trace.ActionHighlight, fmt.Sprintf("Partitioning range [%d..%d].", low, high), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: high,
			}, map[string]any{"low": low, "high": high})

			p := partitionHelper(low, high)

			add("L4", trace.ActionHighlight, fmt.Sprintf("Recursively sorting left partition [%d..%d].", low, p-1), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: -1,
			}, map[string]any{"low": low, "p": p})
			sortHelper(low, p-1)

			add("L5", trace.ActionHighlight, fmt.Sprintf("Recursively sorting right partition [%d..%d].", p+1, high), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: -1,
			}, map[string]any{"high": high, "p": p})
			sortHelper(p+1, high)
		} else if low == high {
			sortedSet[low] = true
			add("L6", trace.ActionSet, fmt.Sprintf("Single element at index %d is sorted.", low), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  getSortedSlice(),
				Current: low,
			}, map[string]any{"index": low})
		}
	}

	partitionHelper = func(low, high int) int {
		add("L8", trace.ActionHighlight, fmt.Sprintf("Starting partition on range [%d..%d].", low, high), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: high,
		}, map[string]any{"low": low, "high": high})

		pivot := arr[high]
		context.PivotIndex, context.PivotID, context.PivotValue, context.PivotActive = high, ids[high], pivot, true
		add("L9", trace.ActionSet, fmt.Sprintf("Selected pivot arr[%d] = %d.", high, pivot), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: high,
		}, map[string]any{"pivot": pivot, "pivotIndex": high})

		i := low - 1
		context.Boundary = i
		add("L10", trace.ActionSet, fmt.Sprintf("Initialized boundary i = %d.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Current: high,
		}, map[string]any{"i": i})

		for j := low; j < high; j++ {
			context.Scan = j
			add("L11", trace.ActionHighlight, fmt.Sprintf("Examining arr[%d] = %d.", j, arr[j]), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Sorted:    getSortedSlice(),
				Comparing: []int{j, high},
				Current:   j,
			}, map[string]any{"j": j, "pivot": pivot})

			b.IncComparisons()
			add("L12", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) < pivot (%d).", j, arr[j], pivot), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Sorted:    getSortedSlice(),
				Comparing: []int{j, high},
				Current:   j,
			}, map[string]any{"val": arr[j], "pivot": pivot})

			if arr[j] < pivot {
				i++
				context.Boundary = i
				add("L13", trace.ActionSet, fmt.Sprintf("Incremented boundary i = %d.", i), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  getSortedSlice(),
					Current: i,
				}, map[string]any{"i": i})

				arr[i], arr[j] = arr[j], arr[i]
				ids[i], ids[j] = ids[j], ids[i]
				b.IncSwaps()
				add("L14", trace.ActionSwap, fmt.Sprintf("Swapped arr[%d] and arr[%d].", i, j), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  getSortedSlice(),
					Swapped: []int{i, j},
					Current: i,
				}, map[string]any{"i": i, "j": j})
			}
		}

		arr[i+1], arr[high] = arr[high], arr[i+1]
		ids[i+1], ids[high] = ids[high], ids[i+1]
		b.IncSwaps()
		p := i + 1
		context.PivotIndex, context.Scan = p, high
		sortedSet[p] = true
		add("L17", trace.ActionSwap, fmt.Sprintf("Placed pivot into sorted position at index %d.", p), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  getSortedSlice(),
			Swapped: []int{p, high},
			Current: p,
		}, map[string]any{"pivotIndex": p})

		add("L18", trace.ActionSet, fmt.Sprintf("Partition complete. Pivot index is %d.", p), renderers.ArrayState{
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

	add("L7", trace.ActionDone, "Quick Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  getSortedSlice(),
		Current: -1,
	}, map[string]any{"status": "complete"})

	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L13" && steps[i+1].LineID == "L14" {
			steps[i].HoldMS, steps[i+1].HoldMS = 600, 1200
		}
	}
	return steps
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}
