package renderers

import (
	"encoding/xml"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestFloydWarshallRenderer(t *testing.T) {
	s := FloydWarshallLessonState{Nodes: []string{`<script onload="x">&界`, "B"}, Distance: [][]int64{{0, 3}, {math.MaxInt64, 0}}, I: 0, J: 1, K: 0, LeftDistance: 0, RightDistance: 3, OldDistance: 3, Candidate: 3, HasCandidate: true, DiagonalIndex: -1}
	for _, w := range []int{320, 960} {
		step := trace.Step{LineID: "L7", State: s, Explanation: `<unsafe & "quote">`}
		out := RenderFloydWarshallLessonSVG(step, w, 640)
		if out != RenderFloydWarshallLessonSVG(step, w, 640) {
			t.Fatal("nondeterministic geometry")
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + out + "</svg>"))
		entities := map[string]bool{}
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			start, ok := token.(xml.StartElement)
			if !ok {
				continue
			}
			if start.Name.Local == "script" {
				t.Fatal("injected element")
			}
			for _, a := range start.Attr {
				if strings.HasPrefix(a.Name.Local, "on") {
					t.Fatal("injected event")
				}
				if a.Name.Local == "data-entity" {
					if entities[a.Value] {
						t.Fatal("duplicate identity")
					}
					entities[a.Value] = true
				}
			}
		}
		for _, want := range []string{`data-lesson="floydwarshall"`, `data-phase="L7"`, `data-scene-width=`, `data-min-width=`, `data-role="left"`, `data-role="dest"`, `data-from-entity="floydwarshall:cell:0:0"`, `min(3, 0 + 3) = 3`, `∞`, `&lt;script`} {
			if !strings.Contains(out, want) {
				t.Errorf("missing %s", want)
			}
		}
		if strings.Contains(out, "NaN") || strings.Contains(out, "+Inf") {
			t.Fatal("nonfinite geometry")
		}
	}
	s.NegativeCycle = true
	s.CheckingDiagonal = true
	s.HasCandidate = false
	out := RenderFloydWarshallLessonSVG(trace.Step{State: s, LineID: "L16"}, 320, 640)
	if !strings.Contains(out, "finite optimum not defined") {
		t.Fatal("missing negative cycle outcome")
	}
	s.NegativeCycle = false
	s.HasCandidate = true
	s.Improved = true
	out = RenderFloydWarshallLessonSVG(trace.Step{State: s, LineID: "L8"}, 320, 640)
	if !strings.Contains(out, `data-from-entity="floydwarshall:operand:left"`) || !strings.Contains(out, "Commit candidate") {
		t.Fatal("missing commit motion")
	}
}

func TestFloydWarshallWideDistanceCells(t *testing.T) {
	s := FloydWarshallLessonState{Nodes: []string{"A", "B"}, Distance: [][]int64{{-3134566563840000, -3134566563840000}, {0, 0}}, I: -1, J: -1, K: -1}
	for _, width := range []int{320, 960} {
		out := RenderFloydWarshallLessonSVG(trace.Step{State: s}, width, 640)
		match := regexp.MustCompile(`data-entity="floydwarshall:cell:0:[01]" data-x="(\d+)"`).FindAllStringSubmatch(out, -1)
		if len(match) != 2 {
			t.Fatal("missing cells")
		}
		x1, _ := strconv.Atoi(match[0][1])
		x2, _ := strconv.Atoi(match[1][1])
		if x2-x1 < len("-3134566563840000")*9+16 {
			t.Fatalf("width %d: long values only %d units apart", width, x2-x1)
		}
	}
}

func TestFloydWarshallLongOperandLayout(t *testing.T) {
	value := int64(-3134566563840000)
	s := FloydWarshallLessonState{Nodes: []string{"A", "B"}, Distance: [][]int64{{value, value}, {value, value}}, I: 0, J: 1, K: 0, LeftDistance: value, RightDistance: value, OldDistance: value, Candidate: value, HasCandidate: true}
	for _, width := range []int{320, 960} {
		out := RenderFloydWarshallLessonSVG(trace.Step{State: s, LineID: "L7"}, width, 640)
		if !strings.Contains(out, `data-operand-layout="stacked"`) {
			t.Errorf("width %d: operands not stacked", width)
		}
		s.Improved = true
		out = RenderFloydWarshallLessonSVG(trace.Step{State: s, LineID: "L8"}, width, 640)
		if !strings.Contains(out, `data-commit-marker="check"`) {
			t.Errorf("width %d: long value used as commit token", width)
		}
		s.Improved = false
	}
}
