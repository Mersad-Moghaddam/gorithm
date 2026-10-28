package renderers

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestGeneticLessonShowsExactMutationRate(t *testing.T) {
	s := GeneticLessonState{Target: "1", MutationRate: 0.001}
	svg := RenderGeneticLessonSVG(trace.Step{State: s}, 320, 0)
	if !strings.Contains(svg, "Mutation rate 0.001") {
		t.Fatal("mutation rate rounded in scene")
	}
}

func TestGeneticLessonSVG(t *testing.T) {
	s := GeneticLessonState{Target: "<&\"'>", Phase: "mutation", Cut: 2, MutationRate: 1, Population: []GeneticIndividualView{{ID: "g0-i0", Genes: "1000", Fitness: 3}, {ID: "g0-i1", Genes: "0111", Fitness: 1}}, Parents: [2]GeneticIndividualView{{ID: "g0-i0", Genes: "1000"}, {ID: "g0-i1", Genes: "0111"}}, Crossed: [2]GeneticIndividualView{{ID: "g1-i1", Genes: "1011"}, {ID: "g1-i2", Genes: "0100"}}, Mutated: [2]GeneticIndividualView{{ID: "g1-i1", Genes: "1010", Fitness: 4}, {ID: "g1-i2", Genes: "0100", Fitness: 1}}, ChangedLoci: [2][]int{{3}, nil}, BestFitnessHistory: []int{3, 4}}
	for _, w := range []int{-1, 320, 390, 768, 1440} {
		svg := RenderGeneticLessonSVG(trace.Step{State: s}, w, 0)
		for _, want := range []string{`data-lesson="genetic"`, `data-scene-height=`, `data-entity="population:g0-i0"`, `data-from-entity="population:g0-i0"`, `data-origin="g0-i1"`, `3: 1 → 0`, `&lt;`, `&amp;`, `&#34;`, `fitness(g)`} {
			if !strings.Contains(svg, want) {
				t.Errorf("missing %q at %d", want, w)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "<svg") {
			t.Fatal("invalid fragment")
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		ids := map[string]bool{}
		for {
			tok, err := d.Token()
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
							t.Fatal("duplicate entity", a.Value)
						}
						ids[a.Value] = true
					}
				}
			}
		}
	}
}
