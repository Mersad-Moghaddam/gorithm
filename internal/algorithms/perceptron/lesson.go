package perceptron

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
		ID:       "perceptron",
		Name:     "Perceptron (Neural Unit)",
		Category: catalog.CategoryAI,
		Renderer: "array",
		Documentation: catalog.Documentation{
			Summary:         "The foundational artificial neural network unit for supervised binary classification of linearly separable patterns.",
			TimeComplexity:  "O(epochs · n · d)",
			SpaceComplexity: "O(d)",
			Difficulty:      "Easy",
			KeyConcepts: []string{
				"Artificial Neuron",
				"Activation Function",
				"Linear Separability",
				"Weight Update Rule",
			},
			HowItWorks: "The Perceptron initializes weight coefficients and a bias term to zero. For each training epoch, it processes input feature vectors by computing the weighted linear combination z = bias + w · x and passing it through a Heaviside step activation function. When the predicted binary label differs from the true target, weights and bias are updated proportionally to the learning rate and error. The process repeats until all samples are classified correctly or the maximum epoch limit is reached.",
		},
		DefaultInput: `{"inputs": [[0, 0], [0, 1], [1, 0], [1, 1]], "labels": [0, 1, 1, 1], "lr": 0.1, "epochs": 10}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Train(inputs [][]float64, labels []int, lr float64, epochs int) ([]float64, float64) {"},
		{ID: "L2", Number: 2, Code: "    weights := make([]float64, len(inputs[0]))"},
		{ID: "L3", Number: 3, Code: "    bias := 0.0"},
		{ID: "L4", Number: 4, Code: "    for epoch := 0; epoch < epochs; epoch++ {"},
		{ID: "L5", Number: 5, Code: "        converged := true"},
		{ID: "L6", Number: 6, Code: "        for i, x := range inputs {"},
		{ID: "L7", Number: 7, Code: "            linear := bias + dot(weights, x)"},
		{ID: "L8", Number: 8, Code: "            pred := step(linear)"},
		{ID: "L9", Number: 9, Code: "            err := labels[i] - pred"},
		{ID: "L10", Number: 10, Code: "            if err != 0 {"},
		{ID: "L11", Number: 11, Code: "                converged = false"},
		{ID: "L12", Number: 12, Code: "                weights += lr * err * x"},
		{ID: "L13", Number: 13, Code: "                bias += lr * err"},
		{ID: "L14", Number: 14, Code: "            }"},
		{ID: "L15", Number: 15, Code: "        }"},
		{ID: "L16", Number: 16, Code: "        if converged { break }"},
		{ID: "L17", Number: 17, Code: "    }"},
		{ID: "L18", Number: 18, Code: "    return weights, bias"},
		{ID: "L19", Number: 19, Code: "}"},
	}
}

type PerceptronInput struct {
	Inputs [][]float64 `json:"inputs"`
	Labels []int       `json:"labels"`
	LR     float64     `json:"lr"`
	Epochs int         `json:"epochs"`
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	inputs := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	labels := []int{0, 1, 1, 1}
	lr := 0.1
	epochs := 10

	payload := PerceptronInput{Inputs: inputs, Labels: labels, LR: lr, Epochs: epochs}
	if len(bytes.TrimSpace(input)) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil {
			return nil, err
		}
		if fields == nil {
			return nil, fmt.Errorf("perceptron input must be an object")
		}
		for name, value := range fields {
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return nil, fmt.Errorf("%s cannot be null", name)
			}
		}
		if err := json.Unmarshal(input, &payload); err != nil {
			return nil, err
		}
	}
	if err := validateLessonInput(payload.Inputs, payload.Labels, payload.LR, payload.Epochs); err != nil {
		return nil, err
	}
	steps := Trace(payload.Inputs, payload.Labels, payload.LR, payload.Epochs)
	if steps[len(steps)-1].State.(renderers.PerceptronLessonState).StopReason == "numerical-error" {
		return nil, fmt.Errorf("perceptron arithmetic overflow; use smaller features")
	}
	return steps, nil
}

func init() {
	catalog.Register(Lesson{})
}
