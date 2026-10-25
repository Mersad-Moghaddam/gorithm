package hanoitower

// Move records a single disk transition between pegs.
type Move struct {
	Disk int
	From int
	To   int
}

// Solve returns the optimal sequence of moves to solve Tower of Hanoi for n disks.
func Solve(n int) []Move {
	if n <= 0 {
		return nil
	}
	var moves []Move
	var hanoi func(n, from, to, aux int)
	hanoi = func(n, from, to, aux int) {
		if n == 0 {
			return
		}
		hanoi(n-1, from, aux, to)
		moves = append(moves, Move{Disk: n, From: from, To: to})
		hanoi(n-1, aux, to, from)
	}
	hanoi(n, 0, 2, 1)
	return moves
}
