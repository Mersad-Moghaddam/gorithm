package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestToposortScene(t *testing.T) {
	s := ToposortLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"`, X: 0, Y: 0}, {ID: "B", Label: "dependency", X: 1, Y: 0}}, Edges: []GraphEdge{{From: "A", To: "B"}, {From: "A", To: "B"}, {From: "B", To: "A"}, {From: "A", To: "A"}}}, InDegree: map[string]int{"A": 1, "B": 0}, Processed: map[string]bool{"A": true}, Order: []string{"A"}, Queue: []string{"B"}, Current: "A", Neighbor: "B", EdgeIndex: 1, OldInDegree: 1, NewInDegree: 0, Decremented: true, ResolvedEdges: []int{0, 1}}
	for _, width := range []int{320, 960} {
		step := trace.Step{LineID: "L10", State: s}
		got := RenderToposortLessonSVG(step, width, 640)
		if got != RenderToposortLessonSVG(step, width, 640) {
			t.Fatal("nondeterministic renderer")
		}
		for _, want := range []string{`data-lesson="toposort"`, `data-phase="L10"`, `data-scene-height=`, `data-entity="toposort:node:0"`, `data-entity="toposort:token:1"`, `Remaining prerequisites`, `Ready FIFO`, `Output order`, `1 − 1 = 0`, `stroke-dasharray`, `&lt;script&gt;`} {
			if !strings.Contains(got, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(got, "<script>") || strings.Contains(got, "NaN") || strings.Contains(got, "+Inf") {
			t.Fatal("unsafe output")
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + got + "</svg>"))
		entities := map[string]bool{}
		for {
			token, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if start, ok := token.(xml.StartElement); ok {
				for _, attr := range start.Attr {
					if attr.Name.Local == "data-entity" {
						if entities[attr.Value] {
							t.Fatalf("duplicate entity %s", attr.Value)
						}
						entities[attr.Value] = true
					}
				}
			}
		}
	}
}

func TestToposortEmptyScene(t *testing.T) {
	got := RenderToposortLessonSVG(trace.Step{State: ToposortLessonState{Terminal: true, Valid: true}}, 320, 300)
	if !strings.Contains(got, "Valid topological order") {
		t.Fatal(got)
	}
}
