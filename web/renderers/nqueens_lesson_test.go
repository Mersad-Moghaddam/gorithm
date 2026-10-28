package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestNQueensLessonSVG(t *testing.T) {
	s := NQueensLessonState{N: 4, Board: []int{0, -1, -1, -1}, Row: 1, Col: 1, Candidate: [2]int{1, 1}, Conflict: [2]int{0, 0}, ConflictKind: "diagonal", Phase: "conflict"}
	for _, w := range []int{0, 1, 320, 390, 768, 1440} {
		svg := RenderNQueensLessonSVG(trace.Step{State: s, Explanation: `<script> & "x"`}, w, 450)
		for _, want := range []string{`data-lesson="nqueens"`, `data-phase="conflict"`, `data-scene-width=`, `data-scene-height=`, `data-min-width=`, `data-entity="queen:0"`, `data-entity="candidate:1"`, `data-motion="linear"`, `|Δcol|=Δrow`, `&lt;script&gt;`, `stroke-dasharray=`} {
			if !strings.Contains(svg, want) {
				t.Errorf("width %d missing %s", w, want)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "<script>") {
			t.Fatal("invalid SVG")
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
	s.Phase = "done"
	s.Board = []int{-1, -1, -1, -1}
	s.Candidate = [2]int{-1, -1}
	s.Solutions = 2
	s.LastSolution = []int{2, 0, 3, 1}
	svg := RenderNQueensLessonSVG(trace.Step{State: s}, 320, 450)
	if !strings.Contains(svg, "Search complete") || !strings.Contains(svg, "2 solutions") || !strings.Contains(svg, "Last recorded solution") {
		t.Fatal("missing honest completion")
	}
	if strings.Contains(svg, `data-entity="queen:`) {
		t.Fatal("past solution painted as live")
	}
	if !strings.Contains(RenderNQueensLessonSVG(trace.Step{}, -1, -1), "No N-Queens state") {
		t.Fatal("missing fallback")
	}
	s.Phase = "try"
	s.Row, s.Col = 1, 2
	s.Candidate = [2]int{1, 2}
	try := RenderNQueensLessonSVG(trace.Step{State: s}, 320, 450)
	s.Col = 3
	s.Candidate = [2]int{1, 3}
	next := RenderNQueensLessonSVG(trace.Step{State: s}, 320, 450)
	if try == next || !strings.Contains(next, `data-entity="candidate:1"`) || !strings.Contains(next, `data-emphasis="compare"`) {
		t.Fatal("candidate movement lacks stable identity")
	}
	s.Phase = "place"
	s.Board[1] = 3
	if !strings.Contains(RenderNQueensLessonSVG(trace.Step{State: s}, 320, 450), `data-from-entity="candidate:1"`) {
		t.Fatal("placement lacks candidate alias")
	}
}
