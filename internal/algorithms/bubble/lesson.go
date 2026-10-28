package bubble

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
		ID:       "bubble",
		Name:     "Bubble Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Repeatedly steps through the list, compares adjacent elements, and swaps them if they are in the wrong order.",
			TimeComplexity:  "O(n²)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Adjacent Comparison", "In-Place", "Stable Sorting", "Quadratic Time"},
			HowItWorks:      "Passes through the array n-1 times. On each pass, the largest unsorted element 'bubbles up' to its final position at the end of the array.",
		},
		DefaultInput: "[29, 10, 14, 37, 13]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func BubbleSort(arr []int) {"},
		{ID: "L2", Number: 2, Code: "    n := len(arr)"},
		{ID: "L3", Number: 3, Code: "    for i := 0; i < n-1; i++ {"},
		{ID: "L4", Number: 4, Code: "        for j := 0; j < n-i-1; j++ {"},
		{ID: "L5", Number: 5, Code: "            if arr[j] > arr[j+1] {"},
		{ID: "L6", Number: 6, Code: "                arr[j], arr[j+1] = arr[j+1], arr[j]"},
		{ID: "L7", Number: 7, Code: "            }"},
		{ID: "L8", Number: 8, Code: "        }"},
		{ID: "L9", Number: 9, Code: "    }"},
		{ID: "L10", Number: 10, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{29, 10, 14, 37, 13}
	if len(input) > 0 {
		var custom []int
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("Bubble Sort input must be an array of integers: %w", err)
		}
		if bytes.Contains(input, []byte("null")) {
			return nil, fmt.Errorf("Bubble Sort input must contain integers, not null")
		}
		if len(custom) == 0 || len(custom) > 64 {
			return nil, fmt.Errorf("Bubble Sort input must contain 1–64 integers")
		}
		for _, value := range custom {
			if value < -1000000 || value > 1000000 {
				return nil, fmt.Errorf("Bubble Sort values must be between -1000000 and 1000000")
			}
		}
		arr = custom
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
