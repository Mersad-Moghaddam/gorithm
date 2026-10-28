package floydwarshall

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestFloydWarshallTypedSnapshots(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatal(err)
	}
	want := [][]int64{{0, 3, 5, 6}, {5, 0, 2, 3}, {3, 6, 0, 1}, {2, 5, 7, 0}}
	final := steps[len(steps)-1].State.(renderers.FloydWarshallLessonState)
	if !reflect.DeepEqual(final.Distance, want) {
		t.Fatalf("matrix %v", final.Distance)
	}
	found := false
	for i, step := range steps {
		s := step.State.(renderers.FloydWarshallLessonState)
		if step.LineID == "L7" && s.I == 0 && s.J == 2 && s.K == 1 {
			if s.LeftDistance != 3 || s.RightDistance != 2 || s.OldDistance != 8 || s.Candidate != 5 || !s.HasCandidate {
				t.Fatalf("operands %+v", s)
			}
			next := steps[i+1].State.(renderers.FloydWarshallLessonState)
			if steps[i+1].LineID != "L8" || !next.Improved || next.Distance[0][2] != 5 || next.OldDistance != 8 {
				t.Fatal("missing commit")
			}
			if step.HoldMS != 600 || steps[i+1].HoldMS != 1200 {
				t.Fatal("pair timing")
			}
			found = true
		}
	}
	if !found {
		t.Fatal("missing A-B-C comparison")
	}
	first := steps[0].State.(renderers.FloydWarshallLessonState)
	final.Distance[0][0] = 99
	final.Nodes[0] = "changed"
	final.Edges[0].From = "changed"
	final.Cells[0].Value = "changed"
	final.RowHeaders[0] = "changed"
	if first.Distance[0][0] != 0 || first.Nodes[0] != "A" || first.Edges[0].From != "A" || first.Cells[0].Value != "0" || first.RowHeaders[0] != "A" {
		t.Fatal("mutable snapshot alias")
	}
}

func TestFloydWarshallExplicitEmptyEdges(t *testing.T) {
	steps, err := (Lesson{}).Build(json.RawMessage(`{"nodes":["A","B"],"edges":[]}`))
	if err != nil {
		t.Fatal(err)
	}
	first := steps[0]
	if first.LineID != "L2" {
		t.Fatalf("initial matrix points to %s, want L2", first.LineID)
	}
	s := first.State.(renderers.FloydWarshallLessonState)
	if len(s.Edges) != 0 || s.Distance[0][1] != math.MaxInt64 {
		t.Fatalf("empty graph initialized with edges: %+v", s)
	}
}

func TestFloydWarshallBoundedNegativeCycleArithmetic(t *testing.T) {
	nodes := make([]string, 16)
	for i := range nodes {
		nodes[i] = string(rune('A' + i))
	}
	edges := make([]WeightedEdge, 0, 120)
	for i := range nodes {
		for j := i + 1; j < len(nodes); j++ {
			edges = append(edges, WeightedEdge{nodes[i], nodes[j], -10000, false})
		}
	}
	input, err := json.Marshal(LessonInput{Nodes: nodes, Edges: edges})
	if err != nil {
		t.Fatal(err)
	}
	steps, err := (Lesson{}).Build(input)
	if err != nil {
		t.Fatal(err)
	}
	last := steps[len(steps)-1].State.(renderers.FloydWarshallLessonState)
	const want int64 = -3134566563840000
	if got := int64(last.Distance[15][15]); got != want || !last.NegativeCycle {
		t.Fatalf("distance=%d negativeCycle=%v; want %d and true", got, last.NegativeCycle, want)
	}
	dist, neg := AllPairsShortestPath(nodes, edges)
	wantAPI := want
	if wantAPI < int64(math.MinInt) {
		wantAPI = int64(math.MinInt)
	}
	if int64(dist[15][15]) != wantAPI || !neg {
		t.Fatalf("algorithm distance=%d negativeCycle=%v", dist[15][15], neg)
	}
}

func TestFloydWarshallAPIReturnType(t *testing.T) {
	var api func([]string, []WeightedEdge) ([][]int, bool) = AllPairsShortestPath
	_, _ = api(nil, nil)
}

func TestFloydWarshallSourceDescribesTraceDistances(t *testing.T) {
	source := (Lesson{}).Source()
	if !strings.Contains(source[0].Code, "[][]int64") || !strings.Contains(source[1].Code, "int64") {
		t.Fatalf("source does not describe 64-bit trace: %q / %q", source[0].Code, source[1].Code)
	}
}

func TestFloydWarshallGuardAndAlias(t *testing.T) {
	for _, fixture := range []struct {
		nodes []string
		edges []WeightedEdge
	}{
		{[]string{"A", "B", "C"}, []WeightedEdge{{"A", "B", 8, true}, {"A", "B", 3, true}, {"B", "C", 1, false}}},
		{[]string{"A"}, []WeightedEdge{{"A", "A", -1, true}}},
		{[]string{"A", "B"}, nil}, {nil, nil},
	} {
		steps := Trace(fixture.nodes, fixture.edges)
		want, neg := AllPairsShortestPath(fixture.nodes, fixture.edges)
		last := steps[len(steps)-1].State.(renderers.FloydWarshallLessonState)
		converted := make([][]int64, len(want))
		for i := range want {
			converted[i] = make([]int64, len(want[i]))
			for j, v := range want[i] {
				if v == math.MaxInt {
					converted[i][j] = math.MaxInt64
				} else {
					converted[i][j] = int64(v)
				}
			}
		}
		if !reflect.DeepEqual(last.Distance, converted) || last.NegativeCycle != neg {
			t.Fatalf("final %+v vs %v %v", last, want, neg)
		}
		if neg && steps[len(steps)-1].LineID != "L16" {
			t.Fatal("negative cycle reported successful")
		}
		for i, step := range steps {
			s := step.State.(renderers.FloydWarshallLessonState)
			if step.LineID == "L6" && s.HasCandidate {
				t.Fatal("infinity produced candidate")
			}
			if step.LineID == "L7" {
				if s.LeftDistance == math.MaxInt64 || s.Candidate != s.LeftDistance+s.RightDistance || s.OldDistance != s.Distance[s.I][s.J] {
					t.Fatal("wrong saved operands")
				}
			}
			if step.LineID == "L8" {
				before := steps[i-1].State.(renderers.FloydWarshallLessonState)
				if s.LeftDistance != before.LeftDistance || s.RightDistance != before.RightDistance || s.OldDistance != before.OldDistance {
					t.Fatal("aliased operands reread after write")
				}
			}
		}
	}
}

func TestFloydWarshallInputBounds(t *testing.T) {
	bad := []string{`null`, `{`, `{"nodes":null}`, `{"nodes":["A","A"]}`, `{"nodes":["A"],"edges":[{"from":"A","to":"B","weight":1}]}`, `{"edges":[{"from":"A","to":"B","weight":10001}]}`, `{"graphNodes":[{"id":"A","x":1000001}]}`, `{"nodes":["` + strings.Repeat("界", 65) + `"]}`, `{"edges":[null]}`}
	for _, raw := range bad {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	nodes := make([]string, 16)
	for i := range nodes {
		nodes[i] = strings.Repeat("x", i+1)
	}
	input, _ := json.Marshal(LessonInput{Nodes: nodes, Edges: []WeightedEdge{{nodes[0], nodes[1], 10000, true}}})
	if _, err := (Lesson{}).Build(input); err != nil {
		t.Fatal(err)
	}
	nodes = append(nodes, "seventeen")
	input, _ = json.Marshal(LessonInput{Nodes: nodes})
	if _, err := (Lesson{}).Build(input); err == nil {
		t.Fatal("accepted 17 nodes")
	}
	longID := strings.Repeat("界", 64)
	for _, limit := range []int{128, 129} {
		edges := make([]WeightedEdge, limit)
		for i := range edges {
			edges[i] = WeightedEdge{longID, "B", -10000, true}
		}
		input, _ := json.Marshal(LessonInput{Nodes: []string{longID, "B"}, Edges: edges})
		_, err := (Lesson{}).Build(input)
		if (err == nil) != (limit == 128) {
			t.Fatalf("edge boundary %d: %v", limit, err)
		}
	}
	for _, coordinate := range []float64{-1000000, 1000000, 1000001} {
		input, _ := json.Marshal(LessonInput{GraphNodes: []renderers.GraphNode{{ID: "A", Label: strings.Repeat("界", 64), X: coordinate}, {ID: "B"}}, GraphEdges: []renderers.GraphEdge{{From: "A", To: "B", Weight: 10000}}})
		_, err := (Lesson{}).Build(input)
		if (err == nil) != (coordinate <= 1000000) {
			t.Fatalf("coordinate boundary %v: %v", coordinate, err)
		}
	}
	for _, raw := range []string{`{"edges":[{"from":"A","to":"B","weight":-10001}]}`, `{"edges":[{"from":"A","to":"B","weight":null}]}`, `{"graphNodes":[{"id":"A","label":"` + strings.Repeat("界", 65) + `"}]}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
}
