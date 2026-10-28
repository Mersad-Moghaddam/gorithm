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
	lcsCells map[[2]int]bool,
) renderers.LCSLessonState {
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
			if lcsCells != nil && lcsCells[[2]int{r, c}] {
				state = "computed"
			} else if isSource[[2]int{r, c}] {
				state = "source"
			} else if r == activeRow && c == activeCol {
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

	table := renderers.TableState{
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
	state := renderers.LCSLessonState{TableState: table, S1: s1, S2: s2, Row: activeRow, Col: activeCol,
		From: [2]int{activeRow, activeCol}, To: [2]int{activeRow, activeCol}, Sources: append([][2]int(nil), sources...)}
	for r := range dpTable {
		state.DP = append(state.DP, append([]int(nil), dpTable[r]...))
		state.Filled = append(state.Filled, append([]bool(nil), filled[r]...))
		for c := range dpTable[r] {
			if lcsCells[[2]int{r, c}] && r > 0 && c > 0 {
				state.Matches = append(state.Matches, [2]int{r - 1, c - 1})
				state.Sequence += s1[r-1 : r]
			}
		}
	}
	return state
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

	b.AddStep(
		"L3",
		trace.ActionInit,
		fmt.Sprintf("Initialize (%d+1)x(%d+1) DP table with 0s for base cases.", m, n),
		buildTableState(s1, s2, dp, filled, -1, -1, nil, nil),
		map[string]any{"m": m, "n": n, "s1": s1, "s2": s2},
	)

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			b.IncComparisons()
			char1 := s1[i-1]
			char2 := s2[j-1]

			if char1 == char2 {
				sources := [][2]int{{i - 1, j - 1}}
				b.AddStep(
					"L7",
					trace.ActionCompare,
					fmt.Sprintf("s1[%d] ('%c') == s2[%d] ('%c'): match! Source is diagonal dp[%d][%d].", i-1, char1, j-1, char2, i-1, j-1),
					buildTableState(s1, s2, dp, filled, i, j, sources, nil),
					map[string]any{"i": i, "j": j, "char1": string(char1), "char2": string(char2)},
				)

				dp[i][j] = dp[i-1][j-1] + 1
				filled[i][j] = true
				b.IncOperations()
				b.AddStep(
					"L8",
					trace.ActionSet,
					fmt.Sprintf("Set dp[%d][%d] = dp[%d][%d] + 1 = %d.", i, j, i-1, j-1, dp[i][j]),
					buildTableState(s1, s2, dp, filled, i, j, sources, nil),
					map[string]any{"i": i, "j": j, "val": dp[i][j]},
				)
			} else {
				topVal := dp[i-1][j]
				leftVal := dp[i][j-1]
				sources := [][2]int{{i - 1, j}, {i, j - 1}}
				b.AddStep(
					"L7",
					trace.ActionCompare,
					fmt.Sprintf("s1[%d] ('%c') != s2[%d] ('%c'): mismatch. Compare top dp[%d][%d]=%d and left dp[%d][%d]=%d.", i-1, char1, j-1, char2, i-1, j, topVal, i, j-1, leftVal),
					buildTableState(s1, s2, dp, filled, i, j, sources, nil),
					map[string]any{"i": i, "j": j, "char1": string(char1), "char2": string(char2)},
				)

				if topVal >= leftVal {
					dp[i][j] = topVal
				} else {
					dp[i][j] = leftVal
				}
				filled[i][j] = true
				b.IncOperations()
				b.AddStep(
					"L10",
					trace.ActionSet,
					fmt.Sprintf("Set dp[%d][%d] = max(top %d, left %d) = %d.", i, j, topVal, leftVal, dp[i][j]),
					buildTableState(s1, s2, dp, filled, i, j, sources, nil),
					map[string]any{"i": i, "j": j, "val": dp[i][j]},
				)
			}
		}
	}

	lcsCells := make(map[[2]int]bool)
	var lcsChars []byte
	bi, bj := m, n

	b.AddStep(
		"L14",
		trace.ActionInit,
		fmt.Sprintf("Begin backtracking from dp[%d][%d] = %d to reconstruct LCS.", bi, bj, dp[bi][bj]),
		buildTableState(s1, s2, dp, filled, bi, bj, nil, lcsCells),
		map[string]any{"i": bi, "j": bj, "lcs": ""},
	)

	for bi > 0 && bj > 0 {
		b.IncComparisons()
		c1 := s1[bi-1]
		c2 := s2[bj-1]

		b.AddStep(
			"L16",
			trace.ActionCompare,
			fmt.Sprintf("Backtrack at (%d, %d): compare s1[%d] ('%c') with s2[%d] ('%c').", bi, bj, bi-1, c1, bj-1, c2),
			buildTableState(s1, s2, dp, filled, bi, bj, nil, lcsCells),
			map[string]any{"i": bi, "j": bj, "char1": string(c1), "char2": string(c2)},
		)

		if c1 == c2 {
			lcsChars = append([]byte{c1}, lcsChars...)
			lcsCells[[2]int{bi, bj}] = true
			b.IncOperations()
			b.AddStep(
				"L17",
				trace.ActionHighlight,
				fmt.Sprintf("Match '%c'! Part of LCS. Current sequence: %q. Move diagonally to (%d, %d).", c1, string(lcsChars), bi-1, bj-1),
				buildTableState(s1, s2, dp, filled, bi, bj, nil, lcsCells),
				map[string]any{"i": bi, "j": bj, "lcs": string(lcsChars)},
			)
			bi--
			bj--
		} else if dp[bi-1][bj] >= dp[bi][bj-1] {
			b.AddStep(
				"L19",
				trace.ActionVisit,
				fmt.Sprintf("dp[%d][%d] (%d) >= dp[%d][%d] (%d). Move up to row %d.", bi-1, bj, dp[bi-1][bj], bi, bj-1, dp[bi][bj-1], bi-1),
				buildTableState(s1, s2, dp, filled, bi-1, bj, [][2]int{{bi, bj}}, lcsCells),
				map[string]any{"i": bi, "j": bj},
			)
			bi--
		} else {
			b.AddStep(
				"L21",
				trace.ActionVisit,
				fmt.Sprintf("dp[%d][%d] (%d) < dp[%d][%d] (%d). Move left to col %d.", bi-1, bj, dp[bi-1][bj], bi, bj-1, dp[bi][bj-1], bj-1),
				buildTableState(s1, s2, dp, filled, bi, bj-1, [][2]int{{bi, bj}}, lcsCells),
				map[string]any{"i": bi, "j": bj},
			)
			bj--
		}
	}

	finalLCS := string(lcsChars)
	b.AddStep(
		"L24",
		trace.ActionDone,
		fmt.Sprintf("Completed! Longest Common Subsequence is %q (length %d).", finalLCS, len(finalLCS)),
		buildTableState(s1, s2, dp, filled, -1, -1, nil, lcsCells),
		map[string]any{
			"lcs":       finalLCS,
			"lcsLength": len(finalLCS),
			"s1":        s1,
			"s2":        s2,
		},
	)

	steps := b.Steps()
	for i := range steps {
		step := &steps[i]
		s := step.State.(renderers.LCSLessonState)
		switch step.LineID {
		case "L3":
			s.Phase = "allocate"
		case "L7":
			s.Phase = "compare"
			step.HoldMS = 600
		case "L8", "L10":
			s.Phase = "write"
			step.HoldMS = 1200
		case "L14", "L16":
			s.Phase = "backtrack"
			if step.LineID == "L16" {
				step.HoldMS = 600
			}
		case "L17":
			s.Phase = "match"
			s.To = [2]int{s.Row - 1, s.Col - 1}
			step.HoldMS = 1200
		case "L19", "L21":
			s.From = [2]int{step.Variables["i"].(int), step.Variables["j"].(int)}
			s.Phase = "up"
			if step.LineID == "L21" {
				s.Phase = "left"
			}
			step.HoldMS = 1200
		case "L24":
			s.Phase = "done"
		}
		if s.From[0] > 0 && s.From[1] > 0 {
			s.Match = s1[s.From[0]-1] == s2[s.From[1]-1]
		}
		step.State = s
	}
	return steps
}
