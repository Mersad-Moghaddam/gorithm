package linearsearch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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
		DefaultInput: "[15, 28, 42, 8, 33, 91, 55]",
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

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	arr := []int{15, 28, 42, 8, 33, 91, 55}
	target := 33

	if len(input) > 0 {
		input = bytes.TrimSpace(input)
		if len(input) == 0 || bytes.Equal(input, []byte("null")) {
			return nil, fmt.Errorf("input must be an array or object")
		}
		if err := checkLinearSearchNulls(input); err != nil {
			return nil, err
		}
		var payload struct {
			Arr    []int `json:"arr"`
			Array  []int `json:"array"`
			Target *int  `json:"target"`
		}
		if input[0] == '{' {
			if err := json.Unmarshal(input, &payload); err != nil {
				return nil, fmt.Errorf("invalid search input: %w", err)
			}
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(input, &fields); err != nil {
				return nil, err
			}
			for _, key := range []string{"arr", "array", "target"} {
				if value, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
					return nil, fmt.Errorf("%s must not be null", key)
				}
			}
			for _, values := range [][]int{payload.Arr, payload.Array} {
				if err := validateLinearSearchValues(values); err != nil {
					return nil, err
				}
			}
			if len(payload.Arr) > 0 {
				arr = payload.Arr
			} else if len(payload.Array) > 0 {
				arr = payload.Array
			}
			if payload.Target != nil {
				target = *payload.Target
			}
		} else {
			var custom []int
			if err := json.Unmarshal(input, &custom); err != nil {
				return nil, fmt.Errorf("invalid search array: %w", err)
			}
			if err := validateLinearSearchValues(custom); err != nil {
				return nil, err
			}
			if len(custom) > 0 {
				arr = custom
			}
		}
	}
	if target < -1000000 || target > 1000000 {
		return nil, fmt.Errorf("target must be within [-1000000,1000000]")
	}

	return Trace(arr, target), nil
}

func validateLinearSearchValues(values []int) error {
	if len(values) > 128 {
		return fmt.Errorf("search array must contain at most 128 integers")
	}
	for _, value := range values {
		if value < -1000000 || value > 1000000 {
			return fmt.Errorf("array values must be within [-1000000,1000000]")
		}
	}
	return nil
}

func checkLinearSearchNulls(input []byte) error {
	dec := json.NewDecoder(bytes.NewReader(input))
	t, err := dec.Token()
	if err != nil {
		return nil
	}
	if delim, ok := t.(json.Delim); ok && delim == '[' {
		for dec.More() {
			elem, err := dec.Token()
			if err != nil {
				return nil
			}
			if elem == nil {
				return fmt.Errorf("array elements must not be null")
			}
		}
		return nil
	}
	if delim, ok := t.(json.Delim); ok && delim == '{' {
		for dec.More() {
			keyTok, err := dec.Token()
			if err != nil {
				return nil
			}
			key, _ := keyTok.(string)
			lower := strings.ToLower(key)
			valTok, err := dec.Token()
			if err != nil {
				return nil
			}
			if lower == "arr" || lower == "array" || lower == "target" {
				if valTok == nil {
					return fmt.Errorf("%s must not be null", key)
				}
			}
			if d, ok := valTok.(json.Delim); ok && d == '[' {
				for dec.More() {
					elem, err := dec.Token()
					if err != nil {
						return nil
					}
					if (lower == "arr" || lower == "array") && elem == nil {
						return fmt.Errorf("%s elements must not be null", key)
					}
					if subD, ok := elem.(json.Delim); ok && (subD == '[' || subD == '{') {
						skipDelim(dec, subD)
					}
				}
				dec.Token()
			} else if d, ok := valTok.(json.Delim); ok && d == '{' {
				skipDelim(dec, d)
			}
		}
	}
	return nil
}

func skipDelim(dec *json.Decoder, open json.Delim) {
	depth := 1
	for depth > 0 && dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return
		}
		if d, ok := tok.(json.Delim); ok {
			if d == '{' || d == '[' {
				depth++
			} else if d == '}' || d == ']' {
				depth--
			}
		}
	}
	if depth > 0 {
		dec.Token()
	}
}

func init() {
	catalog.Register(Lesson{})
}
