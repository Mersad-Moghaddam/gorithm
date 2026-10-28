package merge

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func mergeSceneEntities(t *testing.T, svg string) map[string]map[string]string {
	t.Helper()
	entities := map[string]map[string]string{}
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	for {
		token, err := d.Token()
		if err == io.EOF { return entities }
		if err != nil { t.Fatal(err) }
		if element, ok := token.(xml.StartElement); ok {
			attrs := map[string]string{}
			for _, attr := range element.Attr { attrs[attr.Name.Local] = attr.Value }
			if key := attrs["data-entity"]; key != "" {
				if entities[key] != nil { t.Fatalf("duplicate scene key %s", key) }
				entities[key] = attrs
			}
		}
	}
}

func TestMergeLessonCursorHalves(t *testing.T) {
	for _, input := range [][]int{{1,2}, {1,2,3,4}} {
		drained := false
		for _, step := range Trace(input) {
			s := step.State.(renderers.MergeLessonState)
			if step.LineID != "L22" || s.I <= s.Mid { continue }
			drained = true
			svg := renderers.RenderMergeLessonSVG(step,320,360)
			t.Logf("input=%v line=%s range=%d..%d mid=%d i=%d j=%d",input,step.LineID,s.Left,s.Right,s.Mid,s.I,s.J)
			if strings.Contains(svg,"↑ i") { t.Errorf("exhausted i rendered over right half: %v",input) }
			if !strings.Contains(svg,"↑ j") { t.Errorf("active j missing: %v",input) }
		}
		if !drained { t.Fatal("fixture did not drain left",input) }
	}
}

func TestMergeCopySnapshotsMatchHighlightedSource(t *testing.T) {
	source := map[string]string{}
	for _, line := range (Lesson{}).Source() {
		source[line.ID] = strings.TrimSpace(line.Code)
	}
	want := map[string]string{"L15": "aux[k] = arr[i]", "L17": "aux[k] = arr[j]", "L21": "aux[k] = arr[i]", "L22": "aux[k] = arr[j]"}
	seen := map[string]bool{}
	for _, input := range [][]int{{1, 2}, {2, 1}, {1, 1}} {
		for _, step := range Trace(input) {
			code, ok := want[step.LineID]
			if !ok {
				continue
			}
			seen[step.LineID] = true
			if source[step.LineID] != code {
				t.Errorf("%s highlights %q, want %q", step.LineID, source[step.LineID], code)
			}
			s := step.State.(renderers.MergeLessonState)
			if !s.AuxFilled[s.K] || s.AuxIDs[s.K] != s.IDs[s.Current] || s.Aux[s.K] != s.Values[s.Current] {
				t.Errorf("%s: snapshot does not contain selected copy", step.LineID)
			}
			key := fmt.Sprintf("aux:%d:%d", s.K, s.IDs[s.Current])
			entity := mergeSceneEntities(t, renderers.RenderMergeLessonSVG(step, 320, 360))[key]
			if entity["data-from-entity"] != fmt.Sprintf("main:%d:%d", s.Current, s.IDs[s.Current]) {
				t.Errorf("%s: missing copy motion source", step.LineID)
			}
		}
	}
	for line := range want {
		if !seen[line] {
			t.Errorf("fixture did not exercise %s", line)
		}
	}
}

func TestMergeLessonCopyArrival(t *testing.T) {
	for _, input := range [][]int{{1,2},{1,1},{2,1}} {
		steps:=Trace(input); writes:=0
		for index,step:=range steps {
			if step.LineID!="L23" {continue};writes++
			s:=step.State.(renderers.MergeLessonState)
			before:=mergeSceneEntities(t,renderers.RenderMergeLessonSVG(steps[index-1],320,360))
			svg:=renderers.RenderMergeLessonSVG(step,320,360)
			after:=mergeSceneEntities(t,svg)
			key:=fmt.Sprintf("incoming:%d:%d",s.CopyIndex,s.AuxIDs[s.CopyIndex])
			source:=fmt.Sprintf("aux:%d:%d",s.CopyIndex,s.AuxIDs[s.CopyIndex])
			main:=fmt.Sprintf("main:%d:%d",s.Left+s.CopyIndex,s.AuxIDs[s.CopyIndex])
			t.Logf("input=%v copy=%d main already present=%t aux=%s",input,s.CopyIndex,before[main]!=nil,source)
			arrival:=after[key]
			if arrival==nil {t.Errorf("missing incoming copy entity %s",key);continue}
			if before[key]!=nil {t.Errorf("arrival existed preceding write: %s",key)}
			if arrival["data-from-entity"]!=source||before[source]==nil||after[source]==nil {t.Errorf("copy source absent or wrong: %s",source)}
			if after[main]==nil {t.Errorf("authoritative main missing: %s",main)}
			if arrival["data-x"]!=after[main]["data-x"]||arrival["data-y"]!=after[main]["data-y"] {t.Error("arrival endpoint not main slot")}
			if arrival["data-x"]==before[source]["data-x"]&&arrival["data-y"]==before[source]["data-y"] {t.Error("zero displacement copy")}
			if !strings.Contains(svg,"Copy aux[")||!strings.Contains(svg,"write arrival") {t.Error("copy representation lacks visible explanation")}
		}
		if writes!=2 {t.Fatal("fixture requires two writes",input,writes)}
		for _,step:=range steps {if step.LineID!="L23"&&strings.Contains(renderers.RenderMergeLessonSVG(step,320,360),`data-entity="incoming:`) {t.Fatal("arrival outside actual write")}}
	}
}

func TestMergeLessonSequence(t *testing.T) {
	steps := Trace([]int{2, 1})
	lines := strings.Fields("L1 L2 L5 L6 L2 L3 L7 L2 L3 L8 L10 L11 L12 L14 L17 L21 L23 L23 L24 L9")
	actions := strings.Fields("init highlight set visit highlight visit visit highlight visit highlight visit set set compare set set set set highlight done")
	ops := []int{0, 1, 2, 2, 3, 3, 3, 4, 4, 4, 4, 5, 6, 7, 8, 9, 10, 11, 11, 11}
	if len(steps) != len(lines) {
		t.Fatalf("steps: %d", len(steps))
	}
	for i, step := range steps {
		comparisons := 0
		if i >= 13 {
			comparisons = 1
		}
		if step.LineID != lines[i] || string(step.Action) != actions[i] || step.Metrics != (trace.Metrics{Operations: ops[i], Comparisons: comparisons}) {
			t.Fatalf("step %d: %+v", i, step)
		}
	}
}

func TestMergeLessonValidation(t *testing.T) {
	for _, raw := range []string{"[]", "null", "[null]", "{}", "broken", "[1.5]", "[1000001]", "[-1000001]", "[" + strings.Repeat("0,", 64) + "0]"} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, raw := range []json.RawMessage{nil, []byte("[-1000000,0,1000000]"), []byte("[" + strings.Repeat("0,", 63) + "0]")} {
		if _, err := (Lesson{}).Build(raw); err != nil {
			t.Fatal(err)
		}
	}
}

func TestMergeLessonSortUnchanged(t *testing.T) {
	input := []int{2, 1, 2, 1}
	before := append([]int(nil), input...)
	Trace(input)
	if !reflect.DeepEqual(input, before) {
		t.Fatal("input mutated")
	}
}

func TestMergeLessonSnapshots(t *testing.T) {
	for _, input := range [][]int{{}, {7}, {2, 1}, {2, 1, 2, 1}, {3, 1, 2}, {-2, 0, -3}, {0, 0}} {
		steps := Trace(input)
		valid := map[string]bool{}
		for _, line := range (Lesson{}).Source() {
			valid[line.ID] = true
		}
		for i, step := range steps {
			s, ok := step.State.(renderers.MergeLessonState)
			if !ok {
				t.Fatalf("typed state missing: %T", step.State)
			}
			if !valid[step.LineID] {
				t.Fatal(step.LineID)
			}
			if step.HoldMS != 0 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 350 && step.HoldMS != 1800 {
				t.Fatal("hold", step.HoldMS)
			}
			if step.LineID == "L14" && (step.HoldMS != 600 || steps[i+1].HoldMS != 1200) {
				t.Fatal("compare/commit budget")
			}
			if step.LineID == "L24" && len(s.Sorted) != 0 {
				t.Fatal("local merge claims global sorted")
			}
			if step.LineID == "L2" && len(s.Aux) != 0 {
				t.Fatal("stale aux entering invocation")
			}
		}
		last := steps[len(steps)-1].State.(renderers.MergeLessonState)
		if !reflect.DeepEqual(last.Values, Sort(input)) {
			t.Fatal(last.Values)
		}
		seen := map[int]bool{}
		for _, id := range last.IDs {
			if seen[id] {
				t.Fatal("duplicate final identity")
			}
			seen[id] = true
			if input[id] != last.Values[len(seen)-1] {
				t.Fatal("identity/value mismatch")
			}
		}
		if len(input) == 4 && !reflect.DeepEqual(last.IDs, []int{1, 3, 0, 2}) {
			t.Fatal("unstable", last.IDs)
		}
		if len(input) == 4 && !reflect.DeepEqual(last.Values, []int{1, 1, 2, 2}) {
			t.Fatal("literal result", last.Values)
		}
	}
	for _, step := range Trace([]int{3, 1, 2}) {
		s := step.State.(renderers.MergeLessonState)
		if step.LineID == "L5" && s.Left == 0 && s.Right == 2 && s.Mid != 1 {
			t.Fatal("root midpoint", s)
		}
	}
	zero := Trace([]int{0, 0})
	equalLeft := false
	for _, step := range zero {
		s := step.State.(renderers.MergeLessonState)
		if step.LineID == "L11" && (s.AuxFilled[0] || s.AuxIDs[0] != -1) {
			t.Fatal("unwritten zero")
		}
		if step.LineID == "L15" {
			equalLeft = true
			if s.Aux[0] != 0 || !s.AuxFilled[0] || s.AuxIDs[0] != 0 {
				t.Fatal("filled zero", s)
			}
		}
	}
	if !equalLeft {
		t.Fatal("equal did not take left")
	}
	steps := Trace([]int{2, 1})
	for _, step := range steps {
		s := step.State.(renderers.MergeLessonState)
		if step.LineID == "L17" && (!reflect.DeepEqual(s.Values, []int{2, 1}) || !reflect.DeepEqual(s.AuxIDs, []int{1, -1}) || !reflect.DeepEqual(s.AuxFilled, []bool{true, false}) || s.I != 0 || s.J != 1 || s.K != 0) {
			t.Fatal("right transfer", s)
		}
		if step.LineID == "L21" && !reflect.DeepEqual(s.Aux, []int{1, 2}) {
			t.Fatal("drain", s)
		}
		if step.LineID == "L23" && s.CopyIndex == 0 && (!reflect.DeepEqual(s.Values, []int{1, 1}) || !reflect.DeepEqual(s.IDs, []int{1, 1})) {
			t.Fatal("real memory copy", s)
		}
	}
	a := steps[14].State.(renderers.MergeLessonState)
	later := steps[15].State.(renderers.MergeLessonState)
	later.Values[0] = 99
	later.IDs[0] = 99
	later.Aux[0] = 99
	later.AuxIDs[0] = 99
	later.AuxFilled[0] = false
	later.Stack[0].Left = 99
	if a.Values[0] != 2 || a.IDs[0] != 0 || a.Aux[0] != 1 || a.AuxIDs[0] != 1 || !a.AuxFilled[0] || a.Stack[0].Left != 0 {
		t.Fatal("snapshot aliases")
	}
}
