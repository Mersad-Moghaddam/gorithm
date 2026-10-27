package countingsort

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: count slice tracks non-negative range; upgrade if sparse or negative range needed.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, "Initialize Counting Sort with empty array.", renderers.ArrayState{
			Values:  []int{},
			IDs:     []string{},
			Sorted:  []int{},
			Current: -1,
			Caption: "Empty array: nothing to sort",
		}, map[string]any{"n": 0})
		b.AddStep("L3", trace.ActionHighlight, "Array is empty (n <= 1), nothing to sort.", renderers.ArrayState{
			Values:  []int{},
			IDs:     []string{},
			Sorted:  []int{},
			Current: -1,
			Caption: "Empty array: n <= 1 base case reached",
		}, map[string]any{"n": 0})
		b.AddStep("L4", trace.ActionDone, "Counting Sort completed! Array is empty.", renderers.ArrayState{
			Values:  []int{},
			IDs:     []string{},
			Sorted:  []int{},
			Current: -1,
			Caption: "Counting Sort complete: empty array",
		}, map[string]any{"status": "complete", "result": []int{}})
		return b.Steps()
	}

	if n == 1 {
		ids := []string{fmt.Sprintf("countingsort-%d", 0)}
		b.AddStep("L1", trace.ActionInit, "Initialize Counting Sort with input array.", renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  []int{0},
			Current: -1,
			Caption: "Single element array: already sorted",
		}, map[string]any{"n": 1})
		b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  []int{0},
			Current: -1,
			Caption: "Array length is 1",
		}, map[string]any{"n": 1})
		b.AddStep("L3", trace.ActionHighlight, "Single element array (n <= 1), already sorted.", renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  []int{0},
			Current: -1,
			Caption: "Base case: single element needs no sorting",
		}, map[string]any{"n": 1})
		b.AddStep("L4", trace.ActionDone, "Counting Sort completed! Single element is sorted.", renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  []int{0},
			Current: -1,
			Caption: "Counting Sort complete: single element sorted",
		}, map[string]any{"status": "complete", "result": slices.Clone(arr)})
		return b.Steps()
	}

	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("countingsort-%d", i)
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Counting Sort with input array.", renderers.ArrayState{
		Values:  slices.Clone(arr),
		IDs:     slices.Clone(ids),
		Current: -1,
		Caption: "Initialize Counting Sort: non-comparison integer sort",
	}, map[string]any{"n": n})

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values:  slices.Clone(arr),
		IDs:     slices.Clone(ids),
		Current: -1,
		Caption: fmt.Sprintf("%d values to sort", n),
	}, map[string]any{"n": n})

	b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Array length %d > 1, proceeding with sorting.", n), renderers.ArrayState{
		Values:  slices.Clone(arr),
		IDs:     slices.Clone(ids),
		Current: -1,
		Caption: fmt.Sprintf("Length %d > 1: scan input to find maximum value", n),
	}, map[string]any{"n": n})

	// Phase 1: Scanning input to find max
	max := arr[0]
	b.IncOperations()
	b.AddStep("L6", trace.ActionSet, fmt.Sprintf("Initialize max with arr[0] = %d.", max), renderers.ArrayState{
		Values:  slices.Clone(arr),
		IDs:     slices.Clone(ids),
		Current: 0,
		Caption: fmt.Sprintf("Phase 1: Initial candidate max = %d at [0]", max),
	}, map[string]any{"max": max, "i": 0})

	for i := 1; i < n; i++ {
		b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Inspecting element at index %d (%d) to find maximum value.", i, arr[i]), renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Current: i,
			Caption: fmt.Sprintf("Phase 1: Scan [%d] = %d; compare with current max = %d", i, arr[i], max),
		}, map[string]any{"i": i, "val": arr[i], "max": max})

		b.IncComparisons()
		if arr[i] > max {
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("arr[%d] (%d) > max (%d): new maximum found.", i, arr[i], max), renderers.ArrayState{
				Values:    slices.Clone(arr),
				IDs:       slices.Clone(ids),
				Comparing: []int{i},
				Current:   i,
				Caption:   fmt.Sprintf("[%d] = %d > %d: new maximum found", i, arr[i], max),
			}, map[string]any{"i": i, "val": arr[i], "max": max})

			max = arr[i]
			b.IncOperations()
			b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Updated max to %d.", max), renderers.ArrayState{
				Values:  slices.Clone(arr),
				IDs:     slices.Clone(ids),
				Current: i,
				Caption: fmt.Sprintf("Updated max to %d", max),
			}, map[string]any{"i": i, "max": max})
		} else {
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("arr[%d] (%d) <= max (%d): max remains %d.", i, arr[i], max, max), renderers.ArrayState{
				Values:    slices.Clone(arr),
				IDs:       slices.Clone(ids),
				Comparing: []int{i},
				Current:   i,
				Caption:   fmt.Sprintf("[%d] = %d <= %d: keep max %d", i, arr[i], max, max),
			}, map[string]any{"i": i, "val": arr[i], "max": max})
		}
	}

	// Phase 2: Counting frequencies of each value
	count := make([]int, max+1)
	b.IncOperations()
	b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Allocate count array of size %d (range 0 to %d).", max+1, max), renderers.ArrayState{
		Values:  slices.Clone(arr),
		IDs:     slices.Clone(ids),
		Current: -1,
		Caption: fmt.Sprintf("Phase 2: Count array of size %d (range 0..%d) allocated to store frequency counts of each distinct value", max+1, max),
	}, map[string]any{"max": max, "countSize": max + 1, "count": slices.Clone(count)})

	for i := 0; i < n; i++ {
		val := arr[i]
		b.AddStep("L13", trace.ActionHighlight, fmt.Sprintf("Tallying frequency for arr[%d] = %d.", i, val), renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Current: i,
			Caption: fmt.Sprintf("Phase 2: Scan arr[%d] = %d to count frequency in count[%d]", i, val, val),
		}, map[string]any{"i": i, "val": val, "count": slices.Clone(count)})

		count[val]++
		b.IncOperations()
		b.AddStep("L14", trace.ActionSet, fmt.Sprintf("Increment frequency count[%d] to %d.", val, count[val]), renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Current: i,
			Caption: fmt.Sprintf("Phase 2: count[%d] = %d (%d occurrence(s) seen so far)", val, count[val], count[val]),
		}, map[string]any{"i": i, "val": val, "count": slices.Clone(count), "frequencies": slices.Clone(count)})
	}

	// Phase 3: Accumulating prefix sums
	b.AddStep("L16", trace.ActionHighlight, "Accumulate frequencies into prefix sums to determine positions.", renderers.ArrayState{
		Values:  slices.Clone(arr),
		IDs:     slices.Clone(ids),
		Current: -1,
		Caption: "Phase 3: Accumulate prefix sums; count[v] becomes total count of elements <= v (determining 1-based output positions)",
	}, map[string]any{"count": slices.Clone(count)})

	for i := 1; i <= max; i++ {
		count[i] += count[i-1]
		b.IncOperations()
		b.AddStep("L17", trace.ActionSet, fmt.Sprintf("Prefix sum: count[%d] += count[%d] = %d.", i, i-1, count[i]), renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Current: -1,
			Caption: fmt.Sprintf("Phase 3: Prefix sum count[%d] += count[%d] = %d (%d element(s) <= %d; 1-based positions up to %d)", i, i-1, count[i], count[i], i, count[i]),
		}, map[string]any{"i": i, "count": slices.Clone(count), "prefixSums": slices.Clone(count)})
	}

	// Phase 4: Placing elements into output array with backwards scan, preserving stability
	output := make([]int, n)
	outputIDs := make([]string, n)
	for k := range outputIDs {
		outputIDs[k] = fmt.Sprintf("countingsort-out-%d", k)
	}
	b.IncOperations()
	b.AddStep("L19", trace.ActionSet, fmt.Sprintf("Allocate output array of size %d.", n), renderers.ArrayState{
		Values:  slices.Clone(output),
		IDs:     slices.Clone(outputIDs),
		Current: -1,
		Caption: fmt.Sprintf("Phase 4: Output array of size %d allocated for stable placement", n),
	}, map[string]any{"n": n, "count": slices.Clone(count), "output": slices.Clone(output)})

	sortedIndices := make([]int, 0, n)
	for i := n - 1; i >= 0; i-- {
		val := arr[i]
		targetIdx := count[val] - 1
		b.AddStep("L20", trace.ActionHighlight, fmt.Sprintf("Examining arr[%d] = %d from right to preserve stability.", i, val), renderers.ArrayState{
			Values:  slices.Clone(output),
			IDs:     slices.Clone(outputIDs),
			Sorted:  slices.Clone(sortedIndices),
			Current: -1,
			Caption: fmt.Sprintf("Phase 4: Backwards scan arr[%d] = %d (ID %s). Prefix sum count[%d] = %d gives 1-based position %d (index %d); right-to-left preserves duplicate stability", i, val, ids[i], val, count[val], count[val], targetIdx),
		}, map[string]any{"i": i, "val": val, "targetIndex": targetIdx, "count": slices.Clone(count), "output": slices.Clone(output), "input": slices.Clone(arr)})

		output[targetIdx] = val
		outputIDs[targetIdx] = ids[i]
		sortedIndices = append(sortedIndices, targetIdx)
		b.IncOperations()
		b.AddStep("L21", trace.ActionSet, fmt.Sprintf("Place arr[%d] (%d) at output[%d].", i, val, targetIdx), renderers.ArrayState{
			Values:  slices.Clone(output),
			IDs:     slices.Clone(outputIDs),
			Current: targetIdx,
			Swapped: []int{targetIdx},
			Sorted:  slices.Clone(sortedIndices),
			Caption: fmt.Sprintf("Phase 4: Place arr[%d] (%d, ID %s) into output[%d]. Rightmost duplicates placed first at higher positions, preserving stability", i, val, ids[i], targetIdx),
		}, map[string]any{"i": i, "val": val, "targetIndex": targetIdx, "count": slices.Clone(count), "output": slices.Clone(output)})

		count[val]--
		b.IncOperations()
		b.AddStep("L22", trace.ActionSet, fmt.Sprintf("Decrement count[%d] to %d.", val, count[val]), renderers.ArrayState{
			Values:  slices.Clone(output),
			IDs:     slices.Clone(outputIDs),
			Current: targetIdx,
			Sorted:  slices.Clone(sortedIndices),
			Caption: fmt.Sprintf("Phase 4: Decrement count[%d] to %d so preceding duplicates with value %d will land at index %d", val, count[val], val, count[val]-1),
		}, map[string]any{"i": i, "val": val, "count": slices.Clone(count), "output": slices.Clone(output)})
	}

	b.AddStep("L24", trace.ActionDone, "Counting Sort completed! All elements sorted stably.", renderers.ArrayState{
		Values:  slices.Clone(output),
		IDs:     slices.Clone(outputIDs),
		Sorted:  makeRange(0, n-1),
		Current: -1,
		Caption: "Counting Sort complete: all elements placed into output array stably with prefix sum positions",
	}, map[string]any{"status": "complete", "result": slices.Clone(output)})

	return b.Steps()
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
