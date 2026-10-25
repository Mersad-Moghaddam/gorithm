package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type TreeNode struct {
	ID       string  `json:"id"`
	Val      int     `json:"val"`
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
		return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No tree state</text>`
	}

	nodeMap := make(map[string]TreeNode)
	for _, n := range state.Nodes {
		nodeMap[n.ID] = n
	}

	var sb strings.Builder

	// Edges to parents
	for _, n := range state.Nodes {
		if n.ParentID != "" {
			if parent, exists := nodeMap[n.ParentID]; exists {
				sb.WriteString(fmt.Sprintf(
					`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#2d333b" stroke-width="2" />`,
					parent.X, parent.Y, n.X, n.Y,
				))
			}
		}
	}

	// Nodes
	for _, n := range state.Nodes {
		fillColor := "#1f242c"
		strokeColor := "#444c56"

		switch n.State {
		case "active":
			fillColor = "#3b82f6"
			strokeColor = "#93c5fd"
		case "matched":
			fillColor = "#22c55e"
			strokeColor = "#86efac"
		case "inserted":
			fillColor = "#a855f7"
			strokeColor = "#d8b4fe"
		case "deleted":
			fillColor = "#ef4444"
			strokeColor = "#fca5a5"
		}

		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="20" fill="%s" stroke="%s" stroke-width="2" style="transition: all 0.25s ease;" />`,
			n.X, n.Y, fillColor, strokeColor,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#f0f6fc" font-size="12" font-family="ui-monospace, monospace" font-weight="600" text-anchor="middle" dy="4">%d</text>`,
			n.X, n.Y, n.Val,
		))
	}

	return sb.String()
}
