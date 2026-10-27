package sudoku

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func sampleSolvableBoard() [9][9]int {
	return [9][9]int{
		{5, 3, 0, 0, 7, 0, 0, 0, 0},
		{6, 0, 0, 1, 9, 5, 0, 0, 0},
		{0, 9, 8, 0, 0, 0, 0, 6, 0},
		{8, 0, 0, 0, 6, 0, 0, 0, 3},
		{4, 0, 0, 8, 0, 3, 0, 0, 1},
		{7, 0, 0, 0, 2, 0, 0, 0, 6},
		{0, 6, 0, 0, 0, 0, 2, 8, 0},
		{0, 0, 0, 4, 1, 9, 0, 0, 5},
		{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
}

func isBoardCompleteAndValid(board [9][9]int) bool {
	for r := 0; r < 9; r++ {
		rowSeen := [10]bool{}
		for c := 0; c < 9; c++ {
			v := board[r][c]
			if v < 1 || v > 9 || rowSeen[v] {
				return false
			}
			rowSeen[v] = true
		}
	}

	for c := 0; c < 9; c++ {
		colSeen := [10]bool{}
		for r := 0; r < 9; r++ {
			v := board[r][c]
			if v < 1 || v > 9 || colSeen[v] {
				return false
			}
			colSeen[v] = true
		}
	}

	for br := 0; br < 9; br += 3 {
		for bc := 0; bc < 9; bc += 3 {
			boxSeen := [10]bool{}
			for r := 0; r < 3; r++ {
				for c := 0; c < 3; c++ {
					v := board[br+r][bc+c]
					if v < 1 || v > 9 || boxSeen[v] {
						return false
					}
					boxSeen[v] = true
				}
			}
		}
	}
	return true
}

func TestGridValidation(t *testing.T) {
	board := sampleSolvableBoard()

	// Row conflict: 5 is already in row 0
	if IsValid(board, 0, 2, 5) {
		t.Error("expected 5 to be invalid in row 0")
	}

	// Col conflict: 6 is already in col 0
	if IsValid(board, 2, 0, 6) {
		t.Error("expected 6 to be invalid in col 0")
	}

	// Box conflict: 9 is in top-left 3x3 box
	if IsValid(board, 0, 2, 9) {
		t.Error("expected 9 to be invalid in top-left box")
	}

	// Valid placement: 1 at (0, 2) is valid for row 0, col 2, and top-left box
	if !IsValid(board, 0, 2, 1) && !IsValid(board, 0, 2, 2) && !IsValid(board, 0, 2, 4) {
		t.Error("expected at least one valid candidate at empty cell (0,2)")
	}
}

func TestSudokuSolverCorrectness(t *testing.T) {
	t.Run("solves valid puzzle", func(t *testing.T) {
		initial := sampleSolvableBoard()
		solution, solved := Solve(initial)
		if !solved {
			t.Fatal("expected puzzle to be solved")
		}

		if !isBoardCompleteAndValid(solution) {
			t.Fatal("solution does not satisfy sudoku constraints")
		}

		// Verify initial clues preserved
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if initial[r][c] != 0 && initial[r][c] != solution[r][c] {
					t.Fatalf("clue at (%d,%d) modified from %d to %d", r, c, initial[r][c], solution[r][c])
				}
			}
		}
	})

	t.Run("unsolvable puzzle returns false", func(t *testing.T) {
		var unsolvable [9][9]int
		// 9 in row 0 col 0, and another 9 in row 0 col 1
		unsolvable[0][0] = 9
		unsolvable[0][1] = 9
		_, solved := Solve(unsolvable)
		if solved {
			t.Fatal("expected unsolvable puzzle to fail")
		}
	})
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "sudoku" {
		t.Errorf("expected ID 'sudoku', got %q", meta.ID)
	}
	if meta.Name != "Sudoku Solver" {
		t.Errorf("expected Name 'Sudoku Solver', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryBacktracking {
		t.Errorf("expected Category %q, got %q", catalog.CategoryBacktracking, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(9^(empty_cells))" {
		t.Errorf("expected TimeComplexity 'O(9^(empty_cells))', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("expected Difficulty 'Hard', got %q", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{"Exact Cover", "Row/Column/Box Invariant", "Recursive Backtracking", "Constraint Propagation"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("sudoku")
	if !ok {
		t.Fatal("lesson 'sudoku' not registered in catalog")
	}
	if l.Meta().ID != "sudoku" {
		t.Errorf("expected registered ID 'sudoku', got %q", l.Meta().ID)
	}
}

func TestSudokuTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}

	var foundCandidateVar bool
	for i, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
		}
		if len(tableState.RowHeaders) != 9 || len(tableState.ColHeaders) != 9 {
			t.Fatalf("step %d table dimensions mismatch: %d rows, %d cols", i, len(tableState.RowHeaders), len(tableState.ColHeaders))
		}
		if len(tableState.Cells) != 81 {
			t.Fatalf("step %d table has %d cells, want 81", i, len(tableState.Cells))
		}
		if _, hasCand := step.Variables["candidate"]; hasCand {
			foundCandidateVar = true
		}
	}

	if !foundCandidateVar {
		t.Error("expected at least one step to contain 'candidate' in Variables")
	}
}

func TestLessonBuildCustomInput(t *testing.T) {
	lesson := Lesson{}
	raw, err := json.Marshal(map[string]any{
		"board": sampleSolvableBoard(),
	})
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	steps, err := lesson.Build(raw)
	if err != nil {
		t.Fatalf("Build with custom board failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("Build returned empty steps")
	}
}
