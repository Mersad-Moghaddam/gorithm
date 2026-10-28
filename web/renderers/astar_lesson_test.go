package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestAStarLessonSVG(t *testing.T) {
	state := AStarLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"`, X: 0, Y: 0}, {ID: "B", Label: "Goal", X: 3, Y: 4}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 0, Directed: true}}}, Start: "A", Goal: "B", Current: "A", G: map[string]int{"A": 0}, H: map[string]int{"A": 5, "B": 0}, F: map[string]int{"A": 5}, EdgeIndex: -1, Cost: -1, Open: []AStarOpenEntry{{Node: "A", G: 0, H: 5, F: 5}}}
	for _, width := range []int{320, 960} {
		step := trace.Step{LineID: "L5", State: state}
		got := RenderAStarLessonSVG(step, width, 640)
		if got != RenderAStarLessonSVG(step, width, 640) {
			t.Fatal("nondeterministic")
		}
		for _, want := range []string{`data-lesson="astar"`, `data-scene-height=`, `data-entity="astar-open:0"`, `data-from-entity="astar-open:0"`, `∞`, `estimate h = 5`, `not guaranteed shortest`, `&lt;script&gt;`, `>0</text>`} {
			if !strings.Contains(got, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(got, "<script>") || strings.Contains(got, "NaN") {
			t.Fatal("unsafe SVG")
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
