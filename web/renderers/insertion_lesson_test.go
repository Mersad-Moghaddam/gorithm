package renderers

import (
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestInsertionLessonSVG(t *testing.T) {
	state := InsertionLessonState{ArrayState: ArrayState{Values: []int{3, 3, 2}, Current: 1}, IDs: []int{0, 0, 2}, Pass: 1, Scan: 0, KeyValue: 1, KeyID: 1, Hole: 0, PrefixEnd: 0, KeyHeld: true, MinValue: 0, MaxValue: 3}
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		svg := RenderInsertionLessonSVG(trace.Step{LineID: "L7", Action: trace.ActionSwap, State: state}, size[0], size[1])
		for _, want := range []string{`data-lesson="insertion"`, `data-phase="shift"`, `data-operation="copy"`, `data-from-entity="main:0:0"`, `data-entity="held:1"`, `Conceptual hole`, `memory: 3 #0`, `Sorted prefix`, `data-motion="linear"`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(svg, "swap-up") || strings.Contains(svg, "swap-down") {
			t.Fatal("copy rendered as exchange")
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
			start, ok := token.(xml.StartElement)
			if !ok {
				continue
			}
			for _, a := range start.Attr {
				switch a.Name.Local {
				case "data-entity":
					if entities[a.Value] {
						t.Fatal("duplicate", a.Value)
					}
					entities[a.Value] = true
				case "data-x", "data-y", "width", "height":
					n, err := strconv.ParseFloat(a.Value, 64)
					if err != nil || n != n || n < 0 || n > 32768 {
						t.Fatalf("bad geometry %s=%s", a.Name.Local, a.Value)
					}
				}
			}
		}
	}
}

func TestInsertionLessonSignedAndEmpty(t *testing.T) {
	for _, s := range []InsertionLessonState{{ArrayState: ArrayState{Values: []int{-3, 0, 2}}, IDs: []int{0, 1, 2}, MinValue: -3, MaxValue: 2, Hole: -1, Scan: -1, PrefixEnd: -1}, {ArrayState: ArrayState{}, Hole: -1, Scan: -1, PrefixEnd: -1}} {
		svg := RenderInsertionLessonSVG(trace.Step{State: s}, 320, 360)
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
			t.Fatal(svg)
		}
		if len(s.Values) == 0 && !strings.Contains(svg, "Empty array") {
			t.Fatal(svg)
		}
	}
}
