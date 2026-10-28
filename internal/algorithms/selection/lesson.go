package selection

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "selection",
		Name:     "Selection Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Repeatedly finds minimum element from unsorted portion and places it at beginning of sorted partition.",
			TimeComplexity:  "O(n²)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Min Element Scanning", "In-Place", "Unstable Swapping"},
			HowItWorks:      "Divides input list into two parts: sorted sublist built up from left to right and unsorted sublist occupying rest. In each pass, scans unsorted sublist to locate smallest element, swaps it with first unsorted element, and extends sorted boundary by one.",
		},
		DefaultInput: "[64, 25, 12, 22, 11]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func SelectionSort(arr []int) {"},
		{ID: "L2", Number: 2, Code: "    n := len(arr)"},
		{ID: "L3", Number: 3, Code: "    for i := 0; i < n-1; i++ {"},
		{ID: "L4", Number: 4, Code: "        minIdx := i"},
		{ID: "L5", Number: 5, Code: "        for j := i + 1; j < n; j++ {"},
		{ID: "L6", Number: 6, Code: "            if arr[j] < arr[minIdx] {"},
		{ID: "L7", Number: 7, Code: "                minIdx = j"},
		{ID: "L8", Number: 8, Code: "            }"},
		{ID: "L9", Number: 9, Code: "        }"},
		{ID: "L10", Number: 10, Code: "        arr[i], arr[minIdx] = arr[minIdx], arr[i]"},
		{ID: "L11", Number: 11, Code: "    }"},
		{ID: "L12", Number: 12, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{64, 25, 12, 22, 11}
	if input != nil {
		if bytes.Contains(input, []byte("null")) {
			return nil, fmt.Errorf("selection input must contain integers, not null")
		}
		var custom []*int
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("selection input must be an array of integers: %w", err)
		}
		if len(custom) == 0 || len(custom) > 64 {
			return nil, fmt.Errorf("selection input must contain 1–64 integers")
		}
		arr = make([]int, len(custom))
		for i, value := range custom {
			if value == nil {
				return nil, fmt.Errorf("selection input must contain integers, not null")
			}
			if *value < -1000000 || *value > 1000000 {
				return nil, fmt.Errorf("selection values must be between -1000000 and 1000000")
			}
			arr[i] = *value
		}
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
