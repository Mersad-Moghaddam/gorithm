package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"strconv"
	"strings"
)

type RadixsortLessonState struct {
	ArrayState
	IDs, Output, OutputIDs                        []int
	OutputFilled                                  []bool
	Counts                                        [10]int
	Exp, MaxValue, Scan, Digit, Target, CopyIndex int
	Phase                                         string
}

func RenderRadixsortLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(RadixsortLessonState)
	if !ok {
		return `<g data-lesson="radixsort" data-phase="unavailable"><title>Radix sort state unavailable</title></g>`
	}
	n := len(s.Values)
	w := max(720, 64+n*92)
	w = max(w, min(width, 32768))
	h := max(450, min(height, 32768))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="radixsort" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Radix sort: %s; exponent %d. Colors and original IDs follow items; output is a copy.</title><text x="32" y="28">Radix Sort · exp = %d · %s</text>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(s.Phase), s.Exp, s.Exp, html.EscapeString(s.Phase))
	if n == 0 {
		b.WriteString(`<text x="32" y="70">Empty array</text></g>`)
		return b.String()
	}
	b.WriteString(`<text x="32" y="58">Input · digit = (value / exp) % 10</text>`)
	scale := float64(max(1, s.MaxValue))
	slot := float64(w-64) / float64(n)
	barW := min(72.0, slot-16)
	for i, v := range s.Values {
		if i >= len(s.IDs) {
			continue
		}
		id := s.IDs[i]
		// Slot distinguishes simultaneous real copies of one logical item during L17.
		key := fmt.Sprintf("main:%d:slot:%d", id, i)
		x := 32 + slot*(float64(i)+.5)
		y := 210.0
		barH := max(2.0, float64(v)/scale*105)
		stroke := "#bbcab9"
		if i == s.Current {
			stroke = "#ffe29a"
		}
		if s.Phase == "done" {
			stroke = "#9bf0bb"
		}
		alias := ""
		if s.Phase == "copy" && i == s.CopyIndex {
			alias = fmt.Sprintf(` data-from-entity="output:%d"`, id)
		}
		fmt.Fprintf(&b, `<g data-entity="%s" data-id="%d" data-role="main" data-slot="%d" data-operation="%s" data-motion="linear"%s data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="6" fill="%s" stroke="%s" stroke-width="3"/>`, key, id, i, html.EscapeString(s.Phase), alias, x, y, x, y, -barW/2, -barH, barW, barH, radixsortLessonColor(id), stroke)
		fmt.Fprintf(&b, `<text x="0" y="%.2f" text-anchor="middle">`, -barH-12)
		label := strconv.Itoa(v)
		place := len(label) - 1
		for exp := s.Exp; exp > 1; exp /= 10 {
			place--
		}
		for j, c := range label {
			if s.Exp > 0 && j == place {
				fmt.Fprintf(&b, `<tspan text-decoration="underline" font-weight="700">%c</tspan>`, c)
			} else {
				fmt.Fprintf(&b, `<tspan>%c</tspan>`, c)
			}
		}
		b.WriteString(`</text>`)
		fmt.Fprintf(&b, `<text x="0" y="22" text-anchor="middle">#%d [%d]</text>`, id, i)
		if s.Exp > 0 && place < 0 && s.Phase != "done" {
			b.WriteString(`<text x="0" y="42" text-anchor="middle" fill="#ffe29a">digit 0</text>`)
		}
		if s.Phase == "done" {
			b.WriteString(`<text x="0" y="42" text-anchor="middle" fill="#9bf0bb">✓</text>`)
		}
		b.WriteString(`</g>`)
	}
	fmt.Fprintf(&b, `<path d="M32 210 H%d" stroke="#bbcab9"/><text x="32" y="278">Counters 0–9 · %s</text>`, w-32, radixsortLessonCounterLabel(s.Phase, s.Exp))
	for digit, c := range s.Counts {
		x := 32 + float64(digit)*float64(w-64)/10
		cw := float64(w-64)/10 - 8
		stroke := "#bbcab9"
		if digit == s.Digit {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&b, `<rect x="%.2f" y="290" width="%.2f" height="46" rx="6" fill="#192923" stroke="%s"/><text x="%.2f" y="308" text-anchor="middle">%d</text><text x="%.2f" y="328" text-anchor="middle">%d</text>`, x, cw, stroke, x+cw/2, digit, x+cw/2, c)
	}
	b.WriteString(`<text x="32" y="365">Output · stable reverse placement · unwritten slots marked —</text>`)
	for i := 0; i < n; i++ {
		x := 32 + slot*(float64(i)+.5)
		if i >= len(s.OutputFilled) || !s.OutputFilled[i] || i >= len(s.Output) || i >= len(s.OutputIDs) {
			fmt.Fprintf(&b, `<rect x="%.2f" y="380" width="%.2f" height="42" rx="6" fill="none" stroke="#bbcab9" stroke-dasharray="4 4"/><text x="%.2f" y="406" text-anchor="middle">—</text>`, x-barW/2, barW, x)
			continue
		}
		id := s.OutputIDs[i]
		sourceSlot := -1
		for slot, sourceID := range s.IDs {
			if sourceID == id {
				sourceSlot = slot
				break
			}
		}
		fmt.Fprintf(&b, `<g data-entity="output:%d" data-id="%d" data-role="output" data-slot="%d" data-operation="copy" data-motion="linear" data-from-entity="main:%d:slot:%d" data-x="%.2f" data-y="401" transform="translate(%.2f 401)"><rect x="%.2f" y="-21" width="%.2f" height="42" rx="6" fill="%s" stroke="#8bd5ff"/><text x="0" y="-2" text-anchor="middle">%d</text><text x="0" y="15" text-anchor="middle">#%d</text></g>`, id, id, i, id, sourceSlot, x, x, -barW/2, barW, radixsortLessonColor(id), s.Output[i], id)
	}
	b.WriteString(`</g>`)
	return b.String()
}

func radixsortLessonColor(id int) string {
	colors := [5]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	if id < 0 {
		return "#bbcab9"
	}
	return colors[id%5]
}
func radixsortLessonCounterLabel(phase string, exp int) string {
	if phase == "done" && exp == 0 {
		return "digit frequencies"
	}
	switch phase {
	case "prefix", "output", "copy", "pass-complete", "done":
		return "cumulative positions (decrement after placement)"
	default:
		return "digit frequencies"
	}
}
