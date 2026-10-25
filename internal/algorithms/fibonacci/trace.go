package fibonacci

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func buildTableState(n int, values map[int]int, cellStates map[int]string) renderers.TableState {
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

	return renderers.TableState{
		RowHeaders: []string{"dp"},
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace records the execution steps of bottom-up Fibonacci DP tabulation.
func Trace(n int) []trace.Step {
	b := trace.NewBuilder()

	if n <= 0 {
		state0 := buildTableState(0, map[int]int{0: 0}, map[int]string{0: "normal"})
		b.AddStep("L2", trace.ActionInit, "n <= 0: check base condition.", state0, map[string]any{"n": n})
		stateDone := buildTableState(0, map[int]int{0: 0}, map[int]string{0: "computed"})
		b.AddStep("L3", trace.ActionDone, "Return 0 for n <= 0.", stateDone, map[string]any{"n": n, "result": 0})
		return b.Steps()
	}

	values := make(map[int]int, n+1)
	cellStates := make(map[int]string, n+1)

	// Step 1: L5 - Allocate DP table
	initState := buildTableState(n, values, cellStates)
	b.AddStep("L5", trace.ActionInit, fmt.Sprintf("Initialize DP table of size %d.", n+1), initState, map[string]any{"n": n})

	// Step 2: L6 - Base case dp[0] = 0
	values[0] = 0
	cellStates[0] = "computed"
	b.IncOperations()
	base0State := buildTableState(n, values, cellStates)
	b.AddStep("L6", trace.ActionSet, "Set base case: dp[0] = 0.", base0State, map[string]any{"n": n, "dp[0]": 0})

	// Step 3: L7 - Base case dp[1] = 1
	values[1] = 1
	cellStates[1] = "computed"
	b.IncOperations()
	base1State := buildTableState(n, values, cellStates)
	b.AddStep("L7", trace.ActionSet, "Set base case: dp[1] = 1.", base1State, map[string]any{"n": n, "dp[1]": 1})

	dp := make([]int, n+1)
	dp[0] = 0
	dp[1] = 1

	for i := 2; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]

		// Step A: Highlight sources fib(i-1) and fib(i-2) when computing fib(i)
		iterStates := make(map[int]string, n+1)
		for k, v := range cellStates {
			iterStates[k] = v
		}
		iterStates[i-2] = "source"
		iterStates[i-1] = "source"
		iterStates[i] = "active"

		b.IncOperations()
		hlState := buildTableState(n, values, iterStates)
		b.AddStep(
			"L8",
			trace.ActionHighlight,
			fmt.Sprintf("Compute dp[%d] from source cells dp[%d]=%d and dp[%d]=%d.", i, i-1, dp[i-1], i-2, dp[i-2]),
			hlState,
			map[string]any{"n": n, "i": i, "dp[i-1]": dp[i-1], "dp[i-2]": dp[i-2]},
		)

		// Step B: Set computed dp[i]
		values[i] = dp[i]
		cellStates[i-2] = "computed"
		cellStates[i-1] = "computed"
		cellStates[i] = "computed"

		b.IncOperations()
		setState := buildTableState(n, values, cellStates)
		b.AddStep(
			"L9",
			trace.ActionSet,
			fmt.Sprintf("Set dp[%d] = dp[%d] + dp[%d] = %d + %d = %d.", i, i-1, i-2, dp[i-1], dp[i-2], dp[i]),
			setState,
			map[string]any{"n": n, "i": i, "dp[i]": dp[i]},
		)
	}

	doneState := buildTableState(n, values, cellStates)
	b.AddStep(
		"L11",
		trace.ActionDone,
		fmt.Sprintf("Fibonacci computation complete: fib(%d) = %d.", n, dp[n]),
		doneState,
		map[string]any{"n": n, "result": dp[n]},
	)

	return b.Steps()
}
