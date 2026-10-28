package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type QuickLessonRange struct{ Low, High int }

type QuickLessonState struct {
	ArrayState
	IDs                                                        []int
	Stack                                                      []QuickLessonRange
	Low, High, Boundary, Scan, PivotIndex, PivotID, PivotValue int
	PivotActive                                                bool
}

func RenderQuickLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(QuickLessonState)
	if width < 320 || width > 32768 {
		width = 800
	}
	if height < 360 || height > 32768 {
		height = 450
	}
	n := len(s.Values)
	// Keep eight-digit signed values readable at 1:1; stage owns local scrolling.
	w := math.Max(640, math.Max(float64(width), float64(n)*100+64))
	h := float64(height)
	var out strings.Builder
	fmt.Fprintf(&out, `<g data-lesson="quick" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui, sans-serif"><title>%s</title>`, html.EscapeString(step.LineID), w, h, w, html.EscapeString("Quick Sort: "+step.Explanation))
	text := func(x, y float64, value string, color string, size int) {
		fmt.Fprintf(&out, `<text x="%.2f" y="%.2f" fill="%s" font-size="%d" text-anchor="middle">%s</text>`, x, y, color, size, html.EscapeString(value))
	}
	if !ok || n == 0 {
		text(w/2, h/2, "Empty array — no partition needed", "#eff5ed", 16)
		out.WriteString(`</g>`)
		return out.String()
	}
	minV, maxV := 0.0, 0.0
	for _, v := range s.Values {
		minV = math.Min(minV, float64(v))
		maxV = math.Max(maxV, float64(v))
	}
	if minV == maxV {
		maxV = minV + 1
	}
	top, bottom := 140.0, h-100
	y := func(v float64) float64 { return bottom - (v-minV)/(maxV-minV)*(bottom-top) }
	baseline := y(0)
	spacing := (w - 64) / float64(n)
	x := func(i int) float64 { return 32 + (float64(i)+0.5)*spacing }
	fmt.Fprintf(&out, `<path d="M32 %.2f H%.2f" stroke="#bbcab9"/>`, baseline, w-32)
	text(w/2, 24, fmt.Sprintf("Range [%d..%d] · recursion depth %d", s.Low, s.High, len(s.Stack)), "#eff5ed", 16)
	if s.Low > s.High {
		text(w/2, 50, "Empty range: base case", "#bbcab9", 14)
	} else if s.Low >= 0 && s.High < n {
		fmt.Fprintf(&out, `<path d="M%.2f 70 V60 H%.2f V70" fill="none" stroke="#8bd5ff" stroke-width="2"/>`, x(s.Low)-spacing*.4, x(s.High)+spacing*.4)
		if s.PivotActive {
			text(w/2, 48, fmt.Sprintf("pivot %d#%d · boundary i=%d · scan j=%d", s.PivotValue, s.PivotID, s.Boundary, s.Scan), "#ffe29a", 14)
		}
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	contains := func(list []int, v int) bool {
		for _, item := range list {
			if item == v {
				return true
			}
		}
		return false
	}
	for i, v := range s.Values {
		if i >= len(s.IDs) || s.IDs[i] < 0 {
			continue
		}
		id := s.IDs[i]
		motion := ""
		if step.Action == trace.ActionSwap && len(s.Swapped) == 2 && s.Swapped[0] != s.Swapped[1] {
			if i == s.Swapped[0] {
				motion = ` data-motion="swap-up"`
			} else if i == s.Swapped[1] {
				motion = ` data-motion="swap-down"`
			}
		}
		fmt.Fprintf(&out, `<g data-entity="main:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s>`, id, x(i), baseline, x(i), baseline, motion)
		stroke := "#bbcab9"
		if contains(s.Comparing, i) {
			stroke = "#ffe29a"
		}
		if contains(s.Sorted, i) {
			stroke = "#9bf0bb"
		}
		barY := math.Min(y(float64(v))-baseline, 0)
		barH := math.Max(2, math.Abs(y(float64(v))-baseline))
		fmt.Fprintf(&out, `<rect x="-32" y="%.2f" width="64" height="%.2f" rx="5" fill="%s" stroke="%s" stroke-width="3"/>`, barY, barH, palette[id%5], stroke)
		text(0, barY-10, fmt.Sprintf("%d", v), "#eff5ed", 16)
		text(0, bottom-baseline+26, fmt.Sprintf("#%d [%d]", id, i), "#bbcab9", 14)
		if contains(s.Sorted, i) {
			text(0, barY-32, "✓", "#9bf0bb", 16)
		}
		if s.PivotActive && id == s.PivotID {
			text(0, 88-baseline, "pivot", "#ffe29a", 14)
		}
		out.WriteString(`</g>`)
	}
	if s.PivotActive && step.LineID != "L9" {
		scannedEnd := s.Scan - 1
		boundary := s.Boundary
		if step.LineID == "L12" && s.Scan >= 0 && s.Scan < n && s.Values[s.Scan] >= s.PivotValue || step.LineID == "L14" {
			scannedEnd = s.Scan
		}
		if step.LineID == "L13" {
			boundary--
		}
		if s.Scan == s.High {
			scannedEnd = s.High - 1
		}
		if s.Scan < 0 {
			scannedEnd = s.Low - 1
		}
		greaterStart := boundary + 1
		if s.Scan == s.High {
			greaterStart++
		}
		text(w/2, h-42, fmt.Sprintf("< pivot: [%d..%d] · scanned ≥ pivot: [%d..%d]", s.Low, boundary, greaterStart, scannedEnd), "#bbcab9", 14)
		unscanned := scannedEnd + 1
		if s.Scan < 0 {
			unscanned = s.Low
		}
		text(w/2, h-20, fmt.Sprintf("Unscanned: [%d..%d] · pivot slot [%d]", unscanned, s.High-1, s.PivotIndex), "#bbcab9", 14)
	}
	out.WriteString(`</g>`)
	return out.String()
}
