package countingsort

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: count slice tracks non-negative range; upgrade if sparse or negative range needed.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, "Initialize Counting Sort with empty array.", renderers.ArrayState{
			Values:  []int{},
			Current: -1,
		}, map[string]any{"n": 0})
		b.AddStep("L3", trace.ActionHighlight, "Array is empty (n <= 1), nothing to sort.", renderers.ArrayState{
			Values:  []int{},
			Current: -1,
		}, map[string]any{"n": 0})
		b.AddStep("L4", trace.ActionDone, "Counting Sort completed! Array is empty.", renderers.ArrayState{
			Values:  []int{},
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	if n == 1 {
		b.AddStep("L1", trace.ActionInit, "Initialize Counting Sort with input array.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"n": 1})
		b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"n": 1})
		b.AddStep("L3", trace.ActionHighlight, "Single element array (n <= 1), already sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"n": 1})
		b.AddStep("L4", trace.ActionDone, "Counting Sort completed! Single element is sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Counting Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Array length %d > 1, proceeding with sorting.", n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	max := arr[0]
	b.IncOperations()
	b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Initialize max with arr[0] = %d.", max), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: 0,
	}, map[string]any{"max": max, "i": 0})

	for i := 1; i < n; i++ {
		b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Inspecting element at index %d (%d) to find maximum value.", i, arr[i]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
		}, map[string]any{"i": i, "val": arr[i], "max": max})

		b.IncComparisons()
		if arr[i] > max {
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("arr[%d] (%d) > max (%d): new maximum found.", i, arr[i], max), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{i},
				Current:   i,
			}, map[string]any{"i": i, "val": arr[i], "max": max})

			max = arr[i]
			b.IncOperations()
			b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Updated max to %d.", max), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"i": i, "max": max})
		} else {
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("arr[%d] (%d) <= max (%d): max remains %d.", i, arr[i], max, max), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{i},
				Current:   i,
			}, map[string]any{"i": i, "val": arr[i], "max": max})
		}
	}

	count := make([]int, max+1)
	b.IncOperations()
	b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Allocate count array of size %d (range 0 to %d).", max+1, max), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"max": max, "countSize": max + 1, "count": cloneSlice(count)})

	for i := 0; i < n; i++ {
		val := arr[i]
		b.AddStep("L13", trace.ActionHighlight, fmt.Sprintf("Tallying frequency for arr[%d] = %d.", i, val), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
		}, map[string]any{"i": i, "val": val, "count": cloneSlice(count)})

		count[val]++
		b.IncOperations()
		b.AddStep("L14", trace.ActionSet, fmt.Sprintf("Increment frequency count[%d] to %d.", val, count[val]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: i,
		}, map[string]any{"i": i, "val": val, "count": cloneSlice(count), "frequencies": cloneSlice(count)})
	}

	b.AddStep("L16", trace.ActionHighlight, "Accumulate frequencies into prefix sums to determine positions.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"count": cloneSlice(count)})

	for i := 1; i <= max; i++ {
		count[i] += count[i-1]
		b.IncOperations()
		b.AddStep("L17", trace.ActionSet, fmt.Sprintf("Prefix sum: count[%d] += count[%d] = %d.", i, i-1, count[i]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"i": i, "count": cloneSlice(count), "prefixSums": cloneSlice(count)})
	}

	output := make([]int, n)
	b.IncOperations()
	b.AddStep("L19", trace.ActionSet, fmt.Sprintf("Allocate output array of size %d.", n), renderers.ArrayState{
		Values:  cloneSlice(output),
		Current: -1,
	}, map[string]any{"n": n, "count": cloneSlice(count), "output": cloneSlice(output)})

	sortedIndices := make([]int, 0, n)
	for i := n - 1; i >= 0; i-- {
		val := arr[i]
		b.AddStep("L20", trace.ActionHighlight, fmt.Sprintf("Examining arr[%d] = %d from right to preserve stability.", i, val), renderers.ArrayState{
			Values:  cloneSlice(output),
			Sorted:  cloneSlice(sortedIndices),
			Current: -1,
		}, map[string]any{"i": i, "val": val, "count": cloneSlice(count)})

		targetIdx := count[val] - 1
		output[targetIdx] = val
		sortedIndices = append(sortedIndices, targetIdx)
		b.IncOperations()
		b.AddStep("L21", trace.ActionSet, fmt.Sprintf("Place arr[%d] (%d) at output[%d].", i, val, targetIdx), renderers.ArrayState{
			Values:  cloneSlice(output),
			Current: targetIdx,
			Swapped: []int{targetIdx},
			Sorted:  cloneSlice(sortedIndices),
		}, map[string]any{"i": i, "val": val, "targetIndex": targetIdx, "count": cloneSlice(count), "output": cloneSlice(output)})

		count[val]--
		b.IncOperations()
		b.AddStep("L22", trace.ActionSet, fmt.Sprintf("Decrement count[%d] to %d.", val, count[val]), renderers.ArrayState{
			Values:  cloneSlice(output),
			Current: targetIdx,
			Sorted:  cloneSlice(sortedIndices),
		}, map[string]any{"i": i, "val": val, "count": cloneSlice(count)})
	}

	b.AddStep("L24", trace.ActionDone, "Counting Sort completed! All elements sorted stably.", renderers.ArrayState{
		Values:  cloneSlice(output),
		Sorted:  makeRange(0, n-1),
		Current: -1,
	}, map[string]any{"status": "complete", "result": cloneSlice(output)})

	return b.Steps()
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}

func makeRange(start, end int) []int {
	if start > end {
		return []int{}
	}
	res := make([]int, end-start+1)
	for i := range res {
		res[i] = start + i
	}
	return res
}
