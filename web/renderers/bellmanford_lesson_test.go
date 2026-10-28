package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestBellmanFordLessonSVG(t *testing.T) {
	st := BellmanFordLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"界`, X: 1, Y: 1}, {ID: "B", Label: "B", X: 1, Y: 1}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 0}, {From: "B", To: "A", Weight: -2}, {From: "A", To: "A", Weight: 1}}}, Start: "A", From: "A", To: "B", Distance: map[string]int{"A": 0, "B": 3}, PassStartDistance: map[string]int{"A": 0, "B": 5}, Pass: 1, MaxPasses: 1, EdgeIndex: 0, HasCandidate: true, Candidate: 0, OldDistance: 3, Improved: true, WitnessEdge: -1}
	for _, w := range []int{320, 960} {
		for _, line := range []string{"L1", "L6", "L7", "L18"} {
			step := trace.Step{LineID: line, State: st}
			svg := RenderBellmanFordLessonSVG(step, w, 640)
			if svg != RenderBellmanFordLessonSVG(step, w, 640) {
				t.Fatal("nondeterministic")
			}
			for _, want := range []string{`data-lesson="bellmanford"`, `data-scene-height=`, `data-min-width=`, `bellmanford:cursor`, `marker-end=`, `&lt;script&gt;`, `>0</text>`, `Pass start`, `Current`} {
				if !strings.Contains(svg, want) {
					t.Errorf("missing %s", want)
				}
			}
			if strings.Contains(svg, "<script>") || strings.Contains(svg, "NaN") || strings.Contains(svg, "+Inf") {
				t.Fatal("unsafe SVG")
			}
			d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
			for {
				_, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	st.CheckingCycle = true
	st.NegativeCycle = true
	st.WitnessEdge = 0
	svg := RenderBellmanFordLessonSVG(trace.Step{LineID: "L15", State: st}, 320, 640)
	if !strings.Contains(svg, "Reachable negative cycle") || strings.Contains(svg, "✓ shortest") {
		t.Fatal("cycle celebrated")
	}
}
