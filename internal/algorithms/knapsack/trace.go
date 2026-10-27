package knapsack

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type cellCoord struct {
	row int
	col int
}

// ponytail: table state builds full matrix for rendering; sparse representation if W grows large.
func makeTableState(
	rowHeaders []string,
	colHeaders []string,
	dp [][]int,
	activeRow, activeCol int,
	activeState string,
	sources []cellCoord,
	isDone bool,
	caption string,
	deps []renderers.TableDependency,
) renderers.TableState {
	numRows := len(rowHeaders)
	numCols := len(colHeaders)

	clonedRowHeaders := make([]string, numRows)
	copy(clonedRowHeaders, rowHeaders)

	clonedColHeaders := make([]string, numCols)
	copy(clonedColHeaders, colHeaders)

	sourceMap := make(map[cellCoord]bool, len(sources))
	for _, s := range sources {
		sourceMap[s] = true
	}

	cells := make([]renderers.TableCell, 0, numRows*numCols)
	for r := 0; r < numRows; r++ {
		for c := 0; c < numCols; c++ {
			coord := cellCoord{row: r, col: c}
			val := ""
			state := "normal"

			isComputed := isDone || r == 0 || r < activeRow || (r == activeRow && c < activeCol) || (r == activeRow && c == activeCol && activeState == "computed")

			if isComputed && r < len(dp) && c < len(dp[r]) {
				val = strconv.Itoa(dp[r][c])
			}

			if isDone {
				if r == numRows-1 && c == numCols-1 {
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

	var clonedDeps []renderers.TableDependency
	if len(deps) > 0 {
		clonedDeps = make([]renderers.TableDependency, len(deps))
		copy(clonedDeps, deps)
	}

	return renderers.TableState{
		Caption:      caption,
		Dependencies: clonedDeps,
		RowHeaders:   clonedRowHeaders,
		ColHeaders:   clonedColHeaders,
		Cells:        cells,
	}
}

func findSelectedItems(weights, values []int, capacity int, dp [][]int) []int {
	n := min(len(weights), len(values))
	c := capacity
	var items []int
	for i := n; i >= 1; i-- {
		w := weights[i-1]
		v := values[i-1]
		if c >= w && dp[i][c] == dp[i-1][c-w]+v && dp[i][c] != dp[i-1][c] {
			items = append(items, i)
			c -= w
		}
	}
	for l, r := 0, len(items)-1; l < r; l, r = l+1, r-1 {
		items[l], items[r] = items[r], items[l]
	}
	return items
}

// Trace records the step-by-step execution of 0/1 Knapsack dynamic programming.
func Trace(weights, values []int, capacity int) []trace.Step {
	b := trace.NewBuilder()
	n := min(len(weights), len(values))
	if capacity < 0 {
		capacity = 0
	}

	rowHeaders := make([]string, n+1)
	rowHeaders[0] = "Item 0 (w:0, v:0)"
	for i := 1; i <= n; i++ {
		rowHeaders[i] = fmt.Sprintf("Item %d (w:%d, v:%d)", i, weights[i-1], values[i-1])
	}

	colHeaders := make([]string, capacity+1)
	for c := 0; c <= capacity; c++ {
		colHeaders[c] = strconv.Itoa(c)
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, capacity+1)
	}

	initState := makeTableState(rowHeaders, colHeaders, dp, -1, -1, "normal", nil, false, "Initialize 0/1 Knapsack DP solver", nil)
	b.AddStep("L1", trace.ActionInit, "Initialize 0/1 Knapsack dynamic programming solver.", initState, map[string]any{"n": n, "capacity": capacity})

	allocCaption := fmt.Sprintf("DP matrix allocated (%dx%d)", n+1, capacity+1)
	allocState := makeTableState(rowHeaders, colHeaders, dp, -1, -1, "normal", nil, false, allocCaption, nil)
	b.AddStep("L3", trace.ActionSet, fmt.Sprintf("Allocate DP matrix of size %dx%d with 0-item base cases.", n+1, capacity+1), allocState, map[string]any{"n": n, "capacity": capacity})

	for i := 1; i <= n; i++ {
		w := weights[i-1]
		v := values[i-1]

		itemCaption := fmt.Sprintf("Evaluating item %d (weight=%d, value=%d)", i, w, v)
		itemState := makeTableState(rowHeaders, colHeaders, dp, -1, -1, "normal", nil, false, itemCaption, nil)
		b.AddStep("L5", trace.ActionHighlight, fmt.Sprintf("Evaluating item %d: weight = %d, value = %d.", i, w, v), itemState, map[string]any{"i": i, "w": w, "v": v})

		for c := 0; c <= capacity; c++ {
			excludeCell := cellCoord{row: i - 1, col: c}
			sources := []cellCoord{excludeCell}
			deps := []renderers.TableDependency{
				{FromRow: i - 1, FromCol: c, ToRow: i, ToCol: c},
			}

			b.IncComparisons()
			if w <= c {
				includeCell := cellCoord{row: i - 1, col: c - w}
				if includeCell != excludeCell {
					sources = append(sources, includeCell)
					deps = append(deps, renderers.TableDependency{
						FromRow: i - 1,
						FromCol: c - w,
						ToRow:   i,
						ToCol:   c,
					})
				}

				compCaption := fmt.Sprintf("Item %d: compare exclude (%d) vs include (%d+%d=%d)", i, dp[i-1][c], dp[i-1][c-w], v, dp[i-1][c-w]+v)
				compState := makeTableState(rowHeaders, colHeaders, dp, i, c, "active", sources, false, compCaption, deps)
				b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Item %d weight %d <= capacity %d: compare exclude (%d) vs include (%d + %d = %d).", i, w, c, dp[i-1][c], dp[i-1][c-w], v, dp[i-1][c-w]+v), compState, map[string]any{"i": i, "c": c, "w": w, "v": v, "exclude": dp[i-1][c], "include": dp[i-1][c-w] + v})

				dp[i][c] = max(dp[i-1][c], dp[i-1][c-w]+v)
				b.IncOperations()

				decision := "exclude"
				if dp[i-1][c-w]+v > dp[i-1][c] {
					decision = "include"
				}
				setCaption := fmt.Sprintf("dp[%d][%d] = max(%d, %d) = %d (%s)", i, c, dp[i-1][c], dp[i-1][c-w]+v, dp[i][c], decision)
				setState := makeTableState(rowHeaders, colHeaders, dp, i, c, "computed", sources, false, setCaption, deps)
				b.AddStep("L9", trace.ActionSet, fmt.Sprintf("dp[%d][%d] = max(exclude=%d, include=%d) = %d (%s item %d).", i, c, dp[i-1][c], dp[i-1][c-w]+v, dp[i][c], decision, i), setState, map[string]any{"i": i, "c": c, "dp": dp[i][c]})
			} else {
				compCaption := fmt.Sprintf("Item %d weight %d > capacity %d: exclude (%d)", i, w, c, dp[i-1][c])
				compState := makeTableState(rowHeaders, colHeaders, dp, i, c, "active", sources, false, compCaption, deps)
				b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Item %d weight %d > capacity %d: item cannot fit, exclude choice takes dp[%d][%d] = %d.", i, w, c, i-1, c, dp[i-1][c]), compState, map[string]any{"i": i, "c": c, "w": w, "v": v, "exclude": dp[i-1][c]})

				dp[i][c] = dp[i-1][c]
				b.IncOperations()

				setCaption := fmt.Sprintf("dp[%d][%d] = dp[%d][%d] = %d (excluded)", i, c, i-1, c, dp[i][c])
				setState := makeTableState(rowHeaders, colHeaders, dp, i, c, "computed", sources, false, setCaption, deps)
				b.AddStep("L11", trace.ActionSet, fmt.Sprintf("dp[%d][%d] = dp[%d][%d] = %d (item %d excluded).", i, c, i-1, c, dp[i][c], i), setState, map[string]any{"i": i, "c": c, "dp": dp[i][c]})
			}
		}
	}

	maxVal := 0
	if n > 0 && capacity >= 0 {
		maxVal = dp[n][capacity]
	}

	selected := findSelectedItems(weights, values, capacity, dp)
	var selectedStr string
	if len(selected) == 0 {
		selectedStr = "none"
	} else {
		var parts []string
		for _, item := range selected {
			parts = append(parts, fmt.Sprintf("Item %d", item))
		}
		selectedStr = strings.Join(parts, ", ")
	}

	doneCaption := fmt.Sprintf("Optimal value: %d | Selected items: %s", maxVal, selectedStr)
	doneState := makeTableState(rowHeaders, colHeaders, dp, -1, -1, "normal", nil, true, doneCaption, nil)
	b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Knapsack solved: optimal value for capacity %d is %d (selected: %s).", capacity, maxVal, selectedStr), doneState, map[string]any{"maxValue": maxVal, "capacity": capacity, "selectedItems": selected})

	return b.Steps()
}
