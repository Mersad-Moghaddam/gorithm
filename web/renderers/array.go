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
		// If raw map from JSON unmarshal
		if m, isMap := step.State.(map[string]any); isMap {
			state = parseArrayState(m)
		} else {
			return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No visual state</text>`
		}
	}

	n := len(state.Values)
	if n == 0 {
		return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">Empty array</text>`
	}

	maxVal := 1
	for _, v := range state.Values {
		if v > maxVal {
			maxVal = v
		}
	}

	var sb strings.Builder
	padX := 50
	padY := 60
	usableW := width - 2*padX
	usableH := height - 2*padY

	barWidth := float64(usableW) / float64(n)
	spacing := barWidth * 0.15
	if spacing < 2 {
		spacing = 2
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

	for i, v := range state.Values {
		barH := (float64(v) / float64(maxVal)) * float64(usableH)
		if barH < 12 {
			barH = 12
		}
		x := float64(padX) + float64(i)*barWidth
		y := float64(height) - float64(padY) - barH

		fillColor := "#38bdf8" // default cyan
		strokeColor := "none"
		strokeWidth := 0

		if isSort[i] {
			fillColor = "#22c55e" // green sorted
		} else if isSwap[i] {
			fillColor = "#ef4444" // red swapped
		} else if isComp[i] {
			fillColor = "#eab308" // yellow comparing
			strokeColor = "#fef08a"
			strokeWidth = 2
		} else if i == state.Current {
			fillColor = "#a855f7" // purple current
		}

		// Bar rect
		if strokeWidth > 0 {
			sb.WriteString(fmt.Sprintf(
				`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="%s" stroke="%s" stroke-width="%d" style="transition: all 0.25s ease;" />`,
				x, y, actualBarW, barH, fillColor, strokeColor, strokeWidth,
			))
		} else {
			sb.WriteString(fmt.Sprintf(
				`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="4" fill="%s" style="transition: all 0.25s ease;" />`,
				x, y, actualBarW, barH, fillColor,
			))
		}

		// Value label
		centerX := x + actualBarW/2
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#f0f6fc" font-size="12" font-family="ui-monospace, monospace" font-weight="600" text-anchor="middle">%.d</text>`,
			centerX, y-8, v,
		))

		// Index label
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#6e7681" font-size="11" font-family="ui-monospace, monospace" text-anchor="middle">[%d]</text>`,
			centerX, float64(height)-float64(padY)+20, i,
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
