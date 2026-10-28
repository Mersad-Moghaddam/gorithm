package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type SegmentTreeLessonNode struct {
	ID, ParentID string
	L, R, Sum    int
	Built        bool
}

type SegmentTreeLessonState struct {
	Nodes                              []SegmentTreeLessonNode
	Values                             []int
	Phase, Overlap, ActiveID           string
	QL, QR, UpdateIndex, UpdateValue   int
	AcceptedIDs                        []string
	Result, LeftResult, RightResult    int
	ResultKnown, LeftKnown, RightKnown bool
	PathIDs                            []string
}

func RenderSegmentTreeLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(SegmentTreeLessonState)
	if !ok || len(s.Nodes) == 0 || len(s.Values) == 0 || len(s.Values) > 64 {
		return `<g data-lesson="segmenttree" data-phase="empty" data-scene-width="800" data-scene-height="500"><title>Segment tree unavailable</title><text x="32" y="44" fill="#eff5ed" font-size="14">No segment tree state available</text></g>`
	}
	w := max(800, 64+116*len(s.Values))
	positions := make(map[string][2]float64, len(s.Nodes))
	nodes := make(map[string]SegmentTreeLessonNode, len(s.Nodes))
	for _, n := range s.Nodes {
		nodes[n.ID] = n
	}
	depthMax := 0
	for _, n := range s.Nodes {
		depth := 0
		parent := n.ParentID
		for parent != "" && depth < len(s.Nodes) {
			p, found := nodes[parent]
			if !found {
				break
			}
			depth++
			parent = p.ParentID
		}
		depthMax = max(depthMax, depth)
		x := float64(w-116*len(s.Values))/2 + 116*float64(n.L+n.R+1)/2
		positions[n.ID] = [2]float64{x, float64(152 + 92*depth)}
	}
	h := max(500, 300+92*depthMax+24*((len(s.AcceptedIDs)+5)/6))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="segmenttree" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14"><title>Segment tree: %s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(s.Phase))
	text := func(x, y float64, value, color string, size int) {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" text-anchor="middle" fill="%s" font-size="%d">%s</text>`, x, y, color, size, html.EscapeString(value))
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	start := float64(w-116*len(s.Values)) / 2
	for i, v := range s.Values {
		x := start + 116*(float64(i)+0.5)
		stroke := "#eff5ed"
		if s.Phase == "update" && i == s.UpdateIndex {
			stroke = "#8bd5ff"
		}
		fmt.Fprintf(&b, `<g data-entity="array:%d" data-x="%.1f" data-y="44" transform="translate(%.1f 44)"><rect x="-50" y="-22" width="100" height="44" rx="8" fill="%s" stroke="%s" stroke-width="2"/><text y="6" text-anchor="middle" fill="#ffffff" font-size="16">%d</text></g>`, i, x, x, palette[i%5], stroke, v)
		text(x, 86, fmt.Sprintf("[%d]", i), "#bbcab9", 14)
	}
	if s.Phase == "query" && s.QL >= 0 && s.QR >= s.QL && s.QR < len(s.Values) {
		fmt.Fprintf(&b, `<path d="M %.1f 94 v 8 H %.1f v -8" fill="none" stroke="#ffe29a" stroke-width="3"/>`, start+116*float64(s.QL)+4, start+116*float64(s.QR+1)-4)
		text(float64(w)/2, 124, fmt.Sprintf("Query [%d, %d] inclusive", s.QL, s.QR), "#ffe29a", 14)
	} else if s.Phase == "update" {
		text(float64(w)/2, 124, fmt.Sprintf("Update [%d] = %d; leaf and ancestors only", s.UpdateIndex, s.UpdateValue), "#8bd5ff", 14)
	}
	for _, n := range s.Nodes {
		if p, found := positions[n.ParentID]; n.ParentID != "" && found {
			c := positions[n.ID]
			fmt.Fprintf(&b, `<path id="segmenttree-edge-%s" data-child="%s" data-parent="%s" d="M %.1f %.1f L %.1f %.1f" stroke="#bbcab9" fill="none"/>`, html.EscapeString(n.ID), html.EscapeString(n.ID), html.EscapeString(n.ParentID), p[0], p[1]+26, c[0], c[1]-26)
		}
	}
	accepted := make(map[string]bool)
	for _, id := range s.AcceptedIDs {
		accepted[id] = true
	}
	path := make(map[string]bool)
	for _, id := range s.PathIDs {
		path[id] = true
	}
	for ordinal, n := range s.Nodes {
		p := positions[n.ID]
		stroke := "#bbcab9"
		cue := ""
		dash := ""
		if path[n.ID] {
			stroke = "#8bd5ff"
		}
		if accepted[n.ID] {
			stroke = "#9bf0bb"
			cue = "checked"
		}
		if n.ID == s.ActiveID {
			stroke = "#ffe29a"
			cue = "active"
			switch s.Overlap {
			case "full":
				stroke = "#9bf0bb"
				cue = "checked"
			case "none":
				cue = "+0 disjoint"
				dash = ` stroke-dasharray="5 4"`
			case "partial":
				cue = "split"
			}
		}
		sum := "?"
		if n.Built {
			sum = fmt.Sprint(n.Sum)
		}
		// Wider signed sums keep their full labels; spacing remains leaf-aligned.
		boxWidth := max(68, 12+9*len(sum))
		fmt.Fprintf(&b, `<g id="segmenttree-%s" data-entity="tree:%s" data-x="%.1f" data-y="%.1f" transform="translate(%.1f %.1f)"><title>%s interval [%d, %d]</title><rect x="%d" y="-26" width="%d" height="52" rx="8" fill="%s" stroke="%s" stroke-width="3"%s/><text y="-6" text-anchor="middle" fill="#ffffff">[%d, %d]</text><text y="16" text-anchor="middle" fill="#ffffff" font-size="16">%s</text>`, html.EscapeString(n.ID), html.EscapeString(n.ID), p[0], p[1], p[0], p[1], html.EscapeString(n.ID), n.L, n.R, -boxWidth/2, boxWidth, palette[ordinal%5], stroke, dash, n.L, n.R, sum)
		if cue != "" {
			fmt.Fprintf(&b, `<text y="46" text-anchor="middle" fill="%s">%s</text>`, stroke, cue)
		}
		b.WriteString(`</g>`)
	}
	y := float64(222 + 92*depthMax)
	if s.LeftKnown || s.RightKnown {
		left, right, result := "?", "?", "?"
		if s.LeftKnown {
			left = fmt.Sprint(s.LeftResult)
		}
		if s.RightKnown {
			right = fmt.Sprint(s.RightResult)
		}
		if s.ResultKnown {
			result = fmt.Sprint(s.Result)
		}
		text(float64(w)/2, y, left+" + "+right+" = "+result, "#ffe29a", 16)
		y += 28
	} else if s.ResultKnown {
		text(float64(w)/2, y, fmt.Sprintf("Result: %d", s.Result), "#9bf0bb", 16)
		y += 28
	}
	if s.Phase == "query" {
		text(float64(w)/2, y, "Accepted cover (each interval counted once)", "#9bf0bb", 14)
		for i, id := range s.AcceptedIDs {
			if n, found := nodes[id]; found {
				text(100+float64(i%6)*126, y+24+float64(i/6)*24, fmt.Sprintf("[%d,%d]: %d", n.L, n.R, n.Sum), "#eff5ed", 14)
			}
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}
