package hanoitower

import (
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"strings"
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
		{n: 5, wantMoves: 31}, // 2^5 - 1
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

func TestHanoiRegistration(t *testing.T) {
	l, ok := catalog.Get("hanoitower")
	if !ok {
		t.Fatal("lesson 'hanoitower' not registered in catalog")
	}
	if l.Meta().ID != "hanoitower" {
		t.Fatalf("expected meta.ID 'hanoitower', got %q", l.Meta().ID)
	}
}

func TestHanoiInputValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []struct {
		name  string
		input json.RawMessage
	}{
		{"null", []byte("null")},
		{"whitespace only", []byte("   ")},
		{"not json", []byte("not json")},
		{"malformed json object", []byte("{disks: 3}")},
		{"json array", []byte("[1, 2, 3]")},
		{"json boolean", []byte("true")},
		{"json string", []byte(`"three"`)},
		{"float number", []byte("3.5")},
		{"zero disks", []byte("0")},
		{"negative disks", []byte("-1")},
		{"too many disks (8)", []byte("8")},
		{"too many disks (100)", []byte("100")},
		{"object missing known fields", []byte(`{"unknown": 4}`)},
		{"object disks zero", []byte(`{"disks": 0}`)},
		{"object disks negative", []byte(`{"disks": -2}`)},
		{"object disks too large", []byte(`{"disks": 8}`)},
		{"object n negative", []byte(`{"n": -1}`)},
		{"object n too large", []byte(`{"n": 9}`)},
		{"object count zero", []byte(`{"count": 0}`)},
		{"object count too large", []byte(`{"count": 8}`)},
		{"object disks null", []byte(`{"disks": null}`)},
		{"trailing data after integer", []byte("3 extra")},
		{"trailing data after object", []byte(`{"disks": 3} extra`)},
	}

	for _, tc := range invalidInputs {
		t.Run("invalid_"+tc.name, func(t *testing.T) {
			if _, err := lesson.Build(tc.input); err == nil {
				t.Fatalf("expected error for input %q, got nil", string(tc.input))
			}
		})
	}

	validInputs := []struct {
		name      string
		input     json.RawMessage
		wantDisks int
	}{
		{"nil input (default 3)", nil, 3},
		{"empty input (default 3)", []byte(""), 3},
		{"direct int 1", []byte("1"), 1},
		{"direct int 2", []byte("2"), 2},
		{"direct int 3", []byte("3"), 3},
		{"direct int 4", []byte("4"), 4},
		{"direct int 5", []byte("5"), 5},
		{"direct int 6", []byte("6"), 6},
		{"direct int 7", []byte("7"), 7},
		{"object disks 4", []byte(`{"disks": 4}`), 4},
		{"object n 5", []byte(`{"n": 5}`), 5},
		{"object count 2", []byte(`{"count": 2}`), 2},
	}

	for _, tc := range validInputs {
		t.Run("valid_"+tc.name, func(t *testing.T) {
			steps, err := lesson.Build(tc.input)
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", string(tc.input), err)
			}
			if len(steps) == 0 {
				t.Fatal("expected non-empty steps")
			}
			lastState := steps[len(steps)-1].State.(renderers.HanoiState)
			if len(lastState.Pegs[2]) != tc.wantDisks {
				t.Fatalf("expected %d disks on peg 2 at completion, got %d", tc.wantDisks, len(lastState.Pegs[2]))
			}
		})
	}
}

func TestHanoiTraceAndInvariant(t *testing.T) {
	lesson := Lesson{}

	for n := 1; n <= 5; n++ {
		t.Run(fmt.Sprintf("disks_%d", n), func(t *testing.T) {
			steps := Trace(n)
			if len(steps) == 0 {
				t.Fatalf("Trace(%d) returned 0 steps", n)
			}

			// 1. Source Line IDs validation
			if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
				t.Fatalf("catalog.ValidateSourceLineIDs failed for n=%d: %v", n, err)
			}

			expectedMoves := (1 << n) - 1
			moveStepCount := 0

			// 2. Initial state check
			firstState, ok := steps[0].State.(renderers.HanoiState)
			if !ok {
				t.Fatalf("initial step state is not renderers.HanoiState: %T", steps[0].State)
			}
			if len(firstState.Pegs[0]) != n || len(firstState.Pegs[1]) != 0 || len(firstState.Pegs[2]) != 0 {
				t.Fatalf("initial pegs wrong: %v", firstState.Pegs)
			}

			// Verify initial peg 0 has disks in strictly decreasing order from bottom to top
			for i := 0; i < n; i++ {
				expectedDisk := n - i
				if firstState.Pegs[0][i] != expectedDisk {
					t.Fatalf("initial peg 0 disk at %d = %d, want %d", i, firstState.Pegs[0][i], expectedDisk)
				}
			}

			var sawStep1, sawStep2, sawStep3 bool

			// 3. Invariant check on EVERY single step snapshot
			for stepIdx, step := range steps {
				state, ok := step.State.(renderers.HanoiState)
				if !ok {
					t.Fatalf("step %d state is not renderers.HanoiState: %T", stepIdx, step.State)
				}

				// Check total disk count
				totalDisks := len(state.Pegs[0]) + len(state.Pegs[1]) + len(state.Pegs[2])
				if totalDisks != n {
					t.Fatalf("step %d total disk count = %d, want %d", stepIdx, totalDisks, n)
				}

				// Check that disk set across all pegs is exactly {1..n}
				seenDisks := make(map[int]bool, n)
				for p := 0; p < 3; p++ {
					for _, d := range state.Pegs[p] {
						if d < 1 || d > n {
							t.Fatalf("step %d: invalid disk ID %d", stepIdx, d)
						}
						if seenDisks[d] {
							t.Fatalf("step %d: duplicate disk ID %d", stepIdx, d)
						}
						seenDisks[d] = true
					}
				}
				if len(seenDisks) != n {
					t.Fatalf("step %d: expected %d distinct disks, saw %d", stepIdx, n, len(seenDisks))
				}

				// Verify Peg Invariant: no larger disk ever on top of a smaller disk
				// Bottom to top must be strictly decreasing in size
				for p := 0; p < 3; p++ {
					peg := state.Pegs[p]
					for j := 0; j < len(peg)-1; j++ {
						if peg[j] <= peg[j+1] {
							t.Fatalf("step %d peg %d invariant violation: disk %d under smaller/equal disk %d (peg: %v)",
								stepIdx, p, peg[j], peg[j+1], peg)
						}
					}
				}

				// Check SVG rendering
				svg := renderers.RenderHanoiSVG(step, 800, 400)
				for d := 1; d <= n; d++ {
					expectedKey := fmt.Sprintf(`data-key="disk-%d"`, d)
					if strings.Count(svg, expectedKey) != 1 {
						t.Fatalf("step %d: SVG missing or duplicate %q", stepIdx, expectedKey)
					}
				}

				// Check recursive subproblems narration
				if step.LineID == "L5" && strings.Contains(step.Explanation, "Step 1") {
					sawStep1 = true
				}
				if step.LineID == "L6" {
					moveStepCount++
					if strings.Contains(step.Explanation, "Step 2") &&
						strings.Contains(step.Explanation, "leaves Peg") &&
						strings.Contains(step.Explanation, "lands on Peg") {
						sawStep2 = true
					}
				}
				if step.LineID == "L7" && strings.Contains(step.Explanation, "Step 3") {
					sawStep3 = true
				}
			}

			// Verify move count equals 2^n - 1
			if moveStepCount != expectedMoves {
				t.Fatalf("n=%d: counted %d move steps (L6), want %d (2^n - 1)", n, moveStepCount, expectedMoves)
			}

			lastStep := steps[len(steps)-1]
			if lastStep.Metrics.Operations != expectedMoves {
				t.Fatalf("n=%d: final metrics operations = %d, want %d (2^n - 1)", n, lastStep.Metrics.Operations, expectedMoves)
			}
			if lastStep.Action != trace.ActionDone {
				t.Fatalf("expected last action to be ActionDone, got %v", lastStep.Action)
			}

			lastState := lastStep.State.(renderers.HanoiState)
			if len(lastState.Pegs[0]) != 0 || len(lastState.Pegs[1]) != 0 || len(lastState.Pegs[2]) != n {
				t.Fatalf("final pegs wrong: %v, want all %d disks on peg 2", lastState.Pegs, n)
			}

			if n > 1 {
				if !sawStep1 || !sawStep2 || !sawStep3 {
					t.Fatalf("n=%d: missing recursive subproblem explanations (step1=%v, step2=%v, step3=%v)",
						n, sawStep1, sawStep2, sawStep3)
				}
			}
		})
	}
}

func TestHanoiSnapshotImmutability(t *testing.T) {
	n := 4
	steps := Trace(n)

	// Capture deep copies of all initial snapshots
	type pegCopy [3][]int
	savedPegs := make([]pegCopy, len(steps))
	for i, step := range steps {
		s := step.State.(renderers.HanoiState)
		for p := 0; p < 3; p++ {
			savedPegs[i][p] = slices.Clone(s.Pegs[p])
		}
	}

	// Mutate every slice of every step's state directly
	for _, step := range steps {
		s := step.State.(renderers.HanoiState)
		for p := 0; p < 3; p++ {
			for j := range s.Pegs[p] {
				s.Pegs[p][j] = -9999
			}
			s.Pegs[p] = append(s.Pegs[p], 12345)
		}
	}

	// Run trace again to compare against original savedPegs
	freshSteps := Trace(n)
	for i, fresh := range freshSteps {
		freshState := fresh.State.(renderers.HanoiState)
		for p := 0; p < 3; p++ {
			if !slices.Equal(freshState.Pegs[p], savedPegs[i][p]) {
				t.Fatalf("step %d peg %d: original snapshot was mutated! got %v, want %v",
					i, p, freshState.Pegs[p], savedPegs[i][p])
			}
		}
	}
}
