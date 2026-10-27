package coinchange

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// cloneTableState returns an independent deep copy of TableState so step snapshots never share backing arrays.
// ponytail: explicit slice copies for immutability; deltas or persistent structure if table is huge.
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
		Dependencies: deps,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

// makeTableState builds a renderers.TableState snapshot for Coin Change DP visualization.
func makeTableState(
	amount int,
	dp []int,
	activeCol int,
	activeState string,
	sourceCols []int,
	initialized bool,
	caption string,
	dependencies []renderers.TableDependency,
) renderers.TableState {
	colHeaders := make([]string, amount+1)
	for c := 0; c <= amount; c++ {
		colHeaders[c] = strconv.Itoa(c)
	}

	sourceMap := make(map[int]bool, len(sourceCols))
	for _, sc := range sourceCols {
		sourceMap[sc] = true
	}

	cells := make([]renderers.TableCell, amount+1)
	for c := 0; c <= amount; c++ {
		val := ""
		if initialized {
			if dp[c] > amount {
				val = "∞"
			} else {
				val = strconv.Itoa(dp[c])
			}
		} else if c == 0 && len(dp) > 0 && dp[0] == 0 {
			val = "0"
		}

		st := "normal"
		if c == activeCol {
			st = activeState
		} else if sourceMap[c] {
			st = "source"
		}

		cells[c] = renderers.TableCell{
			Row:   0,
			Col:   c,
			Value: val,
			State: st,
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
		RowHeaders:   []string{"dp[amount]"},
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

// Trace records the step-by-step DP execution for Coin Change.
func Trace(coins []int, amount int) []trace.Step {
	b := trace.NewBuilder()

	if amount < 0 {
		emptyState := makeTableState(0, []int{0}, -1, "normal", nil, false, "Invalid negative amount", nil)
		b.AddStep(
			"L2",
			trace.ActionDone,
			fmt.Sprintf("Invalid amount %d (< 0). Return -1.", amount),
			cloneTableState(emptyState),
			map[string]any{"amount": amount, "minCoins": -1, "result": -1},
		)
		return b.Steps()
	}

	dp := make([]int, amount+1)

	// Step 1: L1 - Initialization
	initCaption := fmt.Sprintf("Initialize dp table for amount %d with coins %v", amount, coins)
	initState := makeTableState(amount, dp, -1, "normal", nil, false, initCaption, nil)
	b.AddStep(
		"L1",
		trace.ActionInit,
		fmt.Sprintf("Initialize Coin Change solver for amount %d with coins %v.", amount, coins),
		cloneTableState(initState),
		map[string]any{"amount": amount, "coins": coins},
	)

	// Step 2: L3 - Base case dp[0] = 0
	b.IncOperations()
	baseCaption := "Base case: dp[0] = 0 coins required to form amount 0"
	baseState := makeTableState(amount, dp, 0, "computed", nil, false, baseCaption, nil)
	b.AddStep(
		"L3",
		trace.ActionSet,
		"Base case: dp[0] = 0 coins required to form amount 0.",
		cloneTableState(baseState),
		map[string]any{"amount": 0, "subproblem": 0, "dp[0]": 0},
	)

	// Step 3: L4 - Set dp[1..amount] to infinity
	if amount > 0 {
		for i := 1; i <= amount; i++ {
			dp[i] = amount + 1
		}
		b.IncOperations()
		infCaption := fmt.Sprintf("Initialize dp[1..%d] with ∞ (unreachable sentinel)", amount)
		infState := makeTableState(amount, dp, -1, "normal", nil, true, infCaption, nil)
		b.AddStep(
			"L4",
			trace.ActionSet,
			fmt.Sprintf("Initialize dp[1..%d] with ∞ (unreachable sentinel).", amount),
			cloneTableState(infState),
			map[string]any{"amount": amount},
		)
	}

	// Outer loop: iterate over subproblem amounts 1..amount
	for i := 1; i <= amount; i++ {
		for _, c := range coins {
			b.IncComparisons()
			if c <= i {
				subproblem := i - c
				sources := []int{subproblem}
				deps := []renderers.TableDependency{{
					FromRow: 0,
					FromCol: subproblem,
					ToRow:   0,
					ToCol:   i,
				}}

				var caption string
				if dp[subproblem] <= amount {
					caption = fmt.Sprintf("Testing coin %d for amount %d: dp[%d - %d] + 1 = %d + 1 = %d coins", c, i, i, c, dp[subproblem], dp[subproblem]+1)
				} else {
					caption = fmt.Sprintf("Testing coin %d for amount %d: dp[%d - %d] = ∞ (cannot form %d)", c, i, i, c, subproblem)
				}

				compState := makeTableState(amount, dp, i, "active", sources, true, caption, deps)
				b.AddStep(
					"L7",
					trace.ActionCompare,
					caption,
					cloneTableState(compState),
					map[string]any{
						"i":          i,
						"coin":       c,
						"subproblem": subproblem,
						"canTake":    true,
					},
				)

				if dp[subproblem]+1 < dp[i] {
					dp[i] = dp[subproblem] + 1
					b.IncOperations()

					setCaption := fmt.Sprintf("Update dp[%d] = dp[%d - %d] + 1 = %d using coin %d", i, i, c, dp[i], c)
					setState := makeTableState(amount, dp, i, "computed", sources, true, setCaption, deps)
					b.AddStep(
						"L8",
						trace.ActionSet,
						fmt.Sprintf("Update dp[%d] = dp[%d - %d] + 1 = %d using coin %d.", i, i, c, dp[i], c),
						cloneTableState(setState),
						map[string]any{
							"i":          i,
							"coin":       c,
							"subproblem": subproblem,
							"newMin":     dp[i],
						},
					)
				}
			} else {
				caption := fmt.Sprintf("Testing coin %d for amount %d: coin exceeds amount (%d > %d), skip", c, i, c, i)
				compState := makeTableState(amount, dp, i, "active", nil, true, caption, nil)
				b.AddStep(
					"L7",
					trace.ActionCompare,
					caption,
					cloneTableState(compState),
					map[string]any{
						"i":       i,
						"coin":    c,
						"canTake": false,
					},
				)
			}
		}
	}

	// Final step
	if dp[amount] > amount {
		finalCaption := fmt.Sprintf("Amount %d cannot be formed with coins %v. Return -1.", amount, coins)
		doneState := makeTableState(amount, dp, amount, "normal", nil, true, finalCaption, nil)
		b.AddStep(
			"L12",
			trace.ActionDone,
			finalCaption,
			cloneTableState(doneState),
			map[string]any{"amount": amount, "minCoins": -1, "result": -1},
		)
	} else {
		finalCaption := fmt.Sprintf("Completed! Minimum coins needed for amount %d is %d.", amount, dp[amount])
		doneState := makeTableState(amount, dp, amount, "computed", nil, true, finalCaption, nil)
		b.AddStep(
			"L13",
			trace.ActionDone,
			finalCaption,
			cloneTableState(doneState),
			map[string]any{"amount": amount, "minCoins": dp[amount], "result": dp[amount]},
		)
	}

	return b.Steps()
}
