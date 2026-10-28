package renderers_test

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"math"
	"strconv"
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

func TestBSTLongValuesHaveRoom(t *testing.T) {
	steps := bst.Trace([]int{0, -1000000, 1000000, -999999, 999999}, []bst.Operation{{Type: "delete", Value: 0}})
	for _, step := range steps {
		s := step.State.(renderers.BSTLessonState)
		svg := renderers.RenderBSTLessonSVG(step, 320, 420)
		d := xml.NewDecoder(strings.NewReader(`<svg>` + svg + `</svg>`))
		var positions [][2]float64
		var radius float64
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
			attrs := map[string]string{}
			for _, a := range el.Attr {
				attrs[a.Name.Local] = a.Value
			}
			if strings.HasPrefix(attrs["data-entity"], "bst:node-") {
				x, _ := strconv.ParseFloat(attrs["data-x"], 64)
				y, _ := strconv.ParseFloat(attrs["data-y"], 64)
				positions = append(positions, [2]float64{x, y})
			}
			if el.Name.Local == "ellipse" {
				radius, _ = strconv.ParseFloat(attrs["rx"], 64)
			}
		}
		if radius < 46 {
			t.Fatalf("node background too narrow for long values: radius %v", radius)
		}
		for i, p := range positions {
			for _, q := range positions[:i] {
				if p[1] == q[1] && math.Abs(p[0]-q[0]) < 2*radius+16 {
					t.Fatalf("node labels lack horizontal space: %v %v", p, q)
				}
			}
		}
		for _, n := range s.Nodes {
			if !strings.Contains(svg, ">"+strconv.Itoa(n.Val)+"</text>") || !strings.Contains(svg, `data-entity="bst:`+n.ID+`"`) {
				t.Fatalf("lost node %s value %d", n.ID, n.Val)
			}
		}
		if s.Phase == "copy" && !strings.Contains(svg, `width="150"`) {
			t.Fatal("copy cue too narrow for full value")
		}
	}
}

func TestBSTDeleteResultNotPremature(t *testing.T) {
	steps := bst.Trace([]int{2, 1, 3}, []bst.Operation{{Type: "delete", Value: 2}})
	for _, step := range steps {
		s := step.State.(renderers.BSTLessonState)
		if s.Phase == "compare" || s.Phase == "successor" || s.Phase == "copy" || s.Phase == "descend" {
			if strings.Contains(renderers.RenderBSTLessonSVG(step, 800, 480), `>deleted</text>`) {
				t.Fatalf("premature delete result in %s", s.Phase)
			}
		}
	}
}
