package heap

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Heap Sort on the initial array.
// ponytail: in-place 0-indexed binary max-heap trace; upgrade to ternary or d-ary heap if visual demonstration needs alternative arities.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := slices.Clone(initial)
	n := len(arr)
	ids := make([]string, n)
	for i := range ids {
		ids[i] = fmt.Sprintf("heap-%d", i)
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

	b.AddStep("L1", trace.ActionInit, "Initialize Heap Sort with input array.", snapshot("Initialize Heap Sort: convert array to max-heap, then extract root to sorted suffix", -1), map[string]any{"n": n})

	if n == 0 {
		b.AddStep("L10", trace.ActionDone, "Heap Sort completed! Empty array.", snapshot("Empty array: already sorted", -1), map[string]any{"status": "complete"})
		return b.Steps()
	}

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), snapshot(fmt.Sprintf("%d elements to sort", n), -1), map[string]any{"n": n})

	if n == 1 {
		sorted = append(sorted, 0)
		b.AddStep("L10", trace.ActionDone, "Heap Sort completed! Single element sorted.", snapshot("Single element: already sorted", -1), map[string]any{"status": "complete"})
		return b.Steps()
	}

	// Phase 1: Build max-heap
	for i := n/2 - 1; i >= 0; i-- {
		captionSubtree := fmt.Sprintf("Build max-heap: inspect subtree rooted at parent [%d]", i)
		b.AddStep("L3", trace.ActionHighlight, captionSubtree, snapshot(captionSubtree, i), map[string]any{"i": i})

		captionHeapify := fmt.Sprintf("Heapify subtree at parent [%d] on heap size %d.", i, n)
		b.AddStep("L4", trace.ActionVisit, captionHeapify, snapshot(captionHeapify, i), map[string]any{"i": i, "n": n})

		heapifyTrace(b, arr, ids, n, i, sorted)
	}

	// Phase 2: Extract elements from heap
	for i := n - 1; i > 0; i-- {
		captionExtract := fmt.Sprintf("Extract max: root [0] (%d) is maximum of active heap; swap with last heap element [%d] (%d).", arr[0], i, arr[i])
		stateExtract := snapshot(captionExtract, 0)
		stateExtract.Comparing = []int{0, i}
		b.AddStep("L6", trace.ActionHighlight, captionExtract, stateExtract, map[string]any{"i": i, "root": 0, "max": arr[0]})

		arr[0], arr[i] = arr[i], arr[0]
		ids[0], ids[i] = ids[i], ids[0]
		b.IncSwaps()

		captionSwap := fmt.Sprintf("Swapped root [0] with last heap element [%d] to place maximum (%d) at settled position.", i, arr[i])
		stateSwap := snapshot(captionSwap, i)
		stateSwap.Swapped = []int{0, i}
		b.AddStep("L7", trace.ActionSwap, captionSwap, stateSwap, map[string]any{"i": i, "swapped": []int{0, i}})

		sorted = append([]int{i}, sorted...)
		captionSettle := fmt.Sprintf("Element at [%d] (%d) settled; joining growing sorted suffix [%d..%d].", i, arr[i], i, n-1)
		b.AddStep("L7", trace.ActionSet, captionSettle, snapshot(captionSettle, -1), map[string]any{"i": i, "sortedIndex": i})

		captionReheap := fmt.Sprintf("Re-heapify: restore max-heap at root [0] on reduced heap size %d.", i)
		b.AddStep("L8", trace.ActionVisit, captionReheap, snapshot(captionReheap, 0), map[string]any{"n": i, "root": 0})

		heapifyTrace(b, arr, ids, i, 0, sorted)
	}

	sorted = append([]int{0}, sorted...)
	b.AddStep("L10", trace.ActionDone, "Heap Sort completed! All elements sorted.", snapshot("Heap Sort completed! All elements sorted.", -1), map[string]any{"status": "complete"})

	return b.Steps()
}

func heapifyTrace(b *trace.Builder, arr []int, ids []string, n, i int, sorted []int) {
	snapshot := func(caption string, current int) renderers.ArrayState {
		return renderers.ArrayState{
			Values:  slices.Clone(arr),
			IDs:     slices.Clone(ids),
			Sorted:  slices.Clone(sorted),
			Current: current,
			Caption: caption,
		}
	}

	largest := i
	l := 2*i + 1
	r := 2*i + 2

	comp := []int{i}
	if l < n {
		comp = append(comp, l)
	}
	if r < n {
		comp = append(comp, r)
	}

	var captionSubtree string
	switch {
	case l < n && r < n:
		captionSubtree = fmt.Sprintf("Max-heapify: check parent [%d] (%d) with left child [%d] (%d) and right child [%d] (%d).", i, arr[i], l, arr[l], r, arr[r])
	case l < n:
		captionSubtree = fmt.Sprintf("Max-heapify: check parent [%d] (%d) with left child [%d] (%d).", i, arr[i], l, arr[l])
	default:
		captionSubtree = fmt.Sprintf("Max-heapify: node [%d] (%d) has no children in heap size %d.", i, arr[i], n)
	}

	state12 := snapshot(captionSubtree, i)
	state12.Comparing = slices.Clone(comp)
	b.AddStep("L12", trace.ActionHighlight, captionSubtree, state12, map[string]any{"i": i, "l": l, "r": r, "largest": largest, "heapSize": n})

	if l < n {
		b.IncComparisons()
		prevLargest := largest
		isLarger := arr[l] > arr[largest]
		if isLarger {
			largest = l
		}
		var captionL string
		if isLarger {
			captionL = fmt.Sprintf("Left child [%d] (%d) > parent/largest [%d] (%d); new largest is [%d].", l, arr[l], prevLargest, arr[prevLargest], l)
		} else {
			captionL = fmt.Sprintf("Left child [%d] (%d) <= parent/largest [%d] (%d); largest remains [%d].", l, arr[l], prevLargest, arr[prevLargest], prevLargest)
		}
		state13 := snapshot(captionL, i)
		state13.Comparing = []int{prevLargest, l}
		b.AddStep("L13", trace.ActionCompare, captionL, state13, map[string]any{"l": l, "largest": largest, "i": i})
	}

	if r < n {
		b.IncComparisons()
		prevLargest := largest
		isLarger := arr[r] > arr[largest]
		if isLarger {
			largest = r
		}
		var captionR string
		if isLarger {
			captionR = fmt.Sprintf("Right child [%d] (%d) > largest [%d] (%d); new largest is [%d].", r, arr[r], prevLargest, arr[prevLargest], r)
		} else {
			captionR = fmt.Sprintf("Right child [%d] (%d) <= largest [%d] (%d); largest remains [%d].", r, arr[r], prevLargest, arr[prevLargest], prevLargest)
		}
		state14 := snapshot(captionR, i)
		state14.Comparing = []int{prevLargest, r}
		b.AddStep("L14", trace.ActionCompare, captionR, state14, map[string]any{"r": r, "largest": largest, "i": i})
	}

	if largest != i {
		captionViolated := fmt.Sprintf("Max-heap property violated: child [%d] (%d) > parent [%d] (%d); sift-down swap required.", largest, arr[largest], i, arr[i])
		state15 := snapshot(captionViolated, largest)
		state15.Comparing = []int{i, largest}
		b.AddStep("L15", trace.ActionHighlight, captionViolated, state15, map[string]any{"i": i, "largest": largest})

		arr[i], arr[largest] = arr[largest], arr[i]
		ids[i], ids[largest] = ids[largest], ids[i]
		b.IncSwaps()

		captionSwap := fmt.Sprintf("Sift-down swap: swapped parent [%d] and child [%d] to restore max-heap property (parent %d >= child %d).", i, largest, arr[i], arr[largest])
		state16 := snapshot(captionSwap, largest)
		state16.Swapped = []int{i, largest}
		b.AddStep("L16", trace.ActionSwap, captionSwap, state16, map[string]any{"i": i, "largest": largest})

		captionRec := fmt.Sprintf("Recursively heapify affected subtree at child index [%d].", largest)
		b.AddStep("L17", trace.ActionVisit, captionRec, snapshot(captionRec, largest), map[string]any{"largest": largest})

		heapifyTrace(b, arr, ids, n, largest, sorted)
	} else if len(comp) > 1 {
		captionSatisfied := fmt.Sprintf("Max-heap property satisfied: parent [%d] (%d) >= children; no swap needed.", i, arr[i])
		b.AddStep("L15", trace.ActionHighlight, captionSatisfied, snapshot(captionSatisfied, i), map[string]any{"i": i, "largest": largest})
	}
}
