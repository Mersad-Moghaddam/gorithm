package selection

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
	low, high := 0, 0
	for i, value := range arr {
		ids[i] = i
		low = min(low, value)
		high = max(high, value)
	}
	pass, scan, minimum := -1, -1, -1
	add := func(line string, action trace.ActionType, explanation string, state renderers.ArrayState, variables map[string]any) {
		b.AddStep(line, action, explanation, renderers.SelectionLessonState{
			ArrayState: state, IDs: cloneSlice(ids), Pass: pass, Scan: scan,
			MinIndex: minimum, SortedPrefix: len(state.Sorted), MinValue: low, MaxValue: high,
		}, variables)
	}

	add("L1", trace.ActionInit, "Initialize Selection Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	sorted := make([]int, 0, n)

	add("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n <= 1 {
		if n == 1 {
			sorted = append(sorted, 0)
		}
		add("L12", trace.ActionDone, "Selection Sort completed! All elements sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	for i := 0; i < n-1; i++ {
		pass, scan, minimum = i, -1, -1
		add("L3", trace.ActionHighlight, fmt.Sprintf("Starting pass i = %d: finding minimum in unsorted subarray.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
			Sorted:  cloneSlice(sorted),
		}, map[string]any{"i": i})

		minIdx := i
		minimum = minIdx
		add("L4", trace.ActionSet, fmt.Sprintf("Assume minimum is at index %d (value %d).", i, arr[i]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: minIdx,
			Sorted:  cloneSlice(sorted),
		}, map[string]any{"i": i, "minIdx": minIdx})

		for j := i + 1; j < n; j++ {
			scan = j
			b.IncComparisons()
			add("L6", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) with current minimum arr[%d] (%d).", j, arr[j], minIdx, arr[minIdx]), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{minIdx, j},
				Sorted:    cloneSlice(sorted),
				Current:   minIdx,
			}, map[string]any{"i": i, "j": j, "minIdx": minIdx, "val": arr[j], "minVal": arr[minIdx]})

			if arr[j] < arr[minIdx] {
				minIdx = j
				minimum = minIdx
				add("L7", trace.ActionSet, fmt.Sprintf("Found smaller element at index %d (%d); update minIdx to %d.", j, arr[j], minIdx), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  cloneSlice(sorted),
					Current: minIdx,
				}, map[string]any{"i": i, "j": j, "minIdx": minIdx})
			}
		}

		scan = -1
		if minIdx != i {
			arr[i], arr[minIdx] = arr[minIdx], arr[i]
			ids[i], ids[minIdx] = ids[minIdx], ids[i]
			b.IncSwaps()
			add("L10", trace.ActionSwap, fmt.Sprintf("Swapped arr[%d] (%d) and arr[%d] (%d).", i, arr[i], minIdx, arr[minIdx]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Swapped: []int{i, minIdx},
				Sorted:  cloneSlice(sorted),
				Current: i,
			}, map[string]any{"i": i, "minIdx": minIdx})
		} else {
			add("L10", trace.ActionSet, fmt.Sprintf("arr[%d] (%d) is already minimum; no swap needed.", i, arr[i]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Sorted:  cloneSlice(sorted),
				Current: i,
			}, map[string]any{"i": i, "minIdx": minIdx})
		}

		sorted = append(sorted, i)
		scan, minimum = -1, -1
		add("L11", trace.ActionSet, fmt.Sprintf("Element at index %d is now in sorted position.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  cloneSlice(sorted),
			Current: -1,
		}, map[string]any{"i": i, "sortedIndex": i})
	}

	sorted = append(sorted, n-1)
	pass, scan, minimum = -1, -1, -1
	add("L12", trace.ActionDone, "Selection Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  cloneSlice(sorted),
		Current: -1,
	}, map[string]any{"status": "complete"})

	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L6" && steps[i+1].LineID == "L7" {
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
