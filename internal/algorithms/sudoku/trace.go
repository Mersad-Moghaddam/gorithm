package sudoku

import (
	"fmt"
	"strconv"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func buildTableState(board [9][9]int, activeRow, activeCol int, cellState string) renderers.TableState {
	rowHeaders := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
	colHeaders := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8"}
	cells := make([]renderers.TableCell, 0, 81)
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			st := "normal"
			if r == activeRow && c == activeCol {
				st = cellState
			} else if board[r][c] != 0 {
				st = "computed"
			}
			val := ""
			if board[r][c] != 0 {
				val = strconv.Itoa(board[r][c])
			}
			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: val,
				State: st,
			})
		}
	}
	return renderers.TableState{
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace records execution steps of solving a Sudoku board.
func Trace(board [9][9]int) []trace.Step {
	b := trace.NewBuilder()
	b.AddStep("L1", trace.ActionInit, "Initialize Sudoku grid state", buildTableState(board, -1, -1, "normal"), map[string]any{"row": -1, "col": -1})

	if !validateBoard(board) {
		b.AddStep("L15", trace.ActionDone, "Invalid initial puzzle configuration violates constraints", buildTableState(board, -1, -1, "normal"), map[string]any{"row": -1, "col": -1})
		return b.Steps()
	}

	work := board
	solved := traceBacktrack(&work, b)
	if !solved {
		b.AddStep("L15", trace.ActionDone, "No valid solution exists for this puzzle", buildTableState(work, -1, -1, "normal"), map[string]any{"row": -1, "col": -1})
	}
	return b.Steps()
}

func traceBacktrack(board *[9][9]int, b *trace.Builder) bool {
	r, c, found := findEmpty(board)
	if !found {
		b.AddStep("L4", trace.ActionDone, "All cells filled successfully; puzzle solved", buildTableState(*board, -1, -1, "normal"), map[string]any{"row": -1, "col": -1})
		return true
	}

	b.AddStep("L2", trace.ActionVisit, fmt.Sprintf("Locate empty cell at (%d, %d)", r, c), buildTableState(*board, r, c, "active"), map[string]any{"row": r, "col": c})

	for num := 1; num <= 9; num++ {
		vars := map[string]any{
			"row":       r,
			"col":       c,
			"candidate": num,
			"num":       num,
		}
		b.IncComparisons()
		b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Evaluating candidate %d for cell (%d, %d)", num, r, c), buildTableState(*board, r, c, "active"), vars)

		if IsValid(*board, r, c, num) {
			board[r][c] = num
			b.IncOperations()
			b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Candidate %d satisfies constraints; place at (%d, %d)", num, r, c), buildTableState(*board, r, c, "computed"), vars)

			b.AddStep("L9", trace.ActionVisit, fmt.Sprintf("Advance to next empty cell with %d at (%d, %d)", num, r, c), buildTableState(*board, r, c, "computed"), vars)

			if traceBacktrack(board, b) {
				return true
			}

			board[r][c] = 0
			b.IncOperations()
			b.AddStep("L12", trace.ActionPop, fmt.Sprintf("Backtrack: reset cell (%d, %d) from %d to 0", r, c, num), buildTableState(*board, r, c, "source"), vars)
		} else {
			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Candidate %d violates row, column, or box constraints at (%d, %d)", num, r, c), buildTableState(*board, r, c, "source"), vars)
		}
	}

	b.AddStep("L15", trace.ActionPop, fmt.Sprintf("Exhausted candidates 1-9 at (%d, %d); trigger backtrack", r, c), buildTableState(*board, r, c, "source"), map[string]any{"row": r, "col": c})
	return false
}
