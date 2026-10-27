package kmp

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	Text    string `json:"text"`
	Pattern string `json:"pattern"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "kmp",
		Name:     "KMP String Matching",
		Category: catalog.CategorySearching,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "An efficient pattern matching algorithm that skips unnecessary comparisons by precomputing a Longest Proper Prefix/Suffix (LPS) array.",
			TimeComplexity:  "O(n + m)",
			SpaceComplexity: "O(m)",
			KeyConcepts:     []string{"Longest Proper Prefix/Suffix", "No Backtracking", "Preprocessing", "Linear Search in Strings"},
			HowItWorks:      "Precomputes an LPS array for the pattern in O(m) time. During search, when a character mismatch occurs, the LPS value specifies how many characters in the pattern can be safely skipped without backtracking the text index, guaranteeing linear O(n + m) total runtime.",
		},
		DefaultInput: `{"text": "ABABDABACDABABCABAB", "pattern": "ABABCABAB"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func KMPSearch(text, pattern string) int {"},
		{ID: "L2", Number: 2, Code: "    n, m := len(text), len(pattern)"},
		{ID: "L3", Number: 3, Code: "    lps := computeLPS(pattern)"},
		{ID: "L4", Number: 4, Code: "    i, j := 0, 0"},
		{ID: "L5", Number: 5, Code: "    for i < n {"},
		{ID: "L6", Number: 6, Code: "        if text[i] == pattern[j] {"},
		{ID: "L7", Number: 7, Code: "            i++; j++"},
		{ID: "L8", Number: 8, Code: "        }"},
		{ID: "L9", Number: 9, Code: "        if j == m {"},
		{ID: "L10", Number: 10, Code: "            return i - j // Match found"},
		{ID: "L11", Number: 11, Code: "        } else if i < n && text[i] != pattern[j] {"},
		{ID: "L12", Number: 12, Code: "            if j != 0 {"},
		{ID: "L13", Number: 13, Code: "                j = lps[j-1]"},
		{ID: "L14", Number: 14, Code: "            } else {"},
		{ID: "L15", Number: 15, Code: "                i++"},
		{ID: "L16", Number: 16, Code: "            }"},
		{ID: "L17", Number: 17, Code: "        }"},
		{ID: "L18", Number: 18, Code: "    }"},
		{ID: "L19", Number: 19, Code: "    return -1 // Not found"},
		{ID: "L20", Number: 20, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	text := "ABABDABACDABABCABAB"
	pattern := "ABABCABAB"

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) > 0 {
		if bytes.Equal(trimmed, []byte("null")) {
			return nil, fmt.Errorf("input cannot be null")
		}

		if trimmed[0] == '{' {
			var custom LessonInput
			if err := json.Unmarshal(trimmed, &custom); err != nil {
				return nil, fmt.Errorf("malformed JSON object: %w", err)
			}
			if custom.Text == "" {
				return nil, fmt.Errorf("text cannot be empty")
			}
			if custom.Pattern == "" {
				return nil, fmt.Errorf("pattern cannot be empty")
			}
			text = custom.Text
			pattern = custom.Pattern
		} else if trimmed[0] == '[' {
			var sliceInput []string
			if err := json.Unmarshal(trimmed, &sliceInput); err != nil {
				return nil, fmt.Errorf("malformed JSON array: %w", err)
			}
			if len(sliceInput) < 2 {
				return nil, fmt.Errorf("input array must contain at least [text, pattern]")
			}
			if sliceInput[0] == "" {
				return nil, fmt.Errorf("text cannot be empty")
			}
			if sliceInput[1] == "" {
				return nil, fmt.Errorf("pattern cannot be empty")
			}
			text = sliceInput[0]
			pattern = sliceInput[1]
		} else {
			return nil, fmt.Errorf("malformed JSON input: expected JSON object or array")
		}
	}

	return Trace(text, pattern), nil
}

func init() {
	catalog.Register(Lesson{})
}
