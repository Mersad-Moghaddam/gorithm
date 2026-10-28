package renderers

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type FloydWarshallLessonState struct {
	TableState
	Nodes                                                   []string
	Edges                                                   []GraphEdge
	Distance                                                [][]int64
	I, J, K                                                 int
	LeftDistance, RightDistance, OldDistance, Candidate     int64
	HasCandidate, Improved, CheckingDiagonal, NegativeCycle bool
	DiagonalIndex                                           int
}

func floydWarshallDistance(v int64) string {
	if v == math.MaxInt64 {
		return "∞"
	}
	return strconv.FormatInt(v, 10)
}

// RenderFloydWarshallLessonSVG renders only the selected stored matrix and operands.
func RenderFloydWarshallLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(FloydWarshallLessonState)
	if !ok {
		return `<g data-lesson="floydwarshall"><title>Missing Floyd–Warshall state</title></g>`
	}
	n := len(s.Nodes)
	cellW, cellH, x0, y0 := 58, 48, 72, 100
	sceneW, routeX, routeY := 960, 420, 176
	mobile := width < 600
	if mobile {
		cellW = 48
		x0 = 48
		y0 = 80
		sceneW = 360
		routeX = 48
		routeY = y0 + n*cellH + 100
	}
	// Negative cycles can grow distances throughout playback; reserve width up front.
	for _, edge := range s.Edges {
		if edge.Weight < 0 {
			cellW = 19*9 + 48
			break
		}
	}
	for _, row := range s.Distance {
		for _, value := range row {
			if needed := len(floydWarshallDistance(value))*9 + 48; needed > cellW {
				cellW = needed
			}
		}
	}
	matrixRight := x0 + n*cellW
	if !mobile && matrixRight+56 > routeX {
		routeX = matrixRight + 56
	}
	if mobile {
		if matrixRight+24 > sceneW {
			sceneW = matrixRight + 24
		}
	} else if routeX+396 > sceneW {
		sceneW = routeX + 396
	}
	selectorY := routeY + 250
	if y0+n*cellH+48 > selectorY {
		selectorY = y0 + n*cellH + 48
	}
	if mobile {
		selectorY = routeY + 250
	}
	chipCols := (sceneW - 48) / 80
	if chipCols < 1 {
		chipCols = 1
	}
	equationY := selectorY + ((n+chipCols-1)/chipCols)*48 + 56
	sceneH := equationY + 230
	if sceneH < 640 {
		sceneH = 640
	}
	phase := step.LineID
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="floydwarshall" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>%s</title><rect width="%d" height="%d" rx="16" fill="#101b25"/>`, html.EscapeString(phase), sceneW, sceneH, sceneW, html.EscapeString(step.Explanation), sceneW, sceneH)
	floydWarshallText(&b, 24, 30, "Floyd–Warshall · distance matrix")
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, name := range s.Nodes {
		fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="end"><title>%s</title>%d</text><text x="%d" y="%d" text-anchor="middle"><title>%s</title>%d</text>`, x0-16, y0+i*cellH+28, html.EscapeString(name), i, x0+i*cellW+cellW/2, y0-18, html.EscapeString(name), i)
		for j := range s.Nodes {
			x, y := x0+j*cellW, y0+i*cellH
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" rx="5" fill="#172735" stroke="#617285"/>`, x, y, cellW-3, cellH-3)
			roles := []string{}
			if s.I >= 0 && s.K >= 0 && s.J >= 0 {
				if i == s.I && j == s.K {
					roles = append(roles, "left")
				}
				if i == s.K && j == s.J {
					roles = append(roles, "right")
				}
				if i == s.I && j == s.J {
					roles = append(roles, "dest")
				}
			}
			for k, role := range roles {
				fmt.Fprintf(&b, `<rect data-role="%s" data-emphasis="compare" x="%d" y="%d" width="%d" height="%d" rx="4" fill="none" stroke="#ffe29a" stroke-width="1.5"><title>%s operand role</title></rect>`, role, x+2+k*3, y+2+k*3, cellW-7-k*6, cellH-7-k*6, role)
			}
			if s.CheckingDiagonal && i == j {
				fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="none" stroke="#ff9b9b" stroke-width="3"><title>Diagonal check</title></rect>`, x+1, y+1, cellW-5, cellH-5)
			}
			value := "?"
			if i < len(s.Distance) && j < len(s.Distance[i]) {
				value = floydWarshallDistance(s.Distance[i][j])
			}
			fmt.Fprintf(&b, `<g data-entity="floydwarshall:cell:%d:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"><text text-anchor="middle" y="5">%s</text></g>`, i, j, x+cellW/2, y+cellH/2, x+cellW/2, y+cellH/2, html.EscapeString(value))
		}
	}
	if s.I >= 0 && s.J >= 0 && s.K >= 0 && s.I < n && s.J < n && s.K < n {
		gap := 170
		if mobile {
			gap = 120
		}
		fmt.Fprintf(&b, `<defs><marker id="floydwarshall-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#8bd5ff"/></marker></defs><path d="M%d %dL%d %d M%d %dL%d %d" stroke="#8bd5ff" fill="none" marker-end="url(#floydwarshall-arrow)"/>`, routeX+22, routeY-10, routeX+gap-22, routeY-70, routeX+gap+22, routeY-70, routeX+2*gap-22, routeY-10)
		for pos, id := range []int{s.I, s.K, s.J} {
			y := routeY
			if pos == 1 {
				y -= 80
			}
			x := routeX + pos*gap
			fmt.Fprintf(&b, `<circle cx="%d" cy="%d" r="22" fill="%s"/><text x="%d" y="%d" text-anchor="middle">%s</text>`, x, y, palette[id%5], x, y+5, []string{"i", "k", "j"}[pos])
			floydWarshallLabel(&b, x, y+42, s.Nodes[id], "middle", 14)
		}
		floydWarshallText(&b, routeX, routeY+78, "Distance routes, not direct input edges")
		fmt.Fprintf(&b, `<path d="M%d %dQ%d %d %d %d" fill="none" stroke="#bbcab9" stroke-dasharray="5 4"/>`, routeX, routeY+52, routeX+gap, routeY+125, routeX+2*gap, routeY+52)
		floydWarshallText(&b, routeX, routeY+128, "left: "+floydWarshallDistance(s.LeftDistance))
		floydWarshallText(&b, routeX, routeY+150, "right: "+floydWarshallDistance(s.RightDistance))
		floydWarshallText(&b, routeX, routeY+172, "old: "+floydWarshallDistance(s.OldDistance))
	}
	for i, name := range s.Nodes {
		x, y := 24+(i%chipCols)*80, selectorY+(i/chipCols)*48
		stroke := "#617285"
		if i == s.K {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="72" height="36" rx="9" fill="#172735" stroke="%s"/>`, x, y, stroke)
		floydWarshallLabel(&b, x+36, y+16, fmt.Sprintf("k=%d %s", i, name), "middle", 8)
	}
	if s.HasCandidate {
		floydWarshallText(&b, 24, equationY, fmt.Sprintf("d[%d][%d] = min(%s, %s + %s) = %s", s.I, s.J, floydWarshallDistance(s.OldDistance), floydWarshallDistance(s.LeftDistance), floydWarshallDistance(s.RightDistance), floydWarshallDistance(min(s.OldDistance, s.Candidate))))
		// Operand copies alias source positions; commit token returns to the destination.
		if s.Improved {
			fmt.Fprintf(&b, `<g data-entity="floydwarshall:commit" data-from-entity="floydwarshall:operand:left" data-x="%d" data-y="%d" transform="translate(%d %d)"><circle r="12" fill="#172735" stroke="#9bf0bb"/><text data-commit-marker="check" y="5" text-anchor="middle">✓</text><title>Commit candidate %s to destination</title></g>`, x0+s.J*cellW+cellW-16, y0+s.I*cellH+14, x0+s.J*cellW+cellW-16, y0+s.I*cellH+14, floydWarshallDistance(s.Candidate))
		} else {
			for p, role := range []string{"left", "right"} {
				r, c, value := s.I, s.K, s.LeftDistance
				if p == 1 {
					r = s.K
					c = s.J
					value = s.RightDistance
				}
				x, y := 24, equationY+80+p*22
				fmt.Fprintf(&b, `<g data-operand-layout="stacked" data-entity="floydwarshall:operand:%s" data-from-entity="floydwarshall:cell:%d:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"><text y="5">%s = %s</text></g>`, role, r, c, x, y, x, y, role, floydWarshallDistance(value))
			}
		}
	} else if step.LineID == "L6" {
		floydWarshallText(&b, 24, equationY, "Unreachable half-path (∞): no candidate, no addition.")
	}
	status := "Gold rings: left / right / destination (concentric when shared)."
	if s.NegativeCycle {
		status = "Negative cycle; finite optimum not defined for affected pairs."
	} else if step.LineID == "L19" {
		status = "✓ Complete: no negative cycle; all-pairs distances settled."
	} else if s.CheckingDiagonal {
		status = "Check diagonal d[i][i] for a negative cycle."
	}
	floydWarshallText(&b, 24, equationY+166, status)
	b.WriteString(`</g>`)
	return b.String()
}

func floydWarshallText(b *strings.Builder, x, y int, text string) {
	fmt.Fprintf(b, `<text x="%d" y="%d"><title>%s</title>`, x, y, html.EscapeString(text))
	words := strings.Fields(text)
	line := ""
	row := 0
	for _, word := range words {
		if len([]rune(line+" "+word)) > 39 && line != "" {
			fmt.Fprintf(b, `<tspan x="%d" y="%d">%s</tspan>`, x, y+row*18, html.EscapeString(line))
			row++
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	fmt.Fprintf(b, `<tspan x="%d" y="%d">%s</tspan></text>`, x, y+row*18, html.EscapeString(line))
}
func floydWarshallLabel(b *strings.Builder, x, y int, text, anchor string, limit int) {
	runes := []rune(text)
	visible := text
	if len(runes) > limit*2 {
		visible = string(runes[:limit*2-1]) + "…"
	}
	r := []rune(visible)
	fmt.Fprintf(b, `<text x="%d" y="%d" text-anchor="%s" font-size="14"><title>%s</title>`, x, y, anchor, html.EscapeString(text))
	for i := 0; i < len(r); i += limit {
		end := i + limit
		if end > len(r) {
			end = len(r)
		}
		fmt.Fprintf(b, `<tspan x="%d" dy="%d">%s</tspan>`, x, func() int {
			if i == 0 {
				return 0
			}
			return 16
		}(), html.EscapeString(string(r[i:end])))
	}
	b.WriteString(`</text>`)
}
