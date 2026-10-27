package bucketsort

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBucketSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "default input",
			input:    []int{78, 17, 39, 26, 72, 94, 21, 12, 23, 68},
			expected: []int{12, 17, 21, 23, 26, 39, 68, 72, 78, 94},
		},
		{
			name:     "already sorted",
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "reverse sorted",
			input:    []int{5, 4, 3, 2, 1},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name:     "with duplicates",
			input:    []int{3, 1, 2, 3, 1},
			expected: []int{1, 1, 2, 3, 3},
		},
		{
			name:     "all identical",
			input:    []int{7, 7, 7, 7},
			expected: []int{7, 7, 7, 7},
		},
		{
			name:     "with negatives",
			input:    []int{-5, -1, -10, 0, 3, -2},
			expected: []int{-10, -5, -2, -1, 0, 3},
		},
		{
			name:     "single element",
			input:    []int{42},
			expected: []int{42},
		},
		{
			name:     "empty slice",
			input:    []int{},
			expected: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sort(tt.input)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestBucketSortLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "bucketsort" {
		t.Errorf("expected ID 'bucketsort', got %q", meta.ID)
	}
	if meta.Name != "Bucket Sort" {
		t.Errorf("expected Name 'Bucket Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n + k) average, O(n²) worst" {
		t.Errorf("expected TimeComplexity 'O(n + k) average, O(n²) worst', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(n + k)" {
		t.Errorf("expected SpaceComplexity 'O(n + k)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", meta.Documentation.Difficulty)
	}
	expectedConcepts := []string{"Scatter-Gather", "Uniform Distribution", "Local Insertion Sorting", "Range Hashing"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}
}

func TestBucketSortCatalogRegistration(t *testing.T) {
	l, ok := catalog.Get("bucketsort")
	if !ok {
		t.Fatal("expected 'bucketsort' to be registered in catalog")
	}
	if l.Meta().ID != "bucketsort" {
		t.Fatalf("expected ID 'bucketsort', got %q", l.Meta().ID)
	}
}

func TestBucketSortTraceValidation(t *testing.T) {
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
	if lastStep.Metrics.Operations == 0 {
		t.Error("expected non-zero operations")
	}

	for _, edgeArr := range [][]int{nil, {}, {42}, {5, 5, 5}} {
		edgeSteps := Trace(edgeArr)
		if err := catalog.ValidateSourceLineIDs(lesson, edgeSteps); err != nil {
			t.Fatalf("edge case trace validation failed: %v", err)
		}
	}
}

func TestBucketSortBuildInputValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []json.RawMessage{
		[]byte(`not json`),
		[]byte(`[1, 2`),
		[]byte(`{"bad": 1}`),
		[]byte(`42`),
		[]byte(`"string"`),
		[]byte(`true`),
		[]byte(`null`),
		[]byte(`[1, "two"]`),
		[]byte(`[1.5, 2]`),
		[]byte(`[1.0]`),
		[]byte(`[3.14]`),
		[]byte(`[1, null]`),
		[]byte(`[1, 2] trailing`),
	}

	for _, bad := range invalidInputs {
		if _, err := lesson.Build(bad); err == nil {
			t.Fatalf("expected error for invalid input %s, got nil", bad)
		}
	}

	// Valid inputs
	validEmpty, err := lesson.Build(json.RawMessage(`[]`))
	if err != nil {
		t.Fatalf("expected success for empty array, got %v", err)
	}
	if got := validEmpty[len(validEmpty)-1].State.(renderers.ArrayState).Values; len(got) != 0 {
		t.Fatalf("expected empty values, got %v", got)
	}

	validCustom, err := lesson.Build(json.RawMessage(`[5, 4, 3, 2, 1]`))
	if err != nil {
		t.Fatalf("expected success for valid custom array, got %v", err)
	}
	if got := validCustom[len(validCustom)-1].State.(renderers.ArrayState).Values; !reflect.DeepEqual(got, []int{1, 2, 3, 4, 5}) {
		t.Fatalf("expected sorted [1 2 3 4 5], got %v", got)
	}
}

func TestBucketSortStableObjectIDsAndDuplicateStability(t *testing.T) {
	input := []int{3, 1, 2, 3, 1}
	steps := Trace(input)
	if len(steps) == 0 {
		t.Fatal("expected steps, got 0")
	}

	for stepIdx, step := range steps {
		state := step.State.(renderers.ArrayState)
		if len(state.IDs) != len(input) {
			t.Fatalf("step %d (%s): expected %d IDs, got %d", stepIdx, step.LineID, len(input), len(state.IDs))
		}
		seen := make(map[string]bool)
		for _, id := range state.IDs {
			if id == "" {
				t.Fatalf("step %d (%s): found empty ID", stepIdx, step.LineID)
			}
			if !strings.HasPrefix(id, "bucketsort-") {
				t.Fatalf("step %d (%s): invalid ID prefix %q", stepIdx, step.LineID, id)
			}
			if seen[id] {
				t.Fatalf("step %d (%s): duplicate ID %q in state %v", stepIdx, step.LineID, id, state.IDs)
			}
			seen[id] = true
		}
	}

	firstState := steps[0].State.(renderers.ArrayState)
	expectedFirstIDs := []string{"bucketsort-0", "bucketsort-1", "bucketsort-2", "bucketsort-3", "bucketsort-4"}
	if !reflect.DeepEqual(firstState.IDs, expectedFirstIDs) {
		t.Fatalf("expected initial IDs %v, got %v", expectedFirstIDs, firstState.IDs)
	}

	lastState := steps[len(steps)-1].State.(renderers.ArrayState)
	expectedValues := []int{1, 1, 2, 3, 3}
	if !reflect.DeepEqual(lastState.Values, expectedValues) {
		t.Fatalf("expected sorted values %v, got %v", expectedValues, lastState.Values)
	}

	// Verify duplicate stability:
	// For value 1: bucketsort-1 must appear before bucketsort-4
	// For value 3: bucketsort-0 must appear before bucketsort-3
	id1First := lastState.IDs[0]
	id1Second := lastState.IDs[1]
	if id1First != "bucketsort-1" || id1Second != "bucketsort-4" {
		t.Fatalf("duplicate stability broken for value 1: got [%s, %s], expected [bucketsort-1, bucketsort-4]", id1First, id1Second)
	}

	id3First := lastState.IDs[3]
	id3Second := lastState.IDs[4]
	if id3First != "bucketsort-0" || id3Second != "bucketsort-3" {
		t.Fatalf("duplicate stability broken for value 3: got [%s, %s], expected [bucketsort-0, bucketsort-3]", id3First, id3Second)
	}
}

func TestBucketSortSnapshotImmutability(t *testing.T) {
	steps := Trace([]int{78, 17, 39, 26, 72})
	if len(steps) < 3 {
		t.Fatalf("expected at least 3 steps, got %d", len(steps))
	}

	first := steps[0].State.(renderers.ArrayState)
	last := steps[len(steps)-1].State.(renderers.ArrayState)

	firstValuesOrig := cloneSlice(first.Values)
	firstIDsOrig := cloneSlice(first.IDs)

	// Mutate the last state
	last.Values[0] = 99999
	last.IDs[0] = "corrupted-id"
	if len(last.Sorted) > 0 {
		last.Sorted[0] = 99999
	}

	// Verify first state unchanged
	if !reflect.DeepEqual(first.Values, firstValuesOrig) {
		t.Fatalf("first Values mutated: expected %v, got %v", firstValuesOrig, first.Values)
	}
	if !reflect.DeepEqual(first.IDs, firstIDsOrig) {
		t.Fatalf("first IDs mutated: expected %v, got %v", firstIDsOrig, first.IDs)
	}
}

func TestBucketSortEdgeCases(t *testing.T) {
	lesson := Lesson{}

	// Edge case 1: empty slice
	emptySteps := Trace([]int{})
	if len(emptySteps) == 0 {
		t.Fatal("empty input produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, emptySteps); err != nil {
		t.Fatalf("empty slice line validation failed: %v", err)
	}
	emptyEnd := emptySteps[len(emptySteps)-1].State.(renderers.ArrayState)
	if len(emptyEnd.Values) != 0 || len(emptyEnd.IDs) != 0 {
		t.Fatalf("expected empty final state, got values=%v ids=%v", emptyEnd.Values, emptyEnd.IDs)
	}

	// Edge case 2: single item
	singleSteps := Trace([]int{42})
	if len(singleSteps) == 0 {
		t.Fatal("single item produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, singleSteps); err != nil {
		t.Fatalf("single item line validation failed: %v", err)
	}
	singleEnd := singleSteps[len(singleSteps)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(singleEnd.Values, []int{42}) || !reflect.DeepEqual(singleEnd.IDs, []string{"bucketsort-0"}) {
		t.Fatalf("single item unexpected final state: values=%v ids=%v", singleEnd.Values, singleEnd.IDs)
	}

	// Edge case 3: uniform range (all identical)
	uniformSteps := Trace([]int{7, 7, 7, 7})
	if len(uniformSteps) == 0 {
		t.Fatal("uniform range produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, uniformSteps); err != nil {
		t.Fatalf("uniform range line validation failed: %v", err)
	}
	for i, s := range uniformSteps {
		st := s.State.(renderers.ArrayState)
		if len(st.IDs) != 4 {
			t.Fatalf("step %d: expected 4 IDs, got %d", i, len(st.IDs))
		}
	}
	uniformEnd := uniformSteps[len(uniformSteps)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(uniformEnd.Values, []int{7, 7, 7, 7}) {
		t.Fatalf("uniform range unexpected values: %v", uniformEnd.Values)
	}
	if !reflect.DeepEqual(uniformEnd.IDs, []string{"bucketsort-0", "bucketsort-1", "bucketsort-2", "bucketsort-3"}) {
		t.Fatalf("uniform range unexpected IDs: %v", uniformEnd.IDs)
	}

	// Edge case 4: duplicate values
	dupSteps := Trace([]int{10, 20, 10, 20})
	dupEnd := dupSteps[len(dupSteps)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(dupEnd.Values, []int{10, 10, 20, 20}) {
		t.Fatalf("duplicate values unexpected result: %v", dupEnd.Values)
	}
	if dupEnd.IDs[0] != "bucketsort-0" || dupEnd.IDs[1] != "bucketsort-2" {
		t.Fatalf("stability broken for 10: got %v", dupEnd.IDs[:2])
	}
	if dupEnd.IDs[2] != "bucketsort-1" || dupEnd.IDs[3] != "bucketsort-3" {
		t.Fatalf("stability broken for 20: got %v", dupEnd.IDs[2:])
	}
}

func TestBucketSortTeachingVisualStatesAndCaptions(t *testing.T) {
	steps := Trace([]int{78, 17, 39, 26, 72, 94, 21, 12, 23, 68})
	if len(steps) == 0 {
		t.Fatal("expected steps")
	}

	var foundRange, foundFormula, foundBucketReceive, foundSortPhase, foundGatherPhase, foundSettledPrefix bool

	for _, step := range steps {
		st := step.State.(renderers.ArrayState)
		caption := st.Caption

		if step.LineID == "L4" && strings.Contains(caption, "range") {
			foundRange = true
		}
		if step.LineID == "L8" && strings.Contains(caption, "Formula:") {
			foundFormula = true
		}
		if step.LineID == "L9" && strings.Contains(caption, "receives") {
			foundBucketReceive = true
		}
		if step.LineID == "L11" || (step.LineID == "L12" && strings.Contains(caption, "sorted")) {
			foundSortPhase = true
		}
		if step.LineID == "L15" && strings.Contains(caption, "Concatenating") {
			foundGatherPhase = true
		}
		if step.LineID == "L17" && strings.Contains(caption, "settled") {
			foundSettledPrefix = true
		}
	}

	if !foundRange {
		t.Error("missing range determination in captions")
	}
	if !foundFormula {
		t.Error("missing bucket index formula in captions")
	}
	if !foundBucketReceive {
		t.Error("missing bucket receive in captions")
	}
	if !foundSortPhase {
		t.Error("missing bucket sorting phase in captions")
	}
	if !foundGatherPhase {
		t.Error("missing gather concatenation in captions")
	}
	if !foundSettledPrefix {
		t.Error("missing settled prefix in captions")
	}
}
