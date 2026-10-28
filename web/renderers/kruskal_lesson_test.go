package renderers

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestKruskalLessonScene(t *testing.T) {
	s := KruskalLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"`, X: 0, Y: 0}, {ID: "B", Label: "B", X: 1, Y: 1}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 0}}}, Parent: map[string]string{"A": "A", "B": "A"}, Rank: map[string]int{"A": 1}, Components: [][]string{{"A", "B"}}, SortedEdgeIndices: []int{0}, SelectedEdges: []int{0}, ActiveEdge: 0, Connected: true}
	for _, width := range []int{320, 960} {
		for _, line := range []string{"L1", "L2", "L5", "L6", "L7", "L10"} {
			step := trace.Step{State: s, LineID: line}
			got := RenderKruskalLessonSVG(step, width, 640)
			if got != RenderKruskalLessonSVG(step, width, 640) {
				t.Fatal("nondeterministic scene")
			}
			for _, want := range []string{`data-lesson="kruskal"`, `data-scene-height=`, `data-entity="kruskal-node:0"`, `data-entity="kruskal-edge-chip:0"`, `DSU`, `&lt;script&gt;`, `>0</text>`} {
				if !strings.Contains(got, want) {
					t.Fatalf("missing %s", want)
				}
			}
			if strings.Contains(got, "<script>") || strings.Contains(got, "NaN") || strings.Contains(got, "Inf") {
				t.Fatal("unsafe scene")
			}
			decoder := xml.NewDecoder(strings.NewReader("<svg>" + got + "</svg>"))
			for {
				_, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
}

func TestKruskalParallelLoopsAndDenseLabels(t *testing.T) {
	s := KruskalLessonState{ActiveEdge: 1, Cursor: -1}
	for i := 0; i < 12; i++ {
		id := fmt.Sprintf("node%d", i)
		s.Nodes = append(s.Nodes, GraphNode{ID: id, Label: "long label " + id, X: float64(i), Y: float64(i % 2)})
	}
	s.Edges = []GraphEdge{{From: "node0", To: "node0", Weight: 1}, {From: "node0", To: "node0", Weight: 2}, {From: "node0", To: "node1", Weight: -10000}}
	s.SortedEdgeIndices = []int{0, 1, 2}
	for _, width := range []int{320, 960} {
		got := RenderKruskalLessonSVG(trace.Step{State: s}, width, 640)
		if strings.Contains(got, `data-edge-index="0"`) || !strings.Contains(got, `data-edge-index="1"`) || !strings.Contains(got, "Edges · all listed; active loop shown") {
			t.Fatalf("width %d: active loop not isolated", width)
		}
		s.ActiveEdge = 0
		other := RenderKruskalLessonSVG(trace.Step{State: s}, width, 640)
		if !strings.Contains(other, `data-edge-index="0"`) || strings.Contains(other, `data-edge-index="1"`) {
			t.Fatalf("width %d: first loop not isolated", width)
		}
		s.ActiveEdge = 1
		if strings.Contains(got, `data-graph-label=`) || !strings.Contains(got, "#0 → #1") || !strings.Contains(got, "-10000") || !strings.Contains(got, "long label node0") {
			t.Fatalf("width %d: dense labels not separated with edge identities", width)
		}
	}
}

func TestKruskalParallelEdgesShowGraphIdentities(t *testing.T) {
	s := KruskalLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", X: 1, Y: 1}, {ID: "B", X: 9, Y: 1}, {ID: "C", X: 5, Y: 5}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 1}, {From: "B", To: "A", Weight: 2}, {From: "A", To: "A", Weight: 3}, {From: "A", To: "A", Weight: 4}}}, SortedEdgeIndices: []int{0, 1, 2, 3}, ActiveEdge: 3, Cursor: -1}
	for _, width := range []int{320, 960} {
		got := RenderKruskalLessonSVG(trace.Step{State: s}, width, 640)
		for _, id := range []int{0, 1, 3} {
			if !strings.Contains(got, fmt.Sprintf(`data-edge-label="%d"`, id)) {
				t.Errorf("width %d: missing graph label for edge %d", width, id)
			}
		}
	}
}

func TestKruskalMaximumParallelEdgesStayInGraph(t *testing.T) {
	s := KruskalLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", X: 1, Y: 1}, {ID: "B", X: 9, Y: 1}}}, ActiveEdge: 127, Cursor: 127}
	for i := 0; i < 128; i++ {
		s.Edges = append(s.Edges, GraphEdge{From: "A", To: "B", Weight: i})
		s.SortedEdgeIndices = append(s.SortedEdgeIndices, i)
	}
	for _, width := range []int{320, 960} {
		got := RenderKruskalLessonSVG(trace.Step{State: s}, width, 640)
		if !strings.Contains(got, `data-edge-label="127"`) {
			t.Errorf("width %d: active edge label missing", width)
		}
		if strings.LastIndex(got, `data-edge-index="127"`) < strings.LastIndex(got, `data-edge-index="126"`) {
			t.Errorf("width %d: active edge drawn behind later paths", width)
		}
		for i := 0; i < 128; i++ {
			if !strings.Contains(got, fmt.Sprintf(`data-entity="kruskal-edge-chip:%d"`, i)) {
				t.Fatalf("width %d: edge %d missing from queue", width, i)
			}
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + got + "</svg>"))
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			el, ok := token.(xml.StartElement)
			if !ok || el.Name.Local != "path" {
				continue
			}
			var id, d string
			for _, a := range el.Attr {
				if a.Name.Local == "data-edge-index" {
					id = a.Value
				}
				if a.Name.Local == "d" {
					d = a.Value
				}
			}
			if id == "" {
				continue
			}
			var x1, y1, cx, cy, x2, y2 float64
			if n, _ := fmt.Sscanf(d, "M %f %f Q %f %f %f %f", &x1, &y1, &cx, &cy, &x2, &y2); n != 6 {
				t.Fatalf("bad edge path: %s", d)
			}
			gx, gy, gw, gh := 24.0, 64.0, 592.0, 348.0
			if width < 600 {
				gx, gy, gw, gh = 16, 56, 328, 240
			}
			for _, p := range [][2]float64{{x1, y1}, {cx, cy}, {x2, y2}} {
				if p[0] < gx || p[0] > gx+gw || p[1] < gy || p[1] > gy+gh {
					t.Errorf("width %d: edge %s outside graph: %v", width, id, p)
					break
				}
			}
		}
	}
}

func TestKruskalDenseParallelFocusAndLedger(t *testing.T) {
	s := KruskalLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", X: 1, Y: 1}, {ID: "B", X: 9, Y: 1}}}, ActiveEdge: 127, Cursor: 127, SelectedEdges: []int{0}, RejectedEdges: []int{126}}
	for i := 0; i < 128; i++ {
		s.Edges = append(s.Edges, GraphEdge{From: "A", To: "B", Weight: i})
		s.SortedEdgeIndices = append(s.SortedEdgeIndices, i)
	}
	for _, width := range []int{320, 960} {
		got := RenderKruskalLessonSVG(trace.Step{State: s}, width, 640)
		if strings.Count(got, `data-edge-index=`) != 2 || !strings.Contains(got, `data-edge-index="0"`) || !strings.Contains(got, `data-edge-index="127"`) {
			t.Errorf("width %d: dense group must show selected and active edges only", width)
		}
		if !strings.Contains(got, `data-edge-label="0"`) || !strings.Contains(got, `data-edge-label="127"`) || !strings.Contains(got, "Parallel edges · graph shows focus") {
			t.Errorf("width %d: missing focus explanation or identities", width)
		}
		for _, want := range []string{"Edge #0 A — B, weight 0: accepted", "Edge #126 A — B, weight 126: rejected", "Edge #127 A — B, weight 127: active"} {
			if !strings.Contains(got, want) {
				t.Errorf("width %d: missing accessible ledger description %q", width, want)
			}
		}
		for i := 0; i < 128; i++ {
			if !strings.Contains(got, fmt.Sprintf(`data-entity="kruskal-edge-chip:%d"`, i)) {
				t.Fatalf("width %d: missing edge %d", width, i)
			}
		}
		s.ActiveEdge = -1
		s.SelectedEdges = nil
		before := RenderKruskalLessonSVG(trace.Step{State: s}, width, 640)
		if strings.Count(before, `data-edge-index=`) != 1 || !strings.Contains(before, `data-edge-index="0"`) {
			t.Errorf("width %d: unselected group needs one representative", width)
		}
		s.ActiveEdge = 127
		s.SelectedEdges = []int{0}
	}
}

func TestKruskalSortingMotionAndParentPointers(t *testing.T) {
	s := KruskalLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 9}, {From: "A", To: "B", Weight: 0}}}, SortedEdgeIndices: []int{0, 1}, Parent: map[string]string{"A": "A", "B": "B"}, Components: [][]string{{"A"}, {"B"}}, ActiveEdge: -1, Cursor: -1}
	before := RenderKruskalLessonSVG(trace.Step{State: s, LineID: "L1"}, 320, 640)
	s.SortedEdgeIndices = []int{1, 0}
	s.Parent["B"] = "A"
	s.Components = [][]string{{"A", "B"}}
	after := RenderKruskalLessonSVG(trace.Step{State: s, LineID: "L6"}, 320, 640)
	if !strings.Contains(before, `data-entity="kruskal-edge-chip:0" data-x="24"`) || !strings.Contains(after, `data-entity="kruskal-edge-chip:0" data-x="168"`) {
		t.Fatal("sort lacks stable identity moving between recorded slots")
	}
	if !strings.Contains(after, `data-entity="kruskal-parent-pointer:1"`) || !strings.Contains(after, "parent is #0") || !strings.Contains(after, "✓ Merged; acceptance follows") {
		t.Fatal("union lacks pointer/teaching witness")
	}
}
