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
	}

	b := trace.NewBuilder()
	pegs := [3][]int{}
	for d := n; d >= 1; d-- {
		pegs[0] = append(pegs[0], d)
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Tower of Hanoi with %d disks on Peg A.", n), renderers.HanoiState{
		Pegs: clonePegs(pegs),
		Move: "",
	}, map[string]any{"n": n})

	var hanoiTraced func(k, from, to, aux int)
	hanoiTraced = func(k, from, to, aux int) {
		b.AddStep("L2", trace.ActionCompare, fmt.Sprintf("Check base case: n = %d.", k), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: "",
		}, map[string]any{"n": k, "from": pegLabel(from), "to": pegLabel(to), "aux": pegLabel(aux)})

		if k == 0 {
			b.AddStep("L3", trace.ActionVisit, "Base case reached: n = 0, returning.", renderers.HanoiState{
				Pegs: clonePegs(pegs),
				Move: "",
			}, map[string]any{"n": k})
			return
		}

		b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Solve subproblem: move %d disks from Peg %s to Peg %s.", k-1, pegLabel(from), pegLabel(aux)), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: "",
		}, map[string]any{"n": k - 1, "from": pegLabel(from), "to": pegLabel(aux)})
		hanoiTraced(k-1, from, aux, to)

		// Move disk k from peg 'from' to peg 'to'
		fromPeg := pegs[from]
		disk := fromPeg[len(fromPeg)-1]
		pegs[from] = fromPeg[:len(fromPeg)-1]
		pegs[to] = append(pegs[to], disk)

		b.IncOperations()
		moveText := fmt.Sprintf("Disk %d: Peg %s -> Peg %s", disk, pegLabel(from), pegLabel(to))
		b.AddStep("L6", trace.ActionSwap, fmt.Sprintf("Moved disk %d from Peg %s to Peg %s.", disk, pegLabel(from), pegLabel(to)), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: moveText,
		}, map[string]any{"disk": disk, "from": pegLabel(from), "to": pegLabel(to)})

		b.AddStep("L7", trace.ActionPush, fmt.Sprintf("Solve subproblem: move %d disks from Peg %s to Peg %s.", k-1, pegLabel(aux), pegLabel(to)), renderers.HanoiState{
			Pegs: clonePegs(pegs),
			Move: "",
		}, map[string]any{"n": k - 1, "from": pegLabel(aux), "to": pegLabel(to)})
		hanoiTraced(k-1, aux, to, from)
	}

	hanoiTraced(n, 0, 2, 1)

	b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Tower of Hanoi completed! All %d disks transferred to Peg C.", n), renderers.HanoiState{
		Pegs: clonePegs(pegs),
		Move: "Done",
	}, map[string]any{"status": "complete", "totalDisks": n})

	return b.Steps()
}

func clonePegs(pegs [3][]int) [3][]int {
	var cloned [3][]int
	for i := range pegs {
		cloned[i] = make([]int, len(pegs[i]))
		copy(cloned[i], pegs[i])
	}
	return cloned
}

func pegLabel(p int) string {
	return string(rune('A' + p))
}
