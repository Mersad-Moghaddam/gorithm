package bst

import (
	"reflect"
	"sort"
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
