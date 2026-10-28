package radixsort

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := cloneSlice(initial)
	n := len(arr)
	ids := makeRange(0, n-1)
	var output, outputIDs []int
	var filled []bool
	var count [10]int
	exponent, scan, digit, target, copyIndex := 0, -1, -1, -1, -1
	domainMax := 0
	for _, v := range initial {
		if v > domainMax {
			domainMax = v
		}
	}
	add := func(line string, action trace.ActionType, explanation string, current int, variables map[string]any) {
		phase := map[string]string{"L1": "init", "L2": "maximum", "L3": "pass", "L4": "pass", "L7": "pass", "L8": "pass", "L9": "pass", "L10": "tally", "L11": "prefix", "L12": "output", "L13": "output", "L14": "output", "L15": "output", "L17": "copy", "L5": "pass-complete", "L6": "done"}[line]
		base := renderers.ArrayState{Values: cloneSlice(arr), Current: current}
		if line == "L6" {
			base.Sorted = makeRange(0, n-1)
		}
		if line == "L17" {
			base.Swapped = []int{copyIndex}
		}
		b.AddStep(line, action, explanation, renderers.RadixsortLessonState{ArrayState: base, IDs: cloneSlice(ids), Output: cloneSlice(output), OutputIDs: cloneSlice(outputIDs), OutputFilled: append([]bool(nil), filled...), Counts: count, Exp: exponent, MaxValue: domainMax, Scan: scan, Digit: digit, Target: target, CopyIndex: copyIndex, Phase: phase}, variables)
	}
	add("L1", trace.ActionInit, "Initialize Radix Sort (LSD) with input array.", -1, map[string]any{"n": n})
	if n <= 1 {
		add("L6", trace.ActionDone, "Radix Sort completed! Array has at most one element.", -1, map[string]any{"status": "complete"})
		return b.Steps()
	}
	maxVal, maxIdx := arr[0], 0
	for i, v := range arr {
		b.IncOperations()
		if v > maxVal {
			maxVal, maxIdx = v, i
		}
	}
	add("L2", trace.ActionSet, fmt.Sprintf("Found maximum value %d at index %d to determine digit count.", maxVal, maxIdx), maxIdx, map[string]any{"maxVal": maxVal, "maxIdx": maxIdx})
	for exp := 1; maxVal/exp > 0; exp *= 10 {
		exponent = exp
		scan, digit, target, copyIndex = -1, -1, -1, -1
		// New pass has no output storage until L8, and no counters until L9.
		output, outputIDs, filled = nil, nil, nil
		count = [10]int{}
		b.IncOperations()
		add("L3", trace.ActionHighlight, fmt.Sprintf("Processing digit place %d (exp = %d).", exp, exp), -1, map[string]any{"exp": exp, "maxVal": maxVal})
		add("L4", trace.ActionVisit, fmt.Sprintf("Invoking counting sort for exponent %d.", exp), -1, map[string]any{"exp": exp})
		add("L7", trace.ActionHighlight, fmt.Sprintf("Starting counting sort for exp = %d.", exp), -1, map[string]any{"exp": exp, "n": n})
		output = make([]int, n)
		outputIDs = make([]int, n)
		filled = make([]bool, n)
		for i := range outputIDs {
			outputIDs[i] = -1
		}
		add("L8", trace.ActionSet, fmt.Sprintf("Allocated output buffer of size %d.", n), -1, map[string]any{"exp": exp, "n": n})
		add("L9", trace.ActionSet, "Initialized 10 digit buckets (0 through 9) to zero.", -1, map[string]any{"exp": exp})
		for i := 0; i < n; i++ {
			scan = i
			digit = (arr[i] / exp) % 10
			count[digit]++
			b.IncOperations()
			add("L10", trace.ActionSet, fmt.Sprintf("arr[%d]=%d has digit %d (exp=%d). count[%d] is now %d.", i, arr[i], digit, exp, digit, count[digit]), i, map[string]any{"exp": exp, "i": i, "val": arr[i], "digit": digit, "count": count[digit]})
		}
		for i := 1; i < 10; i++ {
			count[i] += count[i-1]
			b.IncOperations()
		}
		scan, digit = -1, -1
		add("L11", trace.ActionSet, fmt.Sprintf("Accumulated prefix sums in count array for exp=%d.", exp), -1, map[string]any{"exp": exp})
		for i := n - 1; i >= 0; i-- {
			scan = i
			digit, target = -1, -1
			b.IncOperations()
			add("L12", trace.ActionHighlight, fmt.Sprintf("Processing arr[%d] = %d backwards for stability.", i, arr[i]), i, map[string]any{"exp": exp, "i": i, "val": arr[i]})
			digit = (arr[i] / exp) % 10
			add("L13", trace.ActionSet, fmt.Sprintf("Extracted digit %d from arr[%d]=%d.", digit, i, arr[i]), i, map[string]any{"exp": exp, "i": i, "digit": digit, "val": arr[i]})
			target = count[digit] - 1
			output[target] = arr[i]
			outputIDs[target] = ids[i]
			filled[target] = true
			add("L14", trace.ActionSet, fmt.Sprintf("Placed %d into output[%d] based on prefix sum.", arr[i], target), i, map[string]any{"exp": exp, "i": i, "val": arr[i], "digit": digit, "targetPos": target})
			count[digit]--
			add("L15", trace.ActionSet, fmt.Sprintf("Decremented count[%d] to %d.", digit, count[digit]), i, map[string]any{"exp": exp, "digit": digit, "remainingCount": count[digit]})
		}
		scan, digit, target = -1, -1, -1
		for i := 0; i < n; i++ {
			copyIndex = i
			arr[i] = output[i]
			ids[i] = outputIDs[i]
			b.IncSwaps()
			add("L17", trace.ActionSwap, fmt.Sprintf("Updated arr[%d] = %d from output (copy, not exchange).", i, arr[i]), i, map[string]any{"exp": exp, "i": i, "val": arr[i]})
		}
		copyIndex = -1
		add("L5", trace.ActionHighlight, fmt.Sprintf("Completed pass for exp = %d.", exp), -1, map[string]any{"exp": exp})
	}
	add("L6", trace.ActionDone, "Radix Sort completed! All elements sorted.", -1, map[string]any{"status": "complete"})
	return b.Steps()
}

func cloneSlice(s []int) []int { c := make([]int, len(s)); copy(c, s); return c }
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
