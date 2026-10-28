package radixsort

import (
	"encoding/json"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestRadixsortLessonSnapshots(t *testing.T) {
	input := []int{21, 11, 12}
	steps := Trace(input)
	expected := []string{"L1:init:0:0", "L2:set:3:0"}
	operations, swaps := 3, 0
	for pass := 0; pass < 2; pass++ {
		operations++
		for _, id := range []string{"L3:highlight", "L4:visit", "L7:highlight", "L8:set", "L9:set"} {
			expected = append(expected, fmt.Sprintf("%s:%d:%d", id, operations, swaps))
		}
		for i := 0; i < 3; i++ {
			operations++
			expected = append(expected, fmt.Sprintf("L10:set:%d:%d", operations, swaps))
		}
		operations += 9
		expected = append(expected, fmt.Sprintf("L11:set:%d:%d", operations, swaps))
		for i := 0; i < 3; i++ {
			operations++
			for _, id := range []string{"L12:highlight", "L13:set", "L14:set", "L15:set"} {
				expected = append(expected, fmt.Sprintf("%s:%d:%d", id, operations, swaps))
			}
		}
		for i := 0; i < 3; i++ {
			swaps++
			operations++
			expected = append(expected, fmt.Sprintf("L17:swap:%d:%d", operations, swaps))
		}
		expected = append(expected, fmt.Sprintf("L5:highlight:%d:%d", operations, swaps))
	}
	expected = append(expected, "L6:done:41:6")
	if len(steps) != len(expected) {
		t.Fatalf("steps %d want %d", len(steps), len(expected))
	}
	lines := map[string]bool{}
	for _, l := range (Lesson{}).Source() {
		lines[l.ID] = true
	}
	for i, s := range steps {
		if !lines[s.LineID] || s.Metrics.Comparisons != 0 || fmt.Sprintf("%s:%s:%d:%d", s.LineID, s.Action, s.Metrics.Operations, s.Metrics.Swaps) != expected[i] {
			t.Fatalf("event %d: %+v", i, s)
		}
		if _, ok := s.State.(renderers.RadixsortLessonState); !ok {
			t.Fatalf("typed state missing: %T", s.State)
		}
	}
	var passes []renderers.RadixsortLessonState
	first := false
	for _, s := range steps {
		st := s.State.(renderers.RadixsortLessonState)
		if s.HoldMS != 0 && s.HoldMS != 350 && s.HoldMS != 600 && s.HoldMS != 1200 && s.HoldMS != 1800 {
			t.Fatalf("invalid hold %d", s.HoldMS)
		}
		if s.LineID == "L10" && st.Exp == 1 && st.Scan == 2 && (st.Counts[1] != 2 || st.Counts[2] != 1) {
			t.Fatal(st.Counts)
		}
		if s.LineID == "L14" && !first {
			first = true
			if st.Output[2] != 12 || st.OutputIDs[2] != 2 || st.Target != 2 || !st.OutputFilled[2] {
				t.Fatal(st)
			}
		}
		if s.LineID == "L5" {
			passes = append(passes, st)
		}
		if s.LineID == "L9" {
			for i, occupied := range st.OutputFilled {
				if occupied || st.OutputIDs[i] != -1 {
					t.Fatal("output not reset", st)
				}
			}
		}
	}
	if !reflect.DeepEqual(passes[0].Values, []int{21, 11, 12}) || !reflect.DeepEqual(passes[0].IDs, []int{0, 1, 2}) || !reflect.DeepEqual(passes[1].Values, []int{11, 12, 21}) || !reflect.DeepEqual(passes[1].IDs, []int{1, 2, 0}) {
		t.Fatal(passes)
	}
	if !reflect.DeepEqual(input, []int{21, 11, 12}) {
		t.Fatal("input mutated")
	}
	before, _ := json.Marshal(steps[0].State)
	later := steps[len(steps)-2].State.(renderers.RadixsortLessonState)
	later.Values[0] = 999
	later.IDs[0] = 999
	later.Output[0] = 999
	later.OutputIDs[0] = 999
	later.OutputFilled[0] = false
	after, _ := json.Marshal(steps[0].State)
	if string(before) != string(after) {
		t.Fatal("aliased snapshots")
	}
	for _, in := range [][]int{nil, {7}, {0, 0}, {0, 10, 1}, {12, 12, 2}, {170, 45, 75, 90, 802, 24, 2, 66}} {
		ss := Trace(in)
		last := ss[len(ss)-1].State.(renderers.RadixsortLessonState)
		if !reflect.DeepEqual(last.Values, Sort(in)) {
			t.Fatalf("final %v", last)
		}
		seen := map[int]bool{}
		for i, id := range last.IDs {
			if id < 0 || id >= len(in) || seen[id] || in[id] != last.Values[i] {
				t.Fatal(last)
			}
			seen[id] = true
		}
		if reflect.DeepEqual(in, []int{12, 12, 2}) && !reflect.DeepEqual(last.IDs, []int{2, 0, 1}) {
			t.Fatal(last.IDs)
		}
		if reflect.DeepEqual(in, []int{0, 0}) && len(ss) != 3 {
			t.Fatal("zero digit passes")
		}
		if reflect.DeepEqual(in, []int{170, 45, 75, 90, 802, 24, 2, 66}) {
			if !reflect.DeepEqual(last.Values, []int{2, 24, 45, 66, 75, 90, 170, 802}) {
				t.Fatal(last.Values)
			}
			for _, step := range ss {
				st := step.State.(renderers.RadixsortLessonState)
				if step.LineID == "L5" && st.Exp == 1 && !reflect.DeepEqual(st.Values, []int{170, 90, 802, 2, 24, 45, 75, 66}) {
					t.Fatal(st.Values)
				}
			}
		}
	}
}

func TestRadixsortLessonValidation(t *testing.T) {
	for _, input := range []string{"[]", "null", "{", "{}", "[1.2]", "[-1,2]", "[1000001]", "[" + strings.Repeat("0,", 64) + "0]"} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Fatalf("accepted %s", input)
		}
	}
	for _, input := range []json.RawMessage{nil, []byte("[0,1000000]"), []byte("[" + strings.Repeat("0,", 63) + "0]")} {
		if _, err := (Lesson{}).Build(input); err != nil {
			t.Fatal(err)
		}
	}
}
