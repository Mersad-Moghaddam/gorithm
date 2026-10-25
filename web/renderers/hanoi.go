package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type HanoiState struct {
	Pegs [3][]int `json:"pegs"` // disks on peg 0, 1, 2
	Move string   `json:"move"`
}

func RenderHanoiSVG(step trace.Step, width, height int) string {
	state, ok := step.State.(HanoiState)
	if !ok {
		return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No Hanoi state</text>`
	}

	pegX := [3]float64{180, 400, 620}
	baseY := 360.0
	diskHeight := 24.0

	var sb strings.Builder

	// Base plate
	sb.WriteString(fmt.Sprintf(
		`<rect x="80" y="%.1f" width="640" height="16" rx="4" fill="#30363d" />`,
		baseY,
	))

	// Peg shafts and labels
	for p := range 3 {
		sb.WriteString(fmt.Sprintf(
			`<rect x="%.1f" y="140" width="8" height="220" rx="3" fill="#484f58" />`,
			pegX[p]-4,
		))
		label := fmt.Sprintf("Peg %c", 'A'+p)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="13" font-family="ui-monospace, monospace" font-weight="600" text-anchor="middle">%s</text>`,
			pegX[p], baseY+36, label,
		))
	}

	colors := []string{
		"#ef4444", "#f97316", "#eab308", "#22c55e", "#06b6d4", "#3b82f6", "#a855f7",
	}

	for p := range 3 {
		for level, diskSize := range state.Pegs[p] {
			diskW := 40.0 + float64(diskSize)*24.0
			x := pegX[p] - diskW/2
			y := baseY - float64(level+1)*diskHeight

			color := colors[(diskSize-1)%len(colors)]
			sb.WriteString(fmt.Sprintf(
				`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="5" fill="%s" stroke="#fff" stroke-width="0.5" style="transition: all 0.25s ease;" />`,
				x, y, diskW, diskHeight-3, color,
			))
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#fff" font-size="11" font-family="ui-monospace, monospace" font-weight="700" text-anchor="middle">%d</text>`,
				pegX[p], y+14, diskSize,
			))
		}
	}

	return sb.String()
}
