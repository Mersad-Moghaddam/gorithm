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
		state, ok := step.State.(renderers.SegmentTreeLessonState)
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

func TestSegmentTreeLessonCoverAndUpdate(t *testing.T) {
	arr := []int{1, 3, 5, 7, 9, 11}
	steps := Trace(arr, 1, 4, 2, 6)
	first := steps[0].State.(renderers.SegmentTreeLessonState)
	saved := append([]int(nil), first.Values...)
	var before renderers.SegmentTreeLessonState
	for _, step := range steps {
		s := step.State.(renderers.SegmentTreeLessonState)
		if step.LineID == "L13" {
			before = s
			if !s.ResultKnown || s.Result != 24 {
				t.Fatal("wrong query result")
			}
			cover := make([]bool, len(arr))
			sum := 0
			for _, id := range s.AcceptedIDs {
				for _, node := range s.Nodes {
					if node.ID == id {
						sum += node.Sum
						for i := node.L; i <= node.R; i++ {
							if cover[i] || i < 1 || i > 4 {
								t.Fatal("overlap/outside cover")
							}
							cover[i] = true
						}
					}
				}
			}
			if sum != 24 || !reflect.DeepEqual(cover, []bool{false, true, true, true, true, false}) {
				t.Fatalf("cover %v sum %d", cover, sum)
			}
		}
		if step.LineID == "L20" && (!s.LeftKnown || !s.RightKnown || !s.ResultKnown || s.LeftResult+s.RightResult != s.Result) {
			t.Fatal("wrong subtotal")
		}
	}
	last := steps[len(steps)-1].State.(renderers.SegmentTreeLessonState)
	if !reflect.DeepEqual(last.Values, []int{1, 3, 6, 7, 9, 11}) || last.Nodes[0].Sum != 37 {
		t.Fatalf("final %+v", last)
	}
	for i, node := range last.Nodes {
		if node.Sum != before.Nodes[i].Sum && !(node.L <= 2 && 2 <= node.R) {
			t.Fatal("nonancestor changed")
		}
	}
	arr[0] = 999
	last.Values[0] = 888
	last.Nodes[0].Sum = 888
	last.PathIDs[0] = "changed"
	if !reflect.DeepEqual(first.Values, saved) || first.Nodes[0].Built {
		t.Fatal("snapshot mutated")
	}
}

func TestSegmentTreeLessonBuildKnown(t *testing.T) {
	for _, step := range Trace([]int{0}, 0, 0, 0, 0) {
		s := step.State.(renderers.SegmentTreeLessonState)
		if step.LineID == "L1" && s.Nodes[0].Built {
			t.Fatal("unbuilt zero known")
		}
		if step.LineID == "L5" && (!s.Nodes[0].Built || s.Nodes[0].Sum != 0) {
			t.Fatal("computed zero unknown")
		}
	}
}

func TestSegmentTreeLessonNormalization(t *testing.T) {
	for _, tc := range []struct {
		arr              []int
		ql, qr, idx, val int
	}{
		{[]int{0}, 0, 0, 0, 0}, {[]int{-2, 4, -6, 8, 1}, 1, 3, 4, -3}, {[]int{1, 2, 3, 4, 5}, 4, 1, 99, 6}, {[]int{1, 2, 3, 4, 5}, -3, 99, -1, 6}, {nil, 99, -2, 99, 999},
	} {
		steps := Trace(tc.arr, tc.ql, tc.qr, tc.idx, tc.val)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		s := steps[0].State.(renderers.SegmentTreeLessonState)
		native := BuildTree(s.Values)
		for _, step := range steps {
			q := step.State.(renderers.SegmentTreeLessonState)
			if step.LineID == "L13" && (!q.ResultKnown || q.Result != native.Query(q.QL, q.QR)) {
				t.Fatal("normalized query mismatch")
			}
		}
		native.Update(s.UpdateIndex, s.UpdateValue)
		last := steps[len(steps)-1].State.(renderers.SegmentTreeLessonState)
		if last.Nodes[0].Sum != native.Query(0, len(s.Values)-1) {
			t.Fatal("normalized update mismatch")
		}
	}
}

func TestSegmentTreeLessonInputBounds(t *testing.T) {
	for _, input := range []string{`null`, `{`, `[]`, `{"arr":[1],"query":[0,1]}`, `{"arr":[1],"query":[0]}`, `{"arr":[1],"update":{"index":1,"value":0}}`, `{"arr":[1000001]}`, `{"arr":[1],"update":{"index":0,"value":-1000001}}`, `{"arr":null}`, `{"arr":[1],"query":null}`, `{"arr":[1],"update":null}`, `{"arr":[1],"update":{"index":0}}`} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	for _, input := range []string{`[0]`, `{"arr":[-1000000,1000000],"query":[0,1],"update":{"index":1,"value":1000000}}`} {
		if _, err := (Lesson{}).Build([]byte(input)); err != nil {
			t.Fatal(err)
		}
	}
	data := "[0"
	for i := 1; i < 64; i++ {
		data += ",0"
	}
	data += "]"
	if _, err := (Lesson{}).Build([]byte(data)); err != nil {
		t.Fatal(err)
	}
	if _, err := (Lesson{}).Build([]byte(data[:len(data)-1] + ",0]")); err == nil {
		t.Fatal("accepted 65 values")
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
