package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type KnapsackLessonState struct {
	TableState
	Weights, Values          []int
	DP                       [][]int
	Known                    [][]bool
	Capacity, Item, Column   int
	Sources                  [2][2]int
	Fits                     bool
	Exclude, Include, Chosen int
	Phase                    string
	Result                   int
	HasResult                bool
}

// RenderKnapsackLessonSVG shows only stored evidence; both candidates use the previous row.
func RenderKnapsackLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(KnapsackLessonState)
	if !ok || len(s.DP) == 0 || len(s.DP[0]) == 0 {
		return `<g data-lesson="knapsack" data-phase="empty" data-scene-width="320" data-scene-height="100"><title>Knapsack: no table</title><text x="16" y="40" fill="#eff5ed" font-size="14">No knapsack state available</text></g>`
	}
	// ponytail: full matrix stays locally scrollable; virtualize only if public limits grow.
	cols := len(s.DP[0])
	rows := len(s.DP)
	sharedSource := s.Fits && s.Sources[0] == s.Sources[1] && s.Sources[0][0] >= 0
	cardCols := max(1, (max(320, min(width, 1440))-32)/200)
	cardRows := (len(s.Weights) + cardCols - 1) / cardCols
	top := 76 + cardRows*88
	w := max(320, max(min(width, 1440), 96+cols*84))
	h := max(max(1, min(height, 32768)), top+rows*44+200)
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="knapsack" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(s.Phase), w, h, w, html.EscapeString(step.Explanation))
	fmt.Fprintf(&b, `<text x="16" y="26">0/1 Knapsack · capacity %d</text><text x="16" y="50" fill="#bbcab9">Rows: first i items · columns: capacity c</text>`, s.Capacity)
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, weight := range s.Weights {
		x, y := 16+(i%cardCols)*200, 64+(i/cardCols)*88
		stroke := "#bbcab9"
		if i+1 == s.Item {
			stroke = "#ffe29a"
		}
		value := 0
		if i < len(s.Values) {
			value = s.Values[i]
		}
		fmt.Fprintf(&b, `<g data-entity="item:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"><rect width="188" height="76" rx="10" fill="%s" stroke="%s" stroke-width="2"/><text x="10" y="21">Item %d</text><text x="10" y="43">weight %d</text><text x="10" y="65">value %d</text></g>`, i, x, y, x, y, palette[i%5], stroke, i+1, weight, value)
	}
	for c := 0; c < cols; c++ {
		fmt.Fprintf(&b, `<text x="%d" y="%d" text-anchor="middle" fill="#bbcab9">c=%d</text>`, 80+c*84+40, top-8, c)
	}
	for r, row := range s.DP {
		fmt.Fprintf(&b, `<text x="16" y="%d" fill="#bbcab9">i=%d</text>`, top+r*44+26, r)
		for c, value := range row {
			x, y := 80+c*84, top+r*44
			stroke := "#52634e"
			label := "?"
			if r < len(s.Known) && c < len(s.Known[r]) && s.Known[r][c] {
				label = fmt.Sprint(value)
			}
			for k, src := range s.Sources {
				if sharedSource && k == 1 {
					continue
				}
				if src == [2]int{r, c} && s.Column >= 0 {
					stroke = []string{"#8bd5ff", "#c45d75"}[k]
				}
			}
			if s.HasResult && r == rows-1 && c == cols-1 {
				stroke = "#9bf0bb"
			}
			fmt.Fprintf(&b, `<g data-entity="cell:%d:%d" data-x="%d" data-y="%d" transform="translate(%d %d)"><rect width="80" height="40" rx="6" fill="#192923" stroke="%s" stroke-width="2"/><text x="40" y="26" text-anchor="middle">%s</text></g>`, r, c, x, y, x, y, stroke, html.EscapeString(label))
		}
	}
	if s.Item > 0 && s.Item <= rows-1 && s.Column >= 0 && s.Column < cols {
		x, y := 80+s.Column*84, top+s.Item*44
		fmt.Fprintf(&b, `<g data-entity="cursor:target" data-emphasis="compare" data-x="%d" data-y="%d" transform="translate(%d %d)"><rect width="80" height="40" rx="6" fill="none" stroke="#ffe29a" stroke-width="3"/></g>`, x, y, x, y)
		for k, src := range s.Sources {
			if sharedSource && k == 1 {
				continue
			}
			if src[0] < 0 || src[1] < 0 {
				continue
			}
			color := []string{"#8bd5ff", "#c45d75"}[k]
			fmt.Fprintf(&b, `<path d="M%d %d L%d %d" fill="none" stroke="%s" stroke-width="2" stroke-dasharray="4 3"/><path d="M%d %d l-4 -7 m4 7 l4 -7" fill="none" stroke="%s"/>`, 80+src[1]*84+40, top+src[0]*44+40, x+40, y, color, x+40, y, color)
		}
	}
	y := top + rows*44 + 24
	if s.Column >= 0 {
		fmt.Fprintf(&b, `<text x="16" y="%d" fill="#8bd5ff">Exclude: previous row = %d</text>`, y, s.Exclude)
		if s.Fits {
			fmt.Fprintf(&b, `<text x="16" y="%d" fill="#eea2b3">Include: previous row + value = %d</text>`, y+26, s.Include)
		} else {
			fmt.Fprintf(&b, `<text x="16" y="%d" fill="#ff9b9b">Include blocked: weight exceeds capacity %d</text>`, y+26, s.Column)
		}
		if s.Phase == "write" {
			fmt.Fprintf(&b, `<text x="16" y="%d" fill="#ffe29a">Write %d · no subset backtracking</text>`, y+52, s.Chosen)
		} else {
			fmt.Fprintf(&b, `<text x="16" y="%d" fill="#ffe29a">Compare candidates · target not written yet</text>`, y+52)
		}
		if sharedSource {
			fmt.Fprintf(&b, `<text x="16" y="%d">Both use the same previous-row cell</text>`, y+78)
		} else if s.Fits && s.Include == s.Exclude {
			fmt.Fprintf(&b, `<text x="16" y="%d">Equal scores · no inclusion tie-break</text>`, y+78)
		}
		weight := 0
		if s.Item > 0 && s.Item <= len(s.Weights) {
			weight = s.Weights[s.Item-1]
		}
		fmt.Fprintf(&b, `<text x="16" y="%d" fill="#bbcab9">Current capacity %d · item weight %d</text>`, y+104, s.Column, weight)
		barWidth := 240
		fill := 0
		if s.Capacity > 0 {
			fill = barWidth * max(0, min(s.Column, s.Capacity)) / s.Capacity
		}
		fmt.Fprintf(&b, `<rect x="16" y="%d" width="240" height="8" rx="4" fill="#52634e"/><rect x="16" y="%d" width="%d" height="8" rx="4" fill="#8bd5ff"/>`, y+116, y+116, fill)
	} else if s.HasResult {
		fmt.Fprintf(&b, `<text x="16" y="%d" fill="#9bf0bb">✓ Optimal value: %d</text>`, y, s.Result)
	} else {
		fmt.Fprintf(&b, `<text x="16" y="%d" fill="#bbcab9">Base row known · ? means not computed</text>`, y)
	}
	b.WriteString(`</g>`)
	return b.String()
}
