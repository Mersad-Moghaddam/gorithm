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

func parseAndValidateBoard(input json.RawMessage) ([9][9]int, error) {
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return CanonicalBoard, nil
	}
	if bytes.Equal(trimmed, []byte("null")) {
		return [9][9]int{}, fmt.Errorf("input cannot be null")
	}

	var rawGrid [][]*int
	if trimmed[0] == '{' {
		var payload struct {
			Board *[][]*int `json:"board"`
		}
		if err := json.Unmarshal(trimmed, &payload); err != nil {
			return [9][9]int{}, fmt.Errorf("malformed JSON input: %w", err)
		}
		if payload.Board == nil {
			return [9][9]int{}, fmt.Errorf("board cannot be null")
		}
		rawGrid = *payload.Board
	} else if trimmed[0] == '[' {
		if err := json.Unmarshal(trimmed, &rawGrid); err != nil {
			return [9][9]int{}, fmt.Errorf("malformed JSON input: %w", err)
		}
	} else {
		return [9][9]int{}, fmt.Errorf("malformed JSON input: expected JSON object or array")
	}

	if len(rawGrid) != 9 {
		return [9][9]int{}, fmt.Errorf("invalid board dimensions: expected 9 rows, got %d", len(rawGrid))
	}
	var board [9][9]int
	for r := 0; r < 9; r++ {
		row := rawGrid[r]
		if row == nil {
			return [9][9]int{}, fmt.Errorf("invalid board: row %d cannot be null", r)
		}
		if len(row) != 9 {
			return [9][9]int{}, fmt.Errorf("invalid board dimensions: row %d has %d columns, expected 9", r, len(row))
		}
		for c := 0; c < 9; c++ {
			if row[c] == nil {
				return [9][9]int{}, fmt.Errorf("invalid board: cell (%d, %d) cannot be null", r, c)
			}
			val := *row[c]
			if val < 0 || val > 9 {
				return [9][9]int{}, fmt.Errorf("invalid cell value at (%d, %d): %d (must be 0-9)", r, c, val)
			}
			board[r][c] = val
		}
	}
	return board, nil
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	board, err := parseAndValidateBoard(input)
	if err != nil {
		return nil, err
	}
	return Trace(board), nil
}

func init() {
	catalog.Register(Lesson{})
}
