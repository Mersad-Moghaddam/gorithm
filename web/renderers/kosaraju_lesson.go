package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type KosarajuLessonState struct {
	GraphState
	Pass                                 int
	Transposed                           bool
	Current, Neighbor                    string
	Stack, FinishOrder, CurrentComponent []string
	Components                           [][]string
	ComponentOf                          map[string]int
	VisitedForward, VisitedReverse       map[string]bool
	OrderIndex, EdgeIndex                int
	OriginalEdges                        []GraphEdge
	FinishedVertex, ReturningFrom        string
}

var kosarajuColors = [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}

func RenderKosarajuLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(KosarajuLessonState)
	if !ok {
		return `<g data-lesson="kosaraju"><title>No Kosaraju state</title></g>`
	}
	narrow := width < 600
	w, gx, gy, gw, gh, px, py := 960.0, 24.0, 64.0, 592.0, 348.0, 640.0, 64.0
	if narrow {
		w, gx, gy, gw, gh, px, py = 360, 16, 56, 328, 240, 16, 320
	}
	points := kosarajuPositions(s.Nodes, gx, gy, gw, gh)
	index := map[string]int{}
	for i, n := range s.Nodes {
		index[n.ID] = i
	}
	var body strings.Builder
	text := func(x, y float64, value string, color string) {
		fmt.Fprintf(&body, `<text x="%.2f" y="%.2f" fill="%s" font-size="14">%s</text>`, x, y, color, html.EscapeString(value))
	}
	orientation := "Original graph"
	if s.Transposed {
		orientation = "Transpose graph"
	}
	text(24, 30, fmt.Sprintf("Pass %d · %s", s.Pass, orientation), "#eff5ed")
	body.WriteString(`<defs><marker id="kosaraju-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`)
	for i, e := range s.Edges {
		a, aok := index[e.From]
		c, cok := index[e.To]
		if !aok || !cok {
			continue
		}
		p, q := points[a], points[c]
		color := "#81988f"
		if e.Active {
			color = "#ffe29a"
		}
		d := kosarajuEdgePath(p, q, i, a == c)
		fmt.Fprintf(&body, `<path id="kosaraju-edge-%d" data-edge-index="%d" d="%s" fill="none" stroke="%s" stroke-width="2" marker-end="url(#kosaraju-arrow)"><title>Edge %d: %s → %s</title></path>`, i, i, d, color, i, html.EscapeString(e.From), html.EscapeString(e.To))
		if step.LineID == "L7" || step.LineID == "L16" {
			fmt.Fprintf(&body, `<path d="%s" fill="none" stroke="#8bd5ff" stroke-dasharray="5 5" opacity="0.45"><title>Previous direction; arrows now reversed</title></path>`, kosarajuEdgePath(q, p, i, a == c))
		}
	}
	for i, n := range s.Nodes {
		p := points[i]
		stroke := "#bbcab9"
		if n.ID == s.Current {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&body, `<g data-entity="kosaraju-node:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>%s</title><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><circle r="15" fill="#14241f"/><text y="5" text-anchor="middle" fill="#eff5ed" font-size="14">%d</text>`, i, p[0], p[1], p[0], p[1], html.EscapeString(n.Label), kosarajuColors[i%5], stroke, i+1)
		if n.ID == s.Current {
			body.WriteString(`<text y="-30" text-anchor="middle" fill="#ffe29a" font-size="14">Current</text>`)
		}
		if component, assigned := s.ComponentOf[n.ID]; assigned {
			fmt.Fprintf(&body, `<text y="39" text-anchor="middle" fill="#9bf0bb" font-size="14">SCC %d</text>`, component+1)
		}
		body.WriteString(`</g>`)
	}
	y := py + 20
	text(px, y, "Call stack · top last", "#eff5ed")
	y += 28
	if len(s.Stack) == 0 {
		text(px, y, "Empty", "#bbcab9")
		y += 30
	}
	for _, id := range s.Stack {
		for _, line := range kosarajuWrap(id, 18) {
			text(px, y, line, "#eff5ed")
			y += 20
		}
		y += 10
	}
	text(px, y+10, "Node key · stable identity", "#eff5ed")
	y += 40
	for i, n := range s.Nodes {
		for j, line := range kosarajuWrap(fmt.Sprintf("%d: %s", i+1, n.Label), 18) {
			color := "#eff5ed"
			if j == 0 {
				color = "#bbcab9"
			}
			text(px, y, line, color)
			y += 20
		}
		y += 10
	}
	y = math.Max(y+20, gy+gh+28)
	columns := int((w - 48) / 72)
	if columns < 1 {
		columns = 1
	}
	finishLabel := "Finish stack · latest at right; inspect right to left"
	if narrow {
		finishLabel = "Finish stack · latest last"
	}
	text(24, y, finishLabel, "#eff5ed")
	y += 28
	for j, id := range s.FinishOrder {
		ni := index[id]
		x := 24 + float64(j%columns)*72
		cy := y + float64(j/columns)*48
		alias := ""
		if id == s.FinishedVertex && step.Action == trace.ActionPush {
			alias = fmt.Sprintf(` data-from-entity="kosaraju-node:%d"`, ni)
		}
		fmt.Fprintf(&body, `<g data-entity="kosaraju-finish:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>%s</title><rect width="56" height="32" rx="8" fill="#14241f" stroke="%s"/><text x="28" y="21" text-anchor="middle" fill="#eff5ed" font-size="14">%d</text>`, ni, x, cy, x, cy, alias, html.EscapeString(id), kosarajuColors[ni%5], ni+1)
		body.WriteString(`</g>`)
	}
	if s.OrderIndex >= 0 && s.OrderIndex < len(s.FinishOrder) {
		x, cy := 24+float64(s.OrderIndex%columns)*72, y+float64(s.OrderIndex/columns)*48
		fmt.Fprintf(&body, `<g data-entity="kosaraju-cursor:0" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><path d="M15 -8H41" stroke="#ffe29a" stroke-width="4"/><text x="28" y="46" text-anchor="middle" fill="#ffe29a" font-size="12">Inspect</text></g>`, x, cy, x, cy)
	}
	y += float64((len(s.FinishOrder)+columns-1)/columns)*48 + 28
	componentLabel := "Strong components · mutual reachability"
	if narrow {
		componentLabel = "SCCs · mutual reachability"
	}
	text(24, y, componentLabel, "#eff5ed")
	y += 28
	cards := s.Components
	for ci := 0; ci <= len(cards); ci++ {
		members := s.CurrentComponent
		committed := ci < len(cards)
		if committed {
			members = cards[ci]
		} else if len(members) == 0 || step.LineID == "L13" || step.LineID == "L8" || step.LineID == "L16" {
			continue
		}
		label := fmt.Sprintf("SCC %d · current", ci+1)
		if committed {
			label = fmt.Sprintf("✓ SCC %d · committed", ci+1)
		}
		text(24, y, label, "#9bf0bb")
		y += 24
		for _, id := range members {
			ni := index[id]
			alias := ""
			if !committed && id == s.Current && step.Action == trace.ActionVisit {
				alias = fmt.Sprintf(` data-from-entity="kosaraju-node:%d"`, ni)
			}
			fmt.Fprintf(&body, `<g data-entity="kosaraju-member:%d" data-x="24" data-y="%.2f" transform="translate(24 %.2f)"%s><title>%s</title><rect x="0" y="-14" width="8" height="18" fill="%s"/>`, ni, y, y, alias, html.EscapeString(id), kosarajuColors[ni%5])
			for j, line := range kosarajuWrap(id, int((w-76)/14)) {
				fmt.Fprintf(&body, `<text x="16" y="%d" fill="#eff5ed" font-size="14">%s</text>`, j*20, html.EscapeString(line))
				y += 20
			}
			body.WriteString(`</g>`)
			y += 10
		}
		y += 16
	}
	h := math.Max(640, y+20)
	return fmt.Sprintf(`<g data-lesson="kosaraju" data-phase="pass-%d-%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui, sans-serif"><title>%s</title>%s</g>`, s.Pass, map[bool]string{true: "transpose", false: "original"}[s.Transposed], w, h, w, html.EscapeString(step.Explanation), body.String())
}

func kosarajuWrap(value string, count int) []string {
	r := []rune(value)
	if len(r) == 0 {
		return []string{""}
	}
	out := []string{}
	for len(r) > 0 {
		n := min(count, len(r))
		out = append(out, string(r[:n]))
		r = r[n:]
	}
	return out
}

func kosarajuPositions(nodes []GraphNode, x, y, w, h float64) [][2]float64 {
	p := make([][2]float64, len(nodes))
	if len(nodes) == 0 {
		return p
	}
	xmin, xmax, ymin, ymax := nodes[0].X, nodes[0].X, nodes[0].Y, nodes[0].Y
	coincident := false
	for i, n := range nodes {
		xmin = math.Min(xmin, n.X)
		xmax = math.Max(xmax, n.X)
		ymin = math.Min(ymin, n.Y)
		ymax = math.Max(ymax, n.Y)
		for j := 0; j < i; j++ {
			if n.X == nodes[j].X && n.Y == nodes[j].Y {
				coincident = true
			}
		}
	}
	if coincident || math.IsNaN(xmin+xmax+ymin+ymax) || math.IsInf(xmin+xmax+ymin+ymax, 0) {
		for i := range nodes {
			a := 2*math.Pi*float64(i)/float64(len(nodes)) - math.Pi/2
			p[i] = [2]float64{x + w/2 + (w/2-48)*math.Cos(a), y + h/2 + (h/2-48)*math.Sin(a)}
		}
		return p
	}
	scale := 1.0
	dx, dy := xmax-xmin, ymax-ymin
	if dx > 0 && dy > 0 {
		scale = math.Min((w-96)/dx, (h-96)/dy)
	} else if dx > 0 {
		scale = (w - 96) / dx
	} else if dy > 0 {
		scale = (h - 96) / dy
	}
	for i, n := range nodes {
		p[i] = [2]float64{x + w/2 + (n.X-(xmin+xmax)/2)*scale, y + h/2 + (n.Y-(ymin+ymax)/2)*scale}
	}
	return p
}

func kosarajuEdgePath(p, q [2]float64, i int, loop bool) string {
	if loop {
		r := 26 + float64(i)*0.05
		return fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f", p[0]-16, p[1]-16, p[0]-r*2, p[1]-r*3, p[0]+r*2, p[1]-r*3, p[0]+16, p[1]-16)
	}
	dx, dy := q[0]-p[0], q[1]-p[1]
	length := math.Hypot(dx, dy)
	if length == 0 {
		length = 1
	}
	ux, uy := dx/length, dy/length
	offset := 18 + float64(i)*0.5
	cx, cy := (p[0]+q[0])/2-uy*offset, (p[1]+q[1])/2+ux*offset
	ax, ay := cx-p[0], cy-p[1]
	al := math.Hypot(ax, ay)
	bx, by := cx-q[0], cy-q[1]
	bl := math.Hypot(bx, by)
	return fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f", p[0]+22*ax/al, p[1]+22*ay/al, cx, cy, q[0]+22*bx/bl, q[1]+22*by/bl)
}
