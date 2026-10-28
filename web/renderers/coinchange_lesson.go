package renderers

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type CoinChangeLessonState struct {
	TableState
	Coins []int
	Amount, Index, CoinIndex, SourceIndex int
	DP []int
	Initialized bool
	CanTake, HasCandidate bool
	Previous, Candidate int
	Phase string
	Result int
	HasResult bool
}

func coinChangeCount(value, amount int) string {
	if value > amount { return "∞" }
	return strconv.Itoa(value)
}

// RenderCoinChangeLessonSVG shows only the stored recurrence evidence, not a coin reconstruction.
func RenderCoinChangeLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(CoinChangeLessonState)
	if !ok { return `<g data-lesson="coinchange" data-phase="empty" data-scene-width="320" data-scene-height="180"><title>Coin Change: no snapshot</title><text x="16" y="40" fill="#eff5ed" font-size="14">No coin change snapshot</text></g>` }
	// ponytail: wrap cells at readable 52px pitch; add virtualized rail only beyond Build's 200-amount cap.
	w := width
	if w < 320 { w = 320 }; if w > 1440 { w = 1440 }
	cols := (w-32)/52
	if cols < 1 { cols = 1 }
	coinCols := (w-32)/64
	if coinCols < 1 { coinCols = 1 }
	bankRows := (len(s.Coins)+coinCols-1)/coinCols
	if bankRows < 1 { bankRows = 1 }
	railY := 76 + bankRows*72 + 32
	rows := (len(s.DP)+cols-1)/cols
	if rows < 1 { rows = 1 }
	panelY := railY + rows*64 + 24
	h := panelY + 180
	if height > h && height <= 32768 { h = height }
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="coinchange" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(fmt.Sprintf("Coin Change: %s, target %d", s.Phase, s.Amount)))
	text := func(x, y int, label, color string) { fmt.Fprintf(&b, `<text x="%d" y="%d" fill="%s">%s</text>`, x, y, color, html.EscapeString(label)) }
	text(16, 28, fmt.Sprintf("Fewest coins for amount %d", s.Amount), "#eff5ed")
	text(16, 54, "Denominations · numbered by input order", "#bbcab9")
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, coin := range s.Coins {
		x, y := 44+(i%coinCols)*64, 88+(i/coinCols)*72
		stroke, emphasis := "#bbcab9", ""
		if i == s.CoinIndex { stroke, emphasis = "#ffe29a", ` data-emphasis="compare"` }
		fmt.Fprintf(&b, `<g data-entity="coin:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"%s><circle r="22" fill="%s" stroke="%s" stroke-width="3"/><text text-anchor="middle" y="5" font-weight="700">%d</text><text text-anchor="middle" y="42" font-size="12">#%d</text></g>`, i, x, y, x, y, emphasis, palette[i%len(palette)], stroke, coin, i+1)
	}
	text(16, railY-20, "dp[amount] · ∞ means unreachable", "#bbcab9")
	for i, value := range s.DP {
		x, y := 16+(i%cols)*52, railY+(i/cols)*64
		stroke := "#52634e"
		if i == s.SourceIndex { stroke = "#8bd5ff" }
		if i == s.Index { stroke = "#ffe29a" }
		if s.HasResult && s.Result >= 0 && i == s.Amount { stroke = "#9bf0bb" }
		label := coinChangeCount(value, s.Amount)
		if !s.Initialized && i != 0 { label = "·" }
		fmt.Fprintf(&b, `<g data-entity="amount:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"><rect width="46" height="36" rx="7" fill="#192923" stroke="%s" stroke-width="2"/><text x="23" y="23" text-anchor="middle">%s</text><text x="23" y="53" text-anchor="middle" fill="#bbcab9" font-size="12">[%d]</text></g>`, i, x, y, x, y, stroke, label, i)
	}
	if (s.Phase == "compare" || s.Phase == "update") && s.CanTake && s.SourceIndex >= 0 && s.SourceIndex < len(s.DP) && s.Index >= 0 && s.Index < len(s.DP) {
		x1, y1 := 39+(s.SourceIndex%cols)*52, railY+(s.SourceIndex/cols)*64
		x2, y2 := 39+(s.Index%cols)*52, railY+(s.Index/cols)*64
		fmt.Fprintf(&b, `<path d="M%d %d Q%d %d %d %d" fill="none" stroke="#8bd5ff" stroke-dasharray="4 3"/><path d="M%d %d l-4 -7 m4 7 l4 -7" fill="none" stroke="#8bd5ff"/>`, x1, y1, (x1+x2)/2, min(y1,y2)-18, x2, y2, x2, y2)
	}
	fmt.Fprintf(&b, `<rect x="16" y="%d" width="%d" height="148" rx="12" fill="#192923" stroke="#52634e"/>`, panelY, w-32)
	switch s.Phase {
	case "compare", "update":
		text(28, panelY+26, "Current best: "+coinChangeCount(s.Previous,s.Amount), "#ffe29a")
		if !s.CanTake {
			text(28, panelY+56, "Coin does not fit this amount", "#bbcab9")
		} else if !s.HasCandidate {
			text(28, panelY+56, fmt.Sprintf("dp[%d] = ∞ · unreachable", s.SourceIndex), "#bbcab9")
			text(28, panelY+84, "No finite candidate; keep best", "#bbcab9")
		} else {
			text(28, panelY+56, fmt.Sprintf("dp[%d] + 1 = %d", s.SourceIndex, s.Candidate), "#8bd5ff")
			label := "Keep best · candidate not smaller"
			if s.Candidate == s.Previous { label = "Tie · no write" }
			if s.Candidate < s.Previous { label = "Candidate improves current best" }
			if s.Phase == "update" { label = fmt.Sprintf("Write min(%s, %d) = %d",coinChangeCount(s.Previous,s.Amount),s.Candidate,s.Candidate) }
			text(28, panelY+84, label, "#eff5ed")
			if s.CoinIndex >= 0 && s.CoinIndex < len(s.Coins) {
				// Same candidate identity moves from proposal lane to committed cell.
				x,y := w-52,panelY+118
				if s.Phase == "update" && s.Index >= 0 && s.Index < len(s.DP) { x,y = 39+(s.Index%cols)*52,railY+(s.Index/cols)*64+18 }
				alias := ""
				if s.Phase == "compare" { alias = fmt.Sprintf(` data-from-entity="coin:%d"`,s.CoinIndex) }
				fmt.Fprintf(&b, `<g data-entity="candidate" data-x="%d" data-y="%d" transform="translate(%d %d)"%s><circle r="12" fill="%s" stroke="#8bd5ff"/><text y="5" text-anchor="middle" font-size="14">+1</text></g>`,x,y,x,y,alias,palette[s.CoinIndex%len(palette)])
			}
		}
	case "done":
		if s.Result < 0 { text(28,panelY+30,"Result: -1 · no solution", "#ff9b9b") } else { text(28,panelY+30,fmt.Sprintf("✓ Minimum: %d coins", s.Result), "#9bf0bb") }
		text(28,panelY+60,"Count only · no reconstructed path", "#bbcab9")
	case "base": text(28,panelY+30,"dp[0] = 0 · no coins needed", "#eff5ed")
	case "infinity": text(28,panelY+30,"Positive amounts start at ∞", "#eff5ed")
	default: text(28,panelY+30,"Compare predecessor + one coin", "#eff5ed")
	}
	b.WriteString(`</g>`)
	return b.String()
}
