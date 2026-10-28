package insertion

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: shifts counted as swaps for unified visual metrics; upgrade if shift-specific metric added.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := make([]int, len(initial))
	copy(arr, initial)
	n := len(arr)
	ids := make([]int, n)
	minValue, maxValue := 0, 0
	for i, value := range arr {
		ids[i] = i
		if value < minValue {
			minValue = value
		}
		if value > maxValue {
			maxValue = value
		}
	}
	pass, scan, keyValue, keyID, hole, prefixEnd := -1, -1, 0, -1, -1, -1
	if n > 0 {
		prefixEnd = 0
	}
	keyHeld := false
	addStep := func(line string, action trace.ActionType, explanation string, state renderers.ArrayState, vars map[string]any) {
		b.AddStep(line, action, explanation, renderers.InsertionLessonState{ArrayState: state, IDs: cloneSlice(ids), Pass: pass, Scan: scan, KeyValue: keyValue, KeyID: keyID, Hole: hole, PrefixEnd: prefixEnd, KeyHeld: keyHeld, MinValue: minValue, MaxValue: maxValue}, vars)
	}

	if n == 0 {
		addStep("L1", trace.ActionInit, "Initialize Insertion Sort with empty array.", renderers.ArrayState{
			Current: -1,
		}, map[string]any{"n": 0})
		addStep("L12", trace.ActionDone, "Insertion Sort completed! Array is empty.", renderers.ArrayState{
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	if n == 1 {
		addStep("L1", trace.ActionInit, "Initialize Insertion Sort with input array.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"n": 1})
		addStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"n": 1})
		addStep("L12", trace.ActionDone, "Insertion Sort completed! Single element is sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  []int{0},
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	addStep("L1", trace.ActionInit, "Initialize Insertion Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  []int{0},
		Current: -1,
	}, map[string]any{"n": n})

	addStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  []int{0},
		Current: -1,
	}, map[string]any{"n": n})

	for i := 1; i < n; i++ {
		pass, scan, keyID, hole, prefixEnd = i, -1, -1, -1, i-1
		addStep("L3", trace.ActionHighlight, fmt.Sprintf("Starting outer pass i = %d.", i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, i-1),
			Current: i,
		}, map[string]any{"i": i})

		key := arr[i]
		keyValue, keyID, hole, keyHeld = key, ids[i], i, true
		addStep("L4", trace.ActionSet, fmt.Sprintf("Picked key %d at index %d to insert into sorted sub-array.", key, i), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, i-1),
			Current: i,
		}, map[string]any{"i": i, "key": key})

		j := i - 1
		scan = j
		addStep("L5", trace.ActionSet, fmt.Sprintf("Initialized j = %d to compare backwards.", j), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, i-1),
			Current: i,
		}, map[string]any{"i": i, "key": key, "j": j})

		for {
			if j < 0 {
				addStep("L6", trace.ActionHighlight, "Index j < 0: reached start of array, ready to place key.", renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  makeRange(0, i-1),
					Current: i,
				}, map[string]any{"i": i, "key": key, "j": j})
				break
			}

			b.IncComparisons()
			if arr[j] > key {
				addStep("L6", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) > key (%d): condition is true, need to shift.", j, arr[j], key), renderers.ArrayState{
					Values:    cloneSlice(arr),
					Comparing: []int{j},
					Sorted:    makeRange(0, i-1, j),
					Current:   i,
				}, map[string]any{"i": i, "key": key, "j": j, "val": arr[j]})

				arr[j+1] = arr[j]
				ids[j+1] = ids[j]
				hole = j
				b.IncSwaps()
				addStep("L7", trace.ActionSwap, fmt.Sprintf("Shifted %d from index %d to index %d.", arr[j], j, j+1), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Swapped: []int{j, j + 1},
					Sorted:  makeRange(0, i-1, j, j+1),
					Current: i,
				}, map[string]any{"i": i, "key": key, "j": j})

				j--
				scan = j
				addStep("L8", trace.ActionSet, fmt.Sprintf("Decremented j to %d.", j), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Sorted:  makeRange(0, i-1),
					Current: i,
				}, map[string]any{"i": i, "key": key, "j": j})
			} else {
				addStep("L6", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) > key (%d): condition is false, found insertion position.", j, arr[j], key), renderers.ArrayState{
					Values:    cloneSlice(arr),
					Comparing: []int{j},
					Sorted:    makeRange(0, i-1, j),
					Current:   i,
				}, map[string]any{"i": i, "key": key, "j": j, "val": arr[j]})
				break
			}
		}

		arr[j+1] = key
		ids[j+1] = keyID
		keyHeld, hole, prefixEnd = false, -1, i
		b.IncOperations()
		addStep("L10", trace.ActionSet, fmt.Sprintf("Inserted key %d at index %d.", key, j+1), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  makeRange(0, i),
			Current: j + 1,
		}, map[string]any{"i": i, "key": key, "j": j, "insertedAt": j + 1})
	}

	pass, scan, keyID = -1, -1, -1
	addStep("L12", trace.ActionDone, "Insertion Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  makeRange(0, n-1),
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

func makeRange(start, end int, exclude ...int) []int {
	if start > end {
		return []int{}
	}
	ex := make(map[int]bool, len(exclude))
	for _, x := range exclude {
		ex[x] = true
	}
	res := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		if !ex[i] {
			res = append(res, i)
		}
	}
	return res
}
