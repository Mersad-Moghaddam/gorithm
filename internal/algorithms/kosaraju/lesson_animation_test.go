package kosaraju

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestKosarajuTypedSnapshots(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range steps {
		if got := fmt.Sprintf("%T", step.State); got != "renderers.KosarajuLessonState" {
			t.Fatalf("typed SCC snapshot missing: %s", got)
		}
	}
}

func TestKosarajuOrdersAndImmutableState(t *testing.T) {
	nodes, edges := defaultGraph()
	steps := Trace(nodes, edges)
	last := steps[len(steps)-1].State.(renderers.KosarajuLessonState)
	if !reflect.DeepEqual(last.FinishOrder, []string{"G", "F", "E", "D", "C", "B", "A"}) {
		t.Fatal(last.FinishOrder)
	}
	want := [][]string{{"A", "C", "B"}, {"D", "E"}, {"F", "G"}}
	if !reflect.DeepEqual(last.Components, want) {
		t.Fatal(last.Components)
	}
	var previous renderers.KosarajuLessonState
	for i, step := range steps {
		s := step.State.(renderers.KosarajuLessonState)
		for j, e := range s.Edges {
			want := edges[j]
			if s.Transposed {
				want.From, want.To = want.To, want.From
			}
			if e.From != want.From || e.To != want.To {
				t.Fatalf("edge %d at %s", j, step.LineID)
			}
		}
		if i > 0 && len(s.Components) != len(previous.Components) && step.LineID != "L13" {
			t.Fatal("commit outside L13")
		}
		if i > 0 && len(s.CurrentComponent) > len(previous.CurrentComponent) && step.LineID != "L12" {
			t.Fatal("member outside L12")
		}
		if len(s.Stack) > 0 && step.Action == trace.ActionVisit && s.Stack[len(s.Stack)-1] != s.Current {
			t.Fatal("call stack mismatch")
		}
		previous = s
	}
	before := steps[len(steps)-2].State.(renderers.KosarajuLessonState)
	last.Components[0][0] = "changed"
	last.ComponentOf["A"] = 99
	last.VisitedForward["A"] = false
	last.Nodes[0].Label = "changed"
	last.Edges[0].From = "changed"
	last.OriginalEdges[0].From = "changed"
	last.FinishOrder[0] = "changed"
	if before.Components[0][0] != "A" || before.ComponentOf["A"] != 0 || !before.VisitedForward["A"] || before.Nodes[0].Label != "A" || before.Edges[0].From != "B" || before.OriginalEdges[0].From != "A" || before.FinishOrder[0] != "G" {
		t.Fatal("snapshot alias")
	}
}

func TestKosarajuTraceOracle(t *testing.T) {
	for _, edges := range [][]renderers.GraphEdge{nil, {{From: "A", To: "A"}}, {{From: "A", To: "B"}, {From: "B", To: "A"}, {From: "B", To: "C"}}, {{From: "A", To: "B"}, {From: "C", To: "D"}}, {{From: "A", To: "B"}, {From: "A", To: "B"}, {From: "B", To: "A"}}} {
		nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}}
		pairs := [][2]string{}
		for _, e := range edges {
			pairs = append(pairs, [2]string{e.From, e.To})
		}
		steps := Trace(nodes, edges)
		got := steps[len(steps)-1].State.(renderers.KosarajuLessonState)
		if !reflect.DeepEqual(normalizeSCCs(got.Components), normalizeSCCs(FindSCCs([]string{"A", "B", "C", "D"}, pairs))) {
			t.Fatal(got.Components)
		}
	}
	one := Trace([]renderers.GraphNode{{ID: "A"}}, []renderers.GraphEdge{{From: "A", To: "A"}})
	if len(one[len(one)-1].State.(renderers.KosarajuLessonState).Components) != 1 {
		t.Fatal("singleton")
	}
	if len(Trace(nil, nil)) != 0 {
		t.Fatal("empty trace")
	}
}

func TestKosarajuInputBounds(t *testing.T) {
	valid := LessonInput{Nodes: []renderers.GraphNode{{ID: strings.Repeat("界", 64), Label: strings.Repeat("界", 64), X: 1000000, Y: -1000000}}}
	valid.Edges = []renderers.GraphEdge{{From: valid.Nodes[0].ID, To: valid.Nodes[0].ID, Weight: 10000}}
	check := func(input LessonInput, accept bool) {
		t.Helper()
		data, _ := json.Marshal(input)
		_, err := (Lesson{}).Build(data)
		if (err == nil) != accept {
			t.Fatalf("accepted=%t: %s: %v", accept, data, err)
		}
	}
	check(valid, true)
	valid.Nodes[0].X = 1000001
	check(valid, false)
	valid.Nodes[0].X = 1000000
	valid.Nodes[0].Label += "界"
	check(valid, false)
	valid.Nodes[0].Label = "A"
	valid.Edges[0].Weight = -10000
	check(valid, true)
	valid.Edges[0].Weight = -10001
	check(valid, false)
	nodes := make([]renderers.GraphNode, 32)
	for i := range nodes {
		nodes[i].ID = fmt.Sprint(i)
	}
	edges := make([]renderers.GraphEdge, 128)
	for i := range edges {
		edges[i] = renderers.GraphEdge{From: "0", To: "1"}
	}
	check(LessonInput{Nodes: nodes, Edges: edges}, true)
	check(LessonInput{Nodes: nodes, Edges: append(edges, edges[0])}, false)
	check(LessonInput{Nodes: append(nodes, renderers.GraphNode{ID: "33"})}, false)
}

func TestKosarajuRejectsMalformedInput(t *testing.T) {
	for _, input := range []string{`null`, `{`, `{"nodes":null}`, `{"edges":null}`, `{"nodes":[{"id":"A"},{"id":"A"}]}`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"B"}]}`} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted invalid input: %s", input)
		}
	}
}
