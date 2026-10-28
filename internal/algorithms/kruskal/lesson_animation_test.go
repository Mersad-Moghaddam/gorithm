package kruskal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestKruskalTypedTrace(t *testing.T) {
	nodes, edges := defaultGraph()
	steps := Trace(nodes, edges)
	if reflect.TypeOf(steps[0].State).Name() != "KruskalLessonState" {
		t.Fatalf("need typed DSU state, got %T", steps[0].State)
	}
}

func TestKruskalRejectMalformedInput(t *testing.T) {
	for _, input := range []string{`null`, `{`, `{"nodes":[{"id":"A"},{"id":"A"}],"edges":[{"from":"A","to":"A"}]}`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"missing"}]}`, `{"nodes":[{"id":"A"}],"edges":[{"from":"A","to":"A","weight":10001}]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
}

func TestKruskalUnionStagesAndIsolation(t *testing.T) {
	nodes, edges := defaultGraph()
	steps := Trace(nodes, edges)
	last := steps[len(steps)-1].State.(renderers.KruskalLessonState)
	got := slices.Clone(last.SelectedEdges)
	slices.Sort(got)
	if !slices.Equal(got, []int{1, 3, 4, 6, 8}) || last.TotalWeight != 12 || !last.Connected {
		t.Fatalf("bad MST: %+v", last)
	}
	for i, step := range steps {
		s := step.State.(renderers.KruskalLessonState)
		if step.LineID == "L6" {
			prev := steps[i-1].State.(renderers.KruskalLessonState)
			next := steps[i+1].State.(renderers.KruskalLessonState)
			if len(s.Components) != len(prev.Components)-1 || len(s.SelectedEdges) != len(prev.SelectedEdges) || len(next.SelectedEdges) != len(s.SelectedEdges)+1 {
				t.Fatal("union and acceptance conflated")
			}
			if steps[i-1].HoldMS+step.HoldMS != 1800 {
				t.Fatal("bad pair timing")
			}
		}
	}
	frozen := steps[2].State.(renderers.KruskalLessonState)
	last.Parent["A"] = "mutated"
	last.Rank["A"] = 99
	last.Nodes[0].Label = "mutated"
	last.Edges[0].Weight = 99
	last.Components[0][0] = "mutated"
	last.SortedEdgeIndices[0] = 99
	if frozen.Parent["A"] != "A" || frozen.Rank["A"] == 99 || frozen.Nodes[0].Label == "mutated" || frozen.Edges[0].Weight == 99 || frozen.Components[0][0] == "mutated" || frozen.SortedEdgeIndices[0] == 99 {
		t.Fatal("snapshot mutation leaked")
	}
}

func TestKruskalForestLoopsTiesAndParallel(t *testing.T) {
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}}
	for _, edges := range [][]WeightedEdge{
		{{"A", "B", 1}, {"C", "D", 2}},
		{{"A", "A", -2}, {"A", "B", 0}, {"A", "B", 1}, {"B", "C", 2}},
		{{"A", "B", 1}, {"A", "C", 1}, {"B", "C", 1}, {"C", "D", 1}},
	} {
		steps := Trace(nodes, edges)
		last := steps[len(steps)-1].State.(renderers.KruskalLessonState)
		native, cost := MST([]string{"A", "B", "C", "D"}, edges)
		if cost != last.TotalWeight || len(native) != len(last.SelectedEdges) {
			t.Fatal("native result mismatch")
		}
		if !reflect.DeepEqual(steps, Trace(nodes, edges)) {
			t.Fatal("trace nondeterministic")
		}
		dsu := NewDSU([]string{"A", "B", "C", "D"})
		for _, id := range last.SelectedEdges {
			e := edges[id]
			if !dsu.Union(e.From, e.To) {
				t.Fatal("accepted cycle")
			}
		}
		if len(edges) == 2 && (last.Connected || last.TotalWeight != 3) {
			t.Fatal("forest mislabeled")
		}
		if len(edges) == 4 && edges[0].From == edges[0].To && (!slices.Contains(last.RejectedEdges, 0) || !slices.Contains(last.RejectedEdges, 2) || !last.Edges[1].Active || last.Edges[2].Active) {
			t.Fatal("parallel or self-loop identity lost")
		}
	}
}

func TestKruskalInputLimits(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*LessonInput)
		valid  bool
	}{
		{"max weight", func(in *LessonInput) { in.Edges[0].Weight = 10000 }, true},
		{"min weight", func(in *LessonInput) { in.Edges[0].Weight = -10000 }, true},
		{"weight below", func(in *LessonInput) { in.Edges[0].Weight = -10001 }, false},
		{"coordinate limit", func(in *LessonInput) { in.Nodes[0].X = 1000000; in.Nodes[0].Y = -1000000 }, true},
		{"coordinate over", func(in *LessonInput) { in.Nodes[0].X = 1000001 }, false},
		{"label limit", func(in *LessonInput) { in.Nodes[0].Label = string(make([]rune, 64)) }, true},
		{"label over", func(in *LessonInput) { in.Nodes[0].Label = string(make([]rune, 65)) }, false},
		{"id limit", func(in *LessonInput) { id := string(make([]rune, 64)); in.Nodes[0].ID = id; in.Edges[0].From = id }, true},
		{"id over", func(in *LessonInput) { in.Nodes[0].ID = string(make([]rune, 65)) }, false},
		{"32 nodes", func(in *LessonInput) {
			for i := 2; i < 32; i++ {
				in.Nodes = append(in.Nodes, renderers.GraphNode{ID: fmt.Sprint(i)})
			}
		}, true},
		{"33 nodes", func(in *LessonInput) {
			for i := 2; i < 33; i++ {
				in.Nodes = append(in.Nodes, renderers.GraphNode{ID: fmt.Sprint(i)})
			}
		}, false},
		{"128 edges", func(in *LessonInput) {
			for len(in.Edges) < 128 {
				in.Edges = append(in.Edges, in.Edges[0])
			}
		}, true},
		{"129 edges", func(in *LessonInput) {
			for len(in.Edges) < 129 {
				in.Edges = append(in.Edges, in.Edges[0])
			}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := LessonInput{Nodes: []renderers.GraphNode{{ID: "A"}, {ID: "B"}}, Edges: []WeightedEdge{{"A", "B", 1}}}
			tc.mutate(&in)
			raw, _ := json.Marshal(in)
			_, err := (Lesson{}).Build(raw)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v err=%v", tc.valid, err)
			}
		})
	}
}
