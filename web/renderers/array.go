package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type ArrayState struct {
	Values    []int `json:"values"`
	Comparing []int `json:"comparing"`
	Swapped   []int `json:"swapped"`
	Sorted    []int `json:"sorted"`
	Current   int   `json:"current"`
}

func RenderArraySVG(step trace.Step, width, height int) string {
	state, ok := step.State.(ArrayState)
	if !ok {
		if m, isMap := step.State.(map[string]any); isMap {
			state = parseArrayState(m)
		} else {
			return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">No visual state available</text>`
		}
	}

	n := len(state.Values)
	if n == 0 {
		return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">Empty array</text>`
	}

	maxVal := 1
	for _, v := range state.Values {
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder

	// Shared defs for gradients and glow
	sb.WriteString(`<defs>
		<linearGradient id="bar-default" x1="0%" y1="0%" x2="0%" y2="100%">
			<stop offset="0%" stop-color="#38bdf8" />
			<stop offset="100%" stop-color="#0284c7" />
		</linearGradient>
		<linearGradient id="bar-comp" x1="0%" y1="0%" x2="0%" y2="100%">
			<stop offset="0%" stop-color="#fde047" />
			<stop offset="100%" stop-color="#eab308" />
		</linearGradient>
		<linearGradient id="bar-swap" x1="0%" y1="0%" x2="0%" y2="100%">
			<stop offset="0%" stop-color="#f87171" />
			<stop offset="100%" stop-color="#dc2626" />
		</linearGradient>
		<linearGradient id="bar-sorted" x1="0%" y1="0%" x2="0%" y2="100%">
			<stop offset="0%" stop-color="#4ade80" />
			<stop offset="100%" stop-color="#16a34a" />
		</linearGradient>
		<linearGradient id="bar-current" x1="0%" y1="0%" x2="0%" y2="100%">
			<stop offset="0%" stop-color="#c084fc" />
			<stop offset="100%" stop-color="#9333ea" />
		</linearGradient>
		<filter id="neon-glow" x="-20%" y="-20%" width="140%" height="140%">
			<feGaussianBlur stdDeviation="5" result="blur" />
			<feMerge>
				<feMergeNode in="blur" />
				<feMergeNode in="SourceGraphic" />
			</feMerge>
		</filter>
	</defs>`)

	padX := 60
	padY := 70
	usableW := width - 2*padX
	usableH := height - 2*padY

	barWidth := float64(usableW) / float64(n)
	spacing := barWidth * 0.18
	if spacing < 4 {
		spacing = 4
	}
	if spacing > 16 {
		spacing = 16
	}
	actualBarW := barWidth - spacing

	// Mark sets
	isComp := make(map[int]bool)
	for _, idx := range state.Comparing {
		isComp[idx] = true
	}
	isSwap := make(map[int]bool)
	for _, idx := range state.Swapped {
		isSwap[idx] = true
	}
	isSort := make(map[int]bool)
	for _, idx := range state.Sorted {
		isSort[idx] = true
	}

	// Floor baseline
	baseY := float64(height - padY)
	sb.WriteString(fmt.Sprintf(
		`<line x1="%d" y1="%.1f" x2="%d" y2="%.1f" stroke="#334155" stroke-width="1.5" stroke-dasharray="4 4" opacity="0.6" />`,
		padX-10, baseY+2, width-padX+10, baseY+2,
	))

	for i, v := range state.Values {
		barH := (float64(v) / float64(maxVal)) * float64(usableH)
		if barH < 18 {
			barH = 18
		}
		x := float64(padX) + float64(i)*barWidth + spacing/2
		y := baseY - barH

		fillRef := "url(#bar-default)"
		strokeColor := "rgba(56, 189, 248, 0.4)"
		filterStr := ""

		if isSort[i] {
			fillRef = "url(#bar-sorted)"
			strokeColor = "#86efac"
		} else if isSwap[i] {
			fillRef = "url(#bar-swap)"
			strokeColor = "#fca5a5"
			filterStr = `filter="url(#neon-glow)"`
		} else if isComp[i] {
			fillRef = "url(#bar-comp)"
			strokeColor = "#fef08a"
			filterStr = `filter="url(#neon-glow)"`
		} else if i == state.Current {
			fillRef = "url(#bar-current)"
			strokeColor = "#e9d5ff"
		}

		// Bar container with smooth transitions
		sb.WriteString(fmt.Sprintf(
			`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="6" fill="%s" stroke="%s" stroke-width="1.5" %s style="transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);" />`,
			x, y, actualBarW, barH, fillRef, strokeColor, filterStr,
		))

		// Value label on top of bar
		centerX := x + actualBarW/2
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#f8fafc" font-size="12" font-family="ui-monospace, Menlo, monospace" font-weight="700" text-anchor="middle">%d</text>`,
			centerX, y-10, v,
		))

		// Array Index badge under bar
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#64748b" font-size="11" font-family="ui-monospace, Menlo, monospace" font-weight="500" text-anchor="middle">[%d]</text>`,
			centerX, baseY+20, i,
		))
	}

	return sb.String()
}

func parseArrayState(m map[string]any) ArrayState {
	state := ArrayState{Current: -1}
	if rawVals, ok := m["values"].([]any); ok {
		for _, v := range rawVals {
			if num, ok := v.(float64); ok {
				state.Values = append(state.Values, int(num))
			}
		}
	}
	if rawComp, ok := m["comparing"].([]any); ok {
		for _, v := range rawComp {
			if num, ok := v.(float64); ok {
				state.Comparing = append(state.Comparing, int(num))
			}
		}
	}
	if rawSwap, ok := m["swapped"].([]any); ok {
		for _, v := range rawSwap {
			if num, ok := v.(float64); ok {
				state.Swapped = append(state.Swapped, int(num))
			}
		}
	}
	if rawSort, ok := m["sorted"].([]any); ok {
		for _, v := range rawSort {
			if num, ok := v.(float64); ok {
				state.Sorted = append(state.Sorted, int(num))
			}
		}
	}
	if cur, ok := m["current"].(float64); ok {
		state.Current = int(cur)
	}
	return state
}
