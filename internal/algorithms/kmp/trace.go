package kmp

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func cloneTableState(s renderers.TableState) renderers.TableState {
	rowHeaders := make([]string, len(s.RowHeaders))
	copy(rowHeaders, s.RowHeaders)

	colHeaders := make([]string, len(s.ColHeaders))
	copy(colHeaders, s.ColHeaders)

	cells := make([]renderers.TableCell, len(s.Cells))
	copy(cells, s.Cells)

	var deps []renderers.TableDependency
	if len(s.Dependencies) > 0 {
		deps = make([]renderers.TableDependency, len(s.Dependencies))
		copy(deps, s.Dependencies)
	}

	return renderers.TableState{
		Caption:      s.Caption,
		Dependencies: deps,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

func buildTableState(text, pattern string, lps []int, i, j int, phase, caption string) renderers.TableState {
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
		numCols = 1
	}

	colHeaders := make([]string, numCols)
	for c := 0; c < numCols; c++ {
		colHeaders[c] = strconv.Itoa(c)
	}

	rowHeaders := []string{"Text", "Pattern", "LPS"}
	cells := make([]renderers.TableCell, 0, n+2*m)

	// Row 0: Text
	for c := 0; c < n; c++ {
		cellState := "normal"
		switch phase {
		case "match":
			if c >= shift && c < shift+m {
				cellState = "computed"
			}
		case "compare":
			if c == i {
				cellState = "active"
			} else if c >= shift && c < shift+j {
				cellState = "computed"
			}
		case "char_match":
			if c >= shift && c < shift+j {
				cellState = "computed"
			}
		case "mismatch":
			if c == i {
				cellState = "source"
			} else if c >= shift && c < shift+j {
				cellState = "computed"
			}
		case "lps_lookup":
			if c >= shift && c < shift+j {
				cellState = "computed"
			}
		case "fallback_shift":
			if c == i {
				cellState = "active"
			} else if c >= shift && c < shift+j {
				cellState = "computed"
			}
		default:
			if c >= shift && c < shift+j {
				cellState = "computed"
			}
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
		switch phase {
		case "match":
			cellState = "computed"
		case "compare":
			if k == j {
				cellState = "active"
			} else if k < j {
				cellState = "computed"
			}
		case "char_match":
			if k < j {
				cellState = "computed"
			}
		case "mismatch":
			if k == j {
				cellState = "source"
			} else if k < j {
				cellState = "computed"
			}
		case "lps_lookup":
			if k < j {
				cellState = "computed"
			}
		case "fallback_shift":
			if k == j {
				cellState = "active"
			} else if k < j {
				cellState = "computed"
			}
		default:
			if k < j {
				cellState = "computed"
			}
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
		} else if phase == "fallback_shift" && k == j {
			cellState = "active"
		}

		lpsVal := "0"
		if k < len(lps) {
			lpsVal = strconv.Itoa(lps[k])
		}

		cells = append(cells, renderers.TableCell{
			Row:   2,
			Col:   c,
			Value: lpsVal,
			State: cellState,
		})
	}

	return cloneTableState(renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	})
}

// Trace records the step-by-step execution of Knuth-Morris-Pratt string matching.
func Trace(text, pattern string) []trace.Step {
	b := trace.NewBuilder()
	n := len(text)
	m := len(pattern)

	if m == 0 {
		emptyState := renderers.TableState{
			Caption:    "Empty pattern matches at index 0",
			RowHeaders: []string{"Text", "Pattern", "LPS"},
			ColHeaders: []string{"0"},
			Cells: []renderers.TableCell{
				{Row: 0, Col: 0, Value: "", State: "normal"},
			},
		}
		b.AddStep("L1", trace.ActionInit, "Initialize KMP search with empty pattern.", cloneTableState(emptyState), map[string]any{"i": 0, "j": 0, "shift": 0})
		b.AddStep("L10", trace.ActionDone, "Empty pattern matches at index 0.", cloneTableState(emptyState), map[string]any{"i": 0, "j": 0, "shift": 0, "matchIndex": 0})
		return b.Steps()
	}

	if n < m {
		lps := ComputeLPS(pattern)
		b.AddStep("L1", trace.ActionInit, "Initialize KMP search.",
			buildTableState(text, pattern, lps, 0, 0, "init", fmt.Sprintf("Initialize KMP: text length %d < pattern length %d", n, m)),
			map[string]any{"i": 0, "j": 0, "shift": 0})
		b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Text length %d is less than pattern length %d: match impossible.", n, m),
			buildTableState(text, pattern, lps, 0, 0, "init", fmt.Sprintf("Text length %d < pattern length %d: match impossible", n, m)),
			map[string]any{"i": 0, "j": 0, "shift": 0})
		b.AddStep("L19", trace.ActionDone, "Pattern not found in text.",
			buildTableState(text, pattern, lps, 0, 0, "init", "Pattern not found: text shorter than pattern"),
			map[string]any{"i": 0, "j": 0, "shift": 0, "matchIndex": -1})
		return b.Steps()
	}

	lps := ComputeLPS(pattern)

	b.AddStep("L1", trace.ActionInit, "Initialize KMP search with text and pattern.",
		buildTableState(text, pattern, lps, 0, 0, "init", fmt.Sprintf("Initialize KMP: text length n=%d, pattern length m=%d", n, m)),
		map[string]any{"i": 0, "j": 0, "shift": 0})
	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Text length n = %d, pattern length m = %d.", n, m),
		buildTableState(text, pattern, lps, 0, 0, "init", fmt.Sprintf("Text length n=%d, pattern length m=%d", n, m)),
		map[string]any{"i": 0, "j": 0, "shift": 0})
	b.AddStep("L3", trace.ActionSet, fmt.Sprintf("Precomputed LPS (Longest Proper Prefix/Suffix) array: %v.", lps),
		buildTableState(text, pattern, lps, 0, 0, "init", fmt.Sprintf("Precomputed LPS table: %v", lps)),
		map[string]any{"i": 0, "j": 0, "shift": 0, "lps": lps})
	b.AddStep("L4", trace.ActionSet, "Initialized text pointer i = 0 and pattern pointer j = 0.",
		buildTableState(text, pattern, lps, 0, 0, "init", "Pointers initialized: text index i=0, pattern index j=0"),
		map[string]any{"i": 0, "j": 0, "shift": 0})

	i := 0
	j := 0
	found := false

	for i < n {
		b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Loop check: text index i = %d < %d.", i, n),
			buildTableState(text, pattern, lps, i, j, "normal", fmt.Sprintf("Scanning: text index i=%d < %d (alignment shift=%d)", i, n, i-j)),
			map[string]any{"i": i, "j": j, "shift": i - j})

		b.IncComparisons()
		b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Comparing text[%d] ('%c') with pattern[%d] ('%c') at alignment shift = %d.", i, text[i], j, pattern[j], i-j),
			buildTableState(text, pattern, lps, i, j, "compare", fmt.Sprintf("Compare: text[%d]='%c' vs pattern[%d]='%c' (shift=%d)", i, text[i], j, pattern[j], i-j)),
			map[string]any{"i": i, "j": j, "shift": i - j})

		if text[i] == pattern[j] {
			i++
			j++
			b.AddStep("L7", trace.ActionSet, fmt.Sprintf("Character match: text[%d] ('%c') == pattern[%d] ('%c'). Advance text pointer i to %d and pattern pointer j to %d.", i-1, text[i-1], j-1, pattern[j-1], i, j),
				buildTableState(text, pattern, lps, i, j, "char_match", fmt.Sprintf("Match: '%c' at text[%d] == pattern[%d]. Advanced to i=%d, j=%d", text[i-1], i-1, j-1, i, j)),
				map[string]any{"i": i, "j": j, "shift": i - j})
		}

		if j == m {
			matchIdx := i - j
			b.AddStep("L9", trace.ActionHighlight, fmt.Sprintf("Pattern index j reaches length m = %d: full pattern matched!", m),
				buildTableState(text, pattern, lps, i, j, "match", fmt.Sprintf("Full match! All %d pattern characters matched at shift=%d", m, matchIdx)),
				map[string]any{"i": i, "j": j, "shift": matchIdx})

			b.AddStep("L10", trace.ActionDone, fmt.Sprintf("Full pattern '%s' found in text starting at index %d (i - j = %d - %d = %d).", pattern, matchIdx, i, j, matchIdx),
				buildTableState(text, pattern, lps, i, j, "match", fmt.Sprintf("Match found at index %d", matchIdx)),
				map[string]any{"i": i, "j": j, "shift": matchIdx, "matchIndex": matchIdx})
			found = true
			break
		} else if i < n && text[i] != pattern[j] {
			var mismatchExpl string
			if j > 0 {
				mismatchExpl = fmt.Sprintf("Character mismatch: text[%d] ('%c') != pattern[%d] ('%c'). Crucially, KMP does NOT backtrack text pointer i = %d! Because pattern[0..%d] already matched text[%d..%d], KMP uses the precomputed LPS table to shift pattern safely.", i, text[i], j, pattern[j], i, j-1, i-j, i-1)
			} else {
				mismatchExpl = fmt.Sprintf("Character mismatch: text[%d] ('%c') != pattern[%d] ('%c'). No prefix matched yet (j = 0). Text pointer i = %d does not backtrack; advance will proceed to next text position.", i, text[i], j, pattern[j], i)
			}
			b.AddStep("L11", trace.ActionHighlight, mismatchExpl,
				buildTableState(text, pattern, lps, i, j, "mismatch", fmt.Sprintf("Mismatch: text[%d]='%c' != pattern[%d]='%c'. Text pointer i=%d NOT backtracked", i, text[i], j, pattern[j], i)),
				map[string]any{"i": i, "j": j, "shift": i - j})

			if j != 0 {
				lookupIdx := j - 1
				nextJ := lps[lookupIdx]
				b.AddStep("L12", trace.ActionHighlight, fmt.Sprintf("Pattern pointer j = %d > 0: query LPS table at index j-1 = %d (longest proper prefix that is also a suffix).", j, lookupIdx),
					buildTableState(text, pattern, lps, i, j, "lps_lookup", fmt.Sprintf("LPS query: lps[j-1] = lps[%d] = %d (prefix length to preserve)", lookupIdx, nextJ)),
					map[string]any{"i": i, "j": j, "shift": i - j})

				b.IncOperations()
				prevJ := j
				j = nextJ
				b.AddStep("L13", trace.ActionSet, fmt.Sprintf("LPS fallback: update j = lps[%d] = %d. Pattern alignment shifts from %d to %d (shift = %d). Text pointer i = %d remains unchanged without backtracking.", lookupIdx, nextJ, i-prevJ, i-j, i-j, i),
					buildTableState(text, pattern, lps, i, j, "fallback_shift", fmt.Sprintf("LPS shift: j=%d -> %d. Pattern shifted right to align %d; text i=%d holds", prevJ, j, i-j, i)),
					map[string]any{"i": i, "j": j, "shift": i - j})
			} else {
				b.AddStep("L14", trace.ActionHighlight, fmt.Sprintf("Pattern pointer j = 0: no matched prefix to fall back on. Pattern cannot shift further at text index %d, must advance text pointer.", i),
					buildTableState(text, pattern, lps, i, j, "normal", fmt.Sprintf("Mismatch at j=0: cannot shift pattern further, advance text index i=%d", i)),
					map[string]any{"i": i, "j": j, "shift": i - j})

				b.IncOperations()
				i++
				b.AddStep("L15", trace.ActionSet, fmt.Sprintf("Advanced text pointer i to %d (new shift = %d).", i, i-j),
					buildTableState(text, pattern, lps, i, j, "normal", fmt.Sprintf("Advanced text index i to %d", i)),
					map[string]any{"i": i, "j": j, "shift": i - j})
			}
		}
	}

	if !found {
		b.AddStep("L18", trace.ActionHighlight, fmt.Sprintf("Text exhausted: text index i = %d reached end of text without finding a full match.", i),
			buildTableState(text, pattern, lps, i, j, "normal", "Text exhausted: search loop completed"),
			map[string]any{"i": i, "j": j, "shift": i - j})

		b.AddStep("L19", trace.ActionDone, "KMP search finished: pattern not found in text.",
			buildTableState(text, pattern, lps, i, j, "normal", "Pattern not found in text"),
			map[string]any{"i": i, "j": j, "shift": i - j, "matchIndex": -1})
	}

	return b.Steps()
}
