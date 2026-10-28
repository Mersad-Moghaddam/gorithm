package bubble

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBubbleLessonSequence(t *testing.T) {
	steps := Trace([]int{3, 1, 2})
	swapCount := 0
	for _, step := range steps {
		if step.Action == trace.ActionSwap {
			swapCount++
		}
	}
	if swapCount != 2 {
		t.Fatalf("swaps = %d", swapCount)
	}
	wantLines := []string{"L1/init", "L2/set", "L3/highlight", "L5/compare", "L6/swap", "L5/compare", "L6/swap", "L8/set", "L3/highlight", "L5/compare", "L8/set", "L10/done"}
	wantMetrics := []trace.Metrics{{}, {}, {}, {Comparisons: 1, Operations: 1}, {Comparisons: 1, Swaps: 1, Operations: 2}, {Comparisons: 2, Swaps: 1, Operations: 3}, {Comparisons: 2, Swaps: 2, Operations: 4}, {Comparisons: 2, Swaps: 2, Operations: 4}, {Comparisons: 2, Swaps: 2, Operations: 4}, {Comparisons: 3, Swaps: 2, Operations: 5}, {Comparisons: 3, Swaps: 2, Operations: 5}, {Comparisons: 3, Swaps: 2, Operations: 5}}
	if len(steps) != len(wantLines) {
		t.Fatalf("steps = %d, want %d", len(steps), len(wantLines))
	}
	for i, step := range steps {
		if got := step.LineID + "/" + string(step.Action); got != wantLines[i] || step.Metrics != wantMetrics[i] {
			t.Errorf("step %d: %s %+v, want %s %+v", i, got, step.Metrics, wantLines[i], wantMetrics[i])
		}
	}
}

func TestBubbleLessonSnapshots(t *testing.T) {
	for _, tc := range []struct{ input, output, ids []int }{
		{[]int{3, 1, 2}, []int{1, 2, 3}, []int{1, 2, 0}},
		{[]int{2, 2, 1}, []int{1, 2, 2}, []int{2, 0, 1}},
		{[]int{}, []int{}, []int{}}, {[]int{7}, []int{7}, []int{0}},
		{[]int{-2, 0, -3}, []int{-3, -2, 0}, []int{2, 0, 1}},
		{[]int{1, 2, 3}, []int{1, 2, 3}, []int{0, 1, 2}},
	} {
		steps := Trace(tc.input)
		last := steps[len(steps)-1].State.(renderers.BubbleLessonState)
		if !reflect.DeepEqual(last.Values, tc.output) || !reflect.DeepEqual(last.Values, Sort(tc.input)) || !reflect.DeepEqual(last.IDs, tc.ids) || len(last.Sorted) != len(tc.input) {
			t.Fatalf("input %v: final %+v", tc.input, last)
		}
		lines := map[string]bool{}
		for _, line := range (Lesson{}).Source() {
			lines[line.ID] = true
		}
		for i, step := range steps {
			state := step.State.(renderers.BubbleLessonState)
			first := steps[0].State.(renderers.BubbleLessonState)
			if state.MinValue != first.MinValue || state.MaxValue != first.MaxValue || state.MaxValue <= state.MinValue {
				t.Fatal("domain must remain fixed and nonzero")
			}
			if !lines[step.LineID] {
				t.Fatalf("unknown source %s", step.LineID)
			}
			switch step.HoldMS {
			case 0, 350, 600, 1200, 1800:
			default:
				t.Fatalf("hold %d", step.HoldMS)
			}
			if step.Action == trace.ActionSwap && (i == 0 || steps[i-1].HoldMS+step.HoldMS != 1800) {
				t.Fatal("swap pair budget")
			}
		}
	}
	steps := Trace([]int{3, 1, 2})
	first := steps[0].State.(renderers.BubbleLessonState)
	swap := steps[4].State.(renderers.BubbleLessonState)
	if !reflect.DeepEqual(swap.IDs, []int{1, 0, 2}) || swap.Pass != 0 || swap.Scan != 0 || swap.UnsortedEnd != 2 {
		t.Fatalf("swap %+v", swap)
	}
	if !reflect.DeepEqual(steps[7].State.(renderers.BubbleLessonState).Sorted, []int{2}) {
		t.Fatal("suffix missing")
	}
	swap.Values[0], swap.IDs[0], swap.Swapped[0] = 99, 99, 99
	if !reflect.DeepEqual(first.Values, []int{3, 1, 2}) || !reflect.DeepEqual(first.IDs, []int{0, 1, 2}) {
		t.Fatal("snapshot alias")
	}
	compare := steps[3].State.(renderers.BubbleLessonState)
	compare.Comparing[0] = 99
	if steps[5].State.(renderers.BubbleLessonState).Comparing[0] != 1 {
		t.Fatal("comparison alias")
	}
	later := steps[10].State.(renderers.BubbleLessonState)
	later.Sorted[0] = 99
	if steps[7].State.(renderers.BubbleLessonState).Sorted[0] != 2 {
		t.Fatal("sorted suffix alias")
	}
	passes, swaps := 0, 0
	for _, step := range Trace([]int{1, 2, 3}) {
		if step.LineID == "L3" {
			passes++
		}
		if step.Action == trace.ActionSwap {
			swaps++
		}
	}
	if passes != 2 || swaps != 0 {
		t.Fatalf("sorted fixture passes=%d swaps=%d", passes, swaps)
	}
}

func TestBubbleLessonBuildValidation(t *testing.T) {
	if _, err := (Lesson{}).Build(nil); err != nil {
		t.Fatal(err)
	}
	for _, raw := range []string{"[]", "null", "[null]", "{", "{}", "1", "[1.5]", "[true]", "[1000001]", "[-1000001]", "[" + strings.Repeat("0,", 64) + "0]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("Build(%s) accepted invalid input", raw)
		}
	}
	for _, raw := range []string{"[-1000000,0,1000000]", "[" + strings.Repeat("0,", 63) + "0]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Errorf("Build accepted boundary %s: %v", raw, err)
		}
	}
}

func TestBubbleSortFixtureUnchanged(t *testing.T) {
	input := []int{3, 1, 2}
	Trace(input)
	if !reflect.DeepEqual(input, []int{3, 1, 2}) {
		t.Fatal("Trace mutated input")
	}
}
