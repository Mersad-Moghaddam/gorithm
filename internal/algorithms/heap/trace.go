package heap

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace records real heap operations; identities travel with values, not slots.
func Trace(initial []int) []trace.Step {
	b := trace.NewBuilder()
	arr := cloneSlice(initial)
	ids := make([]int, len(arr))
	for i := range ids {
		ids[i] = i
	}
	n := len(arr)
	s := renderers.HeapLessonState{ArrayState: renderers.ArrayState{Current: -1}, HeapSize: n, Root: -1, LeftChild: -1, RightChild: -1, Largest: -1, ExtractIndex: -1, Building: true}
	add := func(line string, action trace.ActionType, text string, vars map[string]any) {
		snap := s
		snap.Values = cloneSlice(arr)
		snap.IDs = cloneSlice(ids)
		snap.Sorted = cloneSlice(s.Sorted)
		snap.Comparing = cloneSlice(s.Comparing)
		snap.Swapped = cloneSlice(s.Swapped)
		b.AddStep(line, action, text, snap, vars)
		s.Comparing = nil
		s.Swapped = nil
	}
	add("L1", trace.ActionInit, "Initialize Heap Sort with input array.", map[string]any{"n": n})
	if n == 0 {
		s.Building = false
		add("L10", trace.ActionDone, "Heap Sort completed! Empty array.", map[string]any{"status": "complete"})
		return b.Steps()
	}
	add("L2", trace.ActionSet, fmt.Sprintf("Array length is %d.", n), map[string]any{"n": n})
	if n == 1 {
		s.Sorted = []int{0}
		s.Building = false
		s.HeapSize = 0
		add("L10", trace.ActionDone, "Heap Sort completed! Single element sorted.", map[string]any{"status": "complete"})
		return b.Steps()
	}
	var heapifyTrace func(int, int)
	heapifyTrace = func(size, i int) {
		largest, l, r := i, 2*i+1, 2*i+2
		s.Root = i
		s.LeftChild = -1
		s.RightChild = -1
		s.Largest = i
		s.Current = i
		if l < size {
			s.LeftChild = l
		}
		if r < size {
			s.RightChild = r
		}
		add("L12", trace.ActionSet, fmt.Sprintf("Check subtree root %d against children: left %d, right %d.", i, l, r), map[string]any{"i": i, "l": l, "r": r, "largest": largest})
		if l < size {
			b.IncComparisons()
			s.Largest = largest
			s.Comparing = []int{l, largest}
			add("L13", trace.ActionCompare, fmt.Sprintf("Comparing left child arr[%d] (%d) with arr[%d] (%d).", l, arr[l], largest, arr[largest]), map[string]any{"l": l, "largest": largest})
			if arr[l] > arr[largest] {
				largest = l
			}
		}
		if r < size {
			b.IncComparisons()
			s.Largest = largest
			s.Comparing = []int{r, largest}
			add("L14", trace.ActionCompare, fmt.Sprintf("Comparing right child arr[%d] (%d) with arr[%d] (%d).", r, arr[r], largest, arr[largest]), map[string]any{"r": r, "largest": largest})
			if arr[r] > arr[largest] {
				largest = r
			}
		}
		s.Largest = largest
		if largest != i {
			s.Comparing = []int{i, largest}
			add("L15", trace.ActionHighlight, fmt.Sprintf("Largest is at child index %d (%d), swapping with root %d (%d).", largest, arr[largest], i, arr[i]), map[string]any{"i": i, "largest": largest})
			arr[i], arr[largest] = arr[largest], arr[i]
			ids[i], ids[largest] = ids[largest], ids[i]
			b.IncSwaps()
			s.Swapped = []int{i, largest}
			s.Current = largest
			add("L16", trace.ActionSwap, fmt.Sprintf("Swapped arr[%d] and arr[%d] to maintain max-heap.", i, largest), map[string]any{"i": i, "largest": largest})
			add("L17", trace.ActionVisit, fmt.Sprintf("Recursively heapifying subtree at index %d.", largest), map[string]any{"largest": largest})
			heapifyTrace(size, largest)
		}
	}
	for i := n/2 - 1; i >= 0; i-- {
		s.Root = i
		s.Largest = -1
		s.LeftChild = -1
		s.RightChild = -1
		s.Current = i
		add("L3", trace.ActionHighlight, fmt.Sprintf("Build heap: inspecting subtree at index %d.", i), map[string]any{"i": i})
		add("L4", trace.ActionVisit, fmt.Sprintf("Invoking heapify for index %d on heap size %d.", i, n), map[string]any{"i": i, "n": n})
		heapifyTrace(n, i)
	}
	s.Building = false
	for i := n - 1; i > 0; i-- {
		s.Root = 0
		s.ExtractIndex = i
		s.Current = i
		s.Largest = -1
		s.LeftChild = -1
		s.RightChild = -1
		add("L6", trace.ActionHighlight, fmt.Sprintf("Extract max element: root arr[0] (%d) to position %d.", arr[0], i), map[string]any{"i": i})
		arr[0], arr[i] = arr[i], arr[0]
		ids[0], ids[i] = ids[i], ids[0]
		b.IncSwaps()
		s.Sorted = append(s.Sorted, i)
		s.Swapped = []int{0, i}
		s.HeapSize = i
		add("L7", trace.ActionSwap, fmt.Sprintf("Swapped root %d with arr[%d] (%d).", arr[i], i, arr[0]), map[string]any{"i": i})
		s.Current = 0
		add("L8", trace.ActionVisit, fmt.Sprintf("Re-heapify root on reduced heap size %d.", i), map[string]any{"n": i, "root": 0})
		heapifyTrace(i, 0)
	}
	s.Sorted = append(s.Sorted, 0)
	s.Current = -1
	s.Root = -1
	s.LeftChild = -1
	s.RightChild = -1
	s.Largest = -1
	s.ExtractIndex = -1
	s.HeapSize = 0
	add("L10", trace.ActionDone, "Heap Sort completed! All elements sorted.", map[string]any{"status": "complete"})
	steps := b.Steps()
	for i := 1; i < len(steps); i++ {
		if steps[i-1].LineID == "L15" && steps[i].LineID == "L16" || steps[i-1].LineID == "L6" && steps[i].LineID == "L7" {
			steps[i-1].HoldMS = 600
			steps[i].HoldMS = 1200
		}
	}
	return steps
}

func cloneSlice(s []int) []int {
	if s == nil {
		return nil
	}
	c := make([]int, len(s))
	copy(c, s)
	return c
}
