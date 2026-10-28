package renderers

import (
	"encoding/xml"
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
