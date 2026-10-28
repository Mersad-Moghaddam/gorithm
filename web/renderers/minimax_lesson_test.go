package renderers

import (
	"encoding/xml"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strconv"
	"strings"
	"testing"
)

func TestMinimaxLessonSVG(t *testing.T) {
	s := MinimaxLessonState{Nodes: []MinimaxNodeView{{ID: "r", IsMax: true}, {ID: "c", ParentID: "r", Depth: 1, Terminal: true, Value: 9, Pruned: true}}, ActiveID: "r", Alpha: 3, Beta: 2, Cutoff: true, Phase: "prune"}
	for _, w := range []int{0, 320, 390, 768, 1440} {
		svg := RenderMinimaxLessonSVG(trace.Step{State: s, Explanation: `skip <unsafe> & branch`}, w, 640)
		for _, want := range []string{`data-lesson="minimax"`, `data-phase="prune"`, `data-scene-width=`, `data-scene-height=`, `data-min-width=`, `data-entity="node:r"`, `not evaluated`, `β = 2 ≤ α = 3`, `&lt;unsafe&gt;`, `?`, `MAX`, `stroke-dasharray`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %q", want)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
			t.Fatal("nonfinite")
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
}

func TestMinimaxLessonWideGeometry(t *testing.T) {
	s := MinimaxLessonState{Phase: "init", Nodes: []MinimaxNodeView{{ID: "r", IsMax: true}}}
	for i := 0; i < 64; i++ {
		s.Nodes = append(s.Nodes, MinimaxNodeView{ID: fmt.Sprintf("leaf:%d", i), ParentID: "r", Depth: 1, Terminal: true})
	}
	svg := RenderMinimaxLessonSVG(trace.Step{State: s}, 320, 640)
	dec := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	seen := map[string]bool{}
	xs := map[string]bool{}
	for {
		token, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		element, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		attrs := map[string]string{}
		for _, a := range element.Attr {
			attrs[a.Name.Local] = a.Value
		}
		if v := attrs["data-min-width"]; v != "" {
			n, _ := strconv.Atoi(v)
			if n < 64*100 {
				t.Fatal("wide tree shrunk")
			}
		}
		if id := attrs["data-entity"]; id != "" {
			if seen[id] {
				t.Fatal("duplicate identity")
			}
			seen[id] = true
			if strings.HasPrefix(id, "node:leaf:") {
				if xs[attrs["data-x"]] {
					t.Fatal("overlapping leaves")
				}
				xs[attrs["data-x"]] = true
			}
		}
	}
	if len(xs) != 64 {
		t.Fatal("missing leaves")
	}
}

func TestMinimaxLessonSVGReturn(t *testing.T) {
	s := MinimaxLessonState{Nodes: []MinimaxNodeView{{ID: "r", IsMax: true}, {ID: "c", ParentID: "r", Depth: 1, Terminal: true, Value: 3, Evaluated: true}}, ActiveID: "r", ReturnedFromID: "c", Eval: 3, EvalValid: true, BestValue: 3, BestValueValid: true, Alpha: 3, Beta: 2147483647, Phase: "child-return"}
	svg := RenderMinimaxLessonSVG(trace.Step{State: s, LineID: "L9"}, 768, 640)
	for _, want := range []string{`data-entity="utility:c"`, `data-from-entity="node:c"`, `data-motion="linear"`, `max(`, `+∞`} {
		if !strings.Contains(svg, want) {
			t.Errorf("missing %s", want)
		}
	}
	if minimaxBound(1000000) != "1000000" {
		t.Fatal("false infinity")
	}
}

func TestMinimaxReturnBadgeDoesNotCoverParent(t *testing.T) {
	s := MinimaxLessonState{Nodes: []MinimaxNodeView{{ID: "r", IsMax: true}, {ID: "c", ParentID: "r", Depth: 1, Terminal: true}}, ActiveID: "r", ReturnedFromID: "c", Eval: 3, EvalValid: true, Phase: "child-return"}
	svg := RenderMinimaxLessonSVG(trace.Step{State: s}, 320, 640)
	if !strings.Contains(svg, `data-from-entity="node:c" data-motion="linear" data-x="400.0" data-y="70.0"`) {
		t.Fatal("return badge must travel from child to space adjacent to parent")
	}
}

func TestMinimaxReturnBadgeConnectsToReceivingParent(t *testing.T) {
	s := MinimaxLessonState{Nodes: []MinimaxNodeView{{ID: "r", IsMax: true}, {ID: "c", ParentID: "r", Depth: 1, Terminal: true}}, ActiveID: "r", ReturnedFromID: "c", Eval: 3, EvalValid: true, Phase: "child-return"}
	svg := RenderMinimaxLessonSVG(trace.Step{State: s}, 320, 640)
	if !strings.Contains(svg, `data-return-to="node:r"`) || !strings.Contains(svg, `M364.0 70.0L368.0 70.0`) {
		t.Fatal("return badge lacks connector to receiving parent")
	}
}
