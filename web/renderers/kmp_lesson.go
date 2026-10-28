package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"strconv"
	"strings"
)

type KMPLessonState struct {
	Text, Pattern                                    string
	LPS                                              []int
	Phase                                            string
	I, J, Shift                                      int
	LPSIndex, PrefixLength, FallbackFrom, FallbackTo int
	LPSKnownThrough                                  int
	MatchIndex                                       int
	ResultKnown                                      bool
}

func kmpLessonByte(b byte) string {
	if b == ' ' {
		return "SP"
	}
	if b >= 33 && b <= 126 {
		return string(b)
	}
	return fmt.Sprintf("0x%02X", b)
}

// RenderKMPLessonSVG emits destination geometry; shared playback moves stable groups.
func RenderKMPLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(KMPLessonState)
	if !ok {
		return `<g data-lesson="kmp" data-phase="empty" data-scene-width="960" data-scene-height="420" data-min-width="960"><title>KMP byte offsets</title><text x="32" y="48" fill="#eff5ed" font-size="14">No KMP state available</text></g>`
	}
	if width <= 0 || height <= 0 {
		return `<g data-lesson="kmp" data-phase="empty" data-scene-width="960" data-scene-height="420" data-min-width="960"><title>KMP byte offsets</title><text x="32" y="48" fill="#eff5ed" font-size="14">Empty viewport; KMP byte offsets</text></g>`
	}
	shift := s.Shift
	if shift < 0 {
		shift = 0
	}
	// Build bounds keep this below 32768; direct oversized traces receive readable fallback.
	extent := max(len(s.Text), shift+len(s.Pattern))
	if extent > 670 || extent < 0 {
		return `<g data-lesson="kmp" data-phase="empty" data-scene-width="960" data-scene-height="420"><title>KMP byte offsets</title><text x="32" y="48" fill="#eff5ed" font-size="14">KMP byte offsets: input exceeds display bounds</text></g>`
	}
	w := max(960, 64+48*extent)
	h := max(420, min(height, 32768))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="kmp" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="ui-monospace, monospace" font-size="14" fill="#eff5ed"><title>KMP byte offsets: %s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(step.Explanation))
	label := func(x, y int, text, color string) {
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="%s">%s</text>`, x, y, color, html.EscapeString(text))
	}
	label(32, 28, "KMP · byte offsets · "+s.Phase, "#eff5ed")
	// Original strings use wrapped, escaped text so byte cells never masquerade as runes.
	original := []rune("Text: " + s.Text + "   Pattern: " + s.Pattern)
	for start, row := 0, 0; start < len(original); start, row = start+max(1, (w-64)/9), row+1 {
		end := min(len(original), start+max(1, (w-64)/9))
		label(32, 372+row*18, string(original[start:end]), "#bbcab9")
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	cell := func(role string, index, x, y int, value string, active, matched bool) {
		stroke := "#bbcab9"
		if active {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&b, `<g id="kmp-%s-%d" data-entity="kmp-%s-%d" data-x="%d" data-y="%d" transform="translate(%d %d)"`, role, index, role, index, x, y, x, y)
		if active {
			b.WriteString(` data-emphasis="compare"`)
		}
		fmt.Fprintf(&b, `><rect width="44" height="44" rx="7" fill="%s" stroke="%s" stroke-width="3"/><text x="22" y="27" fill="#101820" text-anchor="middle" font-size="16">%s</text><text x="22" y="-7" text-anchor="middle" fill="#bbcab9">%d</text>`, palette[index%5], stroke, html.EscapeString(value), index)
		if matched {
			b.WriteString(`<path d="M2 48H42" stroke="#9bf0bb" stroke-width="4"/><text x="22" y="65" text-anchor="middle" fill="#9bf0bb">=</text>`)
		}
		b.WriteString(`</g>`)
	}
	label(32, 64, "Text", "#bbcab9")
	label(32, 156, "Pattern · shift = "+strconv.Itoa(shift), "#bbcab9")
	label(32, 272, "LPS · fixed byte indices", "#bbcab9")
	search := s.Phase != "lps" && s.Phase != "init"
	for i := 0; i < len(s.Text); i++ {
		cell("text", i, 32+48*i, 96, kmpLessonByte(s.Text[i]), search && i == s.I && s.I < len(s.Text), search && i >= shift && i < shift+s.J)
	}
	for j := 0; j < len(s.Pattern); j++ {
		cell("pattern", j, 32+48*(shift+j), 184, kmpLessonByte(s.Pattern[j]), (search && j == s.J) || (s.Phase == "lps" && (j == s.LPSIndex || j == s.PrefixLength)), search && j < s.J)
		value := "?"
		if j <= s.LPSKnownThrough && j < len(s.LPS) {
			value = strconv.Itoa(s.LPS[j])
		}
		cell("lps", j, 32+48*j, 300, value, j == s.LPSIndex, false)
	}
	if s.Phase == "lps" && s.LPSIndex >= 0 && s.LPSIndex < len(s.Pattern) {
		prefix := s.PrefixLength
		label(32, 248, fmt.Sprintf("computeLPS · prefix length = %d", prefix), "#ffe29a")
		if prefix > 0 {
			prefix = min(prefix, s.LPSIndex+1)
			fmt.Fprintf(&b, `<path d="M32 286v-6h%dv6 M%d 286v-6h%dv6" fill="none" stroke="#ffe29a" stroke-width="2"/>`, 48*prefix-4, 32+48*(s.LPSIndex-prefix+1), 48*prefix-4)
		}
	}
	if s.FallbackFrom > 0 && s.FallbackTo >= 0 && s.LPSIndex >= 0 && s.LPSIndex < len(s.Pattern) && s.Phase == "fallback" {
		label(32, 248, fmt.Sprintf("j = lps[%d] = %d · i stays %d", s.FallbackFrom-1, s.FallbackTo, s.I), "#8bd5ff")
		oldX := 32 + 48*(s.I-s.FallbackFrom+s.FallbackFrom-1) + 22
		lookupX := 32 + 48*s.LPSIndex + 22
		newX := 32 + 48*(shift+s.J) + 22
		fmt.Fprintf(&b, `<path d="M%d 228Q%d 252 %d 294 M%d 294Q%d 254 %d 228" stroke="#8bd5ff" stroke-width="2" stroke-dasharray="5 4" fill="none"/>`, oldX, oldX, lookupX, lookupX, newX, newX)
	}
	if search {
		i := max(0, min(s.I, len(s.Text)))
		j := max(0, min(s.J, len(s.Pattern)))
		label(32+48*i, 84, "i ↓", "#ffe29a")
		label(32+48*(shift+j), 176, "j ↓", "#ffe29a")
	}
	if s.Phase == "compare" {
		label(32, 248, "Compare current bytes", "#ffe29a")
	}
	if s.Phase == "mismatch" {
		label(32, 248, "≠ mismatch · reuse LPS or advance text", "#ff9b9b")
	}
	if s.ResultKnown {
		result := fmt.Sprintf("Found · byte offset %d", s.MatchIndex)
		if s.MatchIndex < 0 {
			result = "Not found · -1"
		}
		if len(s.Pattern) == 0 {
			result = "Empty pattern · returns 0 · no byte comparisons"
		}
		label(32, 248, result, "#9bf0bb")
	}
	b.WriteString(`</g>`)
	return b.String()
}
