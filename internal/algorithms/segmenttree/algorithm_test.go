package segmenttree

import (
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
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
