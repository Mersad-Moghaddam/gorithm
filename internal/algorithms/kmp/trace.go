package kmp

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func buildTableState(text, pattern string, lps []int, i, j int, phase string) renderers.TableState {
	shift := i - j
	if shift < 0 {
		shift = 0
	}
	n := len(text)
	m := len(pattern)
	numCols := n
	if shift+m > numCols {
		numCols = shift + m
	}
	if numCols == 0 {
		return renderers.TableState{}
	}

	colHeaders := make([]string, numCols)
	for c := 0; c < numCols; c++ {
		colHeaders[c] = strconv.Itoa(c)
	}

	cells := make([]renderers.TableCell, 0, n+2*m)

	// Row 0: Text
	for c := 0; c < n; c++ {
		cellState := "normal"
		if phase == "match" && c >= shift && c < shift+m {
			cellState = "computed"
		} else if phase == "compare" && c == i {
			cellState = "active"
		} else if phase == "char_match" && c == i-1 {
			cellState = "computed"
		} else if phase == "mismatch" && c == i {
			cellState = "source"
		} else if c >= shift && c < shift+j {
			cellState = "computed"
		}

		cells = append(cells, renderers.TableCell{
			Row:   0,
			Col:   c,
			Value: string(text[c]),
			State: cellState,
		})
	}

	// Row 1: Pattern
	for k := 0; k < m; k++ {
		c := shift + k
		cellState := "normal"
		if phase == "match" {
			cellState = "computed"
		} else if phase == "compare" && k == j {
			cellState = "active"
		} else if phase == "char_match" && k == j-1 {
			cellState = "computed"
		} else if phase == "mismatch" && k == j {
			cellState = "source"
		} else if k < j {
			cellState = "computed"
		}

		cells = append(cells, renderers.TableCell{
			Row:   1,
			Col:   c,
			Value: string(pattern[k]),
			State: cellState,
		})
	}

	// Row 2: LPS
	for k := 0; k < m; k++ {
		c := shift + k
		cellState := "normal"
		if phase == "lps_lookup" && k == j-1 {
			cellState = "source"
		}

		cells = append(cells, renderers.TableCell{
			Row:   2,
			Col:   c,
			Value: strconv.Itoa(lps[k]),
			State: cellState,
		})
	}

	return renderers.TableState{
		RowHeaders: []string{"Text", "Pattern", "LPS"},
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace records the step-by-step execution of Knuth-Morris-Pratt string matching.
func Trace(text, pattern string) []trace.Step {
	b := trace.NewBuilder()
	n := len(text)
	m := len(pattern)

	if m == 0 {
		emptyState := renderers.TableState{
			RowHeaders: []string{"Text", "Pattern", "LPS"},
			ColHeaders: []string{"0"},
		}
		b.AddStep("L1", trace.ActionInit, "Initialize KMP search with empty pattern.", emptyState, map[string]any{"i": 0, "j": 0, "shift": 0})
		b.AddStep("L10", trace.ActionDone, "Empty pattern matches at index 0.", emptyState, map[string]any{"i": 0, "j": 0, "shift": 0, "matchIndex": 0})
		return b.Steps()
	}

	if n < m {
		lps := ComputeLPS(pattern)
		state := buildTableState(text, pattern, lps, 0, 0, "init")
		b.AddStep("L1", trace.ActionInit, "Initialize KMP search.", state, map[string]any{"i": 0, "j": 0, "shift": 0})
		b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Text length %d is less than pattern length %d: match impossible.", n, m), state, map[string]any{"i": 0, "j": 0, "shift": 0})
		b.AddStep("L19", trace.ActionDone, "Pattern not found in text.", state, map[string]any{"i": 0, "j": 0, "shift": 0, "matchIndex": -1})
		return b.Steps()
	}

	lps := ComputeLPS(pattern)
	initState := buildTableState(text, pattern, lps, 0, 0, "init")

	b.AddStep("L1", trace.ActionInit, "Initialize KMP search with text and pattern.", initState, map[string]any{"i": 0, "j": 0, "shift": 0})
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Text length n = %d, pattern length m = %d.", n, m), initState, map[string]any{"i": 0, "j": 0, "shift": 0})
	b.AddStep("L3", trace.ActionSet, fmt.Sprintf("Precomputed LPS (Longest Proper Prefix/Suffix) array: %v.", lps), initState, map[string]any{"i": 0, "j": 0, "shift": 0})
	b.AddStep("L4", trace.ActionSet, "Initialized text pointer i = 0 and pattern pointer j = 0.", initState, map[string]any{"i": 0, "j": 0, "shift": 0})

	i := 0
	j := 0
	found := false

	for i < n {
		b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Loop check: text index i = %d < %d.", i, n),
			buildTableState(text, pattern, lps, i, j, "normal"),
			map[string]any{"i": i, "j": j, "shift": i - j})

		b.IncComparisons()
		b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Comparing text[%d] ('%c') with pattern[%d] ('%c').", i, text[i], j, pattern[j]),
			buildTableState(text, pattern, lps, i, j, "compare"),
			map[string]any{"i": i, "j": j, "shift": i - j})

		if text[i] == pattern[j] {
			i++
			j++
			b.AddStep("L7", trace.ActionSet, fmt.Sprintf("Match: '%c' == '%c'. Advanced text pointer i to %d and pattern pointer j to %d.", text[i-1], pattern[j-1], i, j),
				buildTableState(text, pattern, lps, i, j, "char_match"),
				map[string]any{"i": i, "j": j, "shift": i - j})
		}

		if j == m {
			matchIdx := i - j
			b.AddStep("L9", trace.ActionHighlight, fmt.Sprintf("Pattern index j reaches length %d: full pattern matched!", m),
				buildTableState(text, pattern, lps, i, j, "match"),
				map[string]any{"i": i, "j": j, "shift": matchIdx})

			b.AddStep("L10", trace.ActionDone, fmt.Sprintf("Pattern found starting at index %d in text.", matchIdx),
				buildTableState(text, pattern, lps, i, j, "match"),
				map[string]any{"i": i, "j": j, "shift": matchIdx, "matchIndex": matchIdx})
			found = true
			break
		} else if i < n && text[i] != pattern[j] {
			b.AddStep("L11", trace.ActionHighlight, fmt.Sprintf("Mismatch: text[%d] ('%c') != pattern[%d] ('%c').", i, text[i], j, pattern[j]),
				buildTableState(text, pattern, lps, i, j, "mismatch"),
				map[string]any{"i": i, "j": j, "shift": i - j})

			if j != 0 {
				lookupIdx := j - 1
				nextJ := lps[lookupIdx]
				b.AddStep("L12", trace.ActionHighlight, fmt.Sprintf("Pattern pointer j = %d > 0: query LPS table at index j-1 = %d.", j, lookupIdx),
					buildTableState(text, pattern, lps, i, j, "lps_lookup"),
					map[string]any{"i": i, "j": j, "shift": i - j})

				b.IncOperations()
				j = nextJ
				b.AddStep("L13", trace.ActionSet, fmt.Sprintf("LPS[%d] = %d: shift pattern alignment without backtracking text pointer i (new j = %d, shift = %d).", lookupIdx, nextJ, j, i-j),
					buildTableState(text, pattern, lps, i, j, "normal"),
					map[string]any{"i": i, "j": j, "shift": i - j})
			} else {
				b.AddStep("L14", trace.ActionHighlight, "Pattern pointer j = 0: cannot shift pattern further, advance text index i.",
					buildTableState(text, pattern, lps, i, j, "normal"),
					map[string]any{"i": i, "j": j, "shift": i - j})

				b.IncOperations()
				i++
				b.AddStep("L15", trace.ActionSet, fmt.Sprintf("Advanced text index i to %d.", i),
					buildTableState(text, pattern, lps, i, j, "normal"),
					map[string]any{"i": i, "j": j, "shift": i - j})
			}
		}
	}

	if !found {
		b.AddStep("L18", trace.ActionHighlight, "Text exhausted without finding a full pattern match.",
			buildTableState(text, pattern, lps, i, j, "normal"),
			map[string]any{"i": i, "j": j, "shift": i - j})

		b.AddStep("L19", trace.ActionDone, "KMP search finished: pattern not found in text.",
			buildTableState(text, pattern, lps, i, j, "normal"),
			map[string]any{"i": i, "j": j, "shift": i - j, "matchIndex": -1})
	}

	return b.Steps()
}
