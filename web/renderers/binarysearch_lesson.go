package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type BinarySearchLessonState struct {
	ArrayState
	Phase                          string
	Target, Low, High, Mid, Result int
	PreviousLow, PreviousHigh      int
	DiscardLow, DiscardHigh        int
	SortedInput, ResultKnown       bool
}

// RenderBinarySearchLessonSVG returns a complete fragment; pointer geometry is
// final destination geometry for the shared player, never a second search.
func RenderBinarySearchLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(BinarySearchLessonState)
	w := max(800, 64+76*len(s.Values))
	h := 360
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="binarysearch" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Binary search: %s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(s.Phase))
	if !ok || width <= 0 || height <= 0 {
		b.WriteString(`<text x="32" y="40">No binary search state or usable viewport</text></g>`)
		return b.String()
	}
	text := func(x, y int, label string) {
		fmt.Fprintf(&b, `<text x="%d" y="%d">%s</text>`, x, y, html.EscapeString(label))
	}
	text(32, 28, fmt.Sprintf("Target: %d", s.Target))
	if s.Mid >= 0 && s.Mid < len(s.Values) {
		text(32, 64, fmt.Sprintf("mid = %d + (%d − %d)/2 = %d", s.Low, s.High, s.Low, s.Mid))
	} else {
		text(32, 64, "mid: pending")
	}
	badge := "sorted input verified"
	if !s.SortedInput {
		badge = "sorted input required — exclusions unproven"
	}
	text(32, 92, badge)
	b.WriteString(`<defs><pattern id="binarysearch-excluded" width="8" height="8" patternUnits="userSpaceOnUse"><path d="M0 8L8 0" stroke="#eff5ed" stroke-opacity="0.35"/></pattern></defs>`)
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, v := range s.Values {
		x := 32 + 76*i
		stroke := "#bbcab9"
		cue := ""
		if i == s.Mid {
			stroke = "#ffe29a"
			cue = "mid"
			if s.Phase == "compare" {
				cue = "compare"
				if step.LineID == "L7" || step.LineID == "L9" {
					cue = "≠ target"
				}
			}
		}
		if s.ResultKnown && i == s.Result {
			stroke = "#9bf0bb"
			cue = "found"
		}
		fmt.Fprintf(&b, `<g id="binarysearch-item-%d" data-entity="binarysearch-item-%d" data-x="%d" data-y="168" transform="translate(%d 168)"`, i, i, x, x)
		if s.Phase == "compare" && i == s.Mid {
			b.WriteString(` data-emphasis="compare"`)
		}
		fmt.Fprintf(&b, `><rect width="64" height="56" rx="10" fill="%s" stroke="%s" stroke-width="3"/><text x="32" y="33" text-anchor="middle" font-size="16" fill="#101820" textLength="%d" lengthAdjust="spacingAndGlyphs">%d</text><text x="32" y="77" text-anchor="middle">[%d]</text>`, palette[i%len(palette)], stroke, min(56, len(fmt.Sprint(v))*9), v, i)
		if i < s.Low || i > s.High {
			fmt.Fprintf(&b, `<rect width="64" height="56" rx="10" fill="url(#binarysearch-excluded)"/><text x="32" y="140" text-anchor="middle">%s</text>`, map[bool]string{true: "excluded", false: "outside"}[s.SortedInput])
		}
		if cue != "" {
			fmt.Fprintf(&b, `<text x="32" y="49" text-anchor="middle" font-size="12">%s</text>`, html.EscapeString(cue))
		}
		b.WriteString(`</g>`)
	}
	// Separate vertical lanes keep coincident low/high/mid readable.
	marker := func(name string, index, y int) {
		x := 64 + 76*index
		if index < 0 {
			x = 32
		}
		if index >= len(s.Values) {
			x = 32 + 76*len(s.Values)
		}
		fmt.Fprintf(&b, `<g data-entity="binarysearch-%s" data-x="%d" data-y="%d" transform="translate(%d %d)"><path d="M-5 6L0 0L5 6" fill="none" stroke="#8bd5ff" stroke-width="2"/><text x="0" y="-4" text-anchor="middle">%s=%d</text></g>`, name, x, y, x, y, name, index)
	}
	marker("low", s.Low, 128)
	marker("high", s.High, 280)
	if s.Mid >= 0 && s.Mid < len(s.Values) {
		marker("mid", s.Mid, 150)
	}
	if s.Low <= s.High && s.Low >= 0 && s.High < len(s.Values) {
		fmt.Fprintf(&b, `<path id="binarysearch-active-range" d="M%d 240v8H%dv-8" fill="none" stroke="#8bd5ff" stroke-width="3"/>`, 32+76*s.Low, 96+76*s.High)
	} else {
		text(32, 332, "empty interval")
	}
	if s.ResultKnown {
		if s.Result >= 0 {
			text(360, 332, fmt.Sprintf("found index %d", s.Result))
		} else {
			text(360, 332, "result −1: no match returned")
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}
