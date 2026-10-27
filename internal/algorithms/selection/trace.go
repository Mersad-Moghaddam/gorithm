package selection

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("selection-%d", i)
	}
	sorted := make([]int, 0, n)
	snapshot := func(caption string, candidate int) renderers.ArrayState {
		return renderers.ArrayState{Values: slices.Clone(arr), IDs: slices.Clone(ids), Sorted: slices.Clone(sorted), Current: candidate, Caption: caption}
	}
	b.AddStep("L1", trace.ActionInit, "Find minimum in unsorted suffix; place it at start of suffix.", snapshot("Find minimum; green prefix is settled", -1), map[string]any{"n": n})
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), snapshot(fmt.Sprintf("%d values to sort", n), -1), map[string]any{"n": n})
	for i := 0; i < n-1; i++ {
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Start pass at index %d.", i), snapshot(fmt.Sprintf("Pass %d: scan unsorted suffix [%d..%d]", i+1, i, n-1), -1), map[string]any{"i": i})
		minIdx := i
		caption := fmt.Sprintf("Candidate minimum: [%d] = %d", minIdx, arr[minIdx])
		b.AddStep("L4", trace.ActionSet, caption, snapshot(caption, minIdx), map[string]any{"i": i, "minIdx": minIdx})
		for j := i + 1; j < n; j++ {
			b.IncComparisons()
			caption = fmt.Sprintf("Scan [%d] = %d; compare with candidate [%d] = %d", j, arr[j], minIdx, arr[minIdx])
			state := snapshot(caption, minIdx)
			state.Comparing = []int{j}
			b.AddStep("L6", trace.ActionCompare, caption, state, map[string]any{"i": i, "j": j, "minIdx": minIdx, "val": arr[j], "minVal": arr[minIdx]})
			if arr[j] < arr[minIdx] {
				minIdx = j
				caption = fmt.Sprintf("Smaller: [%d] = %d is new candidate minimum", j, arr[j])
				b.AddStep("L7", trace.ActionSet, caption, snapshot(caption, minIdx), map[string]any{"i": i, "j": j, "minIdx": minIdx})
			} else {
				caption = fmt.Sprintf("Keep candidate [%d] = %d; [%d] = %d is not smaller", minIdx, arr[minIdx], j, arr[j])
				state = snapshot(caption, minIdx)
				state.Comparing = []int{j}
				b.AddStep("L6", trace.ActionHighlight, caption, state, map[string]any{"i": i, "j": j, "minIdx": minIdx})
			}
		}
		if minIdx != i {
			selected := arr[minIdx]
			arr[i], arr[minIdx] = arr[minIdx], arr[i]
			ids[i], ids[minIdx] = ids[minIdx], ids[i]
			b.IncSwaps()
			caption = fmt.Sprintf("Place minimum %d at [%d]; displaced value moves to [%d]", selected, i, minIdx)
			state := snapshot(caption, i)
			state.Swapped = []int{i, minIdx}
			b.AddStep("L10", trace.ActionSwap, caption, state, map[string]any{"i": i, "minIdx": minIdx})
		} else {
			caption = fmt.Sprintf("Minimum already at [%d]; no swap", i)
			b.AddStep("L10", trace.ActionSet, caption, snapshot(caption, i), map[string]any{"i": i, "minIdx": minIdx})
		}
		sorted = append(sorted, i)
		caption = fmt.Sprintf("[%d] settled; green prefix [0..%d] stays fixed", i, i)
		b.AddStep("L11", trace.ActionSet, caption, snapshot(caption, -1), map[string]any{"i": i, "sortedIndex": i})
	}
	if n > 0 {
		sorted = append(sorted, n-1)
	}
	caption := "Sorted: every position settled"
	if n == 0 {
		caption = "Empty array: nothing to scan"
	}
	b.AddStep("L12", trace.ActionDone, caption, snapshot(caption, -1), map[string]any{"status": "complete"})
	return b.Steps()
}
