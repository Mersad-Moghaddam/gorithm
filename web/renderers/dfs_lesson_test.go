package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestDFSSceneSettledEscapedAndMotion(t *testing.T) {
	s := DFSLessonState{GraphState: GraphState{Nodes: []GraphNode{{ID: "A", Label: `<script>&"long node label`, X: 0, Y: 0}, {ID: "B", Label: "B", X: 0, Y: 0}}, Edges: []GraphEdge{{From: "A", To: "B"}, {From: "B", To: "A"}, {From: "A", To: "A"}}}, Current: "B", Neighbor: "B", Stack: []DFSCallFrame{{Node: "A"}, {Node: "B", Caller: "A"}}, Order: []string{"A", "B"}, Discovered: map[string]bool{"A": true, "B": true}, EdgeIndex: 0}
	for _, w := range []int{320, 960} {
		step := trace.Step{LineID: "L10", State: s}
		got := RenderDFSLessonSVG(step, w, 640)
		if got != RenderDFSLessonSVG(step, w, 640) {
			t.Fatal("nondeterministic scene")
		}
		for _, want := range []string{`data-lesson="dfs"`, `data-phase="call"`, `data-entity="dfs:cursor"`, `data-entity="dfs:frame:1"`, `data-from-entity="dfs:node:1"`, `Preorder`, `Finish order`, `TOP`, `&lt;script&gt;`} {
			if !strings.Contains(got, want) {
				t.Fatalf("missing %s", want)
			}
		}
		if strings.Contains(got, "<script>") || strings.Contains(got, "NaN") || strings.Contains(got, "Inf") {
			t.Fatal("unsafe geometry or text")
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + got + "</svg>"))
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
