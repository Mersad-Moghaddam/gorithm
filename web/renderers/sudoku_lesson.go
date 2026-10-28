package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type SudokuLessonState struct {
	TableState
	Board                      [9][9]int
	Given                      [9][9]bool
	Row, Col, Candidate, Depth int
	HasCandidate               bool
	Conflicts                  [][2]int
	ConflictKinds              []string
	Phase                      string
	Status                     string
}

func RenderSudokuLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(SudokuLessonState)
	if !ok {
		return `<g data-lesson="sudoku" data-phase="empty" data-scene-width="320" data-scene-height="120"><title>Sudoku state unavailable</title><text x="16" y="40" fill="#eff5ed" font-size="14">Sudoku state unavailable</text></g>`
	}
	// ponytail: fixed nine-cell board; wider Sudoku variants need a different lesson.
	if width < 320 {
		width = 320
	}
	if width > 1440 {
		width = 1440
	}
	cell := 32.0
	if width < 390 {
		cell = 30
	}
	boardSize := cell * 9
	x := (float64(width) - boardSize) / 2
	y := 76.0
	sceneHeight := int(y + boardSize + 152)
	if height > sceneHeight && height < 32769 {
		sceneHeight = height
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="sudoku" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="320" font-family="system-ui, sans-serif"><title>Sudoku: %s; depth %d</title>`, html.EscapeString(s.Phase), width, sceneHeight, html.EscapeString(s.Status), s.Depth)
	heading := "Searching"
	switch s.Status {
	case "solved":
		heading = "✓ Solved"
	case "invalid":
		heading = "Invalid givens"
	case "unsatisfiable":
		heading = "No solution"
	}
	fmt.Fprintf(&b, `<text x="16" y="24" fill="#eff5ed" font-size="16">%s · depth %d</text>`, html.EscapeString(heading), s.Depth)
	label := "Select first empty cell"
	if s.Phase == "done" {
		label = "Search finished"
	}
	if s.HasCandidate {
		label = fmt.Sprintf("Candidate %d · cell (%d,%d)", s.Candidate, s.Row, s.Col)
	}
	if s.Phase == "remove" {
		label = fmt.Sprintf("Undo %d at (%d,%d)", s.Candidate, s.Row, s.Col)
	}
	fmt.Fprintf(&b, `<text x="16" y="49" fill="#bbcab9" font-size="14">%s</text>`, html.EscapeString(label))
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			cx, cy := x+float64(c)*cell, y+float64(r)*cell
			fill := "#192923"
			if s.Row == r || s.Col == c || (s.Row >= 0 && s.Col >= 0 && r/3 == s.Row/3 && c/3 == s.Col/3) {
				fill = "#343c2b"
			}
			fmt.Fprintf(&b, `<g data-entity="sudoku-cell:%d:%d" data-x="%.1f" data-y="%.1f" transform="translate(%.1f %.1f)"><rect width="%.1f" height="%.1f" fill="%s" stroke="#52634e"/>`, r, c, cx, cy, cx, cy, cell, cell, fill)
			if s.Board[r][c] != 0 {
				color := palette[(r*9+c)%5]
				weight := "500"
				if s.Given[r][c] {
					weight = "700"
				}
				fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" fill="#eff5ed" font-size="14" font-weight="%s">%d</text>`, cell/2, cell/2+5, weight, s.Board[r][c])
				if s.Given[r][c] {
					fmt.Fprintf(&b, `<circle cx="%.1f" cy="5" r="1.5" fill="#bbcab9"/>`, cell-5)
				} else {
					fmt.Fprintf(&b, `<rect x="3" y="3" width="%.1f" height="%.1f" rx="4" fill="none" stroke="%s"/>`, cell-6, cell-6, color)
				}
			}
			b.WriteString(`</g>`)
		}
	}
	for i := 0; i <= 9; i++ {
		stroke := 1
		if i%3 == 0 {
			stroke = 3
		}
		fmt.Fprintf(&b, `<path d="M%.1f %.1fH%.1f M%.1f %.1fV%.1f" fill="none" stroke="#bbcab9" stroke-width="%d"/>`, x, y+float64(i)*cell, x+boardSize, x+float64(i)*cell, y, y+boardSize, stroke)
	}
	active := s.Row >= 0 && s.Row < 9 && s.Col >= 0 && s.Col < 9
	if active {
		fmt.Fprintf(&b, `<g data-emphasis="compare"><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#ffe29a" stroke-width="2"/><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#ffe29a" stroke-dasharray="4 3"/><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#ffe29a" stroke-dasharray="2 3"/></g>`, x, y+float64(s.Row)*cell, boardSize, cell, x+float64(s.Col)*cell, y, cell, boardSize, x+float64(s.Col/3*3)*cell, y+float64(s.Row/3*3)*cell, cell*3, cell*3)
	}
	stripY := y + boardSize + 30
	fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="30" rx="4" fill="#192923"/>`, x, stripY-21, boardSize)
	for n := 1; n <= 9; n++ {
		cx := x + float64(n-1)*cell
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" fill="#eff5ed" font-size="14" text-anchor="middle">%d</text>`, cx+cell/2, stripY, n)
	}
	if s.HasCandidate && active {
		cx, cy := x+float64(s.Candidate-1)*cell+cell/2, stripY
		if s.Phase == "place" || s.Phase == "descend" {
			cx = x + float64(s.Col)*cell + cell/2
			cy = y + float64(s.Row)*cell + cell/2 + 5
		}
		fmt.Fprintf(&b, `<g data-entity="sudoku-candidate" data-x="%.1f" data-y="%.1f" transform="translate(%.1f %.1f)" data-motion="linear"><circle cy="-5" r="12" fill="none" stroke="#ffe29a" stroke-width="2"/></g>`, cx, cy, cx, cy)
		if (s.Phase == "try" || s.Phase == "reject") && s.Board[s.Row][s.Col] == 0 {
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" fill="#ffe29a" font-size="14" text-anchor="middle">%d?</text>`, x+float64(s.Col)*cell+cell/2, y+float64(s.Row)*cell+cell/2+5, s.Candidate)
		}
	}
	reason := "Row / column / 3×3 box"
	if s.Status == "invalid" {
		reason = "Givens violate Sudoku rules"
	}
	if s.Status == "unsatisfiable" {
		reason = "All branches failed; givens retained"
	}
	if s.Status == "solved" {
		reason = "All 81 cells satisfy Sudoku rules"
	}
	if len(s.Conflicts) > 0 && len(s.ConflictKinds) > 0 {
		p := s.Conflicts[0]
		reason = fmt.Sprintf("%s conflict at (%d,%d)", s.ConflictKinds[0], p[0], p[1])
		if p[0] >= 0 && p[0] < 9 && p[1] >= 0 && p[1] < 9 {
			fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" fill="#ff9b9b" font-size="14">×</text>`, x+float64(p[1])*cell+2, y+float64(p[0])*cell+12)
		}
	}
	if s.Phase == "exhausted" {
		reason = "All candidates exhausted; backtrack"
	}
	if s.Phase == "place" {
		reason = "Fits all three checks; placed"
	}
	if s.Phase == "remove" {
		reason = "Branch failed; guessed cell cleared"
	}
	fmt.Fprintf(&b, `<text x="16" y="%.1f" fill="#bbcab9" font-size="14">%s</text><text x="16" y="%.1f" fill="#bbcab9" font-size="14">Given: bold + dot · Guess: outline</text>`, stripY+32, html.EscapeString(reason), stripY+60)
	if s.Status == "solved" {
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="#9bf0bb" stroke-width="4"/>`, x, y, boardSize, boardSize)
	}
	b.WriteString(`</g>`)
	return b.String()
}
