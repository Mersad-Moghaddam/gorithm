package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type DFSCallFrame struct {
	Node, Caller string
	NextNeighbor int
}
type DFSLessonState struct {
	GraphState
	Start, Current, Neighbor, ReturningFrom string
	Stack                                   []DFSCallFrame
	Order, FinishOrder                      []string
	Parent                                  map[string]string
	Discovered, Finished                    map[string]bool
	EdgeIndex                               int
	BackEdge, Skip                          bool
}

func dfsLessonText(b *strings.Builder, x, y float64, text string) {
	fmt.Fprintf(b, `<text x="%.2f" y="%.2f" fill="#eff5ed" font-size="14">%s</text>`, x, y, html.EscapeString(text))
}

// RenderDFSLessonSVG renders only the selected immutable recursion snapshot.
func RenderDFSLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(DFSLessonState)
	if !ok {
		return `<g data-lesson="dfs"><title>No DFS state</title></g>`
	}
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	w := 960.0
	gx, gy, gw, gh := 24.0, 64.0, 592.0, 348.0
	px, py, pw := 640.0, 64.0, 296.0
	panelH := math.Max(348, 44*float64(len(s.Stack))+64)
	if width < 600 {
		w = 360
		gx, gy, gw, gh = 16, 56, 328, 240
		px, py, pw = 16, 320, 328
		panelH = math.Max(100, 44*float64(len(s.Stack))+64)
	}
	stripY := math.Max(440, py+panelH+28)
	chipW := func(id string) float64 { return math.Max(48, float64(utf8.RuneCountInString(id))*8+20) }
	stripHeight := func(list []string) float64 {
		x, rows := 24.0, 1.0
		for _, id := range list {
			cw := math.Min(w-48, chipW(id))
			if x+cw > w-24 {
				rows++
				x = 24
			}
			x += cw + 8
		}
		return rows*56 + 48
	}
	finishY := stripY + stripHeight(s.Order) + 16
	statusY := finishY + stripHeight(s.FinishOrder) + 20
	sceneH := statusY + 80
	phase := map[string]string{"L1": "init", "L2": "init", "L3": "init", "L14": "root-call", "L6": "discover", "L7": "preorder", "L8": "neighbors", "L9": "inspect", "L10": "call", "L11": "return", "L12": "finish", "L15": "done"}[step.LineID]
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="dfs" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif"><title>%s</title><rect width="%.0f" height="%.0f" fill="#101923"/>`, phase, w, sceneH, w, html.EscapeString("DFS: "+step.Explanation), w, sceneH)
	dfsLessonText(&b, 24, 30, "DFS · newest unfinished call first")
	fmt.Fprintf(&b, `<defs><marker id="dfs-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`)
	pos := map[string][2]float64{}
	idx := map[string]int{}
	minX, maxX, minY, maxY := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	coincident := false
	for i, n := range s.Nodes {
		idx[n.ID] = i
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			coincident = true
			continue
		}
		minX = math.Min(minX, n.X)
		maxX = math.Max(maxX, n.X)
		minY = math.Min(minY, n.Y)
		maxY = math.Max(maxY, n.Y)
		for j := 0; j < i; j++ {
			if s.Nodes[j].X == n.X && s.Nodes[j].Y == n.Y {
				coincident = true
			}
		}
	}
	scale := 1.0
	if maxX > minX {
		scale = (gw - 96) / (maxX - minX)
	}
	if maxY > minY {
		sy := (gh - 96) / (maxY - minY)
		if maxX == minX || sy < scale {
			scale = sy
		}
	}
	for i, n := range s.Nodes {
		x, y := gx+gw/2, gy+gh/2
		if coincident {
			a := 2*math.Pi*float64(i)/float64(len(s.Nodes)) - math.Pi/2
			x += (gw/2 - 40) * math.Cos(a)
			y += (gh/2 - 40) * math.Sin(a)
		} else {
			if maxX > minX {
				x += (n.X - (minX+maxX)/2) * scale
			}
			if maxY > minY {
				y += (n.Y - (minY+maxY)/2) * scale
			}
		}
		pos[n.ID] = [2]float64{x, y}
	}
	for i, e := range s.Edges {
		a, oa := pos[e.From]
		z, oz := pos[e.To]
		if !oa || !oz {
			continue
		}
		stroke := "#7f919e"
		if e.Active {
			stroke = "#ffe29a"
		}
		path := ""
		if e.From == e.To {
			path = fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f", a[0]-16, a[1]-16, a[0]-52, a[1]-70, a[0]+52, a[1]-70, a[0]+16, a[1]-16)
		} else {
			dx, dy := z[0]-a[0], z[1]-a[1]
			d := math.Hypot(dx, dy)
			if d == 0 {
				continue
			}
			nx, ny := -dy/d, dx/d
			offset := 0.0
			parallel := 0
			reciprocal := false
			for j, other := range s.Edges {
				if j < i && other.From == e.From && other.To == e.To {
					parallel++
				}
				if other.From == e.To && other.To == e.From {
					reciprocal = true
				}
			}
			if reciprocal {
				offset = 18
			}
			offset += float64(parallel) * 24
			cx, cy := (a[0]+z[0])/2+nx*offset, (a[1]+z[1])/2+ny*offset
			da, dz := math.Hypot(cx-a[0], cy-a[1]), math.Hypot(cx-z[0], cy-z[1])
			path = fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f", a[0]+22*(cx-a[0])/da, a[1]+22*(cy-a[1])/da, cx, cy, z[0]+22*(cx-z[0])/dz, z[1]+22*(cy-z[1])/dz)
		}
		fmt.Fprintf(&b, `<path data-edge="dfs:edge:%d" d="%s" fill="none" stroke="%s" stroke-width="2" marker-end="url(#dfs-arrow)"/>`, i, path, stroke)
	}
	for i, n := range s.Nodes {
		p := pos[n.ID]
		outline := "#bbcab9"
		state := "untouched"
		if s.Discovered[n.ID] {
			outline = colors[i%5]
			state = "active"
		}
		if s.Finished[n.ID] {
			outline = "#9bf0bb"
			state = "✓ finished"
		}
		if n.ID == s.Neighbor {
			outline = "#ffe29a"
		}
		emphasis := ""
		if step.LineID == "L9" && n.ID == s.Neighbor {
			emphasis = ` data-emphasis="compare"`
		}
		fmt.Fprintf(&b, `<g data-entity="dfs:node:%d"%s data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><rect x="-38" y="-11" width="76" height="21" rx="4" fill="#101923"/>`, i, emphasis, p[0], p[1], p[0], p[1], html.EscapeString(n.Label+" ("+n.ID+"): "+state), colors[i%5], outline)
		label := n.Label
		if label == "" {
			label = n.ID
		}
		runes := []rune(label)
		if len(runes) > 9 {
			dfsLessonText(&b, -34, -2, string(runes[:9]))
			tail := runes[9:]
			if len(tail) > 8 {
				tail = append(tail[:7], '…')
			}
			dfsLessonText(&b, -34, 14, string(tail))
		} else {
			dfsLessonText(&b, -float64(len(runes))*4, 5, label)
		}
		if n.ID == s.Current {
			dfsLessonText(&b, -28, 40, "current")
		} else if s.Finished[n.ID] {
			dfsLessonText(&b, -32, 40, "✓ finished")
		}
		b.WriteString(`</g>`)
	}
	if p, exists := pos[s.Current]; exists {
		fmt.Fprintf(&b, `<g data-entity="dfs:cursor" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><path d="M-7 -37L7 -37L0 -27Z" fill="#8bd5ff"/></g>`, p[0], p[1], p[0], p[1])
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="16" fill="#192733" stroke="#647887"/>`, px, py, pw, panelH)
	dfsLessonText(&b, px+16, py+24, "Call stack · root at bottom")
	for i, f := range s.Stack {
		x, y := px+12, py+panelH-16-float64(i+1)*44
		alias := ""
		if step.LineID == "L10" && i == len(s.Stack)-1 {
			alias = fmt.Sprintf(` data-from-entity="dfs:node:%d"`, idx[f.Node])
		}
		fmt.Fprintf(&b, `<g data-entity="dfs:frame:%d"%s data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><rect width="%.0f" height="40" rx="8" fill="#101923" stroke="%s"/>`, idx[f.Node], alias, x, y, x, y, html.EscapeString(f.Node+" called by "+f.Caller), pw-24, colors[idx[f.Node]%5])
		name := []rune(f.Node)
		if len(name) > 13 {
			name = append(name[:12], '…')
		}
		top := ""
		if i == len(s.Stack)-1 {
			top = " ← TOP"
		}
		dfsLessonText(&b, 8, 16, string(name)+top)
		caller := f.Caller
		if caller == "" {
			caller = "root"
		}
		cr := []rune(caller)
		if len(cr) > 12 {
			caller = string(cr[:11]) + "…"
		}
		dfsLessonText(&b, 8, 33, fmt.Sprintf("from %s · next %d", caller, f.NextNeighbor))
		b.WriteString(`</g>`)
	}
	if len(s.Stack) == 0 {
		dfsLessonText(&b, px+16, py+60, "Empty stack")
	}
	strip := func(title string, list []string, y float64) {
		dfsLessonText(&b, 24, y, title)
		x := 24.0
		y += 12
		for _, id := range list {
			cw := math.Min(w-48, chipW(id))
			if x+cw > w-24 {
				x = 24
				y += 56
			}
			fmt.Fprintf(&b, `<g><title>%s</title><rect x="%.0f" y="%.0f" width="%.0f" height="48" rx="8" fill="#101923" stroke="%s"/>`, html.EscapeString(id), x, y, cw, colors[idx[id]%5])
			runes := []rune(id)
			count := int((cw - 20) / 8)
			if len(runes) > count {
				dfsLessonText(&b, x+10, y+19, string(runes[:count]))
				dfsLessonText(&b, x+10, y+37, string(runes[count:]))
			} else {
				dfsLessonText(&b, x+10, y+25, id)
			}
			b.WriteString(`</g>`)
			x += cw + 8
		}
	}
	strip("Preorder · first discovery", s.Order, stripY)
	strip("Finish order · all neighbors processed", s.FinishOrder, finishY)
	status := "Explore newest unfinished call first."
	if s.Skip {
		if s.BackEdge {
			status = "Already active: back-edge; skip."
		} else {
			status = "Already finished: skip; not a back-edge."
		}
	}
	if phase == "return" {
		status = "Return to parent; resume its neighbor cursor."
	}
	if phase == "done" {
		status = "✓ Done. Unreachable vertices stay untouched."
	}
	// Two lines retain readable text in the narrow scene.
	words := strings.Fields(status)
	line := ""
	y := statusY
	for _, word := range words {
		if len(line)+len(word) > 40 && w < 600 {
			dfsLessonText(&b, 24, y, line)
			line = ""
			y += 20
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	dfsLessonText(&b, 24, y, line)
	b.WriteString(`</g>`)
	return b.String()
}
