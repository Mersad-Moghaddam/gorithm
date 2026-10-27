package perceptron

import (
	"encoding/json"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestPerceptronConvergence_OR(t *testing.T) {
	inputs := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	labels := []int{0, 1, 1, 1}

	weights, bias := Train(inputs, labels, 0.1, 20)

	for i, x := range inputs {
		pred := Predict(x, weights, bias)
		if pred != labels[i] {
			t.Fatalf("OR gate input %v: expected %d, got %d (weights=%v, bias=%v)", x, labels[i], pred, weights, bias)
		}
	}
}

func TestPerceptronConvergence_AND(t *testing.T) {
	inputs := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	labels := []int{0, 0, 0, 1}

	weights, bias := Train(inputs, labels, 0.1, 20)

	for i, x := range inputs {
		pred := Predict(x, weights, bias)
		if pred != labels[i] {
			t.Fatalf("AND gate input %v: expected %d, got %d (weights=%v, bias=%v)", x, labels[i], pred, weights, bias)
		}
	}
}

func TestPerceptronEmptyInputs(t *testing.T) {
	w, b := Train([][]float64{}, []int{}, 0.1, 10)
	if w != nil || b != 0.0 {
		t.Fatalf("expected nil weights and 0 bias for empty inputs, got %v, %v", w, b)
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "perceptron" {
		t.Errorf("expected ID 'perceptron', got %q", meta.ID)
	}
	if meta.Name != "Perceptron (Neural Unit)" {
		t.Errorf("expected Name 'Perceptron (Neural Unit)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryAI {
		t.Errorf("expected Category %q, got %q", catalog.CategoryAI, meta.Category)
	}
	if meta.Renderer != "array" {
		t.Errorf("expected Renderer 'array', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(epochs · n · d)" {
		t.Errorf("expected TimeComplexity 'O(epochs · n · d)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(d)" {
		t.Errorf("expected SpaceComplexity 'O(d)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Easy" {
		t.Errorf("expected Difficulty 'Easy', got %q", meta.Documentation.Difficulty)
	}
	if len(meta.Documentation.KeyConcepts) != 4 {
		t.Fatalf("expected 4 KeyConcepts, got %d", len(meta.Documentation.KeyConcepts))
	}

	expectedConcepts := []string{
		"Artificial Neuron",
		"Activation Function",
		"Linear Separability",
		"Weight Update Rule",
	}
	for i, concept := range expectedConcepts {
		if meta.Documentation.KeyConcepts[i] != concept {
			t.Errorf("expected KeyConcept[%d] = %q, got %q", i, concept, meta.Documentation.KeyConcepts[i])
		}
	}

	registered, ok := catalog.Get("perceptron")
	if !ok || registered == nil {
		t.Fatal("perceptron not registered in catalog")
	}
}

func TestLessonTraceValidation(t *testing.T) {
	lesson := Lesson{}

	// Default build
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) error: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected at least one step in default trace")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on default build: %v", err)
	}

	// Verify state and variables in steps
	for i, step := range steps {
		state, ok := step.State.(renderers.ArrayState)
		if !ok {
			t.Fatalf("step %d state is not renderers.ArrayState, got %T", i, step.State)
		}
		if len(state.Values) == 0 {
			t.Fatalf("step %d ArrayState.Values is empty", i)
		}
		if _, ok := step.Variables["weights"]; !ok {
			t.Fatalf("step %d missing 'weights' variable", i)
		}
		if _, ok := step.Variables["bias"]; !ok {
			t.Fatalf("step %d missing 'bias' variable", i)
		}
	}

	// Verify SVG renderer works on the first step
	svg := renderers.RenderArraySVG(steps[0], 800, 400)
	if len(svg) == 0 {
		t.Fatal("RenderArraySVG produced empty SVG")
	}

	// Custom dataset build
	customInput, _ := json.Marshal(map[string]any{
		"inputs": [][]float64{
			{0, 0},
			{1, 1},
		},
		"labels": []int{0, 1},
		"lr":     0.2,
		"epochs": 5,
	})
	customSteps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("Build(customInput) error: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, customSteps); err != nil {
		t.Fatalf("line validation failed on custom build: %v", err)
	}
}
