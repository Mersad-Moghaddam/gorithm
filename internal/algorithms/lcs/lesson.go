package lcs

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	S1      string `json:"s1"`
	S2      string `json:"s2"`
	Text    string `json:"text"`
	Pattern string `json:"pattern"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "lcs",
		Name:     "Longest Common Subsequence",
		Category: catalog.CategoryDP,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Finds the longest subsequence present in two given strings in the same relative order using 2D dynamic programming and backtracking.",
			TimeComplexity:  "O(m · n)",
			SpaceComplexity: "O(m · n)",
			KeyConcepts: []string{
				"2D Table",
				"Optimal Substructure",
				"Character Match/Mismatch Transition",
				"Backtracking",
			},
			HowItWorks: "Builds an (m+1) x (n+1) table where dp[i][j] stores the LCS length of prefixes s1[0..i-1] and s2[0..j-1]. When characters match (s1[i-1] == s2[j-1]), dp[i][j] = dp[i-1][j-1] + 1. On mismatch, dp[i][j] = max(dp[i-1][j], dp[i][j-1]). After filling the table, backtracking from dp[m][n] reconstructs the LCS sequence.",
		},
		DefaultInput: `{"s1": "AGGTAB", "s2": "GXTXAYB"}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func LCS(s1, s2 string) string {"},
		{ID: "L2", Number: 2, Code: "    m, n := len(s1), len(s2)"},
		{ID: "L3", Number: 3, Code: "    dp := make([][]int, m+1)"},
		{ID: "L4", Number: 4, Code: "    for i := range dp { dp[i] = make([]int, n+1) }"},
		{ID: "L5", Number: 5, Code: "    for i := 1; i <= m; i++ {"},
		{ID: "L6", Number: 6, Code: "        for j := 1; j <= n; j++ {"},
		{ID: "L7", Number: 7, Code: "            if s1[i-1] == s2[j-1] {"},
		{ID: "L8", Number: 8, Code: "                dp[i][j] = dp[i-1][j-1] + 1"},
		{ID: "L9", Number: 9, Code: "            } else {"},
		{ID: "L10", Number: 10, Code: "                dp[i][j] = max(dp[i-1][j], dp[i][j-1])"},
		{ID: "L11", Number: 11, Code: "            }"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "    }"},
		{ID: "L14", Number: 14, Code: "    i, j := m, n; var lcs []byte"},
		{ID: "L15", Number: 15, Code: "    for i > 0 && j > 0 {"},
		{ID: "L16", Number: 16, Code: "        if s1[i-1] == s2[j-1] {"},
		{ID: "L17", Number: 17, Code: "            lcs = append([]byte{s1[i-1]}, lcs...); i--; j--"},
		{ID: "L18", Number: 18, Code: "        } else if dp[i-1][j] >= dp[i][j-1] {"},
		{ID: "L19", Number: 19, Code: "            i--"},
		{ID: "L20", Number: 20, Code: "        } else {"},
		{ID: "L21", Number: 21, Code: "            j--"},
		{ID: "L22", Number: 22, Code: "        }"},
		{ID: "L23", Number: 23, Code: "    }"},
		{ID: "L24", Number: 24, Code: "    return string(lcs)"},
		{ID: "L25", Number: 25, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	s1 := "AGGTAB"
	s2 := "GXTXAYB"

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
			if custom.S1 != "" || custom.S2 != "" {
				if custom.S1 == "" || custom.S2 == "" {
					return nil, fmt.Errorf("both s1 and s2 must be non-empty")
				}
				s1 = custom.S1
				s2 = custom.S2
			} else if custom.Text != "" || custom.Pattern != "" {
				if custom.Text == "" || custom.Pattern == "" {
					return nil, fmt.Errorf("both text and pattern must be non-empty")
				}
				s1 = custom.Text
				s2 = custom.Pattern
			} else {
				return nil, fmt.Errorf("input strings cannot be empty")
			}
		} else if trimmed[0] == '[' {
			var sliceInput []string
			if err := json.Unmarshal(trimmed, &sliceInput); err != nil {
				return nil, fmt.Errorf("malformed JSON array: %w", err)
			}
			if len(sliceInput) < 2 {
				return nil, fmt.Errorf("input array must contain at least 2 strings")
			}
			if sliceInput[0] == "" || sliceInput[1] == "" {
				return nil, fmt.Errorf("input strings cannot be empty")
			}
			s1 = sliceInput[0]
			s2 = sliceInput[1]
		} else {
			return nil, fmt.Errorf("malformed JSON input: expected JSON object or array")
		}
	}

	return Trace(s1, s2), nil
}

func init() {
	catalog.Register(Lesson{})
}
