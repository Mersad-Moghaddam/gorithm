package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type DijkstraQueueEntry struct {
	Node             string
	Distance, Serial int
}
type DijkstraLessonState struct {
	GraphState
	Start, Target, Current, Neighbor                   string
	Distance                                           map[string]int
	Previous                                           map[string]string
	Settled                                            map[string]bool
	Queue                                              []DijkstraQueueEntry
	Path                                               []string
	EdgeIndex, OldDistance, Candidate, PoppedDistance  int
	HasCandidate, Improved, Stale, Found, InvalidStart bool
	// PoppedSerial preserves the extracted chip's identity for heap-to-node motion.
	PoppedSerial int
}

func dijkstraValue(v int) string {
	if v == math.MaxInt {
		return "∞"
	}
	return fmt.Sprint(v)
}

func RenderDijkstraLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(DijkstraLessonState)
	if !ok {
		return `<g data-lesson="dijkstra"><title>Dijkstra state unavailable</title></g>`
	}
	narrow := width < 600
	sw, gx, gy, gw, gh, lx, ly, lw := 960.0, 24.0, 64.0, 592.0, 348.0, 640.0, 64.0, 296.0
	if narrow {
		sw, gx, gy, gw, gh, lx, ly, lw = 360, 16, 56, 328, 240, 16, 320, 328
	}
	ledgerH := 28.0 + 30*float64(len(s.Nodes))
	heapY := math.Max(440, ly+ledgerH+28)
	if narrow {
		heapY = ly + ledgerH + 32
	}
	cols := int((sw - 48) / 120)
	if cols < 1 {
		cols = 1
	}
	rows := (len(s.Queue) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	eqY := heapY + 32 + float64(rows)*48 + 24
	pathRows := (len(s.Path) + cols - 1) / cols
	sceneH := math.Max(640, eqY+math.Max(150, 70+float64(pathRows)*48))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="dijkstra" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif" font-size="14" fill="#eff5ed"><title>%s</title><rect width="%.0f" height="%.0f" fill="#101c23"/>`, html.EscapeString(step.LineID), sw, sceneH, sw, html.EscapeString("Dijkstra: "+step.Explanation), sw, sceneH)
	text := func(x, y float64, v string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f">%s</text>`, x, y, html.EscapeString(v))
	}
	text(24, 30, "Dijkstra · minimum extraction and relaxation")
	fmt.Fprintf(&b, `<defs><marker id="dijkstra-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#bbcab9"/></marker></defs>`)
	points := make([][2]float64, len(s.Nodes))
	indexes := map[string]int{}
	xmin, xmax, ymin, ymax := math.Inf(1), math.Inf(-1), math.Inf(1), math.Inf(-1)
	ring := false
	for i, n := range s.Nodes {
		indexes[n.ID] = i
		if math.IsNaN(n.X) || math.IsNaN(n.Y) || math.IsInf(n.X, 0) || math.IsInf(n.Y, 0) {
			ring = true
			continue
		}
		xmin = math.Min(xmin, n.X)
		xmax = math.Max(xmax, n.X)
		ymin = math.Min(ymin, n.Y)
		ymax = math.Max(ymax, n.Y)
		for j := 0; j < i; j++ {
			if n.X == s.Nodes[j].X && n.Y == s.Nodes[j].Y {
				ring = true
			}
		}
	}
	scale := 1.0
	if xmax > xmin {
		scale = (gw - 100) / (xmax - xmin)
	}
	if ymax > ymin {
		sy := (gh - 100) / (ymax - ymin)
		if xmax <= xmin || sy < scale {
			scale = sy
		}
	}
	for i, n := range s.Nodes {
		if ring {
			a := 2*math.Pi*float64(i)/float64(len(s.Nodes)) - math.Pi/2
			points[i] = [2]float64{gx + gw/2 + (gw/2-50)*math.Cos(a), gy + gh/2 + (gh/2-50)*math.Sin(a)}
		} else {
			points[i] = [2]float64{gx + gw/2 + (n.X-(xmin+xmax)/2)*scale, gy + gh/2 + (n.Y-(ymin+ymax)/2)*scale}
		}
	}
	edgeMid := make([][2]float64, len(s.Edges))
	for i, e := range s.Edges {
		ai, aok := indexes[e.From]
		bi, bok := indexes[e.To]
		if !aok || !bok {
			continue
		}
		a, c := points[ai], points[bi]
		dx, dy := c[0]-a[0], c[1]-a[1]
		length := math.Hypot(dx, dy)
		offset := 0.0
		count := 0
		for j, f := range s.Edges {
			if j == i {
				continue
			}
			if (f.From == e.From && f.To == e.To) || (f.From == e.To && f.To == e.From) {
				if j < i {
					count++
				}
				offset = 18
			}
		}
		offset *= float64(count + 1)
		color := "#84949f"
		if e.Active {
			color = "#ffe29a"
			if s.Found {
				color = "#9bf0bb"
			}
		}
		marker := ""
		if e.Directed {
			marker = ` marker-end="url(#dijkstra-arrow)"`
		}
		var path string
		if length < 0.01 {
			path = fmt.Sprintf("M%.2f %.2f C%.2f %.2f %.2f %.2f %.2f %.2f", a[0]-16, a[1]-16, a[0]-55, a[1]-78, a[0]+55, a[1]-78, a[0]+16, a[1]-16)
			edgeMid[i] = [2]float64{a[0], a[1] - 65}
		} else {
			nx, ny := -dy/length, dx/length
			// Endpoint order supplies reciprocal curvature; indexed offsets separate parallel edges.
			cx, cy := (a[0]+c[0])/2+nx*offset, (a[1]+c[1])/2+ny*offset
			path = fmt.Sprintf("M%.2f %.2f Q%.2f %.2f %.2f %.2f", a[0]+dx/length*22, a[1]+dy/length*22, cx, cy, c[0]-dx/length*22, c[1]-dy/length*22)
			edgeMid[i] = [2]float64{(a[0]+c[0])/2 + nx*(offset/2+10), (a[1]+c[1])/2 + ny*(offset/2+10)}
		}
		fmt.Fprintf(&b, `<path id="dijkstra-edge-%d" d="%s" fill="none" stroke="%s" stroke-width="2.5"%s/>`, i, path, color, marker)
		p := edgeMid[i]
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="48" height="24" rx="8" fill="#101c23" stroke="#84949f"/>`, p[0]-24, p[1]-17)
		text(p[0]-16, p[1], fmt.Sprint(e.Weight))
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	entity := func(id string, x, y float64, extra string) {
		fmt.Fprintf(&b, `<g data-entity="%s" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s>`, id, x, y, x, y, extra)
	}
	for i, n := range s.Nodes {
		p := points[i]
		extra := ""
		if n.ID == s.Neighbor && step.LineID == "L10" {
			extra = ` data-emphasis="compare"`
		}
		entity(fmt.Sprintf("dijkstra:node:%d", i), p[0], p[1], extra)
		fmt.Fprintf(&b, `<title>%s</title><circle r="22" fill="%s" stroke="#eff5ed" stroke-width="2"/>`, html.EscapeString(n.Label), palette[i%5])
		if n.ID == s.Target {
			b.WriteString(`<circle r="28" fill="none" stroke="#8bd5ff" stroke-dasharray="3 3"/>`)
		}
		if n.ID == s.Current || n.ID == s.Neighbor {
			b.WriteString(`<circle r="25" fill="none" stroke="#ffe29a" stroke-width="3"/>`)
		}
		// Dark label backing keeps text contrast independent of identity color.
		b.WriteString(`<rect x="-21" y="-16" width="42" height="32" rx="7" fill="#101c23"/>`)
		label := []rune(n.Label)
		if len(label) == 0 {
			label = []rune(n.ID)
		}
		if len(label) > 8 {
			label = append(label[:7], '…')
		}
		if len(label) > 4 {
			text(-18, -2, string(label[:4]))
			text(-18, 13, string(label[4:]))
		} else {
			text(-15, 5, string(label))
		}
		if s.Settled[n.ID] {
			text(-24, 44, "✓ final")
		}
		if n.ID == s.Target {
			text(-24, -34, "target")
		}
		b.WriteString(`</g>`)
	}
	fmt.Fprintf(&b, `<rect x="%.0f" y="%.0f" width="%.0f" height="%.0f" rx="10" fill="#182832" stroke="#84949f"/>`, lx, ly, lw, ledgerH)
	text(lx+8, ly+20, "Vertex      d       prev     status")
	for i, n := range s.Nodes {
		y := ly + 50 + float64(i)*30
		label := []rune(n.ID)
		if len(label) > 6 {
			label = append(label[:5], '…')
		}
		text(lx+8, y, string(label))
		d, exists := s.Distance[n.ID]
		value := "unknown"
		if exists {
			value = dijkstraValue(d)
		}
		text(lx+78, y, value)
		prev := s.Previous[n.ID]
		if prev == "" {
			prev = "—"
		}
		pr := []rune(prev)
		if len(pr) > 6 {
			pr = append(pr[:5], '…')
		}
		text(lx+135, y, string(pr))
		status := "tentative"
		if !exists || d == math.MaxInt {
			status = "unreached"
		}
		if s.Settled[n.ID] {
			status = "✓ final"
		}
		text(lx+206, y, status)
	}
	text(24, heapY, "Actual heap array · root = min (not sorted)")
	for i, q := range s.Queue {
		x, y := 24+float64(i%cols)*120, heapY+24+float64(i/cols)*48
		entity(fmt.Sprintf("dijkstra:queue:%d", q.Serial), x, y, "")
		b.WriteString(`<rect width="112" height="36" rx="8" fill="#182832" stroke="#84949f"/>`)
		node := []rune(q.Node)
		if len(node) > 5 {
			node = append(node[:4], '…')
		}
		text(6, 16, fmt.Sprintf("[%d] %s:%s", i, string(node), dijkstraValue(q.Distance)))
		if i == 0 {
			text(6, 31, "root / min")
		}
		b.WriteString(`</g>`)
	}
	if len(s.Queue) == 0 {
		text(24, heapY+48, "empty heap")
	}
	if step.LineID == "L6" && s.PoppedSerial >= 0 {
		if i, exists := indexes[s.Current]; exists {
			p := points[i]
			x, y := p[0], p[1]+60
			if s.Stale {
				x, y = 24, eqY+80
			}
			entity(fmt.Sprintf("dijkstra:queue:%d", s.PoppedSerial), x, y, "")
			b.WriteString(`<rect x="-4" y="-19" width="164" height="30" rx="8" fill="#182832" stroke="#ffe29a"/>`)
			label := "extracted / final"
			if s.Stale {
				label = "already settled"
			}
			text(0, 0, label)
			b.WriteString(`</g>`)
		}
	}
	if s.HasCandidate {
		text(24, eqY, "old d[v] = "+dijkstraValue(s.OldDistance))
		relation := " ≥ "
		if s.Improved {
			relation = " < "
		}
		text(24, eqY+26, "candidate = "+dijkstraValue(s.Candidate)+relation+dijkstraValue(s.OldDistance))
		if s.EdgeIndex >= 0 && s.EdgeIndex < len(s.Edges) {
			pos := [2]float64{24, eqY + 52}
			if step.LineID == "L8" {
				pos = edgeMid[s.EdgeIndex]
			}
			entity(fmt.Sprintf("dijkstra:cost:%d", s.EdgeIndex), pos[0], pos[1], "")
			text(0, 0, fmt.Sprintf("d[u] + w = %s + %d", dijkstraValue(s.Distance[s.Current]), s.Edges[s.EdgeIndex].Weight))
			b.WriteString(`</g>`)
		}
	} else if step.LineID == "L8" && s.EdgeIndex >= 0 && s.EdgeIndex < len(edgeMid) {
		p := edgeMid[s.EdgeIndex]
		entity(fmt.Sprintf("dijkstra:cost:%d", s.EdgeIndex), p[0], p[1], "")
		text(0, 0, fmt.Sprintf("w = %d", s.Edges[s.EdgeIndex].Weight))
		b.WriteString(`</g>`)
	} else if step.LineID == "L17" {
		result := "No path · ∞"
		if s.Target == "" {
			result = "✓ All reachable distances finalized"
		}
		if s.InvalidStart {
			result = "Invalid start · no path"
		}
		if s.Found {
			result = "✓ Cost " + dijkstraValue(s.Distance[s.Target])
			text(24, eqY, result)
			for i, id := range s.Path {
				label := []rune(id)
				if len(label) > 12 {
					label = append(label[:11], '…')
				}
				fmt.Fprintf(&b, `<g><title>%s</title>`, html.EscapeString(id))
				text(24+float64(i%cols)*120, eqY+30+float64(i/cols)*48, string(label))
				b.WriteString(`</g>`)
			}
		} else {
			text(24, eqY, result)
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}
