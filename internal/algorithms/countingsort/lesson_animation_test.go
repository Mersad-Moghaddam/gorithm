package countingsort

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestCountingsortLessonSnapshots(t *testing.T) {
	steps := Trace([]int{2, 1, 2})
	type event struct {
		Line    string
		Action  trace.ActionType
		Metrics trace.Metrics
	}
	events := make([]event, len(steps))
	for i, step := range steps {
		events[i] = event{step.LineID, step.Action, step.Metrics}
	}
	data, _ := json.Marshal(events)
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); got != "f99bf8b800c4dbb1c8b628b609aa4c698492258eee31e92a0fc1d69a6d8999ba" {
		t.Fatalf("source/action/metrics changed: %s", data)
	}
	lines := map[string]bool{}
	for _, line := range (Lesson{}).Source() {
		lines[line.ID] = true
	}
	var tally, prefix, placed bool
	for _, step := range steps {
		if !lines[step.LineID] {
			t.Fatal("unknown source", step.LineID)
		}
		s := step.State.(renderers.CountingsortLessonState)
		if step.HoldMS != 0 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
			t.Fatal("unexpected operation hold", step.HoldMS)
		}
		if step.LineID == "L14" && s.Scan == 2 {
			tally = reflect.DeepEqual(s.Counts, []int{0, 1, 2})
		}
		if step.LineID == "L17" && s.CountIndex == 2 {
			prefix = reflect.DeepEqual(s.Counts, []int{0, 1, 3})
		}
		if step.LineID == "L21" && s.Scan == 2 {
			placed = reflect.DeepEqual(s.Counts, []int{0, 1, 3}) && reflect.DeepEqual(s.OutputFilled, []bool{false, false, true}) && s.OutputIDs[2] == 2
		}
		if step.LineID == "L22" && s.Scan == 2 && !reflect.DeepEqual(s.Counts, []int{0, 1, 2}) {
			t.Fatal("decrement", s.Counts)
		}
	}
	if !tally || !prefix || !placed {
		t.Fatal("missing teaching state", tally, prefix, placed)
	}
	paired := false
	pairSteps := Trace([]int{1, 2})
	for i := 0; i+1 < len(pairSteps); i++ {
		if pairSteps[i].LineID == "L8" && pairSteps[i+1].LineID == "L9" {
			paired = pairSteps[i].HoldMS == 600 && pairSteps[i+1].HoldMS == 1200
		}
	}
	if !paired {
		t.Fatal("maximum compare/commit must share 1800ms")
	}
	for _, fixture := range []struct{ input, want, ids []int }{
		{[]int{2, 1, 2}, []int{1, 2, 2}, []int{1, 0, 2}},
		{[]int{0, 2, 0}, []int{0, 0, 2}, []int{0, 2, 1}},
		{[]int{}, []int{}, []int{}}, {[]int{7}, []int{7}, []int{0}},
		{[]int{0, 0}, []int{0, 0}, []int{0, 1}}, {[]int{0, 8, 0}, []int{0, 0, 8}, []int{0, 2, 1}},
	} {
		before := cloneSlice(fixture.input)
		run := Trace(fixture.input)
		final := run[len(run)-1].State.(renderers.CountingsortLessonState)
		if !reflect.DeepEqual(fixture.input, before) || !reflect.DeepEqual(final.Values, fixture.want) || !reflect.DeepEqual(final.Values, Sort(fixture.input)) || !reflect.DeepEqual(final.OutputIDs, fixture.ids) {
			t.Fatal("final state", fixture, final)
		}
		for _, filled := range final.OutputFilled {
			if !filled {
				t.Fatal("final hole")
			}
		}
	}
	// Mutating later snapshots cannot rewrite any earlier buffer, IDs, or occupancy.
	for field := 0; field < 7; field++ {
		run := Trace([]int{2, 1, 2})
		early := run[len(run)-2].State.(renderers.CountingsortLessonState)
		later := run[len(run)-1].State.(renderers.CountingsortLessonState)
		before, _ := json.Marshal(early)
		switch field {
		case 0:
			later.Input[0] = 99
		case 1:
			later.InputIDs[0] = 99
		case 2:
			later.Counts[0] = 99
		case 3:
			later.Output[0] = 99
		case 4:
			later.OutputIDs[0] = 99
		case 5:
			later.OutputFilled[0] = false
		case 6:
			later.Values[0] = 99
		}
		after, _ := json.Marshal(early)
		if string(before) != string(after) {
			t.Fatal("aliased snapshot field", field)
		}
	}
}

func TestCountingsortLessonValidation(t *testing.T) {
	if _, err := (Lesson{}).Build(nil); err != nil {
		t.Fatal(err)
	}
	for _, input := range []string{"[]", "null", "{", "{}", "[1.5]", "[-1,2]", "[1001]"} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
	for _, n := range []int{64, 65} {
		input, _ := json.Marshal(make([]int, n))
		_, err := (Lesson{}).Build(input)
		if (err == nil) != (n == 64) {
			t.Error(fmt.Sprintf("length %d error: %v", n, err))
		}
	}
	if _, err := (Lesson{}).Build([]byte("[0,1000]")); err != nil {
		t.Fatal(err)
	}
}
