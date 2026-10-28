package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// LinearSearchLessonState distinguishes pending checks from unequal and found values.
type LinearSearchLessonState struct {
	ArrayState
	Phase                                 string
	Target, Index, CheckedThrough, Result int
	Compared, Equal, ResultKnown          bool
}

// RenderLinearSearchLessonSVG renders fixed identity tiles, never magnitude bars.
func RenderLinearSearchLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(LinearSearchLessonState)
	w := max(800, 64+76*len(s.Values))
	h := max(320, min(height, 32768))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="linearsearch" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Linear search: %s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(s.Phase))
	if !ok {
		b.WriteString(`<text x="32" y="60">No visual state available</text></g>`)
		return b.String()
	}
	fmt.Fprintf(&b, `<text x="32" y="30" fill="#ffe29a">Target = %d</text>`, s.Target)
	if s.ResultKnown {
		fmt.Fprintf(&b, `<text x="32" y="306" fill="#9bf0bb">result = %d</text>`, s.Result)
	}
	if len(s.Values) == 0 {
		b.WriteString(`<text x="32" y="160">Empty array: no comparisons</text></g>`)
		return b.String()
	}
	colors := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	offset := (w - (64 + 76*len(s.Values))) / 2
	for i, value := range s.Values {
		x := 32 + offset + 76*i
		outline, caption := "#bbcab9", ""
		if i <= s.CheckedThrough {
			outline, caption = "#9bf0bb", "checked"
		}
		if s.ResultKnown && s.Result == i {
			outline, caption = "#9bf0bb", "found"
		}
		emphasis := ""
		if s.Phase == "compare" && s.Index == i {
			outline, emphasis = "#ffe29a", ` data-emphasis="compare"`
			operator := " ≠ "
			if s.Equal {
				operator = " = "
			}
			// Keep the equation inside scene bounds, even for extreme signed operands.
			equationX := max(140, min(x+32, w-140))
			fmt.Fprintf(&b, `<text x="%d" y="82" text-anchor="middle" fill="#ffe29a">%d%s%d</text>`, equationX, value, operator, s.Target)
		}
		fontSize := 16
		if value <= -100000 || value >= 1000000 {
			fontSize = 14
		}
		fmt.Fprintf(&b, `<g id="linearsearch-item-%d" data-entity="linearsearch-item-%d" data-x="%d" data-y="144" transform="translate(%d 144)"%s><rect data-role="tile" width="64" height="56" rx="10" fill="%s" stroke="%s" stroke-width="3"/><text x="32" y="33" text-anchor="middle" font-size="%d" letter-spacing="-0.5" fill="#101820">%d</text><text x="32" y="80" text-anchor="middle">[%d]</text><text x="32" y="103" text-anchor="middle" fill="#9bf0bb">%s</text>`, i, i, x, x, emphasis, colors[i%len(colors)], outline, fontSize, value, i, html.EscapeString(caption))
		if i <= s.CheckedThrough {
			b.WriteString(`<text x="32" y="123" text-anchor="middle" fill="#9bf0bb">≠ target</text>`)
		}
		b.WriteString(`</g>`)
	}
	if s.Index >= 0 && s.Index < len(s.Values) {
		x := 64 + offset + 76*s.Index
		fmt.Fprintf(&b, `<g data-entity="linearsearch-pointer" data-x="%d" data-y="116" transform="translate(%d 116)" data-motion="linear"><path d="M-8 0 L8 0 L0 12 Z" fill="#8bd5ff"/><text x="0" y="-6" text-anchor="middle" fill="#8bd5ff">i</text></g>`, x, x)
	}
	b.WriteString(`</g>`)
	return b.String()
}
