package renderers

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// InsertionLessonState records real memory copies separately from the held key.
type InsertionLessonState struct {
	ArrayState
	IDs                                          []int
	Pass, Scan, KeyValue, KeyID, Hole, PrefixEnd int
	KeyHeld                                      bool
	MinValue, MaxValue                           int
}

func RenderInsertionLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(InsertionLessonState)
	if !ok {
		return `<g data-lesson="insertion" data-phase="unavailable"><title>Insertion sort state unavailable</title><text x="32" y="48" fill="#eff5ed" font-size="16">Insertion sort state unavailable</text></g>`
	}
	n := len(s.Values)
	width = max(560, max(width, 64+n*96))
	width = min(width, 32768)
	height = min(32768, max(height, 450))
	phase := "inspect"
	switch step.LineID {
	case "L4":
		phase = "hold-key"
	case "L6":
		phase = "compare"
	case "L7":
		phase = "shift"
	case "L10":
		phase = "insert"
	case "L12":
		phase = "done"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="insertion" data-phase="%s" data-operation="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Insertion sort: %s. Key shelf and actual array memory.</title>`, phase, insertionOperation(phase), width, height, width, phase)
	b.WriteString(`<text x="32" y="30" font-size="16">Held key — copy kept outside array</text>`)
	if n == 0 {
		b.WriteString(`<text x="32" y="150" font-size="16">Empty array — already sorted</text></g>`)
		return b.String()
	}
	if len(s.IDs) != n {
		b.WriteString(`<text x="32" y="150">Missing item identities</text></g>`)
		return b.String()
	}
	pitch := float64(width-64) / float64(n)
	center := func(i int) float64 { return 32 + (float64(i)+0.5)*pitch }
	top, bottom := 150.0, float64(height-150)
	lo, hi := float64(s.MinValue), float64(s.MaxValue)
	if hi <= lo {
		hi = lo + 1
	}
	scale := (bottom - top) / (hi - lo)
	baseline := top + hi*scale
	fmt.Fprintf(&b, `<path d="M32 %.2f H%d" stroke="#bbcab9"/><text x="32" y="%.2f" fill="#bbcab9">0 baseline</text>`, baseline, width-32, bottom+28)
	for i, v := range s.Values {
		id := s.IDs[i]
		x := center(i)
		valueY := baseline - float64(v)*scale
		barTop := math.Min(baseline, valueY)
		barHeight := math.Max(2, math.Abs(valueY-baseline))
		stroke := "#bbcab9"
		sw := 1
		if step.Action == trace.ActionDone {
			stroke = "#9bf0bb"
			sw = 3
		}
		if phase == "compare" && s.Scan == i {
			stroke = "#ffe29a"
			sw = 3
		}
		alias := ""
		if phase == "shift" && i == s.Scan+1 {
			alias = fmt.Sprintf(` data-from-entity="main:%d:%d"`, id, s.Scan)
		}
		if phase == "insert" && i == s.Current {
			alias = fmt.Sprintf(` data-from-entity="held:%d"`, id)
		}
		fmt.Fprintf(&b, `<g data-entity="main:%d:%d" data-id="%d" data-role="main" data-slot="%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)" data-motion="linear"%s><rect x="-30" y="%.2f" width="60" height="%.2f" rx="5" fill="%s" stroke="%s" stroke-width="%d"/><text x="0" y="%.2f" text-anchor="middle">%d #%d</text>`, id, i, id, i, x, baseline, x, baseline, alias, barTop-baseline, barHeight, insertionColor(id), stroke, sw, barTop-baseline-12, v, id)
		if step.Action == trace.ActionDone {
			b.WriteString(`<text x="0" y="24" text-anchor="middle" fill="#9bf0bb">✓</text>`)
		}
		b.WriteString(`</g>`)
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" text-anchor="middle" fill="#bbcab9">[%d]</text>`, x, bottom+48, i)
	}
	if s.KeyHeld {
		alias := ""
		if phase == "hold-key" {
			alias = fmt.Sprintf(` data-from-entity="main:%d:%d"`, s.KeyID, s.Pass)
		}
		fmt.Fprintf(&b, `<g data-entity="held:%d" data-id="%d" data-role="held" data-x="100" data-y="80" transform="translate(100 80)" data-motion="linear"%s><rect x="-64" y="-25" width="128" height="50" rx="8" fill="%s" stroke="#ffe29a" stroke-width="3"/><text x="0" y="5" text-anchor="middle">%d #%d</text></g>`, s.KeyID, s.KeyID, alias, insertionColor(s.KeyID), s.KeyValue, s.KeyID)
	} else {
		b.WriteString(`<text x="36" y="85" fill="#bbcab9">No held key</text>`)
	}
	if s.KeyHeld && s.Hole >= 0 && s.Hole < n {
		x := center(s.Hole)
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="76" height="%.2f" fill="none" stroke="#8bd5ff" stroke-dasharray="5 4"/><text x="32" y="%d" fill="#8bd5ff">Conceptual hole [%d] — memory: %d #%d</text>`, x-38, top-35, bottom-top+45, height-72, s.Hole, s.Values[s.Hole], s.IDs[s.Hole])
	}
	if s.PrefixEnd >= 0 && s.PrefixEnd < n {
		y := float64(height - 40)
		fmt.Fprintf(&b, `<path d="M32 %.2f v8 H%.2f v-8" stroke="#9bf0bb" fill="none"/><text x="32" y="%d" fill="#9bf0bb">Sorted prefix [0..%d] — %s</text>`, y, center(s.PrefixEnd)+30, height-12, s.PrefixEnd, insertionPrefixMeaning(step.Action))
	}
	if phase == "compare" && s.Scan < 0 {
		b.WriteString(`<text x="230" y="85" fill="#ffe29a">j = -1: boundary reached</text>`)
	}
	if phase == "shift" {
		fmt.Fprintf(&b, `<text x="230" y="85" fill="#8bd5ff">Copy [%d] to [%d]; no exchange</text>`, s.Scan, s.Scan+1)
	}
	b.WriteString(`</g>`)
	return b.String()
}

func insertionColor(id int) string {
	colors := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	return colors[(id%5+5)%5]
}
func insertionOperation(phase string) string {
	if phase == "shift" {
		return "copy"
	}
	if phase == "insert" {
		return "write-key"
	}
	return phase
}
func insertionPrefixMeaning(action trace.ActionType) string {
	if action == trace.ActionDone {
		return "complete ✓"
	}
	return "locally sorted, not final"
}
