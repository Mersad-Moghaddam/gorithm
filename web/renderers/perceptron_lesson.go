package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type PerceptronLessonState struct {
	ArrayState
	Inputs                                             [][]float64
	Labels                                             []int
	Weights, BeforeWeights, Products                   []float64
	Bias, BeforeBias, Linear, LR                       float64
	Epoch, Sample, Target, Prediction, Error, Mistakes int
	PredictionValid, Converged                         bool
	Phase, StopReason                                  string
	XMin, XMax, YMin, YMax                             float64
}

func perceptronNumber(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "unavailable"
	}
	return fmt.Sprintf("%.5g", v)
}

func RenderPerceptronLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(PerceptronLessonState)
	if !ok {
		return `<g data-lesson="perceptron" data-phase="unavailable" data-scene-width="320" data-scene-height="100"><title>Perceptron state unavailable</title></g>`
	}
	w := width
	if w < 320 {
		w = 320
	}
	if w > 1000 {
		w = 1000
	}
	mobile := w < 640
	pad := 24.0
	if mobile {
		pad = 16
	}
	d := len(s.Weights)
	neuronY := 70.0
	neuronX := float64(w)/2 + 24
	if mobile {
		neuronY = 370
		neuronX = pad
	}
	if d != 2 && mobile && len(s.Inputs) > 8 {
		neuronY += float64((len(s.Inputs)-1)/8) * 28
	}
	neuronH := float64(d)*66 + 180
	correctionY := math.Max(390, neuronY+neuronH+24)
	if d != 2 {
		correctionY = math.Max(correctionY, 380+float64((len(s.Inputs)-1)/8)*28)
	}
	sceneH := int(correctionY + float64(d+1)*110 + 130)
	if d == 2 {
		sceneH += 40 + len(s.Inputs)*66
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="perceptron" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14" fill="#eff5ed"><title>%s</title>`, html.EscapeString(s.Phase), w, sceneH, w, html.EscapeString("Perceptron: "+step.Explanation))
	text := func(x, y float64, value, color string) {
		fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" fill="%s">%s</text>`, x, y, color, html.EscapeString(value))
	}
	text(pad, 30, fmt.Sprintf("Epoch %d · %s · mistakes %d", s.Epoch, s.Phase, s.Mistakes), "#ffe29a")
	planeW := float64(w)/2 - 48
	if mobile {
		planeW = float64(w) - 2*pad
	}
	px, py, pw, ph := pad+24, 100.0, planeW-48, 210.0
	if len(s.Weights) == 2 && s.XMin < s.XMax && s.YMin < s.YMax {
		text(pad, 65, "Decision boundary · x0 / x1", "#bbcab9")
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#192923" stroke="#52634e"/>`, px, py, pw, ph)
		ratio := func(v, lo, hi float64) float64 { return (v/2 - lo/2) / (hi/2 - lo/2) }
		mapX := func(v float64) float64 { return px + ratio(v, s.XMin, s.XMax)*pw }
		mapY := func(v float64) float64 { return py + ph - ratio(v, s.YMin, s.YMax)*ph }
		// Intersect the actual boundary with the fixed data rectangle, not an
		// unbounded slope. Normalizing coefficients avoids product overflow.
		scale := math.Max(math.Abs(s.Bias), math.Max(math.Abs(s.Weights[0]), math.Abs(s.Weights[1])))
		if scale == 0 {
			text(pad, 340, "constant prediction: class 1 (z = 0)", "#8bd5ff")
		} else {
			a, c, bias := s.Weights[0]/scale, s.Weights[1]/scale, s.Bias/scale
			points := [][2]float64{}
			push := func(x, y float64) {
				if !math.IsNaN(x) && !math.IsInf(x, 0) && !math.IsNaN(y) && !math.IsInf(y, 0) && x >= s.XMin && x <= s.XMax && y >= s.YMin && y <= s.YMax {
					points = append(points, [2]float64{mapX(x), mapY(y)})
				}
			}
			if c != 0 {
				push(s.XMin, -(a*s.XMin+bias)/c)
				push(s.XMax, -(a*s.XMax+bias)/c)
			}
			if a != 0 {
				push(-(c*s.YMin+bias)/a, s.YMin)
				push(-(c*s.YMax+bias)/a, s.YMax)
			}
			if len(points) >= 2 {
				fmt.Fprintf(&b, `<path d="M%.2f %.2f L%.2f %.2f" stroke="#8bd5ff" stroke-width="3"/>`, points[0][0], points[0][1], points[1][0], points[1][1])
			}
			if a == 0 && c == 0 {
				text(pad, 340, "constant prediction (bias only)", "#8bd5ff")
			}
		}
		for i, row := range s.Inputs {
			if len(row) != 2 {
				continue
			}
			duplicate := false
			active := i == s.Sample
			mixed := false
			for j, other := range s.Inputs {
				if len(other) == 2 && row[0] == other[0] && row[1] == other[1] {
					duplicate = duplicate || j < i
					active = active || j == s.Sample
					if i < len(s.Labels) && j < len(s.Labels) && s.Labels[i] != s.Labels[j] {
						mixed = true
					}
				}
			}
			if duplicate {
				continue
			}
			x, y := mapX(row[0]), mapY(row[1])
			if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
				continue
			}
			fmt.Fprintf(&b, `<g data-entity="sample:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)">`, i, x, y, x, y)
			label := 0
			if i < len(s.Labels) {
				label = s.Labels[i]
			}
			if mixed {
				b.WriteString(`<path d="M-7 0 L0 -7 L7 0 L0 7 Z" fill="#eff5ed"/>`)
			} else if label == 1 {
				b.WriteString(`<rect x="-6" y="-6" width="12" height="12" fill="#387faa"/>`)
			} else {
				b.WriteString(`<circle r="6" fill="#c45d75"/>`)
			}
			if active {
				b.WriteString(`<circle r="11" fill="none" stroke="#ffe29a" stroke-width="3"/>`)
			}
			b.WriteString(`</g>`)
		}
		text(px, py+ph+20, perceptronNumber(s.XMin)+" … "+perceptronNumber(s.XMax), "#bbcab9")
	} else {
		text(pad, 65, "Feature vector · no 2D classifier plot", "#bbcab9")
		for j := range s.Weights {
			v := 0.0
			if s.Sample >= 0 && s.Sample < len(s.Inputs) && j < len(s.Inputs[s.Sample]) {
				v = s.Inputs[s.Sample][j]
			}
			text(pad, 95+float64(j)*26, fmt.Sprintf("x%d = %s", j, perceptronNumber(v)), "#eff5ed")
		}
		for i := range s.Inputs {
			x := pad + float64(i%8)*32
			y := 340 + float64(i/8)*28
			fmt.Fprintf(&b, `<g data-entity="sample:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)">`, i, x, y, x, y)
			color := "#bbcab9"
			if i == s.Sample {
				color = "#ffe29a"
			}
			text(0, 0, fmt.Sprintf("%d", i), color)
			b.WriteString(`</g>`)
		}
	}
	text(neuronX, neuronY, "Neuron · weighted input links", "#ffe29a")
	for j, wj := range s.BeforeWeights {
		if !s.PredictionValid && s.Phase != "sum" && j < len(s.Weights) {
			wj = s.Weights[j]
		}
		xj := 0.0
		if s.Sample >= 0 && s.Sample < len(s.Inputs) && j < len(s.Inputs[s.Sample]) {
			xj = s.Inputs[s.Sample][j]
		}
		product := 0.0
		if j < len(s.Products) {
			product = s.Products[j]
		}
		y := neuronY + 32 + float64(j)*66
		text(neuronX, y, fmt.Sprintf("x%d = %s", j, perceptronNumber(xj)), "#eff5ed")
		text(neuronX, y+20, fmt.Sprintf("w%d = %s", j, perceptronNumber(wj)), "#eff5ed")
		value := "not evaluated"
		if s.PredictionValid || s.Phase == "sum" {
			value = perceptronNumber(product)
		}
		text(neuronX+16, y+40, "↳ w × x = "+value, "#8bd5ff")
	}
	y := neuronY + float64(d)*66 + 38
	bias := s.BeforeBias
	if !s.PredictionValid && s.Phase != "sum" {
		bias = s.Bias
	}
	text(neuronX, y, "Σ products + bias "+perceptronNumber(bias), "#aa92e7")
	if s.PredictionValid || s.Phase == "sum" {
		text(neuronX, y+24, "z = "+perceptronNumber(s.Linear)+"; z ≥ 0 ⇒ 1, else 0", "#ffe29a")
	} else {
		text(neuronX, y+24, "z: not evaluated · threshold ≥ 0", "#ffe29a")
	}
	prediction := "prediction: not evaluated"
	if s.PredictionValid {
		prediction = fmt.Sprintf("prediction %d · target %d", s.Prediction, s.Target)
	}
	text(neuronX, y+48, prediction, "#eff5ed")
	if s.Phase == "weights" || s.Phase == "bias" {
		text(neuronX, y+72, "prediction before correction", "#8bd5ff")
	}
	if s.PredictionValid && s.Phase != "threshold" {
		text(neuronX, y+96, fmt.Sprintf("error = %d − %d = %d", s.Target, s.Prediction, s.Error), "#eff5ed")
	}
	text(pad, correctionY, "Correction · old + delta = current", "#ffe29a")
	for j, wj := range s.Weights {
		before := wj
		if j < len(s.BeforeWeights) {
			before = s.BeforeWeights[j]
		}
		rowY := correctionY + 34 + float64(j)*110
		text(pad, rowY, fmt.Sprintf("w%d: %s + (%s)", j, perceptronNumber(before), perceptronNumber(wj-before)), "#eff5ed")
		text(pad, rowY+22, "current = "+perceptronNumber(wj), "#eff5ed")
		xj := 0.0
		if s.Sample >= 0 && s.Sample < len(s.Inputs) && j < len(s.Inputs[s.Sample]) {
			xj = s.Inputs[s.Sample][j]
		}
		if s.Phase == "weights" || s.Phase == "bias" {
			text(pad, rowY+44, fmt.Sprintf("Δw%d = %s × %d × %s", j, perceptronNumber(s.LR), s.Error, perceptronNumber(xj)), "#bbcab9")
		}
		x := pad + planeW/2 + math.Tanh(wj)*math.Min(80, planeW/3)
		ey := rowY + 66
		fmt.Fprintf(&b, `<g data-entity="weight:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><circle r="5" fill="#ee8b9b"/></g>`, j, x, ey, x, ey)
	}
	by := correctionY + 34 + float64(d)*110
	text(pad, by, "bias: "+perceptronNumber(s.BeforeBias)+" + ("+perceptronNumber(s.Bias-s.BeforeBias)+")", "#eff5ed")
	text(pad, by+22, "current = "+perceptronNumber(s.Bias), "#eff5ed")
	if s.Phase == "bias" {
		text(pad, by+44, fmt.Sprintf("Δb = %s × %d", perceptronNumber(s.LR), s.Error), "#bbcab9")
	}
	bx := pad + planeW/2 + math.Tanh(s.Bias)*math.Min(80, planeW/3)
	fmt.Fprintf(&b, `<g data-entity="bias:0" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"><path d="M-6 0 L0 -6 L6 0 L0 6 Z" fill="#8bd5ff"/></g>`, bx, by+66, bx, by+66)
	if s.StopReason != "" {
		text(pad, by+80, "Stop: "+s.StopReason, "#9bf0bb")
	}
	if d == 2 {
		listY := by + 120
		text(pad, listY, "Samples · shared points use one marker", "#ffe29a")
		for i, row := range s.Inputs {
			if len(row) != 2 || i >= len(s.Labels) {
				continue
			}
			active := ""
			if i == s.Sample {
				active = " · active"
			}
			y := listY + 30 + float64(i)*66
			text(pad, y, fmt.Sprintf("Sample %d · class %d%s", i, s.Labels[i], active), "#eff5ed")
			text(pad, y+22, "x0 = "+perceptronNumber(row[0]), "#bbcab9")
			text(pad, y+44, "x1 = "+perceptronNumber(row[1]), "#bbcab9")
		}
	}
	b.WriteString(`</g>`)
	return b.String()
}
