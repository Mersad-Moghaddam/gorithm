package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type GraphNode struct {
	ID    string  `json:"id"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	State string  `json:"state"` // "unvisited", "visiting", "visited", "queued", "path"
}

type GraphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Weight int    `json:"weight,omitempty"`
	Active bool   `json:"active"`
}

type GraphState struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}

func RenderGraphSVG(step trace.Step, width, height int) string {
	state, ok := step.State.(GraphState)
	if !ok {
		return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No graph state</text>`
	}

	nodeMap := make(map[string]GraphNode)
	for _, n := range state.Nodes {
		nodeMap[n.ID] = n
	}

	var sb strings.Builder

	// Render Edges
	for _, e := range state.Edges {
		from, ok1 := nodeMap[e.From]
		to, ok2 := nodeMap[e.To]
		if !ok1 || !ok2 {
			continue
		}

		edgeColor := "#2d333b"
		strokeW := 2
		if e.Active {
			edgeColor = "#eab308"
			strokeW = 3
		}

		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%d" style="transition: all 0.25s ease;" />`,
			from.X, from.Y, to.X, to.Y, edgeColor, strokeW,
		))

		if e.Weight > 0 {
			midX := (from.X + to.X) / 2
			midY := (from.Y + to.Y) / 2
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="11" font-family="ui-monospace, monospace" text-anchor="middle" dy="-5">%d</text>`,
				midX, midY, e.Weight,
			))
		}
	}

	// Render Nodes
	for _, n := range state.Nodes {
		fillColor := "#1f242c"
		strokeColor := "#444c56"

		switch n.State {
		case "visiting":
			fillColor = "#3b82f6"
			strokeColor = "#93c5fd"
		case "visited":
			fillColor = "#22c55e"
			strokeColor = "#86efac"
		case "queued":
			fillColor = "#eab308"
			strokeColor = "#fde047"
		case "path":
			fillColor = "#a855f7"
			strokeColor = "#d8b4fe"
		}

		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="22" fill="%s" stroke="%s" stroke-width="2.5" style="transition: all 0.25s ease;" />`,
			n.X, n.Y, fillColor, strokeColor,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#f0f6fc" font-size="13" font-family="ui-monospace, monospace" font-weight="700" text-anchor="middle" dy="4">%s</text>`,
			n.X, n.Y, n.Label,
		))
	}

	return sb.String()
}
