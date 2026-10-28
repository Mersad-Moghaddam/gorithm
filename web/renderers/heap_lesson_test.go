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

func TestHeapLessonValueContrast(t *testing.T) {
	luminance := func(color string) float64 {
		rgb, err := strconv.ParseUint(strings.TrimPrefix(color, "#"), 16, 24)
		if err != nil || len(color) != 7 {
			t.Fatalf("invalid color %q", color)
		}
		var sum float64
		for i, weight := range []float64{0.2126, 0.7152, 0.0722} {
			channel := float64((rgb>>uint(16-8*i))&255) / 255
			if channel <= 0.04045 {
				channel /= 12.92
			} else {
				channel = math.Pow((channel+0.055)/1.055, 2.4)
			}
			sum += weight * channel
		}
		return sum
	}
	state := HeapLessonState{ArrayState: ArrayState{Values: []int{1, 2, 3, 4, 5}}, IDs: []int{0, 1, 2, 3, 4}, HeapSize: 5}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	decoder := xml.NewDecoder(strings.NewReader("<svg>" + RenderHeapLessonSVG(trace.Step{State: state}, 320, 360) + "</svg>"))
	entity, background := "", ""
	checked := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		switch element := token.(type) {
		case xml.StartElement:
			attributes := map[string]string{}
			for _, attr := range element.Attr {
				attributes[attr.Name.Local] = attr.Value
			}
			if key := attributes["data-entity"]; key != "" {
				entity = key
			}
			if entity == "" {
				continue
			}
			if element.Name.Local == "rect" {
				background = attributes["fill"]
				id, err := strconv.Atoi(strings.Split(entity, ":")[1])
				if err != nil {
					t.Fatal(err)
				}
				if background != palette[id] {
					t.Fatalf("identity fill changed: %s %s", entity, background)
				}
			}
			if element.Name.Local == "text" && attributes["y"] == "5" {
				fg, bg := luminance(attributes["fill"]), luminance(background)
				ratio := (math.Max(fg, bg) + 0.05) / (math.Min(fg, bg) + 0.05)
				t.Logf("%s text=%s fill=%s contrast=%.3f", entity, attributes["fill"], background, ratio)
				if ratio < 4.5 {
					t.Errorf("%s value contrast %.3f < 4.5", entity, ratio)
				}
				checked++
			}
		case xml.EndElement:
			if element.Name.Local == "g" {
				entity = ""
			}
		}
	}
	if checked != 10 {
		t.Fatalf("checked %d value labels, want all five in both projections", checked)
	}
}

func TestHeapLessonSVG(t *testing.T) {
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		for _, values := range [][]int{nil, {7}, {2, 1}, {-1, -3, -2}, {0, 0, 0}, make([]int, 64)} {
			ids := make([]int, len(values))
			for i := range ids {
				ids[i] = i
			}
			state := HeapLessonState{ArrayState: ArrayState{Values: values, Current: -1}, IDs: ids, HeapSize: len(values), Root: -1, LeftChild: -1, RightChild: -1, Largest: -1, ExtractIndex: -1, Building: true}
			svg := RenderHeapLessonSVG(trace.Step{State: state}, size[0], size[1])
			if !strings.Contains(svg, `data-lesson="heap"`) {
				t.Fatal("missing root")
			}
			decoder := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
			entities := map[string]bool{}
			edges := 0
			for {
				tok, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				start, ok := tok.(xml.StartElement)
				if !ok {
					continue
				}
				for _, a := range start.Attr {
					if a.Name.Local == "data-entity" {
						if entities[a.Value] {
							t.Fatal("duplicate entity", a.Value)
						}
						entities[a.Value] = true
					}
					if a.Name.Local == "data-child" {
						edges++
						child, err := strconv.Atoi(a.Value)
						if err != nil || child >= state.HeapSize {
							t.Fatal("invalid edge")
						}
					}
					if a.Name.Local == "width" || a.Name.Local == "height" || a.Name.Local == "data-x" || a.Name.Local == "data-y" {
						f, err := strconv.ParseFloat(a.Value, 64)
						if err != nil || f < 0 || f != f {
							t.Fatalf("bad geometry %s", a.Value)
						}
					}
				}
			}
			if len(entities) != 2*len(values) || edges != max(0, len(values)-1) {
				t.Fatalf("entities=%d edges=%d n=%d", len(entities), edges, len(values))
			}
			if len(values) == 0 && !strings.Contains(svg, "Empty array") {
				t.Fatal("missing empty label")
			}
		}
	}
}

func TestHeapLessonSwapGeometry(t *testing.T) {
	s := HeapLessonState{ArrayState: ArrayState{Values: []int{3, 2, 1}, Swapped: []int{0, 1}, Current: 1}, IDs: []int{1, 0, 2}, HeapSize: 3, Building: true}
	svg := RenderHeapLessonSVG(trace.Step{State: s, Action: trace.ActionSwap}, 320, 360)
	if strings.Count(svg, `data-motion="swap-`) != 4 {
		t.Fatal("both projections need swap geometry", svg)
	}
	for _, want := range []string{`data-entity="tree:1"`, `data-entity="array:1"`, `#387faa`, `#0`, `data-min-width=`} {
		if !strings.Contains(svg, want) {
			t.Fatal("missing", want)
		}
	}
}

func TestHeapLessonExtractedSuffix(t *testing.T) {
	s := HeapLessonState{ArrayState: ArrayState{Values: []int{1, 2, 3}, Sorted: []int{2}, Swapped: []int{0, 2}}, IDs: []int{2, 0, 1}, HeapSize: 2, ExtractIndex: 2}
	svg := RenderHeapLessonSVG(trace.Step{State: s, Action: trace.ActionSwap}, 320, 360)
	if strings.Contains(svg, `data-entity="tree:1"`) || strings.Contains(svg, `data-child="2"`) {
		t.Fatal("extracted identity or edge remains in active tree")
	}
	for _, want := range []string{`data-entity="array:1"`, `✓ final`, `data-phase="extract"`} {
		if !strings.Contains(svg, want) {
			t.Fatal("missing extraction cue", want)
		}
	}
}

func TestHeapLessonRootLabelClearsSubtitle(t *testing.T) {
	_, y := heapLessonPoint(0, 800)
	if candidateBaseline := y - 31; candidateBaseline < 84 {
		t.Fatalf("root candidate baseline %.0f too close to subtitle baseline 54", candidateBaseline)
	}
}
