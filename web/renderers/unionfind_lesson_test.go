package renderers_test

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/unionfind"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"io"
	"strings"
	"testing"
)

func TestUnionFindLessonForest(t *testing.T) {
	steps, _ := (unionfind.Lesson{}).Build(nil)
	for _, step := range steps {
		state := step.State.(renderers.UnionFindLessonState)
		before, _ := json.Marshal(state)
		for _, size := range [][2]int{{320, 420}, {800, 480}, {0, 0}, {-1, -2}} {
			svg := renderers.RenderUnionFindLessonSVG(step, size[0], size[1])
			if svg != renderers.RenderUnionFindLessonSVG(step, size[0], size[1]) {
				t.Fatal("nondeterministic")
			}
			if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
				t.Fatal("nonfinite SVG")
			}
			ids, entities, edges := map[string]bool{}, map[string]bool{}, map[string]string{}
			decoder := xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">` + svg + `</svg>`))
			for {
				tok, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				start, ok := tok.(xml.StartElement)
				if !ok {
					continue
				}
				attrs := map[string]string{}
				for _, a := range start.Attr {
					attrs[a.Name.Local] = a.Value
				}
				if id := attrs["id"]; id != "" {
					if ids[id] {
						t.Fatal("duplicate ID")
					}
					ids[id] = true
				}
				if id := attrs["data-entity"]; id != "" {
					if entities[id] {
						t.Fatal("duplicate entity")
					}
					entities[id] = true
					if attrs["data-x"] == "" || attrs["data-y"] == "" || attrs["transform"] == "" {
						t.Fatal("missing motion geometry")
					}
				}
				if attrs["data-edge"] == "parent" {
					edges[attrs["data-child"]] = attrs["data-parent"]
					if attrs["marker-end"] == "" {
						t.Fatal("undirected edge")
					}
				}
				for _, key := range []string{"width", "height", "r", "data-scene-width", "data-scene-height"} {
					if strings.HasPrefix(attrs[key], "-") {
						t.Fatal("negative dimension")
					}
				}
			}
			if size[0] <= 0 {
				continue
			}
			for i, p := range state.Parent {
				id := fmt.Sprintf("unionfind-node-%d", i)
				if !ids[id] {
					t.Fatal("missing node")
				}
				if p != i {
					if edges[id] != fmt.Sprintf("unionfind-node-%d", p) {
						t.Fatalf("wrong parent edge %d → %d: %v", i, p, edges)
					}
				} else if edges[id] != "" {
					t.Fatal("root self-loop")
				}
			}
			if !strings.Contains(svg, ">rank<") || !strings.Contains(svg, ">parent<") || !strings.Contains(svg, "Components:") || !strings.Contains(svg, `data-min-width="800"`) {
				t.Fatal("missing table/scene")
			}
			if state.Phase == "compress" && !strings.Contains(svg, `stroke-dasharray="6 5"`) {
				t.Fatal("missing old-parent witness")
			}
		}
		after, _ := json.Marshal(state)
		if string(before) != string(after) {
			t.Fatal("render mutated state")
		}
	}
}

func TestUnionFindLessonFallbackAndEscaping(t *testing.T) {
	for _, state := range []any{nil, "bad", renderers.UnionFindLessonState{}, renderers.UnionFindLessonState{Parent: []int{1}, Rank: []int{0}}, renderers.UnionFindLessonState{Parent: []int{1, 0}, Rank: []int{0, 0}}} {
		if svg := renderers.RenderUnionFindLessonSVG(trace.Step{State: state}, 320, 420); !strings.Contains(svg, "No union-find state") {
			t.Fatal(svg)
		}
	}
	s := renderers.UnionFindLessonState{Parent: []int{0}, Rank: []int{0}, Phase: `<>&"日本`, Operation: `<script>`, Active: -1, ChangedNode: -1}
	svg := renderers.RenderUnionFindLessonSVG(trace.Step{State: s}, 800, 480)
	if strings.Contains(svg, "<script>") || !strings.Contains(svg, "&lt;") || !strings.Contains(svg, `fill="#c45d75"`) {
		t.Fatal("escaping/palette")
	}
	if err := xml.Unmarshal([]byte("<svg>"+svg+"</svg>"), new(any)); err != nil {
		t.Fatal(err)
	}
}
