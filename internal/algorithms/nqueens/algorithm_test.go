package nqueens

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
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
		t.Run(fmt.Sprintf("n=%d", tt.n), func(t *testing.T) {
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
		tbl, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
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

func TestTraceTableState(t *testing.T) {
	steps := Trace(4)
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}

	for i, step := range steps {
		tbl, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d: state is not TableState, got %T", i, step.State)
		}

		if len(tbl.RowHeaders) != 4 || len(tbl.ColHeaders) != 4 {
			t.Fatalf("step %d: expected 4x4 headers, got %d rows and %d cols", i, len(tbl.RowHeaders), len(tbl.ColHeaders))
		}
		for r, rh := range tbl.RowHeaders {
			if rh != fmt.Sprintf("Row %d", r) {
				t.Errorf("step %d: expected RowHeader %q, got %q", i, fmt.Sprintf("Row %d", r), rh)
			}
		}
		for c, ch := range tbl.ColHeaders {
			if ch != fmt.Sprintf("Col %d", c) {
				t.Errorf("step %d: expected ColHeader %q, got %q", i, fmt.Sprintf("Col %d", c), ch)
			}
		}

		if len(tbl.Cells) != 16 {
			t.Fatalf("step %d: expected 16 cells, got %d", i, len(tbl.Cells))
		}
		if tbl.Caption == "" {
			t.Errorf("step %d: expected non-empty caption in TableState", i)
		}

		for _, cell := range tbl.Cells {
			if cell.Row < 0 || cell.Row >= 4 || cell.Col < 0 || cell.Col >= 4 {
				t.Errorf("step %d: cell out of bounds: (%d, %d)", i, cell.Row, cell.Col)
			}
			if cell.Value != "👑" && cell.Value != "." {
				t.Errorf("step %d: invalid cell value %q", i, cell.Value)
			}
			if cell.Value == "👑" && cell.State != "computed" && cell.State != "active" {
				t.Errorf("step %d: queen cell at (%d, %d) has invalid state %q", i, cell.Row, cell.Col, cell.State)
			}
			if cell.Value == "." && cell.State != "normal" {
				t.Errorf("step %d: empty cell at (%d, %d) has state %q, want normal", i, cell.Row, cell.Col, cell.State)
			}
		}
	}
}

func TestSnapshotImmutability(t *testing.T) {
	steps := Trace(4)
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	// Verify slice pointers are unique across steps
	for i := 0; i < len(steps)-1; i++ {
		tblA := steps[i].State.(renderers.TableState)
		tblB := steps[i+1].State.(renderers.TableState)

		if &tblA.Cells[0] == &tblB.Cells[0] {
			t.Fatalf("step %d and %d share Cells backing array (aliased snapshot)", i, i+1)
		}
		if &tblA.RowHeaders[0] == &tblB.RowHeaders[0] {
			t.Fatalf("step %d and %d share RowHeaders backing array", i, i+1)
		}
		if &tblA.ColHeaders[0] == &tblB.ColHeaders[0] {
			t.Fatalf("step %d and %d share ColHeaders backing array", i, i+1)
		}
	}

	// Mutate step 0 and ensure step 1 remains unmodified
	tbl0 := steps[0].State.(renderers.TableState)
	origVal := steps[1].State.(renderers.TableState).Cells[0].Value
	origHeader := steps[1].State.(renderers.TableState).RowHeaders[0]

	tbl0.Cells[0].Value = "MUTATED"
	tbl0.RowHeaders[0] = "MUTATED"

	if steps[1].State.(renderers.TableState).Cells[0].Value != origVal {
		t.Errorf("mutating step 0 affected step 1 Cells: got %q, want %q",
			steps[1].State.(renderers.TableState).Cells[0].Value, origVal)
	}
	if steps[1].State.(renderers.TableState).RowHeaders[0] != origHeader {
		t.Errorf("mutating step 0 affected step 1 RowHeaders: got %q, want %q",
			steps[1].State.(renderers.TableState).RowHeaders[0], origHeader)
	}
}

func TestN4Solution(t *testing.T) {
	l := Lesson{}
	steps, err := l.Build(json.RawMessage(`{"n": 4}`))
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if err := catalog.ValidateSourceLineIDs(l, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}

	solutionSteps := 0
	for i, step := range steps {
		if step.LineID == "L7" {
			solutionSteps++
			if step.Action != trace.ActionDone {
				t.Errorf("step %d (L7): action is %s, want %s", i, step.Action, trace.ActionDone)
			}
			if !strings.Contains(step.Explanation, "Valid 4x4 board configuration") {
				t.Errorf("step %d (L7): explanation does not celebrate valid board: %q", i, step.Explanation)
			}

			tbl := step.State.(renderers.TableState)
			if !strings.Contains(tbl.Caption, "Valid 4x4 board configuration") {
				t.Errorf("step %d (L7): caption does not celebrate valid board: %q", i, tbl.Caption)
			}

			// Verify exactly 4 queens and valid non-attacking configuration
			queenCols := make([]int, 4)
			for r := range queenCols {
				queenCols[r] = -1
			}
			queensFound := 0
			for _, cell := range tbl.Cells {
				if cell.Value == "👑" {
					queensFound++
					queenCols[cell.Row] = cell.Col
				}
			}
			if queensFound != 4 {
				t.Errorf("step %d (L7): expected 4 queens on board, found %d", i, queensFound)
			}
			if !isValidSolution(4, queenCols) {
				t.Errorf("step %d (L7): board is not a valid non-attacking N-Queens solution: %v", i, queenCols)
			}
		}
	}

	if solutionSteps != 2 {
		t.Fatalf("expected 2 solution steps for N=4, got %d", solutionSteps)
	}
}

func TestThreatDetectionAndConflicts(t *testing.T) {
	steps := Trace(4)

	hasColConflict := false
	hasDiagConflict := false

	for i, step := range steps {
		if step.LineID == "L21" {
			if step.Action != trace.ActionCompare {
				t.Errorf("step %d (L21): expected action %s, got %s", i, trace.ActionCompare, step.Action)
			}

			isCol := strings.Contains(step.Explanation, "shares column with queen at")
			isDiag := strings.Contains(step.Explanation, "shares diagonal with queen at")
			if !isCol && !isDiag {
				t.Errorf("step %d (L21): explanation missing threat details: %q", i, step.Explanation)
			}
			if isCol {
				hasColConflict = true
			}
			if isDiag {
				hasDiagConflict = true
			}

			tbl, ok := step.State.(renderers.TableState)
			if !ok {
				t.Fatalf("step %d: state is not TableState", i)
			}
			if tbl.Caption != step.Explanation {
				t.Errorf("step %d: caption %q != explanation %q", i, tbl.Caption, step.Explanation)
			}

			// Verify both conflicting queen and candidate are highlighted in active state
			activeCount := 0
			for _, cell := range tbl.Cells {
				if cell.State == "active" {
					activeCount++
				}
			}
			if activeCount < 2 {
				t.Errorf("step %d (L21): expected at least 2 active cells (conflict highlight), got %d", i, activeCount)
			}

			// Verify conflict variables
			if _, ok := step.Variables["conflictRow"]; !ok {
				t.Errorf("step %d (L21): missing conflictRow variable", i)
			}
			if _, ok := step.Variables["conflictCol"]; !ok {
				t.Errorf("step %d (L21): missing conflictCol variable", i)
			}
			if _, ok := step.Variables["threat"]; !ok {
				t.Errorf("step %d (L21): missing threat variable", i)
			}
		}
	}

	if !hasColConflict {
		t.Error("expected at least one column conflict step in N=4 trace")
	}
	if !hasDiagConflict {
		t.Error("expected at least one diagonal conflict step in N=4 trace")
	}
}

func TestBacktracking(t *testing.T) {
	steps := Trace(4)

	hasBacktrackCaption := false
	for i, step := range steps {
		if step.LineID == "L14" {
			if step.Action != trace.ActionPop {
				t.Errorf("step %d (L14): expected action %s, got %s", i, trace.ActionPop, step.Action)
			}

			tbl, ok := step.State.(renderers.TableState)
			if !ok {
				t.Fatalf("step %d: state is not TableState", i)
			}

			if strings.Contains(step.Explanation, "No valid placement in row") &&
				strings.Contains(step.Explanation, "backtrack to row") {
				hasBacktrackCaption = true
			}

			// Ensure queen at (row, col) has been removed
			row := step.Variables["row"].(int)
			col := step.Variables["col"].(int)
			for _, cell := range tbl.Cells {
				if cell.Row == row && cell.Col == col {
					if cell.Value != "." {
						t.Errorf("step %d: backtracked cell (%d, %d) still has value %q", i, row, col, cell.Value)
					}
					if cell.State != "normal" {
						t.Errorf("step %d: backtracked cell (%d, %d) has state %q, want normal", i, row, col, cell.State)
					}
				}
			}
		}
	}

	if !hasBacktrackCaption {
		t.Errorf("expected backtrack step with caption 'No valid placement in row %%d; backtrack to row %%d'")
	}
}

func TestInputValidation(t *testing.T) {
	l := Lesson{}

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		wantN   int
	}{
		{name: "null input", input: []byte("null"), wantErr: true},
		{name: "null n field", input: []byte(`{"n": null}`), wantErr: true},
		{name: "empty json object", input: []byte(`{}`), wantErr: true},
		{name: "n zero", input: []byte(`{"n": 0}`), wantErr: true},
		{name: "n negative", input: []byte(`{"n": -1}`), wantErr: true},
		{name: "n too large", input: []byte(`{"n": 9}`), wantErr: true},
		{name: "string n field", input: []byte(`{"n": "4"}`), wantErr: true},
		{name: "malformed json", input: []byte(`{broken json`), wantErr: true},
		{name: "direct zero", input: []byte(`0`), wantErr: true},
		{name: "direct negative", input: []byte(`-2`), wantErr: true},
		{name: "direct too large", input: []byte(`9`), wantErr: true},
		{name: "json array", input: []byte(`[4]`), wantErr: true},
		{name: "boolean", input: []byte(`true`), wantErr: true},
		{name: "valid object n=4", input: []byte(`{"n": 4}`), wantErr: false, wantN: 4},
		{name: "valid object n=1", input: []byte(`{"n": 1}`), wantErr: false, wantN: 1},
		{name: "valid object n=8", input: []byte(`{"n": 8}`), wantErr: false, wantN: 8},
		{name: "valid direct integer 4", input: []byte(`4`), wantErr: false, wantN: 4},
		{name: "valid direct integer 1", input: []byte(`1`), wantErr: false, wantN: 1},
		{name: "nil input defaults to 4", input: nil, wantErr: false, wantN: 4},
		{name: "empty bytes defaults to 4", input: []byte(""), wantErr: false, wantN: 4},
		{name: "whitespace defaults to 4", input: []byte("   "), wantErr: false, wantN: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, err := l.Build(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", string(tt.input))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", string(tt.input), err)
			}
			if len(steps) == 0 {
				t.Fatalf("expected non-empty steps for input %q", string(tt.input))
			}
			tbl, ok := steps[0].State.(renderers.TableState)
			if !ok {
				t.Fatalf("step 0 state is not TableState")
			}
			if len(tbl.RowHeaders) != tt.wantN {
				t.Errorf("expected %d rows, got %d", tt.wantN, len(tbl.RowHeaders))
			}
		})
	}
}
