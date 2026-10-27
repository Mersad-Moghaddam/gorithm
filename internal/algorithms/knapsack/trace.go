package knapsack

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type cellCoord struct {
	row int
	col int
}

// ponytail: table state builds full matrix for rendering; sparse representation if W grows large.
func makeTableState(
	n, capacity int,
	dp [][]int,
	activeRow, activeCol int,
	activeState string,
	sources []cellCoord,
	isDone bool,
) renderers.TableState {
	rowHeaders := make([]string, n+1)
	rowHeaders[0] = "Item 0"
	for i := 1; i <= n; i++ {
		rowHeaders[i] = fmt.Sprintf("Item %d", i)
	}

	colHeaders := make([]string, capacity+1)
	for c := 0; c <= capacity; c++ {
		colHeaders[c] = strconv.Itoa(c)
	}

	sourceMap := make(map[cellCoord]bool, len(sources))
	for _, s := range sources {
		sourceMap[s] = true
	}

	cells := make([]renderers.TableCell, 0, (n+1)*(capacity+1))
	for r := 0; r <= n; r++ {
		for c := 0; c <= capacity; c++ {
			coord := cellCoord{row: r, col: c}
			val := ""
			state := "normal"

			isComputed := isDone || r == 0 || r < activeRow || (r == activeRow && c < activeCol) || (r == activeRow && c == activeCol && activeState == "computed")

			if isComputed && r < len(dp) && c < len(dp[r]) {
				val = strconv.Itoa(dp[r][c])
			}

			if isDone {
				if r == n && c == capacity {
					state = "computed"
				}
			} else if r == activeRow && c == activeCol {
				state = activeState
			} else if sourceMap[coord] {
				state = "source"
			}

			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: val,
				State: state,
			})
		}
	}

	return renderers.TableState{
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace records the step-by-step execution of 0/1 Knapsack dynamic programming.
func Trace(weights, values []int, capacity int) []trace.Step {
	b := trace.NewBuilder()
	n := min(len(weights), len(values))
	if capacity < 0 {
		capacity = 0
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, capacity+1)
	}

	initState := makeTableState(n, capacity, dp, -1, -1, "normal", nil, false)
	b.AddStep("L1", trace.ActionInit, "Initialize 0/1 Knapsack dynamic programming solver.", initState, map[string]any{"n": n, "capacity": capacity})

	allocState := makeTableState(n, capacity, dp, -1, -1, "normal", nil, false)
	b.AddStep("L3", trace.ActionSet, fmt.Sprintf("Allocate DP matrix of size %dx%d with 0-item base cases.", n+1, capacity+1), allocState, map[string]any{"n": n, "capacity": capacity})

	for i := 1; i <= n; i++ {
		w := weights[i-1]
		v := values[i-1]

		itemState := makeTableState(n, capacity, dp, -1, -1, "normal", nil, false)
		b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Evaluating item %d: weight = %d, value = %d.", i, w, v), itemState, map[string]any{"i": i, "w": w, "v": v})

		for c := 0; c <= capacity; c++ {
			excludeCell := cellCoord{row: i - 1, col: c}
			sources := []cellCoord{excludeCell}

			b.IncComparisons()
			if w <= c {
				includeCell := cellCoord{row: i - 1, col: c - w}
				if includeCell != excludeCell {
					sources = append(sources, includeCell)
				}
				compState := makeTableState(n, capacity, dp, i, c, "active", sources, false)
				b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Item %d weight %d <= capacity %d: compare exclude (%d) vs include (%d + %d = %d).", i, w, c, dp[i-1][c], dp[i-1][c-w], v, dp[i-1][c-w]+v), compState, map[string]any{"i": i, "c": c, "w": w, "v": v, "exclude": dp[i-1][c], "include": dp[i-1][c-w] + v})

				dp[i][c] = max(dp[i-1][c], dp[i-1][c-w]+v)
				b.IncOperations()

				setState := makeTableState(n, capacity, dp, i, c, "computed", sources, false)
				b.AddStep("L9", trace.ActionSet, fmt.Sprintf("dp[%d][%d] = max(%d, %d) = %d.", i, c, dp[i-1][c], dp[i-1][c-w]+v, dp[i][c]), setState, map[string]any{"i": i, "c": c, "dp": dp[i][c]})
			} else {
				compState := makeTableState(n, capacity, dp, i, c, "active", sources, false)
				b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Item %d weight %d > capacity %d: item cannot fit, exclude.", i, w, c), compState, map[string]any{"i": i, "c": c, "w": w, "v": v})

				dp[i][c] = dp[i-1][c]
				b.IncOperations()

				setState := makeTableState(n, capacity, dp, i, c, "computed", sources, false)
				b.AddStep("L11", trace.ActionSet, fmt.Sprintf("dp[%d][%d] = dp[%d][%d] = %d (item excluded).", i, c, i-1, c, dp[i][c]), setState, map[string]any{"i": i, "c": c, "dp": dp[i][c]})
			}
		}
	}

	maxVal := 0
	if n > 0 && capacity >= 0 {
		maxVal = dp[n][capacity]
	}
	doneState := makeTableState(n, capacity, dp, -1, -1, "normal", nil, true)
	b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Knapsack solved: optimal value for capacity %d is %d.", capacity, maxVal), doneState, map[string]any{"maxValue": maxVal, "capacity": capacity})

	return b.Steps()
}
