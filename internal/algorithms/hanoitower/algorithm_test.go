package hanoitower

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestHanoiSolveMoveCount(t *testing.T) {
	tests := []struct {
		n         int
		wantMoves int
	}{
		{n: 1, wantMoves: 1},
		{n: 2, wantMoves: 3},
		{n: 3, wantMoves: 7},  // 2^3 - 1
		{n: 4, wantMoves: 15}, // 2^4 - 1
		{n: 5, wantMoves: 31},
	}

	for _, tc := range tests {
		moves := Solve(tc.n)
		expected := int(math.Pow(2, float64(tc.n))) - 1
		if len(moves) != expected {
			t.Fatalf("Solve(%d) = %d moves, want %d (2^n - 1)", tc.n, len(moves), expected)
		}
		if len(moves) != tc.wantMoves {
			t.Fatalf("Solve(%d) = %d moves, want %d", tc.n, len(moves), tc.wantMoves)
		}

		// Verify game rules by simulation
		pegs := [3][]int{}
		for d := tc.n; d >= 1; d-- {
			pegs[0] = append(pegs[0], d)
		}

		for moveIdx, m := range moves {
			if m.From < 0 || m.From > 2 || m.To < 0 || m.To > 2 || m.From == m.To {
				t.Fatalf("invalid move pegs at %d: from %d to %d", moveIdx, m.From, m.To)
			}
			fromPeg := pegs[m.From]
			if len(fromPeg) == 0 {
				t.Fatalf("move %d attempted from empty peg %d", moveIdx, m.From)
			}
			topDisk := fromPeg[len(fromPeg)-1]
			if topDisk != m.Disk {
				t.Fatalf("move %d disk mismatch: expected %d on top of peg %d, got %d", moveIdx, m.Disk, m.From, topDisk)
			}
			pegs[m.From] = fromPeg[:len(fromPeg)-1]

			toPeg := pegs[m.To]
			if len(toPeg) > 0 && toPeg[len(toPeg)-1] < m.Disk {
				t.Fatalf("move %d placed larger disk %d on top of smaller disk %d on peg %d", moveIdx, m.Disk, toPeg[len(toPeg)-1], m.To)
			}
			pegs[m.To] = append(toPeg, m.Disk)
		}

		if len(pegs[0]) != 0 || len(pegs[1]) != 0 {
			t.Fatalf("pegs 0 and 1 not empty after Solve(%d): peg0=%v peg1=%v", tc.n, pegs[0], pegs[1])
		}
		if len(pegs[2]) != tc.n {
			t.Fatalf("peg 2 disk count = %d, want %d", len(pegs[2]), tc.n)
		}
	}
}

func TestHanoiTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("source line validation failed: %v", err)
	}

	// Verify all steps produce valid renderers.HanoiState
	for i, step := range steps {
		state, ok := step.State.(renderers.HanoiTowerLessonState)
		if !ok {
			t.Fatalf("step %d state is not renderers.HanoiState: %T", i, step.State)
		}
		totalDisks := len(state.Pegs[0]) + len(state.Pegs[1]) + len(state.Pegs[2])
		if totalDisks != 3 {
			t.Fatalf("step %d disk count mismatch: got %d, want 3", i, totalDisks)
		}
	}

	// Verify metrics tracking
	lastStep := steps[len(steps)-1]
	if lastStep.Metrics.Operations == 0 {
		t.Fatal("expected operations > 0 in final step metrics")
	}
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be ActionDone, got %v", lastStep.Action)
	}
}

func TestHanoiCustomInput(t *testing.T) {
	lesson := Lesson{}

	// Test n=4
	input4, _ := json.Marshal(4)
	steps4, err := lesson.Build(input4)
	if err != nil {
		t.Fatalf("Build(4) failed: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps4); err != nil {
		t.Fatalf("source line validation failed for n=4: %v", err)
	}
	lastState4 := steps4[len(steps4)-1].State.(renderers.HanoiTowerLessonState)
	if len(lastState4.Pegs[2]) != 4 {
		t.Fatalf("peg 2 disks = %d, want 4", len(lastState4.Pegs[2]))
	}

	// Test object input {"disks": 5}
	inputObj := []byte(`{"disks": 5}`)
	steps5, err := lesson.Build(inputObj)
	if err != nil {
		t.Fatalf("Build failed for object input: %v", err)
	}
	lastState5 := steps5[len(steps5)-1].State.(renderers.HanoiTowerLessonState)
	if len(lastState5.Pegs[2]) != 5 {
		t.Fatalf("peg 2 disks = %d, want 5", len(lastState5.Pegs[2]))
	}

	// Explicit out-of-range lesson input must fail before trace allocation.
	inputInvalid := []byte(`{"disks": 10}`)
	_, err = lesson.Build(inputInvalid)
	if err == nil {
		t.Fatal("Build accepted invalid disk count")
	}
}

func TestHanoiMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "hanoitower" {
		t.Fatalf("expected meta.ID 'hanoitower', got %q", meta.ID)
	}
	if meta.Name != "Tower of Hanoi" {
		t.Fatalf("expected meta.Name 'Tower of Hanoi', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryRecursion {
		t.Fatalf("expected CategoryRecursion, got %q", meta.Category)
	}
	if meta.Renderer != "hanoi" {
		t.Fatalf("expected Renderer 'hanoi', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(2ⁿ)" {
		t.Fatalf("expected TimeComplexity 'O(2ⁿ)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Fatalf("expected SpaceComplexity 'O(n)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{
		"Mathematical Induction",
		"Call Stack",
		"Exponential Growth",
		"Recursive Subproblems",
	}
	for _, exp := range expectedConcepts {
		found := false
		for _, concept := range meta.Documentation.KeyConcepts {
			if concept == exp {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected KeyConcept %q missing in %v", exp, meta.Documentation.KeyConcepts)
		}
	}
}
