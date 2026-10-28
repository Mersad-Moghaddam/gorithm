package renderers_test

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kmp"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestKMPLessonSVG(t *testing.T) {
	for _, v := range [][2]string{{"ABABDABACDABABCABAB", "ABABCABAB"}, {"é猫é", "猫"}, {"x<&\"y", "<&\""}, {"abc", ""}, {"", "a"}} {
		for _, step := range kmp.Trace(v[0], v[1]) {
			for _, size := range [][2]int{{320, 420}, {800, 480}, {0, -1}} {
				s := step.State.(renderers.KMPLessonState)
				saved := append([]int(nil), s.LPS...)
				svg := renderers.RenderKMPLessonSVG(step, size[0], size[1])
				if svg != renderers.RenderKMPLessonSVG(step, size[0], size[1]) || !reflect.DeepEqual(saved, s.LPS) {
					t.Fatal("mutation/nondeterminism")
				}
				if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || !strings.Contains(svg, `data-lesson="kmp"`) || !strings.Contains(svg, "byte offsets") {
					t.Fatal(svg)
				}
				dec := xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">` + svg + `</svg>`))
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
							if a.Name.Local == "id" || a.Name.Local == "data-entity" {
								key := a.Name.Local + ":" + a.Value
								if ids[key] {
									t.Fatal("duplicate", key)
								}
								ids[key] = true
							}
							if (a.Name.Local == "width" || a.Name.Local == "height") && strings.HasPrefix(a.Value, "-") {
								t.Fatal("negative geometry")
							}
						}
					}
				}
			}
		}
	}
	for _, s := range []trace.Step{{}, {State: "wrong"}} {
		if !strings.Contains(renderers.RenderKMPLessonSVG(s, 320, 420), "No KMP state") {
			t.Fatal("fallback")
		}
	}
}

func TestKMPLessonFixedLPSAndSlidingPattern(t *testing.T) {
	for _, step := range kmp.Trace("ABABDABACDABABCABAB", "ABABCABAB") {
		if step.LineID == "L13" {
			svg := renderers.RenderKMPLessonSVG(step, 320, 420)
			if !strings.Contains(svg, `data-entity="kmp-lps-0" data-x="32" data-y="300"`) || !strings.Contains(svg, `j = lps[`) || !strings.Contains(svg, `data-entity="kmp-pattern-0" data-x="128"`) {
				t.Fatal(svg)
			}
			return
		}
	}
	t.Fatal("missing fallback")
}
