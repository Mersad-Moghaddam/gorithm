package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestGradientDescentLessonSVG(t *testing.T) {
	s := GradientDescentLessonState{X: []float64{1, 2, 3, 4}, Y: []float64{3, 5, 7, 9}, Predictions: []float64{0, 0, 0, 0}, Residuals: []float64{-3, -5, -7, -9}, Phase: "weight", W: 1.75, BeforeW: 0, LR: .05, GradientW: -35, GradientB: -12, LossBefore: 41, LossBeforeValid: true, DataXMin: 0, DataXMax: 5, DataYMin: -1, DataYMax: 10, SliceWMin: -1, SliceWMax: 4, SliceLossMin: 0, SliceLossMax: 100, HistoryLossMin: 0, HistoryLossMax: 50, History: []GradientDescentEpoch{{Epoch: 0, Loss: 41}}}
	for _, width := range []int{0, 320, 390, 768, 1440} {
		svg := RenderGradientDescentLessonSVG(trace.Step{State: s, Explanation: "weight <update>"}, width, 640)
		for _, want := range []string{`data-lesson="gradientdescent"`, `data-phase="weight"`, `MSE versus w`, `b held at`, `Last measured MSE`, `Current diagnostic MSE`, `slice minimum; b fixed`, `data-motion="path"`, `data-entity="parameter:w"`, `data-path=`, `tangent`, `weight target`, `&lt;update&gt;`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "global minimum") {
			t.Fatal(svg)
		}
		dec := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		for {
			_, err := dec.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	s.X = []float64{0, 0}
	s.Y = []float64{2, 2}
	s.Phase = "bias"
	s.B = .4
	svg := RenderGradientDescentLessonSVG(trace.Step{State: s}, 320, 640)
	if !strings.Contains(svg, "no unique weight minimum") || strings.Contains(svg, `data-motion="path"`) {
		t.Fatal("flat/bias slice misleading")
	}
}
