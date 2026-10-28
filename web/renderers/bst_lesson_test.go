package renderers_test

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bst"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestBSTLessonSVG(t *testing.T) {
	initial := make([]int, 20)
	for i := range initial {
		initial[i] = i - 10
	}
	fixtures := [][]trace.Step{bst.Trace(nil, nil), bst.Trace(initial, []bst.Operation{{Type: "search", Value: 20}}), bst.Trace(nil, []bst.Operation{{Type: "search", Value: 0}})}
	for _, steps := range fixtures {
		for _, step := range steps {
			for _, size := range [][2]int{{320, 420}, {800, 480}, {0, -1}} {
				before, _ := json.Marshal(step)
				svg := renderers.RenderBSTLessonSVG(step, size[0], size[1])
				after, _ := json.Marshal(step)
				if string(before) != string(after) || svg != renderers.RenderBSTLessonSVG(step, size[0], size[1]) {
					t.Fatal("mutation or nondeterminism")
				}
				if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || !strings.Contains(svg, `data-lesson="bst"`) || !strings.Contains(svg, `data-min-width=`) {
					t.Fatal(svg)
				}
				ids := map[string]bool{}
				d := xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">` + svg + `</svg>`))
				for {
					tok, err := d.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if start, ok := tok.(xml.StartElement); ok {
						for _, a := range start.Attr {
							if a.Name.Local == "data-entity" {
								if ids[a.Value] {
									t.Fatal("duplicate entity")
								}
								ids[a.Value] = true
							}
						}
					}
				}
			}
		}
	}
	for _, s := range []trace.Step{{}, {State: "wrong"}} {
		if !strings.Contains(renderers.RenderBSTLessonSVG(s, 320, 420), "No BST state") {
			t.Fatal("missing fallback")
		}
	}
	s := renderers.BSTLessonState{TreeState: renderers.TreeState{Nodes: []renderers.TreeNode{{ID: `node-1`, Label: `<>&"' سلام`}}}, LeftIDs: []string{""}, RightIDs: []string{""}, Phase: `compare`, ActiveID: "node-1", Operation: `<search>`}
	svg := renderers.RenderBSTLessonSVG(trace.Step{State: s}, 320, 420)
	if strings.Contains(svg, `<search>`) || strings.Contains(svg, `<>&`) || !strings.Contains(svg, "compare") || !strings.Contains(svg, "#c45d75") {
		t.Fatal(svg)
	}
}
