package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestHanoiTowerLessonSVG(t *testing.T) {
	s := HanoiTowerLessonState{HanoiState: HanoiState{Pegs: [3][]int{{3, 2}, nil, {1}}}, TotalDisks: 3, MoveCount: 1, Phase: "move", Disk: 1, From: 0, To: 2, Calls: []HanoiTowerCall{{1, 0, 2, 1}}}
	for _, size := range [][2]int{{0, 0}, {1, 1}, {320, 450}, {390, 450}, {768, 450}, {1440, 450}} {
		out := RenderHanoiTowerLessonSVG(trace.Step{State: s}, size[0], size[1])
		for _, want := range []string{`data-lesson="hanoitower"`, `data-phase="move"`, `data-entity="disk:1"`, `data-motion="lift"`, `data-clear-y=`, `Move 1`, `from`, `aux`, `to`, `Pending`} {
			if !strings.Contains(out, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(out, "NaN") || strings.Contains(out, "Inf") {
			t.Fatal("nonfinite geometry")
		}
		dec := xml.NewDecoder(strings.NewReader("<svg>" + out + "</svg>"))
		seen := map[string]bool{}
		for {
			token, err := dec.Token()
			if err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
				break
			}
			if el, ok := token.(xml.StartElement); ok {
				attrs := map[string]string{}
				for _, a := range el.Attr {
					attrs[a.Name.Local] = a.Value
				}
				if id := attrs["data-entity"]; id != "" {
					if seen[id] {
						t.Fatal("duplicate identity")
					}
					seen[id] = true
					for _, key := range []string{"data-x", "data-y"} {
						if _, err := strconv.ParseFloat(attrs[key], 64); err != nil {
							t.Fatal(err)
						}
					}
					if id == "disk:1" {
						clear, _ := strconv.ParseFloat(attrs["data-clear-y"], 64)
						y, _ := strconv.ParseFloat(attrs["data-y"], 64)
						if clear+22 >= 300-3*24-16 || clear >= y {
							t.Fatal("lift does not clear tallest stack")
						}
					}
				}
				if el.Name.Local == "rect" {
					for _, key := range []string{"width", "height"} {
						v, err := strconv.ParseFloat(attrs[key], 64)
						if err != nil || v <= 0 {
							t.Fatal("invalid rectangle")
						}
					}
				}
			}
		}
		if len(seen) != 3 {
			t.Fatal("missing disks")
		}
	}
	s.Phase = `<bad>&"`
	out := RenderHanoiTowerLessonSVG(trace.Step{State: s}, 320, 450)
	if strings.Contains(out, `<bad>`) {
		t.Fatal("unescaped phase")
	}
}
