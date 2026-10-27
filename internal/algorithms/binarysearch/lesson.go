package binarysearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "binarysearch",
		Name:     "Binary Search",
		Category: catalog.CategorySearching,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "An efficient search algorithm that finds the position of a target value within a sorted array by repeatedly halving the search space.",
			TimeComplexity:  "O(log n)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Divide and Conquer", "Sorted Precondition", "Range Bisection", "Logarithmic Search"},
			HowItWorks:      "Maintains low and high pointers defining the active search window. Evaluates the middle element against the target: if equal, search succeeds; if the target is greater, search continues in the right subarray; if smaller, in the left subarray. Each comparison discards half of the remaining elements.",
		},
		DefaultInput: `{"array": [3, 8, 14, 22, 35, 47, 59, 71, 88], "target": 47}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func BinarySearch(arr []int, target int) int {"},
		{ID: "L2", Number: 2, Code: "    low, high := 0, len(arr)-1"},
		{ID: "L3", Number: 3, Code: "    for low <= high {"},
		{ID: "L4", Number: 4, Code: "        mid := low + (high-low)/2"},
		{ID: "L5", Number: 5, Code: "        if arr[mid] == target {"},
		{ID: "L6", Number: 6, Code: "            return mid"},
		{ID: "L7", Number: 7, Code: "        } else if arr[mid] < target {"},
		{ID: "L8", Number: 8, Code: "            low = mid + 1"},
		{ID: "L9", Number: 9, Code: "        } else {"},
		{ID: "L10", Number: 10, Code: "            high = mid - 1"},
		{ID: "L11", Number: 11, Code: "        }"},
		{ID: "L12", Number: 12, Code: "    }"},
		{ID: "L13", Number: 13, Code: "    return -1"},
		{ID: "L14", Number: 14, Code: "}"},
	}
}

type buildInput struct {
	Array  *[]int `json:"array"`
	Arr    *[]int `json:"arr"`
	Target *int   `json:"target"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{3, 8, 14, 22, 35, 47, 59, 71, 88}
	target := 47

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) > 0 && string(trimmed) != "null" {
		var custom buildInput
		if err := json.Unmarshal(trimmed, &custom); err != nil {
			return nil, fmt.Errorf("invalid input format: %w", err)
		}
		if custom.Target == nil {
			return nil, fmt.Errorf("target is required and cannot be null")
		}
		if custom.Array != nil {
			arr = *custom.Array
		} else if custom.Arr != nil {
			arr = *custom.Arr
		} else {
			return nil, fmt.Errorf("array is required and cannot be null")
		}
		target = *custom.Target

		if !slices.IsSorted(arr) {
			for i := 1; i < len(arr); i++ {
				if arr[i] < arr[i-1] {
					return nil, fmt.Errorf("array must be sorted in ascending order (arr[%d]=%d > arr[%d]=%d)", i-1, arr[i-1], i, arr[i])
				}
			}
			return nil, fmt.Errorf("array must be sorted in ascending order")
		}
	}

	return Trace(arr, target), nil
}

func init() {
	catalog.Register(Lesson{})
}
