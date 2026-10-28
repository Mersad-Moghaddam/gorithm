package renderers

import (
	"encoding/xml"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestQuickLessonSVG(t *testing.T) {
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		s := QuickLessonState{ArrayState: ArrayState{Values: []int{-3, 0, 2}, Swapped: []int{0, 2}, Comparing: []int{0, 2}}, IDs: []int{1, 0, 2}, Low: 0, High: 2, Boundary: -1, Scan: 0, PivotActive: true, PivotID: 2, PivotIndex: 2, PivotValue: 2}
		svg := RenderQuickLessonSVG(trace.Step{LineID: "L17", Action: trace.ActionSwap, Explanation: "<script>&", State: s}, size[0], size[1])
		for _, word := range []string{`data-lesson="quick"`, `data-entity="main:1"`, `data-motion="swap-up"`, `data-motion="swap-down"`, `#387faa`, `#ffe29a`, `&lt;script&gt;&amp;`, `pivot`, `Unscanned`, `&lt; pivot`} {
			if !strings.Contains(svg, word) {
				t.Fatal("missing", word)
			}
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		seen := map[string]bool{}
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			e, ok := token.(xml.StartElement)
			if !ok {
				continue
			}
			for _, a := range e.Attr {
				if a.Name.Local == "data-entity" {
					if seen[a.Value] {
						t.Fatal("duplicate entity")
					}
					seen[a.Value] = true
				}
				switch a.Name.Local {
				case "x", "y", "width", "height", "data-x", "data-y", "data-scene-width", "data-scene-height", "data-min-width":
					v, err := strconv.ParseFloat(a.Value, 64)
					if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
						t.Fatal(a)
					}
					if (a.Name.Local == "width" || a.Name.Local == "height") && v < 0 {
						t.Fatal(a)
					}
				}
			}
		}
		if len(seen) != 3 {
			t.Fatal(seen)
		}
	}
	empty := RenderQuickLessonSVG(trace.Step{State: QuickLessonState{}}, 320, 360)
	if !strings.Contains(empty, "Empty array") {
		t.Fatal(empty)
	}
	pending := QuickLessonState{ArrayState: ArrayState{Values: []int{3, 1, 2}}, IDs: []int{0, 1, 2}, Low: 0, High: 2, Boundary: 0, Scan: 1, PivotActive: true, PivotID: 2, PivotIndex: 2, PivotValue: 2}
	beforeSwap := RenderQuickLessonSVG(trace.Step{LineID: "L13", State: pending}, 320, 360)
	if !strings.Contains(beforeSwap, "&lt; pivot: [0..-1]") {
		t.Fatal("boundary increment must not claim values moved", beforeSwap)
	}
}
