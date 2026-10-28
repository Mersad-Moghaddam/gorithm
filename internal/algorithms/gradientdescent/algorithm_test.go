package gradientdescent

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestTrainConvergence(t *testing.T) {
	// y = 2x + 1
	x := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	y := []float64{3.0, 5.0, 7.0, 9.0, 11.0}
	lr := 0.02
	epochs := 2000

	w, b, costHistory := Train(x, y, lr, epochs)

	if len(costHistory) != epochs {
		t.Fatalf("expected cost history length %d, got %d", epochs, len(costHistory))
	}

	if costHistory[0] <= costHistory[len(costHistory)-1] {
		t.Errorf("cost did not decrease: initial %f, final %f", costHistory[0], costHistory[len(costHistory)-1])
	}

	if math.Abs(w-2.0) > 0.05 {
		t.Errorf("weight did not converge to 2.0: got %f", w)
	}

	if math.Abs(b-1.0) > 0.15 {
		t.Errorf("bias did not converge to 1.0: got %f", b)
	}
}

func TestTrainEdgeCases(t *testing.T) {
	w, b, history := Train(nil, nil, 0.01, 100)
	if w != 0 || b != 0 || history != nil {
		t.Errorf("expected 0, 0, nil for nil slices, got %f, %f, %v", w, b, history)
	}

	w, b, history = Train([]float64{1.0}, []float64{1.0, 2.0}, 0.01, 100)
	if w != 0 || b != 0 || history != nil {
		t.Errorf("expected 0, 0, nil for mismatched lengths, got %f, %f, %v", w, b, history)
	}

	w, b, history = Train([]float64{1.0}, []float64{2.0}, 0.01, 0)
	if w != 0 || b != 0 || history != nil {
		t.Errorf("expected 0, 0, nil for non-positive epochs, got %f, %f, %v", w, b, history)
	}
}

func TestLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "gradientdescent" {
		t.Errorf("expected ID 'gradientdescent', got %q", meta.ID)
	}
	if meta.Name != "Gradient Descent (Linear Regression)" {
		t.Errorf("expected Name 'Gradient Descent (Linear Regression)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryAI {
		t.Errorf("expected Category %q, got %q", catalog.CategoryAI, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}

	doc := meta.Documentation
	if doc.TimeComplexity != "O(epochs · n)" {
		t.Errorf("expected TimeComplexity 'O(epochs · n)', got %q", doc.TimeComplexity)
	}
	if doc.SpaceComplexity != "O(1)" {
		t.Errorf("expected SpaceComplexity 'O(1)', got %q", doc.SpaceComplexity)
	}
	if doc.Difficulty != "Medium" {
		t.Errorf("expected Difficulty 'Medium', got %q", doc.Difficulty)
	}

	requiredConcepts := []string{
		"Loss Function MSE",
		"Learning Rate",
		"Partial Derivatives",
		"Parameter Update",
	}
	for _, req := range requiredConcepts {
		found := false
		for _, kc := range doc.KeyConcepts {
			if kc == req {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing required KeyConcept: %q in %v", req, doc.KeyConcepts)
		}
	}

	if doc.Summary == "" {
		t.Error("Summary should not be empty")
	}
	if doc.HowItWorks == "" {
		t.Error("HowItWorks should not be empty")
	}
}

func TestValidateSourceLineIDs(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed: %v", err)
	}
}

func TestTraceStepsAndTableState(t *testing.T) {
	x := []float64{1.0, 2.0, 3.0, 4.0}
	y := []float64{3.0, 5.0, 7.0, 9.0}
	lr := 0.05
	epochs := 5

	steps := Trace(x, y, lr, epochs)
	if len(steps) == 0 {
		t.Fatal("expected steps from Trace")
	}

	for i, step := range steps {
		if step.LineID == "" {
			t.Errorf("step %d missing LineID", i)
		}
		if step.Action == "" {
			t.Errorf("step %d missing Action", i)
		}
		if step.Explanation == "" {
			t.Errorf("step %d missing Explanation", i)
		}

		vars := step.Variables
		if _, ok := vars["loss"]; !ok {
			t.Errorf("step %d missing 'loss' variable", i)
		}
		if _, ok := vars["w"]; !ok {
			t.Errorf("step %d missing 'w' variable", i)
		}
		if _, ok := vars["b"]; !ok {
			t.Errorf("step %d missing 'b' variable", i)
		}

		lessonState, ok := step.State.(renderers.GradientDescentLessonState)
		tableState := lessonState.TableState
		if !ok {
			t.Fatalf("step %d state is not renderers.GradientDescentLessonState: %T", i, step.State)
		}
		if len(tableState.ColHeaders) == 0 {
			t.Errorf("step %d TableState ColHeaders is empty", i)
		}
		if len(tableState.RowHeaders) == 0 {
			t.Errorf("step %d TableState RowHeaders is empty", i)
		}
		if len(tableState.Cells) == 0 {
			t.Errorf("step %d TableState Cells is empty", i)
		}
	}
}

func TestBuildCustomInput(t *testing.T) {
	lesson := Lesson{}

	customJSON := `{"x": [1.0, 2.0, 3.0], "y": [2.0, 4.0, 6.0], "lr": 0.01, "epochs": 3}`
	steps, err := lesson.Build(json.RawMessage(customJSON))
	if err != nil {
		t.Fatalf("Build with custom input failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected non-empty steps with custom input")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed for custom input: %v", err)
	}
}

func TestCatalogRegistration(t *testing.T) {
	registered, ok := catalog.Get("gradientdescent")
	if !ok {
		t.Fatal("Lesson not registered in catalog")
	}
	if registered.Meta().ID != "gradientdescent" {
		t.Errorf("expected ID 'gradientdescent', got %q", registered.Meta().ID)
	}
}
