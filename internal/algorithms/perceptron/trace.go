package perceptron

import (
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// Trace executes perceptron training and emits visualization steps.
func Trace(inputs [][]float64, labels []int, lr float64, epochs int) []trace.Step {
	b := trace.NewBuilder()

	if len(inputs) == 0 || len(labels) == 0 {
		b.AddStep("L1", trace.ActionInit, "Empty dataset provided; perceptron cannot train.", renderers.ArrayState{
			Values:  []int{0},
			Current: -1,
		}, map[string]any{"weights": []float64{}, "bias": 0.0})
		b.AddStep("L18", trace.ActionDone, "Training completed with empty inputs.", renderers.ArrayState{
			Values:  []int{0},
			Current: -1,
		}, map[string]any{"weights": []float64{}, "bias": 0.0})
		return b.Steps()
	}

	d := len(inputs[0])
	weights := make([]float64, d)
	bias := 0.0

	b.AddStep("L1", trace.ActionInit, fmt.Sprintf("Initialize Perceptron training: %d samples, %d features, lr=%.2f, max epochs=%d.", len(inputs), d, lr, epochs), renderers.ArrayState{
		Values:  toBarValues(weights, bias),
		Current: -1,
	}, varsMap(weights, bias, 0, -1, 0, 0, lr))

	b.AddStep("L2", trace.ActionSet, fmt.Sprintf("Initialized weight vector of length %d with zeros.", d), renderers.ArrayState{
		Values:  toBarValues(weights, bias),
		Current: -1,
	}, varsMap(weights, bias, 0, -1, 0, 0, lr))

	b.AddStep("L3", trace.ActionSet, "Initialized bias scalar to 0.0.", renderers.ArrayState{
		Values:  toBarValues(weights, bias),
		Current: d,
	}, varsMap(weights, bias, 0, -1, 0, 0, lr))

	for epoch := 0; epoch < epochs; epoch++ {
		b.AddStep("L4", trace.ActionHighlight, fmt.Sprintf("Starting epoch %d / %d.", epoch+1, epochs), renderers.ArrayState{
			Values:  toBarValues(weights, bias),
			Current: -1,
		}, varsMap(weights, bias, epoch+1, -1, 0, 0, lr))

		converged := true
		b.AddStep("L5", trace.ActionSet, "Set convergence flag = true (assuming no classification errors).", renderers.ArrayState{
			Values:  toBarValues(weights, bias),
			Current: -1,
		}, varsMap(weights, bias, epoch+1, -1, 0, 0, lr))

		for i, x := range inputs {
			y := labels[i]
			b.AddStep("L6", trace.ActionHighlight, fmt.Sprintf("Sample %d: input vector x = %v, target label y = %d.", i, x, y), renderers.ArrayState{
				Values:  toBarValues(weights, bias),
				Current: -1,
			}, varsMap(weights, bias, epoch+1, i, y, 0, lr))

			linear := bias
			for j, val := range x {
				linear += weights[j] * val
			}
			b.IncOperations()

			b.AddStep("L7", trace.ActionHighlight, fmt.Sprintf("Linear combination: z = bias + w·x = %.2f + dot(%v, %v) = %.2f.", bias, weights, x, linear), renderers.ArrayState{
				Values:    toBarValues(weights, bias),
				Comparing: allIndices(d + 1),
				Current:   -1,
			}, varsMap(weights, bias, epoch+1, i, y, 0, lr))

			pred := 0
			if linear >= 0 {
				pred = 1
			}
			b.IncComparisons()

			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Heaviside step activation: z = %.2f >= 0 -> predicted label = %d.", linear, pred), renderers.ArrayState{
				Values:    toBarValues(weights, bias),
				Comparing: allIndices(d + 1),
				Current:   -1,
			}, varsMap(weights, bias, epoch+1, i, y, pred, lr))

			err := y - pred
			b.IncComparisons()

			b.AddStep("L9", trace.ActionCompare, fmt.Sprintf("Compute error: err = target (%d) - pred (%d) = %d.", y, pred, err), renderers.ArrayState{
				Values:  toBarValues(weights, bias),
				Current: -1,
			}, varsMap(weights, bias, epoch+1, i, y, pred, lr))

			if err != 0 {
				b.AddStep("L10", trace.ActionCompare, fmt.Sprintf("Misclassification detected (err = %d). Weights and bias must be adjusted.", err), renderers.ArrayState{
					Values:  toBarValues(weights, bias),
					Swapped: allIndices(d + 1),
					Current: -1,
				}, varsMap(weights, bias, epoch+1, i, y, pred, lr))

				converged = false
				b.AddStep("L11", trace.ActionSet, "Set convergence flag = false due to misclassification.", renderers.ArrayState{
					Values:  toBarValues(weights, bias),
					Current: -1,
				}, varsMap(weights, bias, epoch+1, i, y, pred, lr))

				for j := range weights {
					weights[j] += lr * float64(err) * x[j]
				}
				b.IncSwaps()

				b.AddStep("L12", trace.ActionSwap, fmt.Sprintf("Updated weights via rule w += lr * err * x -> %v.", weights), renderers.ArrayState{
					Values:  toBarValues(weights, bias),
					Swapped: allIndices(d),
					Current: -1,
				}, varsMap(weights, bias, epoch+1, i, y, pred, lr))

				bias += lr * float64(err)
				b.IncSwaps()

				b.AddStep("L13", trace.ActionSet, fmt.Sprintf("Updated bias via rule bias += lr * err -> %.2f.", bias), renderers.ArrayState{
					Values:  toBarValues(weights, bias),
					Swapped: []int{d},
					Current: d,
				}, varsMap(weights, bias, epoch+1, i, y, pred, lr))
			} else {
				b.AddStep("L10", trace.ActionVisit, fmt.Sprintf("Sample %d classified correctly (pred %d == target %d). No weight update needed.", i, pred, y), renderers.ArrayState{
					Values:  toBarValues(weights, bias),
					Sorted:  allIndices(d + 1),
					Current: -1,
				}, varsMap(weights, bias, epoch+1, i, y, pred, lr))
			}
		}

		if converged {
			b.AddStep("L16", trace.ActionHighlight, fmt.Sprintf("All samples classified correctly in epoch %d. Perceptron converged early!", epoch+1), renderers.ArrayState{
				Values:  toBarValues(weights, bias),
				Sorted:  allIndices(d + 1),
				Current: -1,
			}, varsMap(weights, bias, epoch+1, -1, 0, 0, lr))
			break
		}
	}

	b.AddStep("L18", trace.ActionDone, fmt.Sprintf("Perceptron training complete! Final weights: %v, bias: %.2f.", weights, bias), renderers.ArrayState{
		Values:  toBarValues(weights, bias),
		Sorted:  allIndices(d + 1),
		Current: -1,
	}, varsMap(weights, bias, epochs, -1, 0, 0, lr))

	return b.Steps()
}

func toBarValues(weights []float64, bias float64) []int {
	scale := 100.0
	vals := make([]int, len(weights)+1)
	for i, w := range weights {
		vals[i] = int(math.Round(w * scale))
	}
	vals[len(weights)] = int(math.Round(bias * scale))
	return vals
}

func allIndices(n int) []int {
	idx := make([]int, n)
	for i := 0; i < n; i++ {
		idx[i] = i
	}
	return idx
}

func varsMap(weights []float64, bias float64, epoch, sample, target, pred int, lr float64) map[string]any {
	cp := make([]float64, len(weights))
	copy(cp, weights)
	return map[string]any{
		"weights": cp,
		"bias":    bias,
		"epoch":   epoch,
		"sample":  sample,
		"target":  target,
		"pred":    pred,
		"lr":      lr,
	}
}
