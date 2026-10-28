package renderers_test

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/algorithms/trie"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestTrieLessonSVG(t *testing.T) {
	steps := trie.Trace([]string{"é猫", "é", "<&\"", ""}, "absent", "é")
	missing := false
	for _, step := range steps {
		before := renderers.RenderTrieLessonSVG(step, 800, 480)
		for _, size := range [][2]int{{320, 420}, {800, 480}, {0, -1}} {
			svg := renderers.RenderTrieLessonSVG(step, size[0], size[1])
			if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, `<&"`) {
				t.Fatal("unsafe SVG", svg)
			}
			decoder := xml.NewDecoder(strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg">` + svg + `</svg>`))
			entities := map[string]bool{}
			parents := []string{}
			for {
				tok, err := decoder.Token()
				if err != nil {
					if err.Error() != "EOF" {
						t.Fatal(err)
					}
					break
				}
				if el, ok := tok.(xml.StartElement); ok {
					for _, a := range el.Attr {
						if a.Name.Local == "data-parent" {
							parents = append(parents, "node:"+a.Value)
						}
						if a.Name.Local == "data-scene-width" || a.Name.Local == "data-scene-height" || a.Name.Local == "data-min-width" || a.Name.Local == "data-x" || a.Name.Local == "data-y" {
							v, err := strconv.ParseFloat(a.Value, 64)
							if err != nil || v < 0 || v > 32768 {
								t.Fatal("invalid geometry", a)
							}
						}
						if a.Name.Local == "data-entity" {
							if entities[a.Value] {
								t.Fatal("duplicate entity")
							}
							entities[a.Value] = true
						}
					}
				}
			}
			for _, parent := range parents {
				if !entities[parent] {
					t.Fatal("edge references missing parent", parent)
				}
			}
		}
		if before != renderers.RenderTrieLessonSVG(step, 800, 480) {
			t.Fatal("nondeterministic rendering")
		}
		if strings.Contains(before, `data-missing="true"`) {
			missing = true
		}
		s := step.State.(renderers.TrieLessonState)
		for _, id := range s.TerminalIDs {
			if !strings.Contains(before, `data-terminal="`+id+`"`) {
				t.Fatal("lost persistent terminal")
			}
		}
		copyState := s
		copyState.Nodes = append([]renderers.TreeNode(nil), s.Nodes...)
		copyState.Runes = append([]rune(nil), s.Runes...)
		copyState.PathIDs = append([]string(nil), s.PathIDs...)
		copyState.TerminalIDs = append([]string(nil), s.TerminalIDs...)
		renderers.RenderTrieLessonSVG(step, 320, 420)
		if !reflect.DeepEqual(copyState, s) {
			t.Fatal("renderer mutated state")
		}
	}
	if !missing {
		t.Fatal("missing edge must have dashed stub")
	}
	for _, s := range []trace.Step{{}, {State: 42}} {
		if !strings.Contains(renderers.RenderTrieLessonSVG(s, 320, 420), "No trie state") {
			t.Fatal("missing fallback")
		}
	}
}

func TestTrieLessonLayoutMotion(t *testing.T) {
	steps := trie.Trace([]string{"z", "a", "b"}, "z", "a")
	positions := map[string]string{}
	moved := false
	for _, step := range steps {
		svg := renderers.RenderTrieLessonSVG(step, 320, 420)
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		for {
			tok, err := d.Token()
			if err != nil {
				break
			}
			if el, ok := tok.(xml.StartElement); ok {
				attrs := map[string]string{}
				for _, a := range el.Attr {
					attrs[a.Name.Local] = a.Value
				}
				if id := attrs["data-entity"]; id != "" {
					if old, ok := positions[id]; ok && old != attrs["data-x"] {
						moved = true
					}
					positions[id] = attrs["data-x"]
					if attrs["transform"] == "" {
						t.Fatal("missing destination transform")
					}
				}
			}
		}
	}
	if !moved {
		t.Fatal("growing trie must expose stable identities with changed destination positions")
	}
	last := renderers.RenderTrieLessonSVG(steps[len(steps)-1], 320, 420)
	for _, want := range []string{`data-min-width="800"`, `font-size="14"`, `#c45d75`, `#387faa`, `#9bf0bb`, `end`, `path exists`} {
		if !strings.Contains(last, want) {
			t.Fatal("missing readable identity/result cue", want)
		}
	}
}

func TestTrieLessonMaliciousLabels(t *testing.T) {
	s := renderers.TrieLessonState{TreeState: renderers.TreeState{Nodes: []renderers.TreeNode{{ID: "root", Label: `<script>&"`, X: 400, Y: 112}}}, Word: `<script>&"`, Runes: []rune(`<>&"`), RuneIndex: 0, ActiveID: "root", Phase: `compare"<`, Operation: "search"}
	svg := renderers.RenderTrieLessonSVG(trace.Step{State: s}, 800, 480)
	if strings.Contains(svg, "<script>") || !strings.Contains(svg, "&lt;script&gt;") {
		t.Fatal("raw malicious label", svg)
	}
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	for {
		_, err := d.Token()
		if err != nil {
			if err.Error() != "EOF" {
				t.Fatal(err)
			}
			break
		}
	}
}

func TestTrieTerminalSiblingsHaveLabelSpace(t *testing.T) {
	steps := trie.Trace([]string{"a", "b"}, "a", "b")
	svg := renderers.RenderTrieLessonSVG(steps[len(steps)-1], 800, 480)
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	var xs []float64
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		if el, ok := tok.(xml.StartElement); ok && el.Name.Local == "g" {
			attrs := map[string]string{}
			for _, a := range el.Attr {
				attrs[a.Name.Local] = a.Value
			}
			if strings.HasPrefix(attrs["data-entity"], "node:node-") {
				x, err := strconv.ParseFloat(attrs["data-x"], 64)
				if err != nil {
					t.Fatal(err)
				}
				xs = append(xs, x)
			}
		}
	}
	if len(xs) != 2 || xs[1]-xs[0] < 90 {
		t.Fatalf("terminal sibling labels need room: x positions %v", xs)
	}
}

func TestTrieEdgesDoNotDuplicateNodeLabels(t *testing.T) {
	steps := trie.Trace([]string{"ab"}, "ab", "a")
	svg := renderers.RenderTrieLessonSVG(steps[len(steps)-1], 800, 480)
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	inEdge := false
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch el := tok.(type) {
		case xml.StartElement:
			for _, a := range el.Attr {
				if a.Name.Local == "data-edge" {
					inEdge = true
				}
			}
			if inEdge && el.Name.Local == "text" {
				t.Fatal("edge label overlaps parent identity on vertical edges")
			}
		case xml.EndElement:
			if el.Name.Local == "g" && inEdge {
				inEdge = false
			}
		}
	}
}

func TestTrieVerticalLabelsHaveSpace(t *testing.T) {
	steps := trie.Trace([]string{"ab"}, "ab", "a")
	svg := renderers.RenderTrieLessonSVG(steps[len(steps)-1], 800, 480)
	d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
	var ys []float64
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		if el, ok := tok.(xml.StartElement); ok {
			for _, a := range el.Attr {
				if a.Name.Local == "data-y" {
					y, err := strconv.ParseFloat(a.Value, 64)
					if err != nil {
						t.Fatal(err)
					}
					ys = append(ys, y)
				}
			}
		}
	}
	if len(ys) != 3 || ys[1]-ys[0] < 96 || ys[2]-ys[1] < 96 {
		t.Fatalf("parent identity and child cue need separate baselines: %v", ys)
	}
}
