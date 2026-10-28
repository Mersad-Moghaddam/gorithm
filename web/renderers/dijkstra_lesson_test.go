package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestDijkstraLessonSVG(t *testing.T) {
	s := DijkstraLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"long label here`, X: 1, Y: 1}, {ID: "B", Label: "B", X: 1, Y: 1}}, Edges: []GraphEdge{{From: "A", To: "B", Weight: 0}, {From: "A", To: "B", Weight: 2}, {From: "A", To: "A", Weight: 1}}}, Start: "A", Target: "B", Current: "A", Neighbor: "B", Distance: map[string]int{"A": 0, "B": 4}, Previous: map[string]string{"B": "A"}, Settled: map[string]bool{"A": true}, Queue: []DijkstraQueueEntry{{Node: "B", Distance: 4, Serial: 1}}, EdgeIndex: 0, HasCandidate: true, OldDistance: 4, Candidate: 0, PoppedSerial: 0}
	for _, width := range []int{320, 960} {
		step := trace.Step{LineID: "L10", State: s}
		svg := RenderDijkstraLessonSVG(step, width, 640)
		if svg != RenderDijkstraLessonSVG(step, width, 640) {
			t.Fatal("nondeterministic")
		}
		for _, want := range []string{`data-lesson="dijkstra"`, `data-scene-height=`, `data-entity="dijkstra:queue:1"`, `heap array`, `old d[v]`, `&lt;script&gt;`, `>0</text>`} {
			if !strings.Contains(svg, want) {
				t.Fatalf("missing %s", want)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "<script>") {
			t.Fatal("unsafe geometry/markup")
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

func TestDijkstraCandidateDoesNotPrejudgeComparison(t *testing.T) {
	s := DijkstraLessonState{HasCandidate: true, Candidate: 3, OldDistance: 4}
	svg := RenderDijkstraLessonSVG(trace.Step{LineID: "L9", State: s}, 320, 640)
	if strings.Contains(svg, "3 ≥ 4") || strings.Contains(svg, "3 &lt; 4") {
		t.Fatal("candidate step displays comparison before it happens")
	}
}

func TestDijkstraRootChipTextHasLineSpacing(t *testing.T) {
	s := DijkstraLessonState{Queue: []DijkstraQueueEntry{{Node: "E", Distance: 12}}}
	svg := RenderDijkstraLessonSVG(trace.Step{State: s}, 320, 640)
	if !strings.Contains(svg, `y="14.00">[0] E:12</text>`) || !strings.Contains(svg, `y="32.00">root / min</text>`) {
		t.Fatal("root chip labels need separate baselines")
	}
}
