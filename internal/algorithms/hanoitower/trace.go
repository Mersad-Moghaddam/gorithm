package hanoitower

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Tower of Hanoi with n disks.
func Trace(n int) []trace.Step {
	if n < 1 {
		n = 3
	} else if n > 7 {
		n = 7
	}

	b := trace.NewBuilder()
	pegs := [3][]int{
		make([]int, 0, n),
		make([]int, 0, n),
		make([]int, 0, n),
	}
	for d := n; d >= 1; d-- {
		pegs[0] = append(pegs[0], d)
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Tower of Hanoi with %d disk(s) on Peg A. Goal: transfer all disks to Peg C using Peg B.", n), renderers.HanoiState{
		Pegs: clonePegs(pegs),
		Move: "",
	}, map[string]any{"n": n, "from": "A", "to": "C", "aux": "B", "invariant": "valid"})

	var hanoiTraced func(k, from, to, aux int)
	hanoiTraced = func(k, from, to, aux int) {
		b.AddStep("L2", trace.ActionCompare, fmt.Sprintf("Check base case for subproblem: n = %d.", k), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: "",
		}, map[string]any{"n": k, "from": pegLabel(from), "to": pegLabel(to), "aux": pegLabel(aux)})

		if k == 0 {
			b.AddStep("L3", trace.ActionVisit, "Base case reached: n = 0 (no disks to move). Return to caller.", renderers.HanoiState{
				Pegs: clonePegs(pegs),
				Move: "",
			}, map[string]any{"n": 0})
			return
		}

		// Step 1: Move top k-1 disks from source to auxiliary peg
		b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Step 1 (Recursive Subproblem): Move top %d disk(s) from source Peg %s to auxiliary Peg %s using target Peg %s.", k-1, pegLabel(from), pegLabel(aux), pegLabel(to)), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: "",
		}, map[string]any{"step": 1, "n": k - 1, "from": pegLabel(from), "to": pegLabel(aux), "aux": pegLabel(to)})
		hanoiTraced(k-1, from, aux, to)

		// Step 2: Move largest disk k from source to target peg
		fromPeg := pegs[from]
		if len(fromPeg) == 0 {
			panic(fmt.Sprintf("illegal move: source peg %s is empty", pegLabel(from)))
		}
		disk := fromPeg[len(fromPeg)-1]
		toPeg := pegs[to]

		// Verify invariant: no larger disk ever placed on a smaller disk!
		if len(toPeg) > 0 && toPeg[len(toPeg)-1] < disk {
			panic(fmt.Sprintf("invariant violation: cannot place larger disk %d on top of smaller disk %d on Peg %s", disk, toPeg[len(toPeg)-1], pegLabel(to)))
		}

		// Move disk: leaves source peg and lands on target peg
		pegs[from] = fromPeg[:len(fromPeg)-1]
		pegs[to] = append(pegs[to], disk)

		b.IncOperations()
		moveText := fmt.Sprintf("Disk %d: Peg %s -> Peg %s", disk, pegLabel(from), pegLabel(to))
		step2Desc := fmt.Sprintf("Step 2: Move disk %d from source Peg %s to target Peg %s (disk leaves Peg %s and lands on Peg %s). Invariant verified: legal move onto Peg %s.", disk, pegLabel(from), pegLabel(to), pegLabel(from), pegLabel(to), pegLabel(to))

		b.AddStep("L6", trace.ActionSwap, step2Desc, renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: moveText,
		}, map[string]any{"step": 2, "disk": disk, "from": pegLabel(from), "to": pegLabel(to), "invariant": "valid"})

		// Step 3: Move k-1 disks from auxiliary to target peg
		b.AddStep("L7", trace.ActionPush, fmt.Sprintf("Step 3 (Recursive Subproblem): Move %d disk(s) from auxiliary Peg %s to target Peg %s using source Peg %s.", k-1, pegLabel(aux), pegLabel(to), pegLabel(from)), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: "",
		}, map[string]any{"step": 3, "n": k - 1, "from": pegLabel(aux), "to": pegLabel(to), "aux": pegLabel(from)})
		hanoiTraced(k-1, aux, to, from)
	}

	hanoiTraced(n, 0, 2, 1)

	b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Tower of Hanoi completed! All %d disk(s) transferred to Peg C in %d moves.", n, (1<<n)-1), renderers.HanoiState{
		Pegs: clonePegs(pegs),
		Move: "Done",
	}, map[string]any{"status": "complete", "totalDisks": n, "operations": (1<<n)-1, "invariant": "valid"})

	return b.Steps()
}

func clonePegs(pegs [3][]int) [3][]int {
	var cloned [3][]int
	for i := range pegs {
		if pegs[i] == nil {
			cloned[i] = []int{}
		} else {
			cloned[i] = make([]int, len(pegs[i]))
			copy(cloned[i], pegs[i])
		}
	}
	return cloned
}

func pegLabel(p int) string {
	return string(rune('A' + p))
}
