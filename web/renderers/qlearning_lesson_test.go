package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestQLearningLessonSVG(t *testing.T) {
	s := QLearningLessonState{Rows: 3, Cols: 3, Start: [2]int{0, 0}, Goal: [2]int{2, 2}, State: [2]int{2, 1}, NextState: [2]int{2, 2}, Obstacles: [][2]int{{1, 1}}, Q: make([][]float64, 9), Action: 1, BestNextAction: -1, Alpha: .5, Gamma: .9, Epsilon: .3, Reward: 10, TDTarget: 10, NewQ: 5, TransitionValid: true, UpdateValid: true, Terminal: true, Phase: "update"}
	for i := range s.Q {
		s.Q[i] = make([]float64, 4)
	}
	for _, w := range []int{0, 320, 390, 768, 1440} {
		svg := RenderQLearningLessonSVG(trace.Step{State: s}, w, 650)
		for _, want := range []string{`data-lesson="qlearning"`, `data-entity="agent:0"`, `current Q estimates`, `target = 10`, `new Q = 0 + 0.5`, `blocked`, `candidate Up`, `candidate Left`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "<svg") {
			t.Fatal("invalid fragment")
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		for {
			_, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	s.Phase = `<script>&`
	svg := RenderQLearningLessonSVG(trace.Step{State: s}, 320, 650)
	if strings.Contains(svg, "<script>") {
		t.Fatal("unescaped phase")
	}
}

func TestQLearningAssignmentDoesNotSelectDestinationAction(t *testing.T) {
	s := QLearningLessonState{Rows: 1, Cols: 2, Goal: [2]int{0, 1}, State: [2]int{0, 1}, NextState: [2]int{0, 1}, Q: [][]float64{{0, 0, 0, 0}, {0, 0, 0, 0}}, Action: 1, Phase: "assign-state", TransitionValid: true, UpdateValid: true}
	svg := RenderQLearningLessonSVG(trace.Step{State: s}, 320, 650)
	if strings.Contains(svg, `stroke="#ffe29a"`) || strings.Contains(svg, `fill="#ffe29a"`) {
		t.Fatal("previous action highlighted at destination")
	}
}

func TestQLearningResetDoesNotAnimateAcrossGrid(t *testing.T) {
	s := QLearningLessonState{Rows: 1, Cols: 2, Goal: [2]int{0, 1}, State: [2]int{0, 0}, NextState: [2]int{0, 0}, Action: -1, Phase: "reset"}
	svg := RenderQLearningLessonSVG(trace.Step{State: s}, 320, 650)
	if strings.Contains(svg, `data-entity="agent:0"`) {
		t.Fatal("episode reset must not animate agent from previous goal")
	}
}

func TestQLearningUncalculatedValuesAreNotPresentedAsResults(t *testing.T) {
	s := QLearningLessonState{Rows: 1, Cols: 2, Goal: [2]int{0, 1}, Q: [][]float64{{0, 0, 0, 0}, {0, 0, 0, 0}}, Action: -1, Phase: "initialize"}
	svg := RenderQLearningLessonSVG(trace.Step{State: s}, 320, 650)
	if strings.Contains(svg, "Next row (0,0)") || strings.Contains(svg, "reward 0 · old 0 · new 0") {
		t.Fatal("uncalculated transition presented as result")
	}
}
