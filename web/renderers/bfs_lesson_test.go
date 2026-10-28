package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestBFSLessonSVG(t *testing.T) {
	s := BFSLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"你好`}, {ID: "B", Label: "B"}}, Edges: []GraphEdge{{From: "A", To: "B"}, {From: "B", To: "A"}, {From: "A", To: "A"}}}, Start: "A", Current: "A", Neighbor: "B", Queue: []string{"B"}, Order: []string{"A"}, Distance: map[string]int{"A": 0, "B": 1}, Parent: map[string]string{"B": "A"}, Discovered: map[string]bool{"A": true, "B": true}, Finished: map[string]bool{}, EdgeIndex: 0, Enqueued: "B"}
	for _, width := range []int{320, 960} {
		for _, line := range []string{"L1", "L11", "L12", "L17"} {
			step := trace.Step{State: s, LineID: line}
			got := RenderBFSLessonSVG(step, width, 640)
			if got != RenderBFSLessonSVG(step, width, 640) {
				t.Fatal("nondeterministic")
			}
			if strings.Contains(got, "<script>") || strings.Contains(got, "NaN") || strings.Contains(got, "Inf") {
				t.Fatal("unsafe SVG")
			}
			d := xml.NewDecoder(strings.NewReader("<svg>" + got + "</svg>"))
			entities := map[string]bool{}
			for {
				tok, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				if e, ok := tok.(xml.StartElement); ok {
					for _, a := range e.Attr {
						if a.Name.Local == "data-entity" {
							if entities[a.Value] {
								t.Fatal("duplicate identity")
							}
							entities[a.Value] = true
						}
						if strings.HasPrefix(a.Name.Local, "on") {
							t.Fatal("event injection")
						}
					}
				}
			}
			for _, want := range []string{`data-lesson="bfs"`, `data-from-entity="bfs-node:1"`, `data-min-width=`, `FIFO`, `Visit order`, `level`, `parent`, `bfs-edge-2`, `&lt;script&gt;`} {
				if !strings.Contains(got, want) {
					t.Errorf("missing %s", want)
				}
			}
		}
	}
	s.Enqueued = ""
	s.Dequeued = "A"
	if got := RenderBFSLessonSVG(trace.Step{State: s, LineID: "L7"}, 320, 640); !strings.Contains(got, `data-from-entity="bfs-queue:0"`) {
		t.Fatal("missing dequeue alias")
	}
}
