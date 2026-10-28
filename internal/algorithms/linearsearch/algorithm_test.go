package linearsearch

import (
	"encoding/json"
	"reflect"
	"strings"
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

func TestLinearSearchTraceValidation(t *testing.T) {
	lesson := Lesson{}

	// Test default build
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on default build: %v", err)
	}

	// Verify found element is marked in Sorted/Green
	lastStep := steps[len(steps)-1]
	state, ok := lastStep.State.(renderers.LinearSearchLessonState)
	if !ok {
		t.Fatalf("expected state to be ArrayState, got %T", lastStep.State)
	}
	if len(state.Sorted) == 0 || state.Sorted[0] != 4 {
		t.Fatalf("expected target index 4 in Sorted, got %v", state.Sorted)
	}

	// Test not found case
	customInput, _ := json.Marshal(map[string]any{
		"arr":    []int{1, 2, 3},
		"target": 99,
	})
	notFoundSteps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("failed to build custom trace: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, notFoundSteps); err != nil {
		t.Fatalf("line validation failed on not found trace: %v", err)
	}
	lastNotFound := notFoundSteps[len(notFoundSteps)-1]
	notFoundState := lastNotFound.State.(renderers.LinearSearchLessonState)
	if len(notFoundState.Sorted) != 0 {
		t.Fatalf("expected empty Sorted on not found, got %v", notFoundState.Sorted)
	}

	// Test empty array trace
	emptySteps := Trace([]int{}, 10)
	if err := catalog.ValidateSourceLineIDs(lesson, emptySteps); err != nil {
		t.Fatalf("line validation failed on empty array: %v", err)
	}

	// Test single element found trace
	singleSteps := Trace([]int{10}, 10)
	if err := catalog.ValidateSourceLineIDs(lesson, singleSteps); err != nil {
		t.Fatalf("line validation failed on single element found: %v", err)
	}

	// Test single element not found trace
	singleNotFoundSteps := Trace([]int{10}, 20)
	if err := catalog.ValidateSourceLineIDs(lesson, singleNotFoundSteps); err != nil {
		t.Fatalf("line validation failed on single element not found: %v", err)
	}
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

func TestLinearSearchLessonFirstMatch(t *testing.T) {
	for _, tc := range []struct {
		arr                         []int
		target, result, comparisons int
	}{
		{[]int{15, 28, 42, 8, 33, 91, 55}, 33, 4, 5}, {[]int{7, 7, 7}, 7, 0, 1},
	} {
		steps := Trace(tc.arr, tc.target)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		for _, step := range steps {
			s, ok := step.State.(renderers.LinearSearchLessonState)
			if !ok {
				t.Fatalf("state %T", step.State)
			}
			if s.Phase == "scan" && (s.Compared || s.CheckedThrough != s.Index-1) {
				t.Fatalf("pending scan: %+v", s)
			}
			if s.Phase == "compare" && (!s.Compared || s.Equal != (tc.arr[s.Index] == tc.target)) {
				t.Fatalf("comparison: %+v", s)
			}
			if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
				t.Fatalf("hold %d", step.HoldMS)
			}
		}
		last := steps[len(steps)-1]
		s := last.State.(renderers.LinearSearchLessonState)
		if !s.ResultKnown || s.Result != tc.result || s.CheckedThrough != tc.result-1 || last.Metrics.Comparisons != tc.comparisons || s.Result != Search(tc.arr, tc.target) {
			t.Fatalf("final %+v, metrics %+v", s, last.Metrics)
		}
		if steps[len(steps)-2].HoldMS+last.HoldMS != 1800 {
			t.Fatal("compare/result budget")
		}
		before := append([]int(nil), steps[0].State.(renderers.LinearSearchLessonState).Values...)
		tc.arr[0] = 999
		s.Values[0] = 888
		if !reflect.DeepEqual(before, steps[0].State.(renderers.LinearSearchLessonState).Values) {
			t.Fatal("snapshot alias")
		}
	}
}

func TestLinearSearchLessonMissAndEmpty(t *testing.T) {
	for _, tc := range []struct {
		arr                                  []int
		target, result, checked, comparisons int
	}{
		{[]int{0, -2, 4}, 9, -1, 2, 3}, {nil, 9, -1, -1, 0}, {[]int{0, -2, 4}, 0, 0, -1, 1},
	} {
		steps := Trace(tc.arr, tc.target)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		for _, step := range steps {
			if _, ok := step.State.(renderers.LinearSearchLessonState); !ok {
				t.Fatalf("state %T", step.State)
			}
		}
		last := steps[len(steps)-1]
		s := last.State.(renderers.LinearSearchLessonState)
		if !s.ResultKnown || s.Result != tc.result || s.CheckedThrough != tc.checked || last.Metrics.Comparisons != tc.comparisons {
			t.Fatalf("final %+v %+v", s, last.Metrics)
		}
	}
}

func TestLinearSearchLessonBuildValidation(t *testing.T) {
	steps, err := (Lesson{}).Build(json.RawMessage(`{"target":0}`))
	if err != nil {
		t.Fatal(err)
	}
	s := steps[0].State.(renderers.LinearSearchLessonState)
	if s.Target != 0 || len(s.Values) != 7 {
		t.Fatalf("target/default %+v", s)
	}
	for _, input := range []string{`{`, `null`, `"x"`, `{"arr":null}`, `{"target":null}`, `{"arr":"x"}`, `[1000001]`, `[-1000001]`, `{"target":1000001}`, `{"target":-1000001}`, `[` + strings.Repeat("0,", 128) + `0]`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	for _, input := range []string{`[]`, `{}`, `{"array":[0],"target":0}`, `[-1000000,1000000]`, `{"target":-1000000}`, `{"target":1000000}`, `[` + strings.Repeat("0,", 127) + `0]`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err != nil {
			t.Errorf("rejected %s: %v", input, err)
		}
	}
}

func TestLinearSearchLessonRejectsNullIntegersAndFoldedFields(t *testing.T) {
	for _, input := range []string{
		`[null]`, `{"arr":[null],"target":0}`, `{"array":[null]}`,
		`{"arr":[0],"array":[null]}`, `{"arr":[null],"array":[0]}`,
		`{"ARR":null}`, `{"ARRAY":null}`, `{"TARGET":null}`,
		`{"aRr":[null]}`, `{"ArRaY":[null]}`,
		`{"arr":[1],"ARR":null}`, `{"ARR":null,"arr":[1]}`,
		`{"target":0,"TARGET":null}`, `{"TARGET":null,"target":0}`,
		`{"arr":[0],"ArRaY":null}`, `{"arr":[0],"ArRaY":[null]}`,
		`{"arr":[null],"ARR":[0]}`, `{"array":[null],"ARRAY":[0]}`,
		`{"arr":[null],"arr":[0]}`, `{"target":null,"target":0}`,
	} {
		t.Run(input, func(t *testing.T) {
			if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
				t.Fatalf("accepted null in known input: %s", input)
			}
		})
	}
}

func TestLinearSearchLessonAcceptedPayloadSemantics(t *testing.T) {
	defaults := []int{15, 28, 42, 8, 33, 91, 55}
	for _, tc := range []struct { input string; values []int; target int }{
		{`{}`, defaults, 33}, {`[]`, defaults, 33},
		{`{"arr":[],"array":[]}`, defaults, 33}, {`{"TARGET":0}`, defaults, 0},
		{`[0,-2,4]`, []int{0,-2,4}, 33},
		{`{"array":[0],"target":0}`, []int{0}, 0},
		{`{"ARR":[0],"ArRaY":[7],"TaRgEt":0}`, []int{0}, 0},
		{`{"ARR":[],"ARRAY":[7],"TARGET":7}`, []int{7}, 7},
		{`{"arr":[1],"ARR":[0],"target":1,"TARGET":0}`, []int{0}, 0},
		{`{"ARR":[1],"arr":[0],"TARGET":1,"target":0}`, []int{0}, 0},
		{`{"arr":[1],"arr":[0],"target":0}`, []int{0}, 0},
		{`{"arr":[0],"unknown":null,"target":0}`, []int{0}, 0},
	} {
		t.Run(tc.input,func(t *testing.T) {
			steps, err := (Lesson{}).Build(json.RawMessage(tc.input))
			if err != nil { t.Fatal(err) }
			first := steps[0].State.(renderers.LinearSearchLessonState)
			last := steps[len(steps)-1].State.(renderers.LinearSearchLessonState)
			if !reflect.DeepEqual(first.Values,tc.values) || first.Target != tc.target || !last.ResultKnown || last.Result != Search(tc.values,tc.target) {
				t.Fatalf("decoded values=%v target=%d result=%d; want values=%v target=%d result=%d",first.Values,first.Target,last.Result,tc.values,tc.target,Search(tc.values,tc.target))
			}
		})
	}
}
