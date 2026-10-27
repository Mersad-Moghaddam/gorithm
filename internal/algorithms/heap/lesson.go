package heap

import (
	"encoding/json"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "heap",
		Name:     "Heap Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Converts the array into a max-heap and repeatedly extracts the maximum root element to the sorted partition at the end.",
			TimeComplexity:  "O(n log n)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Binary Heap", "Max-Heapify", "In-Place", "Priority Queue Duality"},
			HowItWorks:      "First transforms the array into a max-heap where each parent is greater than its children. Then swaps the root (maximum) with the last element, shrinks heap size, and invokes max-heapify to restore heap order until all elements are sorted.",
		},
		DefaultInput: "[12, 11, 13, 5, 6, 7]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func HeapSort(arr []int) {"},
		{ID: "L2", Number: 2, Code: "    n := len(arr)"},
		{ID: "L3", Number: 3, Code: "    for i := n/2 - 1; i >= 0; i-- {"},
		{ID: "L4", Number: 4, Code: "        heapify(arr, n, i)"},
		{ID: "L5", Number: 5, Code: "    }"},
		{ID: "L6", Number: 6, Code: "    for i := n - 1; i > 0; i-- {"},
		{ID: "L7", Number: 7, Code: "        arr[0], arr[i] = arr[i], arr[0]"},
		{ID: "L8", Number: 8, Code: "        heapify(arr, i, 0)"},
		{ID: "L9", Number: 9, Code: "    }"},
		{ID: "L10", Number: 10, Code: "}"},
		{ID: "L11", Number: 11, Code: "func heapify(arr []int, n, i int) {"},
		{ID: "L12", Number: 12, Code: "    largest := i; l := 2*i + 1; r := 2*i + 2"},
		{ID: "L13", Number: 13, Code: "    if l < n && arr[l] > arr[largest] { largest = l }"},
		{ID: "L14", Number: 14, Code: "    if r < n && arr[r] > arr[largest] { largest = r }"},
		{ID: "L15", Number: 15, Code: "    if largest != i {"},
		{ID: "L16", Number: 16, Code: "        arr[i], arr[largest] = arr[largest], arr[i]"},
		{ID: "L17", Number: 17, Code: "        heapify(arr, n, largest)"},
		{ID: "L18", Number: 18, Code: "    }"},
		{ID: "L19", Number: 19, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{12, 11, 13, 5, 6, 7}
	if len(input) > 0 {
		var custom []int
		if err := json.Unmarshal(input, &custom); err == nil && len(custom) > 0 {
			arr = custom
		}
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
