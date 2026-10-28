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

func TestBubbleLessonSVG(t *testing.T) {
	state := BubbleLessonState{ArrayState: ArrayState{Values: []int{0, -3, 2}, Swapped: []int{0, 1}, Sorted: []int{2}}, IDs: []int{1, 0, 2}, Pass: 0, Scan: 0, UnsortedEnd: 1, MinValue: -3, MaxValue: 2}
	for _, size := range [][2]int{{320, 360}, {960, 540}, {0, -1}} {
		svg := RenderBubbleLessonSVG(trace.Step{Action: trace.ActionSwap, State: state}, size[0], size[1])
		for _, want := range []string{`data-lesson="bubble"`, `data-phase="swap"`, `data-motion="swap-up"`, `data-motion="swap-down"`, `data-entity="main:0"`, `#c45d75`, `#387faa`, `#8763bb`, `#9bf0bb`, `Pass 1`, `Sorted suffix`, `Zero`, `#1`, `✓`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		entities := map[string]bool{}
		sceneWidth, minimumWidth := "", ""
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if start, ok := token.(xml.StartElement); ok {
				for _, attr := range start.Attr {
					if attr.Name.Local == "data-scene-width" {
						sceneWidth = attr.Value
					}
					if attr.Name.Local == "data-min-width" {
						minimumWidth = attr.Value
					}
					if attr.Name.Local == "data-entity" {
						if entities[attr.Value] {
							t.Fatal("duplicate entity")
						}
						entities[attr.Value] = true
					}
					switch attr.Name.Local {
					case "x", "y", "width", "height", "data-x", "data-y", "data-scene-width", "data-scene-height", "data-min-width":
						v, err := strconv.ParseFloat(attr.Value, 64)
						if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
							t.Fatalf("invalid %s=%s", attr.Name.Local, attr.Value)
						}
						if (strings.Contains(attr.Name.Local, "width") || strings.Contains(attr.Name.Local, "height")) && v < 0 {
							t.Fatal("negative dimension")
						}
					}
				}
			}
		}
		if len(entities) != 3 {
			t.Fatal("missing entities")
		}
		if minimumWidth != sceneWidth {
			t.Fatal("scene must retain 1:1 text size when viewport narrows")
		}
	}
	compare := state
	compare.Comparing = []int{0, 1}
	compare.Swapped = nil
	svg := RenderBubbleLessonSVG(trace.Step{Action: trace.ActionCompare, State: compare}, 320, 360)
	if !strings.Contains(svg, `data-emphasis="compare"`) || !strings.Contains(svg, "#ffe29a") || strings.Contains(svg, `data-motion="swap-`) {
		t.Fatal("comparison geometry")
	}
	empty := RenderBubbleLessonSVG(trace.Step{State: BubbleLessonState{ArrayState: ArrayState{Sorted: []int{0}}}}, 320, 360)
	if !strings.Contains(empty, "Empty array") {
		t.Fatal("empty state")
	}
	for _, values := range [][]int{{7}, {0, 0}, {-3, -2}, {2, 2}, {-1000000, 0, 1000000}} {
		s := BubbleLessonState{ArrayState: ArrayState{Values: values}, IDs: make([]int, len(values)), Scan: -1, Pass: -1, UnsortedEnd: len(values) - 1}
		for i, value := range values {
			s.IDs[i] = i
			s.MinValue = min(s.MinValue, value)
			s.MaxValue = max(s.MaxValue, value)
		}
		if s.MinValue == s.MaxValue {
			s.MaxValue++
		}
		fragment := RenderBubbleLessonSVG(trace.Step{State: s}, 320, 360)
		if strings.Contains(fragment, "NaN") || strings.Contains(fragment, "Inf") || strings.Contains(fragment, `data-motion="swap-`) {
			t.Fatalf("invalid edge geometry for %v", values)
		}
	}
}
