package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestSudokuLessonSVG(t *testing.T) {
	s := SudokuLessonState{Row: 0, Col: 2, Candidate: 4, HasCandidate: true, Phase: "try", Status: "searching"}
	s.Board[0][0] = 5
	s.Given[0][0] = true
	for _, width := range []int{0, 1, 320, 390, 768, 1440} {
		svg := RenderSudokuLessonSVG(trace.Step{State: s}, width, 0)
		for _, part := range []string{`data-lesson="sudoku"`, `data-phase="try"`, `data-scene-width=`, `data-scene-height=`, `data-entity="sudoku-cell:0:0"`, `data-entity="sudoku-candidate"`, `data-emphasis="compare"`, `Candidate 4`, `font-size="14"`, `Given`} {
			if !strings.Contains(svg, part) {
				t.Errorf("missing %s", part)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "<svg") {
			t.Fatal("invalid fragment")
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		for {
			_, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	s.Phase = `"><script>`
	s.Status = `<script>`
	svg := RenderSudokuLessonSVG(trace.Step{State: s}, 320, 450)
	if strings.Contains(svg, "<script>") {
		t.Fatal("unescaped state")
	}
	s.Phase = "done"
	s.Status = "invalid"
	if strings.Contains(RenderSudokuLessonSVG(trace.Step{State: s}, 320, 450), "✓ Solved") {
		t.Fatal("invalid marked solved")
	}
	s.Status = "solved"
	if !strings.Contains(RenderSudokuLessonSVG(trace.Step{State: s}, 320, 450), "✓ Solved") {
		t.Fatal("missing verified result")
	}
}
