package gradientdescent

import (
	"encoding/json"
	"math"
	"strings"
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

func TestTraceConvergenceAndCostReduction(t *testing.T) {
	// y = 2x + 1
	x := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	y := []float64{3.0, 5.0, 7.0, 9.0, 11.0}
	lr := 0.02
	epochs := 300

	steps := Trace(x, y, lr, epochs)
	if len(steps) == 0 {
		t.Fatal("expected non-empty trace")
	}

	var firstEpochLoss, lastEpochLoss float64
	var foundFirst, foundLast bool
	var hasUpdateCaption bool

	for _, s := range steps {
		tableState, ok := s.State.(renderers.TableState)
		if ok && strings.Contains(tableState.Caption, "updating w (dw=") && strings.Contains(tableState.Caption, ") and b (db=") {
			hasUpdateCaption = true
		}

		if s.LineID == "L13" {
			loss, ok := s.Variables["loss"].(float64)
			if ok {
				if !foundFirst {
					firstEpochLoss = loss
					foundFirst = true
				}
				lastEpochLoss = loss
				foundLast = true
			}
		}
	}

	if !foundFirst || !foundLast {
		t.Fatal("expected L13 MSE loss steps in trace")
	}

	if lastEpochLoss >= firstEpochLoss {
		t.Fatalf("expected cost reduction across epochs: initial=%f, final=%f", firstEpochLoss, lastEpochLoss)
	}

	if !hasUpdateCaption {
		t.Errorf("expected parameter update caption matching 'Epoch %%d: Loss=%%.4f, updating w (dw=%%.4f) and b (db=%%.4f)'")
	}

	finalStep := steps[len(steps)-1]
	finalW, okW := finalStep.Variables["w"].(float64)
	finalB, okB := finalStep.Variables["b"].(float64)
	if !okW || !okB {
		t.Fatal("final step missing w or b variable")
	}
	if math.Abs(finalW-2.0) > 0.2 {
		t.Errorf("expected final w near 2.0, got %f", finalW)
	}
	if math.Abs(finalB-1.0) > 0.5 {
		t.Errorf("expected final b near 1.0, got %f", finalB)
	}
}

func TestSnapshotImmutability(t *testing.T) {
	x := []float64{1.0, 2.0, 3.0}
	y := []float64{2.0, 4.0, 6.0}
	steps := Trace(x, y, 0.05, 3)

	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	// Verify deep copy: mutating cells in one step doesn't affect subsequent steps
	for i := 0; i < len(steps)-1; i++ {
		tsCurrent, okCurrent := steps[i].State.(renderers.TableState)
		tsNext, okNext := steps[i+1].State.(renderers.TableState)
		if !okCurrent || !okNext {
			t.Fatalf("step state is not renderers.TableState: current=%T, next=%T", steps[i].State, steps[i+1].State)
		}

		if len(tsCurrent.Cells) > 0 && len(tsNext.Cells) > 0 {
			// Backing array pointers must not match
			if &tsCurrent.Cells[0] == &tsNext.Cells[0] {
				t.Fatalf("steps %d and %d share Cells slice backing array", i, i+1)
			}
			// RowHeaders backing pointers must not match
			if len(tsCurrent.RowHeaders) > 0 && len(tsNext.RowHeaders) > 0 {
				if &tsCurrent.RowHeaders[0] == &tsNext.RowHeaders[0] {
					t.Fatalf("steps %d and %d share RowHeaders backing array", i, i+1)
				}
			}
			// ColHeaders backing pointers must not match
			if len(tsCurrent.ColHeaders) > 0 && len(tsNext.ColHeaders) > 0 {
				if &tsCurrent.ColHeaders[0] == &tsNext.ColHeaders[0] {
					t.Fatalf("steps %d and %d share ColHeaders backing array", i, i+1)
				}
			}
		}
	}

	// Mutation test
	ts0 := steps[0].State.(renderers.TableState)
	ts1 := steps[1].State.(renderers.TableState)
	origVal1 := ts1.Cells[0].Value
	ts0.Cells[0].Value = "MUTATED_VALUE"

	if ts1.Cells[0].Value != origVal1 {
		t.Errorf("step 1 Cells mutated after modifying step 0: got %q, want %q", ts1.Cells[0].Value, origVal1)
	}
}

func TestStrictInputValidation(t *testing.T) {
	lesson := Lesson{}

	// nil input should use default
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("expected nil input to succeed, got %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps for nil input")
	}

	// empty raw message should use default
	steps, err = lesson.Build(json.RawMessage(""))
	if err != nil {
		t.Fatalf("expected empty input to succeed, got %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps for empty input")
	}

	tests := []struct {
		name  string
		input string
	}{
		{"null input", `null`},
		{"malformed JSON syntax", `{"x": [1, 2], "y": [3`},
		{"not an object", `[1, 2, 3]`},
		{"empty x and y", `{"x": [], "y": []}`},
		{"empty x only", `{"x": [], "y": [1.0]}`},
		{"empty y only", `{"x": [1.0], "y": []}`},
		{"null x array", `{"x": null, "y": [1.0]}`},
		{"null y array", `{"x": [1.0], "y": null}`},
		{"mismatched lengths", `{"x": [1.0, 2.0], "y": [3.0]}`},
		{"zero learning rate", `{"x": [1.0], "y": [2.0], "lr": 0.0}`},
		{"negative learning rate", `{"x": [1.0], "y": [2.0], "lr": -0.05}`},
		{"zero epochs", `{"x": [1.0], "y": [2.0], "epochs": 0}`},
		{"negative epochs", `{"x": [1.0], "y": [2.0], "epochs": -5}`},
		{"trailing data", `{"x": [1.0], "y": [2.0]} extra`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Fatalf("expected error for %s (%s), but got %d steps", tc.name, tc.input, len(s))
			}
		})
	}

	// Valid custom input
	validJSON := `{"x": [1.0, 2.0, 3.0], "y": [2.0, 4.0, 6.0], "lr": 0.01, "epochs": 3}`
	customSteps, err := lesson.Build(json.RawMessage(validJSON))
	if err != nil {
		t.Fatalf("expected valid custom input to succeed, got %v", err)
	}
	if len(customSteps) == 0 {
		t.Fatal("expected non-empty steps for valid custom input")
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
		t.Fatalf("ValidateSourceLineIDs failed for default input: %v", err)
	}

	customJSON := `{"x": [1.0, 2.0, 3.0], "y": [2.0, 4.0, 6.0], "lr": 0.01, "epochs": 2}`
	customSteps, err := lesson.Build(json.RawMessage(customJSON))
	if err != nil {
		t.Fatalf("Build failed for custom: %v", err)
	}
	if err := catalog.ValidateSourceLineIDs(lesson, customSteps); err != nil {
		t.Fatalf("ValidateSourceLineIDs failed for custom input: %v", err)
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
		if _, ok := vars["dw"]; !ok {
			t.Errorf("step %d missing 'dw' variable", i)
		}
		if _, ok := vars["db"]; !ok {
			t.Errorf("step %d missing 'db' variable", i)
		}

		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
		}
		if len(tableState.ColHeaders) != 4 {
			t.Errorf("step %d TableState ColHeaders length %d, want 4", i, len(tableState.ColHeaders))
		}
		if len(tableState.RowHeaders) != len(x) {
			t.Errorf("step %d TableState RowHeaders length %d, want %d", i, len(tableState.RowHeaders), len(x))
		}
		if len(tableState.Cells) != len(x)*4 {
			t.Errorf("step %d TableState Cells length %d, want %d", i, len(tableState.Cells), len(x)*4)
		}
		if tableState.Caption == "" {
			t.Errorf("step %d TableState Caption is empty", i)
		}
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
