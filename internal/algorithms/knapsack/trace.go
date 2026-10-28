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
	known := make([][]bool, n+1)
	for i := range known {
		known[i] = make([]bool, capacity+1)
		for c := range known[i] {
			known[i][c] = i == 0
		}
	}
	snapshot := func(item, column int, phase string) renderers.KnapsackLessonState {
		s := renderers.KnapsackLessonState{Weights: append([]int(nil), weights[:n]...), Values: append([]int(nil), values[:n]...), Capacity: capacity, Item: item, Column: column, Phase: phase, Sources: [2][2]int{{-1, -1}, {-1, -1}}}
		s.TableState = makeTableState(n, capacity, dp, item, column, "normal", nil, phase == "done")
		s.DP = make([][]int, n+1)
		s.Known = make([][]bool, n+1)
		for r := range dp {
			s.DP[r] = append([]int(nil), dp[r]...)
			s.Known[r] = append([]bool(nil), known[r]...)
			for c := range dp[r] {
				cell := &s.Cells[r*(capacity+1)+c]
				cell.Value = ""
				if known[r][c] {
					cell.Value = strconv.Itoa(dp[r][c])
				}
			}
		}
		if item > 0 && column >= 0 {
			s.Sources[0] = [2]int{item - 1, column}
			s.Exclude = dp[item-1][column]
			s.Fits = weights[item-1] <= column
			if s.Fits {
				s.Sources[1] = [2]int{item - 1, column - weights[item-1]}
				s.Include = dp[item-1][column-weights[item-1]] + values[item-1]
			}
			s.Chosen = dp[item][column]
		}
		if phase == "done" {
			s.Result = dp[n][capacity]
			s.HasResult = true
		}
		return s
	}

	initState := snapshot(-1, -1, "init")
	b.AddStep("L1", trace.ActionInit, "Initialize 0/1 Knapsack dynamic programming solver.", initState, map[string]any{"n": n, "capacity": capacity})

	allocState := snapshot(-1, -1, "allocate")
	b.AddStep("L3", trace.ActionSet, fmt.Sprintf("Allocate DP matrix of size %dx%d with 0-item base cases.", n+1, capacity+1), allocState, map[string]any{"n": n, "capacity": capacity})

	for i := 1; i <= n; i++ {
		w := weights[i-1]
		v := values[i-1]

		itemState := snapshot(i, -1, "item")
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
				compState := snapshot(i, c, "compare")
				b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Item %d weight %d <= capacity %d: compare exclude (%d) vs include (%d + %d = %d).", i, w, c, dp[i-1][c], dp[i-1][c-w], v, dp[i-1][c-w]+v), compState, map[string]any{"i": i, "c": c, "w": w, "v": v, "exclude": dp[i-1][c], "include": dp[i-1][c-w] + v})

				dp[i][c] = max(dp[i-1][c], dp[i-1][c-w]+v)
				known[i][c] = true
				b.IncOperations()

				setState := snapshot(i, c, "write")
				b.AddStep("L9", trace.ActionSet, fmt.Sprintf("dp[%d][%d] = max(%d, %d) = %d.", i, c, dp[i-1][c], dp[i-1][c-w]+v, dp[i][c]), setState, map[string]any{"i": i, "c": c, "dp": dp[i][c]})
			} else {
				compState := snapshot(i, c, "compare")
				b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Item %d weight %d > capacity %d: item cannot fit, exclude.", i, w, c), compState, map[string]any{"i": i, "c": c, "w": w, "v": v})

				dp[i][c] = dp[i-1][c]
				known[i][c] = true
				b.IncOperations()

				setState := snapshot(i, c, "write")
				b.AddStep("L11", trace.ActionSet, fmt.Sprintf("dp[%d][%d] = dp[%d][%d] = %d (item excluded).", i, c, i-1, c, dp[i][c]), setState, map[string]any{"i": i, "c": c, "dp": dp[i][c]})
			}
		}
	}

	maxVal := 0
	if n > 0 && capacity >= 0 {
		maxVal = dp[n][capacity]
	}
	doneState := snapshot(-1, -1, "done")
	b.AddStep("L15", trace.ActionDone, fmt.Sprintf("Knapsack solved: optimal value for capacity %d is %d.", capacity, maxVal), doneState, map[string]any{"maxValue": maxVal, "capacity": capacity})

	steps := b.Steps()
	for i := range steps {
		if steps[i].LineID == "L8" {
			steps[i].HoldMS = 600
		}
		if steps[i].LineID == "L9" || steps[i].LineID == "L11" {
			steps[i].HoldMS = 1200
		}
	}
	return steps
}
