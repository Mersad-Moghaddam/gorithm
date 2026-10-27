package knapsack

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSolveCorrectness(t *testing.T) {
	tests := []struct {
		name      string
		weights   []int
		values    []int
		capacity  int
		wantMax   int
		wantTable [][]int
	}{
		{
			name:     "default example",
			weights:  []int{2, 3, 4, 5},
			values:   []int{3, 4, 5, 6},
			capacity: 5,
			wantMax:  7, // items 0 and 1: weights 2+3=5, values 3+4=7
		},
		{
			name:     "zero capacity",
			weights:  []int{1, 2, 3},
			values:   []int{10, 20, 30},
			capacity: 0,
			wantMax:  0,
		},
		{
			name:     "empty items",
			weights:  []int{},
			values:   []int{},
			capacity: 10,
			wantMax:  0,
		},
		{
			name:     "all items fit",
			weights:  []int{1, 2, 3},
			values:   []int{6, 10, 12},
			capacity: 10,
			wantMax:  28,
		},
		{
			name:     "no item fits",
			weights:  []int{5, 6, 7},
			values:   []int{10, 20, 30},
			capacity: 4,
			wantMax:  0,
		},
		{
			name:     "single item fits",
			weights:  []int{3},
			values:   []int{25},
			capacity: 3,
			wantMax:  25,
		},
		{
			name:     "single item does not fit",
			weights:  []int{5},
			values:   []int{25},
			capacity: 3,
			wantMax:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMax, gotTable := Solve(tt.weights, tt.values, tt.capacity)
			if gotMax != tt.wantMax {
				t.Errorf("Solve() gotMax = %d, want %d", gotMax, tt.wantMax)
			}
			if len(gotTable) != len(tt.weights)+1 {
				t.Fatalf("Solve() table rows = %d, want %d", len(gotTable), len(tt.weights)+1)
			}
			for r, row := range gotTable {
				if len(row) != tt.capacity+1 {
					t.Fatalf("Solve() table row %d len = %d, want %d", r, len(row), tt.capacity+1)
				}
			}
			if gotTable[len(tt.weights)][tt.capacity] != tt.wantMax {
				t.Errorf("Solve() table end cell = %d, want %d", gotTable[len(tt.weights)][tt.capacity], tt.wantMax)
			}
		})
	}
}

func TestKnapsackMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "knapsack" {
		t.Errorf("expected ID 'knapsack', got %q", meta.ID)
	}
	if meta.Name != "0/1 Knapsack (DP)" {
		t.Errorf("expected Name '0/1 Knapsack (DP)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryDP {
		t.Errorf("expected Category %q, got %q", catalog.CategoryDP, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n · W)" {
		t.Errorf("expected TimeComplexity 'O(n · W)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n · W)" {
		t.Errorf("expected SpaceComplexity 'O(n · W)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"2D State Matrix", "Include/Exclude Choice", "Pseudo-Polynomial Time"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("knapsack")
	if !ok {
		t.Fatal("lesson 'knapsack' not registered in catalog")
	}
	if l.Meta().ID != "knapsack" {
		t.Errorf("expected registered ID 'knapsack', got %q", l.Meta().ID)
	}
}

func TestKnapsackTraceValidation(t *testing.T) {
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
	if maxVal, ok := lastStep.Variables["maxValue"].(int); !ok || maxVal != 7 {
		t.Fatalf("expected maxValue 7, got %v", lastStep.Variables["maxValue"])
	}
}

func TestKnapsackTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"weights": [1, 2, 3], "values": [6, 10, 12], "capacity": 5}`)
	steps, err := lesson.Build(json.RawMessage(input))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
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
	// Items 1 and 2 (w=2,v=10 and w=3,v=12) give 22. Items 0 and 2 (w=1,v=6 and w=3,v=12) give 18.
	// Weights 1+2=3 < 5 (v=16). Weights 1+2+3=6 > 5. Weights 2+3=5 (v=22).
	if maxVal, ok := lastStep.Variables["maxValue"].(int); !ok || maxVal != 22 {
		t.Fatalf("expected maxValue 22, got %v", lastStep.Variables["maxValue"])
	}
}

func TestKnapsackRenderTableSVG(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	for i, step := range steps {
		svg := renderers.RenderTableSVG(step, 800, 450)
		if svg == "" {
			t.Errorf("step %d produced empty SVG", i)
		}
	}
}
