package renderers

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type BSTLessonState struct {
	TreeState
	Phase, Operation                               string
	Target                                         int
	ActiveID, ParentID, SuccessorID, ReplacementID string
	PathIDs, LeftIDs, RightIDs                     []string
	RemovedID                                      string
	OldValue                                       int
	HasOldValue                                    bool
	Found, ResultKnown                             bool
}

// BSTNodeSpacing leaves room for bounded integer values at 16px.
const BSTNodeSpacing = 112

func RenderBSTLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(BSTLessonState)
	if !ok {
		return `<g data-lesson="bst" data-scene-width="800" data-scene-height="440"><text x="32" y="40" fill="#eff5ed" font-size="14">No BST state available</text></g>`
	}
	w := 800
	if n := BSTNodeSpacing * (len(s.Nodes) + 1); n > w {
		w = n
	}
	h := 440
	positions := map[string][2]int{}
	indexes := map[string]int{}
	for i, n := range s.Nodes {
		indexes[n.ID] = i
	}
	rank := 0
	visiting := map[string]bool{}
	var place func(string, int)
	place = func(id string, depth int) {
		i, exists := indexes[id]
		if id == "" || !exists || visiting[id] {
			return
		}
		visiting[id] = true
		left, right := "", ""
		if i < len(s.LeftIDs) {
			left = s.LeftIDs[i]
		}
		if i < len(s.RightIDs) {
			right = s.RightIDs[i]
		}
		place(left, depth+1)
		positions[id] = [2]int{BSTNodeSpacing*(rank+1) + (w-BSTNodeSpacing*(len(s.Nodes)+1))/2, 88 + 76*depth}
		rank++
		if y := 152 + 76*depth; y > h {
			h = y
		}
		place(right, depth+1)
	}
	for _, n := range s.Nodes {
		if n.ParentID == "" {
			place(n.ID, 0)
		}
	}
	for _, n := range s.Nodes {
		if !visiting[n.ID] {
			place(n.ID, 0)
		}
	}
	if p, exists := positions[s.ParentID]; exists && p[1]+100 > h {
		h = p[1] + 100
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="bst" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString("BST "+s.Operation+" "+s.Phase))
	fmt.Fprintf(&b, `<text x="32" y="24">Target: %d</text><text x="220" y="24">%s · %s</text>`, s.Target, html.EscapeString(s.Operation), html.EscapeString(s.Phase))
	if s.ResultKnown {
		result := "not found"
		if s.Found {
			result = "found"
		}
		if s.Operation == "delete" && s.Found && (s.Phase == "remove" || s.Phase == "done") {
			result = "deleted"
		}
		fmt.Fprintf(&b, `<text x="480" y="24">%s</text>`, result)
	}
	path := map[string]bool{}
	for _, id := range s.PathIDs {
		path[id] = true
	}
	for _, n := range s.Nodes {
		p, ok := positions[n.ParentID]
		if !ok {
			continue
		}
		q := positions[n.ID]
		color := "#bbcab9"
		if path[n.ID] {
			color = "#ffe29a"
		}
		fmt.Fprintf(&b, `<path data-edge-child="%s" data-edge-parent="%s" d="M%d %d L%d %d" stroke="%s" stroke-width="2"/>`, html.EscapeString(n.ID), html.EscapeString(n.ParentID), p[0], p[1], q[0], q[1], color)
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for _, n := range s.Nodes {
		p := positions[n.ID]
		ordinal, _ := strconv.Atoi(strings.TrimPrefix(n.ID, "node-"))
		if ordinal < 1 {
			ordinal = 1
		}
		stroke := "#bbcab9"
		badge := ""
		emphasis := ""
		if path[n.ID] {
			stroke = "#9bf0bb"
			badge = "checked"
		}
		if n.ID == s.ActiveID {
			stroke = "#ffe29a"
			badge = s.Phase
			emphasis = ` data-emphasis="compare"`
		}
		if n.ID == s.ActiveID && s.ResultKnown && s.Found && (s.Phase == "done" || s.Phase == "duplicate") {
			stroke = "#9bf0bb"
			badge = "found"
		}
		label := strconv.Itoa(n.Val)
		if n.Label != "" {
			label = n.Label
		}
		fmt.Fprintf(&b, `<g id="bst-%s" data-entity="bst:%s" data-x="%d" data-y="%d" transform="translate(%d %d)"%s><title>%s</title><ellipse rx="48" ry="22" fill="%s" stroke="%s" stroke-width="3"/><text y="5" text-anchor="middle" font-size="16" fill="#101820">%s</text><text y="39" text-anchor="middle">%s</text><text y="-30" text-anchor="middle">%s</text></g>`, html.EscapeString(n.ID), html.EscapeString(n.ID), p[0], p[1], p[0], p[1], emphasis, html.EscapeString(label), palette[(ordinal-1)%5], stroke, html.EscapeString(label), html.EscapeString(n.ID), html.EscapeString(badge))
	}
	if (s.Phase == "descend" || (s.ResultKnown && !s.Found)) && s.ParentID != "" {
		if i, exists := indexes[s.ParentID]; exists {
			left := s.Target < s.Nodes[i].Val
			child := ""
			if left && i < len(s.LeftIDs) {
				child = s.LeftIDs[i]
			}
			if !left && i < len(s.RightIDs) {
				child = s.RightIDs[i]
			}
			if child == "" {
				p := positions[s.ParentID]
				dx := 32
				if left {
					dx = -32
				}
				fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="middle" fill="#ffe29a">nil</text>`, p[0]+dx, p[1]+76)
			}
		}
	}
	if s.Phase == "copy" {
		if _, exists := positions[s.ActiveID]; exists {
			fmt.Fprintf(&b, `<g data-entity="bst:copy" data-from-entity="bst:%s" data-x="680" data-y="24" transform="translate(680 24)"><rect x="-75" y="-17" width="150" height="30" rx="8" fill="#8bd5ff"/><text text-anchor="middle" y="4" fill="#101820">copy %d</text></g>`, html.EscapeString(s.SuccessorID), s.Nodes[indexes[s.ActiveID]].Val)
		}
	}
	if len(s.Nodes) == 0 {
		b.WriteString(`<text x="32" y="88">Empty tree · nil</text>`)
	}
	b.WriteString(`</g>`)
	return b.String()
}
