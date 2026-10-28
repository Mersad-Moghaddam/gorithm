package renderers_test

import (
	"encoding/xml"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/segmenttree"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestSegmentTreeLessonSVG(t *testing.T) {
	for _, size := range [][2]int{{320, 420}, {800, 480}, {0, -1}} {
		for _, step := range segmenttree.Trace(nil, 0, 0, 0, 0) {
			svg := renderers.RenderSegmentTreeLessonSVG(step, size[0], size[1])
			if svg != renderers.RenderSegmentTreeLessonSVG(step, size[0], size[1]) {
				t.Fatal("nondeterministic")
			}
			if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
				t.Fatal("nonfinite")
			}
			if !strings.Contains(svg, `data-lesson="segmenttree"`) || !strings.Contains(svg, `data-min-width="800"`) {
				t.Fatal("missing scene contract")
			}
			ids := map[string]bool{}
			entities := map[string]bool{}
			edges := []string{}
			d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
			for {
				tok, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				el, ok := tok.(xml.StartElement)
				if !ok {
					continue
				}
				for _, a := range el.Attr {
					switch a.Name.Local {
					case "id":
						if ids[a.Value] {
							t.Fatal("duplicate id")
						}
						ids[a.Value] = true
					case "data-entity":
						if entities[a.Value] {
							t.Fatal("duplicate entity")
						}
						entities[a.Value] = true
					case "data-child", "data-parent":
						edges = append(edges, a.Value)
					case "width", "height":
						n, err := strconv.ParseFloat(a.Value, 64)
						if err != nil || n < 0 {
							t.Fatal("bad dimension")
						}
					}
				}
			}
			for _, id := range edges {
				if !ids["segmenttree-"+id] {
					t.Fatal("dangling edge")
				}
			}
			if step.LineID == "L1" && !strings.Contains(svg, "?") {
				t.Fatal("missing unknown")
			}
			if step.LineID == "L15" && !strings.Contains(svg, "checked") {
				t.Fatal("missing checked")
			}
			if step.LineID == "L16" && !strings.Contains(svg, "+0") {
				t.Fatal("missing disjoint")
			}
			if step.LineID == "L17" && !strings.Contains(svg, "split") {
				t.Fatal("missing split")
			}
			if step.LineID == "L13" && (!strings.Contains(svg, "Accepted cover") || !strings.Contains(svg, "24")) {
				t.Fatal("missing cover")
			}
		}
	}
}

func TestSegmentTreeLessonFallbackAndEscaping(t *testing.T) {
	for _, step := range []trace.Step{{}, {State: "wrong"}} {
		if !strings.Contains(renderers.RenderSegmentTreeLessonSVG(step, 0, 0), "No segment tree state") {
			t.Fatal("missing fallback")
		}
	}
	s := renderers.SegmentTreeLessonState{Phase: `query<>&"`, Values: []int{0}, Nodes: []renderers.SegmentTreeLessonNode{{ID: `node-<>&"`, Built: true}}}
	copyState := s
	copyState.Nodes = append([]renderers.SegmentTreeLessonNode(nil), s.Nodes...)
	svg := renderers.RenderSegmentTreeLessonSVG(trace.Step{State: s}, 320, 420)
	if strings.Contains(svg, "query<") || !strings.Contains(svg, "&lt;") {
		t.Fatal("unescaped")
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
	if !reflect.DeepEqual(s, copyState) {
		t.Fatal("mutated state")
	}
}
