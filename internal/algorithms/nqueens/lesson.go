package nqueens

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type lessonInput struct {
	N *int `json:"n"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "nqueens",
		Name:     "N-Queens Puzzle",
		Category: catalog.CategoryBacktracking,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Places N non-attacking queens on an N×N chessboard using recursive backtracking with state space tree pruning.",
			TimeComplexity:  "O(N!)",
			SpaceComplexity: "O(N)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"Backtracking",
				"Constraint Satisfaction",
				"Diagonal Conflict Checking",
				"State Space Tree Pruning",
			},
			HowItWorks: "Places queens row by row. For each row, iterates through columns 0 to N-1 and checks for conflicts with previously placed queens (same column, main diagonal, anti-diagonal). If a safe column is found, the queen is placed and the algorithm recurses to the next row. If a dead end is reached (no valid column in row), the algorithm backtracks by removing the queen from the previous row and trying the next available column. When all N rows are filled, a valid board configuration is recorded.",
		},
		DefaultInput: `{"n": 4}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func solveNQueens(n int) [][]int {"},
		{ID: "L2", Number: 2, Code: "    board := make([]int, n)"},
		{ID: "L3", Number: 3, Code: "    backtrack(0, board, n)"},
		{ID: "L4", Number: 4, Code: "}"},
		{ID: "L5", Number: 5, Code: "func backtrack(row int, board []int, n int) {"},
		{ID: "L6", Number: 6, Code: "    if row == n {"},
		{ID: "L7", Number: 7, Code: "        recordSolution(board) // all queens placed"},
		{ID: "L8", Number: 8, Code: "        return"},
		{ID: "L9", Number: 9, Code: "    }"},
		{ID: "L10", Number: 10, Code: "    for col := 0; col < n; col++ {"},
		{ID: "L11", Number: 11, Code: "        if isSafe(row, col, board) {"},
		{ID: "L12", Number: 12, Code: "            board[row] = col // place queen"},
		{ID: "L13", Number: 13, Code: "            backtrack(row + 1, board, n)"},
		{ID: "L14", Number: 14, Code: "            board[row] = -1  // backtrack"},
		{ID: "L15", Number: 15, Code: "        }"},
		{ID: "L16", Number: 16, Code: "    }"},
		{ID: "L17", Number: 17, Code: "}"},
		{ID: "L18", Number: 18, Code: "func isSafe(row, col int, board []int) bool {"},
		{ID: "L19", Number: 19, Code: "    for r := 0; r < row; r++ {"},
		{ID: "L20", Number: 20, Code: "        if board[r] == col || abs(board[r]-col) == row-r {"},
		{ID: "L21", Number: 21, Code: "            return false // conflict"},
		{ID: "L22", Number: 22, Code: "        }"},
		{ID: "L23", Number: 23, Code: "    }"},
		{ID: "L24", Number: 24, Code: "    return true"},
		{ID: "L25", Number: 25, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	n := 4
	trimmed := bytes.TrimSpace(input)
	if len(trimmed) > 0 {
		if bytes.Equal(trimmed, []byte("null")) {
			return nil, fmt.Errorf("input cannot be null")
		}

		if trimmed[0] == '{' {
			var in lessonInput
			if err := json.Unmarshal(trimmed, &in); err != nil {
				return nil, fmt.Errorf("malformed JSON object: %w", err)
			}
			if in.N == nil {
				return nil, fmt.Errorf("field 'n' is required and cannot be null")
			}
			if *in.N < 1 || *in.N > 8 {
				return nil, fmt.Errorf("n must be between 1 and 8, got %d", *in.N)
			}
			n = *in.N
		} else if (trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-' {
			var num int
			if err := json.Unmarshal(trimmed, &num); err != nil {
				return nil, fmt.Errorf("malformed JSON integer: %w", err)
			}
			if num < 1 || num > 8 {
				return nil, fmt.Errorf("n must be between 1 and 8, got %d", num)
			}
			n = num
		} else {
			return nil, fmt.Errorf("malformed JSON input: expected JSON object or integer")
		}
	}
	return Trace(n), nil
}

func init() {
	catalog.Register(Lesson{})
}
