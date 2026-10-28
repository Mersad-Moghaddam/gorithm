package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"math"
	"strings"
)

type KMeansLessonState struct {
	GraphState
	Points, Centroids, PreviousCentroids, Sums                      [][2]float64
	Assignments, Counts                                             []int
	DistanceSquared                                                 []float64
	Iteration, ActivePoint, Closest, PreviousAssignment, RequestedK int
	Changed, Moved, Converged                                       bool
	Phase, StopReason                                               string
	XMin, XMax, YMin, YMax                                          float64
}

func kmeansLessonColor(i int) string {
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	if i < 0 {
		return "#bbcab9"
	}
	return colors[i%len(colors)]
}

// RenderKMeansLessonSVG renders only recorded numeric state; no clustering runs here.
func RenderKMeansLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(KMeansLessonState)
	if !ok {
		return `<g data-lesson="kmeans" data-phase="invalid-input" data-scene-width="320" data-scene-height="100"><title>K-means state unavailable</title></g>`
	}
	w := float64(width)
	if w < 320 || w > 1440 {
		w = 320
	}
	mobile := w < 640
	inset := 24.0
	if mobile {
		inset = 16
	}
	plotX, plotY, plotW, plotH := inset+24, 64.0, w-inset*2-48, 300.0
	keyY := 410.0
	cardX, cardY, cardW := inset, keyY+float64(len(s.Points)+len(s.Centroids))*24+28, w-2*inset
	if !mobile {
		plotW = w - 2*inset - 360
	}
	cardHeight := float64(len(s.Centroids)) * 146
	pointY := math.Max(410, cardY+cardHeight+24)
	equationY := pointY + 30 + float64(len(s.Points))*68
	sceneH := equationY + 136
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="kmeans" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="sans-serif" font-size="14" fill="#eff5ed"><title>%s</title><defs><marker id="kmeans-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#8bd5ff"/></marker></defs>`, html.EscapeString(s.Phase), w, sceneH, w, html.EscapeString("K-means: "+s.Phase+" "+s.StopReason))
	text := func(x, y float64, color, value string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" fill="%s">%s</text>`, x, y, color, html.EscapeString(value))
	}
	text(inset, 26, "#ffe29a", fmt.Sprintf("Iteration %d · %s", s.Iteration, s.Phase))
	effectiveK := min(s.RequestedK, len(s.Points))
	text(inset, 48, "#bbcab9", fmt.Sprintf("k requested %d · effective %d", s.RequestedK, effectiveK))
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	xmin, xmax, ymin, ymax := s.XMin, s.XMax, s.YMin, s.YMax
	if !finite(xmin) || !finite(xmax) || xmax <= xmin {
		xmin, xmax = -1, 1
	}
	if !finite(ymin) || !finite(ymax) || ymax <= ymin {
		ymin, ymax = -1, 1
	}
	scale := math.Min(plotW/(xmax-xmin), plotH/(ymax-ymin))
	if !finite(scale) || scale <= 0 {
		scale = 1
	}
	pos := func(p [2]float64) (float64, float64) {
		if !finite(p[0]) || !finite(p[1]) {
			return plotX + plotW/2, plotY + plotH/2
		}
		return plotX + plotW/2 + (p[0]-(xmin/2+xmax/2))*scale, plotY + plotH/2 - (p[1]-(ymin/2+ymax/2))*scale
	}
	fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" rx="12" fill="#192923" stroke="#52634e"/>`, plotX-24, plotY-8, plotW+48, plotH+48)
	for _, group := range []struct {
		kind   string
		points [][2]float64
		offset int
	}{{"point", s.Points, 0}, {"centroid", s.Centroids, len(s.Points)}} {
		for i, p := range group.points {
			x, y := pos(p)
			rowY := keyY + float64(group.offset+i)*24
			fmt.Fprintf(&b, `<path data-marker-link="%s:%d" d="M%.2f %.2fL%.2f %.2f" fill="none" stroke="#bbcab9" opacity=".45"/>`, group.kind, i, x, y, inset+8, rowY-5)
		}
	}
	for i, p := range s.Points {
		x, y := pos(p)
		if i < len(s.Assignments) {
			j := s.Assignments[i]
			if j >= 0 && j < len(s.Centroids) {
				cx, cy := pos(s.Centroids[j])
				fmt.Fprintf(&b, `<path d="M%.2f %.2fL%.2f %.2f" stroke="%s" opacity=".3"/>`, x, y, cx, cy, kmeansLessonColor(j))
			}
		}
	}
	if s.ActivePoint >= 0 && s.ActivePoint < len(s.Points) {
		x, y := pos(s.Points[s.ActivePoint])
		for j, c := range s.Centroids {
			cx, cy := pos(c)
			color := "#ffe29a"
			if j == s.Closest {
				color = "#8bd5ff"
			}
			fmt.Fprintf(&b, `<path d="M%.2f %.2fL%.2f %.2f" stroke="%s" stroke-dasharray="5 5"/>`, x, y, cx, cy, color)
		}
	}
	for i, p := range s.Points {
		x, y := pos(p)
		cluster := -1
		if i < len(s.Assignments) {
			cluster = s.Assignments[i]
		}
		fmt.Fprintf(&b, `<g data-entity="point:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><title>point %d (%s, %s)</title><circle r="6" fill="%s" stroke="#eff5ed"/>`, i, x, y, x, y, i, kmeansLessonNumber(p[0]), kmeansLessonNumber(p[1]), kmeansLessonColor(cluster))
		if i == s.ActivePoint {
			b.WriteString(`<circle r="11" fill="none" stroke="#ffe29a" stroke-width="3" data-emphasis="compare"/>`)
		}
		b.WriteString(`</g>`)
	}
	for j, c := range s.Centroids {
		x, y := pos(c)
		if s.Phase == "means" && j < len(s.PreviousCentroids) {
			ox, oy := pos(s.PreviousCentroids[j])
			fmt.Fprintf(&b, `<path d="M%.2f %.2fL%.2f %.2f" stroke="#8bd5ff" marker-end="url(#kmeans-arrow)"/><path d="M%.2f %.2fl8 -8l8 8l-8 8Z" fill="none" stroke="#8bd5ff" stroke-dasharray="3 3"/>`, ox, oy, x, y, ox-8, oy)
		}
		fmt.Fprintf(&b, `<g data-entity="centroid:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)" data-motion="linear"><title>centroid %d (%s, %s)</title><path d="M0 -10L10 0L0 10L-10 0Z M-14 0H14 M0 -14V14" fill="%s" stroke="#eff5ed" stroke-width="2"/></g>`, j, x, y, x, y, j, kmeansLessonNumber(c[0]), kmeansLessonNumber(c[1]), kmeansLessonColor(j))
		cy := cardY + float64(j)*146
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="138" rx="8" fill="#22372c" stroke="%s"/>`, cardX, cy, cardW, kmeansLessonColor(j))
		text(cardX+8, cy+19, "#eff5ed", fmt.Sprintf("C%d", j))
		text(cardX+8, cy+39, "#eff5ed", "x: "+kmeansLessonNumber(c[0]))
		text(cardX+8, cy+59, "#eff5ed", "y: "+kmeansLessonNumber(c[1]))
		if j < len(s.Counts) && j < len(s.Sums) {
			n := s.Counts[j]
			if n == 0 {
				text(cardX+8, cy+79, "#8bd5ff", "count 0; no mean")
				text(cardX+8, cy+99, "#ffe29a", "retained: no members")
			} else {
				text(cardX+8, cy+79, "#8bd5ff", "sum x: "+kmeansLessonNumber(s.Sums[j][0]))
				text(cardX+8, cy+99, "#8bd5ff", "sum y: "+kmeansLessonNumber(s.Sums[j][1]))
				text(cardX+8, cy+119, "#bbcab9", fmt.Sprintf("count: %d · mean = sum / count", n))
			}
		} else if j < len(s.DistanceSquared) {
			text(cardX+8, cy+79, "#8bd5ff", fmt.Sprintf("d²(P%d,C%d):", s.ActivePoint, j))
			text(cardX+8, cy+99, "#8bd5ff", kmeansLessonNumber(s.DistanceSquared[j]))
			if j == s.Closest {
				text(cardX+8, cy+119, "#ffe29a", "nearest; first on ties")
			}
		}
	}
	for _, group := range []struct {
		kind, prefix  string
		count, offset int
	}{{"point", "P", len(s.Points), 0}, {"centroid", "C", len(s.Centroids), len(s.Points)}} {
		for i := 0; i < group.count; i++ {
			y := keyY + float64(group.offset+i)*24
			fmt.Fprintf(&b, `<g data-marker-key="%s:%d"><rect x="%.2f" y="%.2f" width="64" height="20" rx="4" fill="#192923"/><text x="%.2f" y="%.2f">%s%d</text></g>`, group.kind, i, inset, y-16, inset+12, y, group.prefix, i)
		}
	}
	text(inset, pointY, "#ffe29a", "Points · assigned cluster")
	for i, p := range s.Points {
		cluster := "unassigned"
		if i < len(s.Assignments) && s.Assignments[i] >= 0 {
			cluster = fmt.Sprintf("C%d", s.Assignments[i])
		}
		y := pointY + 24 + float64(i)*68
		fmt.Fprintf(&b, `<g data-point-row="%d">`, i)
		text(inset, y, "#eff5ed", fmt.Sprintf("P%d · %s", i, cluster))
		text(inset, y+20, "#bbcab9", "x: "+kmeansLessonNumber(p[0]))
		text(inset, y+40, "#bbcab9", "y: "+kmeansLessonNumber(p[1]))
		b.WriteString(`</g>`)
	}
	text(inset, equationY, "#ffe29a", "d²(P,C) = (Px − Cx)²")
	text(inset, equationY+22, "#ffe29a", "                 + (Py − Cy)²")
	text(inset, equationY+48, "#8bd5ff", "Cj = sum(member points) / count")
	text(inset, equationY+74, "#bbcab9", fmt.Sprintf("changed %t · moved %t", s.Changed, s.Moved))
	if s.StopReason != "" {
		text(inset, equationY+100, "#9bf0bb", "Stop: "+s.StopReason)
	}
	b.WriteString(`</g>`)
	return b.String()
}

func kmeansLessonNumber(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "invalid"
	}
	return fmt.Sprintf("%.4g", v)
}
