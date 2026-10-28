package radixsort

import (
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "radixsort",
		Name:     "Radix Sort (LSD)",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Non-comparative integer sorting algorithm that sorts elements digit by digit, from least significant digit (LSD) to most significant digit, using a stable sub-sort.",
			TimeComplexity:  "O(d · (n + b))",
			SpaceComplexity: "O(n + b)",
			Difficulty:      "Medium",
			KeyConcepts: []string{
				"Digit-by-Digit Processing",
				"Stable Sub-sorts",
				"Non-Comparison Sorting",
				"Fixed Radix Base 10",
			},
			HowItWorks: "1. Finds the maximum element to determine the maximum number of digits d.\n2. Iterates from the least significant digit (1s place) up to the most significant digit (multiplying the exponent exp by 10 each pass).\n3. In each pass, performs a stable counting sort based on the digit at (value / exp) % 10.\n4. Preserves relative ordering from previous passes to ensure cumulative correctness across all digits.",
		},
		DefaultInput: "[170, 45, 75, 90, 802, 24, 2, 66]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func RadixSort(arr []int) {"},
		{ID: "L2", Number: 2, Code: "    maxVal := getMax(arr)"},
		{ID: "L3", Number: 3, Code: "    for exp := 1; maxVal/exp > 0; exp *= 10 {"},
		{ID: "L4", Number: 4, Code: "        countingSortByDigit(arr, exp)"},
		{ID: "L5", Number: 5, Code: "    }"},
		{ID: "L6", Number: 6, Code: "}"},
		{ID: "L7", Number: 7, Code: "func countingSortByDigit(arr []int, exp int) {"},
		{ID: "L8", Number: 8, Code: "    n := len(arr); output := make([]int, n)"},
		{ID: "L9", Number: 9, Code: "    var count [10]int"},
		{ID: "L10", Number: 10, Code: "    for i := 0; i < n; i++ { count[(arr[i]/exp)%10]++ }"},
		{ID: "L11", Number: 11, Code: "    for i := 1; i < 10; i++ { count[i] += count[i-1] }"},
		{ID: "L12", Number: 12, Code: "    for i := n - 1; i >= 0; i-- {"},
		{ID: "L13", Number: 13, Code: "        digit := (arr[i]/exp) % 10"},
		{ID: "L14", Number: 14, Code: "        output[count[digit]-1] = arr[i]"},
		{ID: "L15", Number: 15, Code: "        count[digit]--"},
		{ID: "L16", Number: 16, Code: "    }"},
		{ID: "L17", Number: 17, Code: "    for i := 0; i < n; i++ { arr[i] = output[i] }"},
		{ID: "L18", Number: 18, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{170, 45, 75, 90, 802, 24, 2, 66}
	if input != nil {
		var custom []int
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("radix sort input must be an array of integers: %w", err)
		}
		if len(custom) == 0 || len(custom) > 64 {
			return nil, fmt.Errorf("radix sort needs 1 to 64 integers")
		}
		for _, value := range custom {
			if value < 0 || value > 1000000 {
				return nil, fmt.Errorf("radix sort values must be between 0 and 1000000")
			}
		}
		arr = custom
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
