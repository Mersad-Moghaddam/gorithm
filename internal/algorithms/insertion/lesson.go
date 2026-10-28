package insertion

import (
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "insertion",
		Name:     "Insertion Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Builds the final sorted array one item at a time by repeatedly taking the next element and inserting it into its correct position among previously sorted elements.",
			TimeComplexity:  "O(n²)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Incremental Build", "Adaptive", "In-Place", "Stable"},
			HowItWorks:      "Divides the array into a sorted sub-array on the left and an unsorted portion on the right. For each unsorted element (the key), it scans backwards through the sorted sub-array, shifts elements greater than the key one position to the right, and inserts the key into the vacant spot.",
		},
		DefaultInput: "[12, 11, 13, 5, 6]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func InsertionSort(arr []int) {"},
		{ID: "L2", Number: 2, Code: "    n := len(arr)"},
		{ID: "L3", Number: 3, Code: "    for i := 1; i < n; i++ {"},
		{ID: "L4", Number: 4, Code: "        key := arr[i]"},
		{ID: "L5", Number: 5, Code: "        j := i - 1"},
		{ID: "L6", Number: 6, Code: "        for j >= 0 && arr[j] > key {"},
		{ID: "L7", Number: 7, Code: "            arr[j+1] = arr[j]"},
		{ID: "L8", Number: 8, Code: "            j--"},
		{ID: "L9", Number: 9, Code: "        }"},
		{ID: "L10", Number: 10, Code: "        arr[j+1] = key"},
		{ID: "L11", Number: 11, Code: "    }"},
		{ID: "L12", Number: 12, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{12, 11, 13, 5, 6}
	if len(input) > 0 {
		var custom []*int
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("insertion input must be an array of integers: %w", err)
		}
		if len(custom) < 1 || len(custom) > 64 {
			return nil, fmt.Errorf("insertion input needs 1–64 integers")
		}
		arr = make([]int, len(custom))
		for i, value := range custom {
			if value == nil {
				return nil, fmt.Errorf("insertion input must be an array of integers")
			}
			if *value < -1000000 || *value > 1000000 {
				return nil, fmt.Errorf("insertion values must be within [-1000000,1000000]")
			}
			arr[i] = *value
		}
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
