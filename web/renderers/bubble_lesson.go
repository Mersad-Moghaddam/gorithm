package renderers

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type BubbleLessonState struct {
	ArrayState
	IDs                     []int
	Pass, Scan, UnsortedEnd int
	MinValue, MaxValue      int
}

func RenderBubbleLessonSVG(step trace.Step, width, height int) string {
	state, ok := step.State.(BubbleLessonState)
	if width < 320 || width > 32768 {
		width = 800
	}
	if height < 360 || height > 32768 {
		height = 450
	}
	n := len(state.Values)
	minimum := max(320, 64+n*100)
	width = max(width, minimum)
	minimum = width
	phase := "ready"
	switch step.Action {
	case trace.ActionCompare:
		phase = "compare"
	case trace.ActionSwap:
		phase = "swap"
	case trace.ActionDone:
		phase = "done"
	case trace.ActionHighlight:
		phase = "pass"
	case trace.ActionSet:
		if len(state.Sorted) > 0 {
			phase = "confirm"
		}
	}
	var out strings.Builder
	fmt.Fprintf(&out, `<g data-lesson="bubble" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" fill="#eff5ed"><title>Bubble Sort: adjacent comparisons, swaps, and sorted suffix</title>`, phase, width, height, minimum)
	if !ok || len(state.IDs) != n {
		out.WriteString(`<text x="32" y="48" font-size="16">No Bubble Sort state available</text></g>`)
		return out.String()
	}
	if n == 0 {
		out.WriteString(`<text x="32" y="48" font-size="16">Empty array — nothing to sort</text></g>`)
		return out.String()
	}
	label := "Ready to sort"
	if state.Pass >= 0 {
		label = fmt.Sprintf("Pass %d", state.Pass+1)
	}
	if step.Action == trace.ActionDone {
		label = "Complete — all positions final"
	}
	fmt.Fprintf(&out, `<text x="32" y="32" font-size="16" font-weight="600">%s</text>`, label)
	low, high := float64(state.MinValue), float64(state.MaxValue)
	if low >= high {
		high = low + 1
	}
	top, bottom := 148.0, float64(height-124)
	scale := (bottom - top) / (high - low)
	baseline := top + high*scale
	pitch := float64(width-64) / float64(n)
	barWidth := math.Min(72, pitch-24)
	fmt.Fprintf(&out, `<path d="M32 %.2fH%d" stroke="#bbcab9"/><text x="32" y="%.2f" font-size="14" fill="#bbcab9">Zero</text>`, baseline, width-32, float64(height-20))
	palette := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, value := range state.Values {
		id := state.IDs[i]
		if id < 0 {
			continue
		}
		x := 32 + (float64(i)+0.5)*pitch
		tip := -float64(value) * scale
		y, barHeight := math.Min(0, tip), math.Abs(tip)
		if value == 0 {
			y, barHeight = -1, 2
		}
		stroke := "#bbcab9"
		sorted := false
		for _, slot := range state.Sorted {
			if slot == i {
				sorted = true
				stroke = "#9bf0bb"
			}
		}
		comparing := step.Action == trace.ActionCompare && (i == state.Scan || i == state.Scan+1) && state.Scan >= 0
		if comparing {
			stroke = "#ffe29a"
		}
		motion := ""
		if step.Action == trace.ActionSwap && state.Scan >= 0 && state.Scan+1 < n {
			if i == state.Scan {
				motion = ` data-motion="swap-down"`
			}
			if i == state.Scan+1 {
				motion = ` data-motion="swap-up"`
			}
		}
		fmt.Fprintf(&out, `<g data-entity="main:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>Value %d, original identity %d, slot %d</title><rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="4" fill="%s" stroke="%s" stroke-width="3"/>`, id, x, baseline, x, baseline, motion, value, id, i, -barWidth/2, y, barWidth, barHeight, palette[id%len(palette)], stroke)
		if comparing {
			fmt.Fprintf(&out, `<rect data-emphasis="compare" x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="4" fill="none" stroke="#ffe29a" stroke-width="3"/>`, -barWidth/2, y, barWidth, barHeight)
		}
		fmt.Fprintf(&out, `<text x="0" y="%.2f" text-anchor="middle" font-size="14" font-weight="600">%d</text><text x="0" y="%.2f" text-anchor="middle" font-size="12" fill="#bbcab9">#%d</text>`, y-22, value, y-7, id)
		if sorted {
			fmt.Fprintf(&out, `<text x="0" y="%.2f" text-anchor="middle" font-size="16" fill="#9bf0bb">✓</text>`, math.Max(0, tip)+22)
		}
		out.WriteString(`</g>`)
		fmt.Fprintf(&out, `<text x="%.2f" y="%d" text-anchor="middle" font-size="12" fill="#bbcab9">[%d]</text>`, x, height-56, i)
	}
	if state.Scan >= 0 && state.Scan+1 < n && (step.Action == trace.ActionCompare || step.Action == trace.ActionSwap) {
		x1 := 32 + (float64(state.Scan)+0.5)*pitch
		fmt.Fprintf(&out, `<path d="M%.2f %dv8h%.2fv-8" fill="none" stroke="#ffe29a" stroke-width="3"/><text x="%.2f" y="64" font-size="14" text-anchor="middle">Adjacent pair [%d, %d]</text>`, x1, height-92, pitch, x1+pitch/2, state.Scan, state.Scan+1)
	}
	if state.UnsortedEnd < n-1 {
		boundary := 32 + float64(max(0, state.UnsortedEnd+1))*pitch
		fmt.Fprintf(&out, `<path d="M%.2f 112V%d" stroke="#9bf0bb" stroke-dasharray="5 5"/><text x="%d" y="%d" text-anchor="end" font-size="14" fill="#9bf0bb">Sorted suffix ✓</text>`, boundary, height-104, width-32, height-20)
	}
	out.WriteString(`</g>`)
	return out.String()
}
