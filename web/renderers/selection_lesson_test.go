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

func TestSelectionLessonSVG(t *testing.T) {
	s := SelectionLessonState{ArrayState: ArrayState{Values: []int{1, 3, 2}, Swapped: []int{0, 2}}, IDs: []int{2, 1, 0}, Pass: 0, Scan: 2, MinIndex: 2, MinValue: 0, MaxValue: 3}
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		svg := RenderSelectionLessonSVG(trace.Step{State: s, Action: trace.ActionSwap, LineID: "L10"}, size[0], size[1])
		for _, text := range []string{`data-lesson="selection"`, `data-motion="swap-up"`, `data-motion="swap-down"`, `main:0`, `#c45d75`, `minimum`, `i = 0`, `j = 2`, `font-size="14"`} {
			if !strings.Contains(svg, text) {
				t.Fatalf("missing %s", text)
			}
		}
		dec := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		ids := map[string]bool{}
		for {
			tok, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if e, ok := tok.(xml.StartElement); ok {
				for _, a := range e.Attr {
					switch a.Name.Local {
					case "data-entity":
						if ids[a.Value] {
							t.Fatal("duplicate identity")
						}
						ids[a.Value] = true
					case "width", "height", "data-x", "data-y", "data-scene-width", "data-scene-height":
						v, err := strconv.ParseFloat(a.Value, 64)
						if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
							t.Fatalf("bad geometry %v", a)
						}
					}
				}
			}
		}
	}
	noOp := RenderSelectionLessonSVG(trace.Step{State: s, Action: trace.ActionSet, LineID: "L10"}, 320, 360)
	if strings.Contains(noOp, `data-motion=`) || !strings.Contains(noOp, "Already minimum") {
		t.Fatal("no-op swap motion")
	}
	if !strings.Contains(RenderSelectionLessonSVG(trace.Step{State: SelectionLessonState{}}, 320, 360), "Empty array") {
		t.Fatal("missing empty state")
	}
	signed := SelectionLessonState{ArrayState: ArrayState{Values: []int{-1, 0, -2}}, IDs: []int{0, 1, 2}, MinValue: -2, MaxValue: 0, Pass: -1, Scan: -1, MinIndex: -1}
	svg := RenderSelectionLessonSVG(trace.Step{State: signed}, 320, 360)
	if !strings.Contains(svg, ">-2</text>") || !strings.Contains(svg, `height="2.00"`) {
		t.Fatal("signed/zero bars lost")
	}
}

func TestSelectionNegativeFinalCheckClearsSlotIndex(t *testing.T) {
	s := SelectionLessonState{ArrayState: ArrayState{Values: []int{-1000000}}, IDs: []int{0}, SortedPrefix: 1, MinValue: -1000000, Pass: -1, Scan: -1, MinIndex: -1}
	svg := RenderSelectionLessonSVG(trace.Step{State: s, Action: trace.ActionDone}, 320, 360)
	if !strings.Contains(svg, `y="116.00" text-anchor="middle" font-size="16" fill="#9bf0bb">✓`) {
		t.Fatalf("final check must clear slot index at y=308: %s", svg)
	}
}
