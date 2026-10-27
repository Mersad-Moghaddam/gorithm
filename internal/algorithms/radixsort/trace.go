package radixsort

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: base-10 LSD sort trace; upgrade to configurable base if arbitrary radices requested.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("radixsort-%d", i)
	}
	sorted := make([]int, 0, n)

	snapshot := func(caption string, current int) renderers.ArrayState {
		return renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  slices.Clone(sorted),
			Current: current,
			Caption: caption,
		}
	}

	if n == 0 {
		caption := "Empty array: nothing to sort"
		b.AddStep("L1", trace.ActionInit, "Initialize Radix Sort (LSD) with empty array.", snapshot(caption, -1), map[string]any{"n": 0})
		b.AddStep("L6", trace.ActionDone, "Radix Sort completed! Array is empty.", snapshot("Radix Sort completed: array is empty", -1), map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L1", trace.ActionInit, "Initialize Radix Sort (LSD) with input array.", snapshot(fmt.Sprintf("Sort %d integers digit-by-digit using stable passes", n), -1), map[string]any{"n": n})

	if n == 1 {
		b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Found maximum value %d at index 0.", arr[0]), snapshot(fmt.Sprintf("Maximum value %d: 1 element is already sorted", arr[0]), 0), map[string]any{"maxVal": arr[0], "maxIdx": 0})
		sorted = append(sorted, 0)
		b.AddStep("L6", trace.ActionDone, "Radix Sort completed! Single element is sorted.", snapshot("Radix Sort completed: single element sorted", -1), map[string]any{"status": "complete"})
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
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Found maximum value %d at index %d to determine digit count.", maxVal, maxIdx), snapshot(fmt.Sprintf("Maximum value %d determines total passes", maxVal), maxIdx), map[string]any{"maxVal": maxVal, "maxIdx": maxIdx})

	for exp := 1; maxVal/exp > 0; exp *= 10 {
		b.IncOperations()
		digitName := digitPlaceName(exp)

		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Processing digit place %d (%s). Array before pass: %s", exp, digitName, formatElementsWithDigit(arr, exp)), snapshot(fmt.Sprintf("Pass exp=%d (%s): inspect %s of each element", exp, digitName, digitName), -1), map[string]any{"exp": exp, "digitPlace": digitName, "maxVal": maxVal})

		b.AddStep("L4", trace.ActionVisit, fmt.Sprintf("Invoking counting sort subroutine for exponent %d (%s).", exp, digitName), snapshot(fmt.Sprintf("Pass exp=%d (%s): run counting sort on %s", exp, digitName, digitName), -1), map[string]any{"exp": exp, "digitPlace": digitName})

		// countingSortByDigit
		b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Starting counting sort subroutine for exp=%d (%s).", exp, digitName), snapshot(fmt.Sprintf("Pass exp=%d (%s): start digit counting sort", exp, digitName), -1), map[string]any{"exp": exp, "digitPlace": digitName, "n": n})

		output := make([]int, n)
		outputIDs := make([]string, n)
		b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Allocated output buffer of size %d to hold stably ordered elements.", n), snapshot(fmt.Sprintf("Pass exp=%d (%s): allocate output buffer of size %d", exp, digitName, n), -1), map[string]any{"exp": exp, "digitPlace": digitName, "n": n})

		var count [10]int
		b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Initialized 10 digit buckets (0 through 9) to zero for exp=%d (%s).", exp, digitName), snapshot(fmt.Sprintf("Pass exp=%d (%s): initialize digit buckets [0..9] to zero", exp, digitName), -1), map[string]any{"exp": exp, "digitPlace": digitName})

		// L10: Count digit frequencies
		for i := 0; i < n; i++ {
			digit := (arr[i] / exp) % 10
			count[digit]++
			b.IncOperations()
			caption := fmt.Sprintf("Pass exp=%d (%s): arr[%d]=%d has digit %d (count[%d]=%d)", exp, digitName, i, arr[i], digit, digit, count[digit])
			desc := fmt.Sprintf("Inspecting arr[%d]=%d: %s is %d. count[%d] incremented to %d.", i, arr[i], digitName, digit, digit, count[digit])
			state := snapshot(caption, i)
			state.Comparing = []int{i}
			b.AddStep("L10", trace.ActionSet, desc, state, map[string]any{
				"exp": exp, "digitPlace": digitName, "i": i, "val": arr[i], "digit": digit, "count": slices.Clone(count[:]),
			})
		}

		// L11: Compute cumulative positions (prefix sums)
		for i := 1; i < 10; i++ {
			count[i] += count[i-1]
			b.IncOperations()
		}
		captionPrefix := fmt.Sprintf("Pass exp=%d (%s): prefix sums determine bucket end positions", exp, digitName)
		descPrefix := "Prefix sums in count array define ending index for each digit bucket 0..9."
		b.AddStep("L11", trace.ActionSet, descPrefix, snapshot(captionPrefix, -1), map[string]any{
			"exp": exp, "digitPlace": digitName, "prefixSums": slices.Clone(count[:]),
		})

		// L12-L16: Build output array backwards to preserve stability
		for i := n - 1; i >= 0; i-- {
			b.IncOperations()
			caption12 := fmt.Sprintf("Pass exp=%d (%s): scan arr[%d]=%d backwards for stability", exp, digitName, i, arr[i])
			desc12 := fmt.Sprintf("Inspecting arr[%d]=%d backwards to preserve relative order of duplicates.", i, arr[i])
			state12 := snapshot(caption12, i)
			state12.Comparing = []int{i}
			b.AddStep("L12", trace.ActionHighlight, desc12, state12, map[string]any{
				"exp": exp, "digitPlace": digitName, "i": i, "val": arr[i],
			})

			digit := (arr[i] / exp) % 10
			caption13 := fmt.Sprintf("Pass exp=%d (%s): arr[%d]=%d has digit %d", exp, digitName, i, arr[i], digit)
			desc13 := fmt.Sprintf("Extracted %s %d from arr[%d]=%d.", digitName, digit, i, arr[i])
			b.AddStep("L13", trace.ActionSet, desc13, snapshot(caption13, i), map[string]any{
				"exp": exp, "digitPlace": digitName, "i": i, "val": arr[i], "digit": digit,
			})

			targetPos := count[digit] - 1
			output[targetPos] = arr[i]
			outputIDs[targetPos] = ids[i]
			caption14 := fmt.Sprintf("Pass exp=%d (%s): place %d into output[%d] (digit %d)", exp, digitName, arr[i], targetPos, digit)
			desc14 := fmt.Sprintf("Placed arr[%d]=%d (ID %s) into output[%d] based on prefix sum for digit %d.", i, arr[i], ids[i], targetPos, digit)
			state14 := snapshot(caption14, i)
			state14.Swapped = []int{i}
			b.AddStep("L14", trace.ActionSet, desc14, state14, map[string]any{
				"exp": exp, "digitPlace": digitName, "i": i, "val": arr[i], "digit": digit, "targetPos": targetPos,
			})

			count[digit]--
			caption15 := fmt.Sprintf("Pass exp=%d (%s): decrement count[%d] to %d", exp, digitName, digit, count[digit])
			desc15 := fmt.Sprintf("Decremented count[%d] to %d so next element with digit %d goes to index %d.", digit, count[digit], digit, count[digit]-1)
			b.AddStep("L15", trace.ActionSet, desc15, snapshot(caption15, i), map[string]any{
				"exp": exp, "digitPlace": digitName, "digit": digit, "remainingCount": count[digit],
			})
		}

		// L17: Copy output back to arr
		copy(arr, output)
		copy(ids, outputIDs)
		b.IncSwaps()
		caption17 := fmt.Sprintf("Pass exp=%d (%s): copy stably grouped output back to array", exp, digitName)
		desc17 := fmt.Sprintf("Copied all %d elements from output buffer back into main array for the next pass.", n)
		state17 := snapshot(caption17, -1)
		state17.Swapped = makeRange(0, n-1)
		b.AddStep("L17", trace.ActionSwap, desc17, state17, map[string]any{
			"exp": exp, "digitPlace": digitName, "arr": slices.Clone(arr), "ids": slices.Clone(ids),
		})

		caption5 := fmt.Sprintf("Pass exp=%d (%s): group by %s: %s", exp, digitName, digitName, formatElementsWithDigit(arr, exp))
		desc5 := fmt.Sprintf("Completed pass for exp=%d (%s). Array stably ordered by %s: %s", exp, digitName, digitName, formatElementsWithDigit(arr, exp))
		b.AddStep("L5", trace.ActionHighlight, desc5, snapshot(caption5, -1), map[string]any{
			"exp": exp, "digitPlace": digitName, "arr": slices.Clone(arr), "ids": slices.Clone(ids),
		})
	}

	sorted = makeRange(0, n-1)
	captionDone := "Radix Sort completed: all digit passes finished and elements stably sorted"
	b.AddStep("L6", trace.ActionDone, "Radix Sort completed! All elements sorted.", snapshot(captionDone, -1), map[string]any{"status": "complete"})

	return b.Steps()
}

func digitPlaceName(exp int) string {
	switch exp {
	case 1:
		return "ones digit"
	case 10:
		return "tens digit"
	case 100:
		return "hundreds digit"
	case 1000:
		return "thousands digit"
	case 10000:
		return "ten thousands digit"
	default:
		return fmt.Sprintf("10^%d digit", intExp(exp))
	}
}

func intExp(exp int) int {
	p := 0
	for exp >= 10 {
		exp /= 10
		p++
	}
	return p
}

func formatElementsWithDigit(arr []int, exp int) string {
	parts := make([]string, len(arr))
	for i, v := range arr {
		parts[i] = fmt.Sprintf("%d(%d)", v, (v/exp)%10)
	}
	return "[" + strings.Join(parts, ", ") + "]"
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
