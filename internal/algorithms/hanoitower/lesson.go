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

type buildInput struct {
	Disks json.RawMessage `json:"disks"`
	N     json.RawMessage `json:"n"`
	Count json.RawMessage `json:"count"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	n := 3
	input = bytes.TrimSpace(input)
	if len(input) > 0 {
		if input[0] == '{' {
			var obj buildInput
			if err := json.Unmarshal(input, &obj); err != nil {
				return nil, fmt.Errorf("invalid Hanoi input: %w", err)
			}
			selected := false
			for _, raw := range []json.RawMessage{obj.Disks, obj.N, obj.Count} {
				if len(raw) == 0 {
					continue
				}
				var value int
				if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
					return nil, fmt.Errorf("disk count must be an integer in [1,8]")
				}
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("disk count must be an integer: %w", err)
				}
				if value < 1 || value > 8 {
					return nil, fmt.Errorf("disk count must be in [1,8]")
				}
				if !selected {
					n = value
					selected = true
				}
			}
		} else {
			if bytes.Equal(input, []byte("null")) {
				return nil, fmt.Errorf("disk count must be an integer in [1,8]")
			}
			if err := json.Unmarshal(input, &n); err != nil {
				return nil, fmt.Errorf("invalid disk count: %w", err)
			}
		}
	}
	if n < 1 || n > 8 {
		return nil, fmt.Errorf("disk count must be in [1,8]")
	}
	return Trace(n), nil
}

func init() {
	catalog.Register(Lesson{})
}
