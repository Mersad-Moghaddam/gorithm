package bucketsort

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
		ID:       "bucketsort",
		Name:     "Bucket Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Distributes elements into buckets based on value ranges, sorts each bucket individually with insertion sort, and concatenates the buckets into a sorted array.",
			TimeComplexity:  "O(n + k) average, O(n²) worst",
			SpaceComplexity: "O(n + k)",
			Difficulty:      "Medium",
			KeyConcepts:     []string{"Scatter-Gather", "Uniform Distribution", "Local Insertion Sorting", "Range Hashing"},
			HowItWorks:      "Finds minimum and maximum values to establish the numeric range. Distributes elements across k buckets using range hashing: (val - min) * (k - 1) / (max - min). Sorts each bucket using insertion sort, then concatenates all buckets in order.",
		},
		DefaultInput: "[78, 17, 39, 26, 72, 94, 21, 12, 23, 68]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func BucketSort(arr []int) []int {"},
		{ID: "L2", Number: 2, Code: "    n := len(arr)"},
		{ID: "L3", Number: 3, Code: "    if n <= 1 { return arr }"},
		{ID: "L4", Number: 4, Code: "    minVal, maxVal := minMax(arr)"},
		{ID: "L5", Number: 5, Code: "    k := 5"},
		{ID: "L6", Number: 6, Code: "    buckets := make([][]int, k)"},
		{ID: "L7", Number: 7, Code: "    for _, x := range arr {"},
		{ID: "L8", Number: 8, Code: "        idx := (x - minVal) * (k - 1) / (maxVal - minVal)"},
		{ID: "L9", Number: 9, Code: "        buckets[idx] = append(buckets[idx], x)"},
		{ID: "L10", Number: 10, Code: "    }"},
		{ID: "L11", Number: 11, Code: "    for i := 0; i < k; i++ {"},
		{ID: "L12", Number: 12, Code: "        insertionSort(buckets[i])"},
		{ID: "L13", Number: 13, Code: "    }"},
		{ID: "L14", Number: 14, Code: "    idx := 0"},
		{ID: "L15", Number: 15, Code: "    for _, bucket := range buckets {"},
		{ID: "L16", Number: 16, Code: "        for _, x := range bucket {"},
		{ID: "L17", Number: 17, Code: "            arr[idx] = x; idx++"},
		{ID: "L18", Number: 18, Code: "        }"},
		{ID: "L19", Number: 19, Code: "    }"},
		{ID: "L20", Number: 20, Code: "    return arr"},
		{ID: "L21", Number: 21, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{78, 17, 39, 26, 72, 94, 21, 12, 23, 68}
	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) > 0 {
			if string(trimmed) == "null" {
				return nil, fmt.Errorf("bucketsort input cannot be null")
			}
			if trimmed[0] != '[' {
				return nil, fmt.Errorf("bucketsort input must be a JSON array of integers")
			}
			dec := json.NewDecoder(bytes.NewReader(trimmed))
			var rawList []json.RawMessage
			if err := dec.Decode(&rawList); err != nil {
				return nil, fmt.Errorf("invalid JSON array: %w", err)
			}
			if dec.More() {
				return nil, fmt.Errorf("unexpected trailing data in input")
			}
			custom := make([]int, 0, len(rawList))
			for i, raw := range rawList {
				rawTrimmed := string(bytes.TrimSpace(raw))
				if rawTrimmed == "null" {
					return nil, fmt.Errorf("element at index %d is null", i)
				}
				val, err := strconv.Atoi(rawTrimmed)
				if err != nil {
					return nil, fmt.Errorf("element %q at index %d is not an integer", rawTrimmed, i)
				}
				custom = append(custom, val)
			}
			arr = custom
		}
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
