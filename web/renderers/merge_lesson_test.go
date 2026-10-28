package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"
)

func TestMergeLessonSVG(t *testing.T) {
	s := MergeLessonState{ArrayState: ArrayState{Values: []int{2, 0, -3}, Current: 1}, IDs: []int{0, 1, 2}, Aux: []int{0, 0}, AuxIDs: []int{1, -1}, AuxFilled: []bool{true, false}, Stack: []MergeLessonRange{{0, 1, 2}}, Left: 0, Mid: 1, Right: 2, I: 0, J: 2, K: 0, CopyIndex: -1, MinValue: -3, MaxValue: 2}
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		svg := RenderMergeLessonSVG(trace.Step{LineID: "L15", State: s}, size[0], size[1])
		for _, want := range []string{`data-lesson="merge"`, `data-phase="to-aux"`, `data-entity="aux:0:1"`, `data-from-entity="main:1:1"`, `empty`, `#c45d75`, `#387faa`, `#8763bb`, `i=0`, `j=2`, `k=0`, `[0..2]`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		seen := map[string]bool{}
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
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
						if seen[a.Value] {
							t.Fatal("duplicate key", a.Value)
						}
						seen[a.Value] = true
					}
					switch a.Name.Local {
					case "data-x", "data-y", "width", "height", "data-scene-width", "data-scene-height":
						v, err := strconv.ParseFloat(a.Value, 64)
						if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
							t.Fatal(a)
						}
					}
				}
			}
		}
	}
	if !strings.Contains(RenderMergeLessonSVG(trace.Step{State: MergeLessonState{}}, 320, 360), "Empty array") {
		t.Fatal("empty")
	}
	s.Values = []int{0, 0, -3}
	s.IDs = []int{1, 1, 2}
	s.CopyIndex = 0
	copyBack := RenderMergeLessonSVG(trace.Step{LineID: "L23", State: s}, 320, 360)
	for _, want := range []string{`data-entity="main:0:1"`, `data-entity="main:1:1"`, `data-from-entity="aux:0:1"`, `data-phase="copy-back"`} {
		if !strings.Contains(copyBack, want) {
			t.Fatal("copy-back representation", want)
		}
	}
	if strings.Contains(copyBack, `swap-up`) || strings.Contains(copyBack, `swap-down`) {
		t.Fatal("copy treated as swap")
	}
}

func TestMergeLessonFinalNegativeCheckDoesNotOverlapIndex(t *testing.T) {
	s := MergeLessonState{ArrayState: ArrayState{Values: []int{-3}}, IDs: []int{0}, MinValue: -3}
	svg := RenderMergeLessonSVG(trace.Step{LineID: "L9", Action: trace.ActionDone, State: s}, 320, 360)
	if !strings.Contains(svg, `[0] ✓</text>`) {
		t.Fatal("final check must share the index label, not overlap it")
	}
}
