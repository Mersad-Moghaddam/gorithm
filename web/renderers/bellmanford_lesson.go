package renderers

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type BellmanFordLessonState struct {
	GraphState
	Start, From, To                                            string
	Distance, PassStartDistance                                map[string]int
	Pass, MaxPasses, EdgeIndex, Weight, OldDistance, Candidate int
	HasCandidate, Improved, ChangedThisPass                    bool
	CheckingCycle, NegativeCycle, InvalidStart                 bool
	WitnessEdge                                                int
}

func BellmanFordDistanceLabel(d int) string {
	if d == math.MaxInt {
		return "∞"
	}
	return strconv.Itoa(d)
}

func bellmanFordShort(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}

func RenderBellmanFordLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(BellmanFordLessonState)
	if !ok {
		return `<g data-lesson="bellmanford"><title>No Bellman-Ford state</title></g>`
	}
	mobile := width < 600
	sceneW, graphX, graphY, graphW, graphH := 960.0, 24.0, 64.0, 592.0, 348.0
	panelX, panelY, panelW := 640.0, 64.0, 296.0
	if mobile {
		sceneW = 360
		graphX = 16
		graphY = 56
		graphW = 328
		graphH = 240
		panelX = 16
		panelY = 320
		panelW = 328
	}
	panelH := math.Max(348, 58+float64(len(s.Nodes))*30)
	stripY := math.Max(graphY+graphH, panelY+panelH) + 28
	cols := int((sceneW - 48) / 104)
	if cols < 1 {
		cols = 1
	}
	rows := (len(s.Edges) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	equationY := stripY + 32 + float64(rows)*64 + 24
	sceneH := equationY + 112
	phase := "relax"
	if step.LineID == "L1" || step.LineID == "L2" {
		phase = "initialize"
	}
	if s.CheckingCycle {
		phase = "cycle-scan"
	}
	if step.LineID == "L18" {
		phase = "complete"
	}
	if s.NegativeCycle {
		phase = "negative-cycle"
	}
	if s.InvalidStart {
		phase = "invalid"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="bellmanford" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif" font-size="14" fill="#eff5ed"><title>%s</title><defs><marker id="bellmanford-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`, phase, sceneW, sceneH, sceneW, html.EscapeString("Bellman-Ford: "+phase))
	text := func(x, y float64, value, color string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" fill="%s">%s</text>`, x, y, color, html.EscapeString(value))
	}
	title := "Bellman-Ford · in-place passes"
	if mobile {
		title = "Bellman-Ford · directed edges"
	}
	text(24, 30, title, "#eff5ed")
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="16" fill="#13211e" stroke="#52634e"/>`, graphX, graphY, graphW, graphH)
	positions := make(map[string][2]float64, len(s.Nodes))
	indices := map[string]int{}
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	ring := false
	for i, n := range s.Nodes {
		indices[n.ID] = i
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			ring = true
			continue
		}
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
		for j := 0; j < i; j++ {
			if n.X == s.Nodes[j].X && n.Y == s.Nodes[j].Y {
				ring = true
			}
		}
	}
	dx, dy := maxX-minX, maxY-minY
	scale := 1.0
	if dx > 0 && dy > 0 {
		scale = math.Min((graphW-100)/dx, (graphH-100)/dy)
	} else if dx > 0 {
		scale = (graphW - 100) / dx
	} else if dy > 0 {
		scale = (graphH - 100) / dy
	}
	for i, n := range s.Nodes {
		x, y := graphX+graphW/2, graphY+graphH/2
		if ring {
			a := 2*math.Pi*float64(i)/float64(len(s.Nodes)) - math.Pi/2
			x += (graphW/2 - 50) * math.Cos(a)
			y += (graphH/2 - 50) * math.Sin(a)
		} else {
			x += (n.X - (minX+maxX)/2) * scale
			y += (n.Y - (minY+maxY)/2) * scale
		}
		positions[n.ID] = [2]float64{x, y}
	}
	for i, e := range s.Edges {
		p, pok := positions[e.From]
		q, qok := positions[e.To]
		if !pok || !qok {
			continue
		}
		color := "#77918a"
		if i == s.EdgeIndex {
			color = "#ffe29a"
		}
		ux, uy := q[0]-p[0], q[1]-p[1]
		length := math.Hypot(ux, uy)
		mx, my := (p[0]+q[0])/2, (p[1]+q[1])/2
		path := ""
		if length < 1 {
			path = fmt.Sprintf("M %.2f %.2f C %.2f %.2f %.2f %.2f %.2f %.2f", p[0]-16, p[1]-16, p[0]-60, p[1]-85, p[0]+60, p[1]-85, p[0]+16, p[1]-16)
			mx = p[0]
			my = p[1] - 67
		} else {
			nx, ny := -uy/length, ux/length
			offset := 0.0
			parallel := 0
			for j, f := range s.Edges {
				if j == i {
					continue
				}
				if f.From == e.To && f.To == e.From {
					offset = 18
				}
				if j < i && f.From == e.From && f.To == e.To {
					parallel++
				}
			}
			offset += float64(parallel) * 28
			cx, cy := mx+nx*offset, my+ny*offset
			sx, sy := p[0], p[1]
			ex, ey := q[0], q[1]
			l1, l2 := math.Hypot(cx-sx, cy-sy), math.Hypot(ex-cx, ey-cy)
			sx += 22 * (cx - p[0]) / l1
			sy += 22 * (cy - p[1]) / l1
			ex -= 22 * (q[0] - cx) / l2
			ey -= 22 * (q[1] - cy) / l2
			path = fmt.Sprintf("M %.2f %.2f Q %.2f %.2f %.2f %.2f", sx, sy, cx, cy, ex, ey)
			mx = (sx+2*cx+ex)/4 + nx*10
			my = (sy+2*cy+ey)/4 + ny*10
		}
		fmt.Fprintf(&b, `<path id="bellmanford-edge-%d" d="%s" fill="none" stroke="%s" stroke-width="2" marker-end="url(#bellmanford-arrow)"/><rect x="%.2f" y="%.2f" width="52" height="22" rx="8" fill="#13211e"/>`, i, path, color, mx-26, my-16)
		text(mx-18, my, strconv.Itoa(e.Weight), "#eff5ed")
	}
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, n := range s.Nodes {
		p := positions[n.ID]
		outline := "#bbcab9"
		if n.ID == s.From || n.ID == s.To {
			outline = "#ffe29a"
		}
		fmt.Fprintf(&b, `<g data-entity="bellmanford:node:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><text y="5" text-anchor="middle" fill="#ffffff" stroke="#13211e" stroke-width="3" paint-order="stroke">%s</text></g>`, i, p[0], p[1], p[0], p[1], html.EscapeString(n.Label+" ("+n.ID+")"), colors[i%5], outline, html.EscapeString(bellmanFordShort(n.ID, 4)))
	}
	// Candidate identity persists from edge focus through comparison and assignment.
	if s.HasCandidate && s.EdgeIndex >= 0 {
		p, q := positions[s.From], positions[s.To]
		x, y := p[0], p[1]
		if step.LineID == "L6" || step.LineID == "L14" || step.LineID == "L7" || step.LineID == "L8" || step.LineID == "L15" {
			x, y = q[0], q[1]
		}
		motion := ""
		if step.LineID == "L6" || step.LineID == "L14" {
			samples := [][2]float64{p, q}
			raw, _ := json.Marshal(samples)
			motion = ` data-motion="path" data-path="` + html.EscapeString(string(raw)) + `"`
		}
		fmt.Fprintf(&b, `<g data-entity="bellmanford:candidate:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>Candidate %d</title><circle cy="-30" r="7" fill="#ffe29a" stroke="#13211e" stroke-width="2"/></g>`, s.EdgeIndex, x, y, x, y, motion, s.Candidate)
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="16" fill="#192923" stroke="#52634e"/>`, panelX, panelY, panelW, panelH)
	text(panelX+12, panelY+24, "Vertex     Pass start    Current", "#bbcab9")
	for i, n := range s.Nodes {
		y := panelY + 56 + float64(i)*30
		fmt.Fprintf(&b, `<g><title>%s</title>`, html.EscapeString(n.Label+" ("+n.ID+")"))
		text(panelX+12, y, bellmanFordShort(n.Label, 8), colors[i%5])
		text(panelX+116, y, BellmanFordDistanceLabel(s.PassStartDistance[n.ID]), "#bbcab9")
		value, exists := s.Distance[n.ID]
		label := "unknown"
		if exists {
			label = BellmanFordDistanceLabel(value)
		}
		text(panelX+218, y, label, "#eff5ed")
		b.WriteString(`</g>`)
	}
	header := fmt.Sprintf("pass %d / %d · changed: %t", s.Pass, s.MaxPasses, s.ChangedThisPass)
	if s.CheckingCycle {
		header = "Can any reachable edge still improve?"
	}
	if mobile && s.CheckingCycle {
		header = "Reachable edge still improves?"
	}
	text(24, stripY, header, "#ffe29a")
	for i, e := range s.Edges {
		x, y := 24+float64(i%cols)*104, stripY+18+float64(i/cols)*64
		stroke := "#52634e"
		if i == s.EdgeIndex {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&b, `<g><title>%s</title><rect x="%.0f" y="%.0f" width="96" height="48" rx="8" fill="#192923" stroke="%s"/>`, html.EscapeString(fmt.Sprintf("Edge %d: %s → %s (%d)", i+1, e.From, e.To, e.Weight)), x, y, stroke)
		text(x+6, y+19, fmt.Sprintf("%d: %s→%s", i+1, bellmanFordShort(e.From, 3), bellmanFordShort(e.To, 3)), "#eff5ed")
		text(x+6, y+38, fmt.Sprintf("weight %d", e.Weight), "#bbcab9")
		b.WriteString(`</g>`)
	}
	cursor := s.EdgeIndex
	if cursor < 0 {
		cursor = 0
	}
	cx, cy := 24+float64(cursor%cols)*104, stripY+18+float64(cursor/cols)*64
	fmt.Fprintf(&b, `<g data-entity="bellmanford:cursor" data-x="%.0f" data-y="%.0f" transform="translate(%.0f %.0f)"><path d="M0 -5H96" stroke="#ffe29a" stroke-width="3"/></g>`, cx, cy, cx, cy)
	equation := "Source distance 0; unknown vertices ∞."
	result := "Distances remain tentative until final scan."
	if s.From != "" {
		if !s.HasCandidate {
			equation = "Unreachable source: skip addition."
			result = "No candidate; target unchanged."
		} else {
			equation = fmt.Sprintf("%d + (%d) = %d", s.Candidate-s.Weight, s.Weight, s.Candidate)
			result = fmt.Sprintf("candidate %d < old %s: %t", s.Candidate, BellmanFordDistanceLabel(s.OldDistance), s.Improved)
		}
	}
	if step.LineID == "L7" {
		result = "✓ updated · not final shortest"
	}
	if step.LineID == "L18" {
		equation = "✓ shortest finite distances confirmed"
		result = "∞ means unreachable from source."
	}
	if s.InvalidStart {
		equation = "Invalid start"
		result = "No shortest-path result."
	}
	if s.NegativeCycle {
		equation = "Reachable negative cycle"
		result = fmt.Sprintf("Witness edge %d still improves.", s.WitnessEdge+1)
	}
	text(24, equationY, equation, "#eff5ed")
	text(24, equationY+28, result, "#ffe29a")
	text(24, equationY+56, "Gold: comparison · ✓: updated/result", "#bbcab9")
	b.WriteString(`</g>`)
	return b.String()
}
