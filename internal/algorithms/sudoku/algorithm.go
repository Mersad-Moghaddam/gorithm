package sudoku

// ponytail: simple depth-first backtracking search; Dancing Links (Algorithm X) or bitboard constraints needed for 16x16 or hard benchmark grids.

// Conflict records a constraint violation in row, column, or 3x3 box.
type Conflict struct {
	Row       int    `json:"row"`
	Col       int    `json:"col"`
	Invariant string `json:"invariant"` // "row", "column", "box"
	Message   string `json:"message"`
}

// FindConflict checks if placing num at board[row][col] violates row, column, or 3x3 box invariants.
func FindConflict(board [9][9]int, row, col, num int) (*Conflict, bool) {
	if num < 1 || num > 9 {
		return nil, false
	}
	// Check row invariant
	for c := 0; c < 9; c++ {
		if c != col && board[row][c] == num {
			return &Conflict{
				Row:       row,
				Col:       c,
				Invariant: "row",
				Message:   "row conflict: digit already exists in row",
			}, true
		}
	}
	// Check column invariant
	for r := 0; r < 9; r++ {
		if r != row && board[r][col] == num {
			return &Conflict{
				Row:       r,
				Col:       col,
				Invariant: "column",
				Message:   "column conflict: digit already exists in column",
			}, true
		}
	}
	// Check 3x3 box invariant
	boxRow := (row / 3) * 3
	boxCol := (col / 3) * 3
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			currR := boxRow + r
			currC := boxCol + c
			if currR == row && currC == col {
				continue
			}
			if board[currR][currC] == num {
				return &Conflict{
					Row:       currR,
					Col:       currC,
					Invariant: "box",
					Message:   "box conflict: digit already exists in 3x3 box",
				}, true
			}
		}
	}
	return nil, false
}

// IsValid checks if placing num at board[row][col] conforms to row, column, and 3x3 box constraints.
func IsValid(board [9][9]int, row, col, num int) bool {
	if num < 1 || num > 9 {
		return false
	}
	_, hasConflict := FindConflict(board, row, col, num)
	return !hasConflict
}

func validateBoard(board [9][9]int) bool {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			val := board[r][c]
			if val != 0 {
				if val < 1 || val > 9 || !IsValid(board, r, c, val) {
					return false
				}
			}
		}
	}
	return true
}

// Solve executes backtracking search to solve a 9x9 Sudoku grid.
func Solve(board [9][9]int) ([9][9]int, bool) {
	if !validateBoard(board) {
		return board, false
	}
	res := board
	solved := solveBacktrack(&res)
	return res, solved
}

func solveBacktrack(board *[9][9]int) bool {
	row, col, found := findEmpty(board)
	if !found {
		return true
	}
	for num := 1; num <= 9; num++ {
		if IsValid(*board, row, col, num) {
			board[row][col] = num
			if solveBacktrack(board) {
				return true
			}
			board[row][col] = 0
		}
	}
	return false
}

func findEmpty(board *[9][9]int) (int, int, bool) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c] == 0 {
				return r, c, true
			}
		}
	}
	return -1, -1, false
}
