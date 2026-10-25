package radixsort

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

	if n == 0 {
		b.AddStep("L1", trace.ActionInit, "Initialize Radix Sort (LSD) with empty array.", renderers.ArrayState{
			Current: -1,
		}, map[string]any{"n": 0})
		b.AddStep("L6", trace.ActionDone, "Radix Sort completed! Array is empty.", renderers.ArrayState{
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Radix Sort (LSD) with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n == 1 {
		b.AddStep("L6", trace.ActionDone, "Radix Sort completed! Single element is sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	// L2: Find maxVal
	maxVal := arr[0]
	maxIdx := 0
	for i, v := range arr {
		b.IncOperations()
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Found maximum value %d at index %d to determine digit count.", maxVal, maxIdx), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: maxIdx,
	}, map[string]any{"maxVal": maxVal, "maxIdx": maxIdx})

	for exp := 1; maxVal/exp > 0; exp *= 10 {
		b.IncOperations()
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Processing digit place %d (exp = %d).", exp, exp), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp, "maxVal": maxVal})

		b.AddStep("L4", trace.ActionVisit, fmt.Sprintf("Invoking counting sort for exponent %d.", exp), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp})

		// countingSortByDigit
		b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Starting counting sort for exp = %d.", exp), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp, "n": n})

		output := make([]int, n)
		b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Allocated output buffer of size %d.", n), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp, "n": n})

		var count [10]int
		b.AddStep("L9", trace.ActionSet, "Initialized 10 digit buckets (0 through 9) to zero.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp})

		// L10: Count digit frequencies
		for i := 0; i < n; i++ {
			digit := (arr[i] / exp) % 10
			count[digit]++
			b.IncOperations()
			b.AddStep("L10", trace.ActionSet, fmt.Sprintf("arr[%d]=%d has digit %d (exp=%d). count[%d] is now %d.", i, arr[i], digit, exp, digit, count[digit]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"exp": exp, "i": i, "val": arr[i], "digit": digit, "count": count[digit]})
		}

		// L11: Compute cumulative positions
		for i := 1; i < 10; i++ {
			count[i] += count[i-1]
			b.IncOperations()
		}
		b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Accumulated prefix sums in count array for exp=%d.", exp), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp})

		// L12-L16: Build output array backwards to preserve stability
		for i := n - 1; i >= 0; i-- {
			b.IncOperations()
			b.AddStep("L12", trace.ActionHighlight, fmt.Sprintf("Processing arr[%d] = %d backwards for stability.", i, arr[i]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"exp": exp, "i": i, "val": arr[i]})

			digit := (arr[i] / exp) % 10
			b.AddStep("L13", trace.ActionSet, fmt.Sprintf("Extracted digit %d from arr[%d]=%d.", digit, i, arr[i]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"exp": exp, "i": i, "digit": digit, "val": arr[i]})

			targetPos := count[digit] - 1
			output[targetPos] = arr[i]
			b.AddStep("L14", trace.ActionSet, fmt.Sprintf("Placed %d into output[%d] based on prefix sum.", arr[i], targetPos), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"exp": exp, "i": i, "val": arr[i], "digit": digit, "targetPos": targetPos})

			count[digit]--
			b.AddStep("L15", trace.ActionSet, fmt.Sprintf("Decremented count[%d] to %d.", digit, count[digit]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"exp": exp, "digit": digit, "remainingCount": count[digit]})
		}

		// L17: Copy output back to arr
		for i := 0; i < n; i++ {
			arr[i] = output[i]
			b.IncSwaps()
			b.AddStep("L17", trace.ActionSwap, fmt.Sprintf("Updated arr[%d] = %d from output.", i, arr[i]), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Swapped: []int{i},
				Current: i,
			}, map[string]any{"exp": exp, "i": i, "val": arr[i]})
		}

		b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Completed pass for exp = %d.", exp), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"exp": exp})
	}

	b.AddStep("L6", trace.ActionDone, "Radix Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  makeRange(0, n-1),
		Current: -1,
	}, map[string]any{"status": "complete"})

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
