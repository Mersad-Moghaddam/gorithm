package linearsearch

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for linear search on initial slice with given target.
func Trace(initial []int, target int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)

	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("linearsearch-%d", i)
	}

	snapshot := func(caption string, current int, comparing []int, sorted []int) renderers.ArrayState {
		return renderers.ArrayState{
			IDs:       slices.Clone(ids),
			Caption:   caption,
			Values:    slices.Clone(arr),
			Comparing: slices.Clone(comparing),
			Sorted:    slices.Clone(sorted),
			Current:   current,
		}
	}

	initCaption := fmt.Sprintf("Initialize Linear Search for target %d in array of %d elements.", target, n)
	b.AddStep("L1", trace.ActionInit, initCaption, snapshot(initCaption, -1, nil, nil), map[string]any{"target": target, "n": n})

	if n == 0 {
		emptyCaption := fmt.Sprintf("Checked all elements: array is empty; target %d not found. Return -1.", target)
		b.AddStep("L7", trace.ActionDone, emptyCaption, snapshot(emptyCaption, -1, nil, nil), map[string]any{"target": target, "result": -1, "found": false})
		return b.Steps()
	}

	for i := 0; i < n; i++ {
		inspectCaption := fmt.Sprintf("Inspecting index %d (value %d) for target %d.", i, arr[i], target)
		b.AddStep("L2", trace.ActionHighlight, inspectCaption, snapshot(inspectCaption, i, nil, nil), map[string]any{"i": i, "target": target, "val": arr[i]})

		b.IncComparisons()
		if arr[i] == target {
			matchCaption := fmt.Sprintf("arr[%d]=%d == target=%d: match found!", i, arr[i], target)
			b.AddStep("L3", trace.ActionCompare, matchCaption, snapshot(matchCaption, i, []int{i}, nil), map[string]any{"i": i, "target": target, "val": arr[i], "match": true})

			foundCaption := fmt.Sprintf("Found target %d at index %d!", target, i)
			b.AddStep("L4", trace.ActionDone, foundCaption, snapshot(foundCaption, i, nil, []int{i}), map[string]any{"i": i, "target": target, "result": i, "found": true})
			return b.Steps()
		}

		mismatchCaption := fmt.Sprintf("arr[%d]=%d != target=%d; advance to next element", i, arr[i], target)
		b.AddStep("L3", trace.ActionCompare, mismatchCaption, snapshot(mismatchCaption, i, []int{i}, nil), map[string]any{"i": i, "target": target, "val": arr[i], "match": false})
	}

	notFoundCaption := fmt.Sprintf("Checked all %d elements; target %d not found. Return -1.", n, target)
	b.AddStep("L7", trace.ActionDone, notFoundCaption, snapshot(notFoundCaption, -1, nil, nil), map[string]any{"target": target, "n": n, "result": -1, "found": false})

	return b.Steps()
}
