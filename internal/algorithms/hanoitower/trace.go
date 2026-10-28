package hanoitower

import (
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace generates execution steps for Tower of Hanoi with n disks.
func Trace(n int) []trace.Step {
	if n < 0 {
		n = 3
	}

	b := trace.NewBuilder()
	var calls []renderers.HanoiTowerCall
	moveCount := 0
	pegs := [3][]int{}
	for d := n; d >= 1; d-- {
		pegs[0] = append(pegs[0], d)
	}
	snapshot := func(phase string, disk, from, to int, before [3][]int, move string) renderers.HanoiTowerLessonState {
		return renderers.HanoiTowerLessonState{HanoiState: renderers.HanoiState{Pegs: clonePegs(pegs), Move: move}, TotalDisks: n, MoveCount: moveCount, Calls: append([]renderers.HanoiTowerCall(nil), calls...), Phase: phase, Disk: disk, From: from, To: to, BeforePegs: clonePegs(before)}
	}

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Tower of Hanoi with %d disks on Peg A.", n), snapshot("init", -1, -1, -1, [3][]int{}, ""), map[string]any{"n": n})

	var hanoiTraced func(k, from, to, aux int)
	hanoiTraced = func(k, from, to, aux int) {
		calls = append(calls, renderers.HanoiTowerCall{N: k, From: from, To: to, Aux: aux})
		defer func() { calls = calls[:len(calls)-1] }()
		b.AddStep("L2", trace.ActionCompare, fmt.Sprintf("Check base case: n = %d.", k), snapshot("base-check", -1, -1, -1, [3][]int{}, ""), map[string]any{"n": k, "from": pegLabel(from), "to": pegLabel(to), "aux": pegLabel(aux)})

		if k == 0 {
			b.AddStep("L3", trace.ActionVisit, "Base case reached: n = 0, returning.", snapshot("return", -1, -1, -1, [3][]int{}, ""), map[string]any{"n": k})
			return
		}

		b.AddStep("L5", trace.ActionPush, fmt.Sprintf("Solve subproblem: move %d disks from Peg %s to Peg %s.", k-1, pegLabel(from), pegLabel(aux)), snapshot("left-call", -1, -1, -1, [3][]int{}, ""), map[string]any{"n": k - 1, "from": pegLabel(from), "to": pegLabel(aux)})
		hanoiTraced(k-1, from, aux, to)

		// Move disk k from peg 'from' to peg 'to'
		before := clonePegs(pegs)
		fromPeg := pegs[from]
		disk := fromPeg[len(fromPeg)-1]
		pegs[from] = fromPeg[:len(fromPeg)-1]
		pegs[to] = append(pegs[to], disk)

		b.IncOperations()
		moveCount++
		moveText := fmt.Sprintf("Disk %d: Peg %s -> Peg %s", disk, pegLabel(from), pegLabel(to))
		b.AddStep("L6", trace.ActionSwap, fmt.Sprintf("Moved disk %d from Peg %s to Peg %s.", disk, pegLabel(from), pegLabel(to)), snapshot("move", disk, from, to, before, moveText), map[string]any{"disk": disk, "from": pegLabel(from), "to": pegLabel(to)})

		b.AddStep("L7", trace.ActionPush, fmt.Sprintf("Solve subproblem: move %d disks from Peg %s to Peg %s.", k-1, pegLabel(aux), pegLabel(to)), snapshot("right-call", -1, -1, -1, [3][]int{}, ""), map[string]any{"n": k - 1, "from": pegLabel(aux), "to": pegLabel(to)})
		hanoiTraced(k-1, aux, to, from)
	}

	hanoiTraced(n, 0, 2, 1)

	b.AddStep("L8", trace.ActionDone, fmt.Sprintf("Tower of Hanoi completed! All %d disks transferred to Peg C.", n), snapshot("done", -1, -1, -1, [3][]int{}, "Done"), map[string]any{"status": "complete", "totalDisks": n})

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
