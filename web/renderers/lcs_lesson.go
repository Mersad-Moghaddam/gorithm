package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"strconv"
	"strings"
)

// LCSLessonState records byte-prefix DP and the exact backtracking transition.
type LCSLessonState struct {
	TableState
	S1, S2           string
	DP               [][]int
	Filled           [][]bool
	Row, Col         int
	From, To         [2]int
	Sources, Matches [][2]int
	Sequence, Phase  string
	Match            bool
}

func lcsByteLabel(b byte) string {
	if b < 32 || b >= 127 {
		return fmt.Sprintf("0x%02X", b)
	}
	return string(b)
}

func RenderLCSLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(LCSLessonState)
	if !ok {
		return `<g data-lesson="lcs" data-phase="empty" data-scene-width="320" data-scene-height="200"><title>LCS snapshot unavailable</title></g>`
	}
	// ponytail: fixed 48px byte cells use stage-local scrolling; no matrix zoom layer.
	cols := len(s.S2) + 1
	rows := len(s.S1) + 1
	sceneW := max(320, 112+48*max(cols, len(s.S1)))
	sceneW = max(sceneW, 48+48*len(s.Matches))
	sceneW = max(sceneW, min(max(width, 0), 1440))
	sceneH := max(400, 368+48*rows)
	sceneH = max(sceneH, min(max(height, 0), 1440))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="lcs" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="ui-monospace,monospace" font-size="14" fill="#eff5ed"><title>LCS byte prefixes: %s</title>`, html.EscapeString(s.Phase), sceneW, sceneH, sceneW, html.EscapeString(s.Phase))
	text := func(x, y int, value string) {
		fmt.Fprintf(&b, `<text x="%d" y="%d">%s</text>`, x, y, html.EscapeString(value))
	}
	text(16, 24, "LCS · byte comparison · "+s.Phase)
	formula := "Boundary prefixes = 0; blank = uncomputed"
	if s.From[0] > 0 && s.From[1] > 0 && s.From[0] < len(s.DP) && s.From[1] < len(s.DP[s.From[0]]) {
		i, j := s.From[0], s.From[1]
		if s.Match {
			formula = fmt.Sprintf("diagonal %d + 1 = %d", s.DP[i-1][j-1], s.DP[i-1][j-1]+1)
		} else {
			formula = fmt.Sprintf("max(top %d, left %d)", s.DP[i-1][j], s.DP[i][j-1])
		}
	}
	text(16, 48, formula)
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	tile := func(id string, x, y, index int, label, alias string, compare, done bool) {
		stroke := "#bbcab9"
		if compare {
			stroke = "#ffe29a"
		}
		if done {
			stroke = "#9bf0bb"
		}
		fmt.Fprintf(&b, `<g data-entity="%s" data-x="%d" data-y="%d" transform="translate(%d %d)"`, html.EscapeString(id), x, y, x, y)
		if alias != "" {
			fmt.Fprintf(&b, ` data-from-entity="%s"`, html.EscapeString(alias))
		}
		if compare {
			b.WriteString(` data-emphasis="compare"`)
		}
		fmt.Fprintf(&b, `><rect width="44" height="40" rx="6" fill="%s" stroke="%s" stroke-width="2"/><text x="22" y="24" text-anchor="middle">%s</text><text x="22" y="55" font-size="12" text-anchor="middle">[%d]</text></g>`, palette[index%5], stroke, html.EscapeString(label), index)
	}
	for side, str := range []string{s.S1, s.S2} {
		y := 80 + 76*side
		text(16, y+24, fmt.Sprintf("s%d", side+1))
		for i := 0; i < len(str); i++ {
			active := false
			if s.Phase == "compare" || s.Phase == "backtrack" {
				active = (side == 0 && i == s.From[0]-1) || (side == 1 && i == s.From[1]-1)
			}
			tile(fmt.Sprintf("s%d:%d", side+1, i), 64+48*i, y, i, lcsByteLabel(str[i]), "", active, false)
		}
	}
	const startX, startY = 64, 264
	for c := 0; c < cols; c++ {
		label := "∅"
		if c > 0 {
			label = lcsByteLabel(s.S2[c-1])
		}
		text(startX+c*48+8, startY-10, label)
	}
	for r := 0; r < rows; r++ {
		label := "∅"
		if r > 0 {
			label = lcsByteLabel(s.S1[r-1])
		}
		text(16, startY+r*48+26, label)
		for c := 0; c < cols; c++ {
			stroke := "#52634e"
			for _, p := range s.Sources {
				if p == [2]int{r, c} {
					stroke = "#8bd5ff"
				}
			}
			if r == s.Row && c == s.Col {
				stroke = "#ffe29a"
			}
			if s.Phase == "done" && r == rows-1 && c == cols-1 {
				stroke = "#9bf0bb"
			}
			fmt.Fprintf(&b, `<rect x="%d" y="%d" width="44" height="44" rx="5" fill="#192923" stroke="%s" stroke-width="2"/>`, startX+c*48, startY+r*48, stroke)
			if r < len(s.DP) && r < len(s.Filled) && c < len(s.DP[r]) && c < len(s.Filled[r]) && s.Filled[r][c] {
				text(startX+c*48+14, startY+r*48+27, strconv.Itoa(s.DP[r][c]))
			}
		}
	}
	for _, p := range s.Sources {
		if p[0] >= 0 && p[1] >= 0 && s.Row >= 0 && s.Col >= 0 {
			fmt.Fprintf(&b, `<path d="M%d %d L%d %d" fill="none" stroke="#8bd5ff" stroke-dasharray="4 3"/><text x="%d" y="%d" fill="#8bd5ff">›</text>`, startX+p[1]*48+22, startY+p[0]*48+22, startX+s.Col*48+22, startY+s.Row*48+22, startX+s.Col*48+28, startY+s.Row*48+20)
		}
	}
	if s.Phase == "backtrack" || s.Phase == "match" || s.Phase == "up" || s.Phase == "left" {
		r, c := s.To[0], s.To[1]
		if r >= 0 && c >= 0 {
			x, y := startX+c*48, startY+r*48
			fmt.Fprintf(&b, `<g data-entity="cursor" data-x="%d" data-y="%d" transform="translate(%d %d)"><rect width="44" height="44" rx="5" fill="none" stroke="#ffe29a" stroke-width="4"/><title>Backtrack cursor</title></g>`, x, y, x, y)
		}
		if s.From != s.To {
			fmt.Fprintf(&b, `<path d="M%d %d L%d %d" stroke="#ffe29a" fill="none" stroke-width="3"/>`, startX+s.From[1]*48+22, startY+s.From[0]*48+22, startX+s.To[1]*48+22, startY+s.To[0]*48+22)
		}
	}
	footer := startY + rows*48 + 24
	result := "Reconstruction (prepend bytes)"
	if s.Phase == "done" {
		result = fmt.Sprintf("✓ Result: %d bytes", len(s.Sequence))
	}
	text(16, footer, result)
	for i, p := range s.Matches {
		if p[0] >= 0 && p[0] < len(s.S1) {
			tile(fmt.Sprintf("output:%d:%d", p[0], p[1]), 16+i*48, footer+16, p[0], lcsByteLabel(s.S1[p[0]]), fmt.Sprintf("s1:%d", p[0]), false, s.Phase == "done")
		}
	}
	if len(s.Matches) == 0 {
		text(16, footer+44, "∅ (empty sequence)")
	}
	b.WriteString(`</g>`)
	return b.String()
}
