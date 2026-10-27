package merge

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)

	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("merge-%d", i)
	}

	snapshot := func(caption string, current int, comparing []int, swapped []int, sorted []int) renderers.ArrayState {
		return renderers.ArrayState{
			Values:    slices.Clone(arr),
			IDs:       slices.Clone(ids),
			Sorted:    slices.Clone(sorted),
			Current:   current,
			Comparing: slices.Clone(comparing),
			Swapped:   slices.Clone(swapped),
			Caption:   caption,
		}
	}

	b.AddStep("L1", trace.ActionInit, "Divide array into halves, sort recursively, and merge sorted runs.", snapshot("Divide array into halves, sort recursively, and merge sorted runs", -1, nil, nil, nil), map[string]any{"n": n})

	if n == 0 {
		b.AddStep("L9", trace.ActionDone, "Merge Sort completed! Empty array.", snapshot("Empty array: nothing to sort", -1, nil, nil, nil), map[string]any{"status": "complete"})
		return b.Steps()
	}

	if n == 1 {
		b.AddStep("L2", trace.ActionHighlight, "Subarray [0..0] has length 1: base case reached.", snapshot("Subarray [0..0] has length 1: already sorted", 0, nil, nil, nil), map[string]any{"left": 0, "right": 0})
		b.AddStep("L3", trace.ActionVisit, "Return base case subarray [0..0].", snapshot("Base case: [0..0] is sorted", 0, nil, nil, []int{0}), map[string]any{"left": 0, "right": 0})
		b.AddStep("L9", trace.ActionDone, "Merge Sort completed! Single element sorted.", snapshot("Merge Sort completed! Single element sorted.", -1, nil, nil, []int{0}), map[string]any{"status": "complete"})
		return b.Steps()
	}

	var mergeSortTraced func(left, right int)
	mergeSortTraced = func(left, right int) {
		b.IncOperations()
		b.AddStep("L2", trace.ActionHighlight, fmt.Sprintf("Checking range [%d..%d]: left (%d) >= right (%d)?", left, right, left, right), snapshot(fmt.Sprintf("Check range [%d..%d]", left, right), -1, nil, nil, nil), map[string]any{"left": left, "right": right})

		if left >= right {
			b.AddStep("L3", trace.ActionVisit, fmt.Sprintf("Subarray [%d..%d] has length 1: base case reached.", left, right), snapshot(fmt.Sprintf("Base case: [%d..%d] has length 1 (sorted)", left, right), left, nil, nil, nil), map[string]any{"left": left, "right": right})
			return
		}

		mid := left + (right-left)/2
		b.IncOperations()
		captionSplit := fmt.Sprintf("Split range [%d..%d] at mid %d -> left [%d..%d], right [%d..%d]", left, right, mid, left, mid, mid+1, right)
		b.AddStep("L5", trace.ActionSet, captionSplit, snapshot(captionSplit, mid, nil, nil, nil), map[string]any{"left": left, "mid": mid, "right": right})

		b.AddStep("L6", trace.ActionVisit, fmt.Sprintf("Recursively calling MergeSort on left half [%d..%d].", left, mid), snapshot(fmt.Sprintf("Recurse left half [%d..%d]", left, mid), mid, nil, nil, nil), map[string]any{"left": left, "mid": mid})
		mergeSortTraced(left, mid)

		b.AddStep("L7", trace.ActionVisit, fmt.Sprintf("Recursively calling MergeSort on right half [%d..%d].", mid+1, right), snapshot(fmt.Sprintf("Recurse right half [%d..%d]", mid+1, right), mid+1, nil, nil, nil), map[string]any{"mid+1": mid + 1, "right": right})
		mergeSortTraced(mid+1, right)

		captionPrep := fmt.Sprintf("Merge sorted runs [%d..%d] and [%d..%d] into [%d..%d]", left, mid, mid+1, right, left, right)
		b.AddStep("L8", trace.ActionHighlight, captionPrep, snapshot(captionPrep, mid, nil, nil, nil), map[string]any{"left": left, "mid": mid, "right": right})

		captionMerge := fmt.Sprintf("Begin merge: source runs [%d..%d] and [%d..%d] -> destination [%d..%d]", left, mid, mid+1, right, left, right)
		b.AddStep("L10", trace.ActionVisit, captionMerge, snapshot(captionMerge, -1, nil, nil, nil), map[string]any{"left": left, "mid": mid, "right": right})

		aux := make([]int, right-left+1)
		auxIDs := make([]string, right-left+1)
		b.IncOperations()
		b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Created auxiliary buffer of size %d.", len(aux)), snapshot(fmt.Sprintf("Allocate aux buffer of size %d", len(aux)), -1, nil, nil, nil), map[string]any{"aux_size": len(aux)})

		i, j, k := left, mid+1, 0
		b.IncOperations()
		b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Initialized pointers: left run at [%d], right run at [%d]", left, mid+1), snapshot(fmt.Sprintf("Initialize pointers: left [%d], right [%d]", i, j), -1, nil, nil, nil), map[string]any{"i": i, "j": j, "k": k})

		for i <= mid && j <= right {
			b.IncComparisons()
			var cmpCaption string
			if arr[i] < arr[j] {
				cmpCaption = fmt.Sprintf("Compare left [%d]=%d with right [%d]=%d; pick smaller %d", i, arr[i], j, arr[j], arr[i])
			} else if arr[i] == arr[j] {
				cmpCaption = fmt.Sprintf("Compare left [%d]=%d with right [%d]=%d; pick left %d to maintain stability", i, arr[i], j, arr[j], arr[i])
			} else {
				cmpCaption = fmt.Sprintf("Compare left [%d]=%d with right [%d]=%d; pick smaller %d", i, arr[i], j, arr[j], arr[j])
			}

			b.AddStep("L14", trace.ActionCompare, cmpCaption, snapshot(cmpCaption, -1, []int{i, j}, nil, nil), map[string]any{"i": i, "j": j, "val_i": arr[i], "val_j": arr[j]})

			if arr[i] <= arr[j] {
				aux[k] = arr[i]
				auxIDs[k] = ids[i]
				b.IncOperations()
				pickCaption := fmt.Sprintf("Pick left [%d]=%d into aux[%d]", i, arr[i], k)
				b.AddStep("L15", trace.ActionSet, pickCaption, snapshot(pickCaption, i, nil, []int{i}, nil), map[string]any{"i": i, "val": arr[i], "k": k})
				i++
			} else {
				aux[k] = arr[j]
				auxIDs[k] = ids[j]
				b.IncOperations()
				pickCaption := fmt.Sprintf("Pick right [%d]=%d into aux[%d]", j, arr[j], k)
				b.AddStep("L17", trace.ActionSet, pickCaption, snapshot(pickCaption, j, nil, []int{j}, nil), map[string]any{"j": j, "val": arr[j], "k": k})
				j++
			}
			k++
		}

		for i <= mid {
			aux[k] = arr[i]
			auxIDs[k] = ids[i]
			b.IncOperations()
			leftCaption := fmt.Sprintf("Copy remaining left element [%d]=%d to aux[%d]", i, arr[i], k)
			b.AddStep("L21", trace.ActionSet, leftCaption, snapshot(leftCaption, i, nil, []int{i}, nil), map[string]any{"i": i, "val": arr[i], "k": k})
			i++
			k++
		}

		for j <= right {
			aux[k] = arr[j]
			auxIDs[k] = ids[j]
			b.IncOperations()
			rightCaption := fmt.Sprintf("Copy remaining right element [%d]=%d to aux[%d]", j, arr[j], k)
			b.AddStep("L22", trace.ActionSet, rightCaption, snapshot(rightCaption, j, nil, []int{j}, nil), map[string]any{"j": j, "val": arr[j], "k": k})
			j++
			k++
		}

		for p := 0; p < len(aux); p++ {
			dest := left + p
			targetID := auxIDs[p]
			loc := dest
			for idx := dest; idx <= right; idx++ {
				if ids[idx] == targetID {
					loc = idx
					break
				}
			}
			arr[dest], arr[loc] = arr[loc], arr[dest]
			ids[dest], ids[loc] = ids[loc], ids[dest]
			b.IncOperations()
			placeCaption := fmt.Sprintf("Place %d into merged range at [%d]", arr[dest], dest)
			b.AddStep("L23", trace.ActionSet, placeCaption, snapshot(placeCaption, dest, nil, []int{dest}, nil), map[string]any{"index": dest, "val": arr[dest]})
		}

		doneMergeCaption := fmt.Sprintf("Merged range [%d..%d]: %v", left, right, arr[left:right+1])
		b.AddStep("L24", trace.ActionHighlight, doneMergeCaption, snapshot(doneMergeCaption, -1, nil, nil, nil), map[string]any{"left": left, "right": right})
	}

	mergeSortTraced(0, n-1)

	allSorted := make([]int, n)
	for i := 0; i < n; i++ {
		allSorted[i] = i
	}

	b.AddStep("L9", trace.ActionDone, "Merge Sort completed! All elements sorted.", snapshot("Merge Sort completed! All elements sorted.", -1, nil, nil, allSorted), map[string]any{"status": "complete"})

	return b.Steps()
}
