package heap

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Heap Sort on the initial array.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)

	b.AddStep("L1", trace.ActionInit, "Initialize Heap Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n == 0 {
		b.AddStep("L10", trace.ActionDone, "Heap Sort completed! Empty array.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n == 1 {
		b.AddStep("L10", trace.ActionDone, "Heap Sort completed! Single element sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	sorted := make([]int, 0, n)

	// Phase 1: Build max-heap
	for i := n/2 - 1; i >= 0; i-- {
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Build heap: inspecting subtree at index %d.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: i,
		}, map[string]any{"i": i})

		b.AddStep("L4", trace.ActionVisit, fmt.Sprintf("Invoking heapify for index %d on heap size %d.", i, n), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: i,
		}, map[string]any{"i": i, "n": n})

		heapifyTrace(b, arr, n, i, sorted)
	}

	// Phase 2: Extract elements from heap
	for i := n - 1; i > 0; i-- {
		b.AddStep("L6", trace.ActionHighlight, fmt.Sprintf("Extract max element: root arr[0] (%d) to position %d.", arr[0], i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: i,
		}, map[string]any{"i": i})

		arr[0], arr[i] = arr[i], arr[0]
		b.IncSwaps()
		sorted = append(sorted, i)

		b.AddStep("L7", trace.ActionSwap, fmt.Sprintf("Swapped root %d with arr[%d] (%d).", arr[i], i, arr[0]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Swapped: []int{0, i},
			Sorted:  cloneSlice(sorted),
			Current: i,
		}, map[string]any{"i": i})

		b.AddStep("L8", trace.ActionVisit, fmt.Sprintf("Re-heapify root on reduced heap size %d.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: 0,
		}, map[string]any{"n": i, "root": 0})

		heapifyTrace(b, arr, i, 0, sorted)
	}

	sorted = append(sorted, 0)
	b.AddStep("L10", trace.ActionDone, "Heap Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  cloneSlice(sorted),
		Current: -1,
	}, map[string]any{"status": "complete"})

	return b.Steps()
}

func heapifyTrace(b *trace.Builder, arr []int, n, i int, sorted []int) {
	largest := i
	l := 2*i + 1
	r := 2*i + 2

	b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Check subtree root %d against children: left %d, right %d.", i, l, r), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  cloneSlice(sorted),
		Current: i,
	}, map[string]any{"i": i, "l": l, "r": r, "largest": largest})

	if l < n {
		b.IncComparisons()
		b.AddStep("L13", trace.ActionCompare, fmt.Sprintf("Comparing left child arr[%d] (%d) with arr[%d] (%d).", l, arr[l], largest, arr[largest]), renderers.ArrayState{
			Values:    cloneSlice(arr),
			Comparing: []int{l, largest},
			Sorted:    cloneSlice(sorted),
			Current:   i,
		}, map[string]any{"l": l, "largest": largest})

		if arr[l] > arr[largest] {
			largest = l
		}
	}

	if r < n {
		b.IncComparisons()
		b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Comparing right child arr[%d] (%d) with arr[%d] (%d).", r, arr[r], largest, arr[largest]), renderers.ArrayState{
			Values:    cloneSlice(arr),
			Comparing: []int{r, largest},
			Sorted:    cloneSlice(sorted),
			Current:   i,
		}, map[string]any{"r": r, "largest": largest})

		if arr[r] > arr[largest] {
			largest = r
		}
	}

	if largest != i {
		b.AddStep("L15", trace.ActionHighlight, fmt.Sprintf("Largest is at child index %d (%d), swapping with root %d (%d).", largest, arr[largest], i, arr[i]), renderers.ArrayState{
			Values:    cloneSlice(arr),
			Comparing: []int{i, largest},
			Sorted:    cloneSlice(sorted),
			Current:   i,
		}, map[string]any{"i": i, "largest": largest})

		arr[i], arr[largest] = arr[largest], arr[i]
		b.IncSwaps()

		b.AddStep("L16", trace.ActionSwap, fmt.Sprintf("Swapped arr[%d] and arr[%d] to maintain max-heap.", i, largest), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Swapped: []int{i, largest},
			Sorted:  cloneSlice(sorted),
			Current: largest,
		}, map[string]any{"i": i, "largest": largest})

		b.AddStep("L17", trace.ActionVisit, fmt.Sprintf("Recursively heapifying subtree at index %d.", largest), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: largest,
		}, map[string]any{"largest": largest})

		heapifyTrace(b, arr, n, largest, sorted)
	}
}

func cloneSlice(s []int) []int {
	if s == nil {
		return nil
	}
	c := make([]int, len(s))
	copy(c, s)
	return c
}
