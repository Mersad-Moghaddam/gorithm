package renderers

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestPerceptronLessonSVG(t *testing.T) {
	for _, d := range []int{1, 2, 3, 8} {
		s := PerceptronLessonState{Inputs: [][]float64{make([]float64, d), make([]float64, d)}, Labels: []int{0, 1}, Weights: make([]float64, d), BeforeWeights: make([]float64, d), Products: make([]float64, d), Sample: 0, Phase: "weights", PredictionValid: true, Prediction: 1, Error: -1, LR: .1, XMin: -1, XMax: 1, YMin: -1, YMax: 1}
		for _, width := range []int{-1, 320, 390, 768, 1440} {
			for _, vertical := range []bool{false, true} {
				if vertical && d == 2 {
					s.Weights[0] = 1
				}
				svg := RenderPerceptronLessonSVG(trace.Step{State: s, Explanation: `<script> & "unsafe"`}, width, 620)
				if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "<script>") {
					t.Fatal(svg)
				}
				if !strings.Contains(svg, `data-lesson="perceptron"`) || !strings.Contains(svg, "prediction before correction") {
					t.Fatal("missing lesson evidence")
				}
				if d == 2 && !strings.Contains(svg, "Decision boundary") {
					t.Fatal("missing plane")
				}
				if d != 2 && !strings.Contains(svg, "Feature vector") {
					t.Fatal("missing vector")
				}
				dec := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
				ids := map[string]bool{}
				for {
					tok, err := dec.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if el, ok := tok.(xml.StartElement); ok {
						for _, a := range el.Attr {
							if a.Name.Local == "data-entity" {
								if ids[a.Value] {
									t.Fatal("duplicate entity")
								}
								ids[a.Value] = true
							}
						}
					}
				}
				if !ids["sample:0"] || !ids["weight:0"] {
					t.Fatal(ids)
				}
			}
		}
	}
}

func TestPerceptronThresholdShowsPredictionBeforeError(t *testing.T) {
	s := PerceptronLessonState{Weights: []float64{1}, BeforeWeights: []float64{1}, Products: []float64{2}, Sample: 0, Inputs: [][]float64{{2}}, Phase: "threshold", PredictionValid: true, Prediction: 1, Target: 0}
	svg := RenderPerceptronLessonSVG(trace.Step{State: s}, 320, 620)
	if !strings.Contains(svg, "prediction 1 · target 0") {
		t.Error("threshold must show computed prediction")
	}
	if strings.Contains(svg, "error =") {
		t.Error("error is not computed until next step")
	}
}

func TestPerceptronIdleShowsCurrentBias(t *testing.T) {
	s := PerceptronLessonState{Weights: []float64{1}, BeforeWeights: []float64{1}, Bias: .2, BeforeBias: .1, Sample: -1, Phase: "done"}
	svg := RenderPerceptronLessonSVG(trace.Step{State: s}, 320, 620)
	if !strings.Contains(svg, "Σ products + bias 0.2") {
		t.Error("idle neuron must show current bias")
	}
}

func TestPerceptronCoincidentSamplesShareMarkerAndKeepDetails(t *testing.T) {
	s := PerceptronLessonState{Inputs: [][]float64{{.25, .5}, {.25, .5}, {1, 1}}, Labels: []int{0, 1, 1}, Weights: []float64{0, 0}, BeforeWeights: []float64{0, 0}, Sample: 1, XMin: 0, XMax: 1, YMin: 0, YMax: 1}
	for _, width := range []int{320, 800} {
		svg := RenderPerceptronLessonSVG(trace.Step{State: s}, width, 620)
		if strings.Count(svg, `data-entity="sample:`) != 2 {
			t.Errorf("width %d: coincident samples must share a marker", width)
		}
		position := `data-entity="sample:0" data-x="100.00" data-y="205.00"`
		if width == 800 {
			position = `data-entity="sample:0" data-x="124.00" data-y="205.00"`
		}
		if !strings.Contains(svg, position) {
			t.Errorf("width %d: marker moved away from true coordinates", width)
		}
		if !strings.Contains(svg, `M-7 0 L0 -7 L7 0 L0 7 Z`) {
			t.Errorf("width %d: mixed-class marker must not imply a single class", width)
		}
		for _, want := range []string{"Sample 0 · class 0", "Sample 1 · class 1 · active", "Sample 2 · class 1", "x0 = 0.25", "x1 = 0.5"} {
			if !strings.Contains(svg, want) {
				t.Errorf("width %d: missing %q", width, want)
			}
		}
	}
}
