package coinchange

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// ponytail: table state builds 1D array for rendering; could stream deltas if amount is huge.
func makeTableState(
	amount int,
	dp []int,
	activeCol int,
	activeState string,
	sourceCols []int,
	initialized bool,
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

	return renderers.TableState{
		RowHeaders: []string{"dp"},
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace records the step-by-step DP execution for Coin Change.
func Trace(coins []int, amount int) []trace.Step {
	b := trace.NewBuilder()

	if amount < 0 {
		emptyState := makeTableState(0, []int{0}, -1, "normal", nil, false)
		b.AddStep(
			"L2",
			trace.ActionDone,
			fmt.Sprintf("Invalid amount %d (< 0). Return -1.", amount),
			emptyState,
			map[string]any{"amount": amount, "minCoins": -1, "result": -1},
		)
		return b.Steps()
	}

	dp := make([]int, amount+1)

	// Step 1: L1 - Initialization
	initState := makeTableState(amount, dp, -1, "normal", nil, false)
	b.AddStep(
		"L1",
		trace.ActionInit,
		fmt.Sprintf("Initialize Coin Change solver for amount %d with coins %v.", amount, coins),
		initState,
		map[string]any{"amount": amount, "coins": coins},
	)

	// Step 2: L3 - Base case dp[0] = 0
	b.IncOperations()
	baseState := makeTableState(amount, dp, 0, "computed", nil, false)
	b.AddStep(
		"L3",
		trace.ActionSet,
		"Base case: dp[0] = 0 coins required to form amount 0.",
		baseState,
		map[string]any{"amount": 0, "subproblem": 0, "dp[0]": 0},
	)

	// Step 3: L4 - Set dp[1..amount] to infinity
	for i := 1; i <= amount; i++ {
		dp[i] = amount + 1
	}
	b.IncOperations()
	infState := makeTableState(amount, dp, -1, "normal", nil, true)
	b.AddStep(
		"L4",
		trace.ActionSet,
		fmt.Sprintf("Initialize dp[1..%d] with ∞ (unreachable sentinel).", amount),
		infState,
		map[string]any{"amount": amount},
	)

	// Outer loop: iterate over subproblem amounts 1..amount
	for i := 1; i <= amount; i++ {
		for _, c := range coins {
			b.IncComparisons()
			if c <= i {
				subproblem := i - c
				sources := []int{subproblem}

				subVal := "∞"
				if dp[subproblem] <= amount {
					subVal = strconv.Itoa(dp[subproblem])
				}
				curVal := "∞"
				if dp[i] <= amount {
					curVal = strconv.Itoa(dp[i])
				}

				compState := makeTableState(amount, dp, i, "active", sources, true)
				b.AddStep(
					"L7",
					trace.ActionCompare,
					fmt.Sprintf("Amount %d: test coin %d. Subproblem dp[%d] = %s (candidate: %s + 1), current dp[%d] = %s.", i, c, subproblem, subVal, subVal, i, curVal),
					compState,
					map[string]any{
						"i":          i,
						"coin":       c,
						"subproblem": subproblem,
						"subVal":     subVal,
						"curVal":     curVal,
						"canTake":    true,
					},
				)

				if dp[subproblem]+1 < dp[i] {
					dp[i] = dp[subproblem] + 1
					b.IncOperations()

					setState := makeTableState(amount, dp, i, "computed", sources, true)
					b.AddStep(
						"L8",
						trace.ActionSet,
						fmt.Sprintf("Update dp[%d] = dp[%d] + 1 = %d using coin %d.", i, subproblem, dp[i], c),
						setState,
						map[string]any{
							"i":          i,
							"coin":       c,
							"subproblem": subproblem,
							"newMin":     dp[i],
						},
					)
				}
			} else {
				compState := makeTableState(amount, dp, i, "active", nil, true)
				b.AddStep(
					"L7",
					trace.ActionCompare,
					fmt.Sprintf("Amount %d: coin %d > %d, cannot use.", i, c, i),
					compState,
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
		doneState := makeTableState(amount, dp, amount, "normal", nil, true)
		b.AddStep(
			"L12",
			trace.ActionDone,
			fmt.Sprintf("Amount %d cannot be made with coins %v. Return -1.", amount, coins),
			doneState,
			map[string]any{"amount": amount, "minCoins": -1, "result": -1},
		)
	} else {
		doneState := makeTableState(amount, dp, amount, "computed", nil, true)
		b.AddStep(
			"L13",
			trace.ActionDone,
			fmt.Sprintf("Completed! Minimum coins needed for amount %d is %d.", amount, dp[amount]),
			doneState,
			map[string]any{"amount": amount, "minCoins": dp[amount], "result": dp[amount]},
		)
	}

	return b.Steps()
}
