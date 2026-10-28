package gradientdescent

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
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

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	x := []float64{1.0, 2.0, 3.0, 4.0}
	y := []float64{3.0, 5.0, 7.0, 9.0}
	lr := 0.05
	epochs := 10

	if len(bytes.TrimSpace(input)) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("gradient descent input must be a JSON object")
		}
		for _, key := range []string{"x", "y", "lr", "epochs"} {
			if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("%s must not be null", key)
			}
		}
		_, hasX := fields["x"]
		_, hasY := fields["y"]
		if hasX != hasY {
			return nil, fmt.Errorf("x and y must be supplied together")
		}
		var custom struct {
			X      []float64 `json:"x"`
			Y      []float64 `json:"y"`
			LR     *float64  `json:"lr"`
			Epochs *int      `json:"epochs"`
		}
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, fmt.Errorf("invalid gradient descent input: %w", err)
		} else {
			if hasX {
				x = custom.X
				y = custom.Y
			}
			if custom.LR != nil {
				lr = *custom.LR
			}
			if custom.Epochs != nil {
				epochs = *custom.Epochs
			}
		}
	}

	if err := gradientDescentValidate(x, y, lr, epochs); err != nil {
		return nil, err
	}
	steps := Trace(x, y, lr, epochs)
	if s := steps[len(steps)-1].State.(renderers.GradientDescentLessonState); s.StopReason == "numerical-error" {
		return nil, fmt.Errorf("numerical overflow during regression; reduce data magnitude, learning rate or epochs")
	}
	return steps, nil
}

func init() {
	catalog.Register(Lesson{})
}
