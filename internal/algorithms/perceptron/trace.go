package perceptron

import (
	"fmt"
	"slices"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// Trace executes perceptron training and emits visualization steps.
func Trace(inputs [][]float64, labels []int, lr float64, epochs int) []trace.Step {
	b := trace.NewBuilder()

	if len(inputs) == 0 || len(labels) == 0 {
		emptyState := PerceptronState{
			Weights:      []float64{},
			Bias:         0.0,
			LR:           lr,
			Epoch:        0,
			MaxEpochs:    epochs,
			ActiveSample: -1,
			Caption:      "Empty dataset provided; perceptron cannot train.",
		}
		b.AddStep("L1", trace.ActionInit, emptyState.Caption, emptyState.Clone(), varsMap([]float64{}, 0.0, 0, -1, 0, 0, lr, false))
		emptyDone := emptyState.Clone()
		emptyDone.Caption = "Training completed with empty inputs."
		b.AddStep("L18", trace.ActionDone, emptyDone.Caption, emptyDone.Clone(), varsMap([]float64{}, 0.0, 0, -1, 0, 0, lr, true))
		return b.Steps()
	}

	d := len(inputs[0])
	weights := make([]float64, d)
	bias := 0.0

	points := make([]Point, len(inputs))
	for i, x := range inputs {
		points[i] = Point{
			X:      slices.Clone(x),
			Target: labels[i],
		}
	}

	state := PerceptronState{
		Weights:      slices.Clone(weights),
		Bias:         bias,
		LR:           lr,
		Epoch:        0,
		MaxEpochs:    epochs,
		ActiveSample: -1,
		Points:       points,
	}

	initCaption := fmt.Sprintf("Initialize Perceptron: %d samples, %d features, lr=%.2f, max epochs=%d.", len(inputs), d, lr, epochs)
	state.Caption = initCaption
	b.AddStep("L1", trace.ActionInit, initCaption, state.Clone(), varsMap(weights, bias, 0, -1, 0, 0, lr, false))

	wInitCaption := fmt.Sprintf("Initialized weight vector w = %v of dimension %d with zeros.", weights, d)
	state.Caption = wInitCaption
	b.AddStep("L2", trace.ActionSet, wInitCaption, state.Clone(), varsMap(weights, bias, 0, -1, 0, 0, lr, false))

	bInitCaption := "Initialized bias scalar b = 0.00."
	state.Caption = bInitCaption
	b.AddStep("L3", trace.ActionSet, bInitCaption, state.Clone(), varsMap(weights, bias, 0, -1, 0, 0, lr, false))

	for epoch := 0; epoch < epochs; epoch++ {
		state.Epoch = epoch + 1
		state.ActiveSample = -1
		state.HasZ = false
		state.HasPred = false
		state.Error = 0

		epochCaption := fmt.Sprintf("Starting epoch %d / %d.", epoch+1, epochs)
		state.Caption = epochCaption
		b.AddStep("L4", trace.ActionHighlight, epochCaption, state.Clone(), varsMap(weights, bias, epoch+1, -1, 0, 0, lr, false))

		converged := true
		convInitCaption := "Set convergence flag = true (assuming no errors this epoch)."
		state.Caption = convInitCaption
		b.AddStep("L5", trace.ActionSet, convInitCaption, state.Clone(), varsMap(weights, bias, epoch+1, -1, 0, 0, lr, false))

		for i, x := range inputs {
			y := labels[i]
			state.ActiveSample = i
			state.ActiveTarget = y
			state.HasZ = false
			state.HasPred = false

			sampleCaption := fmt.Sprintf("Sample %d: input vector x = %v, target label y = %d.", i, x, y)
			state.Caption = sampleCaption
			b.AddStep("L6", trace.ActionHighlight, sampleCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, 0, lr, false))

			linear := bias
			for j, val := range x {
				linear += weights[j] * val
			}
			b.IncOperations()

			state.LinearZ = linear
			state.HasZ = true
			linearCaption := fmt.Sprintf("Linear combination: z = bias + w·x = %.2f + dot(%v, %v) = %.2f.", bias, weights, x, linear)
			state.Caption = linearCaption
			b.AddStep("L7", trace.ActionHighlight, linearCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, 0, lr, false))

			pred := 0
			if linear >= 0 {
				pred = 1
			}
			b.IncComparisons()

			state.ActivePred = pred
			state.HasPred = true
			predCaption := fmt.Sprintf("Heaviside step activation: z = %.2f >= 0 -> predicted label = %d.", linear, pred)
			state.Caption = predCaption
			b.AddStep("L8", trace.ActionCompare, predCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))

			err := y - pred
			b.IncComparisons()

			state.Error = err
			errCaption := fmt.Sprintf("Compute error: err = target (%d) - pred (%d) = %d.", y, pred, err)
			state.Caption = errCaption
			b.AddStep("L9", trace.ActionCompare, errCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))

			if err != 0 {
				// Misclassification: exact required caption
				misclassCaption := fmt.Sprintf("Misclassified! Target=%d, Pred=%d. Updating weights: delta_w = lr * err * x", y, pred)
				state.Caption = misclassCaption
				b.AddStep("L10", trace.ActionCompare, misclassCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))

				converged = false
				flagCaption := "Set convergence flag = false due to misclassification."
				state.Caption = flagCaption
				b.AddStep("L11", trace.ActionSet, flagCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))

				for j := range weights {
					weights[j] += lr * float64(err) * x[j]
				}
				b.IncSwaps()
				state.Weights = slices.Clone(weights)

				wUpCaption := fmt.Sprintf("Updated weights via rule w += lr * err * x -> %v.", weights)
				state.Caption = wUpCaption
				b.AddStep("L12", trace.ActionSwap, wUpCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))

				bias += lr * float64(err)
				b.IncSwaps()
				state.Bias = bias

				bUpCaption := fmt.Sprintf("Updated bias via rule bias += lr * err -> %.2f.", bias)
				state.Caption = bUpCaption
				b.AddStep("L13", trace.ActionSet, bUpCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))
			} else {
				// Correct prediction: exact required caption
				correctCaption := "Correct prediction! Weights remain unchanged."
				state.Caption = correctCaption
				b.AddStep("L10", trace.ActionVisit, correctCaption, state.Clone(), varsMap(weights, bias, epoch+1, i, y, pred, lr, false))
			}
		}

		if converged {
			state.Converged = true
			state.ActiveSample = -1
			state.HasZ = false
			state.HasPred = false
			earlyCaption := fmt.Sprintf("All samples classified correctly in epoch %d. Perceptron converged early!", epoch+1)
			state.Caption = earlyCaption
			b.AddStep("L16", trace.ActionHighlight, earlyCaption, state.Clone(), varsMap(weights, bias, epoch+1, -1, 0, 0, lr, true))
			break
		}
	}

	state.Converged = true
	state.ActiveSample = -1
	state.HasZ = false
	state.HasPred = false
	doneCaption := fmt.Sprintf("Perceptron training complete! Final weights: %v, bias: %.2f.", weights, bias)
	state.Caption = doneCaption
	b.AddStep("L18", trace.ActionDone, doneCaption, state.Clone(), varsMap(weights, bias, epochs, -1, 0, 0, lr, true))

	return b.Steps()
}

func varsMap(weights []float64, bias float64, epoch, sample, target, pred int, lr float64, converged bool) map[string]any {
	cp := make([]float64, len(weights))
	copy(cp, weights)
	return map[string]any{
		"weights":   cp,
		"bias":      bias,
		"epoch":     epoch,
		"sample":    sample,
		"target":    target,
		"pred":      pred,
		"lr":        lr,
		"converged": converged,
	}
}
