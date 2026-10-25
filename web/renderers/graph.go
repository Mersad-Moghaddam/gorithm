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
		return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">No graph state available</text>`
	}

	nodeMap := make(map[string]GraphNode)
	for _, n := range state.Nodes {
		nodeMap[n.ID] = n
	}

	var sb strings.Builder

	// Defs for gradients & glowing nodes
	sb.WriteString(`<defs>
		<filter id="node-glow" x="-50%" y="-50%" width="200%" height="200%">
			<feGaussianBlur in="SourceGraphic" stdDeviation="6" result="blur" />
			<feMerge>
				<feMergeNode in="blur" />
				<feMergeNode in="SourceGraphic" />
			</feMerge>
		</filter>
		<linearGradient id="grad-unvisited" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#1e293b" />
			<stop offset="100%" stop-color="#0f172a" />
		</linearGradient>
		<linearGradient id="grad-visiting" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#38bdf8" />
			<stop offset="100%" stop-color="#0284c7" />
		</linearGradient>
		<linearGradient id="grad-visited" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#4ade80" />
			<stop offset="100%" stop-color="#16a34a" />
		</linearGradient>
		<linearGradient id="grad-queued" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#fde047" />
			<stop offset="100%" stop-color="#ca8a04" />
		</linearGradient>
		<linearGradient id="grad-path" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#c084fc" />
			<stop offset="100%" stop-color="#7e22ce" />
		</linearGradient>
	</defs>`)

	// Render Edges
	for _, e := range state.Edges {
		from, ok1 := nodeMap[e.From]
		to, ok2 := nodeMap[e.To]
		if !ok1 || !ok2 {
			continue
		}

		edgeColor := "#334155"
		strokeW := 2.0
		filterStr := ""
		if e.Active {
			edgeColor = "#38bdf8"
			strokeW = 3.5
			filterStr = `filter="url(#node-glow)"`
		}

		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%.1f" %s stroke-linecap="round" style="transition: all 0.3s ease;" />`,
			from.X, from.Y, to.X, to.Y, edgeColor, strokeW, filterStr,
		))

		if e.Weight > 0 {
			midX := (from.X + to.X) / 2
			midY := (from.Y + to.Y) / 2
			// Pill backdrop for weight
			sb.WriteString(fmt.Sprintf(
				`<rect x="%.1f" y="%.1f" width="22" height="16" rx="8" fill="#090d16" stroke="#334155" stroke-width="1" />`,
				midX-11, midY-8,
			))
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#94a3b8" font-size="10" font-family="ui-monospace, monospace" font-weight="700" text-anchor="middle" dy="4">%d</text>`,
				midX, midY, e.Weight,
			))
		}
	}

	// Render Nodes
	for _, n := range state.Nodes {
		fillRef := "url(#grad-unvisited)"
		strokeColor := "#475569"
		textColor := "#f8fafc"
		filterStr := ""

		switch n.State {
		case "visiting":
			fillRef = "url(#grad-visiting)"
			strokeColor = "#bae6fd"
			filterStr = `filter="url(#node-glow)"`
		case "visited":
			fillRef = "url(#grad-visited)"
			strokeColor = "#bbf7d0"
		case "queued":
			fillRef = "url(#grad-queued)"
			strokeColor = "#fef08a"
			filterStr = `filter="url(#node-glow)"`
		case "path":
			fillRef = "url(#grad-path)"
			strokeColor = "#f3e8ff"
			filterStr = `filter="url(#node-glow)"`
		}

		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="23" fill="%s" stroke="%s" stroke-width="2.5" %s style="transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);" />`,
			n.X, n.Y, fillRef, strokeColor, filterStr,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="%s" font-size="13" font-family="ui-monospace, Menlo, monospace" font-weight="800" text-anchor="middle" dy="5">%s</text>`,
			n.X, n.Y, textColor, n.Label,
		))
	}

	return sb.String()
}
