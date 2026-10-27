package merge

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

	b.AddStep("L1", trace.ActionInit, "Initialize Merge Sort with input array.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Current: -1,
	}, map[string]any{"n": n})

	if n <= 1 {
		sorted := make([]int, n)
		for i := range sorted {
			sorted[i] = i
		}
		b.AddStep("L9", trace.ActionDone, "Merge Sort completed! Array is sorted.", renderers.ArrayState{
			Values:  cloneSlice(arr),
			Sorted:  sorted,
			Current: -1,
		}, map[string]any{"status": "complete"})
		return b.Steps()
	}

	var mergeSortTraced func(left, right int)
	mergeSortTraced = func(left, right int) {
		b.IncOperations()
		b.AddStep("L2", trace.ActionHighlight, fmt.Sprintf("Checking if left (%d) >= right (%d).", left, right), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"left": left, "right": right})

		if left >= right {
			b.AddStep("L3", trace.ActionVisit, fmt.Sprintf("Subarray [%d..%d] has length <= 1, base case reached.", left, right), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: left,
			}, map[string]any{"left": left, "right": right})
			return
		}

		mid := left + (right-left)/2
		b.IncOperations()
		b.AddStep("L5", trace.ActionSet, fmt.Sprintf("Calculated mid = %d for subarray [%d..%d].", mid, left, right), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: mid,
		}, map[string]any{"left": left, "mid": mid, "right": right})

		b.AddStep("L6", trace.ActionVisit, fmt.Sprintf("Recursively calling MergeSort on left half [%d..%d].", left, mid), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: mid,
		}, map[string]any{"left": left, "mid": mid})
		mergeSortTraced(left, mid)

		b.AddStep("L7", trace.ActionVisit, fmt.Sprintf("Recursively calling MergeSort on right half [%d..%d].", mid+1, right), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: mid + 1,
		}, map[string]any{"mid+1": mid + 1, "right": right})
		mergeSortTraced(mid+1, right)

		b.AddStep("L8", trace.ActionHighlight, fmt.Sprintf("Preparing to merge subarrays [%d..%d] and [%d..%d].", left, mid, mid+1, right), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"left": left, "mid": mid, "right": right})

		b.AddStep("L10", trace.ActionVisit, fmt.Sprintf("Begin merge: [%d..%d] and [%d..%d].", left, mid, mid+1, right), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"left": left, "mid": mid, "right": right})

		aux := make([]int, right-left+1)
		b.IncOperations()
		b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Created auxiliary array of size %d.", len(aux)), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"aux_size": len(aux)})

		i, j, k := left, mid+1, 0
		b.IncOperations()
		b.AddStep("L12", trace.ActionSet, fmt.Sprintf("Initialized pointers: i = %d (left), j = %d (right), k = %d.", i, j, k), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"i": i, "j": j, "k": k})

		for i <= mid && j <= right {
			b.IncComparisons()
			b.AddStep("L14", trace.ActionCompare, fmt.Sprintf("Comparing arr[%d] (%d) and arr[%d] (%d).", i, arr[i], j, arr[j]), renderers.ArrayState{
				Values:    cloneSlice(arr),
				Comparing: []int{i, j},
				Current:   -1,
			}, map[string]any{"i": i, "j": j, "val_i": arr[i], "val_j": arr[j]})

			if arr[i] <= arr[j] {
				aux[k] = arr[i]
				b.IncOperations()
				b.AddStep("L15", trace.ActionSet, fmt.Sprintf("arr[%d] (%d) <= arr[%d] (%d): placing %d into aux[%d].", i, arr[i], j, arr[j], arr[i], k), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Current: i,
				}, map[string]any{"i": i, "val": arr[i], "k": k})
				i++
			} else {
				aux[k] = arr[j]
				b.IncOperations()
				b.AddStep("L17", trace.ActionSet, fmt.Sprintf("arr[%d] (%d) > arr[%d] (%d): placing %d into aux[%d].", i, arr[i], j, arr[j], arr[j], k), renderers.ArrayState{
					Values:  cloneSlice(arr),
					Current: j,
				}, map[string]any{"j": j, "val": arr[j], "k": k})
				j++
			}
			k++
		}

		for i <= mid {
			aux[k] = arr[i]
			b.IncOperations()
			b.AddStep("L21", trace.ActionSet, fmt.Sprintf("Copying remaining left element arr[%d] (%d) to aux[%d].", i, arr[i], k), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: i,
			}, map[string]any{"i": i, "val": arr[i], "k": k})
			i++
			k++
		}

		for j <= right {
			aux[k] = arr[j]
			b.IncOperations()
			b.AddStep("L22", trace.ActionSet, fmt.Sprintf("Copying remaining right element arr[%d] (%d) to aux[%d].", j, arr[j], k), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: j,
			}, map[string]any{"j": j, "val": arr[j], "k": k})
			j++
			k++
		}

		for p := 0; p < len(aux); p++ {
			arr[left+p] = aux[p]
			b.IncOperations()
			b.AddStep("L23", trace.ActionSet, fmt.Sprintf("Copied aux[%d] (%d) back to arr[%d].", p, aux[p], left+p), renderers.ArrayState{
				Values:  cloneSlice(arr),
				Current: left + p,
				Swapped: []int{left + p},
			}, map[string]any{"index": left + p, "val": aux[p]})
		}

		b.AddStep("L24", trace.ActionHighlight, fmt.Sprintf("Merged subarray [%d..%d]: %v.", left, right, arr[left:right+1]), renderers.ArrayState{
			Values:  cloneSlice(arr),
			Current: -1,
		}, map[string]any{"left": left, "right": right})
	}

	mergeSortTraced(0, n-1)

	allSorted := make([]int, n)
	for i := 0; i < n; i++ {
		allSorted[i] = i
	}

	b.AddStep("L9", trace.ActionDone, "Merge Sort completed! All elements sorted.", renderers.ArrayState{
		Values:  cloneSlice(arr),
		Sorted:  allSorted,
		Current: -1,
	}, map[string]any{"status": "complete"})

	return b.Steps()
}

func cloneSlice(s []int) []int {
	c := make([]int, len(s))
	copy(c, s)
	return c
}
