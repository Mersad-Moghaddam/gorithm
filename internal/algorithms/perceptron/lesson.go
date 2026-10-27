package perceptron

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
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
	defaultInputs := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	defaultLabels := []int{0, 1, 1, 1}
	defaultLR := 0.1
	defaultEpochs := 10

	trimmed := bytes.TrimSpace(input)
	if len(trimmed) == 0 {
		return Trace(defaultInputs, defaultLabels, defaultLR, defaultEpochs), nil
	}

	if bytes.Equal(trimmed, []byte("null")) {
		return nil, fmt.Errorf("perceptron: input cannot be null")
	}

	var payload PerceptronInput
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return nil, fmt.Errorf("perceptron: invalid input json: %w", err)
	}

	if payload.Inputs == nil {
		return nil, fmt.Errorf("perceptron: inputs field is required")
	}
	if len(payload.Inputs) == 0 {
		return nil, fmt.Errorf("perceptron: inputs cannot be empty")
	}
	if payload.Labels == nil {
		return nil, fmt.Errorf("perceptron: labels field is required")
	}
	if len(payload.Labels) != len(payload.Inputs) {
		return nil, fmt.Errorf("perceptron: labels length (%d) must match inputs length (%d)", len(payload.Labels), len(payload.Inputs))
	}

	d := len(payload.Inputs[0])
	if d == 0 {
		return nil, fmt.Errorf("perceptron: sample feature vector cannot be empty")
	}
	for i, x := range payload.Inputs {
		if len(x) != d {
			return nil, fmt.Errorf("perceptron: sample %d dimension %d does not match first sample %d", i, len(x), d)
		}
	}

	normLabels := make([]int, len(payload.Labels))
	for i, lbl := range payload.Labels {
		if lbl == -1 {
			normLabels[i] = 0
		} else if lbl == 0 || lbl == 1 {
			normLabels[i] = lbl
		} else {
			return nil, fmt.Errorf("perceptron: invalid label %d at index %d; must be 0 or 1 (or -1 and 1)", lbl, i)
		}
	}

	lr := defaultLR
	if payload.LR < 0 {
		return nil, fmt.Errorf("perceptron: learning rate cannot be negative")
	}
	if payload.LR > 0 {
		lr = payload.LR
	}

	epochs := defaultEpochs
	if payload.Epochs < 0 {
		return nil, fmt.Errorf("perceptron: epochs cannot be negative")
	}
	if payload.Epochs > 0 {
		epochs = payload.Epochs
	}
	if epochs > 50 {
		epochs = 50
	}

	return Trace(payload.Inputs, normLabels, lr, epochs), nil
}

func init() {
	catalog.Register(Lesson{})
}
