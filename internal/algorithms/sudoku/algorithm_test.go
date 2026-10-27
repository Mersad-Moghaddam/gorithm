package sudoku

import (
	"encoding/json"
	"reflect"
	"strings"
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

func TestConflictDetection(t *testing.T) {
	var board [9][9]int
	board[0][5] = 7 // row 0 has 7 at col 5
	board[5][0] = 6 // col 0 has 6 at row 5
	board[1][1] = 8 // box 0 has 8 at (1, 1)

	// Row conflict: placing 7 at (0, 2)
	conf, hasConf := FindConflict(board, 0, 2, 7)
	if !hasConf || conf == nil || conf.Invariant != "row" || conf.Row != 0 || conf.Col != 5 {
		t.Fatalf("expected row conflict at (0,5), got %+v", conf)
	}

	// Column conflict: placing 6 at (2, 0)
	conf, hasConf = FindConflict(board, 2, 0, 6)
	if !hasConf || conf == nil || conf.Invariant != "column" || conf.Row != 5 || conf.Col != 0 {
		t.Fatalf("expected column conflict at (5,0), got %+v", conf)
	}

	// Box conflict: placing 8 at (0, 2)
	conf, hasConf = FindConflict(board, 0, 2, 8)
	if !hasConf || conf == nil || conf.Invariant != "box" || conf.Row != 1 || conf.Col != 1 {
		t.Fatalf("expected box conflict at (1,1), got %+v", conf)
	}

	// Valid placement: placing 1 at (0, 2)
	conf, hasConf = FindConflict(board, 0, 2, 1)
	if hasConf || conf != nil {
		t.Fatalf("expected no conflict for digit 1 at (0,2), got %+v", conf)
	}

	// Invalid digits outside 1..9
	if _, hasConf := FindConflict(board, 0, 2, 0); hasConf {
		t.Error("expected FindConflict(0) to return false")
	}
	if _, hasConf := FindConflict(board, 0, 2, 10); hasConf {
		t.Error("expected FindConflict(10) to return false")
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

	t.Run("solves canonical board", func(t *testing.T) {
		solution, solved := Solve(CanonicalBoard)
		if !solved {
			t.Fatal("expected CanonicalBoard to be solved")
		}
		if !isBoardCompleteAndValid(solution) {
			t.Fatal("canonical board solution does not satisfy sudoku constraints")
		}
		for r := 0; r < 9; r++ {
			for c := 0; c < 9; c++ {
				if CanonicalBoard[r][c] != 0 && CanonicalBoard[r][c] != solution[r][c] {
					t.Fatalf("canonical clue at (%d,%d) modified from %d to %d", r, c, CanonicalBoard[r][c], solution[r][c])
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

func TestSnapshotImmutability(t *testing.T) {
	steps := Trace(CanonicalBoard)
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	state0, ok := steps[0].State.(renderers.TableState)
	if !ok {
		t.Fatalf("step 0 state is not TableState: %T", steps[0].State)
	}
	state1, ok := steps[1].State.(renderers.TableState)
	if !ok {
		t.Fatalf("step 1 state is not TableState: %T", steps[1].State)
	}

	// Mutating step 0 cells must not mutate step 1
	originalVal := state1.Cells[0].Value
	originalState := state1.Cells[0].State
	state0.Cells[0].Value = "MUTATED"
	state0.Cells[0].State = "MUTATED"

	if state1.Cells[0].Value != originalVal || state1.Cells[0].State != originalState {
		t.Fatal("step 1 state was mutated when step 0 cells were modified: snapshot immutability violated")
	}

	// Test CloneTableState
	cloned := CloneTableState(state1)
	cloned.Cells[0].Value = "CLONE_MUTATED"
	if state1.Cells[0].Value == "CLONE_MUTATED" {
		t.Fatal("CloneTableState did not deep copy Cells")
	}
}

func TestInputValidation(t *testing.T) {
	lesson := Lesson{}

	t.Run("nil input uses canonical board", func(t *testing.T) {
		steps, err := lesson.Build(nil)
		if err != nil {
			t.Fatalf("expected success on nil input, got: %v", err)
		}
		if len(steps) == 0 {
			t.Fatal("expected non-empty steps")
		}
	})

	t.Run("empty input uses canonical board", func(t *testing.T) {
		steps, err := lesson.Build([]byte{})
		if err != nil {
			t.Fatalf("expected success on empty input, got: %v", err)
		}
		if len(steps) == 0 {
			t.Fatal("expected non-empty steps")
		}
	})

	t.Run("reject null input", func(t *testing.T) {
		_, err := lesson.Build([]byte("null"))
		if err == nil {
			t.Fatal("expected error on null input")
		}
	})

	t.Run("reject null board field", func(t *testing.T) {
		_, err := lesson.Build([]byte(`{"board": null}`))
		if err == nil {
			t.Fatal("expected error on null board field")
		}
	})

	t.Run("reject malformed JSON", func(t *testing.T) {
		_, err := lesson.Build([]byte(`{"board": [[5,3,0`))
		if err == nil {
			t.Fatal("expected error on malformed JSON")
		}
		_, err = lesson.Build([]byte(`not a json`))
		if err == nil {
			t.Fatal("expected error on non-json input")
		}
	})

	t.Run("reject invalid row count", func(t *testing.T) {
		shortBoard := `{"board": [[1,2,3,4,5,6,7,8,9]]}`
		if _, err := lesson.Build([]byte(shortBoard)); err == nil {
			t.Fatal("expected error on board with 1 row")
		}
	})

	t.Run("reject invalid column count", func(t *testing.T) {
		board := sampleSolvableBoard()
		var rawRows [][]int
		for r := 0; r < 9; r++ {
			if r == 0 {
				rawRows = append(rawRows, []int{1, 2, 3}) // short col
			} else {
				row := make([]int, 9)
				copy(row, board[r][:])
				rawRows = append(rawRows, row)
			}
		}
		raw, _ := json.Marshal(map[string]any{"board": rawRows})
		if _, err := lesson.Build(raw); err == nil {
			t.Fatal("expected error on board with short column count")
		}
	})

	t.Run("reject negative numbers", func(t *testing.T) {
		board := sampleSolvableBoard()
		board[0][2] = -1
		raw, _ := json.Marshal(map[string]any{"board": board})
		if _, err := lesson.Build(raw); err == nil {
			t.Fatal("expected error on board with negative number")
		}
	})

	t.Run("reject numbers above 9", func(t *testing.T) {
		board := sampleSolvableBoard()
		board[0][2] = 10
		raw, _ := json.Marshal(map[string]any{"board": board})
		if _, err := lesson.Build(raw); err == nil {
			t.Fatal("expected error on board with number > 9")
		}
	})

	t.Run("reject null cell", func(t *testing.T) {
		raw := []byte(`{"board": [[5,null,0,0,7,8,9,1,2],[6,7,2,1,9,5,3,4,8],[0,9,8,3,4,2,5,6,7],[8,5,9,7,6,1,4,2,3],[4,2,6,8,0,3,7,9,1],[7,1,3,9,2,4,8,5,6],[9,6,1,5,3,7,0,8,4],[2,8,7,4,1,9,6,3,5],[3,4,5,2,8,6,1,0,0]]}`)
		if _, err := lesson.Build(raw); err == nil {
			t.Fatal("expected error on null cell")
		}
	})

	t.Run("accept direct 2D array", func(t *testing.T) {
		raw, _ := json.Marshal(CanonicalBoard)
		steps, err := lesson.Build(raw)
		if err != nil {
			t.Fatalf("expected direct 2D array to succeed, got: %v", err)
		}
		if len(steps) == 0 {
			t.Fatal("expected non-empty steps")
		}
	})
}

func TestTraceVisualizationRequirements(t *testing.T) {
	steps := Trace(CanonicalBoard)

	var (
		foundEmptyHighlight    bool
		foundConflictHighlight bool
		foundValidPlacement    bool
		foundSolvedStep        bool
	)

	for _, step := range steps {
		ts, ok := step.State.(renderers.TableState)
		if !ok {
			continue
		}

		// Empty cell finding: highlight current empty cell (row, col) (state "active")
		if step.LineID == "L2" {
			r := step.Variables["row"].(int)
			c := step.Variables["col"].(int)
			for _, cell := range ts.Cells {
				if cell.Row == r && cell.Col == c && cell.State == "active" {
					foundEmptyHighlight = true
					break
				}
			}
		}

		// Conflict highlight: conflicting cell has state "source", candidate cell has state "active"
		if step.LineID == "L7" && step.Action == trace.ActionHighlight {
			confRow, hasR := step.Variables["conflictRow"].(int)
			confCol, hasC := step.Variables["conflictCol"].(int)
			row, hasRow := step.Variables["row"].(int)
			col, hasCol := step.Variables["col"].(int)
			if hasR && hasC && hasRow && hasCol {
				var srcFound, actFound bool
				for _, cell := range ts.Cells {
					if cell.Row == confRow && cell.Col == confCol && cell.State == "source" {
						srcFound = true
					}
					if cell.Row == row && cell.Col == col && cell.State == "active" {
						actFound = true
					}
				}
				if srcFound && actFound && strings.Contains(ts.Caption, "violates") {
					foundConflictHighlight = true
				}
			}
		}

		// Valid placement: place digit (state "computed")
		if step.LineID == "L8" && step.Action == trace.ActionSet {
			row := step.Variables["row"].(int)
			col := step.Variables["col"].(int)
			for _, cell := range ts.Cells {
				if cell.Row == row && cell.Col == col && cell.State == "computed" {
					foundValidPlacement = true
					break
				}
			}
		}

		// Solved: all cells filled correctly
		if step.Action == trace.ActionDone && strings.Contains(ts.Caption, "puzzle solved") {
			allFilled := true
			for _, cell := range ts.Cells {
				if cell.Value == " " || cell.Value == "" {
					allFilled = false
					break
				}
			}
			if allFilled {
				foundSolvedStep = true
			}
		}
	}

	if !foundEmptyHighlight {
		t.Error("trace missing empty cell finding step with cell state 'active'")
	}
	if !foundConflictHighlight {
		t.Error("trace missing conflict step highlighting conflicting cell with 'source' and explaining conflict")
	}
	if !foundValidPlacement {
		t.Error("trace missing valid placement step with cell state 'computed'")
	}
	if !foundSolvedStep {
		t.Error("trace missing solved step with all cells filled")
	}

	// Backtracking test: puzzle that triggers dead ends and backtracking
	sol, _ := Solve(sampleSolvableBoard())
	backtrackPuzzle := sol
	backtrackPuzzle[0] = sampleSolvableBoard()[0]
	backtrackPuzzle[1] = sampleSolvableBoard()[1]

	deSteps := Trace(backtrackPuzzle)
	var foundBacktrack bool
	for _, step := range deSteps {
		ts, ok := step.State.(renderers.TableState)
		if !ok {
			continue
		}
		if strings.Contains(ts.Caption, "Dead end: no valid digit for cell") {
			r := step.Variables["row"].(int)
			c := step.Variables["col"].(int)
			for _, cell := range ts.Cells {
				if cell.Row == r && cell.Col == c && cell.State == "normal" && cell.Value == " " {
					foundBacktrack = true
					break
				}
			}
		}
	}
	if !foundBacktrack {
		t.Error("trace missing backtrack step with caption 'Dead end: no valid digit for cell (row, col); backtrack' and cell state 'normal' with value ' '")
	}
}
