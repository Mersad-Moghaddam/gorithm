package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"math"
	"strings"
)

type ToposortLessonState struct {
	GraphState
	InDegree                            map[string]int
	Queue, Order, Blocked               []string
	Processed                           map[string]bool
	ResolvedEdges                       []int
	Current, Neighbor                   string
	EdgeIndex, OldInDegree, NewInDegree int
	Decremented, Ready, Valid, Terminal bool
}

// RenderToposortLessonSVG renders stored prerequisites, never recomputes a sort.
func RenderToposortLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(ToposortLessonState)
	if !ok {
		return `<g data-lesson="toposort"><title>No prerequisite snapshot</title></g>`
	}
	w := 960
	graphY, graphH := 88.0, 220.0
	laneY := 344
	laneW := 432
	orderX := 480
	ledgerY := 448
	cols := laneW / 104
	if cols < 1 {
		cols = 1
	}
	rows := func(n int) int { return max(1, (n+cols-1)/cols) }
	queueH := 40 + rows(len(s.Queue))*52
	orderH := 40 + rows(len(s.Order))*52
	orderY := laneY
	ledgerY = max(448, laneY+max(queueH, orderH)+24)
	sceneH := ledgerY + 58 + len(s.Nodes)*30 + 100
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="toposort" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Topological sort: ready tasks and remaining prerequisites</title><defs><marker id="toposort-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`, html.EscapeString(step.LineID), w, sceneH, w)
	text := func(x, y float64, v string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f">%s</text>`, x, y, html.EscapeString(v))
	}
	text(24, 30, "Topological sort · prerequisites")
	pos := make(map[string][2]float64, len(s.Nodes))
	index := map[string]int{}
	minX, maxX, minY, maxY := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	finite := true
	seen := map[[2]float64]bool{}
	coincident := false
	for _, n := range s.Nodes {
		if math.IsNaN(n.X) || math.IsInf(n.X, 0) || math.IsNaN(n.Y) || math.IsInf(n.Y, 0) {
			finite = false
			break
		}
		p := [2]float64{n.X, n.Y}
		if seen[p] {
			coincident = true
		}
		seen[p] = true
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
	}
	boxW := float64(w - 48)
	cx := float64(w) / 2
	cy := graphY + graphH/2
	scale := 1.0
	if maxX > minX {
		scale = (boxW - 110) / (maxX - minX)
	}
	if maxY > minY {
		scale = math.Min(scale, (graphH-74)/(maxY-minY))
	}
	for i, n := range s.Nodes {
		x, y := cx, cy
		if !finite || coincident {
			a := 2*math.Pi*float64(i)/float64(max(1, len(s.Nodes))) - math.Pi/2
			x += (boxW/2 - 56) * math.Cos(a)
			y += (graphH/2 - 40) * math.Sin(a)
		} else {
			if maxX > minX {
				x += (n.X - (minX+maxX)/2) * scale
			}
			if maxY > minY {
				y += (n.Y - (minY+maxY)/2) * scale
			}
		}
		pos[n.ID] = [2]float64{x, y}
		index[n.ID] = i
	}
	resolved := map[int]bool{}
	for _, i := range s.ResolvedEdges {
		resolved[i] = true
	}
	edgeMid := map[int][2]float64{}
	for i, e := range s.Edges {
		p, a := pos[e.From]
		q, c := pos[e.To]
		if !a || !c {
			continue
		}
		color := "#bbcab9"
		dash := ""
		if resolved[i] {
			color = "#789486"
			dash = ` stroke-dasharray="6 5"`
		}
		if i == s.EdgeIndex {
			color = "#ffe29a"
		}
		path := ""
		mid := [2]float64{(p[0] + q[0]) / 2, (p[1] + q[1]) / 2}
		if e.From == e.To {
			path = fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f", p[0]-24, p[1]-22, p[0]-64, p[1]-80, p[0]+64, p[1]-80, p[0]+24, p[1]-22)
			mid = [2]float64{p[0], p[1] - 66}
		} else {
			dx, dy := q[0]-p[0], q[1]-p[1]
			dist := math.Hypot(dx, dy)
			if dist == 0 {
				continue
			}
			ux, uy := dx/dist, dy/dist
			trim := func(x, y float64) float64 {
				return math.Min(46/math.Max(math.Abs(x), 0.000001), 22/math.Max(math.Abs(y), 0.000001))
			}
			t := trim(ux, uy)
			sx, sy, tx, ty := p[0]+ux*t, p[1]+uy*t, q[0]-ux*t, q[1]-uy*t
			parallel := 0
			reciprocal := false
			for j, f := range s.Edges {
				if j < i && f.From == e.From && f.To == e.To {
					parallel++
				}
				if f.From == e.To && f.To == e.From {
					reciprocal = true
				}
			}
			offset := float64(parallel) * 24
			if reciprocal {
				offset += 18
			}
			mx, my := (sx+tx)/2-uy*offset, (sy+ty)/2+ux*offset
			path = fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f", sx, sy, mx, my, tx, ty)
			mid = [2]float64{(sx + 2*mx + tx) / 4, (sy + 2*my + ty) / 4}
		}
		edgeMid[i] = mid
		fmt.Fprintf(&b, `<path id="toposort-edge-%d" d="%s" fill="none" stroke="%s" stroke-width="2"%s marker-end="url(#toposort-arrow)"><title>Dependency %d: %s to %s%s</title></path>`, i, path, color, dash, i, html.EscapeString(e.From), html.EscapeString(e.To), map[bool]string{true: " resolved", false: " unresolved"}[resolved[i]])
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	entity := func(role string, i int, x, y float64, alias string) {
		fmt.Fprintf(&b, `<g data-entity="toposort:%s:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"`, role, i, x, y, x, y)
		if alias != "" {
			fmt.Fprintf(&b, ` data-from-entity="%s"`, alias)
		}
		b.WriteString(`>`)
	}
	label := func(v string) {
		r := []rune(v)
		if len(r) > 20 {
			r = append(r[:19], '…')
		}
		if len(r) > 10 {
			fmt.Fprintf(&b, `<text text-anchor="middle" y="-3">%s</text><text text-anchor="middle" y="13">%s</text>`, html.EscapeString(string(r[:10])), html.EscapeString(string(r[10:])))
		} else {
			fmt.Fprintf(&b, `<text text-anchor="middle" y="5">%s</text>`, html.EscapeString(string(r)))
		}
	}
	for i, n := range s.Nodes {
		p := pos[n.ID]
		entity("node", i, p[0], p[1], "")
		stroke := "#bbcab9"
		if s.Processed[n.ID] {
			stroke = "#9bf0bb"
		}
		if n.ID == s.Current {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&b, `<title>%s</title><rect x="-46" y="-22" width="92" height="44" rx="12" fill="#192923" stroke="%s" stroke-width="3"/><rect x="-45" y="-21" width="7" height="42" rx="3" fill="%s"/>`, html.EscapeString(n.Label), stroke, palette[i%5])
		v := n.Label
		if v == "" {
			v = n.ID
		}
		label(v)
		text(36, -24, fmt.Sprint(s.InDegree[n.ID]))
		b.WriteString(`</g>`)
		if s.Decremented && s.Neighbor == n.ID {
			m := edgeMid[s.EdgeIndex]
			entity("token", i, m[0], m[1], "")
			text(0, 0, "−1")
		} else {
			entity("token", i, p[0]+38, p[1]-34, "")
			text(0, 0, "·")
		}
		b.WriteString(`</g>`)
	}
	lane := func(title, role string, ids []string, x, y, h int) {
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" rx="14" fill="#192923" stroke="#52634e"/>`, x, y, laneW, h)
		text(float64(x+12), float64(y+24), title)
		for j, id := range ids {
			i, exists := index[id]
			if !exists {
				continue
			}
			px, py := float64(x+58+(j%cols)*104), float64(y+60+(j/cols)*52)
			alias := ""
			if role == "order" && step.LineID == "L8" && id == s.Current {
				alias = fmt.Sprintf("toposort:current:%d", i)
			}
			if role == "queue" && s.Ready && id == s.Neighbor {
				alias = fmt.Sprintf("toposort:node:%d", i)
			}
			entity(role, i, px, py, alias)
			fmt.Fprintf(&b, `<title>%s</title><rect x="-46" y="-20" width="92" height="40" rx="10" fill="#192923" stroke="%s"/>`, html.EscapeString(id), palette[i%5])
			label(id)
			b.WriteString(`</g>`)
		}
	}
	x := 24
	lane("Ready FIFO", "queue", s.Queue, x, laneY, queueH)
	lane("Output order", "order", s.Order, orderX, orderY, orderH)
	if i, exists := index[s.Current]; exists {
		p := pos[s.Current]
		alias := ""
		if step.LineID == "L7" {
			alias = fmt.Sprintf("toposort:queue:%d", i)
		}
		entity("current", i, p[0], graphY+graphH+20, alias)
		r := []rune(s.Current)
		if len(r) > 12 {
			r = append(r[:11], '…')
		}
		fmt.Fprintf(&b, `<title>Current: %s</title>`, html.EscapeString(s.Current))
		text(-44, 0, "Current: "+string(r))
		b.WriteString(`</g>`)
	}
	text(24, float64(ledgerY), "Remaining prerequisites")
	text(24, float64(ledgerY+28), "Vertex")
	text(float64(w/2), float64(ledgerY+28), "Count / status")
	for i, n := range s.Nodes {
		y := float64(ledgerY + 58 + i*30)
		v := n.Label
		if v == "" {
			v = n.ID
		}
		r := []rune(v)
		if len(r) > 18 {
			v = string(r[:17]) + "…"
		}
		fmt.Fprintf(&b, `<g><title>%s</title>`, html.EscapeString(n.Label+" ("+n.ID+")"))
		text(24, y, v)
		b.WriteString(`</g>`)
		status := "waiting"
		if s.Processed[n.ID] {
			status = "✓ ordered"
		} else if s.Terminal {
			status = "blocked"
		} else {
			for _, id := range s.Queue {
				if id == n.ID {
					status = "ready"
				}
			}
		}
		text(float64(w/2), y, fmt.Sprintf("%d · %s", s.InDegree[n.ID], status))
	}
	y := float64(sceneH - 48)
	if s.Decremented {
		text(24, y, fmt.Sprintf("%d − 1 = %d", s.OldInDegree, s.NewInDegree))
	} else if s.Terminal {
		if s.Valid {
			text(24, y, "✓ Valid topological order")
		} else {
			text(24, y, "Cycle prevents completion.")
			text(24, y+22, "Blocked ≠ proven cycle member.")
		}
	} else {
		text(24, y, "Resolve each incoming dependency once.")
	}
	b.WriteString(`</g>`)
	return b.String()
}
