package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestPrimLessonSVG(t *testing.T) {
	s := PrimLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"`, X: 0, Y: 0}, {ID: "B", Label: "long label with Unicode 世界", X: 1, Y: 1}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 0}, {From: "A", To: "B", Weight: -2}}}, InTree: map[string]bool{"A": true}, Candidates: []PrimCandidate{{From: "A", To: "B", Weight: -2, EdgeIndex: 1, Serial: 3}}, ActiveEdge: 1, CandidateWeight: -2, CutEdges: []int{0, 1}}
	for _, width := range []int{320, 960} {
		s.ActiveSerial = 3
		step := trace.Step{LineID: "L7", State: s}
		svg := RenderPrimLessonSVG(step, width, 640)
		if svg != RenderPrimLessonSVG(step, width, 640) {
			t.Fatal("nondeterministic")
		}
		for _, want := range []string{`data-lesson="prim"`, `data-entity="prim-node:0"`, `data-from-entity="prim-heap:3"`, `&lt;script&gt;`, `>0</text>`, `>-2</text>`, `Cut edges:`, `Candidate heap`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(svg, "<script>") || strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
			t.Fatal("unsafe geometry/text")
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

func TestPrimMotionAndIndependentSnapshots(t *testing.T) {
	s := PrimLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: "A"}, {ID: "B", Label: "B"}, {ID: "C", Label: "C"}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 1}}}, ActiveEdge: 0, ActiveSerial: 2, InTree: map[string]bool{"A": true, "B": true}, SelectedEdges: []int{0}, TotalWeight: 1, CandidateWeight: 1}
	for _, phase := range []string{"L1", "L7", "L9", "L12", "L13", "L18"} {
		s.Discarded = phase == "L9"
		step := trace.Step{LineID: phase, State: s}
		svg := RenderPrimLessonSVG(step, 320, 640)
		if svg != RenderPrimLessonSVG(step, 320, 640) {
			t.Fatal("history-dependent output")
		}
		if phase == "L12" && !strings.Contains(svg, `data-from-entity="prim-probe:2"`) {
			t.Fatal("missing accepted-edge transfer")
		}
		if phase == "L13" && !strings.Contains(svg, `data-entity="prim-sum:2"`) {
			t.Fatal("missing sum token")
		}
		if phase == "L9" && !strings.Contains(svg, "Reject internal edge") {
			t.Fatal("missing rejection label")
		}
	}
}
