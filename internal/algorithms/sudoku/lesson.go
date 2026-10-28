package sudoku

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

var CanonicalBoard = [9][9]int{
	{5, 3, 0, 0, 7, 8, 9, 1, 2},
	{6, 7, 2, 1, 9, 5, 3, 4, 8},
	{0, 9, 8, 3, 4, 2, 5, 6, 7},
	{8, 5, 9, 7, 6, 1, 4, 2, 3},
	{4, 2, 6, 8, 0, 3, 7, 9, 1},
	{7, 1, 3, 9, 2, 4, 8, 5, 6},
	{9, 6, 1, 5, 3, 7, 0, 8, 4},
	{2, 8, 7, 4, 1, 9, 6, 3, 5},
	{3, 4, 5, 2, 8, 6, 1, 0, 0},
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "sudoku",
		Name:     "Sudoku Solver",
		Category: catalog.CategoryBacktracking,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Solves 9x9 Sudoku puzzles using recursive depth-first backtracking with row, column, and 3x3 subgrid constraint propagation.",
			TimeComplexity:  "O(9^(empty_cells))",
			SpaceComplexity: "O(1)",
			Difficulty:      "Hard",
			KeyConcepts:     []string{"Exact Cover", "Row/Column/Box Invariant", "Recursive Backtracking", "Constraint Propagation"},
			HowItWorks:      "Locates the first unfilled cell and iteratively tests candidates 1 through 9. For each candidate satisfying row, column, and 3x3 box invariants, it places the digit and recursively proceeds. If a dead end is encountered, it resets the cell to zero (backtracks) and attempts alternative branches until the grid is completely satisfied.",
		},
		DefaultInput: `{"board":[[5,3,0,0,7,8,9,1,2],[6,7,2,1,9,5,3,4,8],[0,9,8,3,4,2,5,6,7],[8,5,9,7,6,1,4,2,3],[4,2,6,8,0,3,7,9,1],[7,1,3,9,2,4,8,5,6],[9,6,1,5,3,7,0,8,4],[2,8,7,4,1,9,6,3,5],[3,4,5,2,8,6,1,0,0]]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func SolveSudoku(board *[9][9]int) bool {"},
		{ID: "L2", Number: 2, Code: "    row, col, found := findEmpty(board)"},
		{ID: "L3", Number: 3, Code: "    if !found {"},
		{ID: "L4", Number: 4, Code: "        return true // puzzle solved"},
		{ID: "L5", Number: 5, Code: "    }"},
		{ID: "L6", Number: 6, Code: "    for num := 1; num <= 9; num++ {"},
		{ID: "L7", Number: 7, Code: "        if isValid(board, row, col, num) {"},
		{ID: "L8", Number: 8, Code: "            board[row][col] = num"},
		{ID: "L9", Number: 9, Code: "            if SolveSudoku(board) {"},
		{ID: "L10", Number: 10, Code: "                return true"},
		{ID: "L11", Number: 11, Code: "            }"},
		{ID: "L12", Number: 12, Code: "            board[row][col] = 0 // backtrack"},
		{ID: "L13", Number: 13, Code: "        }"},
		{ID: "L14", Number: 14, Code: "    }"},
		{ID: "L15", Number: 15, Code: "    return false"},
		{ID: "L16", Number: 16, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	board := CanonicalBoard
	input = bytes.TrimSpace(input)
	if len(input) == 0 {
		return Trace(board), nil
	}
	if bytes.Equal(input, []byte("null")) {
		return nil, fmt.Errorf("board must not be null")
	}
	raw := input
	if input[0] == '{' {
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(input, &payload); err != nil {
			return nil, fmt.Errorf("invalid Sudoku JSON: %w", err)
		}
		var exists bool
		raw, exists = payload["board"]
		if !exists {
			return Trace(board), nil
		}
	}
	var rows [][]int
	if bytes.Contains(raw, []byte("null")) {
		return nil, fmt.Errorf("board entries and rows must not be null")
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("board must contain integer rows: %w", err)
	}
	if len(rows) != 9 {
		return nil, fmt.Errorf("board must have exactly 9 rows")
	}
	for r, row := range rows {
		if len(row) != 9 {
			return nil, fmt.Errorf("row %d must have exactly 9 entries", r)
		}
		for c, value := range row {
			if value < 0 || value > 9 {
				return nil, fmt.Errorf("cell (%d,%d) must be in [0,9]", r, c)
			}
			board[r][c] = value
		}
	}
	if !validateBoard(board) {
		return nil, fmt.Errorf("board has conflicting givens")
	}
	return Trace(board), nil
}

func init() {
	catalog.Register(Lesson{})
}
