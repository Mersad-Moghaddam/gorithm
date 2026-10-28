package countingsort

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: direct Trace requires nonnegative bounded values; Lesson.Build validates browser input.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := cloneSlice(initial)
	n := len(arr)
	s := renderers.CountingsortLessonState{Input: arr, InputIDs: makeRange(0, n-1), Max: -1, Scan: -1, CountIndex: -1, Target: -1, Phase: "init"}
	add := func(line string, action trace.ActionType, text string, legacy renderers.ArrayState, vars map[string]any) {
		v := s
		v.ArrayState = legacy
		v.Input, v.InputIDs, v.Counts = cloneSlice(s.Input), cloneSlice(s.InputIDs), cloneSlice(s.Counts)
		v.Output, v.OutputIDs = cloneSlice(s.Output), cloneSlice(s.OutputIDs)
		v.OutputFilled = append([]bool(nil), s.OutputFilled...)
		b.AddStep(line, action, text, v, vars)
	}
	state := func(values []int, current int, sorted []int) renderers.ArrayState {
		return renderers.ArrayState{Values: cloneSlice(values), Current: current, Sorted: cloneSlice(sorted)}
	}
	if n <= 1 {
		sorted := makeRange(0, n-1)
		add("L1", trace.ActionInit, "Initialize Counting Sort.", state(arr, -1, sorted), map[string]any{"n": n})
		if n == 1 {
			add("L2", trace.ActionSet, "Array length is 1.", state(arr, -1, sorted), map[string]any{"n": n})
		}
		add("L3", trace.ActionHighlight, "At most one element: already sorted.", state(arr, -1, sorted), map[string]any{"n": n})
		s.Phase = "done"
		s.Output, s.OutputIDs, s.OutputFilled = cloneSlice(arr), makeRange(0, n-1), make([]bool, n)
		for i := range s.OutputFilled {
			s.OutputFilled[i] = true
		}
		add("L4", trace.ActionDone, "Counting Sort completed.", state(arr, -1, sorted), map[string]any{"status": "complete"})
		return b.Steps()
	}
	add("L1", trace.ActionInit, "Initialize Counting Sort with input array.", state(arr, -1, nil), map[string]any{"n": n})
	add("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), state(arr, -1, nil), map[string]any{"n": n})
	add("L3", trace.ActionHighlight, "More than one element: find the value range.", state(arr, -1, nil), map[string]any{"n": n})
	s.Phase, s.Max, s.Scan = "maximum", arr[0], 0
	b.IncOperations()
	add("L6", trace.ActionSet, fmt.Sprintf("Initialize maximum to %d.", s.Max), state(arr, 0, nil), map[string]any{"max": s.Max, "i": 0})
	for i := 1; i < n; i++ {
		s.Scan = i
		vars := map[string]any{"i": i, "val": arr[i], "max": s.Max}
		add("L7", trace.ActionHighlight, fmt.Sprintf("Inspect input[%d] = %d for maximum.", i, arr[i]), state(arr, i, nil), vars)
		b.IncComparisons()
		comp := state(arr, i, nil)
		comp.Comparing = []int{i}
		add("L8", trace.ActionCompare, fmt.Sprintf("Compare %d with maximum %d.", arr[i], s.Max), comp, vars)
		if arr[i] > s.Max {
			s.Max = arr[i]
			b.IncOperations()
			add("L9", trace.ActionSet, fmt.Sprintf("Maximum becomes %d.", s.Max), state(arr, i, nil), map[string]any{"i": i, "max": s.Max})
		}
	}
	s.Phase, s.Scan = "tally", -1
	s.Counts = make([]int, s.Max+1)
	b.IncOperations()
	add("L12", trace.ActionSet, fmt.Sprintf("Allocate counts for values 0 through %d.", s.Max), state(arr, -1, nil), map[string]any{"max": s.Max, "countSize": s.Max + 1, "count": cloneSlice(s.Counts)})
	for i, value := range arr {
		s.Scan, s.CountIndex = i, value
		add("L13", trace.ActionHighlight, fmt.Sprintf("Tally input[%d] = %d.", i, value), state(arr, i, nil), map[string]any{"i": i, "val": value, "count": cloneSlice(s.Counts)})
		s.Counts[value]++
		b.IncOperations()
		add("L14", trace.ActionSet, fmt.Sprintf("Frequency[%d] becomes %d.", value, s.Counts[value]), state(arr, i, nil), map[string]any{"i": i, "val": value, "count": cloneSlice(s.Counts), "frequencies": cloneSlice(s.Counts)})
	}
	s.Phase, s.Scan, s.CountIndex = "prefix", -1, -1
	add("L16", trace.ActionHighlight, "Accumulate frequencies: number of items <= each value.", state(arr, -1, nil), map[string]any{"count": cloneSlice(s.Counts)})
	for i := 1; i <= s.Max; i++ {
		s.CountIndex = i
		s.Counts[i] += s.Counts[i-1]
		b.IncOperations()
		add("L17", trace.ActionSet, fmt.Sprintf("count[%d] += count[%d]: %d items <= %d.", i, i-1, s.Counts[i], i), state(arr, -1, nil), map[string]any{"i": i, "count": cloneSlice(s.Counts), "prefixSums": cloneSlice(s.Counts)})
	}
	s.Phase, s.CountIndex = "output", -1
	s.Output, s.OutputIDs, s.OutputFilled = make([]int, n), make([]int, n), make([]bool, n)
	for i := range s.OutputIDs {
		s.OutputIDs[i] = -1
	}
	b.IncOperations()
	add("L19", trace.ActionSet, "Allocate output: empty slots are not zero-valued items.", state(s.Output, -1, nil), map[string]any{"n": n, "count": cloneSlice(s.Counts), "output": cloneSlice(s.Output)})
	sorted := make([]int, 0, n)
	for i := n - 1; i >= 0; i-- {
		value := arr[i]
		s.Scan, s.CountIndex, s.Target = i, value, -1
		add("L20", trace.ActionHighlight, fmt.Sprintf("Read %d#%d from right to preserve stability.", value, i), state(s.Output, -1, sorted), map[string]any{"i": i, "val": value, "count": cloneSlice(s.Counts)})
		s.Target = s.Counts[value] - 1
		s.Output[s.Target], s.OutputIDs[s.Target], s.OutputFilled[s.Target] = value, i, true
		sorted = append(sorted, s.Target)
		b.IncOperations()
		placed := state(s.Output, s.Target, sorted)
		placed.Swapped = []int{s.Target}
		add("L21", trace.ActionSet, fmt.Sprintf("Copy %d#%d to output[%d] = count[%d]-1.", value, i, s.Target, value), placed, map[string]any{"i": i, "val": value, "targetIndex": s.Target, "count": cloneSlice(s.Counts), "output": cloneSlice(s.Output)})
		s.Counts[value]--
		b.IncOperations()
		add("L22", trace.ActionSet, fmt.Sprintf("Decrement count[%d] to %d: remaining insertion boundary.", value, s.Counts[value]), state(s.Output, s.Target, sorted), map[string]any{"i": i, "val": value, "count": cloneSlice(s.Counts)})
	}
	s.Phase, s.Scan, s.CountIndex, s.Target = "done", -1, -1, -1
	add("L24", trace.ActionDone, "Counting Sort completed: equal values retain original order.", state(s.Output, -1, makeRange(0, n-1)), map[string]any{"status": "complete", "result": cloneSlice(s.Output)})
	steps := b.Steps()
	for i := 0; i+1 < len(steps); i++ {
		if steps[i].LineID == "L8" && steps[i+1].LineID == "L9" {
			steps[i].HoldMS, steps[i+1].HoldMS = 600, 1200
		}
	}
	return steps
}

func cloneSlice(s []int) []int { c := make([]int, len(s)); copy(c, s); return c }
func makeRange(start, end int) []int {
	if start > end {
		return []int{}
	}
	r := make([]int, end-start+1)
	for i := range r {
		r[i] = start + i
	}
	return r
}
