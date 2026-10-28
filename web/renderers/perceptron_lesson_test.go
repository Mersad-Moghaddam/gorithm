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
