package gradientdescent

import (
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func gradientDescentFinite(v ...float64) bool {
	for _, x := range v {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return false
		}
	}
	return true
}

func gradientDescentValidate(x, y []float64, lr float64, epochs int) error {
	if len(x) < 1 || len(x) > 64 || len(x) != len(y) {
		return fmt.Errorf("x and y require 1–64 paired samples")
	}
	if !gradientDescentFinite(lr) || lr <= 0 || lr > 1 {
		return fmt.Errorf("learning rate must satisfy 0 < lr <= 1")
	}
	if epochs < 1 || epochs > 100 {
		return fmt.Errorf("epochs must be in [1,100]")
	}
	for i := range x {
		if !gradientDescentFinite(x[i], y[i]) {
			return fmt.Errorf("sample %d must be finite", i)
		}
	}
	return nil
}

// Trace records real loop operations. History measures the parameters before updates.
func Trace(x, y []float64, lr float64, epochs int) []trace.Step {
	builder := trace.NewBuilder()
	s := renderers.GradientDescentLessonState{Sample: -1, LR: lr, Phase: "init"}
	emit := func(line string, action trace.ActionType, reason string) {
		snap := s
		snap.X = append([]float64(nil), s.X...)
		snap.Y = append([]float64(nil), s.Y...)
		snap.Predictions = append([]float64(nil), s.Predictions...)
		snap.Residuals = append([]float64(nil), s.Residuals...)
		snap.History = append([]renderers.GradientDescentEpoch(nil), s.History...)
		snap.TableState = gradientDescentTable(s.History)
		var loss any = "not measured"
		if s.LossBeforeValid {
			loss = s.LossBefore
		}
		builder.AddStep(line, action, reason, snap, map[string]any{"epoch": s.Epoch, "sample": s.Sample, "w": s.W, "b": s.B, "loss": loss, "lossBeforeW": s.BeforeW, "lossBeforeB": s.BeforeB, "dw": s.SumErrorX, "db": s.SumError, "gradientW": s.GradientW, "gradientB": s.GradientB})
	}
	if err := gradientDescentValidate(x, y, lr, epochs); err != nil {
		s.LR = 0
		s.Phase = "invalid"
		s.StopReason = "invalid-input"
		emit("L1", trace.ActionHighlight, err.Error())
		return gradientDescentDomains(builder.Steps())
	}
	s.X = append([]float64(nil), x...)
	s.Y = append([]float64(nil), y...)
	s.Predictions = make([]float64, len(x))
	s.Residuals = make([]float64, len(x))
	fail := func(line string) []trace.Step {
		s.Phase = "invalid"
		s.StopReason = "numerical-error"
		emit(line, trace.ActionHighlight, "Numerical overflow: training stopped without applying a nonfinite value.")
		return gradientDescentDomains(builder.Steps())
	}
	emit("L2", trace.ActionInit, "Initialize w = 0 and b = 0. MSE has not been measured.")
	n := float64(len(x))
	for epoch := 0; epoch < epochs; epoch++ {
		s.Epoch = epoch
		s.Sample = -1
		s.Phase = "init"
		s.LossCurrentValid = false
		emit("L4", trace.ActionHighlight, fmt.Sprintf("Epoch %d begins with w = %.6g, b = %.6g.", epoch, s.W, s.B))
		s.BeforeW = s.W
		s.BeforeB = s.B
		s.LossBeforeValid = false
		s.SumSquaredError = 0
		s.SumErrorX = 0
		s.SumError = 0
		s.GradientW = 0
		s.GradientB = 0
		s.Predictions = make([]float64, len(x))
		s.Residuals = make([]float64, len(x))
		emit("L5", trace.ActionSet, "Reset SSE, sum(e×x), and sum(e) for this epoch.")
		for i := range x {
			s.Sample = i
			s.Phase = "predict"
			emit("L6", trace.ActionVisit, fmt.Sprintf("Visit P%d: x = %.6g, y = %.6g.", i, x[i], y[i]))
			pred := s.W*x[i] + s.B
			if !gradientDescentFinite(pred) {
				return fail("L7")
			}
			s.Predictions[i] = pred
			emit("L7", trace.ActionSet, fmt.Sprintf("P%d prediction = %.6g × %.6g + %.6g = %.6g.", i, s.W, x[i], s.B, pred))
			e := pred - y[i]
			if !gradientDescentFinite(e) {
				return fail("L8")
			}
			s.Residuals[i] = e
			s.Phase = "residual"
			emit("L8", trace.ActionCompare, fmt.Sprintf("P%d signed residual = %.6g − %.6g = %.6g.", i, pred, y[i], e))
			cost := s.SumSquaredError + e*e
			if !gradientDescentFinite(cost) {
				return fail("L9")
			}
			s.SumSquaredError = cost
			s.Phase = "accumulate"
			emit("L9", trace.ActionSet, fmt.Sprintf("Add residual squared: SSE = %.6g.", cost))
			dw := s.SumErrorX + e*x[i]
			if !gradientDescentFinite(dw) {
				return fail("L10")
			}
			s.SumErrorX = dw
			emit("L10", trace.ActionSet, fmt.Sprintf("Add residual × x: sum(e×x) = %.6g.", dw))
			db := s.SumError + e
			if !gradientDescentFinite(db) {
				return fail("L11")
			}
			s.SumError = db
			emit("L11", trace.ActionSet, fmt.Sprintf("Add signed residual: sum(e) = %.6g.", db))
		}
		gw, gb := (2/n)*s.SumErrorX, (2/n)*s.SumError
		if !gradientDescentFinite(gw, gb) {
			return fail("L13")
		}
		s.Sample = -1
		s.GradientW = gw
		s.GradientB = gb
		s.LossBefore = s.SumSquaredError / n
		s.LossBeforeValid = true
		s.Phase = "loss"
		s.History = append(s.History, renderers.GradientDescentEpoch{Epoch: epoch, W: s.BeforeW, B: s.BeforeB, Loss: s.LossBefore})
		emit("L13", trace.ActionCompare, fmt.Sprintf("Measured MSE = %.6g / %d = %.6g at w = %.6g, b = %.6g. Gradients: %.6g, %.6g.", s.SumSquaredError, len(x), s.LossBefore, s.BeforeW, s.BeforeB, gw, gb))
		w := s.W - lr*(2/n)*s.SumErrorX
		if !gradientDescentFinite(w) {
			return fail("L14")
		}
		s.W = w
		s.Phase = "weight"
		builder.IncOperations()
		emit("L14", trace.ActionSet, fmt.Sprintf("Weight update: %.6g − %.6g × (%.6g) = %.6g. Bias stays %.6g; motion illustrates this discrete update.", s.BeforeW, lr, gw, w, s.B))
		b := s.B - lr*(2/n)*s.SumError
		if !gradientDescentFinite(b) {
			return fail("L15")
		}
		s.B = b
		s.Phase = "bias"
		builder.IncOperations()
		emit("L15", trace.ActionSet, fmt.Sprintf("Bias update: %.6g − %.6g × (%.6g) = %.6g using the same old gradients. Loss slice changes.", s.BeforeB, lr, gb, b))
	}
	// Diagnostic only: does not append history or perform another optimizer iteration.
	var cost float64
	for i := range x {
		p := s.W*x[i] + s.B
		e := p - y[i]
		next := cost + e*e
		if !gradientDescentFinite(p, e, next) {
			return fail("L17")
		}
		cost = next
	}
	s.LossCurrent = cost / n
	s.LossCurrentValid = true
	s.Phase = "done"
	s.StopReason = "epoch-limit"
	change := "decreased"
	if s.LossCurrent > s.LossBefore {
		change = "increased"
	}
	if s.LossCurrent == s.LossBefore {
		change = "unchanged"
	}
	emit("L17", trace.ActionDone, fmt.Sprintf("Epoch budget complete, not a convergence test. Final w = %.6g, b = %.6g; last pre-update MSE = %.6g. Current diagnostic MSE = %.6g (%s).", s.W, s.B, s.LossBefore, s.LossCurrent, change))
	return gradientDescentDomains(builder.Steps())
}

func gradientDescentTable(history []renderers.GradientDescentEpoch) renderers.TableState {
	t := renderers.TableState{ColHeaders: []string{"Epoch", "Measured w", "Measured b", "Pre-update MSE"}}
	if len(history) == 0 {
		t.RowHeaders = []string{"unmeasured"}
		t.Cells = []renderers.TableCell{{Value: "not measured"}}
		return t
	}
	start := max(0, len(history)-10)
	for _, r := range history[start:] {
		row := len(t.RowHeaders)
		t.RowHeaders = append(t.RowHeaders, fmt.Sprint(r.Epoch))
		for c, v := range []float64{float64(r.Epoch), r.W, r.B, r.Loss} {
			t.Cells = append(t.Cells, renderers.TableCell{Row: row, Col: c, Value: fmt.Sprintf("%.6g", v)})
		}
	}
	return t
}

func gradientDescentPad(lo, hi float64) (float64, float64) {
	if lo == hi {
		d := math.Max(1, math.Abs(lo)*.1)
		return lo - d, hi + d
	}
	d := (hi - lo) * .1
	if !gradientDescentFinite(d, lo-d, hi+d) {
		return lo, hi
	}
	return lo - d, hi + d
}

func gradientDescentDomains(steps []trace.Step) []trace.Step {
	xmin, xmax, ymin, ymax, wmin, wmax, lmax, hmax := 0., 0., 0., 0., 0., 0., 0., 0.
	include := func(lo, hi *float64, v float64) {
		if gradientDescentFinite(v) {
			*lo = math.Min(*lo, v)
			*hi = math.Max(*hi, v)
		}
	}
	for _, st := range steps {
		s := st.State.(renderers.GradientDescentLessonState)
		for i, x := range s.X {
			include(&xmin, &xmax, x)
			include(&ymin, &ymax, s.Y[i])
			include(&ymin, &ymax, s.W*x+s.B)
		}
		include(&wmin, &wmax, s.W)
		include(&wmin, &wmax, s.BeforeW)
		for _, r := range s.History {
			hmax = math.Max(hmax, r.Loss)
		}
		var xx, xy float64
		for i, x := range s.X {
			xx += x * x
			xy += x * (s.Y[i] - s.B)
		}
		if xx > 0 && gradientDescentFinite(xx, xy, xy/xx) {
			include(&wmin, &wmax, xy/xx)
		}
	}
	xmin, xmax = gradientDescentPad(xmin, xmax)
	ymin, ymax = gradientDescentPad(ymin, ymax)
	wmin, wmax = gradientDescentPad(wmin, wmax)
	for _, st := range steps {
		s := st.State.(renderers.GradientDescentLessonState)
		for _, w := range []float64{wmin, wmax, s.W} {
			var loss float64
			for i, x := range s.X {
				e := w*x + s.B - s.Y[i]
				loss += e * e / float64(len(s.X))
			}
			if gradientDescentFinite(loss) {
				lmax = math.Max(lmax, loss)
			}
		}
	}
	_, lmax = gradientDescentPad(0, lmax)
	_, hmax = gradientDescentPad(0, hmax)
	for i := range steps {
		s := steps[i].State.(renderers.GradientDescentLessonState)
		s.DataXMin = xmin
		s.DataXMax = xmax
		s.DataYMin = ymin
		s.DataYMax = ymax
		s.SliceWMin = wmin
		s.SliceWMax = wmax
		s.SliceLossMin = 0
		s.SliceLossMax = lmax
		s.HistoryLossMin = 0
		s.HistoryLossMax = hmax
		steps[i].State = s
	}
	return steps
}
