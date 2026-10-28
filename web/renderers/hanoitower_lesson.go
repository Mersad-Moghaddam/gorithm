package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"strings"
)

type HanoiTowerCall struct{ N, From, To, Aux int }

type HanoiTowerLessonState struct {
	HanoiState
	TotalDisks, MoveCount int
	Calls                 []HanoiTowerCall
	Phase                 string
	Disk, From, To        int
	BeforePegs            [3][]int
}

func RenderHanoiTowerLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(HanoiTowerLessonState)
	// ponytail: retain 320px local scene for tiny viewports; shared viewport supplies scrolling.
	width = max(320, min(width, 32768))
	height = max(450, min(height, 32768))
	height = max(height, 410+len(s.Calls)*26)
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="hanoitower" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Tower of Hanoi: %s, %d legal moves</title>`, html.EscapeString(s.Phase), width, height, width, html.EscapeString(s.Phase), s.MoveCount)
	text := func(x, y float64, label, color string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" fill="%s">%s</text>`, x, y, color, html.EscapeString(label))
	}
	if !ok {
		text(16, 32, "No Hanoi lesson state", "#bbcab9")
		b.WriteString(`</g>`)
		return b.String()
	}
	text(16, 28, fmt.Sprintf("Move %d · %d disks", s.MoveCount, s.TotalDisks), "#eff5ed")
	caption := "Target C · auxiliary B"
	switch s.Phase {
	case "base-check":
		caption = "Check n = 0: no disks to move?"
	case "return":
		caption = "n = 0: return without moving"
	case "move":
		caption = fmt.Sprintf("Disk %d: %s → %s", s.Disk, hanoiTowerPeg(s.From), hanoiTowerPeg(s.To))
	case "done":
		caption = fmt.Sprintf("✓ Complete · %d legal moves", s.MoveCount)
	}
	text(16, 52, caption, "#ffe29a")
	lane := float64(width-32) / 3
	base := 300.0
	diskH := 24.0
	total := max(1, s.TotalDisks)
	pegTop := base - float64(total)*diskH - 16
	clear := pegTop - 32
	fmt.Fprintf(&b, `<path d="M16 %.2f H%d" stroke="#bbcab9" stroke-width="3"/>`, base, width-16)
	for p := 0; p < 3; p++ {
		x := 16 + lane*(float64(p)+.5)
		fmt.Fprintf(&b, `<path d="M%.2f %.2f V%.2f" stroke="#bbcab9" stroke-width="6" stroke-linecap="round"/>`, x, pegTop, base)
		role := []string{"from", "aux", "to"}[p]
		if len(s.Calls) > 0 {
			c := s.Calls[len(s.Calls)-1]
			switch p {
			case c.From:
				role = "from"
			case c.To:
				role = "to"
			case c.Aux:
				role = "aux"
			}
		}
		fmt.Fprintf(&b, `<text x="%.2f" y="326" text-anchor="middle">%s</text><text x="%.2f" y="346" text-anchor="middle" fill="#bbcab9">%s</text>`, x, hanoiTowerPeg(p), x, role)
	}
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for p, peg := range s.Pegs {
		for level, disk := range peg {
			x := 16 + lane*(float64(p)+.5)
			y := base - float64(level+1)*diskH
			w := 30 + (lane-42)*float64(max(1, min(disk, total)))/float64(total)
			stroke := "#eff5ed"
			if s.Phase == "move" && disk == s.Disk {
				stroke = "#ffe29a"
			}
			if s.Phase == "done" {
				stroke = "#9bf0bb"
			}
			fmt.Fprintf(&b, `<g data-entity="disk:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"`, disk, x, y, x, y)
			if s.Phase == "move" && disk == s.Disk {
				fmt.Fprintf(&b, ` data-motion="lift" data-clear-y="%.2f"`, clear)
			}
			fmt.Fprintf(&b, `><rect x="%.2f" y="0" width="%.2f" height="22" rx="6" fill="%s" stroke="%s" stroke-width="2"/><text x="0" y="16" text-anchor="middle" fill="#ffffff" font-weight="700">%d</text></g>`, -w/2, w, colors[(max(1, disk)-1)%len(colors)], stroke, disk)
		}
	}
	pending := "Pending: none"
	if len(s.Calls) > 0 {
		c := s.Calls[len(s.Calls)-1]
		if s.Phase == "left-call" {
			pending = fmt.Sprintf("Pending: n=%d %s → %s", c.N-1, hanoiTowerPeg(c.From), hanoiTowerPeg(c.Aux))
		}
		if s.Phase == "right-call" {
			pending = fmt.Sprintf("Pending: n=%d %s → %s", c.N-1, hanoiTowerPeg(c.Aux), hanoiTowerPeg(c.To))
		}
	}
	text(16, 376, pending, "#8bd5ff")
	text(16, 400, "Call stack · oldest first", "#bbcab9")
	for i, c := range s.Calls {
		text(16, float64(426+i*26), fmt.Sprintf("%d: n=%d from=%s to=%s aux=%s", i+1, c.N, hanoiTowerPeg(c.From), hanoiTowerPeg(c.To), hanoiTowerPeg(c.Aux)), "#eff5ed")
	}
	b.WriteString(`</g>`)
	return b.String()
}

func hanoiTowerPeg(p int) string {
	if p < 0 || p > 2 {
		return "?"
	}
	return string(rune('A' + p))
}
