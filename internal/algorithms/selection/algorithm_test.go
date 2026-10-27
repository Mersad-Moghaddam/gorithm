package selection

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSelectionSortCorrectness(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected []int
	}{
		{
			name:     "standard input",
			input:    []int{64, 25, 12, 22, 11},
			expected: []int{11, 12, 22, 25, 64},
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

func TestSelectionSortTraceValidation(t *testing.T) {
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
}

func TestSelectionSortMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "selection" {
		t.Errorf("expected ID 'selection', got %q", meta.ID)
	}
	if meta.Name != "Selection Sort" {
		t.Errorf("expected Name 'Selection Sort', got %q", meta.Name)
	}
	if meta.Category != catalog.CategorySorting {
		t.Errorf("expected Category %q, got %q", catalog.CategorySorting, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(n²)" {
		t.Errorf("expected TimeComplexity 'O(n²)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", meta.Documentation.SpaceComplexity)
	}

	expectedConcepts := []string{"Min Element Scanning", "In-Place", "Unstable Swapping"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
}

func TestSelectionSortCustomBuild(t *testing.T) {
	lesson := Lesson{}
	customInput := json.RawMessage(`[5, 1, 4]`)
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
}

func TestSelectionSortTraceMetrics(t *testing.T) {
	// For [3, 2, 1]:
	// i=0: compares (0,1)->2, (1,2)->1; swap (0,2) -> [1, 2, 3] (2 comp, 1 swap)
	// i=1: compares (1,2)->3 > 2; no swap (1 comp, 0 swap)
	// Total comparisons: 3, total swaps: 1
	steps := Trace([]int{3, 2, 1})
	lastStep := steps[len(steps)-1]

	if lastStep.Metrics.Comparisons != 3 {
		t.Errorf("expected 3 comparisons, got %d", lastStep.Metrics.Comparisons)
	}
	if lastStep.Metrics.Swaps != 1 {
		t.Errorf("expected 1 swap, got %d", lastStep.Metrics.Swaps)
	}
}

func TestSelectionSortEdgeCaseTraces(t *testing.T) {
	lesson := Lesson{}
	for _, input := range [][]int{nil, {}, {42}} {
		steps := Trace(input)
		if len(steps) == 0 {
			t.Fatalf("expected steps for input %v", input)
		}
		if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
			t.Fatalf("validation failed for input %v: %v", input, err)
		}
	}
}

func TestSelectionSortRegistration(t *testing.T) {
	l, ok := catalog.Get("selection")
	if !ok {
		t.Fatal("lesson 'selection' not registered in catalog")
	}
	if l.Meta().ID != "selection" {
		t.Errorf("expected ID 'selection', got %q", l.Meta().ID)
	}
}

func TestSelectionTeachingDecisionsAndIdentity(t *testing.T) {
	steps := Trace([]int{2, 1, 1})
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatal(err)
	}
	var compared, rejected, chosen, swapped, settled bool
	for _, step := range steps {
		state := step.State.(renderers.ArrayState)
		unique := make(map[string]bool)
		for _, id := range state.IDs {
			if id == "" || unique[id] {
				t.Fatalf("objects need distinct nonempty IDs: %+v", state.IDs)
			}
			unique[id] = true
		}
		if len(unique) != 3 {
			t.Fatalf("objects need distinct IDs: %+v", state.IDs)
		}
		if step.LineID == "L6" && step.Action == trace.ActionCompare && reflect.DeepEqual(state.Comparing, []int{2}) && state.Current == 1 {
			compared = true
		}
		if step.LineID == "L6" && step.Action == trace.ActionHighlight && reflect.DeepEqual(state.Comparing, []int{2}) && state.Current == 1 {
			rejected = true // equal values do not replace candidate
		}
		if step.LineID == "L7" && state.Current == 1 && len(state.Sorted) == 0 {
			chosen = true
		}
		if step.LineID == "L10" && step.Action == trace.ActionSwap && reflect.DeepEqual(state.Values, []int{1, 2, 1}) && reflect.DeepEqual(state.IDs, []string{"selection-1", "selection-0", "selection-2"}) {
			swapped = true
		}
		if step.LineID == "L11" && reflect.DeepEqual(state.Sorted, []int{0}) {
			settled = true
		}
	}
	if !compared || !rejected || !chosen || !swapped || !settled {
		t.Fatalf("missing teaching states: compare=%v reject=%v choose=%v swap=%v settle=%v", compared, rejected, chosen, swapped, settled)
	}
	first := steps[0].State.(renderers.ArrayState)
	last := steps[len(steps)-1].State.(renderers.ArrayState)
	if !reflect.DeepEqual(first.Values, []int{2, 1, 1}) || !reflect.DeepEqual(first.IDs, []string{"selection-0", "selection-1", "selection-2"}) || !reflect.DeepEqual(last.Values, []int{1, 1, 2}) || !reflect.DeepEqual(last.Sorted, []int{0, 1, 2}) {
		t.Fatalf("history or final state wrong: first=%+v final=%+v", first, last)
	}
	// Mutating a later snapshot must not affect any earlier or intermediate snapshot.
	var intermediate renderers.ArrayState
	for _, step := range steps {
		state := step.State.(renderers.ArrayState)
		if step.LineID == "L11" && reflect.DeepEqual(state.Sorted, []int{0}) {
			intermediate = state
			break
		}
	}
	if len(intermediate.Values) == 0 {
		t.Fatal("missing intermediate snapshot")
	}
	last.Values[0], last.IDs[0], last.Sorted[0] = 999, "corrupted", 999
	if !reflect.DeepEqual(first.Values, []int{2, 1, 1}) || !reflect.DeepEqual(first.IDs, []string{"selection-0", "selection-1", "selection-2"}) || !reflect.DeepEqual(intermediate.Values, []int{1, 2, 1}) || !reflect.DeepEqual(intermediate.IDs, []string{"selection-1", "selection-0", "selection-2"}) || !reflect.DeepEqual(intermediate.Sorted, []int{0}) {
		t.Fatalf("mutating final snapshot changed history: first=%+v intermediate=%+v", first, intermediate)
	}
	for _, input := range [][]int{nil, {}, {7}} {
		history := Trace(input)
		end := history[len(history)-1].State.(renderers.ArrayState)
		if !reflect.DeepEqual(end.Values, input) && !(len(input) == 0 && len(end.Values) == 0) {
			t.Fatalf("edge input %v: final %v", input, end.Values)
		}
		if len(end.IDs) != len(input) || len(end.Sorted) != len(input) {
			t.Fatalf("edge input %v: ids=%v sorted=%v", input, end.IDs, end.Sorted)
		}
	}
}

func TestSelectionBuildInput(t *testing.T) {
	lesson := Lesson{}
	for _, input := range []json.RawMessage{[]byte(`not json`), []byte(`{"bad":1}`), []byte(`[1,"two"]`), []byte(`null`)} {
		if _, err := lesson.Build(input); err == nil {
			t.Fatalf("invalid input %s accepted", input)
		}
	}
	steps, err := lesson.Build(json.RawMessage(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if got := steps[len(steps)-1].State.(renderers.ArrayState).Values; len(got) != 0 {
		t.Fatalf("empty input became %v", got)
	}
}

func TestSelectionComparisonVisualRoles(t *testing.T) {
	for _, step := range Trace([]int{2, 1, 1}) {
		state := step.State.(renderers.ArrayState)
		if step.LineID != "L6" || len(state.Comparing) != 1 {
			continue
		}
		svg := renderers.RenderStepSVG(step, "array", 800, 400)
		candidate := svg[strings.Index(svg, `data-key="array-`+state.IDs[state.Current]+`"`):]
		scan := svg[strings.Index(svg, `data-key="array-`+state.IDs[state.Comparing[0]]+`"`):]
		if !strings.Contains(strings.SplitN(candidate, "</g>", 2)[0], `fill="url(#bar-current)"`) || !strings.Contains(strings.SplitN(scan, "</g>", 2)[0], `fill="url(#bar-comp)"`) {
			t.Fatalf("candidate and scan must have distinct visual cues at %s: %s", step.Action, svg)
		}
	}
}
