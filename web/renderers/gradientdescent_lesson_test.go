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

func TestGradientDescentDuplicateSamplesHighlightActiveWithoutOverlappingLabels(t *testing.T) {
	s := GradientDescentLessonState{X: []float64{1, 1, 1}, Y: []float64{2, 2, 2}, Sample: 0, Phase: "residual", DataXMin: 0, DataXMax: 2, DataYMin: 0, DataYMax: 3, SliceWMin: -1, SliceWMax: 1, SliceLossMax: 10, HistoryLossMax: 10}
	svg := RenderGradientDescentLessonSVG(trace.Step{State: s}, 320, 640)
	if strings.Count(svg, `>P0</text>`) != 1 || strings.Contains(svg, `>P1</text>`) || strings.Contains(svg, `>P2</text>`) {
		t.Fatal("only the active sample should have a plot label")
	}
	if strings.LastIndex(svg, `data-entity="sample:P0"`) < strings.LastIndex(svg, `data-entity="sample:P2"`) {
		t.Fatal("active sample must be drawn on top of coincident samples")
	}
}

func TestGradientDescentActiveLabelStaysInsidePlot(t *testing.T) {
	s := GradientDescentLessonState{X: []float64{1}, Y: []float64{1}, Sample: 0, Phase: "residual", DataXMin: 0, DataXMax: 1, DataYMin: 0, DataYMax: 1, SliceWMin: -1, SliceWMax: 1, SliceLossMax: 1, HistoryLossMax: 1}
	svg := RenderGradientDescentLessonSVG(trace.Step{State: s}, 320, 640)
	if !strings.Contains(svg, `<text x="-8" y="20" text-anchor="end">P0</text>`) {
		t.Fatal("top-right point label must sit inside the plot")
	}
}
