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
	var given [9][9]bool
	for r := range board {
		for c := range board[r] {
			given[r][c] = board[r][c] != 0
		}
	}
	b.AddStep("L1", trace.ActionInit, "Initialize Sudoku grid state", sudokuSnapshot(board, given, -1, -1, 0, 0, "init", "searching"), map[string]any{"row": -1, "col": -1})

	if !validateBoard(board) {
		b.AddStep("L15", trace.ActionDone, "Invalid initial puzzle configuration violates constraints", sudokuSnapshot(board, given, -1, -1, 0, 0, "done", "invalid"), map[string]any{"row": -1, "col": -1})
		return b.Steps()
	}

	work := board
	solved := traceBacktrack(&work, given, 0, b)
	if !solved {
		b.AddStep("L15", trace.ActionDone, "No valid solution exists for this puzzle", sudokuSnapshot(work, given, -1, -1, 0, 0, "done", "unsatisfiable"), map[string]any{"row": -1, "col": -1})
	}
	steps := b.Steps()
	for i := range steps {
		if i+1 < len(steps) && steps[i].LineID == "L6" && steps[i+1].LineID == "L8" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}

func traceBacktrack(board *[9][9]int, given [9][9]bool, depth int, b *trace.Builder) bool {
	r, c, found := findEmpty(board)
	if !found {
		b.AddStep("L4", trace.ActionDone, "All cells filled successfully; puzzle solved", sudokuSnapshot(*board, given, -1, -1, 0, depth, "done", "solved"), map[string]any{"row": -1, "col": -1})
		return true
	}

	b.AddStep("L2", trace.ActionVisit, fmt.Sprintf("Locate empty cell at (%d, %d)", r, c), sudokuSnapshot(*board, given, r, c, 0, depth, "select", "searching"), map[string]any{"row": r, "col": c})

	for num := 1; num <= 9; num++ {
		vars := map[string]any{
			"row":       r,
			"col":       c,
			"candidate": num,
			"num":       num,
		}
		b.IncComparisons()
		b.AddStep("L6", trace.ActionCompare, fmt.Sprintf("Evaluating candidate %d for cell (%d, %d)", num, r, c), sudokuSnapshot(*board, given, r, c, num, depth, "try", "searching"), vars)

		if IsValid(*board, r, c, num) {
			board[r][c] = num
			b.IncOperations()
			b.AddStep("L8", trace.ActionSet, fmt.Sprintf("Candidate %d satisfies constraints; place at (%d, %d)", num, r, c), sudokuSnapshot(*board, given, r, c, num, depth, "place", "searching"), vars)

			b.AddStep("L9", trace.ActionVisit, fmt.Sprintf("Advance to next empty cell with %d at (%d, %d)", num, r, c), sudokuSnapshot(*board, given, r, c, num, depth, "descend", "searching"), vars)

			if traceBacktrack(board, given, depth+1, b) {
				return true
			}

			board[r][c] = 0
			b.IncOperations()
			b.AddStep("L12", trace.ActionPop, fmt.Sprintf("Backtrack: reset cell (%d, %d) from %d to 0", r, c, num), sudokuSnapshot(*board, given, r, c, num, depth, "remove", "searching"), vars)
		} else {
			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Candidate %d violates row, column, or box constraints at (%d, %d)", num, r, c), sudokuSnapshot(*board, given, r, c, num, depth, "reject", "searching"), vars)
		}
	}

	b.AddStep("L15", trace.ActionPop, fmt.Sprintf("Exhausted candidates 1-9 at (%d, %d); trigger backtrack", r, c), sudokuSnapshot(*board, given, r, c, 0, depth, "exhausted", "searching"), map[string]any{"row": r, "col": c})
	return false
}

func sudokuSnapshot(board [9][9]int, given [9][9]bool, row, col, candidate, depth int, phase, status string) renderers.SudokuLessonState {
	s := renderers.SudokuLessonState{TableState: buildTableState(board, row, col, "active"), Board: board, Given: given, Row: row, Col: col, Candidate: candidate, HasCandidate: candidate > 0, Depth: depth, Phase: phase, Status: status}
	if (phase != "try" && phase != "reject") || row < 0 || col < 0 {
		return s
	}
	// First witness: entire row, entire column, then box, each in increasing order.
	for c := 0; c < 9; c++ {
		if c != col && board[row][c] == candidate {
			s.Conflicts = [][2]int{{row, c}}
			s.ConflictKinds = []string{"row"}
			return s
		}
	}
	for r := 0; r < 9; r++ {
		if r != row && board[r][col] == candidate {
			s.Conflicts = [][2]int{{r, col}}
			s.ConflictKinds = []string{"column"}
			return s
		}
	}
	for r := row / 3 * 3; r < row/3*3+3; r++ {
		for c := col / 3 * 3; c < col/3*3+3; c++ {
			if (r != row || c != col) && board[r][c] == candidate {
				s.Conflicts = [][2]int{{r, c}}
				s.ConflictKinds = []string{"box"}
				return s
			}
		}
	}
	return s
}
