package fibonacci

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// cloneTableState returns an independent deep copy of TableState.
func cloneTableState(s renderers.TableState) renderers.TableState {
	var rowHeaders []string
	if s.RowHeaders != nil {
		rowHeaders = make([]string, len(s.RowHeaders))
		copy(rowHeaders, s.RowHeaders)
	}
	var colHeaders []string
	if s.ColHeaders != nil {
		colHeaders = make([]string, len(s.ColHeaders))
		copy(colHeaders, s.ColHeaders)
	}
	var cells []renderers.TableCell
	if s.Cells != nil {
		cells = make([]renderers.TableCell, len(s.Cells))
		copy(cells, s.Cells)
	}
	var deps []renderers.TableDependency
	if s.Dependencies != nil {
		deps = make([]renderers.TableDependency, len(s.Dependencies))
		copy(deps, s.Dependencies)
	}
	return renderers.TableState{
		Caption:      s.Caption,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
		Dependencies: deps,
	}
}

func buildTableState(
	n int,
	values map[int]int,
	cellStates map[int]string,
	deps []renderers.TableDependency,
	caption string,
) renderers.TableState {
	numCols := n + 1
	if numCols < 1 {
		numCols = 1
	}

	colHeaders := make([]string, numCols)
	for c := 0; c < numCols; c++ {
		colHeaders[c] = strconv.Itoa(c)
	}

	cells := make([]renderers.TableCell, numCols)
	for c := 0; c < numCols; c++ {
		valStr := ""
		if v, ok := values[c]; ok {
			valStr = strconv.Itoa(v)
		}
		st := "normal"
		if s, ok := cellStates[c]; ok {
			st = s
		}
		cells[c] = renderers.TableCell{
			Row:   0,
			Col:   c,
			Value: valStr,
			State: st,
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
		RowHeaders:   []string{"F(i)"},
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

// Trace records the execution steps of bottom-up Fibonacci DP tabulation.
func Trace(n int) []trace.Step {
	b := trace.NewBuilder()

	if n <= 0 {
		state0 := cloneTableState(buildTableState(0, map[int]int{0: 0}, map[int]string{0: "computed"}, nil, "Base case: F(0) = 0"))
		b.AddStep("L2", trace.ActionInit, "n <= 0: check base condition.", state0, map[string]any{"n": n})

		stateDone := cloneTableState(buildTableState(0, map[int]int{0: 0}, map[int]string{0: "active"}, nil, "Fibonacci computation complete: F(0) = 0"))
		b.AddStep("L3", trace.ActionDone, "Fibonacci computation complete: F(0) = 0.", stateDone, map[string]any{"n": n, "result": 0})
		return b.Steps()
	}

	values := make(map[int]int, n+1)
	cellStates := make(map[int]string, n+1)

	// Step 1: L5 - Allocate DP table
	initState := cloneTableState(buildTableState(n, values, cellStates, nil, fmt.Sprintf("Initialize DP table of size %d", n+1)))
	b.AddStep("L5", trace.ActionInit, fmt.Sprintf("Initialize DP table of size %d.", n+1), initState, map[string]any{"n": n})

	// Step 2: L6 - Base case F(0) = 0
	values[0] = 0
	cellStates[0] = "computed"
	b.IncOperations()
	base0State := cloneTableState(buildTableState(n, values, cellStates, nil, "Base case: F(0) = 0"))
	b.AddStep("L6", trace.ActionSet, "Set base case: F(0) = 0.", base0State, map[string]any{"n": n, "dp[0]": 0})

	// Step 3: L7 - Base case F(1) = 1
	values[1] = 1
	cellStates[1] = "computed"
	b.IncOperations()
	base1State := cloneTableState(buildTableState(n, values, cellStates, nil, "Base case: F(1) = 1"))
	b.AddStep("L7", trace.ActionSet, "Set base case: F(1) = 1.", base1State, map[string]any{"n": n, "dp[1]": 1})

	dp := make([]int, n+1)
	dp[0] = 0
	dp[1] = 1

	for i := 2; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]

		// Step A: Mark cells i-1 and i-2 as source, cell i as active, and add dependency arrows
		iterStates := make(map[int]string, n+1)
		for k, v := range cellStates {
			iterStates[k] = v
		}
		iterStates[i-2] = "source"
		iterStates[i-1] = "source"
		iterStates[i] = "active"

		deps := []renderers.TableDependency{
			{FromRow: 0, FromCol: i - 2, ToRow: 0, ToCol: i},
			{FromRow: 0, FromCol: i - 1, ToRow: 0, ToCol: i},
		}
		calcCaption := fmt.Sprintf("F(%d) = F(%d) + F(%d) = %d + %d = %d", i, i-1, i-2, dp[i-1], dp[i-2], dp[i])

		b.IncOperations()
		hlState := cloneTableState(buildTableState(n, values, iterStates, deps, calcCaption))
		b.AddStep(
			"L8",
			trace.ActionHighlight,
			calcCaption,
			hlState,
			map[string]any{"n": n, "i": i, "dp[i-1]": dp[i-1], "dp[i-2]": dp[i-2]},
		)

		// Step B: Store result in cell i as computed
		values[i] = dp[i]
		cellStates[i-2] = "computed"
		cellStates[i-1] = "computed"
		cellStates[i] = "computed"

		b.IncOperations()
		setState := cloneTableState(buildTableState(n, values, cellStates, nil, calcCaption))
		b.AddStep(
			"L9",
			trace.ActionSet,
			fmt.Sprintf("Set F(%d) = %d.", i, dp[i]),
			setState,
			map[string]any{"n": n, "i": i, "dp[i]": dp[i]},
		)
	}

	finalStates := make(map[int]string, n+1)
	for k, v := range cellStates {
		finalStates[k] = v
	}
	finalStates[n] = "active"

	completionCaption := fmt.Sprintf("Fibonacci computation complete: F(%d) = %d", n, dp[n])
	doneState := cloneTableState(buildTableState(n, values, finalStates, nil, completionCaption))
	b.AddStep(
		"L11",
		trace.ActionDone,
		fmt.Sprintf("Fibonacci computation complete: F(%d) = %d.", n, dp[n]),
		doneState,
		map[string]any{"n": n, "result": dp[n]},
	)

	return b.Steps()
}
