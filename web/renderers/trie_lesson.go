package renderers

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type TrieLessonState struct {
	TreeState
	Phase, Operation, Word string
	Runes                  []rune
	RuneIndex, ByteOffset  int
	ActiveID, NextID       string
	PathIDs, TerminalIDs   []string
	Found, ResultKnown     bool
	Created                bool
}

func RenderTrieLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(TrieLessonState)
	if !ok || len(s.Nodes) == 0 || width <= 0 || height <= 0 {
		return `<g data-lesson="trie" data-phase="empty" data-scene-width="800" data-scene-height="480"><title>Trie unavailable</title><text x="32" y="40" fill="#eff5ed" font-size="14">No trie state available</text></g>`
	}
	children := map[string][]string{}
	nodes := map[string]TreeNode{}
	for _, n := range s.Nodes {
		nodes[n.ID] = n
		if n.ParentID != "" {
			children[n.ParentID] = append(children[n.ParentID], n.ID)
		}
	}
	// ponytail: bounded Build input yields at most 1025 nodes; direct malformed trees fall back rather than recurse forever.
	positions := map[string][2]float64{}
	visiting := map[string]bool{}
	leaves, depthMax := 0, 0
	var layout func(string, int) float64
	layout = func(id string, depth int) float64 {
		if visiting[id] || depth > len(s.Nodes) {
			return 48
		}
		visiting[id] = true
		if depth > depthMax {
			depthMax = depth
		}
		x := 48.0
		if len(children[id]) == 0 {
			x += 72 * float64(leaves)
			leaves++
		} else {
			first := layout(children[id][0], depth+1)
			last := first
			for _, child := range children[id][1:] {
				last = layout(child, depth+1)
			}
			x = (first + last) / 2
		}
		positions[id] = [2]float64{x, 112 + 80*float64(depth)}
		return x
	}
	for _, n := range s.Nodes {
		if n.ParentID == "" {
			layout(n.ID, 0)
		}
	}
	w := max(800, 96+72*(leaves-1), 64+40*len(s.Runes))
	h := max(480, 184+80*depthMax)
	offset := float64(w-(96+72*(leaves-1))) / 2
	for id, p := range positions {
		p[0] += offset
		positions[id] = p
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="trie" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14"><title>%s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(s.Operation+" "+s.Word))
	for i, ch := range s.Runes {
		fill := "#eff5ed"
		if i == s.RuneIndex {
			fill = "#ffe29a"
		}
		x := 32 + 40*i
		fmt.Fprintf(&b, `<rect x="%d" y="12" width="32" height="32" rx="6" fill="%s"/><text x="%d" y="34" text-anchor="middle" fill="#101820" font-size="16">%s</text>`, x, fill, x+16, html.EscapeString(string(ch)))
		if i < len(s.PathIDs)-1 {
			fmt.Fprintf(&b, `<path d="M%d 49h32" stroke="#9bf0bb" stroke-width="3"/>`, x)
		}
	}
	status := s.Operation + " · " + s.Phase
	if s.ResultKnown {
		if s.Operation == "search" {
			status = fmt.Sprintf("path exists; word ends here = %t", s.Found)
			if s.Phase == "missing" {
				status = "missing edge; word not found"
			}
		} else {
			status = fmt.Sprintf("path exists = %t (prefix query)", s.Found)
		}
	}
	if s.RuneIndex >= 0 {
		status += fmt.Sprintf(" · rune %d / byte %d", s.RuneIndex, s.ByteOffset)
	}
	fmt.Fprintf(&b, `<text x="32" y="76" fill="#eff5ed">%s</text>`, html.EscapeString(status))
	for _, n := range s.Nodes {
		p, exists := positions[n.ParentID]
		q := positions[n.ID]
		if n.ParentID != "" && exists {
			fmt.Fprintf(&b, `<g data-edge="%s" data-parent="%s"><path d="M%.1f %.1fL%.1f %.1f" stroke="#bbcab9" stroke-width="2"/><text x="%.1f" y="%.1f" fill="#eff5ed">%s</text></g>`, html.EscapeString(n.ID), html.EscapeString(n.ParentID), p[0], p[1]+20, q[0], q[1]-20, (p[0]+q[0])/2+8, (p[1]+q[1])/2, html.EscapeString(n.Label))
		}
	}
	terminals := map[string]bool{}
	for _, id := range s.TerminalIDs {
		terminals[id] = true
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for _, n := range s.Nodes {
		p := positions[n.ID]
		ordinal := 0
		if n.ID != "root" {
			ordinal, _ = strconv.Atoi(strings.TrimPrefix(n.ID, "node-"))
		}
		if ordinal < 0 {
			ordinal = 0
		}
		stroke := "#eff5ed"
		cue := ""
		emphasis := ""
		if n.ID == s.ActiveID {
			stroke = "#ffe29a"
			cue = "current"
			if s.Phase == "compare" {
				cue = "lookup"
				emphasis = ` data-emphasis="compare"`
			}
		}
		if s.Created && n.ID == s.NextID {
			stroke = "#8bd5ff"
			cue = "new"
		}
		fmt.Fprintf(&b, `<g id="trie-%s" data-entity="node:%s" data-x="%.1f" data-y="%.1f" transform="translate(%.1f %.1f)"%s><circle r="20" fill="%s" stroke="%s" stroke-width="3"/><text y="5" text-anchor="middle" fill="#101820" font-size="16">%s</text>`, html.EscapeString(n.ID), html.EscapeString(n.ID), p[0], p[1], p[0], p[1], emphasis, palette[ordinal%5], stroke, html.EscapeString(n.Label))
		if terminals[n.ID] {
			fmt.Fprintf(&b, `<circle data-terminal="%s" r="25" fill="none" stroke="#9bf0bb" stroke-width="2"/><text x="29" y="5" fill="#9bf0bb">end</text>`, html.EscapeString(n.ID))
		}
		fmt.Fprintf(&b, `<text y="42" text-anchor="middle" fill="#bbcab9" font-size="12">%s</text><text y="-31" text-anchor="middle" fill="#eff5ed">%s</text></g>`, html.EscapeString(n.ID), cue)
	}
	if s.Phase == "missing" && s.RuneIndex >= 0 && s.RuneIndex < len(s.Runes) {
		p := positions[s.ActiveID]
		fmt.Fprintf(&b, `<g data-missing="true"><path d="M%.1f %.1fv48" fill="none" stroke="#ff9b9b" stroke-dasharray="5 4"/><text x="%.1f" y="%.1f" fill="#ff9b9b">%s missing ×</text></g>`, p[0], p[1]+20, p[0]+8, p[1]+64, html.EscapeString(string(s.Runes[s.RuneIndex])))
	}
	b.WriteString(`</g>`)
	return b.String()
}
