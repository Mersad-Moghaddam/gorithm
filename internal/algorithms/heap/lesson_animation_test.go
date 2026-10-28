package heap

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestHeapLessonBaseline(t *testing.T) {
	steps := Trace([]int{2, 3, 1})
	var got []string
	for _, s := range steps {
		got = append(got, fmt.Sprintf("%s/%s/%d/%d/%d", s.LineID, s.Action, s.Metrics.Comparisons, s.Metrics.Swaps, s.Metrics.Operations))
	}
	want := "L1/init/0/0/0 L2/set/0/0/0 L3/highlight/0/0/0 L4/visit/0/0/0 L12/set/0/0/0 L13/compare/1/0/1 L14/compare/2/0/2 L15/highlight/2/0/2 L16/swap/2/1/3 L17/visit/2/1/3 L12/set/2/1/3 L6/highlight/2/1/3 L7/swap/2/2/4 L8/visit/2/2/4 L12/set/2/2/4 L13/compare/3/2/5 L15/highlight/3/2/5 L16/swap/3/3/6 L17/visit/3/3/6 L12/set/3/3/6 L6/highlight/3/3/6 L7/swap/3/4/7 L8/visit/3/4/7 L12/set/3/4/7 L10/done/3/4/7"
	if strings.Join(got, " ") != want {
		t.Fatalf("sequence or metrics changed: %s", strings.Join(got, " "))
	}
}

func TestHeapLessonSnapshots(t *testing.T) {
	source := map[string]bool{}
	for _, line := range (Lesson{}).Source() {
		source[line.ID] = true
	}
	fixtures := []struct{ input, want, ids []int }{
		{[]int{2, 3, 1}, []int{1, 2, 3}, []int{2, 0, 1}},
		{[]int{2, 2, 1}, []int{1, 2, 2}, []int{2, 1, 0}},
		{[]int{}, []int{}, []int{}}, {[]int{7}, []int{7}, []int{0}},
		{[]int{2, 1}, []int{1, 2}, []int{1, 0}}, {[]int{-1, -3, -2}, []int{-3, -2, -1}, []int{1, 2, 0}},
	}
	for _, f := range fixtures {
		original := append([]int{}, f.input...)
		steps := Trace(f.input)
		for _, step := range steps {
			if !source[step.LineID] {
				t.Fatal("unknown source", step.LineID)
			}
			s := step.State.(renderers.HeapLessonState)
			if s.HeapSize < 0 || s.HeapSize > len(f.input) {
				t.Fatal("bad heap size")
			}
			for _, child := range []int{s.LeftChild, s.RightChild} {
				if child < -1 || child >= s.HeapSize {
					t.Fatal("bad child", child)
				}
			}
			if step.HoldMS != 0 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 && step.HoldMS != 350 {
				t.Fatal("bad hold")
			}
			seen := map[int]bool{}
			for i, id := range s.IDs {
				if id < 0 || id >= len(f.input) || seen[id] || s.Values[i] != f.input[id] {
					t.Fatal("identity mismatch")
				}
				seen[id] = true
			}
		}
		final := steps[len(steps)-1].State.(renderers.HeapLessonState)
		if !reflect.DeepEqual(final.Values, f.want) && len(f.want) > 0 {
			t.Fatal("wrong final", final)
		}
		if !reflect.DeepEqual(final.Values, Sort(f.input)) && len(f.want) > 0 {
			t.Fatal("Sort discrepancy")
		}
		if !reflect.DeepEqual(final.IDs, f.ids) {
			t.Fatal("wrong final IDs", final.IDs)
		}
		if !reflect.DeepEqual(f.input, original) {
			t.Fatal("input changed")
		}
		if len(f.input) > 0 {
			first := steps[0].State.(renderers.HeapLessonState)
			saved := append([]int{}, first.Values...)
			final.Values[0] = 999
			final.IDs[0] = 999
			if !reflect.DeepEqual(first.Values, saved) || first.IDs[0] != 0 {
				t.Fatal("aliased snapshot")
			}
		}
	}
	steps := Trace([]int{2, 3, 1})
	compare := steps[5].State.(renderers.HeapLessonState)
	if compare.Largest != 0 || !reflect.DeepEqual(compare.Comparing, []int{1, 0}) {
		t.Fatal("comparison must show pre-branch candidate")
	}
	build := steps[8].State.(renderers.HeapLessonState)
	if !reflect.DeepEqual(build.Values, []int{3, 2, 1}) || !reflect.DeepEqual(build.IDs, []int{1, 0, 2}) {
		t.Fatal("build swap", build)
	}
	extract := steps[12].State.(renderers.HeapLessonState)
	if extract.HeapSize != 2 || !reflect.DeepEqual(extract.IDs, []int{2, 0, 1}) || !reflect.DeepEqual(extract.Sorted, []int{2}) {
		t.Fatal("extraction", extract)
	}
	if steps[11].HoldMS+steps[12].HoldMS != 1800 {
		t.Fatal("pair budget")
	}
}

func TestHeapLessonValidation(t *testing.T) {
	for _, input := range []string{"[]", "null", "{}", "bad", "[1000001]", "[-1000001]", "[1.5]", "[" + strings.Repeat("0,", 64) + "0]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	for _, input := range []json.RawMessage{nil, []byte("[-1000000,0,1000000]"), []byte("[" + strings.Repeat("0,", 63) + "0]")} {
		if _, err := (Lesson{}).Build(input); err != nil {
			t.Fatal(err)
		}
	}
}

func TestHeapLessonTyped(t *testing.T) {
	s := Trace([]int{2, 3, 1})[0].State
	if reflect.TypeOf(s).Name() != "HeapLessonState" {
		t.Fatalf("want HeapLessonState; got %T", s)
	}
}

func TestHeapLessonNullIntegers(t *testing.T) {
	for _, input := range []string{"[null]", "[2,null,1]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted null integer: %s", input)
		}
	}
	steps, err := (Lesson{}).Build(json.RawMessage("[0]"))
	if err != nil {
		t.Fatal(err)
	}
	if got := steps[0].State.(renderers.HeapLessonState).Values; !reflect.DeepEqual(got, []int{0}) {
		t.Fatalf("zero changed: %v", got)
	}
}
