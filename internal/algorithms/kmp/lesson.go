package kmp

import (
	"encoding/json"

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

	if len(input) > 0 {
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err == nil && custom.Text != "" && custom.Pattern != "" {
			text = custom.Text
			pattern = custom.Pattern
		} else {
			var sliceInput []string
			if err := json.Unmarshal(input, &sliceInput); err == nil && len(sliceInput) >= 2 {
				text = sliceInput[0]
				pattern = sliceInput[1]
			}
		}
	}

	return Trace(text, pattern), nil
}

func init() {
	catalog.Register(Lesson{})
}
