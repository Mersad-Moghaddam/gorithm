package perceptron

import (
	"fmt"
	"html"
	"math"
	"strings"
)

// Point represents a training sample with feature vector X and class Target.
type Point struct {
	X      []float64 `json:"x"`
	Target int       `json:"target"`
}

// PerceptronState is the complete snapshot of perceptron state at each trace step.
type PerceptronState struct {
	Weights      []float64 `json:"weights"`
	Bias         float64   `json:"bias"`
	LR           float64   `json:"lr"`
	Epoch        int       `json:"epoch"`
	MaxEpochs    int       `json:"maxEpochs"`
	ActiveSample int       `json:"activeSample"` // -1 when none
	ActiveTarget int       `json:"activeTarget"`
	ActivePred   int       `json:"activePred"`
	LinearZ      float64   `json:"linearZ"`
	HasZ         bool      `json:"hasZ"`
	HasPred      bool      `json:"hasPred"`
	Error        int       `json:"error"`
	Points       []Point   `json:"points"`
	Converged    bool      `json:"converged"`
	Caption      string    `json:"caption"`
}

// Clone returns a deep copy of PerceptronState.
func (s PerceptronState) Clone() PerceptronState {
	cp := s
	cp.Weights = make([]float64, len(s.Weights))
	copy(cp.Weights, s.Weights)
	if s.Points != nil {
		cp.Points = make([]Point, len(s.Points))
		for i, pt := range s.Points {
			xCp := make([]float64, len(pt.X))
			copy(xCp, pt.X)
			cp.Points[i] = Point{
				X:      xCp,
				Target: pt.Target,
			}
		}
	}
	return cp
}

// RenderSVG produces SVG children visualizing the Perceptron model, decision boundary, and state.
func (s PerceptronState) RenderSVG(width, height int) string {
	if width <= 0 {
		width = 800
	}
	if height <= 0 {
		height = 450
	}

	var sb strings.Builder

	// Top header: caption, epoch counter, and learning rate badge
	caption := s.Caption
	if caption == "" {
		caption = "Perceptron Neural Unit"
	}
	fmt.Fprintf(&sb, `<text data-key="caption" x="%d" y="26" fill="#f8fafc" font-size="14" font-weight="600" text-anchor="middle">%s</text>`, width/2, html.EscapeString(caption))

	if s.MaxEpochs > 0 {
		fmt.Fprintf(&sb, `<text data-key="epoch-badge" x="35" y="26" fill="#94a3b8" font-size="12">Epoch: %d / %d</text>`, s.Epoch, s.MaxEpochs)
	}
	if s.LR > 0 {
		fmt.Fprintf(&sb, `<text data-key="lr-badge" x="%d" y="26" fill="#94a3b8" font-size="12" text-anchor="end">Learning Rate: %.2f</text>`, width-35, s.LR)
	}

	// ponytail: 2D projection plots first 2 features; add PCA or projection when d > 2 datasets requested.
	// Left: 2D Plot Container
	plotBoxX := 35.0
	plotBoxY := 42.0
	plotBoxW := 385.0
	plotBoxH := 392.0
	fmt.Fprintf(&sb, `<rect data-key="plot-bg" x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="8" fill="#0f172a" stroke="#334155" stroke-width="1.5"/>`, plotBoxX, plotBoxY, plotBoxW, plotBoxH)
	fmt.Fprintf(&sb, `<text data-key="plot-title" x="50" y="64" fill="#94a3b8" font-size="12" font-weight="700">2D DECISION BOUNDARY</text>`)

	// Coordinate space mapping
	minX, maxX := 0.0, 1.0
	minY, maxY := 0.0, 1.0
	if len(s.Points) > 0 {
		minX, maxX = s.Points[0].X[0], s.Points[0].X[0]
		if len(s.Points[0].X) > 1 {
			minY, maxY = s.Points[0].X[1], s.Points[0].X[1]
		}
		for _, pt := range s.Points {
			if len(pt.X) > 0 {
				if pt.X[0] < minX {
					minX = pt.X[0]
				}
				if pt.X[0] > maxX {
					maxX = pt.X[0]
				}
			}
			if len(pt.X) > 1 {
				if pt.X[1] < minY {
					minY = pt.X[1]
				}
				if pt.X[1] > maxY {
					maxY = pt.X[1]
				}
			}
		}
	}
	spanX := maxX - minX
	if spanX < 1.0 {
		spanX = 1.0
	}
	spanY := maxY - minY
	if spanY < 1.0 {
		spanY = 1.0
	}
	padX := spanX * 0.35
	padY := spanY * 0.35
	plotMinX := minX - padX
	plotMaxX := maxX + padX
	plotMinY := minY - padY
	plotMaxY := maxY + padY

	plotL := 65.0
	plotR := 395.0
	plotW := plotR - plotL
	plotT := 82.0
	plotB := 382.0
	plotH := plotB - plotT

	mapX := func(x float64) float64 {
		return plotL + (x-plotMinX)/(plotMaxX-plotMinX)*plotW
	}
	mapY := func(y float64) float64 {
		return plotB - (y-plotMinY)/(plotMaxY-plotMinY)*plotH
	}

	// Clip path for inner plot drawing
	fmt.Fprintf(&sb, `<clipPath id="perceptron-plot-clip"><rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="6"/></clipPath>`, plotBoxX+2, plotBoxY+2, plotBoxW-4, plotBoxH-4)
	fmt.Fprintf(&sb, `<g clip-path="url(#perceptron-plot-clip)">`)

	// Axes
	if plotMinX <= 0 && plotMaxX >= 0 {
		axX := mapX(0)
		fmt.Fprintf(&sb, `<line data-key="axis-y" x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#334155" stroke-width="1.5" stroke-dasharray="3,3"/>`, axX, plotT, axX, plotB)
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" fill="#64748b" font-size="10" text-anchor="middle">x₁=0</text>`, axX, plotT-4)
	}
	if plotMinY <= 0 && plotMaxY >= 0 {
		axY := mapY(0)
		fmt.Fprintf(&sb, `<line data-key="axis-x" x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#334155" stroke-width="1.5" stroke-dasharray="3,3"/>`, plotL, axY, plotR, axY)
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" fill="#64748b" font-size="10" text-anchor="start">x₂=0</text>`, plotR+4, axY+3)
	}

	// Decision boundary line
	w1 := 0.0
	w2 := 0.0
	if len(s.Weights) > 0 {
		w1 = s.Weights[0]
	}
	if len(s.Weights) > 1 {
		w2 = s.Weights[1]
	}
	bias := s.Bias

	if math.Abs(w1) < 1e-9 && math.Abs(w2) < 1e-9 {
		fmt.Fprintf(&sb, `<text data-key="boundary-status" x="227" y="415" fill="#64748b" font-size="11" text-anchor="middle">Decision Boundary: w = [0, 0] (uninitialized)</text>`)
	} else {
		type pt2D struct{ x, y float64 }
		var inters []pt2D

		if math.Abs(w2) > 1e-9 {
			y1 := -(w1*plotMinX + bias) / w2
			if y1 >= plotMinY-1e-5 && y1 <= plotMaxY+1e-5 {
				inters = append(inters, pt2D{plotMinX, y1})
			}
			y2 := -(w1*plotMaxX + bias) / w2
			if y2 >= plotMinY-1e-5 && y2 <= plotMaxY+1e-5 {
				inters = append(inters, pt2D{plotMaxX, y2})
			}
		}
		if math.Abs(w1) > 1e-9 {
			x1 := -(w2*plotMinY + bias) / w1
			if x1 >= plotMinX-1e-5 && x1 <= plotMaxX+1e-5 {
				inters = append(inters, pt2D{x1, plotMinY})
			}
			x2 := -(w2*plotMaxY + bias) / w1
			if x2 >= plotMinX-1e-5 && x2 <= plotMaxX+1e-5 {
				inters = append(inters, pt2D{x2, plotMaxY})
			}
		}

		var unique []pt2D
		for _, p := range inters {
			dup := false
			for _, u := range unique {
				if math.Hypot(p.x-u.x, p.y-u.y) < 1e-4 {
					dup = true
					break
				}
			}
			if !dup {
				unique = append(unique, p)
			}
		}

		if len(unique) >= 2 {
			pA, pB := unique[0], unique[1]
			sx1, sy1 := mapX(pA.x), mapY(pA.y)
			sx2, sy2 := mapX(pB.x), mapY(pB.y)
			fmt.Fprintf(&sb, `<line data-key="decision-boundary" x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#f59e0b" stroke-width="3" stroke-linecap="round"/>`, sx1, sy1, sx2, sy2)
		}
		fmt.Fprintf(&sb, `<text data-key="boundary-equation" x="227" y="415" fill="#f59e0b" font-size="11" font-weight="600" text-anchor="middle">w·x + b = %.2f·x₁ %+.2f·x₂ %+.2f = 0</text>`, w1, w2, bias)
	}

	// Data points
	for i, pt := range s.Points {
		if len(pt.X) < 2 {
			continue
		}
		px := mapX(pt.X[0])
		py := mapY(pt.X[1])
		pointKey := fmt.Sprintf("sample-point-%d", i)

		fillColor := "#ef4444"
		targetLabel := "0"
		if pt.Target == 1 {
			fillColor = "#22c55e"
			targetLabel = "1"
		}

		fmt.Fprintf(&sb, `<g data-key="%s"><title>Sample %d: x=(%.1f, %.1f), target=%s</title>`, pointKey, i, pt.X[0], pt.X[1], targetLabel)
		fmt.Fprintf(&sb, `<circle cx="%.1f" cy="%.1f" r="11" fill="%s" stroke="#ffffff" stroke-width="1.5"/>`, px, py, fillColor)
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" fill="#ffffff" font-size="10" font-weight="bold" text-anchor="middle" dominant-baseline="central">%s</text>`, px, py, targetLabel)
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" fill="#94a3b8" font-size="9" text-anchor="middle">(%.1f, %.1f)</text>`, px, py+20, pt.X[0], pt.X[1])
		fmt.Fprintf(&sb, `</g>`)
	}

	// Active sample highlight ring
	if s.ActiveSample >= 0 && s.ActiveSample < len(s.Points) && len(s.Points[s.ActiveSample].X) >= 2 {
		apt := s.Points[s.ActiveSample]
		apx := mapX(apt.X[0])
		apy := mapY(apt.X[1])
		fmt.Fprintf(&sb, `<circle data-key="active-ring" cx="%.1f" cy="%.1f" r="18" fill="none" stroke="#fde047" stroke-width="2.5" stroke-dasharray="4,2"/>`, apx, apy)
	}

	fmt.Fprintf(&sb, `</g>`) // End clip group

	// Right: Math & Neural Unit Panel
	rightX := 435.0
	rightY := 42.0
	rightW := 330.0
	rightH := 392.0
	fmt.Fprintf(&sb, `<rect data-key="math-bg" x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="8" fill="#1e293b" stroke="#334155" stroke-width="1.5"/>`, rightX, rightY, rightW, rightH)
	fmt.Fprintf(&sb, `<text data-key="math-title" x="450" y="64" fill="#94a3b8" font-size="12" font-weight="700">NEURAL UNIT (PERCEPTRON)</text>`)

	// Card 1: Parameters (w1, w2, bias)
	cardW := 304.0
	fmt.Fprintf(&sb, `<rect data-key="weights-card" x="448" y="74" width="%.1f" height="66" rx="6" fill="#0f172a" stroke="#475569" stroke-width="1"/>`, cardW)
	fmt.Fprintf(&sb, `<text x="458" y="90" fill="#94a3b8" font-size="10" font-weight="600">WEIGHTS &amp; BIAS</text>`)
	fmt.Fprintf(&sb, `<rect x="458" y="98" width="90" height="32" rx="4" fill="#1e293b"/>`)
	fmt.Fprintf(&sb, `<text data-key="w1-val" x="503" y="119" fill="#38bdf8" font-size="12" font-weight="bold" text-anchor="middle">w₁: %+.2f</text>`, w1)
	fmt.Fprintf(&sb, `<rect x="555" y="98" width="90" height="32" rx="4" fill="#1e293b"/>`)
	fmt.Fprintf(&sb, `<text data-key="w2-val" x="600" y="119" fill="#38bdf8" font-size="12" font-weight="bold" text-anchor="middle">w₂: %+.2f</text>`, w2)
	fmt.Fprintf(&sb, `<rect x="652" y="98" width="90" height="32" rx="4" fill="#1e293b"/>`)
	fmt.Fprintf(&sb, `<text data-key="bias-val" x="697" y="119" fill="#a78bfa" font-size="12" font-weight="bold" text-anchor="middle">bias: %+.2f</text>`, bias)

	// Card 2: Current Input Sample
	fmt.Fprintf(&sb, `<rect data-key="sample-card" x="448" y="148" width="%.1f" height="64" rx="6" fill="#0f172a" stroke="#475569" stroke-width="1"/>`, cardW)
	fmt.Fprintf(&sb, `<text x="458" y="164" fill="#94a3b8" font-size="10" font-weight="600">INPUT SAMPLE</text>`)
	if s.ActiveSample >= 0 && s.ActiveSample < len(s.Points) {
		pt := s.Points[s.ActiveSample]
		fmt.Fprintf(&sb, `<text data-key="sample-val" x="458" y="190" fill="#f8fafc" font-size="12">Sample %d: x = %v</text>`, s.ActiveSample, pt.X)
		fmt.Fprintf(&sb, `<text data-key="target-val" x="742" y="190" fill="#fde047" font-size="12" font-weight="bold" text-anchor="end">Target y = %d</text>`, s.ActiveTarget)
	} else {
		fmt.Fprintf(&sb, `<text data-key="sample-val" x="458" y="190" fill="#64748b" font-size="12">—</text>`)
	}

	// Card 3: Linear Combination & Activation
	fmt.Fprintf(&sb, `<rect data-key="activation-card" x="448" y="220" width="%.1f" height="88" rx="6" fill="#0f172a" stroke="#475569" stroke-width="1"/>`, cardW)
	fmt.Fprintf(&sb, `<text x="458" y="236" fill="#94a3b8" font-size="10" font-weight="600">LINEAR COMBINATION &amp; ACTIVATION</text>`)
	if s.HasZ {
		fmt.Fprintf(&sb, `<text data-key="z-val" x="458" y="258" fill="#e2e8f0" font-size="11">z = bias + w·x = <tspan fill="#38bdf8" font-weight="bold">%+.2f</tspan></text>`, s.LinearZ)
	} else {
		fmt.Fprintf(&sb, `<text data-key="z-val" x="458" y="258" fill="#64748b" font-size="11">z = bias + w·x</text>`)
	}

	if s.HasPred {
		fmt.Fprintf(&sb, `<text data-key="pred-val" x="458" y="284" fill="#e2e8f0" font-size="11">pred = step(z) = <tspan font-weight="bold">%d</tspan></text>`, s.ActivePred)
		if s.ActivePred == s.ActiveTarget {
			fmt.Fprintf(&sb, `<rect x="618" y="268" width="124" height="24" rx="4" fill="#14532d"/>`)
			fmt.Fprintf(&sb, `<text data-key="eval-badge" x="680" y="284" fill="#4ade80" font-size="11" font-weight="bold" text-anchor="middle">MATCH (Correct)</text>`)
		} else {
			fmt.Fprintf(&sb, `<rect x="588" y="268" width="154" height="24" rx="4" fill="#7f1d1d"/>`)
			fmt.Fprintf(&sb, `<text data-key="eval-badge" x="665" y="284" fill="#f87171" font-size="11" font-weight="bold" text-anchor="middle">MISCLASSIFIED (err=%d)</text>`, s.Error)
		}
	} else {
		fmt.Fprintf(&sb, `<text data-key="pred-val" x="458" y="284" fill="#64748b" font-size="11">pred = step(z) = (z ≥ 0 ? 1 : 0)</text>`)
	}

	// Card 4: Weight Update Rule
	fmt.Fprintf(&sb, `<rect data-key="update-card" x="448" y="316" width="%.1f" height="106" rx="6" fill="#0f172a" stroke="#475569" stroke-width="1"/>`, cardW)
	if s.HasPred && s.Error != 0 {
		fmt.Fprintf(&sb, `<text data-key="update-title" x="458" y="334" fill="#f87171" font-size="11" font-weight="bold">Misclassified! Updating Weights</text>`)
		fmt.Fprintf(&sb, `<text data-key="update-formula" x="458" y="352" fill="#94a3b8" font-size="10">Rule: Δw = lr * err * x,  Δbias = lr * err</text>`)
		if s.ActiveSample >= 0 && s.ActiveSample < len(s.Points) {
			pt := s.Points[s.ActiveSample]
			x1Val := 0.0
			x2Val := 0.0
			if len(pt.X) > 0 {
				x1Val = pt.X[0]
			}
			if len(pt.X) > 1 {
				x2Val = pt.X[1]
			}
			dw1 := s.LR * float64(s.Error) * x1Val
			dw2 := s.LR * float64(s.Error) * x2Val
			db := s.LR * float64(s.Error)
			fmt.Fprintf(&sb, `<text data-key="update-vals" x="458" y="374" fill="#fde047" font-size="11">Δw₁ = %+.2f,  Δw₂ = %+.2f</text>`, dw1, dw2)
			fmt.Fprintf(&sb, `<text data-key="update-bias" x="458" y="396" fill="#a78bfa" font-size="11">Δbias = %+.2f</text>`, db)
		}
	} else if s.HasPred && s.Error == 0 {
		fmt.Fprintf(&sb, `<text data-key="update-title" x="458" y="340" fill="#4ade80" font-size="11" font-weight="bold">Correct Prediction! No Update</text>`)
		fmt.Fprintf(&sb, `<text data-key="update-formula" x="458" y="364" fill="#94a3b8" font-size="11">Weights and bias remain unchanged.</text>`)
		fmt.Fprintf(&sb, `<text data-key="update-vals" x="458" y="388" fill="#64748b" font-size="11">err = 0 → Δw = 0, Δbias = 0</text>`)
	} else if s.Converged {
		fmt.Fprintf(&sb, `<text data-key="update-title" x="458" y="345" fill="#38bdf8" font-size="12" font-weight="bold">Perceptron Converged!</text>`)
		fmt.Fprintf(&sb, `<text data-key="update-formula" x="458" y="375" fill="#cbd5e1" font-size="11">All training samples correctly classified.</text>`)
	} else {
		fmt.Fprintf(&sb, `<text data-key="update-title" x="458" y="340" fill="#94a3b8" font-size="11" font-weight="bold">Weight Update Rule</text>`)
		fmt.Fprintf(&sb, `<text data-key="update-formula" x="458" y="364" fill="#64748b" font-size="11">Rule: w += lr * err * x, bias += lr * err</text>`)
		fmt.Fprintf(&sb, `<text data-key="update-vals" x="458" y="388" fill="#64748b" font-size="11">Awaiting forward classification pass...</text>`)
	}

	return sb.String()
}
