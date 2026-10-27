package bst

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBSTCorrectness(t *testing.T) {
	tree := New()

	// Empty tree
	if tree.Search(10) {
		t.Fatalf("expected Search(10) on empty tree to be false")
	}
	tree.Delete(10)
	if len(tree.InOrder()) != 0 {
		t.Fatalf("expected empty InOrder slice, got %v", tree.InOrder())
	}

	// Insert elements
	values := []int{50, 30, 70, 20, 40, 60, 80}
	for _, v := range values {
		tree.Insert(v)
	}

	// Duplicate insert should be no-op
	tree.Insert(50)
	tree.Insert(30)

	// Verify in-order traversal
	expectedOrder := []int{20, 30, 40, 50, 60, 70, 80}
	if !reflect.DeepEqual(tree.InOrder(), expectedOrder) {
		t.Fatalf("InOrder got %v, want %v", tree.InOrder(), expectedOrder)
	}

	// Search tests
	for _, v := range values {
		if !tree.Search(v) {
			t.Fatalf("expected Search(%d) to be true", v)
		}
	}
	for _, missing := range []int{10, 25, 35, 65, 90, 100} {
		if tree.Search(missing) {
			t.Fatalf("expected Search(%d) to be false", missing)
		}
	}

	// Delete leaf node (20)
	tree.Delete(20)
	if tree.Search(20) {
		t.Fatalf("expected 20 to be deleted")
	}
	expectedAfter20 := []int{30, 40, 50, 60, 70, 80}
	if !reflect.DeepEqual(tree.InOrder(), expectedAfter20) {
		t.Fatalf("after deleting 20, InOrder got %v, want %v", tree.InOrder(), expectedAfter20)
	}

	// Delete node with one child
	// Insert 25 (child of 30 since 20 was deleted)
	tree.Insert(25)
	// Now 30 has left child 25 and right child 40.
	// Delete 25 (leaf)
	tree.Delete(25)
	// Now insert 35 under 40
	tree.Insert(35)
	// Now 40 has left child 35 and no right child (one child)
	tree.Delete(40)
	if tree.Search(40) {
		t.Fatalf("expected 40 to be deleted")
	}
	if !tree.Search(35) {
		t.Fatalf("expected child 35 to still exist after parent 40 deleted")
	}

	// Delete node with two children (50 - root)
	tree.Delete(50)
	if tree.Search(50) {
		t.Fatalf("expected root 50 to be deleted")
	}
	expectedAfter50 := []int{30, 35, 60, 70, 80}
	if !reflect.DeepEqual(tree.InOrder(), expectedAfter50) {
		t.Fatalf("after deleting root 50, InOrder got %v, want %v", tree.InOrder(), expectedAfter50)
	}

	// Delete remaining nodes
	for _, v := range expectedAfter50 {
		tree.Delete(v)
	}
	if len(tree.InOrder()) != 0 {
		t.Fatalf("expected empty tree, got %v", tree.InOrder())
	}
}

func TestBSTRandomInsertsAndDeletes(t *testing.T) {
	tree := New()
	input := []int{42, 17, 88, 8, 23, 65, 94, 4, 11, 20, 29, 73}

	for _, v := range input {
		tree.Insert(v)
	}

	sorted := make([]int, len(input))
	copy(sorted, input)
	sort.Ints(sorted)

	if !reflect.DeepEqual(tree.InOrder(), sorted) {
		t.Fatalf("InOrder got %v, want %v", tree.InOrder(), sorted)
	}

	// Delete in mixed order
	deleteOrder := []int{23, 88, 4, 42, 17, 73, 94, 65, 11, 20, 8, 29}
	for _, v := range deleteOrder {
		tree.Delete(v)
		if tree.Search(v) {
			t.Fatalf("deleted node %d still found in search", v)
		}
	}
	if len(tree.InOrder()) != 0 {
		t.Fatalf("expected empty tree after all deletions, got %v", tree.InOrder())
	}
}

func TestBSTLessonValidation(t *testing.T) {
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

	meta := lesson.Meta()
	if meta.ID != "bst" {
		t.Errorf("meta.ID = %q; want 'bst'", meta.ID)
	}
	if meta.Name != "Binary Search Tree (BST)" {
		t.Errorf("meta.Name = %q; want 'Binary Search Tree (BST)'", meta.Name)
	}
	if meta.Category != catalog.CategoryTrees {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryTrees)
	}
	if meta.Renderer != "tree" {
		t.Errorf("meta.Renderer = %q; want 'tree'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(log n) avg, O(n) worst" {
		t.Errorf("TimeComplexity = %q; want 'O(log n) avg, O(n) worst'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Errorf("SpaceComplexity = %q; want 'O(n)'", meta.Documentation.SpaceComplexity)
	}
	expectedConcepts := []string{"Left < Root < Right Invariant", "Recursive Descent", "In-Order Traversal"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("KeyConcepts = %v; want %v", meta.Documentation.KeyConcepts, expectedConcepts)
	}
}

func TestBSTTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	seenStates := make(map[string]bool)

	for i, step := range steps {
		state, ok := step.State.(renderers.TreeState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}

		nodeMap := make(map[string]renderers.TreeNode)
		for _, n := range state.Nodes {
			seenStates[n.State] = true
			nodeMap[n.ID] = n
		}

		// Validate tree geometry for every step
		for _, n := range state.Nodes {
			if n.ParentID == "" {
				// Root position
				if n.X != 400 || n.Y != 60 {
					t.Errorf("step %d: root node %s at (%.1f, %.1f), want (400, 60)", i, n.ID, n.X, n.Y)
				}
			} else {
				parent, exists := nodeMap[n.ParentID]
				if !exists {
					t.Errorf("step %d: node %s has unknown parent %s", i, n.ID, n.ParentID)
					continue
				}
				if n.Y <= parent.Y {
					t.Errorf("step %d: child Y (%.1f) not below parent Y (%.1f)", i, n.Y, parent.Y)
				}
				if n.Val < parent.Val && n.X >= parent.X {
					t.Errorf("step %d: left child %d (X=%.1f) should be left of parent %d (X=%.1f)", i, n.Val, n.X, parent.Val, parent.X)
				}
				if n.Val > parent.Val && n.X <= parent.X {
					t.Errorf("step %d: right child %d (X=%.1f) should be right of parent %d (X=%.1f)", i, n.Val, n.X, parent.Val, parent.X)
				}
			}
		}

		// Verify no overlapping nodes (circles with radius 20 must have center distance > 30)
		for j := 0; j < len(state.Nodes); j++ {
			for k := j + 1; k < len(state.Nodes); k++ {
				n1 := state.Nodes[j]
				n2 := state.Nodes[k]
				if n1.Y == n2.Y {
					dist := n1.X - n2.X
					if dist < 0 {
						dist = -dist
					}
					if dist < 30 {
						t.Errorf("step %d: nodes %d and %d too close on same level (dist=%.1f)", i, n1.Val, n2.Val, dist)
					}
				}
			}
		}
	}

	// Verify all 5 node states were exercised
	for _, reqState := range []string{"normal", "active", "matched", "inserted", "deleted"} {
		if !seenStates[reqState] {
			t.Errorf("missing expected node state %q across trace steps", reqState)
		}
	}
}

func TestBSTCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{
		"initial": [100, 50, 150],
		"ops": [
			{"type": "insert", "value": 75},
			{"type": "search", "value": 150},
			{"type": "delete", "value": 50}
		]
	}`)

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
}

func TestBSTSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	// Capture initial node states
	type nodeSnapshot struct {
		id    string
		val   int
		state string
	}
	stepSnapshots := make([][]nodeSnapshot, len(steps))
	for i, step := range steps {
		ts, ok := step.State.(renderers.TreeState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}
		stepSnapshots[i] = make([]nodeSnapshot, len(ts.Nodes))
		for j, n := range ts.Nodes {
			stepSnapshots[i][j] = nodeSnapshot{id: n.ID, val: n.Val, state: n.State}
		}
	}

	// Mutate all node objects in all steps
	for _, step := range steps {
		ts := step.State.(renderers.TreeState)
		for j := range ts.Nodes {
			ts.Nodes[j].Val = -9999
			ts.Nodes[j].State = "corrupted"
		}
	}

	// Verify that fresh steps from Build are unaffected
	freshSteps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("fresh Build failed: %v", err)
	}
	for i, step := range freshSteps {
		ts := step.State.(renderers.TreeState)
		if len(ts.Nodes) != len(stepSnapshots[i]) {
			t.Fatalf("step %d node count mismatch", i)
		}
		for j, n := range ts.Nodes {
			if n.Val != stepSnapshots[i][j].val || n.State != stepSnapshots[i][j].state {
				t.Fatalf("step %d node %d was corrupted: got val=%d state=%s, want val=%d state=%s",
					i, j, n.Val, n.State, stepSnapshots[i][j].val, stepSnapshots[i][j].state)
			}
		}
	}

	// Verify each step in freshSteps has unique Nodes slice backing array
	for i := 0; i < len(freshSteps); i++ {
		tsI := freshSteps[i].State.(renderers.TreeState)
		if len(tsI.Nodes) == 0 {
			continue
		}
		for j := i + 1; j < len(freshSteps); j++ {
			tsJ := freshSteps[j].State.(renderers.TreeState)
			if len(tsJ.Nodes) == 0 {
				continue
			}
			if &tsI.Nodes[0] == &tsJ.Nodes[0] {
				t.Errorf("steps %d and %d share same Nodes slice memory", i, j)
			}
		}
	}
}

func TestBSTInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Test null
	if _, err := lesson.Build(json.RawMessage("null")); err == nil {
		t.Errorf("expected error for null input, got nil")
	}
	if _, err := lesson.Build(json.RawMessage("  null  \n")); err == nil {
		t.Errorf("expected error for whitespace null input, got nil")
	}

	// Test malformed JSON
	malformed := []string{
		"{bad json",
		`{"initial": [10, 20], "ops": [}`,
		`[1, 2, 3]`,
		`"not an object"`,
		`123`,
	}
	for _, m := range malformed {
		if _, err := lesson.Build(json.RawMessage(m)); err == nil {
			t.Errorf("expected error for malformed input %q, got nil", m)
		}
	}

	// Test empty ops
	emptyOps := []string{
		`{}`,
		`{"ops": []}`,
		`{"initial": [10, 20]}`,
		`{"initial": [10, 20], "ops": []}`,
	}
	for _, e := range emptyOps {
		if _, err := lesson.Build(json.RawMessage(e)); err == nil {
			t.Errorf("expected error for empty ops input %q, got nil", e)
		}
	}

	// Test invalid op type
	if _, err := lesson.Build(json.RawMessage(`{"ops": [{"type": "rotate", "value": 10}]}`)); err == nil {
		t.Errorf("expected error for invalid op type, got nil")
	}

	// Test valid empty initial with ops
	validEmptyInitial := json.RawMessage(`{"initial": [], "ops": [{"type": "insert", "value": 42}]}`)
	steps, err := lesson.Build(validEmptyInitial)
	if err != nil {
		t.Fatalf("unexpected error for valid empty initial: %v", err)
	}
	if len(steps) == 0 {
		t.Fatalf("expected steps for valid empty initial, got 0")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	// Test omitted initial uses default initial
	validOmittedInitial := json.RawMessage(`{"ops": [{"type": "search", "value": 50}]}`)
	steps2, err := lesson.Build(validOmittedInitial)
	if err != nil {
		t.Fatalf("unexpected error for omitted initial: %v", err)
	}
	if len(steps2) == 0 {
		t.Fatalf("expected steps for omitted initial, got 0")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps2); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}

func TestBSTTeachingRequirements(t *testing.T) {
	lesson := Lesson{}

	// Scenario covering:
	// 1. Insert path descent & newly placed leaf
	// 2. Search found & search not found (reaching nil)
	// 3. Deletion Case 1 (leaf)
	// 4. Deletion Case 2 (one child: right)
	// 5. Deletion Case 2 (one child: left)
	// 6. Deletion Case 3 (two children: in-order successor)
	input := json.RawMessage(`{
		"initial": [50, 30, 70, 20, 40, 60, 80],
		"ops": [
			{"type": "insert", "value": 25},
			{"type": "search", "value": 25},
			{"type": "search", "value": 999},
			{"type": "delete", "value": 40},
			{"type": "delete", "value": 20},
			{"type": "insert", "value": 15},
			{"type": "delete", "value": 25},
			{"type": "delete", "value": 50}
		]
	}`)

	steps, err := lesson.Build(input)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line ID validation failed: %v", err)
	}

	var hasInsertLeaf bool
	var hasSearchFound bool
	var hasSearchNotFound bool
	var hasDeleteLeaf bool
	var hasDeleteOneChildRight bool
	var hasDeleteOneChildLeft bool
	var hasDeleteTwoChild bool
	var hasInvariantExplanation bool

	for _, step := range steps {
		expl := step.Explanation
		if strings.Contains(expl, "BST property dictates") {
			hasInvariantExplanation = true
		}
		if strings.Contains(expl, "newly placed leaf") {
			hasInsertLeaf = true
		}
		if strings.Contains(expl, "matched!") && strings.Contains(expl, "25") {
			hasSearchFound = true
		}
		if strings.Contains(expl, "Reached nil") && strings.Contains(expl, "999") {
			hasSearchNotFound = true
		}
		if strings.Contains(expl, "Case 1 (Leaf)") && strings.Contains(expl, "40") {
			hasDeleteLeaf = true
		}
		if strings.Contains(expl, "Case 2 (One Child)") && strings.Contains(expl, "no left child") {
			hasDeleteOneChildRight = true
		}
		if strings.Contains(expl, "Case 2 (One Child)") && strings.Contains(expl, "no right child") {
			hasDeleteOneChildLeft = true
		}
		if strings.Contains(expl, "Case 3 (Two Children)") && strings.Contains(expl, "in-order successor") {
			hasDeleteTwoChild = true
		}
	}

	if !hasInvariantExplanation {
		t.Errorf("missing BST invariant explanation along traversal path")
	}
	if !hasInsertLeaf {
		t.Errorf("missing newly placed leaf explanation during insert")
	}
	if !hasSearchFound {
		t.Errorf("missing search found step for 25")
	}
	if !hasSearchNotFound {
		t.Errorf("missing search not found / reached nil step for 999")
	}
	if !hasDeleteLeaf {
		t.Errorf("missing Case 1 (Leaf) deletion explanation for 40")
	}
	if !hasDeleteOneChildRight {
		t.Errorf("missing Case 2 (One Child: right) deletion explanation")
	}
	if !hasDeleteOneChildLeft {
		t.Errorf("missing Case 2 (One Child: left) deletion explanation")
	}
	if !hasDeleteTwoChild {
		t.Errorf("missing Case 3 (Two Children / in-order successor) deletion explanation")
	}
}
