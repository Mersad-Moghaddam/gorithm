package renderers

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type SelectionLessonState struct {
	ArrayState
	IDs                                []int
	Pass, Scan, MinIndex, SortedPrefix int
	MinValue, MaxValue                 int
}

func RenderSelectionLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(SelectionLessonState)
	if width < 1 || width > 32768 {
		width = 800
	}
	if height < 1 || height > 32768 {
		height = 450
	}
	n := len(s.Values)
	width = max(width, 400, 64+n*112)
	height = max(height, 420)
	phase := "ready"
	switch step.LineID {
	case "L3":
		phase = "pass"
	case "L4", "L7":
		phase = "minimum"
	case "L6":
		phase = "compare"
	case "L10":
		phase = "already-minimum"
		if step.Action == trace.ActionSwap {
			phase = "swap"
		}
	case "L11":
		phase = "confirm"
	case "L12":
		phase = "done"
	}
	var out strings.Builder
	fmt.Fprintf(&out, `<g data-lesson="selection" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" fill="#eff5ed"><title>Selection Sort: scan for minimum, swap into sorted prefix</title>`, phase, width, height, width)
	if !ok || len(s.IDs) != n {
		out.WriteString(`<text x="32" y="48" font-size="16">No Selection Sort state available</text></g>`)
		return out.String()
	}
	if n == 0 {
		out.WriteString(`<text x="32" y="48" font-size="16">Empty array — nothing to sort</text></g>`)
		return out.String()
	}
	label := "Find minimum in unsorted range"
	if step.Action == trace.ActionDone {
		label = "Complete — all positions final"
	}
	if phase == "already-minimum" {
		label = "Already minimum — no swap"
	}
	fmt.Fprintf(&out, `<text x="32" y="32" font-size="16" font-weight="600">%s</text><text x="32" y="60" font-size="14" fill="#9bf0bb">Sorted prefix: %d ✓</text>`, label, s.SortedPrefix)
	low, high := math.Min(0, float64(s.MinValue)), math.Max(0, float64(s.MaxValue))
	if high <= low {
		high = low + 1
	}
	top, bottom := 176.0, float64(height-140)
	scale := (bottom - top) / (high - low)
	baseline := top + high*scale
	pitch := float64(width-64) / float64(n)
	barWidth := math.Min(80, pitch-24)
	xAt := func(i int) float64 { return 32 + (float64(i)+0.5)*pitch }
	fmt.Fprintf(&out, `<path d="M32 %.2fH%d" stroke="#bbcab9"/><text x="32" y="%d" font-size="14" fill="#bbcab9">Zero baseline</text>`, baseline, width-32, height-16)
	palette := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, value := range s.Values {
		id := s.IDs[i]
		if id < 0 {
			continue
		}
		x := xAt(i)
		tip := -float64(value) * scale
		y, bh := math.Min(0, tip), math.Abs(tip)
		if value == 0 {
			y, bh = -1, 2
		}
		stroke := "#bbcab9"
		if i < s.SortedPrefix {
			stroke = "#9bf0bb"
		}
		compare := step.Action == trace.ActionCompare && (i == s.Scan || i == s.MinIndex)
		if compare {
			stroke = "#ffe29a"
		}
		motion := ""
		if step.Action == trace.ActionSwap && len(s.Swapped) == 2 && s.Swapped[0] != s.Swapped[1] {
			if i == s.Swapped[0] {
				motion = ` data-motion="swap-up"`
			}
			if i == s.Swapped[1] {
				motion = ` data-motion="swap-down"`
			}
		}
		fmt.Fprintf(&out, `<g data-entity="main:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>Value %d, original identity %d, slot %d</title><rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="4" fill="%s" stroke="%s" stroke-width="3"/>`, id, x, baseline, x, baseline, motion, value, id, i, -barWidth/2, y, barWidth, bh, palette[id%5], stroke)
		if compare {
			fmt.Fprintf(&out, `<rect data-emphasis="compare" x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="4" fill="none" stroke="#ffe29a" stroke-width="3"/>`, -barWidth/2, y, barWidth, bh)
		}
		fmt.Fprintf(&out, `<text x="0" y="%.2f" text-anchor="middle" font-size="14" font-weight="600">%d</text><text x="0" y="%.2f" text-anchor="middle" font-size="12" fill="#bbcab9">#%d</text>`, y-24, value, y-8, id)
		if i < s.SortedPrefix {
			fmt.Fprintf(&out, `<text x="0" y="%.2f" text-anchor="middle" font-size="16" fill="#9bf0bb">✓</text>`, math.Max(0, tip)+22)
		}
		out.WriteString(`</g>`)
		fmt.Fprintf(&out, `<text x="%.2f" y="%d" text-anchor="middle" font-size="12" fill="#bbcab9">[%d]</text>`, x, height-112, i)
	}
	minimumLabel := "minimum"
	if step.Action == trace.ActionSwap {
		minimumLabel = "minIdx (from)"
	}
	for _, flag := range []struct {
		name     string
		index, y int
		color    string
	}{{"i", s.Pass, height - 88, "#8bd5ff"}, {"j", s.Scan, height - 64, "#ffe29a"}, {minimumLabel, s.MinIndex, height - 40, "#8bd5ff"}} {
		if flag.index >= 0 && flag.index < n {
			fmt.Fprintf(&out, `<text x="%.2f" y="%d" text-anchor="middle" font-size="14" fill="%s">↑ %s = %d</text>`, xAt(flag.index), flag.y, flag.color, flag.name, flag.index)
		}
	}
	if s.SortedPrefix > 0 && s.SortedPrefix < n {
		boundary := 32 + float64(s.SortedPrefix)*pitch
		fmt.Fprintf(&out, `<path d="M%.2f 100V%d" stroke="#9bf0bb" stroke-dasharray="5 5"/>`, boundary, height-108)
	}
	out.WriteString(`</g>`)
	return out.String()
}
