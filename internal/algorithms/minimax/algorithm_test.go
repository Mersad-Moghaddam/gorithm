package minimax

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestMinimaxCorrectness(t *testing.T) {
	// Default tree: depth 3, branching 2, terminal scores [-1, 4, 2, 6, -3, -5, 0, 7]
	scores := []int{-1, 4, 2, 6, -3, -5, 0, 7}
	tree := BuildTree(scores, 2)
	val, bestChild := Solve(tree)
	if val != 4 {
		t.Fatalf("expected optimal root value 4, got %d", val)
	}
	if bestChild == nil || bestChild.Val != 4 {
		t.Fatalf("expected best child value 4, got %v", bestChild)
	}

	// Test with textbook example: [3, 5, 6, 9, 1, 2, 0, -1] -> best val 5
	tree2 := BuildTree([]int{3, 5, 6, 9, 1, 2, 0, -1}, 2)
	val2, _ := Solve(tree2)
	if val2 != 5 {
		t.Fatalf("expected optimal root value 5, got %d", val2)
	}

	// Test small depth 2 tree: [3, 5, 2, 9] -> best val 3
	tree3 := BuildTree([]int{3, 5, 2, 9}, 2)
	val3, _ := Solve(tree3)
	if val3 != 3 {
		t.Fatalf("expected optimal root value 3, got %d", val3)
	}

	// Test single leaf
	tree4 := BuildTree([]int{42}, 2)
	val4, _ := Solve(tree4)
	if val4 != 42 {
		t.Fatalf("expected single leaf value 42, got %d", val4)
	}
}

func TestMinimaxLessonMeta(t *testing.T) {
	lesson, ok := catalog.Get("minimax")
	if !ok {
		t.Fatal("expected minimax lesson to be registered in catalog")
	}
	meta := lesson.Meta()
	if meta.ID != "minimax" {
		t.Errorf("ID = %q; want 'minimax'", meta.ID)
	}
	if meta.Name != "Minimax & Alpha-Beta Pruning" {
		t.Errorf("Name = %q; want 'Minimax & Alpha-Beta Pruning'", meta.Name)
	}
	if meta.Category != catalog.CategoryAI {
		t.Errorf("Category = %q; want %q", meta.Category, catalog.CategoryAI)
	}
	if meta.Renderer != "tree" {
		t.Errorf("Renderer = %q; want 'tree'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(b^(d/2)) best, O(b^d) worst" {
		t.Errorf("TimeComplexity = %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(b · d)" {
		t.Errorf("SpaceComplexity = %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("Difficulty = %q; want 'Hard'", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{"Adversarial Search", "Alpha-Beta Cutoffs", "Game Theory", "Zero-Sum"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("KeyConcepts = %v; want %v", meta.Documentation.KeyConcepts, expectedConcepts)
	}
}

func TestMinimaxTraceAndSourceLineIDs(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected at least one step in trace")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	seenStates := make(map[string]bool)
	hasCutoff := false
	hasOptimalMove := false

	for i, step := range steps {
		state, ok := step.State.(renderers.MinimaxLessonState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}

		if step.Variables == nil {
			t.Fatalf("step %d has nil variables", i)
		}
		if _, ok := step.Variables["alpha"]; !ok {
			t.Errorf("step %d missing 'alpha' in variables", i)
		}
		if _, ok := step.Variables["beta"]; !ok {
			t.Errorf("step %d missing 'beta' in variables", i)
		}

		if cutoff, ok := step.Variables["cutoff"].(bool); ok && cutoff {
			hasCutoff = true
		}

		for _, n := range state.TreeState.Nodes {
			seenStates[n.State] = true
			if n.State == "inserted" {
				hasOptimalMove = true
			}
		}
	}

	if !hasCutoff {
		t.Errorf("expected at least one alpha/beta cutoff step in trace")
	}
	if !hasOptimalMove {
		t.Errorf("expected optimal move to be marked inserted in tree state")
	}

	// Verify all required node states were exercised
	for _, reqState := range []string{"normal", "active", "matched", "deleted", "inserted"} {
		if !seenStates[reqState] {
			t.Errorf("missing expected node state %q across trace steps", reqState)
		}
	}
}

func TestMinimaxCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{"scores": [3, 5, 6, 9, 1, 2, 0, -1], "branching": 2}`)

	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("Build failed with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("custom input trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on custom input: %v", err)
	}

	// Also support raw json array of ints
	rawArrInput := []byte(`[10, 20, 30, 40]`)
	steps2, err := lesson.Build(rawArrInput)
	if err != nil {
		t.Fatalf("Build failed with array input: %v", err)
	}
	if len(steps2) == 0 {
		t.Fatal("array input trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps2); err != nil {
		t.Fatalf("line validation failed on array input: %v", err)
	}
}
