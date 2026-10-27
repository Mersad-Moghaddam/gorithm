package insertion

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: shifts counted as swaps for unified visual metrics; upgrade if shift-specific metric added.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)

	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("insertion-%d", i)
	}

	snapshot := func(caption string, current int, comparing []int, swapped []int, sorted []int) renderers.ArrayState {
		return renderers.ArrayState{
			Values:    slices.Clone(arr),
			IDs:       slices.Clone(ids),
			Sorted:    slices.Clone(sorted),
			Comparing: slices.Clone(comparing),
			Swapped:   slices.Clone(swapped),
			Current:   current,
			Caption:   caption,
		}
	}

	prefixExcept := func(maxIdx int, exclude ...int) []int {
		if maxIdx < 0 {
			return []int{}
		}
		ex := make(map[int]bool, len(exclude))
		for _, x := range exclude {
			ex[x] = true
		}
		res := make([]int, 0, maxIdx+1)
		for k := 0; k <= maxIdx; k++ {
			if !ex[k] {
				res = append(res, k)
			}
		}
		return res
	}

	initSorted := []int{}
	initCaption := "Empty array: nothing to sort"
	if n > 0 {
		initSorted = []int{0}
		initCaption = "Insert next element into growing green prefix"
	}

	b.AddStep("L1", trace.ActionInit, "Insert each element into the sorted prefix to its left.", snapshot(initCaption, -1, nil, nil, initSorted), map[string]any{"n": n})
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), snapshot(fmt.Sprintf("%d values to sort", n), -1, nil, nil, initSorted), map[string]any{"n": n})

	for i := 1; i < n; i++ {
		caption := fmt.Sprintf("Pass %d: insert arr[%d] = %d into prefix [0..%d]", i, i, arr[i], i-1)
		b.AddStep("L3", trace.ActionHighlight, caption, snapshot(caption, i, nil, nil, prefixExcept(i-1)), map[string]any{"i": i})

		key := arr[i]
		keyPos := i
		caption = fmt.Sprintf("Pick key %d at [%d] to insert into sorted prefix", key, i)
		b.AddStep("L4", trace.ActionSet, caption, snapshot(caption, keyPos, nil, nil, prefixExcept(i-1)), map[string]any{"i": i, "key": key})

		j := i - 1
		caption = fmt.Sprintf("Start comparing backwards from index %d", j)
		b.AddStep("L5", trace.ActionSet, caption, snapshot(caption, keyPos, nil, nil, prefixExcept(i-1)), map[string]any{"i": i, "key": key, "j": j})

		for {
			if j < 0 {
				caption = fmt.Sprintf("j < 0: reached start; place key %d at index 0", key)
				b.AddStep("L6", trace.ActionHighlight, caption, snapshot(caption, keyPos, nil, nil, prefixExcept(i, keyPos)), map[string]any{"i": i, "key": key, "j": j})
				break
			}

			b.IncComparisons()
			if arr[j] > key {
				caption = fmt.Sprintf("Compare: arr[%d] (%d) > key (%d); must shift right", j, arr[j], key)
				b.AddStep("L6", trace.ActionCompare, caption, snapshot(caption, keyPos, []int{j}, nil, prefixExcept(i, keyPos, j)), map[string]any{"i": i, "key": key, "j": j, "val": arr[j]})

				arr[j], arr[j+1] = arr[j+1], arr[j]
				ids[j], ids[j+1] = ids[j+1], ids[j]
				keyPos = j
				b.IncSwaps()

				caption = fmt.Sprintf("Shift: moved %d from [%d] to [%d]", arr[j+1], j, j+1)
				b.AddStep("L7", trace.ActionSwap, caption, snapshot(caption, keyPos, nil, []int{j, j + 1}, prefixExcept(i, j, j+1)), map[string]any{"i": i, "key": key, "j": j})

				j--
				caption = fmt.Sprintf("Decremented j to %d", j)
				b.AddStep("L8", trace.ActionSet, caption, snapshot(caption, keyPos, nil, nil, prefixExcept(i, keyPos)), map[string]any{"i": i, "key": key, "j": j})
			} else {
				caption = fmt.Sprintf("Compare: arr[%d] (%d) <= key (%d); found insertion position", j, arr[j], key)
				if arr[j] == key {
					caption = fmt.Sprintf("Compare: arr[%d] (%d) == key (%d); keep relative order for stability", j, arr[j], key)
				}
				b.AddStep("L6", trace.ActionCompare, caption, snapshot(caption, keyPos, []int{j}, nil, prefixExcept(i, keyPos, j)), map[string]any{"i": i, "key": key, "j": j, "val": arr[j]})
				break
			}
		}

		b.IncOperations()
		caption = fmt.Sprintf("Inserted key %d at index %d", key, j+1)
		b.AddStep("L10", trace.ActionSet, caption, snapshot(caption, keyPos, nil, nil, prefixExcept(i, keyPos)), map[string]any{"i": i, "key": key, "j": j, "insertedAt": j + 1})

		caption = fmt.Sprintf("Sorted prefix grew to [0..%d]", i)
		b.AddStep("L11", trace.ActionSet, caption, snapshot(caption, -1, nil, nil, prefixExcept(i)), map[string]any{"i": i, "sortedPrefix": i})
	}

	caption := "Sorted: every position settled"
	if n == 0 {
		caption = "Empty array: nothing to sort"
	}
	allSorted := []int{}
	if n > 0 {
		allSorted = prefixExcept(n - 1)
	}
	b.AddStep("L12", trace.ActionDone, caption, snapshot(caption, -1, nil, nil, allSorted), map[string]any{"status": "complete"})

	return b.Steps()
}
