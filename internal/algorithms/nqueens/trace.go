package nqueens

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func cloneTableState(s renderers.TableState) renderers.TableState {
	rowHeaders := make([]string, len(s.RowHeaders))
	copy(rowHeaders, s.RowHeaders)
	colHeaders := make([]string, len(s.ColHeaders))
	copy(colHeaders, s.ColHeaders)
	cells := make([]renderers.TableCell, len(s.Cells))
	copy(cells, s.Cells)
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

func buildTableState(caption string, n int, board []int, activeCells map[[2]int]string, candidate *[2]int) renderers.TableState {
	rowHeaders := make([]string, n)
	colHeaders := make([]string, n)
	for i := 0; i < n; i++ {
		rowHeaders[i] = fmt.Sprintf("Row %d", i)
		colHeaders[i] = fmt.Sprintf("Col %d", i)
	}

	cells := make([]renderers.TableCell, n*n)
	idx := 0
	for r := 0; r < n; r++ {
		for c := 0; c < n; c++ {
			val := "."
			st := "normal"
			if board[r] == c {
				val = "👑"
				st = "computed"
			}
			if candidate != nil && candidate[0] == r && candidate[1] == c {
				val = "👑"
				st = "active"
			}
			if cellState, ok := activeCells[[2]int{r, c}]; ok {
				st = cellState
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

	return cloneTableState(renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	})
}

// Trace simulates N-Queens recursive backtracking and records visualization steps.
func Trace(n int) []trace.Step {
	if n <= 0 {
		n = 4
	}
	if n > 8 {
		n = 8
	}
	b := trace.NewBuilder()
	board := make([]int, n)
	for i := range board {
		board[i] = -1
	}

	solutionsCount := 0

	initDesc := fmt.Sprintf("Initialize %dx%d chessboard with no queens placed", n, n)
	b.AddStep(
		"L2",
		trace.ActionInit,
		initDesc,
		buildTableState(initDesc, n, board, nil, nil),
		map[string]any{"row": 0, "col": 0, "solutions": 0, "n": n},
	)

	var backtrack func(row int) bool
	backtrack = func(row int) bool {
		checkDesc := fmt.Sprintf("Row %d: check if all %d queens have been placed", row, n)
		b.AddStep(
			"L6",
			trace.ActionCompare,
			checkDesc,
			buildTableState(checkDesc, n, board, nil, nil),
			map[string]any{"row": row, "solutions": solutionsCount, "n": n},
		)

		if row == n {
			solutionsCount++
			solDesc := fmt.Sprintf("Solution #%d found! Valid %dx%d board configuration", solutionsCount, n, n)
			b.AddStep(
				"L7",
				trace.ActionDone,
				solDesc,
				buildTableState(solDesc, n, board, nil, nil),
				map[string]any{"row": row, "solutions": solutionsCount, "n": n},
			)
			return true
		}

		placedAny := false
		for col := 0; col < n; col++ {
			cand := &[2]int{row, col}
			visitDesc := fmt.Sprintf("Row %d: inspect column %d as potential queen placement", row, col)
			b.AddStep(
				"L10",
				trace.ActionVisit,
				visitDesc,
				buildTableState(visitDesc, n, board, nil, cand),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)

			confRow := -1
			confCol := -1
			threatType := ""
			for r := 0; r < row; r++ {
				b.IncComparisons()
				c := board[r]
				if c == col {
					confRow = r
					confCol = c
					threatType = "column"
					break
				}
				diffCol := c - col
				if diffCol < 0 {
					diffCol = -diffCol
				}
				diffRow := row - r
				if diffCol == diffRow {
					confRow = r
					confCol = c
					threatType = "diagonal"
					break
				}
			}

			if confRow != -1 {
				conflictMap := map[[2]int]string{
					{confRow, confCol}: "active",
					{row, col}:         "active",
				}
				threatDesc := fmt.Sprintf("Threat detected: shares %s with queen at (%d, %d)", threatType, confRow, confCol)
				b.AddStep(
					"L21",
					trace.ActionCompare,
					threatDesc,
					buildTableState(threatDesc, n, board, conflictMap, cand),
					map[string]any{
						"row":         row,
						"col":         col,
						"conflictRow": confRow,
						"conflictCol": confCol,
						"threat":      threatType,
						"solutions":   solutionsCount,
						"n":           n,
					},
				)
				continue
			}

			safeDesc := fmt.Sprintf("Position (row %d, col %d) is safe; no conflicts with existing queens", row, col)
			b.AddStep(
				"L24",
				trace.ActionCompare,
				safeDesc,
				buildTableState(safeDesc, n, board, nil, cand),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)

			board[row] = col
			placedAny = true
			b.IncOperations()
			placeDesc := fmt.Sprintf("Place queen at row %d, col %d", row, col)
			b.AddStep(
				"L12",
				trace.ActionSet,
				placeDesc,
				buildTableState(placeDesc, n, board, nil, nil),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)

			recurseDesc := fmt.Sprintf("Recurse to row %d to place next queen", row+1)
			b.AddStep(
				"L13",
				trace.ActionPush,
				recurseDesc,
				buildTableState(recurseDesc, n, board, nil, nil),
				map[string]any{"row": row + 1, "col": 0, "solutions": solutionsCount, "n": n},
			)

			childFound := backtrack(row + 1)

			board[row] = -1
			b.IncOperations()

			var backtrackDesc string
			if row+1 == n && childFound {
				backtrackDesc = fmt.Sprintf("Backtrack from row %d, col %d to search for further solutions", row, col)
			} else {
				backtrackDesc = fmt.Sprintf("No valid placement in row %d; backtrack to row %d", row+1, row)
			}

			b.AddStep(
				"L14",
				trace.ActionPop,
				backtrackDesc,
				buildTableState(backtrackDesc, n, board, nil, nil),
				map[string]any{"row": row, "col": col, "solutions": solutionsCount, "n": n},
			)
		}

		return placedAny
	}

	backtrack(0)

	doneDesc := fmt.Sprintf("N-Queens backtracking complete. Found %d valid solutions", solutionsCount)
	b.AddStep(
		"L1",
		trace.ActionDone,
		doneDesc,
		buildTableState(doneDesc, n, board, nil, nil),
		map[string]any{"solutions": solutionsCount, "n": n},
	)

	return b.Steps()
}
