package dijkstra

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"math"
	"reflect"
	"testing"
)

func TestDijkstraRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{`null`, `{`, `{"start":"Z"}`, `{"target":"Z"}`, `{"start":null}`, `{"target":123}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
}

func TestDijkstraTypedOperations(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	last := steps[len(steps)-1].State.(renderers.DijkstraLessonState)
	if !last.Found || last.Distance["E"] != 10 || !reflect.DeepEqual(last.Path, []string{"A", "C", "B", "D", "E"}) || !last.Settled["E"] {
		t.Fatalf("bad result: %+v", last)
	}
	compared := false
	for i, step := range steps {
		s := step.State.(renderers.DijkstraLessonState)
		if step.LineID == "L10" && s.Current == "C" && s.Neighbor == "B" {
			compared = true
			if s.Distance["B"] != 4 || s.OldDistance != 4 || s.Candidate != 3 {
				t.Fatal("comparison previews mutation")
			}
			next := steps[i+1].State.(renderers.DijkstraLessonState)
			if next.Distance["B"] != 3 || next.Previous["B"] != "A" || step.HoldMS+steps[i+1].HoldMS != 1800 {
				t.Fatal("commit/predecessor/timing mismatch")
			}
		}
	}
	if !compared {
		t.Fatal("missing relaxation")
	}
	first := steps[0].State.(renderers.DijkstraLessonState)
	last.Distance["A"] = 99
	last.Previous["A"] = "bad"
	last.Settled["A"] = false
	last.Nodes[0].Label = "bad"
	last.Edges[0].Weight = 99
	last.Path[0] = "bad"
	if first.Distance["A"] != math.MaxInt || len(first.Previous) != 0 || first.Settled["A"] || first.Nodes[0].Label != "A" || first.Edges[0].Weight != 4 {
		t.Fatal("aliased snapshots")
	}
}

func TestDijkstraStaleAndTerminalCases(t *testing.T) {
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "Z"}}
	edges := []renderers.GraphEdge{{From: "A", To: "B", Weight: 4}, {From: "A", To: "C", Weight: 1}, {From: "C", To: "B", Weight: 1}}
	steps := Trace(nodes, edges, "A", "")
	stale := false
	for _, step := range steps {
		s := step.State.(renderers.DijkstraLessonState)
		if s.Stale {
			stale = true
			if s.Current != "B" || s.PoppedDistance != 4 || s.Distance["B"] != 2 {
				t.Fatal("bad stale entry")
			}
		}
	}
	if !stale {
		t.Fatal("missing stale pop")
	}
	ids := []string{"A", "B", "C", "Z"}
	nativeEdges := []WeightedEdge{{From: "A", To: "B", Weight: 4}, {From: "A", To: "C", Weight: 1}, {From: "C", To: "B", Weight: 1}}
	dist, _ := ShortestPath(ids, nativeEdges, "A", "")
	if !reflect.DeepEqual(dist, steps[len(steps)-1].State.(renderers.DijkstraLessonState).Distance) {
		t.Fatal("trace/native full search mismatch")
	}
	for i, step := range steps {
		s := step.State.(renderers.DijkstraLessonState)
		if len(s.Queue) > 0 {
			before := s.Queue[0]
			s.Queue[0].Node = "mutated"
			for j, other := range steps {
				if j == i {
					continue
				}
				for _, q := range other.State.(renderers.DijkstraLessonState).Queue {
					if q.Serial == before.Serial && q.Node == "mutated" {
						t.Fatal("queue alias")
					}
				}
			}
			break
		}
	}
	for _, target := range []string{"missing", "Z", "A"} {
		ss := Trace(nodes, edges, "A", target)
		s := ss[len(ss)-1].State.(renderers.DijkstraLessonState)
		if s.Found != (target == "A") {
			t.Fatalf("target %s result %+v", target, s)
		}
	}
	ss := Trace(nodes, edges, "missing", "B")
	if !ss[len(ss)-1].State.(renderers.DijkstraLessonState).InvalidStart {
		t.Fatal("invalid start lost")
	}
}

func TestDijkstraZeroWeightAndDefaults(t *testing.T) {
	for _, input := range []string{`{}`, `{"start":"","target":""}`, `{"start":"A","target":"E"}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err != nil {
			t.Fatal(err)
		}
	}
	steps := Trace([]renderers.GraphNode{{ID: "A"}, {ID: "B"}}, []renderers.GraphEdge{{From: "A", To: "B", Weight: 0}}, "A", "B")
	last := steps[len(steps)-1].State.(renderers.DijkstraLessonState)
	if !last.Found || last.Distance["B"] != 0 {
		t.Fatal("zero edge rejected")
	}
}

func TestDijkstraTraceRespectsDirectedEdges(t *testing.T) {
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}}
	edges := []renderers.GraphEdge{{From: "A", To: "B", Weight: 1, Directed: true}}
	steps := Trace(nodes, edges, "B", "A")
	last := steps[len(steps)-1].State.(renderers.DijkstraLessonState)
	if last.Found || last.Distance["A"] != math.MaxInt {
		t.Fatalf("directed edge traversed backward: %+v", last)
	}
}
