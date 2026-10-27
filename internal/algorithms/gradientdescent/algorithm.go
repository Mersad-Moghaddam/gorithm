package gradientdescent

// Train fits a linear regression model y = w*x + b using gradient descent.
func Train(x, y []float64, lr float64, epochs int) (float64, float64, []float64) {
	n := len(x)
	if n == 0 || n != len(y) || epochs <= 0 {
		return 0, 0, nil
	}

	w, b := 0.0, 0.0
	costHistory := make([]float64, epochs)

	for epoch := 0; epoch < epochs; epoch++ {
		var cost, dw, db float64
		for i := 0; i < n; i++ {
			pred := w*x[i] + b
			err := pred - y[i]
			cost += err * err
			dw += err * x[i]
			db += err
		}
		costHistory[epoch] = cost / float64(n)
		w -= lr * (2.0 / float64(n)) * dw
		b -= lr * (2.0 / float64(n)) * db
	}

	return w, b, costHistory
}
