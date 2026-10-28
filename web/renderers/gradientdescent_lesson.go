package renderers

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type GradientDescentLessonState struct {
	TableState
	X, Y, Predictions, Residuals                               []float64
	Epoch, Sample                                              int
	Phase                                                      string
	W, B, BeforeW, BeforeB, LR                                 float64
	SumSquaredError, SumErrorX, SumError, GradientW, GradientB float64
	LossBefore, LossCurrent                                    float64
	LossBeforeValid, LossCurrentValid                          bool
	History                                                    []GradientDescentEpoch
	StopReason                                                 string
	DataXMin, DataXMax, DataYMin, DataYMax                     float64
	SliceWMin, SliceWMax, SliceLossMin, SliceLossMax           float64
	HistoryLossMin, HistoryLossMax                             float64
}

type GradientDescentEpoch struct {
	Epoch      int
	W, B, Loss float64
}

func gradientDescentLessonFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func gradientDescentLessonLoss(s GradientDescentLessonState, w, b float64) float64 {
	if len(s.X) == 0 || len(s.X) != len(s.Y) {
		return 0
	}
	var sum float64
	for i, x := range s.X {
		e := w*x + b - s.Y[i]
		sum += e * e / float64(len(s.X))
	}
	return sum
}

func gradientDescentLessonMap(v, lo, hi, start, span float64) float64 {
	if !gradientDescentLessonFinite(v) || !gradientDescentLessonFinite(lo) || !gradientDescentLessonFinite(hi) || hi <= lo {
		return start + span/2
	}
	// Scaling operands first also handles domains whose difference would overflow.
	scale := math.Max(math.Abs(lo), math.Abs(hi))
	if scale == 0 {
		scale = 1
	}
	t := (v/scale - lo/scale) / (hi/scale - lo/scale)
	if !gradientDescentLessonFinite(t) {
		t = .5
	}
	return start + math.Max(-4, math.Min(5, t))*span
}

// RenderGradientDescentLessonSVG returns one authoritative endpoint, without timers.
func RenderGradientDescentLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(GradientDescentLessonState)
	if !ok {
		return `<g data-lesson="gradientdescent" data-phase="invalid" data-scene-width="320" data-scene-height="160"><title>Gradient descent state unavailable</title><text x="16" y="40" fill="#eff5ed" font-size="14">Gradient descent state unavailable</text></g>`
	}
	mobile := width < 640
	sceneW := float64(width)
	if sceneW < 320 {
		sceneW = 320
	}
	if sceneW > 1440 {
		sceneW = 1440
	}
	sceneH := 920.
	if mobile {
		sceneH = 1440
	}
	inset := 24.
	if mobile {
		inset = 16
	}
	plotW := (sceneW - 3*inset) / 2
	plotH := 230.
	x1, y1 := inset, 64.
	x2, y2 := 2*inset+plotW, 64.
	if mobile {
		plotW = sceneW - 2*inset
		x2 = inset
		y2 = 392
	}
	historyY := 390.
	formulaY := 540.
	if mobile {
		historyY = 728
		formulaY = 868
	}
	var out strings.Builder
	fmt.Fprintf(&out, `<g data-lesson="gradientdescent" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui,sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(s.Phase), sceneW, sceneH, sceneW, html.EscapeString("Linear regression: "+s.Phase+". "+step.Explanation))
	text := func(x, y float64, color, value string) {
		fmt.Fprintf(&out, `<text x="%.3f" y="%.3f" fill="%s">%s</text>`, x, y, color, html.EscapeString(value))
	}
	line := func(xa, ya, xb, yb float64, color, extra string) {
		if gradientDescentLessonFinite(xa) && gradientDescentLessonFinite(ya) && gradientDescentLessonFinite(xb) && gradientDescentLessonFinite(yb) {
			fmt.Fprintf(&out, `<path d="M%.3f %.3fL%.3f %.3f" fill="none" stroke="%s" stroke-width="2" %s/>`, xa, ya, xb, yb, color, extra)
		}
	}
	wrap := func(y float64, value string) float64 {
		maxChars := int((sceneW - 2*inset) / 7.8)
		words := strings.Fields(value)
		row := ""
		for _, word := range words {
			if len([]rune(row+" "+word)) > maxChars && row != "" {
				text(inset, y, "#eff5ed", row)
				y += 22
				row = ""
			}
			if row != "" {
				row += " "
			}
			row += word
		}
		if row != "" {
			text(inset, y, "#eff5ed", row)
			y += 22
		}
		return y
	}
	text(x1, 28, "#eff5ed", fmt.Sprintf("Epoch %d · %s", s.Epoch, s.Phase))
	text(x1, y1-16, "#bbcab9", fmt.Sprintf("Data / model: w=%.5g, b=%.5g", s.W, s.B))
	text(x2, y2-34, "#bbcab9", "MSE versus w")
	text(x2, y2-14, "#bbcab9", fmt.Sprintf("b held at %.5g", s.B))
	fmt.Fprintf(&out, `<defs><clipPath id="gradientdescent-data-clip"><rect x="%.3f" y="%.3f" width="%.3f" height="%.3f"/></clipPath><clipPath id="gradientdescent-slice-clip"><rect x="%.3f" y="%.3f" width="%.3f" height="%.3f"/></clipPath></defs>`, x1, y1, plotW, plotH, x2, y2, plotW, plotH)
	for _, p := range [][2]float64{{x1, y1}, {x2, y2}} {
		fmt.Fprintf(&out, `<rect x="%.3f" y="%.3f" width="%.3f" height="%.3f" rx="8" fill="#16251f" stroke="#52634e"/>`, p[0], p[1], plotW, plotH)
	}
	dx := func(v float64) float64 { return gradientDescentLessonMap(v, s.DataXMin, s.DataXMax, x1+8, plotW-16) }
	dy := func(v float64) float64 {
		return gradientDescentLessonMap(v, s.DataYMin, s.DataYMax, y1+plotH-8, -plotH+16)
	}
	wx := func(v float64) float64 { return gradientDescentLessonMap(v, s.SliceWMin, s.SliceWMax, x2+8, plotW-16) }
	ly := func(v float64) float64 {
		return gradientDescentLessonMap(v, s.SliceLossMin, s.SliceLossMax, y2+plotH-8, -plotH+16)
	}
	// Clip plotted geometry, keep point labels outside clip; movable groups have no transformed parent.
	fmt.Fprintf(&out, `<g clip-path="url(#gradientdescent-data-clip)">`)
	shift := dy(s.B) - dy(0)
	fmt.Fprintf(&out, `<g data-entity="model:line" data-x="0" data-y="%.3f" transform="translate(0 %.3f)">`, shift, shift)
	line(dx(s.DataXMin), dy(s.W*s.DataXMin), dx(s.DataXMax), dy(s.W*s.DataXMax), "#ee8b9b", `aria-label="fitted model"`)
	out.WriteString(`</g>`)
	for i, x := range s.X {
		if i >= len(s.Y) {
			break
		}
		if i < s.Sample || (i == s.Sample && s.Phase != "predict") || s.LossBeforeValid {
			line(dx(x), dy(s.Y[i]), dx(x), dy(s.W*x+s.B), "#ffe29a", `stroke-dasharray="5 4" aria-label="residual"`)
		}
	}
	out.WriteString(`</g>`)
	colors := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for pass := 0; pass < 2; pass++ {
		for i, x := range s.X {
			if i >= len(s.Y) {
				break
			}
			active := i == s.Sample
			if active != (pass == 1) {
				continue
			}
			px, py := dx(x), dy(s.Y[i])
			stroke, strokeWidth := "#eff5ed", 1
			if active {
				stroke, strokeWidth = "#ffe29a", 3
			}
			fmt.Fprintf(&out, `<g data-entity="sample:P%d" data-x="%.3f" data-y="%.3f" transform="translate(%.3f %.3f)"><title>P%d: x=%.6g, y=%.6g</title><circle r="6" fill="%s" stroke="%s" stroke-width="%d"/>`, i, px, py, px, py, i, x, s.Y[i], colors[i%len(colors)], stroke, strokeWidth)
			if active {
				labelX, labelY, anchor := 8, -8, "start"
				if px > x1+plotW-42 {
					labelX, anchor = -8, "end"
				}
				if py < y1+22 {
					labelY = 20
				}
				fmt.Fprintf(&out, `<text x="%d" y="%d" text-anchor="%s">P%d</text>`, labelX, labelY, anchor, i)
			}
			out.WriteString(`</g>`)
		}
	}
	text(x1, y1+plotH+20, "#bbcab9", fmt.Sprintf("x: %.4g … %.4g", s.DataXMin, s.DataXMax))
	text(x1, y1+plotH+40, "#bbcab9", fmt.Sprintf("y: %.4g … %.4g", s.DataYMin, s.DataYMax))
	out.WriteString(`<g clip-path="url(#gradientdescent-slice-clip)">`)
	var curve strings.Builder
	for i := 0; i <= 64; i++ {
		u := s.SliceWMin*(1-float64(i)/64) + s.SliceWMax*(float64(i)/64)
		loss := gradientDescentLessonLoss(s, u, s.B)
		if !gradientDescentLessonFinite(loss) {
			continue
		}
		cmd := "L"
		if curve.Len() == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&curve, "%s%.3f %.3f ", cmd, wx(u), ly(loss))
	}
	fmt.Fprintf(&out, `<path d="%s" fill="none" stroke="#aa92e7" stroke-width="3"/>`, curve.String())
	measuredW := s.W
	if s.Phase == "weight" {
		measuredW = s.BeforeW
	}
	currentLoss := gradientDescentLessonLoss(s, s.W, s.B)
	// Analytic derivative of the displayed slice, not another optimizer iteration.
	var xx, xy, deriv float64
	for i, x := range s.X {
		if i >= len(s.Y) {
			break
		}
		xx += x * x
		xy += x * (s.Y[i] - s.B)
		deriv += 2 * x * (measuredW*x + s.B - s.Y[i]) / float64(len(s.X))
	}
	baseLoss := gradientDescentLessonLoss(s, measuredW, s.B)
	half := (s.SliceWMax - s.SliceWMin) * .1
	if gradientDescentLessonFinite(baseLoss) && gradientDescentLessonFinite(deriv) {
		line(wx(measuredW-half), ly(baseLoss-deriv*half), wx(measuredW+half), ly(baseLoss+deriv*half), "#ffe29a", `aria-label="tangent"`)
	}
	if s.Phase == "loss" || s.Phase == "weight" {
		target := s.BeforeW - s.LR*s.GradientW
		tx, ty := wx(target), ly(gradientDescentLessonLoss(s, target, s.B))
		line(wx(s.BeforeW), ly(gradientDescentLessonLoss(s, s.BeforeW, s.B)), tx, ty, "#8bd5ff", `stroke-dasharray="6 4" aria-label="discrete weight update"`)
		fmt.Fprintf(&out, `<circle cx="%.3f" cy="%.3f" r="9" fill="none" stroke="#8bd5ff" stroke-width="2"><title>weight target</title></circle>`, tx, ty)
	}
	out.WriteString(`</g>`)
	if xx > 0 && gradientDescentLessonFinite(xx) && gradientDescentLessonFinite(xy/xx) {
		u := xy / xx
		v := gradientDescentLessonLoss(s, u, s.B)
		if gradientDescentLessonFinite(v) {
			fmt.Fprintf(&out, `<circle cx="%.3f" cy="%.3f" r="5" fill="#9bf0bb"/>`, wx(u), ly(v))
		}
		text(x2, y2+plotH+20, "#9bf0bb", "slice minimum; b fixed")
		text(x2, y2+plotH+40, "#9bf0bb", fmt.Sprintf("w*=%.5g", u))
	} else {
		text(x2, y2+plotH+20, "#9bf0bb", "flat slice:")
		text(x2, y2+plotH+40, "#9bf0bb", "no unique weight minimum")
	}
	px, py := wx(s.W), ly(currentLoss)
	motion := ""
	if s.Phase == "weight" {
		path := make([][2]float64, 17)
		valid := true
		for i := range path {
			u := s.BeforeW + (s.W-s.BeforeW)*float64(i)/16
			loss := gradientDescentLessonLoss(s, u, s.B)
			if !gradientDescentLessonFinite(loss) {
				valid = false
			}
			path[i] = [2]float64{wx(u), ly(loss)}
		}
		if valid {
			raw, _ := json.Marshal(path)
			motion = ` data-motion="path" data-path="` + html.EscapeString(string(raw)) + `"`
		}
	}
	markerID := "parameter:w"
	if s.Phase == "bias" {
		markerID = "slice:changed-bias"
	}
	fmt.Fprintf(&out, `<g data-entity="%s" data-x="%.3f" data-y="%.3f" transform="translate(%.3f %.3f)"%s><circle r="7" fill="#ee8b9b" stroke="#eff5ed"/><title>Current w on current fixed-b slice</title></g>`, markerID, px, py, px, py, motion)
	// A bias update changes the entire slice: snap this marker rather than inventing a weight-only path.
	text(inset, historyY-12, "#bbcab9", "Measured epoch loss · pre-update parameters")
	stripW := sceneW - 2*inset
	fmt.Fprintf(&out, `<rect x="%.3f" y="%.3f" width="%.3f" height="72" rx="8" fill="#16251f" stroke="#52634e"/>`, inset, historyY, stripW)
	for i, r := range s.History {
		hx := inset + 12 + float64(i)*math.Min(28, (stripW-24)/math.Max(1, float64(len(s.History)-1)))
		hy := gradientDescentLessonMap(r.Loss, s.HistoryLossMin, s.HistoryLossMax, historyY+62, -52)
		fmt.Fprintf(&out, `<circle cx="%.3f" cy="%.3f" r="4" fill="#aa92e7"><title>Epoch %d measured at w=%.6g, b=%.6g: MSE %.6g</title></circle>`, hx, hy, r.Epoch, r.W, r.B, r.Loss)
	}
	y := formulaY
	before := "Last measured MSE: not measured"
	if s.LossBeforeValid {
		before = fmt.Sprintf("Last measured MSE: %.6g at w=%.6g, b=%.6g", s.LossBefore, s.BeforeW, s.BeforeB)
	}
	y = wrap(y, before)
	diagnostic := "Current diagnostic MSE: not evaluated"
	if s.LossCurrentValid {
		diagnostic = fmt.Sprintf("Current diagnostic MSE: %.6g at w=%.6g, b=%.6g", s.LossCurrent, s.W, s.B)
	}
	y = wrap(y, diagnostic)
	y = wrap(y, fmt.Sprintf("SSE=%.6g; sum(e×x)=%.6g; sum(e)=%.6g", s.SumSquaredError, s.SumErrorX, s.SumError))
	y = wrap(y, fmt.Sprintf("g_w=2 sum(e×x)/n=%.6g; g_b=2 sum(e)/n=%.6g", s.GradientW, s.GradientB))
	if s.Sample >= 0 && s.Sample < len(s.X) && s.Sample < len(s.Y) {
		i := s.Sample
		y = wrap(y, fmt.Sprintf("P%d: pred = %.6g × %.6g + %.6g; e = pred − %.6g", i, s.W, s.X[i], s.B, s.Y[i]))
		if step.LineID != "L6" && i < len(s.Predictions) {
			y = wrap(y, fmt.Sprintf("Computed prediction P%d = %.6g", i, s.Predictions[i]))
		}
		if (s.Phase == "residual" || s.Phase == "accumulate") && i < len(s.Residuals) {
			y = wrap(y, fmt.Sprintf("Signed residual P%d = %.6g", i, s.Residuals[i]))
		}
	}
	if s.Phase == "weight" || s.Phase == "bias" || s.Phase == "done" {
		y = wrap(y, fmt.Sprintf("w′=%.6g − %.6g × (%.6g); b′=%.6g − %.6g × (%.6g)", s.BeforeW, s.LR, s.GradientW, s.BeforeB, s.LR, s.GradientB))
	}
	if s.StopReason != "" {
		y = wrap(y, "Stop: "+s.StopReason)
	}
	wrap(y+8, step.Explanation)
	out.WriteString(`</g>`)
	return out.String()
}
