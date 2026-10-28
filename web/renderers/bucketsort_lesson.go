package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type BucketsortLessonState struct {
	ArrayState
	IDs                                                                 []int
	Buckets, BucketIDs                                                  [][]int
	Min, Max, BucketCount, Scan, ActiveBucket, BucketOffset, WriteIndex int
	Phase                                                               string
}

// RenderBucketsortLessonSVG shows actual array memory and bucket copies, not a second sort.
func RenderBucketsortLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(BucketsortLessonState)
	if !ok {
		return `<g data-lesson="bucketsort"><title>Bucket Sort: no typed state</title></g>`
	}
	width = max(320, min(width, 12000))
	n := len(s.Values)
	// ponytail: full-size numeric labels need local scrolling; keep every array slot readable rather than shrinking bars.
	sceneW := max(width, 64+n*88)
	if s.BucketCount > 0 && (s.Min < -999 || s.Max > 999) {
		sceneW = max(sceneW, 560)
	}
	cols := 1
	if width >= 700 {
		cols = 2
	}
	if width >= 1100 {
		cols = 5
	}
	laneW := float64(width-64) / float64(cols)
	laneCols := max(1, int((laneW-24)/88))
	rows := (s.BucketCount + cols - 1) / cols
	rowHeights := make([]float64, rows)
	for i := 0; i < s.BucketCount && i < len(s.Buckets); i++ {
		row := i / cols
		h := float64(60 + ((len(s.Buckets[i])+laneCols-1)/laneCols)*60)
		rowHeights[row] = math.Max(rowHeights[row], math.Max(h, 112))
	}
	sceneH := 360.0
	for _, h := range rowHeights {
		sceneH += h + 16
	}
	sceneH = math.Max(sceneH, float64(max(360, height)))
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="bucketsort" data-phase="%s" data-scene-width="%d" data-scene-height="%.0f" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>Bucket Sort: %s. Array memory and exact integer-hash bucket copies.</title>`, html.EscapeString(s.Phase), sceneW, sceneH, sceneW, html.EscapeString(s.Phase))
	b.WriteString(`<text x="32" y="28">Array memory / bucket copies</text>`)
	if s.BucketCount > 0 {
		fmt.Fprintf(&b, `<text x="32" y="52">(x-min)*4/(max-min)</text><text x="32" y="74">Range %d…%d</text>`, s.Min, s.Max)
		if s.Phase == "scatter" && s.Scan >= 0 && s.Scan < n && s.ActiveBucket >= 0 {
			fmt.Fprintf(&b, `<text x="32" y="96">(%d − %d)*4/(%d − %d)</text><text x="32" y="118">Bucket %d</text>`, s.Values[s.Scan], s.Min, s.Max, s.Min, s.ActiveBucket)
		}
	} else {
		b.WriteString(`<text x="32" y="52">No buckets allocated yet.</text><text x="32" y="74">Equal range: no hashing.</text>`)
	}
	if n == 0 {
		b.WriteString(`<text x="32" y="120">Empty array</text>`)
	}
	low, high := math.Min(0, float64(s.Min)), math.Max(0, float64(s.Max))
	span := high - low
	if span == 0 {
		span = 1
	}
	baseline := 144 + high/span*100
	fmt.Fprintf(&b, `<line x1="24" y1="%.2f" x2="%d" y2="%.2f" stroke="#bbcab9"/><text x="24" y="%.2f" font-size="12">0</text>`, baseline, sceneW-24, baseline, baseline-5)
	for i, v := range s.Values {
		id := i
		if i < len(s.IDs) {
			id = s.IDs[i]
		}
		x := float64(76 + i*88)
		h := math.Abs(float64(v)) / span * 100
		y := baseline
		if v > 0 {
			y -= h
		}
		if v == 0 {
			h = 2
			y -= 1
		}
		stroke := "#bbcab9"
		done := false
		for _, j := range s.Sorted {
			if j == i {
				done = true
				stroke = "#9bf0bb"
			}
		}
		if i == s.Scan || (s.Phase == "gather" && i == s.WriteIndex) {
			stroke = "#ffe29a"
		}
		alias := ""
		if step.LineID == "L17" && i == s.WriteIndex {
			alias = fmt.Sprintf(` data-from-entity="bucket%d:%d"`, s.ActiveBucket, id)
		}
		fmt.Fprintf(&b, `<g data-entity="main%d:%d" data-id="%d" data-role="array" data-x="%.2f" data-y="%.2f" data-motion="linear"%s transform="translate(%.2f %.2f)"><rect x="-36" y="0" width="72" height="%.2f" rx="3" fill="%s" stroke="%s" stroke-width="2"/><text x="0" y="-8" text-anchor="middle">%d</text></g>`, i, id, id, x, y, alias, x, y, h, bucketsortLessonColor(id), stroke, v)
		mark := ""
		if done {
			mark = " ✓"
		}
		fmt.Fprintf(&b, `<text x="%.2f" y="266" text-anchor="middle" font-size="12">[%d] #%d%s</text>`, x, i, id, mark)
	}
	if s.Phase == "gather" {
		fmt.Fprintf(&b, `<text x="32" y="296" fill="#8bd5ff">Gather cursor: write [%d]</text>`, s.WriteIndex)
	} else if s.Phase == "local" {
		b.WriteString(`<text x="32" y="296">Local sort: before / completed</text>`)
	}
	y := 320.0
	for i := 0; i < s.BucketCount && i < len(s.Buckets); i++ {
		row, col := i/cols, i%cols
		if col == 0 && row > 0 {
			y += rowHeights[row-1] + 16
		}
		x := 32 + float64(col)*laneW
		stroke := "#52634e"
		if i == s.ActiveBucket {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="8" fill="#192923" stroke="%s"/><text x="%.2f" y="%.2f">Bucket %d%s</text>`, x, y, laneW-12, rowHeights[row], stroke, x+12, y+22, i, bucketsortLessonBucketLabel(i))
		if len(s.Buckets[i]) == 0 {
			fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" fill="#bbcab9">Empty</text>`, x+12, y+58)
		}
		for j, v := range s.Buckets[i] {
			id := j
			if i < len(s.BucketIDs) && j < len(s.BucketIDs[i]) {
				id = s.BucketIDs[i][j]
			}
			tx, ty := x+50+float64(j%laneCols)*88, y+48+float64(j/laneCols)*60
			alias := ""
			if step.LineID == "L9" && i == s.ActiveBucket && j == s.BucketOffset {
				alias = fmt.Sprintf(` data-from-entity="main%d:%d"`, s.Scan, id)
			}
			fmt.Fprintf(&b, `<g data-entity="bucket%d:%d" data-id="%d" data-role="bucket" data-x="%.2f" data-y="%.2f" data-motion="linear"%s transform="translate(%.2f %.2f)"><rect x="-36" y="0" width="72" height="28" rx="4" fill="%s"/><text x="0" y="19" text-anchor="middle" fill="#ffffff">%d</text><text x="0" y="44" text-anchor="middle" font-size="12">#%d · [%d]</text></g>`, i, id, id, tx, ty, alias, tx, ty, bucketsortLessonColor(id), v, id, j)
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}

func bucketsortLessonColor(id int) string {
	colors := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	return colors[(id%5+5)%5]
}
func bucketsortLessonBucketLabel(i int) string {
	if i == 4 {
		return " · max only"
	}
	return ""
}
