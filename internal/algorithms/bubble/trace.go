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
	ids := make([]int, n)
	for i := range ids {
		ids[i] = i
	}
	pass, scan, end := -1, -1, n-1
	minValue, maxValue := 0, 0
	for _, value := range initial {
		minValue, maxValue = min(minValue, value), max(maxValue, value)
	}
	if minValue == maxValue {
		maxValue++
	}
	snapshot := func(state renderers.ArrayState) renderers.BubbleLessonState {
		state.Current = -1
		return renderers.BubbleLessonState{ArrayState: state, IDs: cloneSlice(ids), Pass: pass, Scan: scan, UnsortedEnd: end, MinValue: minValue, MaxValue: maxValue}
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Bubble Sort with input array.", snapshot(renderers.ArrayState{
		Values: cloneSlice(arr),
	}), map[string]any{"n": n})

	sorted := make([]int, 0, n)

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), snapshot(renderers.ArrayState{
		Values: cloneSlice(arr),
	}), map[string]any{"n": n})

	for i := 0; i < n-1; i++ {
		pass, scan, end = i, -1, n-i-1
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Starting outer pass i = %d.", i), snapshot(renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
			Sorted:  cloneSlice(sorted),
		}), map[string]any{"i": i})

		for j := 0; j < n-i-1; j++ {
			scan = j
			b.IncComparisons()
			b.AddStep("L5", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) and arr[%d] (%d).", j, arr[j], j+1, arr[j+1]), snapshot(renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{j, j + 1},
				Sorted:    cloneSlice(sorted),
			}), map[string]any{"i": i, "j": j, "val1": arr[j], "val2": arr[j+1]})

			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				ids[j], ids[j+1] = ids[j+1], ids[j]
				b.IncSwaps()
				b.AddStep("L6", trace.ActionSwap, fmt.Sprintf("Swapped %d and %d because %d > %d.", arr[j+1], arr[j], arr[j+1], arr[j]), snapshot(renderers.ArrayState{
					Values:  cloneSlice(arr),
					Swapped: []int{j, j + 1},
					Sorted:  cloneSlice(sorted),
				}), map[string]any{"i": i, "j": j})
			}
		}

		sorted = append([]int{n - 1 - i}, sorted...)
		scan, end = -1, n-i-2
		b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Element at index %d is now in sorted position.", n-1-i), snapshot(renderers.ArrayState{
			Values: cloneSlice(arr),
			Sorted: cloneSlice(sorted),
		}), map[string]any{"i": i, "sortedIndex": n - 1 - i})
	}

	// First element is also sorted
	if n > 0 {
		sorted = append([]int{0}, sorted...)
	}
	scan, end = -1, -1
	b.AddStep("L10", trace.ActionDone, "Bubble Sort completed! All elements sorted.", snapshot(renderers.ArrayState{
		Values: cloneSlice(arr),
		Sorted: cloneSlice(sorted),
	}), map[string]any{"status": "complete"})

	steps := b.Steps()
	for i := range steps {
		if steps[i].Action == trace.ActionSwap {
			steps[i-1].HoldMS, steps[i].HoldMS = 600, 1200
		}
	}
	return steps
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}
