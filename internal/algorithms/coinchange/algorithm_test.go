package coinchange

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestMinCoins(t *testing.T) {
	tests := []struct {
		name       string
		coins      []int
		amount     int
		wantCoins  int
		wantDPSize int
	}{
		{
			name:       "standard amount 11",
			coins:      []int{1, 2, 5},
			amount:     11,
			wantCoins:  3,
			wantDPSize: 12,
		},
		{
			name:       "impossible amount",
			coins:      []int{2},
			amount:     3,
			wantCoins:  -1,
			wantDPSize: 4,
		},
		{
			name:       "zero amount",
			coins:      []int{1},
			amount:     0,
			wantCoins:  0,
			wantDPSize: 1,
		},
		{
			name:       "larger amount",
			coins:      []int{1, 2, 5, 10},
			amount:     27,
			wantCoins:  4,
			wantDPSize: 28,
		},
		{
			name:       "greedy failure case 1, 5, 6, 9 amount 11",
			coins:      []int{1, 5, 6, 9},
			amount:     11,
			wantCoins:  2,
			wantDPSize: 12,
		},
		{
			name:       "negative amount",
			coins:      []int{1, 2, 5},
			amount:     -1,
			wantCoins:  -1,
			wantDPSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotDP := MinCoins(tt.coins, tt.amount)
			if gotMin != tt.wantCoins {
				t.Errorf("MinCoins(%v, %d) = %d, want %d", tt.coins, tt.amount, gotMin, tt.wantCoins)
			}
			if len(gotDP) != tt.wantDPSize {
				t.Errorf("len(gotDP) = %d, want %d", len(gotDP), tt.wantDPSize)
			}
		})
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "coinchange" {
		t.Errorf("expected ID 'coinchange', got %q", meta.ID)
	}
	if meta.Name != "Coin Change (DP)" {
		t.Errorf("expected Name 'Coin Change (DP)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryDP {
		t.Errorf("expected Category %q, got %q", catalog.CategoryDP, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(amount · |coins|)" {
		t.Errorf("expected TimeComplexity 'O(amount · |coins|)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(amount)" {
		t.Errorf("expected SpaceComplexity 'O(amount)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"Unbounded Knapsack",
		"Overlapping Subproblems",
		"Bottom-up Tabulation",
		"Min Optimization",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	registered, ok := catalog.Get("coinchange")
	if !ok {
		t.Fatal("lesson 'coinchange' not registered in catalog")
	}
	if registered.Meta().ID != "coinchange" {
		t.Errorf("expected registered ID 'coinchange', got %q", registered.Meta().ID)
	}
}

func TestTraceValidation(t *testing.T) {
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
	if minCoins, ok := lastStep.Variables["minCoins"].(int); !ok || minCoins != 3 {
		t.Fatalf("expected minCoins 3, got %v", lastStep.Variables["minCoins"])
	}

	for i, step := range steps {
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
		}
		if len(tableState.RowHeaders) == 0 || len(tableState.ColHeaders) == 0 {
			t.Fatalf("step %d has empty headers", i)
		}
		if len(tableState.Cells) != 12 {
			t.Fatalf("step %d has %d cells, want 12", i, len(tableState.Cells))
		}
	}
}

func TestTraceCustomInput(t *testing.T) {
	lesson := Lesson{}

	// Case 1: solvable custom input
	input1 := []byte(`{"coins": [1, 5, 6, 9], "amount": 11}`)
	steps1, err := lesson.Build(json.RawMessage(input1))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps1); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
	lastStep1 := steps1[len(steps1)-1]
	if lastStep1.Action != trace.ActionDone {
		t.Fatalf("expected ActionDone, got %s", lastStep1.Action)
	}
	if minCoins, ok := lastStep1.Variables["minCoins"].(int); !ok || minCoins != 2 {
		t.Fatalf("expected minCoins 2, got %v", lastStep1.Variables["minCoins"])
	}

	// Case 2: impossible custom input
	input2 := []byte(`{"coins": [2], "amount": 3}`)
	steps2, err := lesson.Build(json.RawMessage(input2))
	if err != nil {
		t.Fatalf("failed to build impossible custom input: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps2); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
	lastStep2 := steps2[len(steps2)-1]
	if lastStep2.Action != trace.ActionDone {
		t.Fatalf("expected ActionDone, got %s", lastStep2.Action)
	}
	if minCoins, ok := lastStep2.Variables["minCoins"].(int); !ok || minCoins != -1 {
		t.Fatalf("expected minCoins -1, got %v", lastStep2.Variables["minCoins"])
	}
}

func TestRenderTableSVG(t *testing.T) {
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

