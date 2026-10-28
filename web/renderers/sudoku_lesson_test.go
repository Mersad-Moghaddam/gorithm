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

func TestSudokuDigitContrast(t *testing.T) {
	s := SudokuLessonState{Row: 0, Col: 0, Phase: "place", Status: "searching"}
	for r := range s.Board {
		for c := range s.Board[r] {
			s.Board[r][c] = c + 1
		}
	}
	s.Given[0][0] = true
	svg := RenderSudokuLessonSVG(trace.Step{State: s}, 320, 450)
	luminance := func(color string) float64 {
		n, err := strconv.ParseUint(strings.TrimPrefix(color, "#"), 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		var result float64
		for i, weight := range []float64{0.2126, 0.7152, 0.0722} {
			v := float64((n>>uint(16-i*8))&255) / 255
			if v <= 0.04045 {
				v /= 12.92
			} else {
				v = math.Pow((v+0.055)/1.055, 2.4)
			}
			result += weight * v
		}
		return result
	}
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	digits := 0
	for {
		tok, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		start, ok := tok.(xml.StartElement)
		if !ok || start.Name.Local != "text" {
			continue
		}
		var text string
		if err := d.DecodeElement(&text, &start); err != nil {
			t.Fatal(err)
		}
		if _, err := strconv.Atoi(text); err != nil {
			continue
		}
		digits++
		fg := ""
		for _, a := range start.Attr {
			if a.Name.Local == "fill" {
				fg = a.Value
			}
		}
		for _, bg := range []string{"#192923", "#343c2b"} {
			if !strings.Contains(svg, `fill="`+bg+`"`) {
				t.Fatalf("background %s not rendered", bg)
			}
			a, b := luminance(fg), luminance(bg)
			if a < b {
				a, b = b, a
			}
			if ratio := (a + 0.05) / (b + 0.05); ratio < 4.5 {
				t.Errorf("digit %s foreground %s on %s contrast %.2f", text, fg, bg, ratio)
			}
		}
	}
	if digits != 90 {
		t.Fatalf("checked %d digits, want 81 board + 9 strip", digits)
	}
	if !strings.Contains(svg, `font-weight="700"`) || !strings.Contains(svg, `font-weight="500"`) || !strings.Contains(svg, `stroke="#387faa"`) {
		t.Fatal("given/guess cues or identity outline missing")
	}
}
