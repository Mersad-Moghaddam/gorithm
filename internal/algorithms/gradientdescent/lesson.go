package gradientdescent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "gradientdescent",
		Name:     "Gradient Descent (Linear Regression)",
		Category: catalog.CategoryAI,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "Optimizes linear regression parameters (weight and bias) iteratively by calculating the gradient of the Mean Squared Error (MSE) loss function and stepping in the direction of steepest descent.",
			TimeComplexity:  "O(epochs · n)",
			SpaceComplexity: "O(1)",
			Difficulty:      "Medium",
			KeyConcepts: []string{
				"Loss Function MSE",
				"Learning Rate",
				"Partial Derivatives",
				"Parameter Update",
			},
			HowItWorks: "Initializes weight w and bias b to zero. In each epoch, computes model predictions y_pred = w*x + b and evaluates Mean Squared Error loss. Calculates partial derivatives of MSE with respect to w and b across all n training points, then updates parameters in the opposite direction of the gradient scaled by learning rate lr.",
		},
		DefaultInput: `{"x": [1, 2, 3, 4], "y": [3, 5, 7, 9], "lr": 0.05, "epochs": 10}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Train(x, y []float64, lr float64, epochs int) (float64, float64, []float64) {"},
		{ID: "L2", Number: 2, Code: "    w, b := 0.0, 0.0"},
		{ID: "L3", Number: 3, Code: "    costHistory := make([]float64, epochs)"},
		{ID: "L4", Number: 4, Code: "    for epoch := 0; epoch < epochs; epoch++ {"},
		{ID: "L5", Number: 5, Code: "        var cost, dw, db float64"},
		{ID: "L6", Number: 6, Code: "        for i := 0; i < len(x); i++ {"},
		{ID: "L7", Number: 7, Code: "            pred := w*x[i] + b"},
		{ID: "L8", Number: 8, Code: "            err := pred - y[i]"},
		{ID: "L9", Number: 9, Code: "            cost += err * err"},
		{ID: "L10", Number: 10, Code: "           dw += err * x[i]"},
		{ID: "L11", Number: 11, Code: "           db += err"},
		{ID: "L12", Number: 12, Code: "       }"},
		{ID: "L13", Number: 13, Code: "       costHistory[epoch] = cost / float64(len(x))"},
		{ID: "L14", Number: 14, Code: "       w -= lr * (2.0 / float64(len(x))) * dw"},
		{ID: "L15", Number: 15, Code: "       b -= lr * (2.0 / float64(len(x))) * db"},
		{ID: "L16", Number: 16, Code: "   }"},
		{ID: "L17", Number: 17, Code: "   return w, b, costHistory"},
		{ID: "L18", Number: 18, Code: "}"},
	}
}

type LessonInput struct {
	X      *[]float64 `json:"x"`
	Y      *[]float64 `json:"y"`
	LR     *float64   `json:"lr"`
	Epochs *int       `json:"epochs"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	x := []float64{1.0, 2.0, 3.0, 4.0}
	y := []float64{3.0, 5.0, 7.0, 9.0}
	lr := 0.05
	epochs := 10

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace(x, y, lr, epochs), nil
	}
	if bytes.Equal(trimmed, []byte("null")) {
		return nil, fmt.Errorf("gradient descent input cannot be null")
	}

	var custom LessonInput
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	if err := dec.Decode(&custom); err != nil {
		return nil, fmt.Errorf("malformed JSON input: %w", err)
	}
	if dec.More() {
		return nil, fmt.Errorf("unexpected trailing data in input")
	}

	if custom.X == nil || custom.Y == nil {
		return nil, fmt.Errorf("both x and y arrays are required")
	}
	if len(*custom.X) == 0 || len(*custom.Y) == 0 {
		return nil, fmt.Errorf("x and y arrays cannot be empty")
	}
	if len(*custom.X) != len(*custom.Y) {
		return nil, fmt.Errorf("mismatched x and y lengths: %d != %d", len(*custom.X), len(*custom.Y))
	}
	for i, v := range *custom.X {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("x[%d] must be a finite number", i)
		}
	}
	for i, v := range *custom.Y {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("y[%d] must be a finite number", i)
		}
	}

	x = *custom.X
	y = *custom.Y

	if custom.LR != nil {
		if *custom.LR <= 0 || math.IsNaN(*custom.LR) || math.IsInf(*custom.LR, 0) {
			return nil, fmt.Errorf("learning rate must be positive and finite")
		}
		lr = *custom.LR
	}

	if custom.Epochs != nil {
		if *custom.Epochs <= 0 {
			return nil, fmt.Errorf("epochs must be positive")
		}
		epochs = *custom.Epochs
	}

	return Trace(x, y, lr, epochs), nil
}

func init() {
	catalog.Register(Lesson{})
}
