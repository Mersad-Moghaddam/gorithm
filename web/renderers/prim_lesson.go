package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"math"
	"strings"
)

type PrimCandidate struct {
	From, To                  string
	Weight, EdgeIndex, Serial int
}
type PrimLessonState struct {
	GraphState
	Start, Current                                         string
	InTree                                                 map[string]bool
	Candidates                                             []PrimCandidate
	SelectedEdges                                          []int
	ActiveEdge, ActiveSerial, TotalWeight, CandidateWeight int
	CutEdges                                               []int
	Discarded, Connected                                   bool
}

func RenderPrimLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(PrimLessonState)
	if !ok {
		return `<g data-lesson="prim"><title>Prim state unavailable</title></g>`
	}
	w, graphX, graphY, graphW, graphH, panelX, panelY, panelW := 960.0, 24.0, 64.0, 592.0, 348.0, 640.0, 64.0, 296.0
	mobile := width < 600
	if mobile {
		w = 360
		graphX = 16
		graphY = 56
		graphW = 328
		graphH = 240
		panelX = 16
		panelY = 320
		panelW = 328
	}
	panelH := math.Max(348, float64(100+30*len(s.Nodes)))
	stripY := math.Max(graphY+graphH, panelY+panelH) + 32
	cols := int((w - 48) / 184)
	if cols < 1 {
		cols = 1
	}
	heapRows := (len(s.Candidates) + cols - 1) / cols
	if heapRows < 1 {
		heapRows = 1
	}
	acceptedY := stripY + 40 + float64(heapRows)*48
	acceptedRows := (len(s.SelectedEdges) + cols - 1) / cols
	if acceptedRows < 1 {
		acceptedRows = 1
	}
	loopCount := 0
	for _, e := range s.Edges {
		if e.From == e.To {
			loopCount++
		}
	}
	loopY := acceptedY + 40 + float64(acceptedRows)*48
	equationY := loopY
	if loopCount > 0 {
		equationY += 40 + float64((loopCount+cols-1)/cols)*48
	}
	sceneH := math.Max(640, equationY+120)
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="prim" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(step.LineID), w, sceneH, w, html.EscapeString("Prim: "+step.Explanation))
	text := func(x, y float64, v string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f">%s</text>`, x, y, html.EscapeString(v))
	}
	text(24, 30, "Prim · growing tree and its cut")
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="16" fill="#101c27" stroke="#52634e"/>`, graphX, graphY, graphW, graphH)
	pos := primLessonPositions(s.Nodes, graphX, graphY, graphW, graphH)
	ids := map[string]int{}
	for i, n := range s.Nodes {
		ids[n.ID] = i
	}
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	cut := map[int]bool{}
	for _, i := range s.CutEdges {
		cut[i] = true
	}
	selected := map[int]bool{}
	for _, i := range s.SelectedEdges {
		selected[i] = true
	}
	for i, e := range s.Edges {
		a, aok := ids[e.From]
		z, zok := ids[e.To]
		if !aok || !zok {
			continue
		}
		p, q := pos[a], pos[z]
		if a == z {
			continue // Self-loops occupy their own ledger, never the cut or MST.
		}
		color, dash := "#7e929e", ""
		if cut[i] {
			color = "#ffe29a"
			dash = ` stroke-dasharray="7 5"`
		}
		if selected[i] {
			color = colors[a%5]
		}
		rank, count := 0, 0
		for j, other := range s.Edges {
			if (other.From == e.From && other.To == e.To) || (other.From == e.To && other.To == e.From) {
				if j < i {
					rank++
				}
				count++
			}
		}
		dx, dy := q[0]-p[0], q[1]-p[1]
		length := math.Hypot(dx, dy)
		mx, my := (p[0]+q[0])/2, (p[1]+q[1])/2
		path := ""
		if length > 0 {
			offset := float64(rank)*32 - float64(count-1)*16
			// Canonical normal separates reversed parallel input edges too.
			nx, ny := -dy/length, dx/length
			if a > z {
				nx, ny = -nx, -ny
			}
			cx, cy := mx+nx*offset, my+ny*offset
			d1, d2 := math.Hypot(cx-p[0], cy-p[1]), math.Hypot(cx-q[0], cy-q[1])
			path = fmt.Sprintf("M %.2f %.2f Q %.2f %.2f %.2f %.2f", p[0]+22*(cx-p[0])/d1, p[1]+22*(cy-p[1])/d1, cx, cy, q[0]+22*(cx-q[0])/d2, q[1]+22*(cy-q[1])/d2)
			mx = (p[0]+2*cx+q[0])/4 + nx*10
			my = (p[1]+2*cy+q[1])/4 + ny*10
		}
		fmt.Fprintf(&b, `<path id="prim-edge-%d" d="%s" fill="none" stroke="%s" stroke-width="%d"%s><title>Edge #%d: %s–%s, weight %d</title></path>`, i, path, color, 2+2*primLessonBool(selected[i]), dash, i, html.EscapeString(e.From), html.EscapeString(e.To), e.Weight)
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="64" height="22" rx="8" fill="#101c27"/>`, mx-32, my-16)
		text(mx-22, my, fmt.Sprint(e.Weight))
		if selected[i] {
			text(mx+18, my, "✓")
			alias := ""
			if step.LineID == "L12" && i == s.ActiveEdge {
				alias = fmt.Sprintf(` data-from-entity="prim-probe:%d"`, s.ActiveSerial)
			}
			fmt.Fprintf(&b, `<g data-entity="prim-edge-check:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>Accepted original edge #%d</title><circle r="10" fill="#101c27" stroke="#9bf0bb"/><text x="-6" y="5" fill="#9bf0bb">✓</text></g>`, i, mx+28, my+10, mx+28, my+10, alias, i)
		}
	}
	for i, n := range s.Nodes {
		p := pos[i]
		fmt.Fprintf(&b, `<g data-entity="prim-node:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><circle r="22" fill="%s" stroke="#eff5ed" stroke-width="2"/>`, i, p[0], p[1], p[0], p[1], html.EscapeString(n.Label), colors[i%5])
		for _, e := range s.Edges {
			if e.From == n.ID && e.To == n.ID {
				b.WriteString(`<path d="M -16 -16 C -52 -78 52 -78 16 -16" fill="none" stroke="#7e929e"><title>Self-loops listed below</title></path>`)
				break
			}
		}
		label := n.Label
		if label == "" {
			label = n.ID
		}
		r := []rune(label)
		if len(r) > 12 {
			r = r[:12]
		}
		for j := 0; j < len(r); j += 6 {
			end := j + 6
			if end > len(r) {
				end = len(r)
			}
			fmt.Fprintf(&b, `<text x="0" y="%d" text-anchor="middle" fill="#ffffff" stroke="#101c27" stroke-width="3" paint-order="stroke">%s</text>`, -2+(j/6)*15, html.EscapeString(string(r[j:end])))
		}
		if s.InTree[n.ID] {
			text(-22, 42, "✓ in tree")
		}
		b.WriteString(`</g>`)
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="16" fill="#162531" stroke="#52634e"/>`, panelX, panelY, panelW, panelH)
	text(panelX+12, panelY+24, fmt.Sprintf("Selected %d / %d · total %d", len(s.SelectedEdges), max(0, len(s.Nodes)-1), s.TotalWeight))
	text(panelX+12, panelY+50, fmt.Sprintf("Cut edges: %d · dashed gold", len(s.CutEdges)))
	for i, n := range s.Nodes {
		membership := "outside"
		if s.InTree[n.ID] {
			membership = "✓ in tree"
		}
		text(panelX+12, panelY+82+float64(i)*30, fmt.Sprintf("%d · %s: %s", i, primLessonShort(n.ID, 14), membership))
	}
	text(24, stripY, "Candidate heap · array order, not sorted")
	chip := func(entity, alias string, x, y float64, label string) {
		fmt.Fprintf(&b, `<g data-entity="%s" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"`, entity, x, y, x, y)
		if alias != "" {
			fmt.Fprintf(&b, ` data-from-entity="%s"`, alias)
		}
		fmt.Fprintf(&b, `><title>%s</title><rect width="176" height="36" rx="8" fill="#223744" stroke="#ffe29a"/><text x="8" y="23">%s</text></g>`, html.EscapeString(label), html.EscapeString(primLessonShort(label, 22)))
	}
	for i, c := range s.Candidates {
		label := fmt.Sprintf("#%d %s–%s: %d", c.EdgeIndex, c.From, c.To, c.Weight)
		if s.InTree[c.From] && s.InTree[c.To] {
			label = "internal · " + label
		}
		chip(fmt.Sprintf("prim-heap:%d", c.Serial), "", 24+float64(i%cols)*184, stripY+16+float64(i/cols)*48, label)
	}
	text(24, acceptedY, "Accepted edges · original input IDs")
	for i, id := range s.SelectedEdges {
		alias := ""
		if step.LineID == "L12" && id == s.ActiveEdge {
			alias = fmt.Sprintf("prim-probe:%d", s.ActiveSerial)
		}
		chip(fmt.Sprintf("prim-selected:%d", id), alias, 24+float64(i%cols)*184, acceptedY+16+float64(i/cols)*48, fmt.Sprintf("✓ edge #%d", id))
	}
	if loopCount > 0 {
		text(24, loopY, "Self-loops · never cross the cut")
		loop := 0
		for i, e := range s.Edges {
			if e.From != e.To {
				continue
			}
			x, y := 24+float64(loop%cols)*184, loopY+16+float64(loop/cols)*48
			fmt.Fprintf(&b, `<g data-entity="prim-loop:%d" transform="translate(%.2f %.2f)"><title>Edge #%d: %s–%s, weight %d</title><rect width="176" height="36" rx="8" fill="#223744" stroke="#7e929e"/><path d="M 14 25 C 2 5 36 5 24 25" fill="none" stroke="#7e929e"/><text x="38" y="23">#%d · node %d: %d</text></g>`, i, x, y, i, html.EscapeString(e.From), html.EscapeString(e.To), e.Weight, i, ids[e.From], e.Weight)
			loop++
		}
	}
	if s.ActiveEdge >= 0 {
		label := fmt.Sprintf("Probe #%d: %d", s.ActiveEdge, s.CandidateWeight)
		if s.Discarded {
			label = "Reject internal edge"
		}
		// Match heap serial from prior snapshot by stored serial, not endpoint text.
		probeX := 24.0
		if s.Discarded {
			probeX = w - 200
		}
		chip(fmt.Sprintf("prim-probe:%d", s.ActiveSerial), fmt.Sprintf("prim-heap:%d", s.ActiveSerial), probeX, equationY, label)
	}
	if step.LineID == "L13" {
		chip(fmt.Sprintf("prim-sum:%d", s.ActiveSerial), fmt.Sprintf("prim-probe:%d", s.ActiveSerial), 24, equationY+52, fmt.Sprintf("%d + (%d) = %d", s.TotalWeight-s.CandidateWeight, s.CandidateWeight, s.TotalWeight))
	}
	if step.Action == trace.ActionDone && len(s.Nodes) == 0 {
		text(24, equationY+64, "Empty graph; tree is empty")
	} else if step.LineID == "L18" {
		result := "Partial start-component tree; disconnected"
		if s.Connected {
			result = "✓ MST spanning all vertices"
		}
		if mobile && !s.Connected {
			text(24, equationY+64, "Partial start-component tree")
			text(24, equationY+88, "Disconnected: not a full MST")
		} else {
			text(24, equationY+64, result)
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}

func primLessonBool(v bool) int {
	if v {
		return 1
	}
	return 0
}
func primLessonShort(v string, n int) string {
	r := []rune(v)
	if len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return v
}
func primLessonPositions(nodes []GraphNode, x, y, w, h float64) [][2]float64 {
	p := make([][2]float64, len(nodes))
	if len(nodes) == 0 {
		return p
	}
	minX, maxX, minY, maxY := nodes[0].X, nodes[0].X, nodes[0].Y, nodes[0].Y
	ring := false
	seen := map[[2]float64]bool{}
	for _, n := range nodes {
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			ring = true
			continue
		}
		key := [2]float64{n.X, n.Y}
		if seen[key] {
			ring = true
		}
		seen[key] = true
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
	}
	scale := math.Inf(1)
	if maxX > minX {
		scale = (w - 100) / (maxX - minX)
	}
	if maxY > minY {
		scale = math.Min(scale, (h-110)/(maxY-minY))
	}
	if math.IsInf(scale, 0) {
		scale = 1
	}
	for i, n := range nodes {
		if ring {
			a := 2*math.Pi*float64(i)/float64(len(nodes)) - math.Pi/2
			p[i] = [2]float64{x + w/2 + (w/2-50)*math.Cos(a), y + h/2 + (h/2-55)*math.Sin(a)}
		} else {
			p[i] = [2]float64{x + w/2 + (n.X-(minX+maxX)/2)*scale, y + h/2 + (n.Y-(minY+maxY)/2)*scale}
		}
	}
	return p
}
