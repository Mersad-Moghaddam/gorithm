package hanoitower

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
		ID:       "hanoitower",
		Name:     "Tower of Hanoi",
		Category: catalog.CategoryRecursion,
		Renderer: "hanoi",
		Documentation: catalog.Documentation{
			Summary:         "A mathematical puzzle where disks are moved between three pegs under the rule that no larger disk may be placed on top of a smaller disk.",
			TimeComplexity:  "O(2ⁿ)",
			SpaceComplexity: "O(n)",
			Difficulty:      "Medium",
			KeyConcepts: []string{
				"Mathematical Induction",
				"Call Stack",
				"Exponential Growth",
				"Recursive Subproblems",
			},
			HowItWorks: "To move n disks from source peg A to target peg C using auxiliary peg B: 1) Recursively move n-1 disks from A to B. 2) Move the nth (largest) disk directly from A to C. 3) Recursively move the n-1 disks from B to C.",
		},
		DefaultInput: "3",
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Hanoi(n, from, to, aux int) {"},
		{ID: "L2", Number: 2, Code: "    if n == 0 {"},
		{ID: "L3", Number: 3, Code: "        return"},
		{ID: "L4", Number: 4, Code: "    }"},
		{ID: "L5", Number: 5, Code: "    Hanoi(n-1, from, aux, to)"},
		{ID: "L6", Number: 6, Code: "    moveDisk(n, from, to)"},
		{ID: "L7", Number: 7, Code: "    Hanoi(n-1, aux, to, from)"},
		{ID: "L8", Number: 8, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	n := 3
	if len(input) > 0 {
		trimmed := bytes.TrimSpace(input)
		if len(trimmed) == 0 {
			return nil, fmt.Errorf("input cannot be whitespace")
		}
		if bytes.Equal(trimmed, []byte("null")) {
			return nil, fmt.Errorf("input cannot be null")
		}

		if trimmed[0] == '{' {
			var obj struct {
				Disks *int `json:"disks"`
				N     *int `json:"n"`
				Count *int `json:"count"`
			}
			dec := json.NewDecoder(bytes.NewReader(trimmed))
			if err := dec.Decode(&obj); err != nil {
				return nil, fmt.Errorf("malformed JSON object: %w", err)
			}
			if dec.More() {
				return nil, fmt.Errorf("unexpected extra data after JSON object")
			}
			if obj.Disks != nil {
				n = *obj.Disks
			} else if obj.N != nil {
				n = *obj.N
			} else if obj.Count != nil {
				n = *obj.Count
			} else {
				return nil, fmt.Errorf("JSON object must specify disks, n, or count")
			}
		} else {
			var directVal int
			dec := json.NewDecoder(bytes.NewReader(trimmed))
			if err := dec.Decode(&directVal); err != nil {
				return nil, fmt.Errorf("malformed JSON input: expected integer or object with disks/n/count: %w", err)
			}
			if dec.More() {
				return nil, fmt.Errorf("unexpected extra data after JSON integer")
			}
			n = directVal
		}

		if n < 1 || n > 7 {
			return nil, fmt.Errorf("disks must be between 1 and 7, got %d", n)
		}
	}
	return Trace(n), nil
}

func init() {
	catalog.Register(Lesson{})
}
