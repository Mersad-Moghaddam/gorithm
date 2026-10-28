package astar

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"testing"
)

func TestAStarTypedSnapshots(t *testing.T) {
	steps, _ := (Lesson{}).Build(nil)
	first, ok := steps[0].State.(renderers.AStarLessonState)
	if !ok {
		t.Fatalf("want typed A* state, got %T", steps[0].State)
	}
	last := steps[len(steps)-1].State.(renderers.AStarLessonState)
	if last.Cost != 580 || !reflect.DeepEqual(last.Path, []string{"A", "B", "D", "F"}) || first.H["A"] != 520 || first.H["B"] != 374 || first.H["D"] != 189 {
		t.Fatalf("wrong result: %+v", last)
	}
	before := first.Nodes[0]
	last.Nodes[0].Label = "changed"
	last.G["A"] = 99
	last.H["A"] = 99
	last.CameFrom["A"] = "bad"
	last.Closed["A"] = false
	last.EdgeUndirected[0] = false
	if first.Nodes[0] != before || first.H["A"] != 520 || len(first.G) != 0 || len(first.CameFrom) != 0 || len(first.Closed) != 0 || !first.EdgeUndirected[0] {
		t.Fatal("snapshots alias")
	}
	for _, step := range steps {
		s := step.State.(renderers.AStarLessonState)
		seen := map[string]bool{}
		for i, e := range s.Open {
			if seen[e.Node] || e.HeapIndex != i {
				t.Fatal("invalid heap snapshot")
			}
			seen[e.Node] = true
		}
		if step.LineID == "L9" && s.Closed[s.Neighbor] && s.HasTentative {
			t.Fatal("closed neighbor relaxed")
		}
	}
}

func TestAStarInputValidation(t *testing.T) {
	for _, raw := range []string{`null`, `{`, `{"nodes":null}`, `{"nodes":[{"id":"A"},{"id":"A"}]}`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"X","weight":1}]}`, `{"nodes":[{"id":"A"}],"start":"X"}`, `{"nodes":[{"id":"A","x":1000001}]}`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"A","weight":-1}]}`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"A","weight":10001}]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	steps, err := (Lesson{}).Build(json.RawMessage(`{"nodes":[{"id":"A","x":0,"y":0},{"id":"B","x":3,"y":4}],"edges":[{"from":"A","to":"B","weight":10000}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if steps[0].State.(renderers.AStarLessonState).H["A"] != 5 {
		t.Fatal("input coordinates changed")
	}
}

func TestAStarHeapAndGuards(t *testing.T) {
	nodes := []AStarNode{{ID: "S"}, {ID: "A"}, {ID: "B"}, {ID: "G"}}
	edges := []AStarEdge{{From: "S", To: "A", Weight: 10}, {From: "S", To: "B", Weight: 1}, {From: "B", To: "A", Weight: 1}, {From: "A", To: "G", Weight: 1}}
	steps := Trace(nodes, edges, "S", "G")
	decreased := false
	for _, step := range steps {
		s := step.State.(renderers.AStarLessonState)
		if step.LineID == "L15" && s.Neighbor == "A" && s.G["A"] == 2 {
			decreased = true
			count := 0
			for _, e := range s.Open {
				if e.Node == "A" {
					count++
					if e.F != 2 {
						t.Fatal("stale heap priority")
					}
				}
			}
			if count != 1 {
				t.Fatal("decrease-key duplicated entry")
			}
		}
	}
	last := steps[len(steps)-1].State.(renderers.AStarLessonState)
	cost, path := FindPath(nodes, edges, "S", "G")
	if !decreased || last.Cost != cost || !reflect.DeepEqual(last.Path, path) {
		t.Fatal("heap/native mismatch")
	}
	for _, pair := range [][2]string{{"S", "S"}, {"G", "S"}, {"missing", "G"}} {
		ss := Trace(nodes, edges, pair[0], pair[1])
		state := ss[len(ss)-1].State.(renderers.AStarLessonState)
		want, p := FindPath(nodes, edges, pair[0], pair[1])
		if state.Cost != want || !reflect.DeepEqual(state.Path, p) {
			t.Fatalf("guard %v: %+v", pair, state)
		}
	}
	tieNodes := []AStarNode{{ID: "S", X: 0}, {ID: "A", X: 2}, {ID: "B", X: 1}, {ID: "G", X: 3}}
	tieEdges := []AStarEdge{{From: "S", To: "B", Weight: 1}, {From: "S", To: "A", Weight: 2}}
	popped := []string{}
	for _, step := range Trace(tieNodes, tieEdges, "S", "G") {
		if step.LineID == "L5" {
			popped = append(popped, step.State.(renderers.AStarLessonState).Current)
		}
	}
	if !reflect.DeepEqual(popped, []string{"S", "A", "B"}) {
		t.Fatalf("f tie must prefer lower h: %v", popped)
	}
	inconsistent := Trace([]AStarNode{{ID: "S", X: 100}, {ID: "G"}}, []AStarEdge{{From: "S", To: "G", Weight: 1}}, "S", "G")
	if inconsistent[0].State.(renderers.AStarLessonState).HeuristicConsistent {
		t.Fatal("overestimate labeled consistent")
	}
	for i, step := range steps {
		switch step.HoldMS {
		case 0, 350, 600, 1200, 1800:
		default:
			t.Fatal("invalid hold")
		}
		if step.HoldMS == 600 && steps[i+1].HoldMS != 1200 {
			t.Fatal("pair budget")
		}
	}
}
