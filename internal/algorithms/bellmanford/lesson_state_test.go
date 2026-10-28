package bellmanford

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestBellmanFordPassSnapshots(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	final := steps[len(steps)-1].State.(renderers.BellmanFordLessonState)
	want := map[string]int{"A": 0, "B": 4, "C": 2, "D": 6, "E": 3}
	if !reflect.DeepEqual(final.Distance, want) || final.NegativeCycle || final.Pass != 2 {
		t.Fatalf("final: %+v", final)
	}
	paired := false
	for i, s := range steps {
		st := s.State.(renderers.BellmanFordLessonState)
		if s.LineID == "L11" && st.Pass == 1 && !reflect.DeepEqual(st.Distance, want) {
			t.Fatal("pass 1 must propagate in-place")
		}
		if s.LineID == "L11" && st.Pass == 2 && st.ChangedThisPass {
			t.Fatal("pass 2 changed")
		}
		if s.LineID == "L6" && st.Improved {
			if s.HoldMS != 600 || steps[i+1].HoldMS != 1200 {
				t.Fatal("compare/commit budget")
			}
			paired = true
		}
	}
	if !paired {
		t.Fatal("missing compare/commit")
	}
	final.Distance["A"] = 99
	final.PassStartDistance["A"] = 99
	final.Nodes[0].Label = "mutated"
	final.Edges[0].Weight = 99
	first := steps[0].State.(renderers.BellmanFordLessonState)
	if first.Distance["A"] != math.MaxInt || first.Nodes[0].Label != "A" || first.Edges[0].Weight != 4 || first.PassStartDistance["A"] == 99 {
		t.Fatal("aliased snapshot")
	}
}

func TestBellmanFordCycleWitness(t *testing.T) {
	for _, reachable := range []bool{true, false} {
		nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}}
		edges := []renderers.GraphEdge{{From: "B", To: "C", Weight: -2}, {From: "C", To: "B", Weight: -2}}
		if reachable {
			edges = append([]renderers.GraphEdge{{From: "A", To: "B", Weight: 1}}, edges...)
		}
		steps := Trace(nodes, edges, "A")
		st := steps[len(steps)-1].State.(renderers.BellmanFordLessonState)
		if st.NegativeCycle != reachable {
			t.Fatalf("reachable=%t state=%+v", reachable, st)
		}
		if reachable {
			e := edges[st.WitnessEdge]
			if st.Distance[e.From] == math.MaxInt || st.Distance[e.From]+e.Weight >= st.Distance[e.To] || steps[len(steps)-1].LineID != "L15" {
				t.Fatal("invalid witness")
			}
		}
		for _, s := range steps {
			x := s.State.(renderers.BellmanFordLessonState)
			if x.From != "" && x.Distance[x.From] == math.MaxInt && x.HasCandidate {
				t.Fatal("infinity arithmetic")
			}
		}
	}
}

func TestBellmanFordStrictInput(t *testing.T) {
	for _, s := range []string{`null`, `{`, `{"nodes":null}`, `{"nodes":[]}`, `{"start":"Z"}`, `{"nodes":[{"id":"A"},{"id":"A"}]}`, `{"edges":[{"from":"A","to":"Z"}]}`, `{"edges":[{"from":"A","to":"B","weight":10001}]}`, `{"nodes":[{"id":"A","x":1000001}]}`, `{"start":null}`} {
		if _, err := (Lesson{}).Build([]byte(s)); err == nil {
			t.Errorf("accepted %s", s)
		}
	}
}

func TestBellmanFordInputBoundaries(t *testing.T) {
	nodes := make([]renderers.GraphNode, 32)
	for i := range nodes {
		nodes[i] = renderers.GraphNode{ID: string(rune('A' + i)), Label: strings.Repeat("界", 64), X: 1000000, Y: -1000000}
	}
	edges := make([]renderers.GraphEdge, 128)
	for i := range edges {
		edges[i] = renderers.GraphEdge{From: nodes[0].ID, To: nodes[1].ID, Weight: -10000}
	}
	input := LessonInput{Start: nodes[0].ID, Nodes: nodes, Edges: edges}
	raw, _ := json.Marshal(input)
	if _, err := (Lesson{}).Build(raw); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(){func() { input.Nodes = append(nodes, renderers.GraphNode{ID: "extra"}) }, func() { input.Nodes = nodes; input.Edges = append(edges, edges[0]) }, func() {
		input.Edges = edges
		input.Nodes = append([]renderers.GraphNode(nil), nodes...)
		input.Nodes[0].Label = strings.Repeat("界", 65)
	}} {
		mutate()
		raw, _ = json.Marshal(input)
		if _, err := (Lesson{}).Build(raw); err == nil {
			t.Fatal("accepted exceeded limit")
		}
	}
}

func TestBellmanFordReverseEdgeOrderNativeAgreement(t *testing.T) {
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	edges := []renderers.GraphEdge{{From: "B", To: "C", Weight: 2}, {From: "A", To: "B", Weight: 1}}
	steps := Trace(nodes, edges, "A")
	got := steps[len(steps)-1].State.(renderers.BellmanFordLessonState)
	want, cycle := ShortestPath([]string{"A", "B", "C"}, []DirectedEdge{{From: "B", To: "C", Weight: 2}, {From: "A", To: "B", Weight: 1}}, "A")
	if !reflect.DeepEqual(got.Distance, want) || got.NegativeCycle != cycle || got.Pass != 2 {
		t.Fatal("native mismatch")
	}
	for _, ns := range [][]renderers.GraphNode{nil, {{ID: "A"}}} {
		invalid := Trace(ns, nil, "missing")
		if !invalid[len(invalid)-1].State.(renderers.BellmanFordLessonState).InvalidStart {
			t.Fatal("invalid start")
		}
	}
}
