package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type NQueensLessonState struct {
	TableState
	N                   int
	Board               []int
	Row, Col            int
	Candidate, Conflict [2]int
	ConflictKind        string
	Depth, Solutions    int
	Phase               string
	LastSolution        []int
}

// RenderNQueensLessonSVG renders only the selected committed board and its evidence.
func RenderNQueensLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(NQueensLessonState)
	if width < 280 {
		width = 280
	}
	if width > 32768 {
		width = 32768
	}
	if height < 320 {
		height = 320
	}
	if height > 32768 {
		height = 32768
	}
	if !ok || s.N < 1 || s.N > 64 || len(s.Board) != s.N {
		return fmt.Sprintf(`<g data-lesson="nqueens" data-phase="empty" data-scene-width="%d" data-scene-height="%d"><text x="16" y="32" fill="#eff5ed" font-size="16">No N-Queens state</text></g>`, width, height)
	}
	// ponytail: 36-unit cells preserve labels; large boards use shared local scrolling.
	minWidth := s.N*36 + 56
	if width < minWidth {
		width = minWidth
	}
	side := min(width-56, height-212)
	if side < s.N*36 {
		side = s.N * 36
	}
	if height < side+212 {
		height = side + 212
	}
	cell := float64(side) / float64(s.N)
	x0, y0 := float64(width-side)/2, 88.0
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="nqueens" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif"><title>%s</title>`, html.EscapeString(s.Phase), width, height, width, html.EscapeString(step.Explanation))
	text := func(x, y float64, label, color string, size int) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" fill="%s" font-size="%d">%s</text>`, x, y, color, size, html.EscapeString(label))
	}
	heading := "Try, check, place, undo"
	if s.Phase == "done" {
		heading = "Search complete"
	}
	if s.Phase == "solution" {
		heading = "✓ Recorded; search continues"
	}
	text(16, 24, heading, "#eff5ed", 14)
	text(16, 47, fmt.Sprintf("Depth %d / %d · %d solutions", s.Depth, s.N, s.Solutions), "#bbcab9", 14)
	for r := 0; r < s.N; r++ {
		text(x0-22, y0+(float64(r)+0.5)*cell+4, fmt.Sprint(r), "#bbcab9", 12)
		text(x0+(float64(r)+0.5)*cell-4, y0-10, fmt.Sprint(r), "#bbcab9", 12)
		for c := 0; c < s.N; c++ {
			fill := "#202d39"
			if (r+c)%2 == 0 {
				fill = "#364451"
			}
			fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" stroke="#718096" stroke-width="0.5"/>`, x0+float64(c)*cell, y0+float64(r)*cell, cell, cell, fill)
		}
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	token := func(role string, r, c int, ghost bool) {
		if r < 0 || r >= s.N || c < 0 || c >= s.N {
			return
		}
		x, y := x0+(float64(c)+0.5)*cell, y0+(float64(r)+0.5)*cell
		fill, stroke, dash := palette[r%len(palette)], "#eff5ed", ""
		if ghost {
			fill = "none"
			stroke = "#ffe29a"
			dash = ` stroke-dasharray="4 3"`
		}
		if !ghost && s.Phase == "solution" {
			stroke = "#9bf0bb"
		}
		alias := ""
		if s.Phase == "place" && r == s.Row && !ghost {
			alias = fmt.Sprintf(` data-from-entity="candidate:%d"`, r)
		}
		if ghost {
			alias = ` data-emphasis="compare"`
		}
		fmt.Fprintf(&b, `<g data-entity="%s:%d" data-x="%.2f" data-y="%.2f" data-motion="linear"%s transform="translate(%.2f %.2f)"><circle r="%.2f" fill="%s" stroke="%s" stroke-width="2"%s/><text text-anchor="middle" y="5" fill="#eff5ed" font-size="14">Q%d</text></g>`, role, r, x, y, alias, x, y, min(cell*0.4, 20), fill, stroke, dash, r)
	}
	for r, c := range s.Board {
		token("queen", r, c, false)
	}
	if s.Phase == "try" || s.Phase == "safe" || s.Phase == "conflict" {
		token("candidate", s.Candidate[0], s.Candidate[1], true)
	}
	status := "Solid Q: placed · outline: try"
	if s.Phase == "conflict" && s.Conflict[0] >= 0 && s.Conflict[0] < s.N && s.Conflict[1] >= 0 && s.Conflict[1] < s.N {
		fmt.Fprintf(&b, `<path d="M%.2f %.2f L%.2f %.2f" fill="none" stroke="#ffe29a" stroke-width="3" stroke-dasharray="5 4"/>`, x0+(float64(s.Col)+0.5)*cell, y0+(float64(s.Row)+0.5)*cell, x0+(float64(s.Conflict[1])+0.5)*cell, y0+(float64(s.Conflict[0])+0.5)*cell)
		status = "Reject: |Δcol|=Δrow"
		if s.ConflictKind == "column" {
			status = "Reject: same column"
		}
	}
	if s.Phase == "safe" {
		status = "Safe: no earlier queen attacks."
	}
	if s.Phase == "remove" {
		status = fmt.Sprintf("Undo row %d; try its next column.", s.Row)
	}
	if s.Phase == "done" {
		status = "Board empty; search exhausted."
	}
	footer := y0 + float64(side) + 24
	text(16, footer, status, "#ffe29a", 14)
	if len(s.LastSolution) == s.N {
		text(16, footer+24, "Last recorded solution (cols):", "#9bf0bb", 14)
		text(16, footer+46, fmt.Sprint(s.LastSolution), "#eff5ed", 14)
	} else {
		text(16, footer+24, "No solution recorded yet.", "#bbcab9", 14)
	}
	b.WriteString(`</g>`)
	return b.String()
}
