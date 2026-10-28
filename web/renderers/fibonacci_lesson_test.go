package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"regexp"
	"strings"
	"testing"
)

func TestFibonacciLessonSVG(t *testing.T) {
	s := FibonacciLessonState{N: 5, Index: 5, Values: []int{0, 1, 1, 2, 3, 5}, Known: []bool{true, true, true, true, true, true}, Sources: [2]int{3, 4}, Phase: "write"}
	for _, width := range []int{0, 1, 320, 390, 768, 1440} {
		step := trace.Step{State: s}
		svg := RenderFibonacciLessonSVG(step, width, 100)
		for _, want := range []string{`data-lesson="fibonacci"`, `data-phase="write"`, `data-entity="cell:3"`, `data-from-entity="cell:3"`, `2 + 3 = 5`, `i-2`, `i-1`, `data-scene-width=`, `data-min-width=`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, `stroke="#9bf0bb"`) {
			t.Fatal("invalid geometry or premature success")
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
		RenderFibonacciLessonSVG(trace.Step{State: FibonacciLessonState{Phase: "allocate"}}, width, 0)
		if svg != RenderFibonacciLessonSVG(step, width, 100) {
			t.Fatal("render depends on history")
		}
	}
	s.Phase = `<script>&"`
	svg := RenderFibonacciLessonSVG(trace.Step{State: s}, 320, 450)
	if strings.Contains(svg, "<script>") || !strings.Contains(svg, "&lt;script&gt;") {
		t.Fatal("unescaped phase")
	}
	s.Phase = "done"
	s.HasResult = true
	s.Result = 5
	if !strings.Contains(RenderFibonacciLessonSVG(trace.Step{State: s}, 320, 450), `stroke="#9bf0bb"`) {
		t.Fatal("missing final check")
	}
}

func TestFibonacciCellsStayPutWhenValuesGrow(t *testing.T) {
	s := FibonacciLessonState{N: 46, Values: make([]int, 47), Known: make([]bool, 47), Phase: "allocate"}
	position := regexp.MustCompile(`data-entity="cell:1" data-x="\d+" data-y="\d+"`)
	for _, width := range []int{320, 1440} {
		before := position.FindString(RenderFibonacciLessonSVG(trace.Step{State: s}, width, 0))
		s.Values[46], s.Known[46] = 1836311903, true
		after := position.FindString(RenderFibonacciLessonSVG(trace.Step{State: s}, width, 0))
		if before == "" || before != after {
			t.Errorf("width %d: cell moved from %q to %q", width, before, after)
		}
		s.Values[46], s.Known[46] = 0, false
	}
}

func TestFibonacciFinalLabelFitsNarrowScene(t *testing.T) {
	s := FibonacciLessonState{N: 46, Values: []int{1836311903}, Known: []bool{true}, Phase: "done", HasResult: true, Result: 1836311903}
	svg := RenderFibonacciLessonSVG(trace.Step{State: s}, 320, 0)
	if strings.Contains(svg, "Verified result: F(46) = 1836311903") || !strings.Contains(svg, "F(46) = 1836311903") {
		t.Fatal("maximum result label does not fit narrow scene")
	}
}
