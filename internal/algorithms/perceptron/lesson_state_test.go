package perceptron

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestPerceptronLessonThresholdAndCorrection(t *testing.T) {
	inputs := [][]float64{{0, 0}, {0, 1}}
	steps := Trace(inputs, []int{0, 1}, .1, 1)
	if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(steps, Trace(inputs, []int{0, 1}, .1, 1)) {
		t.Fatal("nondeterministic trace")
	}
	seen := map[string]int{}
	for _, step := range steps {
		s := step.State.(renderers.PerceptronLessonState)
		seen[step.LineID]++
		if step.HoldMS != 0 && step.HoldMS != 350 && step.HoldMS != 600 && step.HoldMS != 1200 && step.HoldMS != 1800 {
			t.Fatal("invalid hold")
		}
		if step.LineID == "L8" && step.HoldMS+steps[step.Index+1].HoldMS != 1800 {
			t.Fatal("threshold/error pair budget")
		}
		if step.LineID == "L7" && s.PredictionValid {
			t.Fatal("prediction before threshold")
		}
		if step.LineID == "L8" && (!s.PredictionValid || s.Prediction != 1-s.Sample) {
			t.Fatalf("threshold: %+v", s)
		}
		if step.LineID == "L12" {
			if s.Bias != s.BeforeBias || s.Error != 2*s.Sample-1 || s.Prediction != 1-s.Sample {
				t.Fatalf("weight correction: %+v", s)
			}
			if s.Weights[0] != 0 || s.Weights[1] != float64(s.Sample)*.1 {
				t.Fatalf("weights: %v", s.Weights)
			}
		}
		if step.LineID == "L13" && s.Bias != -.1+float64(s.Sample)*.1 {
			t.Fatalf("bias: %+v", s)
		}
	}
	if seen["L12"] != 2 || seen["L13"] != 2 {
		t.Fatal(seen)
	}
	last := steps[len(steps)-1].State.(renderers.PerceptronLessonState)
	if last.Converged || last.StopReason != "epoch-limit" || last.Epoch != 1 {
		t.Fatalf("stop: %+v", last)
	}
	last.Inputs[0][0] = 99
	last.Weights[0] = 99
	last.Labels[0] = 1
	first := steps[0].State.(renderers.PerceptronLessonState)
	if first.Inputs[0][0] != 0 || first.Weights[0] != 0 || first.Labels[0] != 0 || inputs[0][0] != 0 {
		t.Fatal("aliased snapshots")
	}
}

func TestPerceptronLessonLimitsAndDimensions(t *testing.T) {
	for _, n := range []int{64, 65} {
		for _, d := range []int{8, 9} {
			input := PerceptronInput{Inputs: make([][]float64, n), Labels: make([]int, n), LR: 1, Epochs: 100}
			for i := range input.Inputs {
				input.Inputs[i] = make([]float64, d)
				input.Labels[i] = 1
			}
			raw, _ := json.Marshal(input)
			_, err := (Lesson{}).Build(raw)
			if (err == nil) != (n == 64 && d == 8) {
				t.Fatalf("bounds n=%d d=%d: %v", n, d, err)
			}
		}
	}
	for _, raw := range []string{`null`, `{`, `{"inputs":null}`, `{"inputs":[[1],[1,2]],"labels":[0,1]}`, `{"inputs":[[1]],"labels":[]}`, `{"inputs":[[]],"labels":[0]}`, `{"inputs":[[1]],"labels":[2]}`, `{"lr":0}`, `{"lr":1.1}`, `{"epochs":0}`, `{"epochs":101}`, `{"lr":null}`, `{"epochs":null}`, `{"labels":null}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, d := range []int{1, 2, 3, 8} {
		steps := Trace([][]float64{make([]float64, d)}, []int{1}, 1, 100)
		last := steps[len(steps)-1].State.(renderers.PerceptronLessonState)
		if !last.Converged || last.Epoch != 1 || last.StopReason != "stable" {
			t.Fatalf("stable: %+v", last)
		}
		w, b := Train([][]float64{make([]float64, d)}, []int{1}, 1, 100)
		if !reflect.DeepEqual(w, last.Weights) || b != last.Bias {
			t.Fatal("Train mismatch")
		}
	}
	xor := Trace([][]float64{{0, 0}, {0, 1}, {1, 0}, {1, 1}}, []int{0, 1, 1, 0}, .1, 2)
	end := xor[len(xor)-1].State.(renderers.PerceptronLessonState)
	if end.Converged || end.StopReason != "epoch-limit" || end.Epoch != 2 {
		t.Fatalf("XOR: %+v", end)
	}
	w, b := Train([][]float64{{0, 0}, {0, 1}, {1, 0}, {1, 1}}, []int{0, 1, 1, 0}, .1, 2)
	if !reflect.DeepEqual(w, end.Weights) || b != end.Bias {
		t.Fatal("XOR Train mismatch")
	}
	for _, inputs := range [][][]float64{nil, {{1}, {1, 2}}, {{math.NaN()}}, {{math.Inf(1)}}} {
		steps := Trace(inputs, []int{0}, .1, 1)
		if len(steps) != 1 || steps[0].State.(renderers.PerceptronLessonState).StopReason != "invalid-input" {
			t.Fatal("invalid direct trace")
		}
	}
	overflow := Trace([][]float64{{math.MaxFloat64}, {math.MaxFloat64}}, []int{0, 1}, 1, 2)
	if overflow[len(overflow)-1].State.(renderers.PerceptronLessonState).StopReason != "numerical-error" {
		t.Fatal("overflow not stopped")
	}
}
