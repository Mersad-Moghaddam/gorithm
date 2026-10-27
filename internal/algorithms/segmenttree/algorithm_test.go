package segmenttree

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSegmentTreeCorrectness(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9, 11}
	st := BuildTree(arr)

	// Range sum queries
	if got := st.Query(1, 4); got != 24 {
		t.Fatalf("Query(1, 4) = %d, want 24", got)
	}
	if got := st.Query(0, 5); got != 36 {
		t.Fatalf("Query(0, 5) = %d, want 36", got)
	}
	if got := st.Query(2, 2); got != 5 {
		t.Fatalf("Query(2, 2) = %d, want 5", got)
	}
	if got := st.Query(0, 2); got != 9 {
		t.Fatalf("Query(0, 2) = %d, want 9", got)
	}
	if got := st.Query(3, 5); got != 27 {
		t.Fatalf("Query(3, 5) = %d, want 27", got)
	}

	// Point update: index 2 changed from 5 to 6
	st.Update(2, 6)

	if got := st.Query(1, 4); got != 25 {
		t.Fatalf("after Update(2, 6), Query(1, 4) = %d, want 25", got)
	}
	if got := st.Query(2, 2); got != 6 {
		t.Fatalf("after Update(2, 6), Query(2, 2) = %d, want 6", got)
	}
	if got := st.Query(0, 5); got != 37 {
		t.Fatalf("after Update(2, 6), Query(0, 5) = %d, want 37", got)
	}

	// Another update: index 0 changed from 1 to 10
	st.Update(0, 10)
	if got := st.Query(0, 1); got != 13 {
		t.Fatalf("after Update(0, 10), Query(0, 1) = %d, want 13", got)
	}
	if got := st.Query(0, 5); got != 46 {
		t.Fatalf("after Update(0, 10), Query(0, 5) = %d, want 46", got)
	}

	// Empty array edge case
	emptyST := BuildTree([]int{})
	if got := emptyST.Query(0, 0); got != 0 {
		t.Fatalf("emptyST.Query(0, 0) = %d, want 0", got)
	}
	emptyST.Update(0, 10) // should not panic

	// Single element edge case
	singleST := BuildTree([]int{42})
	if got := singleST.Query(0, 0); got != 42 {
		t.Fatalf("singleST.Query(0, 0) = %d, want 42", got)
	}
	singleST.Update(0, 99)
	if got := singleST.Query(0, 0); got != 99 {
		t.Fatalf("after update singleST.Query(0, 0) = %d, want 99", got)
	}
}

func TestSegmentTreeLessonValidation(t *testing.T) {
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
	if meta.ID != "segmenttree" {
		t.Errorf("meta.ID = %q; want 'segmenttree'", meta.ID)
	}
	if meta.Name != "Segment Tree (Range Queries)" {
		t.Errorf("meta.Name = %q; want 'Segment Tree (Range Queries)'", meta.Name)
	}
	if meta.Category != catalog.CategoryTrees {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryTrees)
	}
	if meta.Renderer != "tree" {
		t.Errorf("meta.Renderer = %q; want 'tree'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(log n) query/update, O(n) build" {
		t.Errorf("TimeComplexity = %q; want 'O(log n) query/update, O(n) build'", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(4n)" {
		t.Errorf("SpaceComplexity = %q; want 'O(4n)'", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("Difficulty = %q; want 'Hard'", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"Range Decomposition",
		"Tree Representation of Array Intervals",
		"Logarithmic Point Updates",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("KeyConcepts = %v; want %v", meta.Documentation.KeyConcepts, expectedConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("Summary should not be empty")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("HowItWorks should not be empty")
	}
}

func TestSegmentTreeTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	seenQueryBoundaries := false
	seenTreeState := false

	for i, step := range steps {
		state, ok := step.State.(renderers.TreeState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TreeState", i, step.State)
		}
		seenTreeState = true
		if len(state.Nodes) == 0 {
			t.Fatalf("step %d: empty tree nodes", i)
		}

		if lVal, hasL := step.Variables["L"]; hasL {
			if rVal, hasR := step.Variables["R"]; hasR {
				if lVal == 1 && rVal == 4 {
					seenQueryBoundaries = true
				}
			}
		}
	}

	if !seenTreeState {
		t.Error("never saw TreeState across steps")
	}
	if !seenQueryBoundaries {
		t.Error("did not find query boundaries [L, R] = [1, 4] in step variables")
	}
}

func TestSegmentTreeCustomInput(t *testing.T) {
	lesson := Lesson{}
	customInput := []byte(`{
		"arr": [2, 4, 6, 8],
		"query": [0, 2],
		"update": {"index": 1, "value": 10}
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

func TestSegmentTreeIntervalLabels(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify all steps contain valid interval labels in form "[L..R]: val"
	for stepIdx, step := range steps {
		state, ok := step.State.(renderers.TreeState)
		if !ok {
			t.Fatalf("step %d state is %T, want TreeState", stepIdx, step.State)
		}
		for _, node := range state.Nodes {
			if node.Label == "" {
				t.Fatalf("step %d node %s has empty label", stepIdx, node.ID)
			}
			if !strings.HasPrefix(node.Label, "[") || !strings.Contains(node.Label, "..") || !strings.Contains(node.Label, "]: ") {
				t.Fatalf("step %d node %s label %q does not match pattern '[L..R]: val'", stepIdx, node.ID, node.Label)
			}
		}
	}

	// Find step L3 (Build completed) and verify root node label "[0..5]: 36"
	foundL3 := false
	for _, step := range steps {
		if step.LineID == "L3" {
			foundL3 = true
			state := step.State.(renderers.TreeState)
			if state.Nodes[0].Label != "[0..5]: 36" {
				t.Errorf("step L3 root label = %q, want '[0..5]: 36'", state.Nodes[0].Label)
			}
			break
		}
	}
	if !foundL3 {
		t.Error("step L3 not found")
	}

	// Find leaf node for index 2 (node-5 with range [2..2]) before and after update
	var leafBeforeUpdate, leafAfterUpdate string
	for _, step := range steps {
		if step.LineID == "L13" { // end of query phase (before update)
			state := step.State.(renderers.TreeState)
			for _, n := range state.Nodes {
				if n.ID == "node-5" {
					leafBeforeUpdate = n.Label
				}
			}
		}
		if step.LineID == "L23" { // end of update phase
			state := step.State.(renderers.TreeState)
			for _, n := range state.Nodes {
				if n.ID == "node-5" {
					leafAfterUpdate = n.Label
				}
			}
		}
	}

	if leafBeforeUpdate != "[2..2]: 5" {
		t.Errorf("leaf before update = %q, want '[2..2]: 5'", leafBeforeUpdate)
	}
	if leafAfterUpdate != "[2..2]: 6" {
		t.Errorf("leaf after update = %q, want '[2..2]: 6'", leafAfterUpdate)
	}

	// Verify root node label reflects new sum 37 after update
	for _, step := range steps {
		if step.LineID == "L23" {
			state := step.State.(renderers.TreeState)
			if state.Nodes[0].Label != "[0..5]: 37" {
				t.Errorf("root label after update = %q, want '[0..5]: 37'", state.Nodes[0].Label)
			}
		}
	}
}

func TestSegmentTreeQueryCoverage(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundCompleteOverlap := false
	foundDisjoint := false
	foundPartialOverlap := false
	var finalQueryStep *trace.Step

	for i := range steps {
		step := &steps[i]
		switch step.LineID {
		case "L15":
			foundCompleteOverlap = true
			if step.Action != trace.ActionHighlight {
				t.Errorf("L15 action = %s, want highlight", step.Action)
			}
		case "L16":
			foundDisjoint = true
		case "L17":
			foundPartialOverlap = true
		case "L13":
			finalQueryStep = step
		}
	}

	if !foundCompleteOverlap {
		t.Error("missing complete overlap step (L15)")
	}
	if !foundDisjoint {
		t.Error("missing disjoint bypass step (L16)")
	}
	if !foundPartialOverlap {
		t.Error("missing partial overlap step (L17)")
	}
	if finalQueryStep == nil {
		t.Fatal("missing final query step (L13)")
	}

	// Verify combined answer in L13
	ans, ok := finalQueryStep.Variables["combinedAnswer"].(int)
	if !ok || ans != 24 {
		t.Errorf("combinedAnswer = %v, want 24", finalQueryStep.Variables["combinedAnswer"])
	}

	// Verify selected nodes in final query step are green ("matched") and sum to 24
	state := finalQueryStep.State.(renderers.TreeState)
	matchedSum := 0
	matchedCount := 0
	for _, n := range state.Nodes {
		if n.State == "matched" {
			matchedCount++
			matchedSum += n.Val
		}
	}

	if matchedCount != 3 {
		t.Errorf("matched nodes count = %d, want 3", matchedCount)
	}
	if matchedSum != 24 {
		t.Errorf("sum of matched nodes = %d, want 24", matchedSum)
	}
}

func TestSegmentTreePointUpdate(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	foundDescent := false
	foundLeafUpdate := false
	foundAncestorRecalc := false
	foundUpdateDone := false

	for _, step := range steps {
		switch step.LineID {
		case "L27", "L28":
			foundDescent = true
		case "L25":
			foundLeafUpdate = true
			if val, ok := step.Variables["val"].(int); !ok || val != 6 {
				t.Errorf("L25 val = %v, want 6", step.Variables["val"])
			}
		case "L29":
			foundAncestorRecalc = true
		case "L23":
			foundUpdateDone = true
			if rootSum, ok := step.Variables["rootSum"].(int); !ok || rootSum != 37 {
				t.Errorf("L23 rootSum = %v, want 37", step.Variables["rootSum"])
			}
		}
	}

	if !foundDescent {
		t.Error("missing point update descent step (L27/L28)")
	}
	if !foundLeafUpdate {
		t.Error("missing leaf update step (L25)")
	}
	if !foundAncestorRecalc {
		t.Error("missing ancestor recalculate step (L29)")
	}
	if !foundUpdateDone {
		t.Error("missing update completion step (L23)")
	}
}

func TestSegmentTreeSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(steps) < 2 {
		t.Fatal("need at least 2 steps for immutability check")
	}

	// Mutate first step's first node
	state0 := steps[0].State.(renderers.TreeState)
	origVal := state0.Nodes[0].Val
	origLabel := state0.Nodes[0].Label
	origState := state0.Nodes[0].State

	state0.Nodes[0].Val = 999999
	state0.Nodes[0].Label = "MUTATED"
	state0.Nodes[0].State = "mutated"

	// Step 1 must be completely unaffected
	state1 := steps[1].State.(renderers.TreeState)
	if state1.Nodes[0].Val == 999999 {
		t.Error("mutating step 0 mutated step 1 Val: not deep-copied!")
	}
	if state1.Nodes[0].Label == "MUTATED" {
		t.Error("mutating step 0 mutated step 1 Label: not deep-copied!")
	}

	// Restore step 0
	state0.Nodes[0].Val = origVal
	state0.Nodes[0].Label = origLabel
	state0.Nodes[0].State = origState

	// Verify that early steps still retain old tree values after later update steps
	var earlyRootVal int
	foundEarly := false
	for _, step := range steps {
		if step.LineID == "L3" {
			state := step.State.(renderers.TreeState)
			earlyRootVal = state.Nodes[0].Val
			foundEarly = true
			break
		}
	}

	if !foundEarly || earlyRootVal != 36 {
		t.Errorf("early root val = %d, want 36", earlyRootVal)
	}
}

func TestSegmentTreeStrictInputValidation(t *testing.T) {
	lesson := Lesson{}

	// nil and empty byte slice should succeed (default input)
	if _, err := lesson.Build(nil); err != nil {
		t.Errorf("Build(nil) failed: %v", err)
	}
	if _, err := lesson.Build([]byte("")); err != nil {
		t.Errorf("Build('') failed: %v", err)
	}
	if _, err := lesson.Build([]byte("   \n\t  ")); err != nil {
		t.Errorf("Build(whitespace) failed: %v", err)
	}

	// null must be rejected
	if _, err := lesson.Build([]byte("null")); err == nil {
		t.Error("expected error for 'null', got nil")
	}

	// malformed JSON must be rejected
	if _, err := lesson.Build([]byte("{broken")); err == nil {
		t.Error("expected error for malformed json, got nil")
	}
	if _, err := lesson.Build([]byte("[1, 2, ")); err == nil {
		t.Error("expected error for truncated array, got nil")
	}

	// empty array must be rejected
	if _, err := lesson.Build([]byte("[]")); err == nil {
		t.Error("expected error for empty array '[]', got nil")
	}
	if _, err := lesson.Build([]byte(`{"arr": []}`)); err == nil {
		t.Error("expected error for empty arr '{\"arr\": []}', got nil")
	}

	// invalid query ranges
	invalidQueries := []string{
		`{"arr": [1, 2, 3], "query": [2, 1]}`,  // ql > qr
		`{"arr": [1, 2, 3], "query": [-1, 2]}`, // ql < 0
		`{"arr": [1, 2, 3], "query": [0, 3]}`,  // qr >= len
		`{"arr": [1, 2, 3], "query": [1]}`,     // len(query) != 2
		`{"arr": [1, 2, 3], "query": [1, 2, 3]}`,
	}
	for _, q := range invalidQueries {
		if _, err := lesson.Build([]byte(q)); err == nil {
			t.Errorf("expected error for invalid query %q, got nil", q)
		}
	}

	// invalid update indices
	invalidUpdates := []string{
		`{"arr": [1, 2, 3], "update": {"index": -1, "value": 10}}`,
		`{"arr": [1, 2, 3], "update": {"index": 3, "value": 10}}`,
	}
	for _, u := range invalidUpdates {
		if _, err := lesson.Build([]byte(u)); err == nil {
			t.Errorf("expected error for invalid update %q, got nil", u)
		}
	}

	// unexpected trailing data
	if _, err := lesson.Build([]byte(`{"arr": [1, 2]} trailing`)); err == nil {
		t.Error("expected error for trailing data, got nil")
	}

	// pure array input succeeds
	steps, err := lesson.Build([]byte(`[5, 10, 15]`))
	if err != nil {
		t.Fatalf("Build pure array failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("pure array trace has 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed for pure array: %v", err)
	}
}
