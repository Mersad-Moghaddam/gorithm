package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestLCSLessonSVG(t *testing.T) {
	s := LCSLessonState{S1: "<>&\"é", S2: "<", DP: [][]int{{0, 0}, {0, 1}}, Filled: [][]bool{{true, true}, {true, true}}, Phase: "match", Row: 1, Col: 1, From: [2]int{1, 1}, To: [2]int{0, 0}, Matches: [][2]int{{0, 0}}, Sequence: "<"}
	for _, w := range []int{0, 1, 320, 390, 768, 1440} {
		out := RenderLCSLessonSVG(trace.Step{State: s}, w, 0)
		for _, want := range []string{`data-lesson="lcs"`, `data-scene-width=`, `data-min-width=`, `data-entity="s1:0"`, `data-entity="cursor"`, `data-entity="output:0:0"`, `data-from-entity="s1:0"`, `0xC3`, `&lt;`} {
			if !strings.Contains(out, want) {
				t.Fatalf("missing %s", want)
			}
		}
		dec := xml.NewDecoder(strings.NewReader("<svg>" + out + "</svg>"))
		ids := map[string]bool{}
		for {
			token, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if e, ok := token.(xml.StartElement); ok {
				if e.Name.Local == "script" {
					t.Fatal("injection")
				}
				for _, a := range e.Attr {
					if a.Name.Local == "data-entity" {
						if ids[a.Value] {
							t.Fatal("duplicate", a.Value)
						}
						ids[a.Value] = true
					}
				}
			}
		}
		if strings.Contains(out, "NaN") || strings.Contains(out, "Inf") {
			t.Fatal("nonfinite")
		}
	}
}

func TestLCSBacktrackDoesNotShowFillFormula(t *testing.T) {
	s := LCSLessonState{S1: "A", S2: "A", DP: [][]int{{0, 0}, {0, 1}}, Filled: [][]bool{{true, true}, {true, true}}, Phase: "match", From: [2]int{1, 1}, To: [2]int{0, 0}, Match: true}
	out := RenderLCSLessonSVG(trace.Step{State: s}, 320, 450)
	if strings.Contains(out, "diagonal 0 + 1 = 1") {
		t.Fatal("backtracking presents a table-fill formula as current operation")
	}
}

func TestLCSEmptySceneFitsBoundaryCaption(t *testing.T) {
	out := RenderLCSLessonSVG(trace.Step{State: LCSLessonState{Phase: "allocate"}}, 320, 450)
	if !strings.Contains(out, `data-min-width="400"`) {
		t.Fatal("boundary caption needs a scrollable 400-unit scene")
	}
}
