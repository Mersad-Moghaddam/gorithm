package bfs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBFSAnimationLimits(t *testing.T) {
	valid := func(n, e int, id string, coord float64, weight int) buildInput {
		v := buildInput{Start: id}
		for i := 0; i < n; i++ {
			key := fmt.Sprint(i)
			if i == 0 {
				key = id
			}
			v.Nodes = append(v.Nodes, renderers.GraphNode{ID: key, Label: id, X: coord})
		}
		for i := 0; i < e; i++ {
			v.Edges = append(v.Edges, renderers.GraphEdge{From: id, To: id, Weight: weight})
		}
		return v
	}
	for _, tc := range []struct {
		name  string
		input buildInput
		ok    bool
	}{
		{"accepted limits", valid(32, 128, strings.Repeat("界", 64), 1000000, 10000), true},
		{"accepted negatives", valid(1, 1, "A", -1000000, -10000), true},
		{"nodes", valid(33, 1, "A", 0, 0), false}, {"edges", valid(1, 129, "A", 0, 0), false},
		{"label", valid(1, 1, strings.Repeat("界", 65), 0, 0), false}, {"coordinates", valid(1, 1, "A", 1000001, 0), false},
		{"weight", valid(1, 1, "A", 0, 10001), false}, {"negative weight", valid(1, 1, "A", 0, -10001), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			_, err = (Lesson{}).Build(raw)
			if (err == nil) != tc.ok {
				t.Fatalf("accepted=%v want %v: %v", err == nil, tc.ok, err)
			}
		})
	}
}

func TestBFSAnimationBoundary(t *testing.T) {
	for _, input := range []string{`null`, `{`, `{"start":"missing"}`, `{"nodes":[{"id":"A"},{"id":"A"}]}`, `{"edges":[{"from":"A","to":"missing"}]}`, `{"nodes":null}`, `{"start":null}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := reflect.TypeOf(steps[0].State).Name(); got != "BFSLessonState" {
		t.Fatalf("state = %s; want BFSLessonState", got)
	}
}

func TestBFSAnimationSnapshots(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	last := steps[len(steps)-1].State.(renderers.BFSLessonState)
	if !reflect.DeepEqual(last.Order, []string{"A", "B", "C", "D", "E", "F"}) || !reflect.DeepEqual(last.Distance, map[string]int{"A": 0, "B": 1, "C": 1, "D": 2, "E": 2, "F": 3}) {
		t.Fatalf("bad BFS result: %+v", last)
	}
	enqueues := 0
	for i, step := range steps {
		s := step.State.(renderers.BFSLessonState)
		if s.Enqueued == "E" {
			enqueues++
		}
		if step.LineID == "L7" {
			for _, id := range s.Queue {
				if id == s.Current {
					t.Fatal("popped current remains queued")
				}
			}
		}
		if step.LineID == "L9" && s.Order[len(s.Order)-1] != s.Current {
			t.Fatal("current missing from order")
		}
		if step.LineID == "L11" && !s.Skip {
			if s.Discovered[s.Neighbor] {
				t.Fatal("comparison prematurely discovers neighbor")
			}
			if step.HoldMS != 600 || steps[i+1].HoldMS != 1200 {
				t.Fatal("compare/commit budget")
			}
		}
	}
	if enqueues != 1 {
		t.Fatalf("E enqueued %d times", enqueues)
	}
	earlier := steps[1].State.(renderers.BFSLessonState)
	last.Distance["A"] = 99
	last.Discovered["A"] = false
	last.Finished["A"] = false
	last.Parent["B"] = "bad"
	last.Nodes[0].Label = "bad"
	last.Edges[0].To = "bad"
	last.Order[0] = "bad"
	if earlier.Distance["A"] != 0 || !earlier.Discovered["A"] || earlier.Nodes[0].Label != "A" || earlier.Edges[0].To != "B" || earlier.Parent["B"] != "" || earlier.Finished["A"] {
		t.Fatal("snapshot alias")
	}
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}}
	edges := []renderers.GraphEdge{{From: "A", To: "B"}, {From: "B", To: "A"}}
	cycle := Trace(nodes, edges, "A")
	end := cycle[len(cycle)-1].State.(renderers.BFSLessonState)
	if !reflect.DeepEqual(end.Order, Traverse(map[string][]string{"A": {"B"}, "B": {"A"}}, "A")) {
		t.Fatal("native mismatch")
	}
	if _, ok := end.Distance["C"]; ok {
		t.Fatal("isolated C has level")
	}
	for _, step := range cycle {
		s := step.State.(renderers.BFSLessonState)
		for i, e := range s.Edges {
			if e.Active != (i == s.EdgeIndex) {
				t.Fatal("reverse edge falsely active")
			}
		}
	}
	if len(Trace(nil, nil, "")) != 0 {
		t.Fatal("empty trace")
	}
}

func TestBFSSourceShowsSnapshotOperations(t *testing.T) {
	lines := map[string]string{}
	for _, line := range (Lesson{}).Source() {
		lines[line.ID] = line.Code
	}
	if !strings.Contains(lines["L7"], "queue[1:]") {
		t.Fatal("dequeue snapshot must highlight queue removal")
	}
	if !strings.Contains(lines["L12"], "append(queue, neighbor)") {
		t.Fatal("discovery snapshot must highlight enqueue")
	}
}

func TestBFSTraversalCorrectness(t *testing.T) {
	adj := map[string][]string{
		"A": {"B", "C"},
		"B": {"D", "E"},
		"C": {"E"},
		"D": {"F"},
		"E": {"F"},
		"F": {},
	}

	got := Traverse(adj, "A")
	expected := []string{"A", "B", "C", "D", "E", "F"}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("Traverse got %v, want %v", got, expected)
	}

	// Cycle graph test to ensure termination and correct order
	adjCycle := map[string][]string{
		"A": {"B"},
		"B": {"C", "A"},
		"C": {"A"},
	}
	gotCycle := Traverse(adjCycle, "A")
	expectedCycle := []string{"A", "B", "C"}
	if !reflect.DeepEqual(gotCycle, expectedCycle) {
		t.Fatalf("cycle Traverse got %v, want %v", gotCycle, expectedCycle)
	}

	// Single node
	gotSingle := Traverse(map[string][]string{"A": {}}, "A")
	if !reflect.DeepEqual(gotSingle, []string{"A"}) {
		t.Fatalf("single node got %v, want [A]", gotSingle)
	}

	// Empty start
	if gotEmpty := Traverse(adj, ""); gotEmpty != nil {
		t.Fatalf("empty start got %v, want nil", gotEmpty)
	}
}

func TestBFSLessonValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	meta := lesson.Meta()
	if meta.ID != "bfs" {
		t.Errorf("meta.ID = %q; want 'bfs'", meta.ID)
	}
	if meta.Category != catalog.CategoryGraphs {
		t.Errorf("meta.Category = %q; want %q", meta.Category, catalog.CategoryGraphs)
	}
	if meta.Renderer != "graph" {
		t.Errorf("meta.Renderer = %q; want 'graph'", meta.Renderer)
	}
}

func TestBFSTraceProperties(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Check that node states transition through unvisited, queued, visiting, visited
	seenStates := make(map[string]bool)
	sawActiveEdge := false

	for _, step := range steps {
		state, ok := step.State.(renderers.BFSLessonState)
		if !ok {
			t.Fatalf("step state is %T, want renderers.GraphState", step.State)
		}
		for _, n := range state.Nodes {
			seenStates[n.State] = true
		}
		for _, e := range state.Edges {
			if e.Active {
				sawActiveEdge = true
			}
		}
	}

	for _, requiredState := range []string{"unvisited", "queued", "visiting", "visited"} {
		if !seenStates[requiredState] {
			t.Errorf("missing expected node state %q across trace steps", requiredState)
		}
	}

	if !sawActiveEdge {
		t.Errorf("expected at least one active edge during traversal")
	}

	// Final step should have all reachable nodes visited and no active edges
	lastState := steps[len(steps)-1].State.(renderers.BFSLessonState)
	for _, n := range lastState.Nodes {
		if n.State != "visited" {
			t.Errorf("node %s final state = %q; want 'visited'", n.ID, n.State)
		}
	}
	for _, e := range lastState.Edges {
		if e.Active {
			t.Errorf("edge %s->%s is active at final step", e.From, e.To)
		}
	}
}
