package perceptron

// Train trains a single-layer perceptron on binary classification data.
func Train(inputs [][]float64, labels []int, lr float64, epochs int) ([]float64, float64) {
	if len(inputs) == 0 || len(labels) == 0 {
		return nil, 0
	}
	d := len(inputs[0])
	weights := make([]float64, d)
	bias := 0.0

	for epoch := 0; epoch < epochs; epoch++ {
		converged := true
		for i, x := range inputs {
			linear := bias
			for j, val := range x {
				linear += weights[j] * val
			}
			pred := 0
			if linear >= 0 {
				pred = 1
			}
			err := labels[i] - pred
			if err != 0 {
				converged = false
				for j := range weights {
					weights[j] += lr * float64(err) * x[j]
				}
				bias += lr * float64(err)
			}
		}
		if converged {
			break
		}
	}
	return weights, bias
}

// Predict calculates the binary classification output (0 or 1) for a single input vector.
func Predict(x []float64, weights []float64, bias float64) int {
	linear := bias
	for j, val := range x {
		linear += weights[j] * val
	}
	if linear >= 0 {
		return 1
	}
	return 0
}
