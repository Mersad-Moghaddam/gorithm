package quick

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "quick",
		Name:     "Quick Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "An efficient, divide-and-conquer sorting algorithm that partitions an array around a pivot element and recursively sorts the sub-arrays.",
			TimeComplexity:  "O(n log n) avg, O(n²) worst",
			SpaceComplexity: "O(log n)",
			KeyConcepts:     []string{"Divide and Conquer", "Partitioning", "Pivot Selection", "In-Place"},
			HowItWorks:      "Selects a pivot element (commonly the last element in Lomuto partition) and partitions the array so elements smaller than the pivot precede it while greater elements follow it. Recursively sorts subarrays before and after the pivot.",
		},
		DefaultInput: "[10, 80, 30, 90, 40, 50, 70]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func QuickSort(arr []int, low, high int) {"},
		{ID: "L2", Number: 2, Code: "    if low < high {"},
		{ID: "L3", Number: 3, Code: "        p := partition(arr, low, high)"},
		{ID: "L4", Number: 4, Code: "        QuickSort(arr, low, p-1)"},
		{ID: "L5", Number: 5, Code: "        QuickSort(arr, p+1, high)"},
		{ID: "L6", Number: 6, Code: "    }"},
		{ID: "L7", Number: 7, Code: "}"},
		{ID: "L8", Number: 8, Code: "func partition(arr []int, low, high int) int {"},
		{ID: "L9", Number: 9, Code: "    pivot := arr[high]"},
		{ID: "L10", Number: 10, Code: "    i := low - 1"},
		{ID: "L11", Number: 11, Code: "    for j := low; j < high; j++ {"},
		{ID: "L12", Number: 12, Code: "        if arr[j] < pivot {"},
		{ID: "L13", Number: 13, Code: "            i++"},
		{ID: "L14", Number: 14, Code: "            arr[i], arr[j] = arr[j], arr[i]"},
		{ID: "L15", Number: 15, Code: "        }"},
		{ID: "L16", Number: 16, Code: "    }"},
		{ID: "L17", Number: 17, Code: "    arr[i+1], arr[high] = arr[high], arr[i+1]"},
		{ID: "L18", Number: 18, Code: "    return i + 1"},
		{ID: "L19", Number: 19, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{10, 80, 30, 90, 40, 50, 70}
	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) == 0 {
			return Trace(arr), nil
		}
		if trimmed[0] != '[' || trimmed[len(trimmed)-1] != ']' {
			return nil, fmt.Errorf("quick sort input must be a JSON array of integers")
		}
		var raw []json.RawMessage
		if err := json.Unmarshal(trimmed, &raw); err != nil {
			return nil, err
		}
		arr = make([]int, len(raw))
		for i, item := range raw {
			itemStr := string(bytes.TrimSpace(item))
			if itemStr == "null" {
				return nil, fmt.Errorf("quick sort element cannot be null")
			}
			val, err := strconv.Atoi(itemStr)
			if err != nil {
				return nil, fmt.Errorf("quick sort element %q is not a valid integer: %w", itemStr, err)
			}
			arr[i] = val
		}
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
