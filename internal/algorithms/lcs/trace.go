package lcs

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func buildTableState(
	s1, s2 string,
	dpTable [][]int,
	filled [][]bool,
	activeRow, activeCol int,
	sources [][2]int,
	dependencies []renderers.TableDependency,
	caption string,
	lcsCells map[[2]int]bool,
) renderers.TableState {
	m := len(s1)
	n := len(s2)

	rowHeaders := make([]string, m+1)
	rowHeaders[0] = "-"
	for i := 0; i < m; i++ {
		rowHeaders[i+1] = string(s1[i])
	}

	colHeaders := make([]string, n+1)
	colHeaders[0] = "-"
	for j := 0; j < n; j++ {
		colHeaders[j+1] = string(s2[j])
	}

	isSource := make(map[[2]int]bool, len(sources))
	for _, s := range sources {
		isSource[s] = true
	}

	cells := make([]renderers.TableCell, 0, (m+1)*(n+1))
	for r := 0; r <= m; r++ {
		for c := 0; c <= n; c++ {
			val := ""
			if filled[r][c] {
				val = strconv.Itoa(dpTable[r][c])
			}

			state := "normal"
			if isSource[[2]int{r, c}] {
				state = "source"
			}
			if lcsCells != nil && lcsCells[[2]int{r, c}] {
				state = "computed"
			}
			if r == activeRow && c == activeCol {
				state = "active"
			}

			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: val,
				State: state,
			})
		}
	}

	var deps []renderers.TableDependency
	if len(dependencies) > 0 {
		deps = make([]renderers.TableDependency, len(dependencies))
		copy(deps, dependencies)
	}

	return renderers.TableState{
		Caption:      caption,
		Dependencies: deps,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

// Trace simulates the LCS dynamic programming algorithm step by step.
// ponytail: quadratic trace table state recorded per step; could be streamed if step count becomes memory bottleneck.
func Trace(s1, s2 string) []trace.Step {
	b := trace.NewBuilder()
	m, n := len(s1), len(s2)

	dp := make([][]int, m+1)
	filled := make([][]bool, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
		filled[i] = make([]bool, n+1)
	}

	for i := 0; i <= m; i++ {
		filled[i][0] = true
	}
	for j := 0; j <= n; j++ {
		filled[0][j] = true
	}

	initDesc := fmt.Sprintf("Initialize (%d+1)x(%d+1) DP table with 0s for base cases.", m, n)
	b.AddStep(
		"L3",
		trace.ActionInit,
		initDesc,
		buildTableState(s1, s2, dp, filled, -1, -1, nil, nil, initDesc, nil),
		map[string]any{"m": m, "n": n, "s1": s1, "s2": s2},
	)

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			b.IncComparisons()
			char1 := s1[i-1]
			char2 := s2[j-1]

			if char1 == char2 {
				sources := [][2]int{{i - 1, j - 1}}
				deps := []renderers.TableDependency{{FromRow: i - 1, FromCol: j - 1, ToRow: i, ToCol: j}}
				caption := fmt.Sprintf("Match '%c' == '%c': extend diagonal LCS length by 1", char1, char2)

				b.AddStep(
					"L7",
					trace.ActionCompare,
					caption,
					buildTableState(s1, s2, dp, filled, i, j, sources, deps, caption, nil),
					map[string]any{"i": i, "j": j, "char1": string(char1), "char2": string(char2)},
				)

				dp[i][j] = dp[i-1][j-1] + 1
				filled[i][j] = true
				b.IncOperations()
				b.AddStep(
					"L8",
					trace.ActionSet,
					fmt.Sprintf("Set dp[%d][%d] = dp[%d][%d] + 1 = %d.", i, j, i-1, j-1, dp[i][j]),
					buildTableState(s1, s2, dp, filled, i, j, sources, deps, caption, nil),
					map[string]any{"i": i, "j": j, "val": dp[i][j]},
				)
			} else {
				topVal := dp[i-1][j]
				leftVal := dp[i][j-1]
				maxVal := topVal
				if leftVal > maxVal {
					maxVal = leftVal
				}
				sources := [][2]int{{i - 1, j}, {i, j - 1}}
				deps := []renderers.TableDependency{
					{FromRow: i - 1, FromCol: j, ToRow: i, ToCol: j},
					{FromRow: i, FromCol: j - 1, ToRow: i, ToCol: j},
				}
				caption := fmt.Sprintf("Mismatch '%c' != '%c': take max(top=%d, left=%d) = %d", char1, char2, topVal, leftVal, maxVal)

				b.AddStep(
					"L7",
					trace.ActionCompare,
					caption,
					buildTableState(s1, s2, dp, filled, i, j, sources, deps, caption, nil),
					map[string]any{"i": i, "j": j, "char1": string(char1), "char2": string(char2)},
				)

				dp[i][j] = maxVal
				filled[i][j] = true
				b.IncOperations()
				b.AddStep(
					"L10",
					trace.ActionSet,
					fmt.Sprintf("Set dp[%d][%d] = max(top %d, left %d) = %d.", i, j, topVal, leftVal, dp[i][j]),
					buildTableState(s1, s2, dp, filled, i, j, sources, deps, caption, nil),
					map[string]any{"i": i, "j": j, "val": dp[i][j]},
				)
			}
		}
	}

	lcsCells := make(map[[2]int]bool)
	var lcsChars []byte
	bi, bj := m, n

	backtrackInitCaption := fmt.Sprintf("Begin backtracking from dp[%d][%d] = %d to reconstruct LCS.", bi, bj, dp[bi][bj])
	b.AddStep(
		"L14",
		trace.ActionInit,
		backtrackInitCaption,
		buildTableState(s1, s2, dp, filled, bi, bj, nil, nil, backtrackInitCaption, lcsCells),
		map[string]any{"i": bi, "j": bj, "lcs": ""},
	)

	for bi > 0 && bj > 0 {
		b.IncComparisons()
		c1 := s1[bi-1]
		c2 := s2[bj-1]

		compCaption := fmt.Sprintf("Backtrack at (%d, %d): compare s1[%d] ('%c') with s2[%d] ('%c').", bi, bj, bi-1, c1, bj-1, c2)
		b.AddStep(
			"L16",
			trace.ActionCompare,
			compCaption,
			buildTableState(s1, s2, dp, filled, bi, bj, nil, nil, compCaption, lcsCells),
			map[string]any{"i": bi, "j": bj, "char1": string(c1), "char2": string(c2), "lcs": string(lcsChars)},
		)

		if c1 == c2 {
			lcsChars = append([]byte{c1}, lcsChars...)
			lcsCells[[2]int{bi, bj}] = true
			b.IncOperations()

			arrow := []renderers.TableDependency{{FromRow: bi, FromCol: bj, ToRow: bi - 1, ToCol: bj - 1}}
			matchCaption := fmt.Sprintf("Match '%c'! Part of LCS. Current sequence: %q. Move diagonally to (%d, %d).", c1, string(lcsChars), bi-1, bj-1)
			b.AddStep(
				"L17",
				trace.ActionHighlight,
				matchCaption,
				buildTableState(s1, s2, dp, filled, bi, bj, nil, arrow, matchCaption, lcsCells),
				map[string]any{"i": bi, "j": bj, "char": string(c1), "lcs": string(lcsChars)},
			)
			bi--
			bj--
		} else if dp[bi-1][bj] >= dp[bi][bj-1] {
			arrow := []renderers.TableDependency{{FromRow: bi, FromCol: bj, ToRow: bi - 1, ToCol: bj}}
			upCaption := fmt.Sprintf("dp[%d][%d] (%d) >= dp[%d][%d] (%d). Move up to row %d.", bi-1, bj, dp[bi-1][bj], bi, bj-1, dp[bi][bj-1], bi-1)
			b.AddStep(
				"L19",
				trace.ActionVisit,
				upCaption,
				buildTableState(s1, s2, dp, filled, bi-1, bj, [][2]int{{bi, bj}}, arrow, upCaption, lcsCells),
				map[string]any{"i": bi, "j": bj, "lcs": string(lcsChars)},
			)
			bi--
		} else {
			arrow := []renderers.TableDependency{{FromRow: bi, FromCol: bj, ToRow: bi, ToCol: bj - 1}}
			leftCaption := fmt.Sprintf("dp[%d][%d] (%d) < dp[%d][%d] (%d). Move left to col %d.", bi-1, bj, dp[bi-1][bj], bi, bj-1, dp[bi][bj-1], bj-1)
			b.AddStep(
				"L21",
				trace.ActionVisit,
				leftCaption,
				buildTableState(s1, s2, dp, filled, bi, bj-1, [][2]int{{bi, bj}}, arrow, leftCaption, lcsCells),
				map[string]any{"i": bi, "j": bj, "lcs": string(lcsChars)},
			)
			bj--
		}
	}

	finalLCS := string(lcsChars)
	doneCaption := fmt.Sprintf("Completed! Longest Common Subsequence is %q (length %d).", finalLCS, len(finalLCS))
	b.AddStep(
		"L24",
		trace.ActionDone,
		doneCaption,
		buildTableState(s1, s2, dp, filled, -1, -1, nil, nil, doneCaption, lcsCells),
		map[string]any{
			"lcs":       finalLCS,
			"lcsLength": len(finalLCS),
			"s1":        s1,
			"s2":        s2,
		},
	)

	return b.Steps()
}
