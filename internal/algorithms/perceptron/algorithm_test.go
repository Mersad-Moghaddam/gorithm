package perceptron

import (
	"encoding/json"
	"fmt"
	"strings"
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

	steps := Trace([][]float64{}, []int{}, 0.1, 10)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps for empty inputs, got %d", len(steps))
	}
	if steps[0].LineID != "L1" || steps[1].LineID != "L18" {
		t.Errorf("unexpected line IDs for empty steps: %s, %s", steps[0].LineID, steps[1].LineID)
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

func TestLessonBuildValidation(t *testing.T) {
	lesson := Lesson{}

	// Valid builds
	if _, err := lesson.Build(nil); err != nil {
		t.Fatalf("Build(nil) unexpected error: %v", err)
	}
	if _, err := lesson.Build([]byte{}); err != nil {
		t.Fatalf("Build([]) unexpected error: %v", err)
	}
	if _, err := lesson.Build([]byte("   ")); err != nil {
		t.Fatalf("Build(spaces) unexpected error: %v", err)
	}

	// -1 and 1 label normalization
	negLabelInput := `{"inputs": [[0, 0], [1, 1]], "labels": [-1, 1]}`
	if steps, err := lesson.Build([]byte(negLabelInput)); err != nil {
		t.Fatalf("Build(negLabels) unexpected error: %v", err)
	} else if len(steps) == 0 {
		t.Fatal("expected steps for negLabels input")
	}

	// Invalid builds: null
	if _, err := lesson.Build([]byte("null")); err == nil {
		t.Fatal("expected error for 'null' input, got nil")
	}

	// Invalid builds: malformed JSON
	if _, err := lesson.Build([]byte("{not a valid json")); err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}

	// Invalid builds: empty inputs
	if _, err := lesson.Build([]byte(`{"inputs": [], "labels": []}`)); err == nil {
		t.Fatal("expected error for empty inputs, got nil")
	}

	// Invalid builds: missing labels
	if _, err := lesson.Build([]byte(`{"inputs": [[1, 2]]}`)); err == nil {
		t.Fatal("expected error for missing labels, got nil")
	}

	// Invalid builds: mismatched counts
	if _, err := lesson.Build([]byte(`{"inputs": [[1, 2]], "labels": [0, 1]}`)); err == nil {
		t.Fatal("expected error for mismatched inputs and labels lengths, got nil")
	}

	// Invalid builds: empty feature vector
	if _, err := lesson.Build([]byte(`{"inputs": [[]], "labels": [0]}`)); err == nil {
		t.Fatal("expected error for empty feature vector, got nil")
	}

	// Invalid builds: mismatched feature dimensions
	if _, err := lesson.Build([]byte(`{"inputs": [[1, 2], [1]], "labels": [0, 1]}`)); err == nil {
		t.Fatal("expected error for mismatched feature dimensions, got nil")
	}

	// Invalid builds: non-binary label
	if _, err := lesson.Build([]byte(`{"inputs": [[1, 2]], "labels": [42]}`)); err == nil {
		t.Fatal("expected error for non-binary label 42, got nil")
	}

	// Invalid builds: negative learning rate
	if _, err := lesson.Build([]byte(`{"inputs": [[1, 2]], "labels": [1], "lr": -0.5}`)); err == nil {
		t.Fatal("expected error for negative learning rate, got nil")
	}

	// Invalid builds: negative epochs
	if _, err := lesson.Build([]byte(`{"inputs": [[1, 2]], "labels": [1], "epochs": -2}`)); err == nil {
		t.Fatal("expected error for negative epochs, got nil")
	}
}

func TestPerceptronSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	inputs := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	labels := []int{0, 1, 1, 1}

	steps := Trace(inputs, labels, 0.1, 10)
	if len(steps) == 0 {
		t.Fatal("expected non-empty trace")
	}

	// Step 0 must have zero weights and bias
	st0, ok := steps[0].State.(PerceptronState)
	if !ok {
		t.Fatalf("expected PerceptronState, got %T", steps[0].State)
	}
	if len(st0.Weights) != 2 || st0.Weights[0] != 0.0 || st0.Weights[1] != 0.0 || st0.Bias != 0.0 {
		t.Fatalf("step 0 parameters corrupted: w=%v, bias=%v", st0.Weights, st0.Bias)
	}

	// Mutate input array passed into Trace
	inputs[0][0] = 999.0
	inputs[0][1] = 888.0

	// Check that snapshot 0 points were NOT affected
	if st0.Points[0].X[0] != 0.0 || st0.Points[0].X[1] != 0.0 {
		t.Fatalf("mutating input array mutated step 0 snapshot: pt[0]=%v", st0.Points[0].X)
	}

	// Mutate state inside later step
	stLast := steps[len(steps)-1].State.(PerceptronState)
	stLast.Weights[0] = -12345.0

	// Step 0 must still be zero
	st0Again := steps[0].State.(PerceptronState)
	if st0Again.Weights[0] != 0.0 {
		t.Fatalf("mutating step last mutated step 0: got %v", st0Again.Weights[0])
	}

	// Verify all steps validate against Source IDs
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("source line validation failed: %v", err)
	}
}

func TestPerceptronWeightUpdates(t *testing.T) {
	inputs := [][]float64{
		{0, 0},
		{0, 1},
		{1, 0},
		{1, 1},
	}
	labels := []int{0, 1, 1, 1}
	lr := 0.1
	steps := Trace(inputs, labels, lr, 10)

	foundMisclass := false
	foundCorrect := false

	for i, step := range steps {
		st, ok := step.State.(PerceptronState)
		if !ok {
			t.Fatalf("step %d state is not PerceptronState, got %T", i, step.State)
		}

		if step.LineID == "L10" && st.Error != 0 {
			foundMisclass = true
			expectedCaption := fmt.Sprintf("Misclassified! Target=%d, Pred=%d. Updating weights: delta_w = lr * err * x", st.ActiveTarget, st.ActivePred)
			if step.Explanation != expectedCaption {
				t.Errorf("step %d misclassification explanation mismatch:\nwant: %q\ngot:  %q", i, expectedCaption, step.Explanation)
			}
		}

		if step.LineID == "L10" && st.Error == 0 {
			foundCorrect = true
			expectedCaption := "Correct prediction! Weights remain unchanged."
			if step.Explanation != expectedCaption {
				t.Errorf("step %d correct prediction explanation mismatch:\nwant: %q\ngot:  %q", i, expectedCaption, step.Explanation)
			}
		}
	}

	if !foundMisclass {
		t.Fatal("expected at least one misclassification step")
	}
	if !foundCorrect {
		t.Fatal("expected at least one correct prediction step")
	}
}

func TestPerceptronSceneRendering(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build(nil) failed: %v", err)
	}

	for i, step := range steps {
		svg := renderers.RenderStepSVG(step, "array", 800, 450)
		if len(svg) == 0 {
			t.Fatalf("step %d rendered empty SVG", i)
		}

		// Check required keys and elements
		requiredKeys := []string{
			`data-key="caption"`,
			`data-key="plot-bg"`,
			`data-key="math-bg"`,
			`data-key="weights-card"`,
			`data-key="sample-card"`,
			`data-key="activation-card"`,
			`data-key="update-card"`,
		}
		for _, key := range requiredKeys {
			if !strings.Contains(svg, key) {
				t.Fatalf("step %d missing SVG element %s", i, key)
			}
		}

		// Ensure points are rendered
		if !strings.Contains(svg, `data-key="sample-point-0"`) {
			t.Fatalf("step %d missing sample-point-0", i)
		}

		// Check decision boundary or boundary status exists
		hasBoundary := strings.Contains(svg, `data-key="decision-boundary"`) ||
			strings.Contains(svg, `data-key="boundary-status"`)
		if !hasBoundary {
			t.Fatalf("step %d missing decision boundary or boundary status", i)
		}
	}
}

func TestCustomDatasetTrace(t *testing.T) {
	lesson := Lesson{}
	customInput, _ := json.Marshal(map[string]any{
		"inputs": [][]float64{
			{0, 0},
			{1, 1},
		},
		"labels": []int{0, 1},
		"lr":     0.2,
		"epochs": 5,
	})
	steps, err := lesson.Build(customInput)
	if err != nil {
		t.Fatalf("Build(customInput) failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps from custom build")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed on custom build: %v", err)
	}
}
