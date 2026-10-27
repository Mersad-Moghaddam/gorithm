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
		return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">No Hanoi state available</text>`
	}

	pegX := [3]float64{180, 400, 620}
	baseY := 360.0
	diskHeight := 26.0

	var sb strings.Builder

	// Defs for metallic peg and glowing disks
	sb.WriteString(`<defs>
		<linearGradient id="peg-grad" x1="0%" y1="0%" x2="100%" y2="0%">
			<stop offset="0%" stop-color="#475569" />
			<stop offset="50%" stop-color="#94a3b8" />
			<stop offset="100%" stop-color="#334155" />
		</linearGradient>
		<linearGradient id="base-grad" x1="0%" y1="0%" x2="0%" y2="100%">
			<stop offset="0%" stop-color="#334155" />
			<stop offset="100%" stop-color="#0f172a" />
		</linearGradient>
		<filter id="hanoi-glow" x="-20%" y="-20%" width="140%" height="140%">
			<feGaussianBlur in="SourceGraphic" stdDeviation="4" result="blur" />
			<feMerge>
				<feMergeNode in="blur" />
				<feMergeNode in="SourceGraphic" />
			</feMerge>
		</filter>
	</defs>`)

	// Base plate
	sb.WriteString(fmt.Sprintf(
		`<rect x="60" y="%.1f" width="680" height="20" rx="8" fill="url(#base-grad)" stroke="#475569" stroke-width="1.5" />`,
		baseY,
	))

	// Peg shafts and labels
	for p := range 3 {
		sb.WriteString(fmt.Sprintf(
			`<rect x="%.1f" y="130" width="10" height="230" rx="5" fill="url(#peg-grad)" />`,
			pegX[p]-5,
		))
		label := fmt.Sprintf("Peg %c", 'A'+p)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#94a3b8" font-size="13" font-family="ui-monospace, Menlo, monospace" font-weight="700" text-anchor="middle">%s</text>`,
			pegX[p], baseY+40, label,
		))
	}

	colors := []string{
		"#ef4444", "#f97316", "#eab308", "#22c55e", "#06b6d4", "#3b82f6", "#a855f7",
	}

	for p := range 3 {
		for level, diskSize := range state.Pegs[p] {
			fmt.Fprintf(&sb, `<g data-key="disk-%d"><title>Disk %d on peg %c</title>`, diskSize, diskSize, 'A'+p)
			diskW := 50.0 + float64(diskSize)*28.0
			x := pegX[p] - diskW/2
			y := baseY - float64(level+1)*diskHeight

			color := colors[(diskSize-1)%len(colors)]
			sb.WriteString(fmt.Sprintf(
				`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="7" fill="%s" stroke="#ffffff" stroke-width="1" stroke-opacity="0.4" style="transition: x 0.25s ease, y 0.25s ease, width 0.25s ease;" />`,
				x, y, diskW, diskHeight-4, color,
			))
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#ffffff" font-size="12" font-family="ui-monospace, Menlo, monospace" font-weight="800" text-anchor="middle">%d</text>`,
				pegX[p], y+15, diskSize,
			))
			sb.WriteString(`</g>`)
		}
	}

	return sb.String()
}
