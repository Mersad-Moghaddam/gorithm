package perceptron

import (
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func validateLessonInput(inputs [][]float64, labels []int, lr float64, epochs int) error {
	if len(inputs) < 1 || len(inputs) > 64 || len(labels) != len(inputs) {
		return fmt.Errorf("use 1–64 samples with matching labels")
	}
	d := len(inputs[0])
	if d < 1 || d > 8 {
		return fmt.Errorf("use 1–8 features")
	}
	if !perceptronFinite(lr) || lr <= 0 || lr > 1 || epochs < 1 || epochs > 100 {
		return fmt.Errorf("require 0 < lr <= 1 and 1–100 epochs")
	}
	for i, row := range inputs {
		if len(row) != d {
			return fmt.Errorf("sample %d has inconsistent feature count", i)
		}
		if labels[i] != 0 && labels[i] != 1 {
			return fmt.Errorf("sample %d label must be 0 or 1", i)
		}
		for _, v := range row {
			if !perceptronFinite(v) {
				return fmt.Errorf("features must be finite")
			}
		}
	}
	return nil
}

func perceptronFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Trace records the real threshold and separate weight/bias corrections.
func Trace(inputs [][]float64, labels []int, lr float64, epochs int) []trace.Step {
	b := trace.NewBuilder()
	s := renderers.PerceptronLessonState{Sample: -1, LR: lr, Phase: "init"}
	add := func(line string, action trace.ActionType, explanation string) {
		cp := s
		cp.Inputs = make([][]float64, len(s.Inputs))
		for i, row := range s.Inputs {
			cp.Inputs[i] = append([]float64(nil), row...)
		}
		cp.Labels = append([]int(nil), s.Labels...)
		cp.Weights = append([]float64(nil), s.Weights...)
		cp.BeforeWeights = append([]float64(nil), s.BeforeWeights...)
		cp.Products = append([]float64(nil), s.Products...)
		cp.ArrayState = renderers.ArrayState{Values: make([]int, len(s.Weights)+1), Current: -1}
		vars := map[string]any{"weights": append([]float64(nil), s.Weights...), "bias": s.Bias, "epoch": s.Epoch, "sample": s.Sample, "target": s.Target, "lr": s.LR, "mistakes": s.Mistakes, "stopReason": s.StopReason}
		if s.PredictionValid {
			vars["pred"] = s.Prediction
			vars["err"] = s.Error
		}
		b.AddStep(line, action, explanation, cp, vars)
		if line == "L8" {
			b.Steps()[len(b.Steps())-1].HoldMS = 600
		}
		if line == "L9" {
			b.Steps()[len(b.Steps())-1].HoldMS = 1200
		}
	}
	if err := validateLessonInput(inputs, labels, lr, epochs); err != nil {
		s.LR = 0
		s.Phase = "invalid-input"
		s.StopReason = "invalid-input"
		add("L1", trace.ActionDone, err.Error())
		return b.Steps()
	}
	s.Inputs = inputs
	s.Labels = labels
	s.Weights = make([]float64, len(inputs[0]))
	s.BeforeWeights = make([]float64, len(s.Weights))
	s.Products = make([]float64, len(s.Weights))
	// Fixed domains frame the dataset, never hidden optimizer history. Half-scaled
	// arithmetic keeps extreme but finite data from overflowing the padding.
	domain := func(feature int) (float64, float64) {
		lo, hi := 0.0, 0.0
		for _, row := range inputs {
			v := row[feature]
			lo = math.Min(lo, v)
			hi = math.Max(hi, v)
		}
		if lo == hi {
			return lo - 1, hi + 1
		}
		pad := (hi/2 - lo/2) * .2
		if perceptronFinite(lo - pad) {
			lo -= pad
		}
		if perceptronFinite(hi + pad) {
			hi += pad
		}
		return lo, hi
	}
	s.XMin, s.XMax = domain(0)
	s.YMin, s.YMax = -1, 1
	if len(s.Weights) == 2 {
		s.YMin, s.YMax = domain(1)
	}
	add("L1", trace.ActionInit, fmt.Sprintf("Train %d samples with %d features; threshold z ≥ 0 predicts class 1.", len(inputs), len(s.Weights)))
	add("L2", trace.ActionSet, "Initialize every weight to zero.")
	add("L3", trace.ActionSet, "Initialize bias to zero.")
	numerical := func(line string) []trace.Step {
		s.Phase = "numerical-error"
		s.StopReason = "numerical-error"
		s.Converged = false
		add(line, trace.ActionDone, "Arithmetic overflow; training stopped before recording nonfinite parameters.")
		return b.Steps()
	}
	for epoch := 1; epoch <= epochs; epoch++ {
		s.Epoch = epoch
		s.Sample = -1
		s.PredictionValid = false
		s.Phase = "epoch"
		add("L4", trace.ActionHighlight, fmt.Sprintf("Start epoch %d of %d.", epoch, epochs))
		s.Mistakes = 0
		add("L5", trace.ActionSet, "Assume no mistakes; only a full error-free epoch proves stability.")
		for i, x := range inputs {
			s.Sample = i
			s.Target = labels[i]
			s.PredictionValid = false
			s.Linear = 0
			s.Error = 0
			s.Phase = "sample"
			copy(s.BeforeWeights, s.Weights)
			s.BeforeBias = s.Bias
			for j := range s.Products {
				s.Products[j] = 0
			}
			add("L6", trace.ActionVisit, fmt.Sprintf("Select sample %d; target class %d.", i, s.Target))
			linear := s.Bias
			products := make([]float64, len(x))
			for j, v := range x {
				products[j] = s.Weights[j] * v
				if !perceptronFinite(products[j]) || !perceptronFinite(linear+products[j]) {
					return numerical("L7")
				}
				linear += products[j]
			}
			s.Products = products
			s.Linear = linear
			s.Phase = "sum"
			b.IncOperations()
			add("L7", trace.ActionHighlight, fmt.Sprintf("z = bias + Σ(wj × xj) = %.6g. Prediction not evaluated yet.", linear))
			s.Prediction = 0
			if linear >= 0 {
				s.Prediction = 1
			}
			s.PredictionValid = true
			s.Phase = "threshold"
			b.IncComparisons()
			add("L8", trace.ActionCompare, fmt.Sprintf("z = %.6g; z ≥ 0 predicts %d.", linear, s.Prediction))
			s.Error = s.Target - s.Prediction
			s.Phase = "error"
			b.IncComparisons()
			add("L9", trace.ActionCompare, fmt.Sprintf("err = target %d − prediction %d = %d.", s.Target, s.Prediction, s.Error))
			s.Phase = "branch"
			if s.Error == 0 {
				add("L10", trace.ActionVisit, "Correct prediction; weights and bias stay unchanged.")
				continue
			}
			add("L10", trace.ActionCompare, "Incorrect prediction; apply signed correction.")
			s.Mistakes++
			s.Phase = "mistake"
			add("L11", trace.ActionSet, "This epoch has a mistake; stability is not proven.")
			next := append([]float64(nil), s.Weights...)
			for j := range next {
				delta := lr * float64(s.Error) * x[j]
				next[j] += delta
				if !perceptronFinite(delta) || !perceptronFinite(next[j]) {
					return numerical("L12")
				}
			}
			s.Weights = next
			s.Phase = "weights"
			b.IncSwaps()
			add("L12", trace.ActionSwap, "Weights corrected: w′j = wj + lr × err × xj. Bias still unchanged; prediction is before correction.")
			nextBias := s.Bias + lr*float64(s.Error)
			if !perceptronFinite(nextBias) {
				return numerical("L13")
			}
			s.Bias = nextBias
			s.Phase = "bias"
			b.IncSwaps()
			add("L13", trace.ActionSet, fmt.Sprintf("Bias corrected: %.6g + %.6g × %d = %.6g. Prediction is before correction.", s.BeforeBias, lr, s.Error, s.Bias))
		}
		s.Sample = -1
		s.PredictionValid = false
		s.Phase = "epoch-check"
		s.Converged = s.Mistakes == 0
		if s.Converged {
			s.StopReason = "stable"
			add("L16", trace.ActionHighlight, "Full epoch without mistakes; training-set predictions stable.")
			break
		}
		add("L16", trace.ActionCompare, fmt.Sprintf("Epoch had %d mistakes; continue if budget remains.", s.Mistakes))
	}
	if !s.Converged {
		s.StopReason = "epoch-limit"
	}
	s.Phase = "done"
	add("L18", trace.ActionDone, fmt.Sprintf("Stopped after %d executed epochs: %s.", s.Epoch, s.StopReason))
	return b.Steps()
}
