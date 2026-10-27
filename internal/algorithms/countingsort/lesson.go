package countingsort

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
		ID:       "countingsort",
		Name:     "Counting Sort",
		Category: catalog.CategorySorting,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Non-comparison integer sorting algorithm that counts occurrences of each distinct element, calculates prefix sums to establish indices, and places elements into a sorted array.",
			TimeComplexity:  "O(n + k)",
			SpaceComplexity: "O(k)",
			Difficulty:      "Easy",
			KeyConcepts: []string{
				"Non-Comparison Sort",
				"Frequency Tally",
				"Prefix Sum Accumulation",
				"Stable Output Mapping",
			},
			HowItWorks: "1. Scans the input array to determine the maximum element value k.\n2. Allocates a count array of size k + 1 and tallies each element's occurrences.\n3. Accumulates prefix sums across the count array so each position stores the count of elements <= that value.\n4. Iterates through the input array backwards, placing each element into its corresponding output position and decrementing the count to maintain stability.",
		},
		DefaultInput: "[4, 2, 2, 8, 3, 3, 1]",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func CountingSort(arr []int) []int {"},
		{ID: "L2", Number: 2, Code: "    n := len(arr)"},
		{ID: "L3", Number: 3, Code: "    if n <= 1 {"},
		{ID: "L4", Number: 4, Code: "        return arr"},
		{ID: "L5", Number: 5, Code: "    }"},
		{ID: "L6", Number: 6, Code: "    max := arr[0]"},
		{ID: "L7", Number: 7, Code: "    for i := 1; i < n; i++ {"},
		{ID: "L8", Number: 8, Code: "        if arr[i] > max {"},
		{ID: "L9", Number: 9, Code: "            max = arr[i]"},
		{ID: "L10", Number: 10, Code: "        }"},
		{ID: "L11", Number: 11, Code: "    }"},
		{ID: "L12", Number: 12, Code: "    count := make([]int, max+1)"},
		{ID: "L13", Number: 13, Code: "    for i := 0; i < n; i++ {"},
		{ID: "L14", Number: 14, Code: "        count[arr[i]]++"},
		{ID: "L15", Number: 15, Code: "    }"},
		{ID: "L16", Number: 16, Code: "    for i := 1; i <= max; i++ {"},
		{ID: "L17", Number: 17, Code: "        count[i] += count[i-1]"},
		{ID: "L18", Number: 18, Code: "    }"},
		{ID: "L19", Number: 19, Code: "    output := make([]int, n)"},
		{ID: "L20", Number: 20, Code: "    for i := n - 1; i >= 0; i-- {"},
		{ID: "L21", Number: 21, Code: "        output[count[arr[i]]-1] = arr[i]"},
		{ID: "L22", Number: 22, Code: "        count[arr[i]]--"},
		{ID: "L23", Number: 23, Code: "    }"},
		{ID: "L24", Number: 24, Code: "    return output"},
		{ID: "L25", Number: 25, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace([]int{4, 2, 2, 8, 3, 3, 1}), nil
	}
	if string(trimmed) == "null" {
		return nil, fmt.Errorf("counting sort input cannot be null")
	}
	if trimmed[0] != '[' {
		return nil, fmt.Errorf("counting sort input must be a JSON array")
	}
	d := json.NewDecoder(bytes.NewReader(trimmed))
	d.UseNumber()
	var numbers []json.Number
	if err := d.Decode(&numbers); err != nil {
		return nil, fmt.Errorf("invalid json array of numbers: %w", err)
	}
	if d.More() {
		return nil, fmt.Errorf("unexpected trailing data in input")
	}
	arr := make([]int, len(numbers))
	for i, num := range numbers {
		val, err := num.Int64()
		if err != nil {
			return nil, fmt.Errorf("element at index %d is not an integer: %s", i, num.String())
		}
		if val < 0 {
			return nil, fmt.Errorf("element at index %d is negative (%d); counting sort requires non-negative integers", i, val)
		}
		arr[i] = int(val)
	}
	return Trace(arr), nil
}

func init() {
	catalog.Register(Lesson{})
}
