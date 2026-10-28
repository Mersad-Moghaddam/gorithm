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
	Array  []int `json:"array"`
	Arr    []int `json:"arr"`
	Target *int  `json:"target"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{3, 8, 14, 22, 35, 47, 59, 71, 88}
	target := 47

	if len(input) > 0 {
		input = bytes.TrimSpace(input)
		if !json.Valid(input) || bytes.Equal(input, []byte("null")) {
			return nil, fmt.Errorf("expected binary search object or integer array")
		}
		var custom buildInput
		if input[0] == '{' {
			if err := json.Unmarshal(input, &custom); err != nil {
				return nil, fmt.Errorf("invalid binary search input: %w", err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(input, &fields); err != nil {
				return nil, err
			}
			for _, key := range []string{"array", "arr", "target"} {
				if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
					return nil, fmt.Errorf("%s must not be null", key)
				}
			}
			for _, values := range [][]int{custom.Array, custom.Arr} {
				if err := validateBinarySearchValues(values); err != nil {
					return nil, err
				}
			}
			if len(custom.Array) > 0 {
				arr = custom.Array
			} else if len(custom.Arr) > 0 {
				arr = custom.Arr
			}
			if custom.Target != nil {
				target = *custom.Target
			}
		} else {
			var customArr []int
			if err := json.Unmarshal(input, &customArr); err != nil {
				return nil, fmt.Errorf("expected integer array: %w", err)
			}
			if err := validateBinarySearchValues(customArr); err != nil {
				return nil, err
			}
			if len(customArr) > 0 {
				arr = customArr
			}
		}
	}
	if target < -1000000 || target > 1000000 {
		return nil, fmt.Errorf("target must be within ±1000000")
	}

	return Trace(arr, target), nil
}

func validateBinarySearchValues(values []int) error {
	if len(values) > 128 {
		return fmt.Errorf("array must contain at most 128 integers")
	}
	for _, value := range values {
		if value < -1000000 || value > 1000000 {
			return fmt.Errorf("array values must be within ±1000000")
		}
	}
	if !slices.IsSorted(values) {
		return fmt.Errorf("binary search requires nondecreasing input")
	}
	return nil
}

func init() {
	catalog.Register(Lesson{})
}
