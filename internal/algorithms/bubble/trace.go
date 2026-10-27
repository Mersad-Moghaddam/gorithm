package bubble

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("bubble-%d", i)
	}
	sorted := make([]int, 0, n)
	snapshot := func(caption string) renderers.ArrayState {
		return renderers.ArrayState{Values: cloneSlice(arr), IDs: slices.Clone(ids), Sorted: cloneSlice(sorted), Current: -1, Caption: caption}
	}

	b.AddStep("L1", trace.ActionInit, "Compare neighbors; move the larger value right. Green marks settled values.", snapshot("Compare neighbors; larger values move right"), map[string]any{"n": n})

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), snapshot(fmt.Sprintf("%d values to sort", n)), map[string]any{"n": n})

	for i := 0; i < n-1; i++ {
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Pass %d scans indices 0 through %d; the green suffix stays fixed.", i+1, n-i-1), snapshot(fmt.Sprintf("Pass %d: scan [0..%d]; green suffix stays fixed", i+1, n-i-1)), map[string]any{"i": i})

		for j := 0; j < n-i-1; j++ {
			b.IncComparisons()
			state := snapshot(fmt.Sprintf("Neighbors [%d], [%d]: is %d > %d?", j, j+1, arr[j], arr[j+1]))
			state.Comparing = []int{j, j + 1}
			b.AddStep("L5", trace.ActionCompare, state.Caption, state, map[string]any{"i": i, "j": j, "val1": arr[j], "val2": arr[j+1]})

			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
				ids[j], ids[j+1] = ids[j+1], ids[j]
				b.IncSwaps()
				state = snapshot(fmt.Sprintf("%d > %d: swap; %d moves right", arr[j+1], arr[j], arr[j+1]))
				state.Swapped = []int{j, j + 1}
				b.AddStep("L6", trace.ActionSwap, state.Caption, state, map[string]any{"i": i, "j": j})
			} else {
				caption := fmt.Sprintf("Keep order: %d < %d", arr[j], arr[j+1])
				if arr[j] == arr[j+1] {
					caption = fmt.Sprintf("Keep order: %d = %d; equal objects stay stable", arr[j], arr[j+1])
				}
				state = snapshot(caption)
				state.Comparing = []int{j, j + 1}
				b.AddStep("L5", trace.ActionHighlight, caption, state, map[string]any{"i": i, "j": j})
			}
		}

		sorted = append([]int{n - 1 - i}, sorted...)
		caption := fmt.Sprintf("%d settled at [%d]; green suffix is final", arr[n-1-i], n-1-i)
		b.AddStep("L8", trace.ActionSet, caption, snapshot(caption), map[string]any{"i": i, "sortedIndex": n - 1 - i})
	}

	caption := "Sorted: every value is in its final position"
	if n > 0 {
		sorted = append([]int{0}, sorted...)
	} else {
		caption = "Empty array: nothing to compare or swap"
	}
	b.AddStep("L10", trace.ActionDone, caption, snapshot(caption), map[string]any{"status": "complete"})

	return b.Steps()
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}
