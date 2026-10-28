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
	dense := len(s.Nodes) >= 3 || len(s.Edges) >= 2
	for i, e := range s.Edges {
		if e.From == e.To {
			for _, prior := range s.Edges[:i] {
				if prior.From == e.From && prior.To == e.To {
					dense = true
				}
			}
		}
	}
	rowHeight := 30
	if dense {
		rowHeight = 96
	}
	panelH := float64(56 + len(s.Nodes)*rowHeight)
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
	type edgeLabel struct {
		id   int
		x, y float64
	}
	var labels []edgeLabel
	// Draw active edge last so other paths cannot obscure it.
	edgeOrder := make([]int, 0, len(s.Edges))
	for i := range s.Edges {
		if i != s.ActiveEdge {
			edgeOrder = append(edgeOrder, i)
		}
	}
	if s.ActiveEdge >= 0 && s.ActiveEdge < len(s.Edges) {
		edgeOrder = append(edgeOrder, s.ActiveEdge)
	}
	for _, i := range edgeOrder {
		e := s.Edges[i]
		if dense && e.From == e.To {
			shown := i
			for j, other := range s.Edges {
				if other.From == e.From && other.To == e.To && (j == s.ActiveEdge || (shown != s.ActiveEdge && j < shown)) {
					shown = j
				}
			}
			if i != shown {
				continue
			}
		}
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
			ordinal := 0
			for _, prior := range s.Edges[:i] {
				if prior.From == e.From && prior.To == e.To {
					ordinal++
				}
			}
			direction := -1.0
			if a[1]-gy < 80 {
				direction = 1
			}
			spread, reach := math.Min(54, 42+float64(ordinal)*0.1), math.Min(70, 54+float64(ordinal)*0.2)
			spread = math.Min(spread, math.Min(a[0]-gx, gx+gw-a[0]))
			if direction < 0 {
				reach = math.Min(reach, a[1]-gy)
			} else {
				reach = math.Min(reach, gy+gh-a[1])
			}
			path = fmt.Sprintf("M %.2f %.2f C %.2f %.2f %.2f %.2f %.2f %.2f", a[0]-16, a[1]+direction*16, a[0]-spread, a[1]+direction*reach, a[0]+spread, a[1]+direction*reach, a[0]+16, a[1]+direction*16)
			mx = a[0]
			my = a[1] + direction*(4+0.75*reach)
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
			nx, ny := -dy/length, dx/length
			if index[e.From] > index[e.To] {
				nx = -nx
				ny = -ny
			}
			// ponytail: compress parallel curves to canvas; use an expanded graph layout if individual curves must remain distinguishable at high counts.
			limit := math.Inf(1)
			if nx != 0 {
				limit = math.Min(limit, math.Min(mx-gx, gx+gw-mx)/math.Abs(nx))
			}
			if ny != 0 {
				limit = math.Min(limit, math.Min(my-gy, gy+gh-my)/math.Abs(ny))
			}
			spacing := 36.0
			if count > 1 {
				spacing = math.Min(spacing, 2*limit/float64(count-1))
			}
			offset := (float64(ordinal) - float64(count-1)/2) * spacing
			cx, cy := mx+nx*offset, my+ny*offset
			path = fmt.Sprintf("M %.2f %.2f Q %.2f %.2f %.2f %.2f", a[0]+dx/length*22, a[1]+dy/length*22, cx, cy, z[0]-dx/length*22, z[1]-dy/length*22)
			mx = (mx + cx) / 2
			my = (my + cy) / 2
		}
		fmt.Fprintf(&b, `<path data-edge-index="%d" d="%s" fill="none" stroke="%s" stroke-width="3"%s><title>%s</title></path>`, i, path, color, dash, html.EscapeString(fmt.Sprintf("Edge #%d %s — %s, weight %d: %s", i, e.From, e.To, e.Weight, status)))
		labels = append(labels, edgeLabel{i, mx, my})
	}
	// Place IDs in distinct grid cells, with leader lines to their actual paths.
	for i := range labels {
		if labels[i].id == s.ActiveEdge {
			labels[0], labels[i] = labels[i], labels[0]
			break
		}
	}
	var occupied [][2]float64
	omitted := 0
	for _, label := range labels {
		best, distance := [2]float64{}, math.Inf(1)
		for y := gy + 18; y <= gy+gh-18; y += 28 {
			for x := gx + 24; x <= gx+gw-24; x += 48 {
				free := true
				for _, p := range pos {
					if math.Abs(x-p[0]) < 44 && math.Abs(y-p[1]) < 36 {
						free = false
						break
					}
				}
				for _, p := range occupied {
					if x == p[0] && y == p[1] {
						free = false
						break
					}
				}
				d := math.Hypot(x-label.x, y-label.y)
				if free && d < distance {
					best, distance = [2]float64{x, y}, d
				}
			}
		}
		if math.IsInf(distance, 1) && label.id == s.ActiveEdge {
			// Dense node layouts may occupy every slot; keep active identity in the heading.
			best, distance = [2]float64{gx + gw - 24, gy - 18}, 0
		}
		if math.IsInf(distance, 1) {
			omitted++
			continue
		}
		occupied = append(occupied, best)
		fmt.Fprintf(&b, `<g data-edge-label="%d"><path d="M %.2f %.2f L %.2f %.2f" fill="none" stroke="#eff5ed" stroke-width="1"/><rect x="%.2f" y="%.2f" width="44" height="22" rx="6" fill="#111d27" stroke="#80938d"/><text x="%.2f" y="%.2f" text-anchor="middle" fill="#eff5ed" font-size="14">#%d</text></g>`, label.id, label.x, label.y, best[0], best[1], best[0]-22, best[1]-11, best[0], best[1]+5, label.id)
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
		for line := 0; !dense && line < 2 && line*12 < len(label); line++ {
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
		y := py + 50 + float64(i*rowHeight)
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
		if dense {
			label := []rune(n.Label)
			if len(label) == 0 {
				label = []rune(n.ID)
			}
			for line := 0; line*18 < len(label); line++ {
				kruskalText(&b, 0, 16+float64(line*16), string(label[line*18:min((line+1)*18, len(label))]))
			}
		}
		b.WriteString(`</g>`)
		if j, ok := index[parent]; ok {
			pointerY := py + 46 + float64(j*rowHeight)
			pointerX := px + pw - 58 + float64(i%4)*9
			fmt.Fprintf(&b, `<g data-entity="kruskal-parent-pointer:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><path d="M -4 -4 L 4 0 L -4 4 Z" fill="%s"/></g>`, i, pointerX, pointerY, pointerX, pointerY, html.EscapeString(fmt.Sprintf("Node #%d parent is #%d", i, j)), palette[i%5])
		}
	}
	// Explicit row brackets remain truthful even when a component has nonadjacent rows.
	for i, n := range s.Nodes {
		if g := component[n.ID]; g > 0 {
			y := py + 34 + float64(i*rowHeight)
			fmt.Fprintf(&b, `<path d="M %.0f %.0f h 5 v 24 h -5" fill="none" stroke="#8bd5ff"><title>Component G%d, node #%d</title></path>`, px+pw-12, y, g, i)
		}
	}
	queueTitle := "Edge queue · #original index : weight"
	if dense {
		queueTitle = "Edges · all listed; active loop shown"
	}
	if omitted > 0 {
		queueTitle = fmt.Sprintf("Edges · %d IDs in queue only", omitted)
	}
	kruskalText(&b, 24, stripY, queueTitle)
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
		if dense {
			e := s.Edges[edgeIndex]
			kruskalText(&b, 5, 15, fmt.Sprintf("#%d: #%d → #%d", edgeIndex, index[e.From], index[e.To]))
			kruskalText(&b, 5, 31, fmt.Sprintf("%d%s", e.Weight, marker))
		} else {
			kruskalText(&b, 5, 23, fmt.Sprintf("#%d:%d%s", edgeIndex, s.Edges[edgeIndex].Weight, marker))
		}
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
