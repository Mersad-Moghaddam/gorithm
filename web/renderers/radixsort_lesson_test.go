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

func TestRadixsortLessonSVG(t *testing.T) {
	state := RadixsortLessonState{ArrayState: ArrayState{Values: []int{0, 12, 12}, Current: 1}, IDs: []int{0, 1, 2}, Output: []int{0, 0, 12}, OutputIDs: []int{-1, -1, 2}, OutputFilled: []bool{false, false, true}, Exp: 10, MaxValue: 12, Scan: 1, Digit: 1, Target: 2, CopyIndex: -1, Phase: "output"}
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		svg := RenderRadixsortLessonSVG(trace.Step{State: state}, size[0], size[1])
		for _, want := range []string{`data-lesson="radixsort"`, `data-entity="main:0:slot:0"`, `data-entity="output:2"`, `data-from-entity="main:2:slot:2"`, `data-motion="linear"`, `data-min-width=`, `exp = 10`, `unwritten`, `#8763bb`, `text-decoration="underline"`} {
			if !strings.Contains(svg, want) {
				t.Fatalf("missing %s", want)
			}
		}
		if strings.Contains(svg, `data-entity="output:-1"`) || strings.Contains(svg, "swap-up") {
			t.Fatal(svg)
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		entities := map[string]bool{}
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if start, ok := token.(xml.StartElement); ok {
				for _, a := range start.Attr {
					if a.Name.Local == "data-entity" {
						if entities[a.Value] {
							t.Fatal("duplicate entity", a.Value)
						}
						entities[a.Value] = true
					}
					switch a.Name.Local {
					case "x", "y", "width", "height", "data-x", "data-y", "data-scene-width", "data-scene-height":
						v, e := strconv.ParseFloat(a.Value, 64)
						if e != nil || math.IsNaN(v) || math.IsInf(v, 0) || (a.Name.Local == "width" || a.Name.Local == "height") && v < 0 {
							t.Fatalf("bad geometry %v", a)
						}
					}
				}
			}
		}
	}
	if svg := RenderRadixsortLessonSVG(trace.Step{State: RadixsortLessonState{}}, 320, 360); !strings.Contains(svg, "Empty array") {
		t.Fatal(svg)
	}
	state.Phase = `<script>`
	if strings.Contains(RenderRadixsortLessonSVG(trace.Step{State: state}, 320, 360), `data-phase="<script>`) {
		t.Fatal("phase unescaped")
	}
}
