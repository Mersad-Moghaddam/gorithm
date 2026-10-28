package renderers

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// FibonacciLessonState records only committed cells; zero is distinct from unknown.
type FibonacciLessonState struct {
	TableState
	N, Index  int
	Values    []int
	Known     []bool
	Sources   [2]int
	Phase     string
	Result    int
	HasResult bool
}

func RenderFibonacciLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(FibonacciLessonState)
	if !ok {
		return `<g data-lesson="fibonacci" data-phase="empty" data-scene-width="320" data-scene-height="160" data-min-width="320"><title>Fibonacci state unavailable</title><text x="16" y="40" fill="#eff5ed" font-size="14">Fibonacci state unavailable</text></g>`
	}
	// Keep essential labels at native size; foundation supplies local scrolling.
	w := width
	if w < 320 {
		w = 320
	}
	if w > 1440 {
		w = 1440
	}
	// Fixed width prevents later, longer values from moving existing cells.
	cellW := 106 // ponytail: fits Fibonacci through F(46); revise with input limit.
	cols := (w - 32) / (cellW + 8)
	if cols < 1 {
		cols = 1
	}
	rows := (len(s.Values) + cols - 1) / cols
	if rows < 1 {
		rows = 1
	}
	boardEnd := 88 + rows*80
	cardY := boardEnd + 52
	narrow := w < 600
	h := cardY + 112 + 60
	if narrow {
		h = cardY + 3*100 + 60
	}
	if height > h && height <= 32768 {
		h = height
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="fibonacci" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="ui-monospace, monospace"><title>Fibonacci F(%d), %s</title>`, html.EscapeString(s.Phase), w, h, w, s.N, html.EscapeString(s.Phase))
	fibonacciText(&b, 16, 28, fmt.Sprintf("Bottom-up DP: F(%d)", s.N), "#eff5ed", 16)
	formula := "Allocate slots; blank means unknown"
	if s.Phase == "base" {
		formula = "Base cases: F(0)=0, F(1)=1"
		if s.N <= 0 {
			formula = "n <= 0: return 0"
		}
	}
	recurrence := (s.Phase == "read" || s.Phase == "write") && s.Index >= 2 && s.Index < len(s.Values) && s.Sources[0] >= 0 && s.Sources[1] >= 0 && s.Sources[0] < len(s.Values) && s.Sources[1] < len(s.Values)
	if recurrence {
		formula = fmt.Sprintf("F(%d) = F(%d) + F(%d)", s.Index, s.Sources[0], s.Sources[1])
	}
	if s.HasResult {
		formula = fmt.Sprintf("F(%d) = %d", max(0, s.N), s.Result)
	}
	fibonacciText(&b, 16, 54, formula, "#bbcab9", 14)
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	position := func(i int) (int, int) { return 16 + (i%cols)*(cellW+8), 88 + (i/cols)*80 }
	for i, v := range s.Values {
		x, y := position(i)
		known := i < len(s.Known) && s.Known[i]
		label := "?"
		if known {
			label = strconv.Itoa(v)
		}
		stroke := "#bbcab9"
		extra := ""
		if recurrence && (i == s.Sources[0] || i == s.Sources[1]) {
			stroke = "#ffe29a"
			extra = ` data-emphasis="compare"`
		}
		if s.HasResult && i == max(0, s.N) {
			stroke = "#9bf0bb"
		}
		fmt.Fprintf(&b, `<g data-entity="cell:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"%s><rect width="%d" height="48" rx="8" fill="%s" stroke="%s" stroke-width="2"/>`, i, x, y, x, y, extra, cellW, colors[i%5], stroke)
		fibonacciText(&b, 8, 30, label, "#ffffff", 14)
		fibonacciText(&b, 0, 66, fmt.Sprintf("F(%d)%s", i, func() string {
			if s.HasResult && i == max(0, s.N) {
				return " ✓"
			}
			return ""
		}()), "#eff5ed", 12)
		b.WriteString(`</g>`)
	}
	if recurrence {
		labels := []string{"i-2", "i-1", "sum"}
		for k := 0; k < 3; k++ {
			x, y := 16+k*((w-44)/3+6), cardY
			cw := (w - 44) / 3
			if narrow {
				x = 16
				y = cardY + k*100
				cw = w - 32
			}
			value := "?"
			entity := fmt.Sprintf("formula:%d", k)
			alias := ""
			color := "#22372c"
			if k < 2 && s.Sources[k] >= 0 && s.Sources[k] < len(s.Values) {
				index := s.Sources[k]
				value = strconv.Itoa(s.Values[index])
				color = colors[index%5]
				if s.Phase == "write" {
					entity = fmt.Sprintf("operand:%d:%d", index, s.Index)
					alias = fmt.Sprintf(` data-from-entity="cell:%d"`, index)
				}
			}
			if k == 2 && s.Phase == "write" {
				value = strconv.Itoa(s.Values[s.Index])
				color = colors[s.Index%5]
			}
			if k < 2 {
				sx, sy := position(s.Sources[k])
				fmt.Fprintf(&b, `<path d="M%d %d Q%d %d %d %d" fill="none" stroke="#ffe29a" stroke-dasharray="5 4"/><text x="%d" y="%d" fill="#ffe29a" font-size="12">%s</text>`, sx+cellW/2, sy+48, sx+cellW/2, boardEnd+24, x+20, y, x+24, y-8, labels[k])
			}
			fmt.Fprintf(&b, `<g data-entity="%s" data-x="%d" data-y="%d" transform="translate(%d %d)"%s><rect width="%d" height="76" rx="10" fill="%s" stroke="#ffe29a"/>`, entity, x, y, x, y, alias, cw, color)
			fibonacciText(&b, 8, 23, labels[k], "#ffffff", 14)
			fibonacciText(&b, 8, 53, value, "#ffffff", 16)
			b.WriteString(`</g>`)
		}
		equation := fmt.Sprintf("%d + %d = ?", s.Values[s.Sources[0]], s.Values[s.Sources[1]])
		if s.Phase == "write" {
			equation = fmt.Sprintf("%d + %d = %d", s.Values[s.Sources[0]], s.Values[s.Sources[1]], s.Values[s.Index])
		}
		if len(equation)*9 > w-32 {
			fibonacciText(&b, 16, h-64, strconv.Itoa(s.Values[s.Sources[0]])+" +", "#eff5ed", 14)
			fibonacciText(&b, 16, h-44, strconv.Itoa(s.Values[s.Sources[1]]), "#eff5ed", 14)
			answer := "= ?"
			if s.Phase == "write" {
				answer = "= " + strconv.Itoa(s.Values[s.Index])
			}
			fibonacciText(&b, 16, h-24, answer, "#eff5ed", 14)
		} else {
			fibonacciText(&b, 16, h-32, equation, "#eff5ed", 14)
		}
	} else {
		fibonacciText(&b, 16, h-32, "Store once; reuse known values.", "#bbcab9", 14)
	}
	b.WriteString(`</g>`)
	return b.String()
}

func fibonacciText(b *strings.Builder, x, y int, text, color string, size int) {
	fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" font-size="%d">%s</text>`, x, y, color, size, html.EscapeString(text))
}
