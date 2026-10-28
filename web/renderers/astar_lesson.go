package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type AStarOpenEntry struct {
	Node               string
	G, H, F, HeapIndex int
}
type AStarLessonState struct {
	GraphState
	Start, Goal, Current, Neighbor                             string
	G, H, F                                                    map[string]int
	CameFrom                                                   map[string]string
	Closed                                                     map[string]bool
	Open                                                       []AStarOpenEntry
	Path                                                       []string
	EdgeUndirected                                             []bool
	EdgeIndex, OldG, TentativeG, Cost                          int
	KnownOldG, HasTentative, Improved, Found, InvalidEndpoints bool
	HeuristicConsistent                                        bool
}

// RenderAStarLessonSVG uses only this snapshot. Display fitting never changes h.
func RenderAStarLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(AStarLessonState)
	if !ok {
		return `<g data-lesson="astar" data-phase="missing" data-scene-width="360" data-scene-height="120"><title>A* state unavailable</title></g>`
	}
	w := 960.0
	gx, gy, gw, gh := 24.0, 64.0, 592.0, 348.0
	lx, ly, lw := 640.0, 64.0, 296.0
	if width < 600 {
		w = 360
		gx = 16
		gy = 56
		gw = 328
		gh = 240
		lx = 16
		ly = 320
		lw = 328
	}
	ledgerH := 28.0 + float64(len(s.Nodes))*30
	heapY := math.Max(gy+gh, ly+ledgerH) + 36
	chips := int((w - 48) / 120)
	if chips < 1 {
		chips = 1
	}
	heapRows := (len(s.Open) + chips - 1) / chips
	if heapRows < 1 {
		heapRows = 1
	}
	equationY := heapY + float64(heapRows)*74 + 24
	sceneH := equationY + 180
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="astar" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(step.LineID), w, sceneH, w, html.EscapeString("A* pathfinding. "+step.Explanation))
	fmt.Fprintf(&b, `<rect width="%.0f" height="%.0f" rx="16" fill="#192923"/><defs><marker id="astar-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`, w, sceneH)
	astarText(&b, 24, 30, "A*: f = g + h", "#eff5ed")
	positions := astarPositions(s.Nodes, gx, gy, gw, gh)
	index := map[string]int{}
	for i, n := range s.Nodes {
		index[n.ID] = i
	}
	for i, e := range s.Edges {
		a, aok := positions[e.From]
		z, zok := positions[e.To]
		if !aok || !zok {
			continue
		}
		color := "#80958d"
		if e.Active {
			color = "#8bd5ff"
		}
		marker := ""
		if e.Directed {
			marker = ` marker-end="url(#astar-arrow)"`
		}
		dx, dy := z[0]-a[0], z[1]-a[1]
		dist := math.Hypot(dx, dy)
		mx, my := (a[0]+z[0])/2, (a[1]+z[1])/2
		path := ""
		if dist < 1 {
			path = fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f", a[0]-15, a[1]-16, a[0]-50, a[1]-80, a[0]+50, a[1]-80, a[0]+15, a[1]-16)
			mx = a[0]
			my = a[1] - 65
		} else {
			offset := 0.0
			parallel := 0
			for j := 0; j < i; j++ {
				other := s.Edges[j]
				if other.From == e.From && other.To == e.To {
					parallel++
				}
			}
			for j, other := range s.Edges {
				if j != i && other.From == e.To && other.To == e.From {
					offset = 18
					break
				}
			}
			offset += float64(parallel) * 24
			nx, ny := -dy/dist, dx/dist
			cx, cy := mx+nx*offset, my+ny*offset
			sx, sy := a[0]+dx/dist*22, a[1]+dy/dist*22
			ex, ey := z[0]-dx/dist*22, z[1]-dy/dist*22
			path = fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f", sx, sy, cx, cy, ex, ey)
			mx = (sx+2*cx+ex)/4 + nx*10
			my = (sy+2*cy+ey)/4 + ny*10
		}
		fmt.Fprintf(&b, `<path data-edge="astar-edge:%d" d="%s" fill="none" stroke="%s" stroke-width="2"%s/><rect x="%.2f" y="%.2f" width="52" height="22" rx="9" fill="#192923"/>`, i, path, color, marker, mx-26, my-16)
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" text-anchor="middle">%d</text>`, mx, my, e.Weight)
	}
	if a, ok := positions[s.Current]; ok {
		if z, ok := positions[s.Goal]; ok && s.Current != s.Goal {
			fmt.Fprintf(&b, `<path d="M%.2f %.2fL%.2f %.2f" stroke="#ffe29a" stroke-dasharray="5 5" fill="none"/>`, a[0], a[1], z[0], z[1])
			astarText(&b, gx+8, gy+20, fmt.Sprintf("estimate h = %d (not an edge)", s.H[s.Current]), "#ffe29a")
		}
	}
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, n := range s.Nodes {
		p := positions[n.ID]
		outline := "#bbcab9"
		if n.ID == s.Current || n.ID == s.Neighbor {
			outline = "#ffe29a"
		}
		if n.State == "path" {
			outline = "#9bf0bb"
		}
		fmt.Fprintf(&b, `<g data-entity="astar-node:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><circle r="22" fill="%s" stroke="%s" stroke-width="3"/>`, i, p[0], p[1], p[0], p[1], html.EscapeString(n.Label), colors[i%5], outline)
		if n.ID == s.Goal {
			b.WriteString(`<circle r="28" fill="none" stroke="#eff5ed" stroke-width="2"/>`)
		}
		fmt.Fprintf(&b, `<text y="5" text-anchor="middle" fill="#fff">%d</text>`, i+1)
		label := n.Label
		if label == "" {
			label = n.ID
		}
		astarLabel(&b, 0, 44, label, 16)
		b.WriteString(`</g>`)
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="12" fill="#22372c"/>`, lx, ly, lw, ledgerH)
	astarText(&b, lx+8, ly+20, "Node       g       h       f    status", "#bbcab9")
	for i, n := range s.Nodes {
		y := ly + 50 + float64(i)*30
		status := "unknown"
		if _, ok := s.G[n.ID]; ok {
			status = "tentative"
		}
		if s.Closed[n.ID] {
			status = "closed"
		}
		if n.ID == s.Current {
			status = "current"
		}
		if n.State == "path" {
			status = "✓ path"
		}
		if step.LineID == "L19" {
			if _, ok := s.G[n.ID]; !ok {
				status = "unreachable"
			}
		}
		astarText(&b, lx+8, y, fmt.Sprintf("%d", i+1), colors[i%5])
		astarText(&b, lx+46, y, astarScore(s.G, n.ID), "#eff5ed")
		astarText(&b, lx+99, y, astarScore(s.H, n.ID), "#eff5ed")
		astarText(&b, lx+152, y, astarScore(s.F, n.ID), "#eff5ed")
		astarText(&b, lx+210, y, status, "#eff5ed")
	}
	astarText(&b, 24, heapY-10, "Open heap · actual slots (f, then h)", "#bbcab9")
	for i, e := range s.Open {
		x := 24 + float64(i%chips)*120
		y := heapY + float64(i/chips)*74
		id := index[e.Node]
		fmt.Fprintf(&b, `<g data-entity="astar-open:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><rect width="112" height="64" rx="12" fill="#22372c" stroke="%s"/>`, id, x, y, x, y, html.EscapeString(e.Node), colors[id%5])
		astarText(&b, 8, 20, fmt.Sprintf("#%d · node %d", e.HeapIndex, id+1), "#bbcab9")
		astarText(&b, 8, 39, fmt.Sprintf("f = %d", e.F), "#eff5ed")
		astarText(&b, 8, 57, fmt.Sprintf("%d + %d", e.G, e.H), "#eff5ed")
		b.WriteString(`</g>`)
	}
	if len(s.Open) == 0 {
		astarText(&b, 24, heapY+24, "Empty heap", "#bbcab9")
	}
	if p, ok := positions[s.Current]; ok {
		alias := ""
		if step.LineID == "L5" {
			alias = fmt.Sprintf(` data-from-entity="astar-open:%d"`, index[s.Current])
		}
		fmt.Fprintf(&b, `<g data-entity="astar-current:%d"%s data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><circle r="33" fill="none" stroke="#ffe29a" stroke-width="2"/><title>Current node</title></g>`, index[s.Current], alias, p[0], p[1], p[0], p[1])
	}
	y := equationY
	if s.HasTentative {
		old := "∞"
		if s.KnownOldG {
			old = fmt.Sprint(s.OldG)
		}
		astarText(&b, 24, y, fmt.Sprintf("tentative g = %d; previous g = %s", s.TentativeG, old), "#ffe29a")
		y += 24
		astarText(&b, 24, y, fmt.Sprintf("f = g %d + h %d = %d", s.TentativeG, s.H[s.Neighbor], s.TentativeG+s.H[s.Neighbor]), "#8bd5ff")
		y += 24
		if step.LineID == "L13" || step.LineID == "L14" {
			for j, value := range []int{s.TentativeG, s.H[s.Neighbor]} {
				x := 24 + float64(j)*110
				if step.LineID == "L14" {
					x = 220 + float64(j)*70
				}
				fmt.Fprintf(&b, `<g data-entity="astar-score:%d" data-x="%.0f" data-y="%.0f" transform="translate(%.0f %.0f)"><title>g/h token joins f</title><text>%d</text></g>`, j, x, y, x, y, value)
			}
			y += 24
		}
	} else if s.Closed[s.Neighbor] {
		astarText(&b, 24, y, "Closed neighbor: skip, no relaxation.", "#ffe29a")
		y += 24
	}
	if s.Found {
		astarText(&b, 24, y, fmt.Sprintf("✓ Path found · cost %d", s.Cost), "#9bf0bb")
		y += 24
	} else if s.InvalidEndpoints {
		astarText(&b, 24, y, "Invalid endpoints · cost -1", "#ff9b9b")
		y += 24
	} else if step.LineID == "L19" {
		astarText(&b, 24, y, "No path · cost -1", "#ff9b9b")
		y += 24
	}
	if s.HeuristicConsistent {
		astarText(&b, 24, y, "Consistent h: closed-set guarantee holds.", "#9bf0bb")
	} else {
		astarText(&b, 24, y, "Inconsistent h: not guaranteed shortest.", "#ff9b9b")
	}
	y += 24
	astarText(&b, 24, y, "h uses original coordinates, not display positions.", "#bbcab9")
	b.WriteString(`</g>`)
	return b.String()
}

func astarScore(values map[string]int, id string) string {
	if v, ok := values[id]; ok {
		return fmt.Sprint(v)
	}
	return "∞"
}
func astarText(b *strings.Builder, x, y float64, text, color string) {
	fmt.Fprintf(b, `<text x="%.2f" y="%.2f" fill="%s">%s</text>`, x, y, color, html.EscapeString(text))
}
func astarLabel(b *strings.Builder, x, y float64, label string, limit int) {
	runes := []rune(label)
	if len(runes) > 2*limit {
		runes = append(runes[:2*limit-1], '…')
	}
	for i := 0; i < len(runes); i += limit {
		end := i + limit
		if end > len(runes) {
			end = len(runes)
		}
		fmt.Fprintf(b, `<text x="%.2f" y="%.2f" text-anchor="middle" fill="#eff5ed">%s</text>`, x, y+float64(i/limit)*16, html.EscapeString(string(runes[i:end])))
	}
}
func astarPositions(nodes []GraphNode, x, y, w, h float64) map[string][2]float64 {
	out := map[string][2]float64{}
	if len(nodes) == 0 {
		return out
	}
	minX, maxX, minY, maxY := nodes[0].X, nodes[0].X, nodes[0].Y, nodes[0].Y
	coincident := false
	seen := map[[2]float64]bool{}
	for _, n := range nodes {
		p := [2]float64{n.X, n.Y}
		if seen[p] || math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			coincident = true
		}
		seen[p] = true
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
	}
	if coincident {
		for i, n := range nodes {
			angle := 2*math.Pi*float64(i)/float64(len(nodes)) - math.Pi/2
			out[n.ID] = [2]float64{x + w/2 + (w/2-48)*math.Cos(angle), y + h/2 + (h/2-64)*math.Sin(angle)}
		}
		return out
	}
	scale := math.Inf(1)
	if maxX > minX {
		scale = (w - 96) / (maxX - minX)
	}
	if maxY > minY {
		scale = math.Min(scale, (h-128)/(maxY-minY))
	}
	if math.IsInf(scale, 0) {
		scale = 1
	}
	for _, n := range nodes {
		out[n.ID] = [2]float64{x + w/2 + (n.X-(minX+maxX)/2)*scale, y + h/2 + (n.Y-(minY+maxY)/2)*scale}
	}
	return out
}
