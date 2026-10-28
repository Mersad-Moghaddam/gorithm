package nqueens

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func isValidSolution(n int, queens []int) bool {
	if len(queens) != n {
		return false
	}
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if queens[i] == queens[j] {
				return false
			}
			diffCol := queens[i] - queens[j]
			if diffCol < 0 {
				diffCol = -diffCol
			}
			diffRow := i - j
			if diffRow < 0 {
				diffRow = -diffRow
			}
			if diffCol == diffRow {
				return false
			}
		}
	}
	return true
}

func TestSolve(t *testing.T) {
	tests := []struct {
		n             int
		wantSolutions int
	}{
		{n: 1, wantSolutions: 1},
		{n: 2, wantSolutions: 0},
		{n: 3, wantSolutions: 0},
		{n: 4, wantSolutions: 2},
		{n: 8, wantSolutions: 92},
	}

	for _, tt := range tests {
		t.Run(string(rune('0'+tt.n)), func(t *testing.T) {
			got := Solve(tt.n)
			if len(got) != tt.wantSolutions {
				t.Fatalf("Solve(%d) returned %d solutions, want %d", tt.n, len(got), tt.wantSolutions)
			}
			for _, sol := range got {
				if !isValidSolution(tt.n, sol) {
					t.Errorf("Solve(%d) produced invalid solution: %v", tt.n, sol)
				}
			}
		})
	}
}

func TestLessonMeta(t *testing.T) {
	l := Lesson{}
	meta := l.Meta()

	if meta.ID != "nqueens" {
		t.Errorf("expected ID 'nqueens', got %q", meta.ID)
	}
	if meta.Name != "N-Queens Puzzle" {
		t.Errorf("expected Name 'N-Queens Puzzle', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryBacktracking {
		t.Errorf("expected Category %q, got %q", catalog.CategoryBacktracking, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(N!)" {
		t.Errorf("expected TimeComplexity 'O(N!)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(N)" {
		t.Errorf("expected SpaceComplexity 'O(N)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("expected Difficulty 'Hard', got %q", meta.Documentation.Difficulty)
	}

	wantConcepts := []string{
		"Backtracking",
		"Constraint Satisfaction",
		"Diagonal Conflict Checking",
		"State Space Tree Pruning",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, wantConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", wantConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	regLesson, ok := catalog.Get("nqueens")
	if !ok {
		t.Fatal("lesson 'nqueens' not registered in catalog")
	}
	if regLesson.Meta().ID != "nqueens" {
		t.Errorf("expected registered ID 'nqueens', got %q", regLesson.Meta().ID)
	}
}

func TestTraceValidation(t *testing.T) {
	l := Lesson{}
	steps, err := l.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(l, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}

	hasVars := false
	hasConflictHighlight := false
	for i, step := range steps {
		tbl, ok := step.State.(renderers.NQueensLessonState)
		if !ok {
			t.Fatalf("step %d state is not renderers.NQueensLessonState: %T", i, step.State)
		}
		if len(tbl.Cells) != 16 { // default n=4 -> 4x4
			t.Fatalf("step %d expected 16 cells, got %d", i, len(tbl.Cells))
		}
		if _, ok := step.Variables["row"]; ok {
			hasVars = true
		}
		if _, ok := step.Variables["col"]; ok {
			hasVars = true
		}
		for _, cell := range tbl.Cells {
			if cell.State == "active" {
				hasConflictHighlight = true
			}
		}
	}

	if !hasVars {
		t.Error("expected steps to contain row/col variables")
	}
	if !hasConflictHighlight {
		t.Error("expected conflict or candidate highlights in active state")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
}

func TestTraceCustomInput(t *testing.T) {
	l := Lesson{}
	input := json.RawMessage(`{"n": 4}`)
	steps, err := l.Build(input)
	if err != nil {
		t.Fatalf("Build with custom input failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps for custom input")
	}
	if err := catalog.ValidateSourceLineIDs(l, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs for custom input failed: %v", err)
	}
}
