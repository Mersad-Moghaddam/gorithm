package fibonacci

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestComputeFibonacciCorrectness(t *testing.T) {
	tests := []struct {
		name      string
		n         int
		wantVal   int
		wantTable []int
	}{
		{name: "negative input", n: -1, wantVal: 0, wantTable: []int{0}},
		{name: "n = 0", n: 0, wantVal: 0, wantTable: []int{0}},
		{name: "n = 1", n: 1, wantVal: 1, wantTable: []int{0, 1}},
		{name: "n = 2", n: 2, wantVal: 1, wantTable: []int{0, 1, 1}},
		{name: "n = 3", n: 3, wantVal: 2, wantTable: []int{0, 1, 1, 2}},
		{name: "n = 4", n: 4, wantVal: 3, wantTable: []int{0, 1, 1, 2, 3}},
		{name: "n = 5", n: 5, wantVal: 5, wantTable: []int{0, 1, 1, 2, 3, 5}},
		{name: "n = 6", n: 6, wantVal: 8, wantTable: []int{0, 1, 1, 2, 3, 5, 8}},
		{name: "n = 7", n: 7, wantVal: 13, wantTable: []int{0, 1, 1, 2, 3, 5, 8, 13}},
		{name: "n = 8", n: 8, wantVal: 21, wantTable: []int{0, 1, 1, 2, 3, 5, 8, 13, 21}},
		{name: "n = 10", n: 10, wantVal: 55, wantTable: []int{0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotTable := Compute(tt.n)
			if gotVal != tt.wantVal {
				t.Fatalf("Compute(%d) gotVal = %d, want %d", tt.n, gotVal, tt.wantVal)
			}
			if !reflect.DeepEqual(gotTable, tt.wantTable) {
				t.Fatalf("Compute(%d) gotTable = %v, want %v", tt.n, gotTable, tt.wantTable)
			}
		})
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "fibonacci" {
		t.Errorf("expected ID 'fibonacci', got %q", meta.ID)
	}
	if meta.Name != "Fibonacci (DP)" {
		t.Errorf("expected Name 'Fibonacci (DP)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryDP {
		t.Errorf("expected Category %q, got %q", catalog.CategoryDP, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n)" {
		t.Errorf("expected TimeComplexity 'O(n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Errorf("expected SpaceComplexity 'O(n)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Optimal Substructure", "Overlapping Subproblems", "Tabulation", "Bottom-Up DP"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("fibonacci")
	if !ok {
		t.Fatal("lesson 'fibonacci' not registered in catalog")
	}
	if l.Meta().ID != "fibonacci" {
		t.Errorf("expected registered ID 'fibonacci', got %q", l.Meta().ID)
	}
}

func TestFibonacciTraceValidation(t *testing.T) {
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
}

func TestFibonacciTraceHighlighting(t *testing.T) {
	steps := Trace(4)
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	var foundHighlight bool
	for _, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step state is not renderers.TableState: %T", step.State)
		}

		var sourceCount, activeCount int
		for _, cell := range tableState.Cells {
			if cell.State == "source" {
				sourceCount++
			}
			if cell.State == "active" {
				activeCount++
			}
		}

		if sourceCount == 2 && activeCount == 1 {
			foundHighlight = true
		}
	}

	if !foundHighlight {
		t.Fatal("expected at least one step highlighting 2 source cells and 1 active cell")
	}

	lesson := Lesson{}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("Trace(4) line validation failed: %v", err)
	}
}

func TestFibonacciTraceEdgeCases(t *testing.T) {
	lesson := Lesson{}

	for _, n := range []int{0, 1} {
		steps := Trace(n)
		if len(steps) == 0 {
			t.Fatalf("Trace(%d) produced 0 steps", n)
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("Trace(%d) line validation failed: %v", n, err)
		}
		last := steps[len(steps)-1]
		if last.Action != trace.ActionDone {
			t.Fatalf("Trace(%d) last action = %s, want %s", n, last.Action, trace.ActionDone)
		}
	}
}

func TestFibonacciBuildCustomInput(t *testing.T) {
	lesson := Lesson{}

	tests := []struct {
		name  string
		input []byte
		wantN int
	}{
		{name: "json object", input: []byte(`{"n": 5}`), wantN: 5},
		{name: "json integer", input: []byte(`6`), wantN: 6},
		{name: "empty input", input: nil, wantN: 8},
		{name: "n = 0", input: []byte(`{"n": 0}`), wantN: 0},
		{name: "n = 1", input: []byte(`{"n": 1}`), wantN: 1},
		{name: "n = 30", input: []byte(`{"n": 30}`), wantN: 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			steps, err := lesson.Build(json.RawMessage(tt.input))
			if err != nil {
				t.Fatalf("Build() returned error: %v", err)
			}
			if len(steps) == 0 {
				t.Fatal("Build() returned 0 steps")
			}
			last := steps[len(steps)-1]
			gotN, ok := last.Variables["n"].(int)
			if !ok || gotN != tt.wantN {
				t.Fatalf("expected variable n=%d in last step, got %v", tt.wantN, last.Variables["n"])
			}
		})
	}
}

func TestFibonacciBuildValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []struct {
		name  string
		input []byte
	}{
		{name: "null literal", input: []byte("null")},
		{name: "field n is null", input: []byte(`{"n": null}`)},
		{name: "missing n in object", input: []byte(`{"other": 10}`)},
		{name: "negative n in object", input: []byte(`{"n": -1}`)},
		{name: "negative direct int", input: []byte(`-5`)},
		{name: "n > 30 in object", input: []byte(`{"n": 31}`)},
		{name: "direct int > 30", input: []byte(`35`)},
		{name: "malformed JSON string", input: []byte(`"not-a-number"`)},
		{name: "malformed JSON array", input: []byte(`[1, 2]`)},
		{name: "broken JSON syntax", input: []byte(`{"n": `)},
	}

	for _, tt := range invalidInputs {
		t.Run(tt.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tt.input))
			if err == nil {
				t.Fatalf("Build(%s) expected error, got nil", string(tt.input))
			}
		})
	}
}

func TestFibonacciTableState_N8(t *testing.T) {
	lesson := Lesson{}
	steps := Trace(8)
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	_, dp := Compute(8)

	var foundBase0, foundBase1 bool
	highlightSteps := make(map[int]trace.Step)
	setSteps := make(map[int]trace.Step)

	for _, step := range steps {
		ts, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step state is not renderers.TableState: %T", step.State)
		}

		if !reflect.DeepEqual(ts.RowHeaders, []string{"F(i)"}) {
			t.Fatalf("expected RowHeaders [\"F(i)\"], got %v", ts.RowHeaders)
		}
		if len(ts.ColHeaders) != 9 {
			t.Fatalf("expected 9 ColHeaders, got %d", len(ts.ColHeaders))
		}
		if len(ts.Cells) != 9 {
			t.Fatalf("expected 9 Cells, got %d", len(ts.Cells))
		}

		if step.LineID == "L6" && ts.Cells[0].Value == "0" && ts.Cells[0].State == "computed" {
			foundBase0 = true
		}
		if step.LineID == "L7" && ts.Cells[1].Value == "1" && ts.Cells[1].State == "computed" {
			foundBase1 = true
		}
		if step.LineID == "L8" && step.Action == trace.ActionHighlight {
			if iVal, ok := step.Variables["i"].(int); ok {
				highlightSteps[iVal] = step
			}
		}
		if step.LineID == "L9" && step.Action == trace.ActionSet {
			if iVal, ok := step.Variables["i"].(int); ok {
				setSteps[iVal] = step
			}
		}
	}

	if !foundBase0 {
		t.Fatal("expected base case F(0)=0 marked computed at L6")
	}
	if !foundBase1 {
		t.Fatal("expected base case F(1)=1 marked computed at L7")
	}

	for i := 2; i <= 8; i++ {
		hlStep, ok := highlightSteps[i]
		if !ok {
			t.Fatalf("missing highlight step for i=%d", i)
		}
		hlState := hlStep.State.(renderers.TableState)

		if hlState.Cells[i-2].State != "source" {
			t.Fatalf("i=%d: expected cell %d to be 'source', got %q", i, i-2, hlState.Cells[i-2].State)
		}
		if hlState.Cells[i-1].State != "source" {
			t.Fatalf("i=%d: expected cell %d to be 'source', got %q", i, i-1, hlState.Cells[i-1].State)
		}
		if hlState.Cells[i].State != "active" {
			t.Fatalf("i=%d: expected cell %d to be 'active', got %q", i, i, hlState.Cells[i].State)
		}

		if len(hlState.Dependencies) != 2 {
			t.Fatalf("i=%d: expected 2 dependencies, got %d", i, len(hlState.Dependencies))
		}
		expectedDep0 := renderers.TableDependency{FromRow: 0, FromCol: i - 2, ToRow: 0, ToCol: i}
		expectedDep1 := renderers.TableDependency{FromRow: 0, FromCol: i - 1, ToRow: 0, ToCol: i}
		if hlState.Dependencies[0] != expectedDep0 || hlState.Dependencies[1] != expectedDep1 {
			t.Fatalf("i=%d: dependencies mismatch: got %v", i, hlState.Dependencies)
		}

		expectedCaption := fmt.Sprintf("F(%d) = F(%d) + F(%d) = %d + %d = %d", i, i-1, i-2, dp[i-1], dp[i-2], dp[i])
		if hlState.Caption != expectedCaption {
			t.Fatalf("i=%d: expected caption %q, got %q", i, expectedCaption, hlState.Caption)
		}

		setStep, ok := setSteps[i]
		if !ok {
			t.Fatalf("missing set step for i=%d", i)
		}
		setState := setStep.State.(renderers.TableState)
		if setState.Cells[i].State != "computed" {
			t.Fatalf("i=%d: expected cell %d to be 'computed' after set, got %q", i, i, setState.Cells[i].State)
		}
		if setState.Cells[i].Value != strconv.Itoa(dp[i]) {
			t.Fatalf("i=%d: expected cell %d value %d, got %q", i, i, dp[i], setState.Cells[i].Value)
		}
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action ActionDone, got %s", lastStep.Action)
	}
	doneState := lastStep.State.(renderers.TableState)
	if doneState.Cells[8].State != "active" {
		t.Fatalf("expected cell 8 to be highlighted 'active' in final step, got %q", doneState.Cells[8].State)
	}
	if doneState.Cells[8].Value != "21" {
		t.Fatalf("expected cell 8 value '21', got %q", doneState.Cells[8].Value)
	}
	for c := 0; c < 8; c++ {
		if doneState.Cells[c].State != "computed" {
			t.Fatalf("expected cell %d to be 'computed' in final step, got %q", c, doneState.Cells[c].State)
		}
	}
}

func TestFibonacciTableState_N0_N1(t *testing.T) {
	lesson := Lesson{}

	// Test n = 0
	steps0 := Trace(0)
	if err := catalog.ValidateSourceLineIDs(lesson, steps0); err != nil {
		t.Fatalf("Trace(0) line validation failed: %v", err)
	}
	last0 := steps0[len(steps0)-1]
	ts0 := last0.State.(renderers.TableState)
	if !reflect.DeepEqual(ts0.RowHeaders, []string{"F(i)"}) {
		t.Fatalf("Trace(0) RowHeaders = %v, want [\"F(i)\"]", ts0.RowHeaders)
	}
	if !reflect.DeepEqual(ts0.ColHeaders, []string{"0"}) {
		t.Fatalf("Trace(0) ColHeaders = %v, want [\"0\"]", ts0.ColHeaders)
	}
	if len(ts0.Cells) != 1 || ts0.Cells[0].Value != "0" || ts0.Cells[0].State != "active" {
		t.Fatalf("Trace(0) unexpected cell 0: %+v", ts0.Cells[0])
	}

	// Test n = 1
	steps1 := Trace(1)
	if err := catalog.ValidateSourceLineIDs(lesson, steps1); err != nil {
		t.Fatalf("Trace(1) line validation failed: %v", err)
	}
	last1 := steps1[len(steps1)-1]
	ts1 := last1.State.(renderers.TableState)
	if !reflect.DeepEqual(ts1.RowHeaders, []string{"F(i)"}) {
		t.Fatalf("Trace(1) RowHeaders = %v, want [\"F(i)\"]", ts1.RowHeaders)
	}
	if !reflect.DeepEqual(ts1.ColHeaders, []string{"0", "1"}) {
		t.Fatalf("Trace(1) ColHeaders = %v, want [\"0\", \"1\"]", ts1.ColHeaders)
	}
	if len(ts1.Cells) != 2 {
		t.Fatalf("Trace(1) expected 2 cells, got %d", len(ts1.Cells))
	}
	if ts1.Cells[0].State != "computed" || ts1.Cells[0].Value != "0" {
		t.Fatalf("Trace(1) unexpected cell 0: %+v", ts1.Cells[0])
	}
	if ts1.Cells[1].State != "active" || ts1.Cells[1].Value != "1" {
		t.Fatalf("Trace(1) unexpected cell 1 (final highlighted): %+v", ts1.Cells[1])
	}
}

func TestFibonacciSnapshotImmutability(t *testing.T) {
	steps := Trace(8)

	for i := 0; i < len(steps); i++ {
		tsI := steps[i].State.(renderers.TableState)
		for j := i + 1; j < len(steps); j++ {
			tsJ := steps[j].State.(renderers.TableState)
			if &tsI.Cells[0] == &tsJ.Cells[0] {
				t.Fatalf("steps %d and %d share identical Cells slice pointer", i, j)
			}
			if len(tsI.Dependencies) > 0 && len(tsJ.Dependencies) > 0 {
				if &tsI.Dependencies[0] == &tsJ.Dependencies[0] {
					t.Fatalf("steps %d and %d share identical Dependencies slice pointer", i, j)
				}
			}
		}
	}

	// Mutate snapshot cells in place and verify other steps do not change
	origVal := steps[0].State.(renderers.TableState).Cells[0].Value
	steps[0].State.(renderers.TableState).Cells[0].Value = "CORRUPTED"

	if steps[1].State.(renderers.TableState).Cells[0].Value == "CORRUPTED" {
		t.Fatal("mutating step 0 corrupted step 1 cells")
	}

	freshSteps := Trace(8)
	if freshSteps[0].State.(renderers.TableState).Cells[0].Value != origVal {
		t.Fatalf("fresh trace was affected by mutation: got %q, want %q", freshSteps[0].State.(renderers.TableState).Cells[0].Value, origVal)
	}
}

func TestFibonacciRenderTableSVG(t *testing.T) {
	for _, n := range []int{0, 1, 8} {
		steps := Trace(n)
		for i, step := range steps {
			svg := renderers.RenderTableSVG(step, 800, 600)
			if svg == "" || svg == `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No table state</text>` {
				t.Fatalf("n=%d step %d rendered invalid table svg: %s", n, i, svg)
			}
		}
	}
}

