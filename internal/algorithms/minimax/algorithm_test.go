package minimax

import (
	"reflect"
	"strings"
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
	hasCutoffCaption := false
	hasOptimalMove := false

	for i, step := range steps {
		state, ok := step.State.(renderers.TreeState)
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
		if _, ok := step.Variables["bounds"]; !ok {
			t.Errorf("step %d missing 'bounds' in variables", i)
		}
		if _, ok := step.Variables["[alpha, beta]"]; !ok {
			t.Errorf("step %d missing '[alpha, beta]' in variables", i)
		}

		if cutoff, ok := step.Variables["cutoff"].(bool); ok && cutoff {
			hasCutoff = true
			if strings.Contains(step.Explanation, "Alpha-Beta Cutoff: beta <= alpha, branch pruned!") {
				hasCutoffCaption = true
			}
		}

		for _, n := range state.Nodes {
			seenStates[n.State] = true
			if n.State == "inserted" {
				hasOptimalMove = true
			}
		}
	}

	if !hasCutoff {
		t.Errorf("expected at least one alpha/beta cutoff step in trace")
	}
	if !hasCutoffCaption {
		t.Errorf("expected cutoff caption containing 'Alpha-Beta Cutoff: beta <= alpha, branch pruned!'")
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

	// Verify final step declares chosen move
	finalStep := steps[len(steps)-1]
	optMove, ok := finalStep.Variables["optimalMove"].(string)
	if !ok || optMove == "" {
		t.Fatalf("expected non-empty optimalMove in final step, got %v", finalStep.Variables["optimalMove"])
	}
	if bestVal, ok := finalStep.Variables["bestValue"].(int); !ok || bestVal != 4 {
		t.Fatalf("expected bestValue 4 in final step, got %v", finalStep.Variables["bestValue"])
	}
}

func TestMinimaxTreeStateAndPruning(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Initial step check: Root should start as MAX player, unvisited internal nodes have MAX/MIN labels
	initialState := steps[0].State.(renderers.TreeState)
	for _, n := range initialState.Nodes {
		if n.ID == "n0" && n.Label != "MAX" {
			t.Errorf("root node expected label 'MAX', got %q", n.Label)
		}
	}

	// Final step check: Root node propagated score should be 4
	finalState := steps[len(steps)-1].State.(renderers.TreeState)
	var rootNode *renderers.TreeNode
	prunedCount := 0
	for i := range finalState.Nodes {
		n := &finalState.Nodes[i]
		if n.ID == "n0" {
			rootNode = n
		}
		if n.State == "deleted" {
			prunedCount++
			if n.Label != "✕" {
				t.Errorf("pruned node %s expected label '✕', got %q", n.ID, n.Label)
			}
		}
	}

	if rootNode == nil {
		t.Fatal("root node n0 not found in final TreeState")
	}
	if rootNode.Val != 4 {
		t.Errorf("root node value = %d; want 4", rootNode.Val)
	}
	if rootNode.Label != "4" {
		t.Errorf("root node label = %q; want '4'", rootNode.Label)
	}
	if prunedCount == 0 {
		t.Error("expected at least one pruned branch with 'deleted' state in final TreeState")
	}
}

func TestMinimaxSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	step0 := steps[0].State.(renderers.TreeState)
	stepLast := steps[len(steps)-1].State.(renderers.TreeState)

	// Step 0 root should still be "normal" or "active" during init, never "inserted"
	if step0.Nodes[0].State == "inserted" {
		t.Fatalf("step 0 was mutated by later steps: state is 'inserted'")
	}
	if stepLast.Nodes[0].State != "inserted" {
		t.Fatalf("last step root node should be 'inserted', got %q", stepLast.Nodes[0].State)
	}

	// Modifying step0 slice must not affect stepLast
	origLastState := stepLast.Nodes[0].State
	step0.Nodes[0].State = "mutated_state"
	if stepLast.Nodes[0].State != origLastState {
		t.Fatalf("step snapshots share memory! Mutating step 0 altered stepLast: %q", stepLast.Nodes[0].State)
	}

	// Verify slice pointers are different
	hdr0 := reflect.ValueOf(step0.Nodes).Pointer()
	hdrLast := reflect.ValueOf(stepLast.Nodes).Pointer()
	if hdr0 == hdrLast {
		t.Fatalf("step 0 and last step share the exact same underlying node slice pointer: %x", hdr0)
	}
}

func TestMinimaxInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Nil / empty input defaults correctly
	if _, err := lesson.Build(nil); err != nil {
		t.Errorf("expected nil input to succeed with defaults, got error: %v", err)
	}
	if _, err := lesson.Build([]byte("")); err != nil {
		t.Errorf("expected empty input to succeed with defaults, got error: %v", err)
	}
	if _, err := lesson.Build([]byte("   ")); err != nil {
		t.Errorf("expected whitespace input to succeed with defaults, got error: %v", err)
	}

	// Rejections
	invalidInputs := []struct {
		name  string
		input string
	}{
		{"null literal", "null"},
		{"whitespace null", "  null  "},
		{"malformed json syntax", "{ scores: [1, 2]"},
		{"unclosed array", "[1, 2, "},
		{"empty array", "[]"},
		{"empty object scores", `{"scores": []}`},
		{"null object scores", `{"scores": null}`},
		{"missing scores object", `{"branching": 2}`},
		{"negative branching", `{"scores": [1, 2], "branching": -1}`},
		{"primitive number", "42"},
		{"primitive string", `"hello"`},
		{"primitive boolean", "true"},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build([]byte(tc.input))
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tc.input)
			}
		})
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
