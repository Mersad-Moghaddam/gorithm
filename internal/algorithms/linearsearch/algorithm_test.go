package linearsearch

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSearchCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		arr      []int
		target   int
		expected int
	}{
		{
			name:     "found in middle",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   33,
			expected: 4,
		},
		{
			name:     "found at start",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   15,
			expected: 0,
		},
		{
			name:     "found at end",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   55,
			expected: 6,
		},
		{
			name:     "not found",
			arr:      []int{15, 28, 42, 8, 33, 91, 55},
			target:   999,
			expected: -1,
		},
		{
			name:     "empty array",
			arr:      []int{},
			target:   10,
			expected: -1,
		},
		{
			name:     "single element found",
			arr:      []int{42},
			target:   42,
			expected: 0,
		},
		{
			name:     "single element not found",
			arr:      []int{42},
			target:   7,
			expected: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Search(tc.arr, tc.target)
			if got != tc.expected {
				t.Fatalf("Search(%v, %d) = %d; want %d", tc.arr, tc.target, got, tc.expected)
			}
		})
	}
}

func TestLinearSearchTeachingTrace(t *testing.T) {
	lesson := Lesson{}

	// 1. Target found at start
	t.Run("target found at start", func(t *testing.T) {
		steps := Trace([]int{10, 20, 30}, 10)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("line validation failed: %v", err)
		}
		// Expect L1 (init), L2 (inspect 0), L3 (match 0), L4 (done 0)
		if len(steps) != 4 {
			t.Fatalf("expected 4 steps, got %d", len(steps))
		}
		stepL2 := steps[1]
		stateL2 := stepL2.State.(renderers.ArrayState)
		if stepL2.LineID != "L2" || stateL2.Current != 0 {
			t.Fatalf("expected L2 inspect index 0, got line %s current %d", stepL2.LineID, stateL2.Current)
		}

		stepL3 := steps[2]
		stateL3 := stepL3.State.(renderers.ArrayState)
		if stepL3.LineID != "L3" || !reflect.DeepEqual(stateL3.Comparing, []int{0}) {
			t.Fatalf("expected L3 compare index 0, got line %s comparing %v", stepL3.LineID, stateL3.Comparing)
		}

		last := steps[len(steps)-1]
		lastState := last.State.(renderers.ArrayState)
		if last.LineID != "L4" || last.Action != trace.ActionDone {
			t.Fatalf("expected L4 ActionDone, got line %s action %v", last.LineID, last.Action)
		}
		if !reflect.DeepEqual(lastState.Sorted, []int{0}) {
			t.Fatalf("expected Sorted=[0], got %v", lastState.Sorted)
		}
		expectedCaption := "Found target 10 at index 0!"
		if lastState.Caption != expectedCaption || last.Explanation != expectedCaption {
			t.Fatalf("expected caption %q, got state caption %q explanation %q", expectedCaption, lastState.Caption, last.Explanation)
		}
	})

	// 2. Target found at middle
	t.Run("target found at middle", func(t *testing.T) {
		steps := Trace([]int{10, 20, 30}, 20)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("line validation failed: %v", err)
		}
		// Expect: L1, L2(0), L3(0 mismatch), L2(1), L3(1 match), L4(done 1)
		if len(steps) != 6 {
			t.Fatalf("expected 6 steps, got %d", len(steps))
		}
		// Step 2 is mismatch at index 0
		stepMismatch := steps[2]
		stateMismatch := stepMismatch.State.(renderers.ArrayState)
		expectedMismatchCaption := "arr[0]=10 != target=20; advance to next element"
		if stepMismatch.LineID != "L3" || stepMismatch.Explanation != expectedMismatchCaption || stateMismatch.Caption != expectedMismatchCaption {
			t.Fatalf("expected mismatch caption %q, got exp %q state %q", expectedMismatchCaption, stepMismatch.Explanation, stateMismatch.Caption)
		}

		last := steps[len(steps)-1]
		lastState := last.State.(renderers.ArrayState)
		if last.LineID != "L4" || !reflect.DeepEqual(lastState.Sorted, []int{1}) {
			t.Fatalf("expected L4 with Sorted=[1], got line %s sorted %v", last.LineID, lastState.Sorted)
		}
		expectedCaption := "Found target 20 at index 1!"
		if lastState.Caption != expectedCaption || last.Explanation != expectedCaption {
			t.Fatalf("expected caption %q, got %q", expectedCaption, lastState.Caption)
		}
	})

	// 3. Target found at end
	t.Run("target found at end", func(t *testing.T) {
		steps := Trace([]int{10, 20, 30}, 30)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("line validation failed: %v", err)
		}
		last := steps[len(steps)-1]
		lastState := last.State.(renderers.ArrayState)
		if last.LineID != "L4" || !reflect.DeepEqual(lastState.Sorted, []int{2}) {
			t.Fatalf("expected L4 with Sorted=[2], got line %s sorted %v", last.LineID, lastState.Sorted)
		}
		expectedCaption := "Found target 30 at index 2!"
		if lastState.Caption != expectedCaption || last.Explanation != expectedCaption {
			t.Fatalf("expected caption %q, got %q", expectedCaption, lastState.Caption)
		}
	})

	// 4. Target not found
	t.Run("target not found", func(t *testing.T) {
		steps := Trace([]int{10, 20, 30}, 99)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("line validation failed: %v", err)
		}
		// Expect L1, then for each of 3 elements: L2, L3 (mismatch). Then L7 (done). Total = 1 + 6 + 1 = 8 steps.
		if len(steps) != 8 {
			t.Fatalf("expected 8 steps, got %d", len(steps))
		}
		last := steps[len(steps)-1]
		lastState := last.State.(renderers.ArrayState)
		if last.LineID != "L7" || last.Action != trace.ActionDone {
			t.Fatalf("expected L7 ActionDone, got line %s action %v", last.LineID, last.Action)
		}
		if len(lastState.Sorted) != 0 {
			t.Fatalf("expected empty Sorted on not found, got %v", lastState.Sorted)
		}
		expectedCaption := "Checked all 3 elements; target 99 not found. Return -1."
		if lastState.Caption != expectedCaption || last.Explanation != expectedCaption {
			t.Fatalf("expected caption %q, got state %q exp %q", expectedCaption, lastState.Caption, last.Explanation)
		}
	})

	// 5. Empty array
	t.Run("empty array", func(t *testing.T) {
		steps := Trace([]int{}, 10)
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("line validation failed on empty array: %v", err)
		}
		if len(steps) != 2 {
			t.Fatalf("expected 2 steps for empty array, got %d", len(steps))
		}
		if steps[0].LineID != "L1" {
			t.Fatalf("expected step 0 to be L1, got %s", steps[0].LineID)
		}
		last := steps[1]
		lastState := last.State.(renderers.ArrayState)
		if last.LineID != "L7" || last.Action != trace.ActionDone {
			t.Fatalf("expected L7 ActionDone, got line %s action %v", last.LineID, last.Action)
		}
		if len(lastState.Sorted) != 0 {
			t.Fatalf("expected empty Sorted, got %v", lastState.Sorted)
		}
		if !strings.Contains(lastState.Caption, "not found") {
			t.Fatalf("expected caption to indicate not found, got %q", lastState.Caption)
		}
	})
}

func TestLinearSearchStableIDsAndImmutability(t *testing.T) {
	initial := []int{15, 28, 42, 8, 33, 91, 55}
	target := 33
	steps := Trace(initial, target)

	if len(steps) == 0 {
		t.Fatal("expected steps, got 0")
	}

	// 1. Verify stable IDs across all steps
	for stepIdx, step := range steps {
		state, ok := step.State.(renderers.ArrayState)
		if !ok {
			t.Fatalf("step %d: state is not ArrayState", stepIdx)
		}
		if len(state.IDs) != len(state.Values) {
			t.Fatalf("step %d: len(IDs)=%d != len(Values)=%d", stepIdx, len(state.IDs), len(state.Values))
		}
		unique := make(map[string]bool)
		for i, id := range state.IDs {
			expectedID := fmt.Sprintf("linearsearch-%d", i)
			if id != expectedID {
				t.Fatalf("step %d: expected ID %q at index %d, got %q", stepIdx, expectedID, i, id)
			}
			if unique[id] {
				t.Fatalf("step %d: duplicate ID %q", stepIdx, id)
			}
			unique[id] = true
		}
	}

	// 2. Verify snapshot immutability
	firstState := steps[0].State.(renderers.ArrayState)
	var intermediateState renderers.ArrayState
	for _, step := range steps {
		s := step.State.(renderers.ArrayState)
		if step.LineID == "L3" && reflect.DeepEqual(s.Comparing, []int{0}) {
			intermediateState = s
			break
		}
	}
	if len(intermediateState.Values) == 0 {
		t.Fatal("missing intermediate comparison step")
	}

	lastState := steps[len(steps)-1].State.(renderers.ArrayState)

	// Mutate lastState slices
	lastState.Values[0] = 9999
	lastState.IDs[0] = "corrupted-id"
	lastState.Sorted[0] = 8888

	// Ensure firstState and intermediateState were untouched
	if firstState.Values[0] != 15 || firstState.IDs[0] != "linearsearch-0" {
		t.Fatalf("mutating last step mutated first step: Values[0]=%d, IDs[0]=%s", firstState.Values[0], firstState.IDs[0])
	}
	if intermediateState.Values[0] != 15 || intermediateState.IDs[0] != "linearsearch-0" {
		t.Fatalf("mutating last step mutated intermediate step: Values[0]=%d, IDs[0]=%s", intermediateState.Values[0], intermediateState.IDs[0])
	}

	// Verify mutating input array does not alter already generated steps
	initial[0] = -777
	step0State := steps[0].State.(renderers.ArrayState)
	if step0State.Values[0] != 15 {
		t.Fatalf("mutating initial array affected trace steps: Values[0]=%d", step0State.Values[0])
	}
}

func TestLinearSearchBuildValidation(t *testing.T) {
	lesson := Lesson{}

	// Invalid inputs that must return an error
	invalidInputs := []struct {
		name  string
		input json.RawMessage
	}{
		{name: "malformed json", input: []byte("not json")},
		{name: "json null", input: []byte("null")},
		{name: "plain json array", input: []byte("[1, 2, 3]")},
		{name: "missing target in object", input: []byte(`{"array": [1, 2, 3]}`)},
		{name: "missing target with arr key", input: []byte(`{"arr": [1, 2, 3]}`)},
		{name: "null target", input: []byte(`{"array": [1, 2, 3], "target": null}`)},
		{name: "empty array under array key", input: []byte(`{"array": [], "target": 5}`)},
		{name: "empty array under arr key", input: []byte(`{"arr": [], "target": 5}`)},
		{name: "missing array entirely", input: []byte(`{"target": 5}`)},
		{name: "null array", input: []byte(`{"array": null, "target": 5}`)},
		{name: "non-integer string in array", input: []byte(`{"array": [1, "two", 3], "target": 5}`)},
		{name: "non-integer float in array", input: []byte(`{"array": [1.5, 2], "target": 5}`)},
		{name: "non-integer string target", input: []byte(`{"array": [1, 2], "target": "five"}`)},
		{name: "non-integer float target", input: []byte(`{"array": [1, 2], "target": 2.5}`)},
	}

	for _, tc := range invalidInputs {
		t.Run("reject_"+tc.name, func(t *testing.T) {
			steps, err := lesson.Build(tc.input)
			if err == nil {
				t.Fatalf("expected error for input %s, but got %d steps", tc.input, len(steps))
			}
		})
	}

	// Valid inputs that must succeed
	t.Run("valid default nil", func(t *testing.T) {
		steps, err := lesson.Build(nil)
		if err != nil {
			t.Fatalf("unexpected error for nil: %v", err)
		}
		if len(steps) == 0 {
			t.Fatal("expected steps, got 0")
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("line validation failed: %v", err)
		}
	})

	t.Run("valid default empty byte slice", func(t *testing.T) {
		steps, err := lesson.Build([]byte(""))
		if err != nil {
			t.Fatalf("unexpected error for empty byte slice: %v", err)
		}
		if len(steps) == 0 {
			t.Fatal("expected steps, got 0")
		}
	})

	t.Run("valid custom with array key", func(t *testing.T) {
		input := []byte(`{"array": [10, 20, 30], "target": 20}`)
		steps, err := lesson.Build(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		last := steps[len(steps)-1]
		state := last.State.(renderers.ArrayState)
		if !reflect.DeepEqual(state.Sorted, []int{1}) {
			t.Fatalf("expected target found at index 1, got sorted %v", state.Sorted)
		}
	})

	t.Run("valid custom with arr key", func(t *testing.T) {
		input := []byte(`{"arr": [7, 8, 9], "target": 9}`)
		steps, err := lesson.Build(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		last := steps[len(steps)-1]
		state := last.State.(renderers.ArrayState)
		if !reflect.DeepEqual(state.Sorted, []int{2}) {
			t.Fatalf("expected target found at index 2, got sorted %v", state.Sorted)
		}
	})
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "linearsearch" {
		t.Errorf("expected ID 'linearsearch', got %q", meta.ID)
	}
	if meta.Name != "Linear Search" {
		t.Errorf("expected Name 'Linear Search', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySearching {
		t.Errorf("expected Category %q, got %q", catalog.CategorySearching, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n)" {
		t.Errorf("expected TimeComplexity 'O(n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}

	// Verify catalog registration
	registered, ok := catalog.Get("linearsearch")
	if !ok || registered == nil {
		t.Fatal("linearsearch not found in catalog registry")
	}
}
