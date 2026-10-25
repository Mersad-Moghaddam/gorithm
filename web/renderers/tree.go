package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type TreeNode struct {
	ID       string  `json:"id"`
	Val      int     `json:"val"`
	Label    string  `json:"label,omitempty"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	ParentID string  `json:"parentId"`
	State    string  `json:"state"` // "normal", "active", "matched", "deleted", "inserted"
}

type TreeState struct {
	Nodes []TreeNode `json:"nodes"`
}

func RenderTreeSVG(step trace.Step, width, height int) string {
	state, ok := step.State.(TreeState)
	if !ok {
		return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">No tree state available</text>`
	}

	nodeMap := make(map[string]TreeNode)
	for _, n := range state.Nodes {
		nodeMap[n.ID] = n
	}

	var sb strings.Builder

	// Defs
	sb.WriteString(`<defs>
		<filter id="tree-glow" x="-40%" y="-40%" width="180%" height="180%">
			<feGaussianBlur in="SourceGraphic" stdDeviation="5" result="blur" />
			<feMerge>
				<feMergeNode in="blur" />
				<feMergeNode in="SourceGraphic" />
			</feMerge>
		</filter>
		<linearGradient id="tree-grad-node" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#1e293b" />
			<stop offset="100%" stop-color="#0f172a" />
		</linearGradient>
		<linearGradient id="tree-grad-active" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#38bdf8" />
			<stop offset="100%" stop-color="#0369a1" />
		</linearGradient>
		<linearGradient id="tree-grad-match" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#4ade80" />
			<stop offset="100%" stop-color="#15803d" />
		</linearGradient>
		<linearGradient id="tree-grad-insert" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#c084fc" />
			<stop offset="100%" stop-color="#7e22ce" />
		</linearGradient>
		<linearGradient id="tree-grad-delete" x1="0%" y1="0%" x2="100%" y2="100%">
			<stop offset="0%" stop-color="#f87171" />
			<stop offset="100%" stop-color="#b91c1c" />
		</linearGradient>
	</defs>`)

	// Smooth curved organic branch connectors
	for _, n := range state.Nodes {
		if n.ParentID != "" {
			if parent, exists := nodeMap[n.ParentID]; exists {
				midY := (parent.Y + n.Y) / 2
				sb.WriteString(fmt.Sprintf(
					`<path d="M %.1f %.1f C %.1f %.1f, %.1f %.1f, %.1f %.1f" fill="none" stroke="#334155" stroke-width="2.5" stroke-linecap="round" />`,
					parent.X, parent.Y, parent.X, midY, n.X, midY, n.X, n.Y,
				))
			}
		}
	}

	// Nodes
	for _, n := range state.Nodes {
		fillRef := "url(#tree-grad-node)"
		strokeColor := "#475569"
		filterStr := ""

		switch n.State {
		case "active":
			fillRef = "url(#tree-grad-active)"
			strokeColor = "#bae6fd"
			filterStr = `filter="url(#tree-glow)"`
		case "matched":
			fillRef = "url(#tree-grad-match)"
			strokeColor = "#bbf7d0"
			filterStr = `filter="url(#tree-glow)"`
		case "inserted":
			fillRef = "url(#tree-grad-insert)"
			strokeColor = "#f3e8ff"
			filterStr = `filter="url(#tree-glow)"`
		case "deleted":
			fillRef = "url(#tree-grad-delete)"
			strokeColor = "#fecaca"
			filterStr = `filter="url(#tree-glow)"`
		}

		displayText := fmt.Sprintf("%d", n.Val)
		if n.Label != "" {
			displayText = n.Label
		}
		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="22" fill="%s" stroke="%s" stroke-width="2.5" %s style="transition: fill 0.25s ease, stroke 0.25s ease;" />`,
			n.X, n.Y, fillRef, strokeColor, filterStr,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#f8fafc" font-size="13" font-family="ui-monospace, Menlo, monospace" font-weight="700" text-anchor="middle" dy="5">%s</text>`,
			n.X, n.Y, displayText,
		))
	}

	return sb.String()
}
