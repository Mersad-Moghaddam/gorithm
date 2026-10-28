package renderers

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestKosarajuSceneDeterministicEscapedAndResponsive(t *testing.T) {
	nodes := []GraphNode{{ID: "A", Label: `<script onload="bad">&界`, X: 1, Y: 1}, {ID: "B", Label: strings.Repeat("界", 64), X: 1, Y: 1}}
	for _, pass := range []int{1, 2} {
		for _, width := range []int{320, 960} {
			s := KosarajuLessonState{GraphState: GraphState{Nodes: nodes, Edges: []GraphEdge{{From: "A", To: "B"}, {From: "B", To: "A"}, {From: "A", To: "B"}, {From: "A", To: "A"}}}, Pass: pass, Transposed: pass == 2, Stack: []string{"A"}, FinishOrder: []string{"B", "A"}, Current: "A", FinishedVertex: "A", OrderIndex: 1, ComponentOf: map[string]int{"A": 0}, CurrentComponent: []string{"A"}}
			step := trace.Step{State: s, LineID: "L5", Action: trace.ActionPush, Explanation: `<script>&`}
			got := RenderKosarajuLessonSVG(step, width, 640)
			if got != RenderKosarajuLessonSVG(step, width, 640) {
				t.Fatal("nondeterministic")
			}
			if strings.Contains(got, "<script") || strings.Contains(got, `onload="bad"`) || strings.Contains(got, "NaN") || strings.Contains(got, "Inf") {
				t.Fatal("unsafe SVG")
			}
			for _, want := range []string{`data-lesson="kosaraju"`, `data-from-entity="kosaraju-node:0"`, `kosaraju-edge-3`, `SCC 1`, `Finish stack`, `Call stack`, `data-scene-height=`, `data-min-width=`} {
				if !strings.Contains(got, want) {
					t.Fatalf("missing %s", want)
				}
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
}

func TestKosarajuFixedGeometryAndReverseWitness(t *testing.T) {
	nodes := []GraphNode{{ID: "A", X: -1000000, Y: 0}, {ID: "B", X: 1000000, Y: 0}}
	p := kosarajuPositions(nodes, 24, 64, 592, 348)
	if p[0][1] != p[1][1] || p[0][0] >= p[1][0] {
		t.Fatal(p)
	}
	step := trace.Step{LineID: "L7", State: KosarajuLessonState{GraphState: GraphState{Nodes: nodes, Edges: []GraphEdge{{From: "B", To: "A"}}}, Pass: 2, Transposed: true}}
	got := RenderKosarajuLessonSVG(step, 960, 640)
	if !strings.Contains(got, "Previous direction; arrows now reversed") || !strings.Contains(got, "Transpose graph") {
		t.Fatal("missing transpose witness")
	}
}
