package quick

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestQuickLessonSnapshots(t *testing.T) {
	steps := Trace([]int{3, 1, 2})
	wantLines := "L1 L2 L3 L8 L9 L10 L11 L12 L11 L12 L13 L14 L17 L18 L4 L2 L6 L5 L2 L6 L7"
	wantActions := "init highlight highlight highlight set set highlight compare highlight compare set swap swap set highlight highlight set highlight highlight set done"
	var lines, actions []string
	for _, s := range steps {
		lines = append(lines, s.LineID)
		actions = append(actions, string(s.Action))
	}
	if strings.Join(lines, " ") != wantLines || strings.Join(actions, " ") != wantActions {
		t.Fatal("changed source/action sequence", lines, actions)
	}
	for i, s := range steps {
		c, swaps := 0, 0
		if i >= 7 {
			c = 1
		}
		if i >= 9 {
			c = 2
		}
		if i >= 11 {
			swaps = 1
		}
		if i >= 12 {
			swaps = 2
		}
		if s.Metrics.Comparisons != c || s.Metrics.Swaps != swaps || s.Metrics.Operations != c+swaps {
			t.Fatalf("metrics frame %d: %+v", i, s.Metrics)
		}
	}
	firstSwap := steps[11].State.(renderers.QuickLessonState)
	if !reflect.DeepEqual(firstSwap.IDs, []int{1, 0, 2}) || firstSwap.Boundary != 0 || firstSwap.Scan != 1 || firstSwap.PivotID != 2 || firstSwap.PivotIndex != 2 {
		t.Fatal(firstSwap)
	}
	pivot := steps[12].State.(renderers.QuickLessonState)
	if !reflect.DeepEqual(pivot.IDs, []int{1, 2, 0}) || pivot.PivotIndex != 1 {
		t.Fatal(pivot)
	}
	if !reflect.DeepEqual(pivot.Values, []int{1, 2, 3}) {
		t.Fatal(pivot.Values)
	}
	base := steps[15].State.(renderers.QuickLessonState)
	if base.PivotActive || base.PivotID != -1 || base.Scan != -1 || len(base.Stack) != 2 || base.Low != 0 || base.High != 0 {
		t.Fatal("recursive call retained pivot", base)
	}
	if steps[10].HoldMS+steps[11].HoldMS != 1800 {
		t.Fatal("pair timing")
	}
	for _, fixture := range [][]int{nil, {7}, {2, 2, 2}, {2, 2, 1}, {-1, -3, 0}, {1, 2}, {3, 2, 1}, {0, 0, 0}} {
		input := append([]int{}, fixture...)
		frames := Trace(input)
		end := frames[len(frames)-1].State.(renderers.QuickLessonState)
		if !reflect.DeepEqual(end.Values, Sort(input)) {
			t.Fatal(end.Values)
		}
		if !reflect.DeepEqual(input, append([]int{}, fixture...)) {
			t.Fatal("input changed")
		}
		seen := map[int]bool{}
		for i, id := range end.IDs {
			if seen[id] || id < 0 || id >= len(input) || end.Values[i] != input[id] {
				t.Fatal("identity", end)
			}
			seen[id] = true
		}
		if reflect.DeepEqual(input, []int{2, 2, 2}) || reflect.DeepEqual(input, []int{2, 2, 1}) {
			if !reflect.DeepEqual(end.IDs, []int{2, 0, 1}) {
				t.Fatal(end.IDs)
			}
		}
		valid := map[string]bool{}
		for _, line := range (Lesson{}).Source() {
			valid[line.ID] = true
		}
		for _, frame := range frames {
			if !valid[frame.LineID] {
				t.Fatal(frame.LineID)
			}
			if frame.HoldMS != 0 && frame.HoldMS != 600 && frame.HoldMS != 1200 {
				t.Fatal(frame.HoldMS)
			}
		}
	}
	last := steps[len(steps)-1].State.(renderers.QuickLessonState)
	last.Values[0] = 99
	last.IDs[0] = 99
	firstSwap.Stack[0].Low = 99
	initial := steps[0].State.(renderers.QuickLessonState)
	if initial.Values[0] != 3 || initial.IDs[0] != 0 || steps[10].State.(renderers.QuickLessonState).Stack[0].Low != 0 {
		t.Fatal("aliased snapshot")
	}
	self := 0
	for _, frame := range Trace([]int{1, 2}) {
		if frame.Action == "swap" {
			self++
			if strings.Contains(renderers.RenderQuickLessonSVG(frame, 320, 360), `data-motion=`) {
				t.Fatal("self-swap motion")
			}
		}
	}
	if self != 2 {
		t.Fatal(self)
	}
}

func TestQuickLessonValidation(t *testing.T) {
	for _, input := range []string{"[]", "null", "[null]", "[1.5]", "[true]", "{}", "oops", "[1000001]", "[-1000001]"} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted %s", input)
		}
	}
	if _, err := (Lesson{}).Build(nil); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{64, 65} {
		b, _ := json.Marshal(make([]int, n))
		_, err := (Lesson{}).Build(b)
		if (err == nil) != (n == 64) {
			t.Fatal(n, err)
		}
	}
	if _, err := (Lesson{}).Build([]byte("[-1000000,0,1000000]")); err != nil {
		t.Fatal(err)
	}
}
