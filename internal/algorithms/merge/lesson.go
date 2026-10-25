package merge

import (
	"encoding/json"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "merge",
		Name:     "Merge Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "A divide-and-conquer algorithm that recursively splits an array into halves, sorts each half, and merges them.",
			TimeComplexity:  "O(n log n)",
			SpaceComplexity: "O(n)",
			KeyConcepts:     []string{"Divide and Conquer", "Stable", "Predictable Performance", "Auxiliary Array"},
			HowItWorks:      "Divides the input array into two halves, recursively sorts them, and then merges the two sorted halves into a single sorted array using an auxiliary buffer.",
		},
		DefaultInput: "[38, 27, 43, 3, 9, 82, 10]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func MergeSort(arr []int, left, right int) {"},
		{ID: "L2", Number: 2, Code: "    if left >= right {"},
		{ID: "L3", Number: 3, Code: "        return"},
		{ID: "L4", Number: 4, Code: "    }"},
		{ID: "L5", Number: 5, Code: "    mid := left + (right-left)/2"},
		{ID: "L6", Number: 6, Code: "    MergeSort(arr, left, mid)"},
		{ID: "L7", Number: 7, Code: "    MergeSort(arr, mid+1, right)"},
		{ID: "L8", Number: 8, Code: "    merge(arr, left, mid, right)"},
		{ID: "L9", Number: 9, Code: "}"},
		{ID: "L10", Number: 10, Code: "func merge(arr []int, left, mid, right int) {"},
		{ID: "L11", Number: 11, Code: "    aux := make([]int, right-left+1)"},
		{ID: "L12", Number: 12, Code: "    i, j, k := left, mid+1, 0"},
		{ID: "L13", Number: 13, Code: "    for i <= mid && j <= right {"},
		{ID: "L14", Number: 14, Code: "        if arr[i] <= arr[j] {"},
		{ID: "L15", Number: 15, Code: "            aux[k] = arr[i]; i++"},
		{ID: "L16", Number: 16, Code: "        } else {"},
		{ID: "L17", Number: 17, Code: "            aux[k] = arr[j]; j++"},
		{ID: "L18", Number: 18, Code: "        }"},
		{ID: "L19", Number: 19, Code: "        k++"},
		{ID: "L20", Number: 20, Code: "    }"},
		{ID: "L21", Number: 21, Code: "    for i <= mid { aux[k] = arr[i]; i++; k++ }"},
		{ID: "L22", Number: 22, Code: "    for j <= right { aux[k] = arr[j]; j++; k++ }"},
		{ID: "L23", Number: 23, Code: "    for p := 0; p < len(aux); p++ { arr[left+p] = aux[p] }"},
		{ID: "L24", Number: 24, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{38, 27, 43, 3, 9, 82, 10}
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
