package nqueens

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func buildTableState(n int, board []int, activeCells map[[2]int]string, candidate *[2]int) renderers.TableState {
	rowHeaders := make([]string, n)
	colHeaders := make([]string, n)
	for i := 0; i < n; i++ {
		rowHeaders[i] = fmt.Sprintf("Row %d", i)
		colHeaders[i] = fmt.Sprintf("Col %d", i)
	}

	cells := make([]renderers.TableCell, 0, n*n)
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			val := "."
			st := "normal"
			if board[r] == c {
				val = "Q"
				st = "computed"
			}
			if candidate != nil && candidate[0] == r && candidate[1] == c {
				val = "Q"
				st = "active"
			}
			if cellState, ok := activeCells[[2]int{r, c}]; ok {
				st = cellState
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

// Trace simulates N-Queens recursive backtracking and records visualization steps.
func Trace(n int) []trace.Step {
	if n <= 0 {
		n = 4
	}
	b := trace.NewBuilder()
	board := make([]int, n)
	for i := range board {
		board[i] = -1
	}

	solutionsCount := 0

	b.AddStep(
		"L2",
		trace.ActionInit,
		fmt.Sprintf("Initialize %dx%d chessboard with no queens placed", n, n),
		buildTableState(n, board, nil, nil),
		map[string]any{"row": 0, "col": 0, "solutions": 0, "n": n},
	)

	var backtrack func(row int)
	backtrack = func(row int) {
		b.AddStep(
			"L6",
			trace.ActionCompare,
			fmt.Sprintf("Row %d: check if all %d queens have been placed", row, n),
			buildTableState(n, board, nil, nil),
			map[string]any{"row": row, "solutions": solutionsCount, "n": n},
		)

		if row == n {
			solutionsCount++
			b.AddStep(
				"L7",
				trace.ActionDone,
				fmt.Sprintf("Solution #%d found! All %d queens placed safely", solutionsCount, n),
				buildTableState(n, board, nil, nil),
				map[string]any{"row": row, "solutions": solutionsCount, "n": n},
			)
			return
		}

		for col := 0; col < n; col++ {
			cand := &[2]int{row, col}
			b.AddStep(
				"L10",
				trace.ActionVisit,
				fmt.Sprintf("Row %d: inspect column %d as potential queen placement", row, col),
				buildTableState(n, board, nil, cand),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)

			confRow := -1
			confCol := -1
			for r := 0; r < row; r++ {
				c := board[r]
				diffCol := c - col
				if diffCol < 0 {
					diffCol = -diffCol
				}
				diffRow := r - row
				if diffRow < 0 {
					diffRow = -diffRow
				}
				if c == col || diffCol == diffRow {
					confRow = r
					confCol = c
					break
				}
			}

			b.IncComparisons()

			if confRow != -1 {
				conflictMap := map[[2]int]string{
					{confRow, confCol}: "active",
					{row, col}:         "active",
				}
				b.AddStep(
					"L21",
					trace.ActionCompare,
					fmt.Sprintf("Conflict detected between (row %d, col %d) and existing queen at (row %d, col %d)", row, col, confRow, confCol),
					buildTableState(n, board, conflictMap, cand),
					map[string]any{"row": row, "col": col, "conflictRow": confRow, "conflictCol": confCol, "solutions": solutionsCount, "n": n},
				)
				continue
			}

			b.AddStep(
				"L24",
				trace.ActionCompare,
				fmt.Sprintf("Position (row %d, col %d) is safe; no column or diagonal conflicts", row, col),
				buildTableState(n, board, nil, cand),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)

			board[row] = col
			b.IncOperations()
			b.AddStep(
				"L12",
				trace.ActionSet,
				fmt.Sprintf("Place queen at row %d, col %d", row, col),
				buildTableState(n, board, nil, nil),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)

			b.AddStep(
				"L13",
				trace.ActionPush,
				fmt.Sprintf("Recurse to row %d to place next queen", row+1),
				buildTableState(n, board, nil, nil),
				map[string]any{"row": row + 1, "col": 0, "solutions": solutionsCount, "n": n},
			)

			backtrack(row + 1)

			board[row] = -1
			b.IncOperations()
			b.AddStep(
				"L14",
				trace.ActionPop,
				fmt.Sprintf("Backtrack: remove queen from row %d, col %d", row, col),
				buildTableState(n, board, nil, nil),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)
		}
	}

	backtrack(0)

	b.AddStep(
		"L1",
		trace.ActionDone,
		fmt.Sprintf("N-Queens backtracking complete. Found %d valid solutions", solutionsCount),
		buildTableState(n, board, nil, nil),
		map[string]any{"solutions": solutionsCount, "n": n},
	)

	return b.Steps()
}
