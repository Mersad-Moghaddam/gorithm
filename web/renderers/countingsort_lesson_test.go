package renderers

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestCountingsortLessonSVG(t *testing.T) {
	s := CountingsortLessonState{Input: []int{0, 2, 0}, InputIDs: []int{0, 1, 2}, Counts: []int{2, 2, 3}, Output: []int{0, 0, 0}, OutputIDs: []int{-1, 2, -1}, OutputFilled: []bool{false, true, false}, Phase: "output", Scan: 2, Target: 1, CountIndex: 0}
	for _, width := range []int{320, 960} {
		height := 360
		if width == 960 {
			height = 540
		}
		svg := RenderCountingsortLessonSVG(trace.Step{State: s, LineID: "L21"}, width, height)
		for _, text := range []string{`data-lesson="countingsort"`, `data-entity="input:2"`, `data-entity="output:2"`, `data-from-entity="input:2"`, "Remaining insertion boundary", "empty", "#8763bb"} {
			if !strings.Contains(svg, text) {
				t.Errorf("missing %q", text)
			}
		}
		if strings.Contains(svg, `data-entity="output:-1"`) || strings.Contains(svg, "swap-") {
			t.Fatal("phantom item or fake swap")
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		ids := map[string]bool{}
		for {
			tok, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if node, ok := tok.(xml.StartElement); ok {
				for _, a := range node.Attr {
					switch a.Name.Local {
					case "data-entity":
						if ids[a.Value] {
							t.Fatal("duplicate identity", a.Value)
						}
						ids[a.Value] = true
					case "width", "height", "data-x", "data-y", "data-scene-width", "data-scene-height":
						v, e := strconv.ParseFloat(a.Value, 64)
						if e != nil || v < 0 || v != v || v > 32768 {
							t.Fatal("invalid geometry", a)
						}
					}
				}
			}
		}
	}
	if svg := RenderCountingsortLessonSVG(trace.Step{State: CountingsortLessonState{}}, 320, 360); !strings.Contains(svg, "Empty array") {
		t.Fatal(svg)
	}
	// Sparse ranges retain every counter, including zero-frequency indices.
	s.Counts = make([]int, 1001)
	s.Counts[1000] = 1
	if svg := RenderCountingsortLessonSVG(trace.Step{State: s}, 320, 360); !strings.Contains(svg, "[1000]") {
		t.Fatal("omitted sparse counter")
	}
}
