package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"math"
	"strings"
)

type MergeLessonRange struct{ Left, Mid, Right int }
type MergeLessonState struct {
	ArrayState
	IDs, Aux, AuxIDs                     []int
	AuxFilled                            []bool
	Stack                                []MergeLessonRange
	Left, Mid, Right, I, J, K, CopyIndex int
	MinValue, MaxValue                   int
}

func RenderMergeLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(MergeLessonState)
	if width < 320 || width > 32768 {
		width = 800
	}
	if height < 360 || height > 32768 {
		height = 450
	}
	width = max(width, max(480, 96+len(s.Values)*144))
	height = max(height, 544+len(s.Stack)*24)
	phase := "control"
	switch step.LineID {
	case "L14":
		phase = "compare"
	case "L15", "L17", "L21", "L22":
		phase = "to-aux"
	case "L23":
		phase = "copy-back"
	case "L24":
		phase = "merged"
	case "L9":
		phase = "done"
	}
	var out strings.Builder
	fmt.Fprintf(&out, `<g data-lesson="merge" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" fill="#eff5ed"><title>Merge Sort: recursive ranges, stable merge, and auxiliary copies</title>`, phase, width, height, width)
	if !ok || len(s.IDs) != len(s.Values) {
		out.WriteString(`<text x="32" y="40" font-size="16">No Merge Sort state available</text></g>`)
		return out.String()
	}
	if len(s.Values) == 0 {
		out.WriteString(`<text x="32" y="40" font-size="16">Empty array — nothing to sort</text></g>`)
		return out.String()
	}
	fmt.Fprintf(&out, `<text x="32" y="28" font-size="16">Recursion stack (caller to current)</text>`)
	for i, r := range s.Stack {
		fmt.Fprintf(&out, `<text x="32" y="%d" font-size="14">%d: [%d..%d] · mid=%d</text>`, 52+i*24, i, r.Left, r.Right, r.Mid)
	}
	rowTop := float64(100 + len(s.Stack)*24)
	rowBottom := rowTop + 160
	auxY := rowBottom + 152
	low, high := math.Min(0, float64(s.MinValue)), math.Max(0, float64(s.MaxValue))
	if low == high {
		high = low + 1
	}
	scale := 160 / (high - low)
	baseline := rowTop + high*scale
	pitch := float64(width-64) / float64(len(s.Values))
	xpos := func(i int) float64 { return 32 + (float64(i)+0.5)*pitch }
	fmt.Fprintf(&out, `<text x="32" y="%.2f" font-size="14">Main array · zero baseline</text><path d="M32 %.2fH%d" stroke="#bbcab9"/>`, rowTop-40, baseline, width-32)
	palette := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, v := range s.Values {
		id := s.IDs[i]
		if id < 0 {
			continue
		}
		tip := -float64(v) * scale
		y, h := math.Min(0, tip), math.Abs(tip)
		if v == 0 {
			y, h = -1, 2
		}
		stroke := "#bbcab9"
		if step.Action == trace.ActionDone {
			stroke = "#9bf0bb"
		}
		if phase == "compare" && (i == s.I || i == s.J) {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&out, `<g data-entity="main:%d:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>Value %d, original identity %d, main slot %d</title><rect x="-42" y="%.2f" width="84" height="%.2f" rx="4" fill="%s" stroke="%s" stroke-width="3"/><text x="0" y="%.2f" text-anchor="middle" font-size="14">%d · #%d</text>`, i, id, xpos(i), baseline, xpos(i), baseline, v, id, i, y, h, palette[id%5], stroke, y-12, v, id)
		if phase == "compare" && (i == s.I || i == s.J) {
			fmt.Fprintf(&out, `<rect data-emphasis="compare" x="-42" y="%.2f" width="84" height="%.2f" fill="none" stroke="#ffe29a" stroke-width="3"/>`, y, h)
		}
		out.WriteString(`</g>`)
		check := ""
		if step.Action == trace.ActionDone {
			check = " ✓"
		}
		fmt.Fprintf(&out, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="12">[%d]%s</text>`, xpos(i), rowBottom+20, i, check)
	}
	if step.LineID == "L23" && s.CopyIndex >= 0 && s.CopyIndex < len(s.AuxIDs) && s.Left+s.CopyIndex >= 0 && s.Left+s.CopyIndex < len(s.Values) {
		k := s.CopyIndex
		auxID := s.AuxIDs[k]
		destIdx := s.Left + k
		destX := xpos(destIdx)
		sourceKey := fmt.Sprintf("aux:%d:%d", k, auxID)
		incomingKey := fmt.Sprintf("incoming:%d:%d", k, auxID)
		fmt.Fprintf(&out, `<g data-entity="%s" data-from-entity="%s" data-motion="linear" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>Copy aux[%d] write arrival</title><text x="0" y="24" text-anchor="middle" font-size="12" fill="#8bd5ff">Copy aux[%d] write arrival</text></g>`, incomingKey, sourceKey, destX, baseline, destX, baseline, k, k)
	}
	if s.Left >= 0 && s.Left <= s.Mid && s.Mid < s.Right && s.Right < len(s.Values) {
		for half, r := range [][2]int{{s.Left, s.Mid}, {s.Mid + 1, s.Right}} {
			a, z := xpos(r[0])-44, xpos(r[1])+44
			label := "Left"
			if half == 1 {
				label = "Right"
			}
			fmt.Fprintf(&out, `<path d="M%.2f %.2fv8H%.2fv-8" fill="none" stroke="#ffe29a" stroke-width="2"/><text x="%.2f" y="%.2f" text-anchor="middle" font-size="14">%s [%d..%d]</text>`, a, rowBottom+30, z, (a+z)/2, rowBottom+60, label, r[0], r[1])
		}
	}
	if s.I >= s.Left && s.I <= s.Mid && s.I >= 0 && s.I < len(s.Values) {
		fmt.Fprintf(&out, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="14" fill="#8bd5ff">↑ i</text>`, xpos(s.I), rowBottom+84)
	}
	if s.J >= s.Mid+1 && s.J <= s.Right && s.J >= 0 && s.J < len(s.Values) {
		fmt.Fprintf(&out, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="14" fill="#8bd5ff">↑ j</text>`, xpos(s.J), rowBottom+84)
	}
	fmt.Fprintf(&out, `<text x="32" y="%.2f" font-size="14">Auxiliary storage · i=%d · j=%d · k=%d</text>`, auxY-32, s.I, s.J, s.K)
	for k := range s.Aux {
		if s.Left+k < 0 || s.Left+k >= len(s.Values) {
			continue
		}
		x := xpos(s.Left + k)
		filled := k < len(s.AuxFilled) && s.AuxFilled[k] && k < len(s.AuxIDs) && s.AuxIDs[k] >= 0
		if !filled {
			fmt.Fprintf(&out, `<rect x="%.2f" y="%.2f" width="84" height="42" rx="5" fill="none" stroke="#bbcab9" stroke-dasharray="4 4"/><text x="%.2f" y="%.2f" text-anchor="middle" font-size="14">empty</text>`, x-42, auxY-21, x, auxY+5)
			continue
		}
		id := s.AuxIDs[k]
		alias := ""
		if phase == "to-aux" && k == s.K && s.Current >= 0 && s.Current < len(s.Values) {
			alias = fmt.Sprintf(` data-from-entity="main:%d:%d" data-motion="linear"`, s.Current, id)
		}
		fmt.Fprintf(&out, `<g data-entity="aux:%d:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s><title>Auxiliary slot %d, copied identity %d</title><rect x="-42" y="-21" width="84" height="42" rx="5" fill="%s" stroke="#8bd5ff" stroke-width="2"/><text x="0" y="5" text-anchor="middle" font-size="14">%d · #%d</text></g>`, k, id, x, auxY, x, auxY, alias, k, id, palette[id%5], s.Aux[k], id)
	}
	if s.K >= 0 && s.K < len(s.Aux) && s.Left+s.K >= 0 && s.Left+s.K < len(s.Values) {
		fmt.Fprintf(&out, `<text x="%.2f" y="%.2f" text-anchor="middle" font-size="14" fill="#8bd5ff">↑ k</text>`, xpos(s.Left+s.K), auxY+48)
	}
	if phase == "merged" {
		fmt.Fprintf(&out, `<text x="32" y="%d" font-size="14" fill="#9bf0bb">✓ Local range [%d..%d] merged; caller may still need work</text>`, height-24, s.Left, s.Right)
	}
	out.WriteString(`</g>`)
	return out.String()
}
