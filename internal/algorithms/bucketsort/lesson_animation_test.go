package bucketsort

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBucketsortLessonSnapshots(t *testing.T) {
	baseline := Trace([]int{3, 2, 0, 10})
	wantLines := []string{"L1", "L2", "L4", "L5", "L6"}
	for i := 0; i < 4; i++ {
		wantLines = append(wantLines, "L7", "L8", "L9")
	}
	for i := 0; i < 5; i++ {
		wantLines = append(wantLines, "L11", "L12")
	}
	wantLines = append(wantLines, "L14")
	for _, count := range []int{2, 1, 0, 0, 1} {
		wantLines = append(wantLines, "L15")
		for i := 0; i < count; i++ {
			wantLines = append(wantLines, "L16", "L17")
		}
	}
	wantLines = append(wantLines, "L20", "L21")
	if len(baseline) != len(wantLines) {
		t.Fatalf("sequence length %d", len(baseline))
	}
	for i, s := range baseline {
		if s.LineID != wantLines[i] {
			t.Fatalf("step %d: %s", i, s.LineID)
		}
		wantAction := "set"
		switch s.LineID {
		case "L1":
			wantAction = "init"
		case "L7", "L11", "L15", "L16", "L20":
			wantAction = "highlight"
		case "L9":
			wantAction = "push"
		case "L21":
			wantAction = "done"
		}
		if string(s.Action) != wantAction {
			t.Fatalf("step %d action %s", i, s.Action)
		}
		if s.LineID == "L8" || s.LineID == "L16" {
			if s.HoldMS+baseline[i+1].HoldMS != 1800 {
				t.Fatal("pair budget")
			}
		}
	}
	m := baseline[len(baseline)-1].Metrics
	if m.Comparisons != 4 || m.Swaps != 1 || m.Operations != 18 {
		t.Fatalf("metrics changed %+v", m)
	}
	for _, tc := range []struct{ input, values, ids []int }{
		{[]int{0, 3, 2, 10}, []int{0, 2, 3, 10}, []int{0, 2, 1, 3}},
		{[]int{3, 2, 0, 10}, []int{0, 2, 3, 10}, []int{2, 1, 0, 3}},
		{[]int{2, 2, 0, 10}, []int{0, 2, 2, 10}, []int{2, 0, 1, 3}},
		{[]int{-3, -1, -2, 0}, []int{-3, -2, -1, 0}, []int{0, 2, 1, 3}},
	} {
		original := append([]int(nil), tc.input...)
		steps := Trace(tc.input)
		final := steps[len(steps)-1].State.(renderers.BucketsortLessonState)
		if !reflect.DeepEqual(final.Values, tc.values) || !reflect.DeepEqual(final.IDs, tc.ids) || !reflect.DeepEqual(final.Values, Sort(tc.input)) {
			t.Fatalf("final %+v", final)
		}
		if !reflect.DeepEqual(original, tc.input) {
			t.Fatal("input mutated")
		}
		source := map[string]bool{}
		for _, l := range (Lesson{}).Source() {
			source[l.ID] = true
		}
		for _, step := range steps {
			if !source[step.LineID] {
				t.Fatalf("unknown source %s", step.LineID)
			}
		}
	}
	steps := Trace([]int{3, 2, 0, 10})
	var before, after renderers.BucketsortLessonState
	for _, step := range steps {
		s := step.State.(renderers.BucketsortLessonState)
		if step.LineID == "L11" && s.ActiveBucket == 0 {
			before = s
		}
		if step.LineID == "L12" && s.ActiveBucket == 0 {
			after = s
		}
	}
	if !reflect.DeepEqual(before.Buckets[0], []int{2, 0}) || !reflect.DeepEqual(after.Buckets[0], []int{0, 2}) || !reflect.DeepEqual(after.BucketIDs[0], []int{2, 1}) {
		t.Fatal("local reorder missing")
	}
	after.Buckets[0][0] = 99
	after.BucketIDs[0][0] = 99
	after.Values[0] = 99
	after.IDs[0] = 99
	if before.Buckets[0][0] != 2 || before.BucketIDs[0][0] != 1 || before.Values[0] != 3 || before.IDs[0] != 0 {
		t.Fatal("aliased snapshot")
	}
	scatter := Trace([]int{0, 3, 2, 10})[16].State.(renderers.BucketsortLessonState)
	if !reflect.DeepEqual(scatter.Buckets, [][]int{{0, 2}, {3}, nil, nil, {10}}) {
		t.Fatalf("hash buckets %v", scatter.Buckets)
	}
	for _, input := range [][]int{nil, {7}, {4, 4, 4}} {
		for _, step := range Trace(input) {
			s := step.State.(renderers.BucketsortLessonState)
			if s.BucketCount != 0 || len(s.Buckets) != 0 || step.LineID == "L8" {
				t.Fatal("early exit allocated buckets")
			}
		}
	}
	equal := Trace([]int{4, 4, 4})
	if equal[len(equal)-2].LineID != "L4_equal" {
		t.Fatal("missing equal-range guard")
	}
}

func TestBucketsortBuildValidation(t *testing.T) {
	for _, raw := range []string{"[]", "null", "{}", "[1,]", "[1000001]", "[-1000001]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	if _, err := (Lesson{}).Build(nil); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{64, 65} {
		raw, _ := json.Marshal(make([]int, n))
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 64) {
			t.Fatalf("length %d: %v", n, err)
		}
	}
	if _, err := (Lesson{}).Build([]byte("[-1000000,0,1000000]")); err != nil {
		t.Fatal(err)
	}
}
