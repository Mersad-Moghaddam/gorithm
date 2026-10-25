package fibonacci

import (
	"encoding/json"
	"reflect"
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

func TestFibonacciRenderTableSVG(t *testing.T) {
	steps := Trace(8)
	for i, step := range steps {
		svg := renderers.RenderTableSVG(step, 800, 600)
		if svg == "" || svg == `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No table state</text>` {
			t.Fatalf("step %d rendered invalid table svg: %s", i, svg)
		}
	}
}

