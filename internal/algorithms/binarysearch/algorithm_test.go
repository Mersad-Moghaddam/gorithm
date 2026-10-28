package binarysearch

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
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
			name:     "found in odd length middle",
			arr:      []int{1, 3, 5, 7, 9},
			target:   5,
			expected: 2,
		},
		{
			name:     "found in odd length left",
			arr:      []int{1, 3, 5, 7, 9},
			target:   1,
			expected: 0,
		},
		{
			name:     "found in odd length right",
			arr:      []int{1, 3, 5, 7, 9},
			target:   9,
			expected: 4,
		},
		{
			name:     "found in even length left half",
			arr:      []int{2, 4, 6, 8, 10, 12},
			target:   4,
			expected: 1,
		},
		{
			name:     "found in even length right half",
			arr:      []int{2, 4, 6, 8, 10, 12},
			target:   10,
			expected: 4,
		},
		{
			name:     "not found smaller than min",
			arr:      []int{2, 4, 6, 8, 10},
			target:   1,
			expected: -1,
		},
		{
			name:     "not found larger than max",
			arr:      []int{2, 4, 6, 8, 10},
			target:   15,
			expected: -1,
		},
		{
			name:     "not found between elements",
			arr:      []int{2, 4, 6, 8, 10},
			target:   5,
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
		{
			name:     "empty slice",
			arr:      []int{},
			target:   10,
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Search(tt.arr, tt.target)
			if got != tt.expected {
				t.Fatalf("Search(%v, %d) = %d, expected %d", tt.arr, tt.target, got, tt.expected)
			}
		})
	}
}

func TestBinarySearchLessonBounds(t *testing.T) {
	steps, _ := (Lesson{}).Build(nil)
	var bounds, discards [][2]int
	var mids []int
	for _, step := range steps {
		s, ok := step.State.(renderers.BinarySearchLessonState)
		if !ok {
			t.Fatalf("untyped state at %s", step.LineID)
		}
		if step.LineID == "L4" {
			bounds = append(bounds, [2]int{s.Low, s.High})
			mids = append(mids, s.Mid)
		}
		if s.Phase == "discard" {
			discards = append(discards, [2]int{s.DiscardLow, s.DiscardHigh})
			if s.Mid != -1 {
				t.Fatal("stale midpoint")
			}
		}
		if !reflect.DeepEqual(s.Values, []int{3, 8, 14, 22, 35, 47, 59, 71, 88}) {
			t.Fatal("values rearranged")
		}
	}
	if !reflect.DeepEqual(bounds, [][2]int{{0, 8}, {5, 8}, {5, 5}}) || !reflect.DeepEqual(mids, []int{4, 6, 5}) || !reflect.DeepEqual(discards, [][2]int{{0, 4}, {6, 8}}) {
		t.Fatalf("bounds %v mids %v discards %v", bounds, mids, discards)
	}
	last := steps[len(steps)-1]
	if last.Metrics.Comparisons != 3 || last.State.(renderers.BinarySearchLessonState).Result != 5 {
		t.Fatal("wrong result/metrics")
	}
	for i, step := range steps {
		if step.LineID == "L7" || step.LineID == "L9" || step.LineID == "L6" {
			if step.LineID == "L6" {
				i--
			}
			if steps[i].HoldMS != 600 || steps[i+1].HoldMS != 1200 {
				t.Fatal("compare/commit budget must total 1800ms")
			}
		}
	}
}

func TestBinarySearchLessonBoundaryCases(t *testing.T) {
	for _, fixture := range []struct {
		a      []int
		target int
	}{{nil, 0}, {[]int{2}, 2}, {[]int{2}, 1}, {[]int{2, 4}, 0}, {[]int{2, 4}, 5}, {[]int{-8, -3, 0}, -3}, {[]int{2, 2, 2}, 2}, {[]int{3, 1, 2}, 1}} {
		steps := Trace(fixture.a, fixture.target)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		for _, step := range steps {
			s := step.State.(renderers.BinarySearchLessonState)
			if s.Mid != -1 && (s.Mid < s.Low || s.Mid > s.High) {
				t.Fatal("mid outside interval")
			}
			if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
				t.Fatal("invalid hold")
			}
		}
		last := steps[len(steps)-1].State.(renderers.BinarySearchLessonState)
		if !last.ResultKnown || last.Result != Search(fixture.a, fixture.target) {
			t.Fatal("result parity")
		}
		if len(fixture.a) == 3 && fixture.a[0] == 3 && last.SortedInput {
			t.Fatal("unsorted claimed sorted")
		}
	}
	input := []int{1, 2, 3}
	steps := Trace(input, 3)
	before := steps[0].State.(renderers.BinarySearchLessonState).Values
	input[0] = 99
	steps[len(steps)-1].State.(renderers.BinarySearchLessonState).Values[0] = 88
	if before[0] != 1 {
		t.Fatal("aliased snapshot")
	}
}

func TestBinarySearchBuildValidation(t *testing.T) {
	for _, input := range []string{`null`, `{`, `true`, `{"array":null}`, `{"arr":null}`, `{"target":null}`, `{"array":[3,1,2]}`, `{"array":[1000001]}`, `{"array":[-1000001]}`, `{"target":1000001}`, `{"target":-1000001}`, `{"array":["1"]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	for _, n := range []int{128, 129} {
		a := make([]int, n)
		raw, _ := json.Marshal(a)
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 128) {
			t.Fatalf("length %d: %v", n, err)
		}
	}
	for _, input := range []string{`{"target":0}`, `{"array":[-1000000,1000000],"target":1000000}`, `{"arr":[0],"target":0}`, `[]`, `{}`} {
		steps, err := (Lesson{}).Build(json.RawMessage(input))
		if err != nil {
			t.Fatal(err)
		}
		if input == `{"target":0}` && steps[0].State.(renderers.BinarySearchLessonState).Target != 0 {
			t.Fatal("zero target lost")
		}
	}
}

func TestBinarySearchMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "binarysearch" {
		t.Errorf("expected ID 'binarysearch', got %q", meta.ID)
	}
	if meta.Name != "Binary Search" {
		t.Errorf("expected Name 'Binary Search', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySearching {
		t.Errorf("expected Category %q, got %q", catalog.CategorySearching, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(log n)" {
		t.Errorf("expected TimeComplexity 'O(log n)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Divide and Conquer", "Sorted Precondition", "Range Bisection", "Logarithmic Search"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}
}

func TestBinarySearchTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build default trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed for default build: %v", err)
	}
}

func TestBinarySearchCustomBuild(t *testing.T) {
	lesson := Lesson{}

	// Test with JSON object containing array and target
	customInput := json.RawMessage(`{"array": [10, 20, 30, 40, 50], "target": 30}`)
	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	// Test with plain array JSON
	arrayInput := json.RawMessage(`[1, 2, 3, 4]`)
	steps2, err := lesson.Build(arrayInput)
	if err != nil {
		t.Fatalf("failed to build trace with array input: %v", err)
	}
	if len(steps2) == 0 {
		t.Fatal("trace produced 0 steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps2); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}
}

func TestBinarySearchEdgeCaseTraces(t *testing.T) {
	lesson := Lesson{}
	cases := []struct {
		arr    []int
		target int
	}{
		{nil, 5},
		{[]int{}, 5},
		{[]int{10}, 10},
		{[]int{10}, 5},
		{[]int{1, 2, 3, 4, 5, 6}, 1},
		{[]int{1, 2, 3, 4, 5, 6}, 6},
		{[]int{1, 2, 3, 4, 5, 6}, 7},
	}

	for _, c := range cases {
		steps := Trace(c.arr, c.target)
		if len(steps) == 0 {
			t.Fatalf("expected steps for input %v, target %d", c.arr, c.target)
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed for input %v, target %d: %v", c.arr, c.target, err)
		}
	}
}

func TestBinarySearchRegistration(t *testing.T) {
	l, ok := catalog.Get("binarysearch")
	if !ok {
		t.Fatal("lesson 'binarysearch' not registered in catalog")
	}
	if l.Meta().ID != "binarysearch" {
		t.Errorf("expected ID 'binarysearch', got %q", l.Meta().ID)
	}
}

func TestBinarySearchDefaultBuildValues(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) error: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != "done" {
		t.Fatalf("expected last step action 'done', got %q", lastStep.Action)
	}
	if lastStep.Variables["result"] != 5 {
		t.Errorf("expected target 47 found at index 5, got %v", lastStep.Variables["result"])
	}
}

func TestBinarySearchTracePointersAndState(t *testing.T) {
	steps := Trace([]int{10, 20, 30, 40, 50}, 40)
	var foundMid, foundPointers bool
	for _, s := range steps {
		if s.LineID == "L4" && !foundMid {
			foundMid = true
			if mid, ok := s.Variables["mid"].(int); !ok || mid != 2 {
				t.Errorf("expected initial mid = 2, got %v", s.Variables["mid"])
			}
		}
		if s.LineID == "L5" {
			if _, ok := s.Variables["low"]; ok {
				if _, ok := s.Variables["high"]; ok {
					foundPointers = true
				}
			}
		}
	}
	if !foundMid {
		t.Error("expected mid calculation step L4 in trace")
	}
	if !foundPointers {
		t.Error("expected low and high pointers in trace variables")
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Variables["found"] != true {
		t.Errorf("expected found == true, got %v", lastStep.Variables["found"])
	}
	if lastStep.Variables["result"] != 3 {
		t.Errorf("expected result index 3, got %v", lastStep.Variables["result"])
	}
}
