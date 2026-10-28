package renderers_test

import (
	"encoding/xml"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/linearsearch"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestLinearSearchLessonSVG(t *testing.T) {
	for _, arr := range [][]int{{-2, 0, 7, 7}, nil} {
		for _, step := range linearsearch.Trace(arr, 7) {
			before := step.State.(renderers.LinearSearchLessonState)
			values := append([]int(nil), before.Values...)
			for _, size := range [][2]int{{320, 420}, {800, 480}, {0, -1}} {
				svg := renderers.RenderLinearSearchLessonSVG(step, size[0], size[1])
				if svg != renderers.RenderLinearSearchLessonSVG(step, size[0], size[1]) {
					t.Fatal("nondeterministic")
				}
				if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
					t.Fatal(svg)
				}
				d := xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">` + svg + `</svg>`))
				ids := map[string]bool{}
				tiles := 0
				for {
					token, err := d.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					e, ok := token.(xml.StartElement)
					if !ok {
						continue
					}
					attrs := map[string]string{}
					for _, a := range e.Attr {
						attrs[a.Name.Local] = a.Value
					}
					if id := attrs["data-entity"]; id != "" {
						if ids[id] {
							t.Fatalf("duplicate %s", id)
						}
						ids[id] = true
						if attrs["data-x"] == "" || attrs["data-y"] == "" || attrs["transform"] == "" {
							t.Fatal("missing motion coordinates")
						}
					}
					if attrs["data-role"] == "tile" {
						tiles++
						if attrs["width"] != "64" || attrs["height"] != "56" {
							t.Fatal("magnitude-scaled tile")
						}
					}
				}
				if tiles != len(arr) {
					t.Fatalf("tiles %d", tiles)
				}
				if !strings.Contains(svg, `data-lesson="linearsearch"`) || !strings.Contains(svg, `data-min-width="800"`) {
					t.Fatal("missing scene contract")
				}
				if before.Phase == "compare" && !strings.Contains(svg, " = ") && !strings.Contains(svg, " ≠ ") {
					t.Fatal("missing equation")
				}
				if before.ResultKnown && before.Result >= 0 && !strings.Contains(svg, "found") {
					t.Fatal("missing found cue")
				}
				if strings.Count(svg, ">checked</text>") != before.CheckedThrough+1 {
					t.Fatal("checked prefix includes unvisited suffix")
				}
				if before.CheckedThrough >= 0 && !strings.Contains(svg, "≠ target") {
					t.Fatal("checked must not imply found")
				}
				if len(arr) == 0 && (strings.Contains(svg, `data-entity="linearsearch-pointer"`) || !strings.Contains(svg, "Empty array")) {
					t.Fatal("empty pointer")
				}
			}
			if !reflect.DeepEqual(values, append([]int(nil), before.Values...)) {
				t.Fatal("render mutation")
			}
		}
	}
	for _, step := range []trace.Step{{}, {State: "wrong"}} {
		if !strings.Contains(renderers.RenderLinearSearchLessonSVG(step, 320, 420), "No visual state") {
			t.Fatal("missing fallback")
		}
	}
}

func TestLinearSearchLessonPointerMotionAndEscaping(t *testing.T) {
	steps := linearsearch.Trace([]int{7, 7, 7}, 9)
	var previous string
	for _, step := range steps {
		s := step.State.(renderers.LinearSearchLessonState)
		if s.Phase != "scan" {
			continue
		}
		svg := renderers.RenderLinearSearchLessonSVG(step, 320, 420)
		start := strings.Index(svg, `data-entity="linearsearch-pointer"`)
		if start < 0 {
			t.Fatal("pointer missing")
		}
		end := strings.Index(svg[start:], ">")
		pointer := svg[start : start+end]
		if pointer == previous {
			t.Fatal("pointer did not move")
		}
		previous = pointer
		for _, id := range []string{"linearsearch-item-0", "linearsearch-item-1", "linearsearch-item-2"} {
			if !strings.Contains(svg, `data-entity="`+id+`"`) {
				t.Fatal("identity missing")
			}
		}
	}
	s := steps[0].State.(renderers.LinearSearchLessonState)
	s.Phase = `<>&"`
	svg := renderers.RenderLinearSearchLessonSVG(trace.Step{State: s}, 320, 420)
	if strings.Contains(svg, `data-phase="<`) || !strings.Contains(svg, "&lt;&gt;&amp;&#34;") {
		t.Fatal("unescaped phase")
	}
}
