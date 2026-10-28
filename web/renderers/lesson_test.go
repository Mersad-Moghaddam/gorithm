package renderers_test

import (
	"encoding/xml"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bubble"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestStepHoldMS(t *testing.T) {
	for _, tt := range []struct {
		hold  int
		speed float64
		want  int
	}{
		{0, 1, 1800}, {600, 2, 300}, {1200, .5, 2400},
		{350, 1, 350}, {1800, 1, 1800}, {600, 0, 600},
		{600, -1, 600}, {600, math.NaN(), 600}, {600, math.Inf(1), 600},
		{999, 1, 1800}, {600, 100, 100}, {350, 3, 117},
	} {
		if got := renderers.StepHoldMS(trace.Step{HoldMS: tt.hold}, tt.speed); got != tt.want {
			t.Errorf("hold %d speed %g: got %d want %d", tt.hold, tt.speed, got, tt.want)
		}
	}
}

func TestLessonMotionMarkup(t *testing.T) {
	steps, err := (bubble.Lesson{}).Build([]byte(`[2,2,1]`))
	if err != nil {
		t.Fatal(err)
	}
	colors := map[string]string{}
	for index, step := range steps {
		fragment, ok := renderers.RenderLessonSVG("bubble", step, 800, 450)
		if !ok {
			t.Fatal("Bubble pilot must use dedicated dispatch")
		}
		decoder := xml.NewDecoder(strings.NewReader("<svg>" + fragment + "</svg>"))
		seen := map[string]bool{}
		var order []string
		entity, checks := "", 0
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("step %d: %v", index, err)
			}
			start, ok := token.(xml.StartElement)
			if !ok {
				continue
			}
			attrs := map[string]string{}
			for _, attr := range start.Attr {
				attrs[attr.Name.Local] = attr.Value
				switch attr.Name.Local {
				case "data-x", "data-y", "data-scene-width", "data-scene-height", "data-min-width", "x", "y", "width", "height":
					value, err := strconv.ParseFloat(attr.Value, 64)
					if err != nil || math.IsNaN(value) || math.IsInf(value, 0) || ((attr.Name.Local == "width" || attr.Name.Local == "height") && value < 0) {
						t.Fatalf("step %d: invalid geometry %s=%q", index, attr.Name.Local, attr.Value)
					}
				}
			}
			if id := attrs["data-entity"]; id != "" {
				if seen[id] {
					t.Fatalf("step %d: duplicate identity %s", index, id)
				}
				seen[id], entity = true, id
				order = append(order, id)
			}
			if start.Name.Local == "rect" && attrs["fill"] != "none" {
				if color, exists := colors[entity]; exists && color != attrs["fill"] {
					t.Fatalf("step %d: identity %s changed color", index, entity)
				}
				colors[entity] = attrs["fill"]
			}
			if start.Name.Local == "text" {
				var text string
				if err := decoder.DecodeElement(&text, &start); err != nil {
					t.Fatal(err)
				}
				if text == "✓" {
					checks++
				}
			}
		}
		if len(seen) != 3 {
			t.Fatalf("step %d: want three original identities, got %v", index, order)
		}
		state := step.State.(renderers.BubbleLessonState)
		for slot, id := range state.IDs {
			if order[slot] != "main:"+strconv.Itoa(id) {
				t.Fatalf("step %d: rendered identity differs from snapshot", index)
			}
		}
		if index == len(steps)-1 && (strings.Join(order, ",") != "main:2,main:0,main:1" || checks != 3 || !strings.Contains(fragment, `data-phase="done"`)) {
			t.Fatalf("final duplicate result not fully visible: %v, %d checks", order, checks)
		}
	}
	if colors["main:0"] == colors["main:1"] || len(renderers.LessonLegend("bubble")) != 3 {
		t.Fatal("duplicate values need distinct identity colors and Bubble legend")
	}
}

func TestAllLessonsHaveDedicatedDispatch(t *testing.T) {
	ids := []string{
		"astar", "avl", "bellmanford", "bfs", "binarysearch", "bst", "bubble",
		"bucketsort", "coinchange", "countingsort", "dfs", "dijkstra", "fibonacci",
		"floydwarshall", "genetic", "gradientdescent", "hanoitower", "heap",
		"insertion", "kmeans", "kmp", "knapsack", "kosaraju", "kruskal", "lcs",
		"linearsearch", "merge", "minimax", "nqueens", "perceptron", "prim",
		"qlearning", "quick", "radixsort", "segmenttree", "selection", "sudoku",
		"toposort", "trie", "unionfind",
	}
	for _, id := range ids {
		if svg, ok := renderers.RenderLessonSVG(id, trace.Step{}, 800, 450); !ok || svg == "" {
			t.Errorf("%s: missing dedicated renderer dispatch", id)
		}
		if legend := renderers.LessonLegend(id); len(legend) < 3 {
			t.Errorf("%s: want at least three legend entries, got %d", id, len(legend))
		}
	}
}

func TestLessonDispatchFallback(t *testing.T) {
	if svg, ok := renderers.RenderLessonSVG("not-converted", trace.Step{}, 800, 450); ok || svg != "" {
		t.Fatal("unconverted lesson must retain generic renderer")
	}
	if len(renderers.LessonLegend("not-converted")) != 0 {
		t.Fatal("unconverted legend must fall back")
	}
}
