package nqueens

// Solve returns all distinct solutions to the N-Queens puzzle on an n x n board.
// Each solution is represented as a slice of length n, where index r is the row
// and the value is the column index of the queen placed on that row.
func Solve(n int) [][]int {
	if n <= 0 {
		return nil
	}
	var solutions [][]int
	board := make([]int, n)
	for i := range board {
		board[i] = -1
	}

	var backtrack func(row int)
	backtrack = func(row int) {
		if row == n {
			sol := make([]int, n)
			copy(sol, board)
			solutions = append(solutions, sol)
			return
		}

		for col := 0; col < n; col++ {
			if isSafe(row, col, board) {
				board[row] = col
				backtrack(row + 1)
				board[row] = -1
			}
		}
	}

	backtrack(0)
	return solutions
}

func isSafe(row, col int, board []int) bool {
	for r := 0; r < row; r++ {
		c := board[r]
		if c == col {
			return false
		}
		diffCol := c - col
		if diffCol < 0 {
			diffCol = -diffCol
		}
		diffRow := r - row
		if diffRow < 0 {
			diffRow = -diffRow
		}
		if diffCol == diffRow {
			return false
		}
	}
	return true
}
