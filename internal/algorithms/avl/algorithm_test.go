package avl

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func isAVL(t *testing.T, n *Node) (int, bool) {
	if n == nil {
		return 0, true
	}

	lh, lok := isAVL(t, n.Left)
	rh, rok := isAVL(t, n.Right)

	if !lok || !rok {
		return 0, false
	}

	if n.Left != nil && n.Left.Val >= n.Val {
		t.Errorf("BST invariant violated: left child %d >= node %d", n.Left.Val, n.Val)
		return 0, false
	}
	if n.Right != nil && n.Right.Val <= n.Val {
		t.Errorf("BST invariant violated: right child %d <= node %d", n.Right.Val, n.Val)
		return 0, false
	}

	diff := lh - rh
	if diff < -1 || diff > 1 {
		t.Errorf("AVL balance violated at node %d: left height %d, right height %d, diff %d", n.Val, lh, rh, diff)
		return 0, false
	}

	expectedHeight := 1
	if lh > rh {
		expectedHeight += lh
	} else {
		expectedHeight += rh
	}

	if n.Height != expectedHeight {
		t.Errorf("node %d stored height %d != expected height %d", n.Val, n.Height, expectedHeight)
		return 0, false
	}

	return expectedHeight, true
}

func TestAVLCorrectness(t *testing.T) {
	tree := New()

	if tree.Search(10) {
		t.Fatalf("expected Search(10) on empty tree to be false")
	}
	if len(tree.InOrder()) != 0 {
		t.Fatalf("expected empty InOrder slice, got %v", tree.InOrder())
	}

	values := []int{10, 20, 30, 40, 50, 25}
	for _, v := range values {
		tree.Insert(v)
	}

	// Duplicate insert should be no-op
	tree.Insert(30)
	tree.Insert(10)

	expectedOrder := []int{10, 20, 25, 30, 40, 50}
	if !reflect.DeepEqual(tree.InOrder(), expectedOrder) {
		t.Fatalf("InOrder got %v, want %v", tree.InOrder(), expectedOrder)
	}

	for _, v := range values {
		if !tree.Search(v) {
			t.Fatalf("expected Search(%d) to be true", v)
		}
	}

	for _, missing := range []int{5, 15, 35, 60, 100} {
		if tree.Search(missing) {
			t.Fatalf("expected Search(%d) to be false", missing)
		}
	}

	if _, ok := isAVL(t, tree.Root); !ok {
		t.Fatalf("tree failed AVL balance property after [10, 20, 30, 40, 50, 25]")
	}
}

func TestAVLBalancePropertySequentialAndRandom(t *testing.T) {
	// 1. Sequential ascending inserts (triggers Right-Right / Left rotations)
	treeAsc := New()
	for i := 1; i <= 100; i++ {
		treeAsc.Insert(i)
	}
	if _, ok := isAVL(t, treeAsc.Root); !ok {
		t.Fatalf("ascending sequential inserts failed AVL invariant")
	}
	if treeAsc.Root.Height > 8 {
		t.Fatalf("tree height %d exceeds O(log n) bound for 100 nodes", treeAsc.Root.Height)
	}

	// 2. Sequential descending inserts (triggers Left-Left / Right rotations)
	treeDesc := New()
	for i := 100; i >= 1; i-- {
		treeDesc.Insert(i)
	}
	if _, ok := isAVL(t, treeDesc.Root); !ok {
		t.Fatalf("descending sequential inserts failed AVL invariant")
	}
	if treeDesc.Root.Height > 8 {
		t.Fatalf("tree height %d exceeds O(log n) bound for 100 nodes", treeDesc.Root.Height)
	}

	// 3. Random inserts (triggers LL, RR, LR, RL rotations)
	rng := rand.New(rand.NewSource(42))
	nums := rng.Perm(200)
	treeRand := New()
	for _, n := range nums {
		treeRand.Insert(n)
	}
	if _, ok := isAVL(t, treeRand.Root); !ok {
		t.Fatalf("random inserts failed AVL invariant")
	}

	sorted := make([]int, len(nums))
	copy(sorted, nums)
	sort.Ints(sorted)
	if !reflect.DeepEqual(treeRand.InOrder(), sorted) {
		t.Fatalf("InOrder does not match sorted elements for random inserts")
	}
}

func TestAVLLessonValidation(t *testing.T) {
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
	if meta.ID != "avl" {
		t.Errorf("meta.ID = %q; want 'avl'", meta.ID)
	}
	if meta.Name != "AVL Tree (Self-Balancing)" {
		t.Errorf("meta.Name = %q; want 'AVL Tree (Self-Balancing)'", meta.Name)
	}
	if meta.Category != catalog.CategoryTrees {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryTrees)
	}
	if meta.Renderer != "tree" {
		t.Errorf("meta.Renderer = %q; want 'tree'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(log n)" {
		t.Errorf("TimeComplexity = %q; want 'O(log n)'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n)" {
		t.Errorf("SpaceComplexity = %q; want 'O(n)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("Difficulty = %q; want 'Hard'", meta.Documentation.Difficulty)
	}

	foundBF := false
	foundRot := false
	foundInv := false
	for _, c := range meta.Documentation.KeyConcepts {
		if c == "Balance Factor [-1, 0, 1]" {
			foundBF = true
		}
		if c == "Single & Double Rotations" {
			foundRot = true
		}
		if c == "Height-Balanced Invariant" {
			foundInv = true
		}
	}
	if !foundBF || !foundRot || !foundInv {
		t.Errorf("missing key concepts in %v", meta.Documentation.KeyConcepts)
	}
}

func TestAVLTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	seenRotationExpl := false
	seenBFInVars := false

	for i, step := range steps {
		state, ok := step.State.(renderers.TreeState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}

		if step.Variables != nil {
			if _, ok := step.Variables["balanceFactor"]; ok {
				seenBFInVars = true
			} else if _, ok := step.Variables["bf"]; ok {
				seenBFInVars = true
			}
		}

		if step.Action == trace.ActionSwap {
			seenRotationExpl = true
		}

		// Validate coordinates
		nodeMap := make(map[string]renderers.TreeNode)
		for _, n := range state.Nodes {
			nodeMap[n.ID] = n
		}

		for _, n := range state.Nodes {
			if n.ParentID == "" {
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
			}
		}
	}

	if !seenRotationExpl {
		t.Errorf("trace did not exercise any rotation actions")
	}
	if !seenBFInVars {
		t.Errorf("trace steps variables did not include balance factor ('bf' or 'balanceFactor')")
	}
}

func TestAVLCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{"values": [30, 20, 10, 25, 40, 50]}`)

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

func TestAVLInputValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid default with nil
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) should succeed, got %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("Build(nil) produced 0 steps")
	}

	// Valid array input
	validArray := []byte(`[15, 10, 20]`)
	if steps, err := lesson.Build(validArray); err != nil || len(steps) == 0 {
		t.Fatalf("Build(%s) failed: %v", validArray, err)
	}

	// Valid object input
	validObj := []byte(`{"values": [15, 10, 20]}`)
	if steps, err := lesson.Build(validObj); err != nil || len(steps) == 0 {
		t.Fatalf("Build(%s) failed: %v", validObj, err)
	}

	// Invalid inputs to reject
	invalidCases := []struct {
		name  string
		input []byte
	}{
		{"empty input", []byte("")},
		{"whitespace input", []byte("   \n\t")},
		{"null literal", []byte("null")},
		{"malformed JSON", []byte("not a json string")},
		{"truncated JSON", []byte("[10, 20,")},
		{"type mismatch in array", []byte(`[10, "twenty", 30]`)},
		{"empty array", []byte("[]")},
		{"empty object", []byte("{}")},
		{"empty values array in object", []byte(`{"values": []}`)},
		{"null values field in object", []byte(`{"values": null}`)},
		{"unknown field in object", []byte(`{"values": [10], "unknown": true}`)},
		{"raw number", []byte("42")},
		{"boolean", []byte("true")},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := lesson.Build(tc.input); err == nil {
				t.Fatalf("expected Build(%s) to return an error, but it succeeded", tc.input)
			}
		})
	}
}

func TestAVLRotationCases(t *testing.T) {
	// Case 1: Left-Left (LL) -> Right Rotation
	t.Run("LL_RightRotation", func(t *testing.T) {
		steps := Trace([]int{30, 20, 10})
		seenLLCase := false
		seenRightRotation := false
		seenPivotRising := false

		for _, s := range steps {
			if s.Variables != nil {
				if s.Variables["case"] == "LL" {
					seenLLCase = true
				}
				if s.Variables["rotation"] == "right" {
					seenRightRotation = true
					if s.Variables["pivot"] == 20 && s.Variables["descending"] == 30 {
						seenPivotRising = true
					}
				}
			}
		}

		if !seenLLCase {
			t.Errorf("LL case not identified in steps")
		}
		if !seenRightRotation {
			t.Errorf("Right rotation not executed for LL case")
		}
		if !seenPivotRising {
			t.Errorf("LL right rotation should show pivot 20 rising and 30 descending")
		}

		// Verify final structure: root is 20, left child is 10, right child is 30
		lastState := steps[len(steps)-1].State.(renderers.TreeState)
		nodeMap := make(map[string]renderers.TreeNode)
		var rootNode *renderers.TreeNode
		for _, n := range lastState.Nodes {
			nodeMap[n.ID] = n
			if n.ParentID == "" {
				cp := n
				rootNode = &cp
			}
		}
		if rootNode == nil || rootNode.Val != 20 {
			t.Fatalf("final root val = %v; want 20", rootNode)
		}
	})

	// Case 2: Right-Right (RR) -> Left Rotation
	t.Run("RR_LeftRotation", func(t *testing.T) {
		steps := Trace([]int{10, 20, 30})
		seenRRCase := false
		seenLeftRotation := false
		seenPivotRising := false

		for _, s := range steps {
			if s.Variables != nil {
				if s.Variables["case"] == "RR" {
					seenRRCase = true
				}
				if s.Variables["rotation"] == "left" {
					seenLeftRotation = true
					if s.Variables["pivot"] == 20 && s.Variables["descending"] == 10 {
						seenPivotRising = true
					}
				}
			}
		}

		if !seenRRCase {
			t.Errorf("RR case not identified in steps")
		}
		if !seenLeftRotation {
			t.Errorf("Left rotation not executed for RR case")
		}
		if !seenPivotRising {
			t.Errorf("RR left rotation should show pivot 20 rising and 10 descending")
		}

		lastState := steps[len(steps)-1].State.(renderers.TreeState)
		var rootVal int
		for _, n := range lastState.Nodes {
			if n.ParentID == "" {
				rootVal = n.Val
			}
		}
		if rootVal != 20 {
			t.Fatalf("final root val = %d; want 20", rootVal)
		}
	})

	// Case 3: Left-Right (LR) -> Child rotates Left, Parent rotates Right
	t.Run("LR_DoubleRotation", func(t *testing.T) {
		steps := Trace([]int{30, 10, 20})
		seenLRCase := false
		seenLR1 := false
		seenLR2 := false

		for _, s := range steps {
			if s.Variables != nil {
				if s.Variables["case"] == "LR" {
					seenLRCase = true
				}
				if s.Variables["step"] == "LR-1" {
					seenLR1 = true
				}
				if s.Variables["step"] == "LR-2" {
					seenLR2 = true
				}
			}
		}

		if !seenLRCase {
			t.Errorf("LR case not identified in steps")
		}
		if !seenLR1 {
			t.Errorf("LR step 1 (child left rotation) not found")
		}
		if !seenLR2 {
			t.Errorf("LR step 2 (root right rotation) not found")
		}

		lastState := steps[len(steps)-1].State.(renderers.TreeState)
		var rootVal int
		for _, n := range lastState.Nodes {
			if n.ParentID == "" {
				rootVal = n.Val
			}
		}
		if rootVal != 20 {
			t.Fatalf("final root val = %d; want 20", rootVal)
		}
	})

	// Case 4: Right-Left (RL) -> Child rotates Right, Parent rotates Left
	t.Run("RL_DoubleRotation", func(t *testing.T) {
		steps := Trace([]int{10, 30, 20})
		seenRLCase := false
		seenRL1 := false
		seenRL2 := false

		for _, s := range steps {
			if s.Variables != nil {
				if s.Variables["case"] == "RL" {
					seenRLCase = true
				}
				if s.Variables["step"] == "RL-1" {
					seenRL1 = true
				}
				if s.Variables["step"] == "RL-2" {
					seenRL2 = true
				}
			}
		}

		if !seenRLCase {
			t.Errorf("RL case not identified in steps")
		}
		if !seenRL1 {
			t.Errorf("RL step 1 (child right rotation) not found")
		}
		if !seenRL2 {
			t.Errorf("RL step 2 (root left rotation) not found")
		}

		lastState := steps[len(steps)-1].State.(renderers.TreeState)
		var rootVal int
		for _, n := range lastState.Nodes {
			if n.ParentID == "" {
				rootVal = n.Val
			}
		}
		if rootVal != 20 {
			t.Fatalf("final root val = %d; want 20", rootVal)
		}
	})
}

func TestAVLSnapshotImmutability(t *testing.T) {
	steps := Trace([]int{10, 20, 30, 40, 50, 25})
	if len(steps) < 5 {
		t.Fatalf("expected at least 5 steps, got %d", len(steps))
	}

	// Deep copy original nodes from all steps
	originalNodes := make([][]renderers.TreeNode, len(steps))
	for i, s := range steps {
		st := s.State.(renderers.TreeState)
		originalNodes[i] = make([]renderers.TreeNode, len(st.Nodes))
		copy(originalNodes[i], st.Nodes)
	}

	// Mutate steps at different indices
	for i := 0; i < len(steps); i += 3 {
		st := steps[i].State.(renderers.TreeState)
		if len(st.Nodes) > 0 {
			st.Nodes[0].Val = 99999
			st.Nodes[0].ID = "mutated-id"
			st.Nodes[0].State = "mutated-state"
			st.Nodes[0].X = -888
			st.Nodes[0].Y = -888
		}
	}

	// Verify unmutated steps remain completely identical to originals
	for i := 0; i < len(steps); i++ {
		if i%3 == 0 {
			continue
		}
		st := steps[i].State.(renderers.TreeState)
		if !reflect.DeepEqual(st.Nodes, originalNodes[i]) {
			t.Fatalf("step %d snapshot was mutated by changes to step %d", i, (i/3)*3)
		}
	}
}

func TestAVLBalanceFactorCaptionsAndVariables(t *testing.T) {
	steps := Trace([]int{10, 20, 30, 40, 50, 25})

	seenFormulaInCaption := false
	seenBSTOrderExplanation := false
	seenImbalanceCaption := false

	for _, s := range steps {
		if strings.Contains(s.Explanation, "bf = left.Height") || strings.Contains(s.Explanation, "bf = 0: left.Height") {
			seenFormulaInCaption = true
		}
		if strings.Contains(s.Explanation, "preserving BST order") || strings.Contains(s.Explanation, "BST invariant preserved") {
			seenBSTOrderExplanation = true
		}
		if strings.Contains(s.Explanation, "Imbalance at critical node") {
			seenImbalanceCaption = true
		}

		if s.Variables != nil {
			if bf, ok := s.Variables["bf"]; ok {
				bfInt, ok := bf.(int)
				if !ok {
					t.Errorf("bf variable is not an int: %T", bf)
				}
				if s.Action == trace.ActionDone && s.LineID == "L16" {
					if bfInt < -1 || bfInt > 1 {
						t.Errorf("step balanced with |bf| > 1: %d", bfInt)
					}
				}
			}
		}
	}

	if !seenFormulaInCaption {
		t.Errorf("explanation did not contain balance factor calculation formula")
	}
	if !seenBSTOrderExplanation {
		t.Errorf("explanation did not contain explanation that rotation preserves BST order")
	}
	if !seenImbalanceCaption {
		t.Errorf("explanation did not contain critical node imbalance notification")
	}
}
