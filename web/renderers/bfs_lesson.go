package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"math"
	"strings"
)

type BFSLessonState struct {
	GraphState
	Start, Current, Neighbor string
	Queue, Order             []string
	Distance                 map[string]int
	Parent                   map[string]string
	Discovered, Finished     map[string]bool
	EdgeIndex                int
	Enqueued, Dequeued       string
	Skip                     bool
}

var bfsLessonColors = [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}

func RenderBFSLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(BFSLessonState)
	if !ok {
		return `<g data-lesson="bfs"><title>No BFS state</title></g>`
	}
	w, gx, gy, gw, gh, px, py, pw := 960.0, 24.0, 64.0, 592.0, 348.0, 640.0, 64.0, 296.0
	if width < 600 {
		w, gx, gy, gw, gh, px, py, pw = 360, 16, 56, 328, 240, 16, 320, 328
	}
	// Reserve full input capacity so FIFO geometry remains fixed across snapshots.
	cols := int((w - 48) / 104)
	if cols < 1 {
		cols = 1
	}
	rows := (len(s.Nodes) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	panelH := 28 + 30*float64(len(s.Nodes))
	stripY := math.Max(440, py+panelH+28)
	stripH := 34 + 42*float64(rows)
	orderY := stripY + stripH + 16
	eqY := orderY + stripH + 16
	sceneH := math.Max(640, eqY+116)
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="bfs" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif" font-size="16" fill="#eff5ed"><title>%s</title><rect width="%.0f" height="%.0f" rx="16" fill="#101b22"/><text x="24" y="30" font-size="19">BFS · FIFO distance layers</text><defs><marker id="bfs-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`, html.EscapeString(step.LineID), w, sceneH, w, html.EscapeString(step.Explanation), w, sceneH)
	positions := bfsLessonPositions(s.Nodes, gx, gy, gw, gh)
	indices := map[string]int{}
	for i, n := range s.Nodes {
		indices[n.ID] = i
	}
	for i, e := range s.Edges {
		a, aok := positions[e.From]
		z, zok := positions[e.To]
		if !aok || !zok {
			continue
		}
		color := "#71858c"
		if i == s.EdgeIndex {
			color = "#ffe29a"
		}
		dx, dy := z[0]-a[0], z[1]-a[1]
		length := math.Hypot(dx, dy)
		path := ""
		if length < 0.01 {
			offset := float64(i%5) * 8
			path = fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f", a[0]-16, a[1]-16, a[0]-52-offset, a[1]-72-offset, a[0]+52+offset, a[1]-72-offset, a[0]+16, a[1]-16)
		} else {
			offset := 0.0
			parallel := 0
			for j, p := range s.Edges {
				if j < i && p.From == e.From && p.To == e.To {
					parallel++
				}
				if p.From == e.To && p.To == e.From {
					offset = 18
				}
			}
			offset += float64(parallel) * 24
			cx, cy := (a[0]+z[0])/2-dy/length*offset, (a[1]+z[1])/2+dx/length*offset
			ux, uy := cx-a[0], cy-a[1]
			ul := math.Hypot(ux, uy)
			vx, vy := z[0]-cx, z[1]-cy
			vl := math.Hypot(vx, vy)
			path = fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f", a[0]+ux/ul*22, a[1]+uy/ul*22, cx, cy, z[0]-vx/vl*24, z[1]-vy/vl*24)
		}
		fmt.Fprintf(&b, `<path id="bfs-edge-%d" data-edge-index="%d" d="%s" stroke="%s" stroke-width="2.5" fill="none" marker-end="url(#bfs-arrow)"><title>%s → %s; forward traversal</title></path>`, i, i, path, color, html.EscapeString(e.From), html.EscapeString(e.To))
	}
	for i, n := range s.Nodes {
		p := positions[n.ID]
		ring := "#71858c"
		status := "unknown"
		if s.Discovered[n.ID] {
			status = "queued"
		}
		if n.ID == s.Current {
			status = "current"
			ring = "#8bd5ff"
		}
		if s.Finished[n.ID] {
			status = "✓ finished"
			ring = "#9bf0bb"
		}
		if n.ID == s.Neighbor && step.LineID == "L11" {
			ring = "#ffe29a"
		}
		fmt.Fprintf(&b, `<g data-entity="bfs-node:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s · %s</title><circle r="22" fill="%s" stroke="%s" stroke-width="3"/>`, i, p[0], p[1], p[0], p[1], html.EscapeString(n.Label), status, bfsLessonColors[i%5], ring)
		if n.ID == s.Neighbor && step.LineID == "L11" {
			b.WriteString(`<circle r="26" fill="none" stroke="#ffe29a" data-emphasis="compare"/>`)
		}
		label := n.Label
		if label == "" {
			label = n.ID
		}
		bfsLessonLabel(&b, 0, 42, fmt.Sprintf("%d: %s", i+1, label), 12)
		level := "?"
		if d, yes := s.Distance[n.ID]; yes {
			level = fmt.Sprint(d)
		}
		fmt.Fprintf(&b, `<rect x="12" y="-34" width="34" height="24" rx="8" fill="#101b22" stroke="%s"/><text x="29" y="-16" text-anchor="middle">%s</text></g>`, ring, level)
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="12" fill="#1d2d35"/><text x="%.0f" y="%.0f">Vertex · level · parent · status</text>`, px, py, pw, panelH, px+8, py+20)
	for i, n := range s.Nodes {
		level, parent := "?", "—"
		if d, yes := s.Distance[n.ID]; yes {
			level = fmt.Sprint(d)
		}
		if p, yes := s.Parent[n.ID]; yes {
			parent = p
		}
		status := "unknown"
		if s.Discovered[n.ID] {
			status = "queued"
		}
		if n.ID == s.Current {
			status = "current"
		}
		if s.Finished[n.ID] {
			status = "✓ finished"
		}
		text := fmt.Sprintf("%d: %s · %s · %s · %s", i+1, bfsLessonShort(n.ID, 4), level, bfsLessonShort(parent, 4), status)
		fmt.Fprintf(&b, `<text x="%.0f" y="%.0f"><title>%s · level %s · parent %s · %s</title>%s</text>`, px+8, py+48+30*float64(i), html.EscapeString(n.ID), level, html.EscapeString(parent), status, html.EscapeString(text))
	}
	for k, list := range [][]string{s.Queue, s.Order} {
		y, title, role := stripY, "FIFO · front → back", "bfs-queue"
		if k == 1 {
			y, title, role = orderY, "Visit order", "bfs-order"
		}
		fmt.Fprintf(&b, `<text x="24" y="%.0f">%s</text>`, y+18, title)
		for j, id := range list {
			i, exists := indices[id]
			if !exists {
				continue
			}
			x, cy := 24+float64(j%cols)*104, y+34+float64(j/cols)*42
			alias := ""
			if k == 0 && s.Enqueued == id {
				alias = fmt.Sprintf(` data-from-entity="bfs-node:%d"`, i)
			}
			fmt.Fprintf(&b, `<g data-entity="%s:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>%s</title><rect width="96" height="32" rx="8" fill="#1d2d35" stroke="%s"/><text x="48" y="22" text-anchor="middle">%s</text></g>`, role, i, x, cy, x, cy, alias, html.EscapeString(id), bfsLessonColors[i%5], html.EscapeString(fmt.Sprintf("%d: %s", i+1, bfsLessonShort(id, 7))))
		}
		if len(list) == 0 {
			fmt.Fprintf(&b, `<text x="24" y="%.0f" fill="#bbcab9">empty</text>`, y+56)
		}
	}
	if i, exists := indices[s.Dequeued]; exists && s.Dequeued != "" {
		p := positions[s.Dequeued]
		fmt.Fprintf(&b, `<g data-entity="bfs-current:%d" data-from-entity="bfs-queue:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><circle r="29" fill="none" stroke="#8bd5ff" stroke-width="3"/><title>Dequeued current vertex</title></g>`, i, i, p[0], p[1], p[0], p[1])
	}
	status := "Levels count edges, not weighted cost."
	if s.Skip {
		status = "Skip: queued/visited. No second enqueue."
	} else if s.Neighbor != "" {
		if d, yes := s.Distance[s.Neighbor]; yes {
			status = fmt.Sprintf("d[%s] = d[%s] + 1 = %d", s.Neighbor, s.Current, d)
		} else {
			status = fmt.Sprintf("%s: level unknown; test discovery first.", s.Neighbor)
		}
	}
	bfsLessonLabel(&b, w/2, eqY+24, status, int((w-48)/9))
	fmt.Fprintf(&b, `<text x="24" y="%.0f" fill="#bbcab9">visited in source = discovered</text></g>`, eqY+90)
	return b.String()
}

func bfsLessonShort(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}
func bfsLessonLabel(b *strings.Builder, x, y float64, label string, limit int) {
	r := []rune(label)
	lines := []string{label}
	if len(r) > limit {
		lines = []string{string(r[:limit]), bfsLessonShort(string(r[limit:]), limit)}
	}
	for i, line := range lines {
		fmt.Fprintf(b, `<text x="%.2f" y="%.2f" text-anchor="middle" fill="#eff5ed" stroke="#101b22" stroke-width="4" paint-order="stroke"><title>%s</title>%s</text>`, x, y+float64(i)*18, html.EscapeString(label), html.EscapeString(line))
	}
}
func bfsLessonPositions(nodes []GraphNode, x, y, w, h float64) map[string][2]float64 {
	result := map[string][2]float64{}
	if len(nodes) == 0 {
		return result
	}
	minX, maxX, minY, maxY := nodes[0].X, nodes[0].X, nodes[0].Y, nodes[0].Y
	ring := false
	for i, n := range nodes {
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			ring = true
		}
		for _, p := range nodes[:i] {
			if n.X == p.X && n.Y == p.Y {
				ring = true
			}
		}
	}
	scale := math.Inf(1)
	if maxX > minX {
		scale = (w - 100) / (maxX - minX)
	}
	if maxY > minY {
		scale = math.Min(scale, (h-140)/(maxY-minY))
	}
	if math.IsInf(scale, 0) {
		scale = 1
	}
	for i, n := range nodes {
		px, py := x+w/2+(n.X-(minX+maxX)/2)*scale, y+h/2+(n.Y-(minY+maxY)/2)*scale
		if ring {
			angle := 2*math.Pi*float64(i)/float64(len(nodes)) - math.Pi/2
			px = x + w/2 + (w/2-55)*math.Cos(angle)
			py = y + h/2 + (h/2-65)*math.Sin(angle)
		}
		result[n.ID] = [2]float64{px, py}
	}
	return result
}
