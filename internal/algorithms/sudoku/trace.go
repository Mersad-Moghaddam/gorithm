package sudoku

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

type cellConfig struct {
	state string
	value string
}

// buildTableState returns a freshly allocated TableState representing the 9x9 board.
// Cells are deep-copied on every step to guarantee snapshot immutability.
func buildTableState(board [9][9]int, overrides map[[2]int]cellConfig, caption string) renderers.TableState {
	rowHeaders := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
	colHeaders := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
	cells := make([]renderers.TableCell, 81)
	idx := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			st := "normal"
			val := " "
			if board[r][c] != 0 {
				st = "computed"
				val = strconv.Itoa(board[r][c])
			}
			if overrides != nil {
				if ov, ok := overrides[[2]int{r, c}]; ok {
					if ov.state != "" {
						st = ov.state
					}
					if ov.value != "" {
						val = ov.value
					}
				}
			}
			cells[idx] = renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: val,
				State: st,
			}
			idx++
		}
	}
	return renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// CloneTableState creates an explicit deep copy of a TableState, cloning its Cells slice.
func CloneTableState(ts renderers.TableState) renderers.TableState {
	cloned := ts
	if ts.RowHeaders != nil {
		cloned.RowHeaders = append([]string(nil), ts.RowHeaders...)
	}
	if ts.ColHeaders != nil {
		cloned.ColHeaders = append([]string(nil), ts.ColHeaders...)
	}
	if ts.Cells != nil {
		cloned.Cells = make([]renderers.TableCell, len(ts.Cells))
		copy(cloned.Cells, ts.Cells)
	}
	if ts.Dependencies != nil {
		cloned.Dependencies = make([]renderers.TableDependency, len(ts.Dependencies))
		copy(cloned.Dependencies, ts.Dependencies)
	}
	return cloned
}

// Trace records execution steps of solving a Sudoku board.
func Trace(board [9][9]int) []trace.Step {
	b := trace.NewBuilder()
	initCaption := "Initialize Sudoku grid state"
	b.AddStep("L1", trace.ActionInit, initCaption, buildTableState(board, nil, initCaption), map[string]any{"row": -1, "col": -1})

	if !validateBoard(board) {
		failCaption := "Invalid initial puzzle configuration violates constraints"
		b.AddStep("L15", trace.ActionDone, failCaption, buildTableState(board, nil, failCaption), map[string]any{"row": -1, "col": -1})
		return b.Steps()
	}

	work := board
	solved := traceBacktrack(&work, b)
	if !solved {
		failCaption := "No valid solution exists for this puzzle"
		b.AddStep("L15", trace.ActionDone, failCaption, buildTableState(work, nil, failCaption), map[string]any{"row": -1, "col": -1})
	}
	return b.Steps()
}

func traceBacktrack(board *[9][9]int, b *trace.Builder) bool {
	r, c, found := findEmpty(board)
	if !found {
		solvedCaption := "All cells filled successfully; puzzle solved"
		b.AddStep("L4", trace.ActionDone, solvedCaption, buildTableState(*board, nil, solvedCaption), map[string]any{"row": -1, "col": -1, "solved": true})
		return true
	}

	visitCaption := fmt.Sprintf("Locate empty cell at (%d, %d)", r, c)
	b.AddStep("L2", trace.ActionVisit, visitCaption, buildTableState(*board, map[[2]int]cellConfig{
		{r, c}: {state: "active", value: " "},
	}, visitCaption), map[string]any{"row": r, "col": c})

	for num := 1; num <= 9; num++ {
		vars := map[string]any{
			"row":       r,
			"col":       c,
			"num":       num,
			"candidate": num,
		}
		b.IncComparisons()
		testCaption := fmt.Sprintf("Evaluating candidate %d for cell (%d, %d)", num, r, c)
		b.AddStep("L6", trace.ActionCompare, testCaption, buildTableState(*board, map[[2]int]cellConfig{
			{r, c}: {state: "active", value: strconv.Itoa(num)},
		}, testCaption), vars)

		conflict, hasConflict := FindConflict(*board, r, c, num)
		if hasConflict {
			conflictVars := map[string]any{
				"row":          r,
				"col":          c,
				"num":          num,
				"candidate":    num,
				"conflictRow":  conflict.Row,
				"conflictCol":  conflict.Col,
				"conflictType": conflict.Invariant,
			}
			conflictCaption := fmt.Sprintf("Candidate %d at (%d, %d) violates %s invariant: conflict at (%d, %d)", num, r, c, conflict.Invariant, conflict.Row, conflict.Col)
			b.AddStep("L7", trace.ActionHighlight, conflictCaption, buildTableState(*board, map[[2]int]cellConfig{
				{r, c}:                       {state: "active", value: strconv.Itoa(num)},
				{conflict.Row, conflict.Col}: {state: "source"},
			}, conflictCaption), conflictVars)
			continue
		}

		board[r][c] = num
		b.IncOperations()
		placeCaption := fmt.Sprintf("Candidate %d satisfies constraints; place at (%d, %d)", num, r, c)
		b.AddStep("L8", trace.ActionSet, placeCaption, buildTableState(*board, map[[2]int]cellConfig{
			{r, c}: {state: "computed", value: strconv.Itoa(num)},
		}, placeCaption), vars)

		advanceCaption := fmt.Sprintf("Advance to next empty cell with %d at (%d, %d)", num, r, c)
		b.AddStep("L9", trace.ActionVisit, advanceCaption, buildTableState(*board, map[[2]int]cellConfig{
			{r, c}: {state: "computed", value: strconv.Itoa(num)},
		}, advanceCaption), vars)

		if traceBacktrack(board, b) {
			return true
		}

		board[r][c] = 0
		b.IncOperations()
		backtrackCaption := fmt.Sprintf("Dead end: no valid digit for cell (%d, %d); backtrack", r, c)
		b.AddStep("L12", trace.ActionPop, backtrackCaption, buildTableState(*board, map[[2]int]cellConfig{
			{r, c}: {state: "normal", value: " "},
		}, backtrackCaption), vars)
	}

	deadEndCaption := fmt.Sprintf("Dead end: no valid digit for cell (%d, %d); backtrack", r, c)
	b.AddStep("L15", trace.ActionPop, deadEndCaption, buildTableState(*board, map[[2]int]cellConfig{
		{r, c}: {state: "normal", value: " "},
	}, deadEndCaption), map[string]any{"row": r, "col": c})
	return false
}
