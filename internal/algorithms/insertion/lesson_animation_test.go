package insertion

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestInsertionLessonSequence(t *testing.T) {
	steps := Trace([]int{3, 1, 2})
	var sequence []string
	valid := map[string]bool{}
	for _, line := range (Lesson{}).Source() {
		valid[line.ID] = true
	}
	for _, step := range steps {
		if !valid[step.LineID] {
			t.Fatal(step.LineID)
		}
		sequence = append(sequence, step.LineID+":"+string(step.Action))
	}
	wantMetrics := []trace.Metrics{{}, {}, {}, {}, {}, {Comparisons: 1, Operations: 1}, {Comparisons: 1, Swaps: 1, Operations: 2}, {Comparisons: 1, Swaps: 1, Operations: 2}, {Comparisons: 1, Swaps: 1, Operations: 2}, {Comparisons: 1, Swaps: 1, Operations: 3}, {Comparisons: 1, Swaps: 1, Operations: 3}, {Comparisons: 1, Swaps: 1, Operations: 3}, {Comparisons: 1, Swaps: 1, Operations: 3}, {Comparisons: 2, Swaps: 1, Operations: 4}, {Comparisons: 2, Swaps: 2, Operations: 5}, {Comparisons: 2, Swaps: 2, Operations: 5}, {Comparisons: 3, Swaps: 2, Operations: 6}, {Comparisons: 3, Swaps: 2, Operations: 7}, {Comparisons: 3, Swaps: 2, Operations: 7}}
	if len(steps) != len(wantMetrics) {
		t.Fatal("step count", len(steps))
	}
	for i, step := range steps {
		if step.Metrics != wantMetrics[i] {
			t.Fatalf("step %d metrics: %+v", i, step.Metrics)
		}
	}
	want := "L1:init L2:set L3:highlight L4:set L5:set L6:compare L7:swap L8:set L6:highlight L10:set L3:highlight L4:set L5:set L6:compare L7:swap L8:set L6:compare L10:set L12:done"
	if strings.Join(sequence, " ") != want {
		t.Fatal(sequence)
	}
	if got := steps[len(steps)-1].Metrics; got != (trace.Metrics{Comparisons: 3, Swaps: 2, Operations: 7}) {
		t.Fatal(got)
	}
}

func TestInsertionLessonValidation(t *testing.T) {
	for _, input := range []string{"[]", "null", "{", "{}", "[1.5]", "[1000001]", "[-1000001]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	for _, n := range []int{64, 65} {
		input, _ := json.Marshal(make([]int, n))
		_, err := (Lesson{}).Build(input)
		if (err == nil) != (n == 64) {
			t.Errorf("length %d: %v", n, err)
		}
	}
	for _, input := range []json.RawMessage{nil, []byte("[-1000000,0,1000000]")} {
		if _, err := (Lesson{}).Build(input); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInsertionLessonSnapshots(t *testing.T) {
	for _, input := range [][]int{nil, {7}, {3, 1, 2}, {2, 2, 1}, {1, 2, 3}, {3, 2, 1}, {-1, -3, 0}} {
		before := append([]int(nil), input...)
		steps := Trace(input)
		valid := map[string]bool{}
		for _, line := range (Lesson{}).Source() {
			valid[line.ID] = true
		}
		boundary := false
		for _, step := range steps {
			if !valid[step.LineID] {
				t.Fatal(step.LineID)
			}
			s := step.State.(renderers.InsertionLessonState)
			if s.MinValue != steps[0].State.(renderers.InsertionLessonState).MinValue || s.MaxValue != steps[0].State.(renderers.InsertionLessonState).MaxValue {
				t.Fatal("scale changed")
			}
			if step.LineID == "L6" && s.Scan == -1 {
				boundary = true
			}
			if fmt.Sprint(input) == "[1 2 3]" && step.LineID == "L7" {
				t.Fatal("sorted input shifted")
			}
		}
		if fmt.Sprint(input) == "[3 2 1]" && !boundary {
			t.Fatal("missing backward boundary")
		}
		var states []struct {
			renderers.ArrayState
			IDs                                          []int
			Pass, Scan, KeyValue, KeyID, Hole, PrefixEnd int
			KeyHeld                                      bool
		}
		for _, step := range steps {
			data, _ := json.Marshal(step.State)
			var state struct {
				renderers.ArrayState
				IDs                                          []int
				Pass, Scan, KeyValue, KeyID, Hole, PrefixEnd int
				KeyHeld                                      bool
			}
			if err := json.Unmarshal(data, &state); err != nil {
				t.Fatal(err)
			}
			states = append(states, state)
		}
		final := states[len(states)-1]
		if fmt.Sprint(final.Values) != fmt.Sprint(Sort(input)) {
			t.Fatalf("%v: %v", input, final.Values)
		}
		if len(final.IDs) != len(input) {
			t.Fatalf("missing identities for %v: %+v", input, final)
		}
		seen := map[int]bool{}
		for i, id := range final.IDs {
			if id < 0 || id >= len(input) || seen[id] || input[id] != final.Values[i] {
				t.Fatal(final)
			}
			seen[id] = true
		}
		if !reflect.DeepEqual(input, before) && len(input) > 0 {
			t.Fatal("input mutated")
		}
		if fmt.Sprint(input) == "[3 1 2]" {
			found := false
			inserted := false
			for i, step := range steps {
				if step.LineID == "L7" && !found {
					s := states[i]
					if fmt.Sprint(s.Values) != "[3 3 2]" || fmt.Sprint(s.IDs) != "[0 0 2]" || !s.KeyHeld || s.KeyValue != 1 || s.KeyID != 1 || s.Hole != 0 {
						t.Fatalf("shift: %+v", s)
					}
					found = true
				}
				if step.LineID == "L10" && !inserted {
					s := states[i]
					if fmt.Sprint(s.Values) != "[1 3 2]" || fmt.Sprint(s.IDs) != "[1 0 2]" || s.KeyHeld || s.Hole != -1 {
						t.Fatalf("insert: %+v", s)
					}
					inserted = true
				}
			}
			if !found || !inserted || fmt.Sprint(final.Values) != "[1 2 3]" {
				t.Fatal("missing expected operation/result")
			}
			if fmt.Sprint(final.IDs) != "[1 2 0]" {
				t.Fatal(final.IDs)
			}
		}
		if fmt.Sprint(input) == "[2 2 1]" && fmt.Sprint(final.IDs) != "[2 0 1]" {
			t.Fatal("unstable", final.IDs)
		}
	}
}

func TestInsertionLessonImmutableAndTiming(t *testing.T) {
	steps := Trace([]int{3, 1, 2})
	first := steps[0].State.(renderers.InsertionLessonState)
	last := steps[len(steps)-1].State.(renderers.InsertionLessonState)
	last.Values[0] = 99
	last.IDs[0] = 99
	last.Sorted[0] = 99
	if first.Values[0] != 3 || first.IDs[0] != 0 || first.Sorted[0] != 0 {
		t.Fatal("snapshot alias")
	}
	paired := false
	for i, step := range steps {
		switch step.HoldMS {
		case 0, 350, 600, 1200, 1800:
		default:
			t.Fatal(step.HoldMS)
		}
		if step.LineID == "L7" {
			if i == 0 || steps[i-1].HoldMS+step.HoldMS != 1800 {
				t.Fatal("copy timing")
			}
			paired = true
		}
	}
	if !paired {
		t.Fatal("missing shift")
	}
}
