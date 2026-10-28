package selection

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestSelectionLessonSnapshots(t *testing.T) {
	input := []int{3, 1, 2}
	steps := Trace(input)
	var sequence []string
	source := map[string]bool{}
	for _, line := range (Lesson{}).Source() {
		source[line.ID] = true
	}
	for _, step := range steps {
		sequence = append(sequence, step.LineID+":"+string(step.Action))
		if !source[step.LineID] {
			t.Fatalf("unknown source %s", step.LineID)
		}
		s, ok := step.State.(renderers.SelectionLessonState)
		if !ok {
			t.Fatalf("untyped state %T", step.State)
		}
		if s.MinValue != 0 || s.MaxValue != 3 {
			t.Fatal("axis changed")
		}
		if step.HoldMS != 0 && step.HoldMS != 600 && step.HoldMS != 1200 {
			t.Fatal("invalid hold")
		}
		if step.LineID == "L10" && step.Action == trace.ActionSwap && s.MinIndex != s.Swapped[1] {
			t.Fatal("minIdx must retain actual source index")
		}
		if step.LineID == "L7" && s.Pass == 0 && (!reflect.DeepEqual(s.IDs, []int{0, 1, 2}) || s.MinIndex != 1 || s.Scan != 1) {
			t.Fatalf("minimum must move without items: %+v", s)
		}
		if step.Action == trace.ActionSwap && s.Pass == 0 && (!reflect.DeepEqual(s.Values, []int{1, 3, 2}) || !reflect.DeepEqual(s.IDs, []int{1, 0, 2})) {
			t.Fatalf("first swap %+v", s)
		}
	}
	want := "L1:init L2:set L3:highlight L4:set L6:compare L7:set L6:compare L10:swap L11:set L3:highlight L4:set L6:compare L7:set L10:swap L11:set L12:done"
	if strings.Join(sequence, " ") != want {
		t.Fatalf("sequence %v", sequence)
	}
	final := steps[len(steps)-1]
	if final.Metrics.Comparisons != 3 || final.Metrics.Swaps != 2 || final.Metrics.Operations != 5 {
		t.Fatalf("metrics %+v", final.Metrics)
	}
	for i, step := range steps {
		if step.LineID == "L6" && i+1 < len(steps) && steps[i+1].LineID == "L7" && step.HoldMS+steps[i+1].HoldMS != 1800 {
			t.Fatal("compare/update budget")
		}
	}
	if !reflect.DeepEqual(input, []int{3, 1, 2}) {
		t.Fatal("input mutated")
	}
	for _, tc := range []struct{ input, values, ids []int }{{[]int{}, []int{}, []int{}}, {[]int{7}, []int{7}, []int{0}}, {[]int{-1, 0, -2}, []int{-2, -1, 0}, []int{2, 0, 1}}, {[]int{2, 2}, []int{2, 2}, []int{0, 1}}, {[]int{2, 2, 1}, []int{1, 2, 2}, []int{2, 1, 0}}} {
		ss := Trace(tc.input)
		s := ss[len(ss)-1].State.(renderers.SelectionLessonState)
		if !reflect.DeepEqual(s.Values, tc.values) || !reflect.DeepEqual(s.Values, Sort(tc.input)) || !reflect.DeepEqual(s.IDs, tc.ids) || s.SortedPrefix != len(tc.input) {
			t.Fatalf("terminal %+v", s)
		}
	}
	for _, step := range Trace([]int{1, 2}) {
		if step.LineID == "L10" && (step.Action != trace.ActionSet || !reflect.DeepEqual(step.State.(renderers.SelectionLessonState).IDs, []int{0, 1})) {
			t.Fatal("fake swap")
		}
	}
	first := steps[0].State.(renderers.SelectionLessonState)
	last := final.State.(renderers.SelectionLessonState)
	last.Values[0] = 99
	last.IDs[0] = 99
	last.Sorted[0] = 99
	if !reflect.DeepEqual(first.Values, input) || !reflect.DeepEqual(first.IDs, []int{0, 1, 2}) {
		t.Fatal("aliased snapshots")
	}
}

func TestSelectionBuildValidation(t *testing.T) {
	for _, input := range []string{"[null]", "[2,null,1]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted null element in %s", input)
		}
	}
	if _, err := (Lesson{}).Build(json.RawMessage("[0]")); err != nil {
		t.Fatalf("zero is an integer: %v", err)
	}
	for _, input := range []string{"[]", "null", "{}", "broken", "[1.5]", "[1000001]", "[-1000001]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Fatalf("accepted %s", input)
		}
	}
	if _, err := (Lesson{}).Build(nil); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{64, 65} {
		data, _ := json.Marshal(make([]int, n))
		_, err := (Lesson{}).Build(data)
		if (err == nil) != (n == 64) {
			t.Fatalf("length %d: %v", n, err)
		}
	}
	if _, err := (Lesson{}).Build(json.RawMessage("[-1000000,0,1000000]")); err != nil {
		t.Fatal(err)
	}
}

func TestSelectionScanEndsBeforeSwap(t *testing.T) {
	for _, step := range Trace([]int{3, 1, 2}) {
		if step.LineID == "L10" && step.State.(renderers.SelectionLessonState).Scan != -1 {
			t.Fatal("scan index remains visible after inner loop ends")
		}
	}
}
