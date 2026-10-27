package linearsearch

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
		ID:       "linearsearch",
		Name:     "Linear Search",
		Category: catalog.CategorySearching,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "Sequentially checks each element of the list until a match is found or the whole list has been searched.",
			TimeComplexity:  "O(n)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Sequential Scan", "Unordered Data", "Early Exit"},
			HowItWorks:      "Starts from the first element of the array and compares each element with the target value sequentially. If a match is found, returns its index immediately. If the end of the array is reached without finding the target, returns -1.",
		},
		DefaultInput: `{"array": [15, 28, 42, 8, 33, 91, 55], "target": 33}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func LinearSearch(arr []int, target int) int {"},
		{ID: "L2", Number: 2, Code: "    for i := 0; i < len(arr); i++ {"},
		{ID: "L3", Number: 3, Code: "        if arr[i] == target {"},
		{ID: "L4", Number: 4, Code: "            return i"},
		{ID: "L5", Number: 5, Code: "        }"},
		{ID: "L6", Number: 6, Code: "    }"},
		{ID: "L7", Number: 7, Code: "    return -1"},
		{ID: "L8", Number: 8, Code: "}"},
	}
}

type buildInput struct {
	Array  *[]int `json:"array"`
	Arr    *[]int `json:"arr"`
	Target *int   `json:"target"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{15, 28, 42, 8, 33, 91, 55}
	target := 33

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace(arr, target), nil
	}

	var payload buildInput
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return nil, fmt.Errorf("linearsearch: invalid input: %w", err)
	}

	var elements []int
	if payload.Array != nil {
		elements = *payload.Array
	} else if payload.Arr != nil {
		elements = *payload.Arr
	} else {
		return nil, fmt.Errorf("linearsearch: missing array")
	}

	if len(elements) == 0 {
		return nil, fmt.Errorf("linearsearch: array cannot be empty")
	}

	if payload.Target == nil {
		return nil, fmt.Errorf("linearsearch: missing or null target")
	}

	return Trace(elements, *payload.Target), nil
}

func init() {
	catalog.Register(Lesson{})
}
