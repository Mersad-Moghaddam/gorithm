package dfs

import (
	"encoding/json"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strings"
	"testing"
)

func TestDFSRequiresTypedSnapshots(t *testing.T) {
	steps, err := (Lesson{}).Build(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range steps {
		if got := fmt.Sprintf("%T", step.State); got != "renderers.DFSLessonState" {
			t.Fatalf("snapshot type = %s; want renderers.DFSLessonState", got)
		}
	}
}

func TestDFSStackOrdersAndIsolation(t *testing.T) {
	steps, _ := (Lesson{}).Build(nil)
	final := steps[len(steps)-1].State.(renderers.DFSLessonState)
	want := Traverse(map[string][]string{"A": {"B", "C"}, "B": {"D", "E"}, "C": {"E", "F"}}, "A")
	if !reflect.DeepEqual(final.Order, want) || !reflect.DeepEqual(final.FinishOrder, []string{"D", "E", "B", "F", "C", "A"}) || len(final.Stack) != 0 {
		t.Fatalf("wrong final state: %+v", final)
	}
	for i, step := range steps {
		s := step.State.(renderers.DFSLessonState)
		if step.LineID == "L10" {
			prev := steps[i-1].State.(renderers.DFSLessonState)
			if len(s.Stack) != len(prev.Stack)+1 || s.Stack[len(s.Stack)-1].Node != s.Current {
				t.Fatal("child call missing")
			}
			if steps[i-1].HoldMS+step.HoldMS != 1800 {
				t.Fatal("pair budget")
			}
		}
		if step.LineID == "L11" {
			if s.Stack[len(s.Stack)-1].Node != s.Current || !s.Finished[s.ReturningFrom] {
				t.Fatal("parent not restored")
			}
		}
		if step.LineID == "L9" && s.Current == "C" && s.Neighbor == "E" {
			if !s.Skip || s.BackEdge {
				t.Fatal("finished neighbor called back-edge")
			}
		}
		got := renderers.RenderDFSLessonSVG(step, 320, 640)
		if got != renderers.RenderDFSLessonSVG(step, 320, 640) {
			t.Fatal("seek nondeterministic")
		}
	}
	first := steps[0].State.(renderers.DFSLessonState)
	final.Nodes[0].Label = "mutated"
	final.Edges[0].From = "mutated"
	final.Order[0] = "mutated"
	final.FinishOrder[0] = "mutated"
	final.Parent["B"] = "mutated"
	final.Discovered["A"] = false
	final.Finished["A"] = false
	if first.Nodes[0].Label != "A" || first.Edges[0].From != "A" || len(first.Order) != 0 || len(first.Parent) != 0 || len(first.Discovered) != 0 || len(first.Finished) != 0 {
		t.Fatal("snapshot alias")
	}
	for i, step := range steps {
		s := step.State.(renderers.DFSLessonState)
		if len(s.Stack) > 1 {
			old := s.Stack[0].Node
			s.Stack[0].Node = "changed"
			for j := i + 1; j < len(steps); j++ {
				other := steps[j].State.(renderers.DFSLessonState)
				if len(other.Stack) > 0 && other.Stack[0].Node != old {
					t.Fatal("frame slice alias")
				}
			}
			break
		}
	}
}

func TestDFSCycleSingletonAndGuards(t *testing.T) {
	nodes := []renderers.GraphNode{{ID: "A"}, {ID: "B"}, {ID: "Z"}}
	steps := Trace(nodes, []renderers.GraphEdge{{From: "A", To: "B"}, {From: "B", To: "A"}}, "A")
	found := false
	for _, step := range steps {
		s := step.State.(renderers.DFSLessonState)
		if step.LineID == "L9" && s.Current == "B" {
			found = s.BackEdge && s.Skip
		}
	}
	final := steps[len(steps)-1].State.(renderers.DFSLessonState)
	if !found || final.Discovered["Z"] || len(final.Order) != 2 {
		t.Fatal("cycle/unreachable incorrect")
	}
	single := Trace(nodes[:1], nil, "")
	s := single[len(single)-1].State.(renderers.DFSLessonState)
	if !reflect.DeepEqual(s.Order, []string{"A"}) || !reflect.DeepEqual(s.FinishOrder, []string{"A"}) || len(s.Stack) != 0 {
		t.Fatal("singleton incorrect")
	}
	if len(Trace(nil, nil, "")) != 0 || len(Trace(nodes, nil, "missing")) != 0 {
		t.Fatal("guard behavior changed")
	}
}

func TestDFSInputLimits(t *testing.T) {
	for _, limit := range []struct {
		key       string
		good, bad any
	}{
		{"x", 1000000, 1000001}, {"y", -1000000, -1000001}, {"label", strings.Repeat("界", 64), strings.Repeat("界", 65)},
	} {
		for _, valid := range []bool{true, false} {
			value := limit.bad
			if valid {
				value = limit.good
			}
			node := map[string]any{"id": "A", limit.key: value}
			raw, _ := json.Marshal(map[string]any{"nodes": []any{node}, "edges": []any{map[string]any{"from": "A", "to": "A"}}})
			_, err := (Lesson{}).Build(raw)
			if (err == nil) != valid {
				t.Fatalf("%s boundary valid=%v: %v", limit.key, valid, err)
			}
		}
	}
	for _, n := range []int{32, 33} {
		nodes := []renderers.GraphNode{}
		for i := 0; i < n; i++ {
			nodes = append(nodes, renderers.GraphNode{ID: fmt.Sprint(i)})
		}
		raw, _ := json.Marshal(map[string]any{"start": "0", "nodes": nodes, "edges": []renderers.GraphEdge{{From: "0", To: "0"}}})
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 32) {
			t.Fatal("node count boundary", err)
		}
	}
	for _, n := range []int{128, 129} {
		edges := make([]renderers.GraphEdge, n)
		for i := range edges {
			edges[i] = renderers.GraphEdge{From: "A", To: "B", Weight: 10000}
		}
		raw, _ := json.Marshal(map[string]any{"edges": edges})
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 128) {
			t.Fatal("edge count boundary", err)
		}
	}
	for _, n := range []int{64, 65} {
		id := strings.Repeat("界", n)
		raw, _ := json.Marshal(map[string]any{"start": id, "nodes": []renderers.GraphNode{{ID: id}}, "edges": []renderers.GraphEdge{{From: id, To: id, Weight: -10000}}})
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 64) {
			t.Fatal("ID boundary", err)
		}
	}
}

func TestDFSRejectsInvalidInput(t *testing.T) {
	for _, input := range []string{`null`, `{`, `123`, `{"start":null}`, `{"nodes":null}`, `{"edges":null}`, `{"start":"missing"}`, `{"nodes":[{"id":"A"},{"id":"A"}]}`, `{"edges":[{"from":"A","to":"missing"}]}`, `{"nodes":[{"id":"A","x":1000001}]}`, `{"edges":[{"from":"A","to":"B","weight":10001}]}`} {
		t.Run(input, func(t *testing.T) {
			if _, err := (Lesson{}).Build(json.RawMessage(input)); err == nil {
				t.Fatal("invalid input accepted")
			}
		})
	}
}
