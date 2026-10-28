package toposort

import (
	"encoding/json"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestToposortTypedSnapshots(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range steps {
		if reflect.TypeOf(step.State).Name() != "ToposortLessonState" {
			t.Fatalf("%s state = %T, want typed prerequisite snapshot", step.LineID, step.State)
		}
	}
	for _, i := range []int{0, len(steps) / 2, len(steps) - 1} {
		first := renderers.RenderToposortLessonSVG(steps[i], 320, 640)
		if first != renderers.RenderToposortLessonSVG(steps[i], 320, 640) {
			t.Fatal("snapshot render depends on history")
		}
	}
	for _, step := range steps {
		s := step.State.(renderers.ToposortLessonState)
		got := renderers.RenderToposortLessonSVG(step, 960, 640)
		if step.LineID == "L7" && !strings.Contains(got, `data-from-entity="toposort:queue:`) {
			t.Fatal("missing FIFO pop motion")
		}
		if step.LineID == "L8" && !strings.Contains(got, `data-from-entity="toposort:current:`) {
			t.Fatal("missing order copy motion")
		}
		if step.LineID == "L12" && s.Ready && !strings.Contains(got, `data-from-entity="toposort:node:`) {
			t.Fatal("missing ready append motion")
		}
	}
}

func TestToposortPrerequisitesAndResults(t *testing.T) {
	steps, _ := (Lesson{}).Build(nil)
	countdown := []int{steps[0].State.(renderers.ToposortLessonState).InDegree["package"]}
	enqueued := 0
	for i, step := range steps {
		s := step.State.(renderers.ToposortLessonState)
		if step.LineID == "L10" && s.Neighbor == "package" {
			countdown = append(countdown, s.NewInDegree)
			if !s.Decremented || s.OldInDegree != s.NewInDegree+1 {
				t.Fatal("bad decrement")
			}
			if i == 0 || steps[i-1].HoldMS+step.HoldMS != 1800 {
				t.Fatal("bad pair timing")
			}
		}
		if step.LineID == "L12" && s.Neighbor == "package" {
			enqueued++
			if s.InDegree["package"] != 0 {
				t.Fatal("premature enqueue")
			}
		}
		if step.LineID == "L8" && s.Current == "clean" && len(s.ResolvedEdges) != 0 {
			t.Fatal("outgoing edges resolved before L10")
		}
	}
	if !reflect.DeepEqual(countdown, []int{3, 2, 1, 0}) || enqueued != 1 {
		t.Fatalf("countdown %v, enqueue %d", countdown, enqueued)
	}
	final := steps[len(steps)-1].State.(renderers.ToposortLessonState)
	want := []string{"clean", "lint", "compile", "test", "docs", "package", "deploy"}
	if !final.Terminal || !final.Valid || !reflect.DeepEqual(final.Order, want) {
		t.Fatalf("result %+v", final)
	}
	positions := map[string]int{}
	for i, id := range final.Order {
		positions[id] = i
	}
	for _, e := range final.Edges {
		if positions[e.From] >= positions[e.To] {
			t.Fatal("edge order violated")
		}
	}
	for _, edges := range [][]renderers.GraphEdge{{{From: "A", To: "B"}, {From: "B", To: "A"}, {From: "B", To: "C"}}, {{From: "A", To: "A"}}, {{From: "A", To: "B"}, {From: "A", To: "B"}}, nil} {
		nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "C"}}
		run := Trace(nodes, edges)
		got := run[len(run)-1].State.(renderers.ToposortLessonState)
		pairs := [][2]string{}
		for _, e := range edges {
			pairs = append(pairs, [2]string{e.From, e.To})
		}
		order, valid := Sort([]string{"A", "B", "C"}, pairs)
		if !slices.Equal(got.Order, order) || got.Valid != valid {
			t.Fatalf("native mismatch %v", edges)
		}
		if len(edges) == 3 && !reflect.DeepEqual(got.Blocked, []string{"A", "B", "C"}) {
			t.Fatal("downstream task not blocked")
		}
	}
}

func TestToposortSnapshotIsolation(t *testing.T) {
	steps, _ := (Lesson{}).Build(nil)
	before, _ := json.Marshal(steps[0])
	last := steps[len(steps)-1].State.(renderers.ToposortLessonState)
	last.InDegree["package"] = 99
	last.Processed["clean"] = false
	last.Nodes[0].Label = "changed"
	last.Edges[0].To = "changed"
	last.Order[0] = "changed"
	last.ResolvedEdges[0] = 99
	for _, step := range steps {
		s := step.State.(renderers.ToposortLessonState)
		if len(s.Queue) > 0 {
			s.Queue[0] = "changed"
			break
		}
	}
	after, _ := json.Marshal(steps[0])
	if string(before) != string(after) {
		t.Fatal("earlier snapshot mutated")
	}
}

func TestToposortInputLimits(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input buildInput
		bad   bool
	}{
		{"bounds", buildInput{Nodes: []renderers.GraphNode{{ID: strings.Repeat("界", 64), Label: strings.Repeat("界", 64), X: 1000000, Y: -1000000}}, Edges: []renderers.GraphEdge{{From: strings.Repeat("界", 64), To: strings.Repeat("界", 64), Weight: 10000}}}, false},
		{"ID", buildInput{Nodes: []renderers.GraphNode{{ID: strings.Repeat("界", 65)}}}, true},
		{"label", buildInput{Nodes: []renderers.GraphNode{{ID: "A", Label: strings.Repeat("界", 65)}}}, true},
		{"weight", buildInput{Edges: []renderers.GraphEdge{{From: "clean", To: "lint", Weight: -10001}}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, _ := json.Marshal(tc.input) // omitted arrays retain defaults; remove null fields.
			fields := map[string]json.RawMessage{}
			_ = json.Unmarshal(raw, &fields)
			for k, v := range fields {
				if string(v) == "null" {
					delete(fields, k)
				}
			}
			raw, _ = json.Marshal(fields)
			_, err := (Lesson{}).Build(raw)
			if (err != nil) != tc.bad {
				t.Fatalf("err=%v bad=%v", err, tc.bad)
			}
		})
	}
	nodes := make([]renderers.GraphNode, 32)
	for i := range nodes {
		nodes[i].ID = string(rune('A' + i))
	}
	edges := make([]renderers.GraphEdge, 128)
	for i := range edges {
		edges[i] = renderers.GraphEdge{From: nodes[0].ID, To: nodes[1].ID}
	}
	for _, bad := range []bool{false, true} {
		n, e := nodes, edges
		if bad {
			n = append(append([]renderers.GraphNode{}, nodes...), renderers.GraphNode{ID: "extra"})
		}
		raw, _ := json.Marshal(buildInput{n, e})
		_, err := (Lesson{}).Build(raw)
		if (err != nil) != bad {
			t.Fatal("node limit", err)
		}
	}
	edges = append(edges, edges[0])
	raw, _ := json.Marshal(buildInput{nodes, edges})
	if _, err := (Lesson{}).Build(raw); err == nil {
		t.Fatal("edge limit")
	}
}

func TestToposortRejectInvalidInput(t *testing.T) {
	for _, input := range []string{`null`, `{`, `[]`, `{"nodes":[{"id":"A"},{"id":"A"}]}`, `{"edges":[{"from":"clean","to":"missing"}]}`, `{"nodes":null}`, `{"nodes":[{"id":"A","x":1000001}]}`} {
		if _, err := (Lesson{}).Build([]byte(input)); err == nil {
			t.Errorf("accepted invalid input %s", input)
		}
	}
}
