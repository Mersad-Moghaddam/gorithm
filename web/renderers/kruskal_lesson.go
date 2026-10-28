package renderers

import (
	"fmt"
	"html"
	"math"
	"slices"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type KruskalLessonState struct {
	GraphState
	SortedEdgeIndices, SelectedEdges, RejectedEdges []int
	Cursor, ActiveEdge, TotalWeight                 int
	Parent                                          map[string]string
	Rank                                            map[string]int
	RootFrom, RootTo                                string
	Components                                      [][]string
	Cycle, Connected                                bool
}

func kruskalText(b *strings.Builder, x, y float64, text string) {
	fmt.Fprintf(b, `<text x="%.2f" y="%.2f" fill="#eff5ed" font-size="14">%s</text>`, x, y, html.EscapeString(text))
}

// RenderKruskalLessonSVG renders only the selected, settled DSU snapshot.
func RenderKruskalLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(KruskalLessonState)
	if !ok {
		return `<g data-lesson="kruskal" data-phase="unavailable" data-scene-width="360" data-scene-height="80"><title>Kruskal state unavailable</title></g>`
	}
	w := 960.0
	gx, gy, gw, gh := 24.0, 64.0, 592.0, 348.0
	px, py, pw := 640.0, 64.0, 296.0
	if width < 600 {
		w = 360
		gx = 16
		gy = 56
		gw = 328
		gh = 240
		px = 16
		py = 320
		pw = 328
	}
	panelH := float64(56 + len(s.Nodes)*30)
	stripY := math.Max(440, py+panelH+28)
	cols := int((w - 48) / 144)
	if cols < 1 {
		cols = 1
	}
	stripRows := (len(s.SortedEdgeIndices) + cols - 1) / cols
	resultY := stripY + 40 + float64(stripRows*48)
	selectedRows := (len(s.SelectedEdges) + cols - 1) / cols
	sceneH := math.Max(640, resultY+112+float64(selectedRows*48))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="kruskal" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif"><title>Kruskal: %s</title>`, html.EscapeString(step.LineID), w, sceneH, w, html.EscapeString(step.Explanation))
	kruskalText(&b, 24, 30, "Kruskal · edges and components")
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="16" fill="#111d27" stroke="#52634e"/>`, gx, gy, gw, gh)
	pos := map[string][2]float64{}
	index := map[string]int{}
	minX, maxX, minY, maxY := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	coincident := false
	for i, n := range s.Nodes {
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
		for j := 0; j < i; j++ {
			if n.X == s.Nodes[j].X && n.Y == s.Nodes[j].Y {
				coincident = true
			}
		}
	}
	scale := 1.0
	if maxX > minX || maxY > minY {
		sx, sy := math.Inf(1), math.Inf(1)
		if maxX > minX {
			sx = (gw - 112) / (maxX - minX)
		}
		if maxY > minY {
			sy = (gh - 112) / (maxY - minY)
		}
		scale = math.Min(sx, sy)
	}
	for i, n := range s.Nodes {
		x, y := gx+gw/2, gy+gh/2
		if coincident {
			a := 2*math.Pi*float64(i)/float64(len(s.Nodes)) - math.Pi/2
			x += (gw/2 - 48) * math.Cos(a)
			y += (gh/2 - 48) * math.Sin(a)
		} else {
			x += (n.X - (minX+maxX)/2) * scale
			y += (n.Y - (minY+maxY)/2) * scale
		}
		pos[n.ID] = [2]float64{x, y}
		index[n.ID] = i
	}
	for i, e := range s.Edges {
		a, okA := pos[e.From]
		z, okZ := pos[e.To]
		if !okA || !okZ {
			continue
		}
		color := "#80938d"
		status := "pending"
		dash := ""
		if slices.Contains(s.SelectedEdges, i) {
			color = "#9bf0bb"
			status = "✓ accepted"
		} else if slices.Contains(s.RejectedEdges, i) {
			color = "#ff9b9b"
			status = "× same component"
			dash = ` stroke-dasharray="5 5"`
		} else if s.ActiveEdge == i {
			color = "#ffe29a"
			status = "candidate"
		}
		dx, dy := z[0]-a[0], z[1]-a[1]
		length := math.Hypot(dx, dy)
		mx, my := (a[0]+z[0])/2, (a[1]+z[1])/2
		path := ""
		if e.From == e.To {
			path = fmt.Sprintf("M %.2f %.2f C %.2f %.2f %.2f %.2f %.2f %.2f", a[0]-16, a[1]-16, a[0]-58, a[1]-74, a[0]+58, a[1]-74, a[0]+16, a[1]-16)
			mx = a[0]
			my = a[1] - 56
		} else if length > 0 {
			// Parallel edge offsets depend on original identity, never current state.
			count, ordinal := 0, 0
			for j, other := range s.Edges {
				if (other.From == e.From && other.To == e.To) || (other.From == e.To && other.To == e.From) {
					if j < i {
						ordinal++
					}
					count++
				}
			}
			offset := (float64(ordinal) - float64(count-1)/2) * 36
			nx, ny := -dy/length, dx/length
			if index[e.From] > index[e.To] {
				nx = -nx
				ny = -ny
			}
			cx, cy := mx+nx*offset, my+ny*offset
			path = fmt.Sprintf("M %.2f %.2f Q %.2f %.2f %.2f %.2f", a[0]+dx/length*22, a[1]+dy/length*22, cx, cy, z[0]-dx/length*22, z[1]-dy/length*22)
			mx += nx * (offset/2 + 10)
			my += ny * (offset/2 + 10)
		}
		fmt.Fprintf(&b, `<path data-edge-index="%d" d="%s" fill="none" stroke="%s" stroke-width="3"%s><title>%s</title></path><rect x="%.2f" y="%.2f" width="52" height="22" rx="8" fill="#111d27"/>`, i, path, color, dash, html.EscapeString(fmt.Sprintf("Edge #%d %s — %s, weight %d: %s", i, e.From, e.To, e.Weight, status)), mx-26, my-16)
		kruskalText(&b, mx-20, my, fmt.Sprint(e.Weight))
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	component := map[string]int{}
	for i, members := range s.Components {
		for _, id := range members {
			component[id] = i + 1
		}
	}
	for i, n := range s.Nodes {
		p := pos[n.ID]
		outline := "#dce9e4"
		emphasis := ""
		if step.LineID == "L5" && (n.ID == s.RootFrom || n.ID == s.RootTo) {
			outline = "#ffe29a"
			emphasis = ` data-emphasis="compare"`
		}
		fmt.Fprintf(&b, `<g data-entity="kruskal-node:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>%s</title><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><text y="5" text-anchor="middle" fill="#ffffff" font-size="14">%d</text></g>`, i, p[0], p[1], p[0], p[1], emphasis, html.EscapeString(n.Label), palette[i%5], outline, i)
		label := []rune(n.Label)
		if len(label) == 0 {
			label = []rune(n.ID)
		}
		for line := 0; line < 2 && line*12 < len(label); line++ {
			end := min((line+1)*12, len(label))
			text := string(label[line*12 : end])
			if line == 1 && end < len(label) {
				text += "…"
			}
			kruskalText(&b, p[0]-36, p[1]+40+float64(line*16), text)
		}
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="12" fill="#192923" stroke="#52634e"/>`, px, py, pw, panelH)
	kruskalText(&b, px+12, py+20, "DSU · node / parent / rank / group")
	for i, n := range s.Nodes {
		y := py + 50 + float64(i*30)
		parent, known := s.Parent[n.ID]
		if !known {
			parent = "—"
		}
		fmt.Fprintf(&b, `<g data-entity="kruskal-parent:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title>`, i, px+12, y, px+12, y, html.EscapeString(fmt.Sprintf("%s / %s / rank %d / group %d", n.ID, parent, s.Rank[n.ID], component[n.ID])))
		parentIndex := "—"
		if j, ok := index[parent]; ok {
			parentIndex = fmt.Sprint(j)
		}
		kruskalText(&b, 0, 0, fmt.Sprintf("#%d    #%s    %d     G%d", i, parentIndex, s.Rank[n.ID], component[n.ID]))
		b.WriteString(`</g>`)
		if j, ok := index[parent]; ok {
			pointerY := py + 46 + float64(j*30)
			pointerX := px + pw - 58 + float64(i%4)*9
			fmt.Fprintf(&b, `<g data-entity="kruskal-parent-pointer:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><path d="M -4 -4 L 4 0 L -4 4 Z" fill="%s"/></g>`, i, pointerX, pointerY, pointerX, pointerY, html.EscapeString(fmt.Sprintf("Node #%d parent is #%d", i, j)), palette[i%5])
		}
	}
	// Explicit row brackets remain truthful even when a component has nonadjacent rows.
	for i, n := range s.Nodes {
		if g := component[n.ID]; g > 0 {
			y := py + 34 + float64(i*30)
			fmt.Fprintf(&b, `<path d="M %.0f %.0f h 5 v 24 h -5" fill="none" stroke="#8bd5ff"><title>Component G%d, node #%d</title></path>`, px+pw-12, y, g, i)
		}
	}
	kruskalText(&b, 24, stripY, "Edge queue · #original index : weight")
	for slot, edgeIndex := range s.SortedEdgeIndices {
		if edgeIndex < 0 || edgeIndex >= len(s.Edges) {
			continue
		}
		x, y := 24+float64(slot%cols)*144, stripY+18+float64(slot/cols)*48
		color := "#52634e"
		marker := ""
		if slot == s.Cursor {
			color = "#ffe29a"
			marker = " ▾"
		}
		if slices.Contains(s.RejectedEdges, edgeIndex) {
			color = "#ff9b9b"
			marker = " ×"
		}
		if slices.Contains(s.SelectedEdges, edgeIndex) {
			color = "#9bf0bb"
			marker = " ✓"
		}
		fmt.Fprintf(&b, `<g data-entity="kruskal-edge-chip:%d" data-x="%.0f" data-y="%.0f" transform="translate(%.0f %.0f)"><rect width="136" height="36" rx="8" fill="#192923" stroke="%s" stroke-width="2"/>`, edgeIndex, x, y, x, y, color)
		kruskalText(&b, 5, 23, fmt.Sprintf("#%d:%d%s", edgeIndex, s.Edges[edgeIndex].Weight, marker))
		b.WriteString(`</g>`)
	}
	status := "Choose edges joining different roots"
	switch step.LineID {
	case "L5":
		status = fmt.Sprintf("Roots: #%d and #%d", index[s.RootFrom], index[s.RootTo])
		if s.Cycle {
			status += " · × same component"
		}
	case "L6":
		status = "✓ Merged; acceptance follows"
	case "L7":
		status = "✓ Edge accepted; cost updated"
	case "L10":
		status = "✓ Minimum spanning forest"
		if s.Connected {
			status = "✓ Minimum spanning tree"
		}
	}
	kruskalText(&b, 24, resultY, status)
	kruskalText(&b, 24, resultY+26, fmt.Sprintf("Accepted %d edges · total cost %d", len(s.SelectedEdges), s.TotalWeight))
	for slot, id := range s.SelectedEdges {
		x, y := 24+float64(slot%cols)*144, resultY+42+float64(slot/cols)*48
		fmt.Fprintf(&b, `<g data-entity="kruskal-cost:%d" data-from-entity="kruskal-edge-chip:%d" data-x="%.0f" data-y="%.0f" transform="translate(%.0f %.0f)"><rect width="136" height="32" rx="8" fill="#192923" stroke="#9bf0bb"/>`, id, id, x, y, x, y)
		kruskalText(&b, 5, 22, fmt.Sprintf("#%d %+d", id, s.Edges[id].Weight))
		b.WriteString(`</g>`)
	}
	b.WriteString(`</g>`)
	return b.String()
}
