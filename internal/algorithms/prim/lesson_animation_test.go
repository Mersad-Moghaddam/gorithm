package prim

import (
	"encoding/json"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestPrimValidationBounds(t *testing.T) {
	base := LessonInput{Nodes: []renderers.GraphNode{{ID: strings.Repeat("界", 64), Label: strings.Repeat("界", 64), X: 1000000, Y: -1000000}, {ID: "B"}}, Edges: []WeightedEdge{{From: strings.Repeat("界", 64), To: "B", Weight: -10000}}}
	check := func(in LessonInput, valid bool) {
		t.Helper()
		raw, _ := json.Marshal(in)
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != valid {
			t.Fatalf("valid=%t err=%v", valid, err)
		}
	}
	check(base, true)
	bad := base
	bad.Nodes = append([]renderers.GraphNode{}, base.Nodes...)
	bad.Nodes[0].X = 1000001
	check(bad, false)
	bad = base
	bad.Nodes = append([]renderers.GraphNode{}, base.Nodes...)
	bad.Nodes[0].Label = strings.Repeat("界", 65)
	check(bad, false)
	bad = base
	bad.Nodes = append([]renderers.GraphNode{}, base.Nodes...)
	bad.Nodes[1].ID = bad.Nodes[0].ID
	check(bad, false)
	for i := 0; i < 32; i++ {
		if i >= len(base.Nodes) {
			base.Nodes = append(base.Nodes, renderers.GraphNode{ID: fmt.Sprint(i)})
		}
	}
	base.Edges = nil
	for i := 0; i < 128; i++ {
		base.Edges = append(base.Edges, WeightedEdge{base.Nodes[0].ID, "B", 10000})
	}
	check(base, true)
	bad = base
	bad.Edges = append(append([]WeightedEdge{}, base.Edges...), base.Edges[0])
	check(bad, false)
	bad = base
	bad.Nodes = append(append([]renderers.GraphNode{}, base.Nodes...), renderers.GraphNode{ID: "extra"})
	check(bad, false)
}

func TestPrimRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{`{"nodes":[null]}`, `{"edges":[{"from":"A","to":"B","weight":null}]}`, `{"nodes":[{"id":"A","x":null}]}`} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted null field: %s", input)
		}
	}
	for _, input := range []string{`null`, `{`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"B","weight":1}]}`, `{"edges":[{"from":"A","to":"B","weight":10001}]}`, `{"edges":[{"from":"A","to":"B","weight":0}],"start":"missing"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
}

func TestPrimTypedTrace(t *testing.T) {
	nodes, edges, start := DefaultGraph()
	steps := Trace(nodes, edges, start)
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatal(err)
	}
	last, ok := steps[len(steps)-1].State.(renderers.PrimLessonState)
	if !ok {
		t.Fatalf("need typed state, got %T", steps[len(steps)-1].State)
	}
	if !reflect.DeepEqual(last.SelectedEdges, []int{1, 2, 3, 6, 8}) || last.TotalWeight != 13 || !last.Connected {
		t.Fatalf("bad result %+v", last)
	}
	previousTotal, previousCount := 0, 0
	obsolete := false
	for _, step := range steps {
		s := step.State.(renderers.PrimLessonState)
		if s.TotalWeight != previousTotal && step.LineID != "L13" {
			t.Fatalf("sum changed at %s", step.LineID)
		}
		if len(s.SelectedEdges) != previousCount && step.LineID != "L12" {
			t.Fatalf("selection changed at %s", step.LineID)
		}
		if step.LineID == "L9" && s.ActiveEdge == 0 && s.Discarded {
			obsolete = true
		}
		previousTotal, previousCount = s.TotalWeight, len(s.SelectedEdges)
	}
	for i, step := range steps {
		if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
			t.Fatal("invalid hold")
		}
		if step.LineID == "L8" && step.HoldMS+steps[i+1].HoldMS != 1800 {
			t.Fatal("comparison budget")
		}
		if step.LineID == "L7" && step.State.(renderers.PrimLessonState).ActiveEdge == 2 && step.State.(renderers.PrimLessonState).Current != "B" {
			t.Fatal("BC original index lost reverse orientation")
		}
	}
	if !obsolete {
		t.Fatal("missing obsolete AB candidate")
	}
	before := steps[2].State.(renderers.PrimLessonState)
	last.Nodes[0].Label = "mutated"
	last.Edges[0].Weight = 999
	last.InTree["A"] = false
	last.SelectedEdges[0] = 999
	if before.Nodes[0].Label == "mutated" || before.Edges[0].Weight == 999 || !before.InTree["A"] {
		t.Fatal("aliased snapshot")
	}
	var queueSteps []int
	for i, step := range steps {
		if len(step.State.(renderers.PrimLessonState).Candidates) > 1 {
			queueSteps = append(queueSteps, i)
		}
	}
	i, j := queueSteps[0], queueSteps[1]
	earlier := steps[i].State.(renderers.PrimLessonState)
	later := steps[j].State.(renderers.PrimLessonState)
	later.Candidates[0].Weight = 12345
	if earlier.Candidates[0].Weight == 12345 {
		t.Fatal("heap aliases")
	}
	if len(later.CutEdges) > 0 {
		later.CutEdges[0] = 12345
		for _, id := range earlier.CutEdges {
			if id == 12345 {
				t.Fatal("cut aliases")
			}
		}
	}
}

func TestPrimEdgeCases(t *testing.T) {
	empty := Trace(nil, nil, "")
	if len(empty) != 1 || empty[0].LineID != "L1" {
		t.Fatal("empty trace")
	}
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	edges := []WeightedEdge{{"A", "B", 0}, {"A", "B", -2}}
	steps := Trace(nodes, edges, "A")
	s := steps[len(steps)-1].State.(renderers.PrimLessonState)
	if s.Connected || s.TotalWeight != -2 || !reflect.DeepEqual(s.SelectedEdges, []int{1}) || !strings.Contains(steps[len(steps)-1].Explanation, "start-component") {
		t.Fatalf("bad disconnected result %+v", s)
	}
	_, total := MST([]string{"A", "B", "C"}, edges, "A")
	if total != s.TotalWeight {
		t.Fatal("native mismatch")
	}
	one := Trace(nodes[:1], nil, "")
	final := one[len(one)-1].State.(renderers.PrimLessonState)
	if !final.Connected || final.TotalWeight != 0 || one[len(one)-1].LineID != "L18" {
		t.Fatal("singleton")
	}
	if !reflect.DeepEqual(Trace(nodes, edges, "A"), Trace(nodes, edges, "A")) {
		t.Fatal("unstable ties")
	}
}
