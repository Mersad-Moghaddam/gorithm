package renderers_test

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/binarysearch"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestBinarySearchLessonSVG(t *testing.T) {
	for _, a := range [][]int{nil, {2}, {2, 2, 2}, {3, 1, 2}, {-1000000, 0, 1000000}} {
		for _, step := range binarysearch.Trace(a, 4) {
			before := step.State.(renderers.BinarySearchLessonState)
			copyState := before
			copyState.Values = append([]int{}, before.Values...)
			copyState.Comparing = append([]int(nil), before.Comparing...)
			copyState.Sorted = append([]int(nil), before.Sorted...)
			for _, size := range [][2]int{{320, 420}, {800, 480}, {0, -1}} {
				svg := renderers.RenderBinarySearchLessonSVG(step, size[0], size[1])
				if svg != renderers.RenderBinarySearchLessonSVG(step, size[0], size[1]) {
					t.Fatal("nondeterministic")
				}
				if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, `width="-`) || strings.Contains(svg, `height="-`) {
					t.Fatal("bad geometry")
				}
				d := xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">` + svg + `</svg>`))
				ids := map[string]bool{}
				for {
					tok, err := d.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if e, ok := tok.(xml.StartElement); ok {
						for _, attr := range e.Attr {
							if attr.Name.Local == "id" || attr.Name.Local == "data-entity" {
								key := attr.Name.Local + attr.Value
								if ids[key] {
									t.Fatalf("duplicate %s", key)
								}
								ids[key] = true
							}
						}
					}
				}
				if size[0] > 0 && before.Low > before.High && (strings.Contains(svg, `id="binarysearch-active-range"`) || !strings.Contains(svg, "empty interval")) {
					t.Fatal("crossed bracket")
				}
				if size[0] > 0 && !before.SortedInput && !strings.Contains(svg, "sorted input required") {
					t.Fatal("missing warning")
				}
			}
			if !reflect.DeepEqual(copyState, step.State) {
				t.Fatal("renderer mutated snapshot")
			}
		}
	}
	for _, step := range []trace.Step{{}, {State: 42}} {
		if !strings.Contains(renderers.RenderBinarySearchLessonSVG(step, 320, 420), "No binary search state") {
			t.Fatal("missing fallback")
		}
	}
}

func TestBinarySearchLessonSVGPointerMotionAndEscaping(t *testing.T) {
	steps := binarysearch.Trace([]int{3, 8, 14, 22, 35, 47, 59, 71, 88}, 47)
	var before, after string
	for _, step := range steps {
		if step.LineID == "L7" && before == "" {
			before = renderers.RenderBinarySearchLessonSVG(step, 320, 420)
		}
		if step.LineID == "L8" && after == "" {
			after = renderers.RenderBinarySearchLessonSVG(step, 320, 420)
		}
	}
	if !strings.Contains(before, `data-entity="binarysearch-low" data-x="64"`) || !strings.Contains(after, `data-entity="binarysearch-low" data-x="444"`) {
		t.Fatal("low marker lacks stable moving identity")
	}
	if !strings.Contains(after, "excluded") {
		t.Fatal("discard cue missing")
	}
	s := steps[0].State.(renderers.BinarySearchLessonState)
	s.Phase = `<script a="x">&`
	svg := renderers.RenderBinarySearchLessonSVG(trace.Step{State: s}, 800, 480)
	if strings.Contains(svg, "<script") || !strings.Contains(svg, "&lt;script") {
		t.Fatal("unescaped phase")
	}
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	for {
		_, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestBinarySearchLessonSVGCompareAndFound(t *testing.T) {
	for _, step := range binarysearch.Trace([]int{2, 2, 2}, 2) {
		svg := renderers.RenderBinarySearchLessonSVG(step, 320, 420)
		if !strings.Contains(svg, `data-min-width="800"`) || !strings.Contains(svg, `data-entity="binarysearch-item-0"`) || !strings.Contains(svg, `fill="#c45d75"`) {
			t.Fatal("identity/readability contract")
		}
		if step.LineID == "L5" && (!strings.Contains(svg, "compare") || !strings.Contains(svg, "#ffe29a")) {
			t.Fatal("compare cue")
		}
		if step.LineID == "L6" && (!strings.Contains(svg, "found index 1") || !strings.Contains(svg, "#9bf0bb")) {
			t.Fatal("found cue")
		}
	}
}

func TestBinarySearchMarkerLabelsStaySeparatedAndInsideScene(t *testing.T) {
	for _, step := range binarysearch.Trace([]int{2}, 1) {
		svg := renderers.RenderBinarySearchLessonSVG(step, 320, 420)
		if step.LineID == "L5" && !strings.Contains(svg, `y="49" text-anchor="middle" font-size="12">compare`) {
			t.Fatal("comparison cue overlaps midpoint marker")
		}
		if step.LineID == "L13" && !strings.Contains(svg, `data-entity="binarysearch-high" data-x="32"`) {
			t.Fatal("negative high marker clips at scene edge")
		}
	}
}
