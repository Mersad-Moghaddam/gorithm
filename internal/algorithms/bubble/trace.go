package bubble

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)

	b.AddStep("L1", trace.ActionInit, "Initialize Bubble Sort with input array.", renderers.ArrayState{
		Values: cloneSlice(arr),
	}, map[string]any{"n": n})

	sorted := make([]int, 0, n)

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values: cloneSlice(arr),
	}, map[string]any{"n": n})

	for i := 0; i < n-1; i++ {
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Starting outer pass i = %d.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
			Sorted:  cloneSlice(sorted),
		}, map[string]any{"i": i})

		for j := 0; j < n-i-1; j++ {
			b.IncComparisons()
			b.AddStep("L5", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) and arr[%d] (%d).", j, arr[j], j+1, arr[j+1]), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{j, j + 1},
				Sorted:    cloneSlice(sorted),
			}, map[string]any{"i": i, "j": j, "val1": arr[j], "val2": arr[j+1]})

			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				b.IncSwaps()
				b.AddStep("L6", trace.ActionSwap, fmt.Sprintf("Swapped %d and %d because %d > %d.", arr[j+1], arr[j], arr[j+1], arr[j]), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Swapped: []int{j, j + 1},
					Sorted:  cloneSlice(sorted),
				}, map[string]any{"i": i, "j": j})
			}
		}

		sorted = append([]int{n - 1 - i}, sorted...)
		b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Element at index %d is now in sorted position.", n-1-i), renderers.ArrayState{
			Values: cloneSlice(arr),
			Sorted: cloneSlice(sorted),
		}, map[string]any{"i": i, "sortedIndex": n - 1 - i})
	}

	// First element is also sorted
	sorted = append([]int{0}, sorted...)
	b.AddStep("L10", trace.ActionDone, "Bubble Sort completed! All elements sorted.", renderers.ArrayState{
		Values: cloneSlice(arr),
		Sorted: cloneSlice(sorted),
	}, map[string]any{"status": "complete"})

	return b.Steps()
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}
