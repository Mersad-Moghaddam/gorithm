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
		{ID: "empty", Number: 3, Code: "    if m == 0 { return 0 }"},
		{ID: "short", Number: 4, Code: "    if n < m { return -1 }"},
		{ID: "L3", Number: 5, Code: "    lps := computeLPS(pattern)"},
		{ID: "L4", Number: 6, Code: "    i, j := 0, 0"},
		{ID: "L5", Number: 7, Code: "    for i < n {"},
		{ID: "L6", Number: 8, Code: "        if text[i] == pattern[j] {"},
		{ID: "L7", Number: 9, Code: "            i++; j++"},
		{ID: "L8", Number: 10, Code: "        }"},
		{ID: "L9", Number: 11, Code: "        if j == m {"},
		{ID: "L10", Number: 12, Code: "            return i - j // Match found"},
		{ID: "L11", Number: 13, Code: "        } else if i < n && text[i] != pattern[j] {"},
		{ID: "L12", Number: 14, Code: "            if j != 0 {"},
		{ID: "L13", Number: 15, Code: "                j = lps[j-1]"},
		{ID: "L14", Number: 16, Code: "            } else {"},
		{ID: "L15", Number: 17, Code: "                i++"},
		{ID: "L16", Number: 18, Code: "            }"},
		{ID: "L17", Number: 19, Code: "        }"},
		{ID: "L18", Number: 20, Code: "    }"},
		{ID: "L19", Number: 21, Code: "    return -1 // Not found"},
		{ID: "L20", Number: 22, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	text := "ABABDABACDABABCABAB"
	pattern := "ABABCABAB"

	if len(input) > 0 {
		raw := bytes.TrimSpace(input)
		if len(raw) == 0 {
			return nil, fmt.Errorf("KMP input must be a JSON object or string array")
		}
		switch raw[0] {
		case '{':
			var fields map[string]json.RawMessage
			if err := json.Unmarshal(raw, &fields); err != nil {
				return nil, err
			}
			for _, field := range []struct {
				name string
				dst  *string
			}{{"text", &text}, {"pattern", &pattern}} {
				if value, ok := fields[field.name]; ok {
					if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
						return nil, fmt.Errorf("%s must be a string", field.name)
					}
					if err := json.Unmarshal(value, field.dst); err != nil {
						return nil, err
					}
				}
			}
		case '[':
			var values []json.RawMessage
			if err := json.Unmarshal(raw, &values); err != nil {
				return nil, err
			}
			if len(values) < 2 {
				return nil, fmt.Errorf("KMP array needs text and pattern")
			}
			for k, value := range values {
				var str string
				if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
					return nil, fmt.Errorf("KMP array entries must be strings")
				}
				if err := json.Unmarshal(value, &str); err != nil {
					return nil, err
				}
				if k == 0 {
					text = str
				}
				if k == 1 {
					pattern = str
				}
			}
		default:
			return nil, fmt.Errorf("KMP input must be a JSON object or string array")
		}
	}
	if len(text) > 256 || len(pattern) > 64 {
		return nil, fmt.Errorf("KMP limits: text 256 bytes, pattern 64 bytes")
	}

	return Trace(text, pattern), nil
}

func init() {
	catalog.Register(Lesson{})
}
