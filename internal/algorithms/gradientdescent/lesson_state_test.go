package gradientdescent

import (
	"encoding/json"
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"io"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestGradientDescentLessonFirstEpoch(t *testing.T) {
	steps := Trace([]float64{1, 2, 3, 4}, []float64{3, 5, 7, 9}, .05, 1)
	for _, step := range steps {
		s, ok := step.State.(renderers.GradientDescentLessonState)
		if !ok {
			t.Fatalf("untyped state %T", step.State)
		}
		if step.LineID == "L2" && s.LossBeforeValid {
			t.Fatal("initial loss not measured")
		}
		if step.LineID == "L13" {
			for _, p := range [][2]float64{{s.SumSquaredError, 164}, {s.LossBefore, 41}, {s.SumErrorX, -70}, {s.SumError, -24}, {s.GradientW, -35}, {s.GradientB, -12}} {
				if math.Abs(p[0]-p[1]) > 1e-10 {
					t.Fatalf("loss/gradient %v", p)
				}
			}
		}
		if step.LineID == "L14" && (s.W != 1.75 || s.B != 0 || step.Variables["w"] != s.W || s.BeforeW != 0 || s.LossBefore != 41) {
			t.Fatalf("weight provenance: %+v", s)
		}
	}
	s := steps[len(steps)-1].State.(renderers.GradientDescentLessonState)
	if math.Abs(s.W-1.75) > 1e-10 || math.Abs(s.B-.6) > 1e-10 || math.Abs(s.LossCurrent-1.12875) > 1e-10 || !s.LossCurrentValid || s.LossBefore != 41 || s.StopReason != "epoch-limit" {
		t.Fatalf("final %+v", s)
	}
	if len(s.History) != 1 || s.History[0].W != 0 || s.History[0].B != 0 || s.History[0].Loss != 41 {
		t.Fatal(s.History)
	}
	w, b, h := Train([]float64{1, 2, 3, 4}, []float64{3, 5, 7, 9}, .05, 1)
	if w != s.W || b != s.B || h[0] != s.History[0].Loss {
		t.Fatal("Train mismatch")
	}
}

func TestGradientDescentLessonDegenerateAndDivergent(t *testing.T) {
	for _, tt := range []struct {
		x, y                    []float64
		lr, w, b, loss, current float64
	}{
		{[]float64{0, 0}, []float64{2, 2}, .1, 0, .4, 4, 2.56},
		{[]float64{1}, []float64{1}, 1, 2, 2, 1, 9},
	} {
		steps := Trace(tt.x, tt.y, tt.lr, 1)
		s := steps[len(steps)-1].State.(renderers.GradientDescentLessonState)
		if s.W != tt.w || math.Abs(s.B-tt.b) > 1e-10 || s.LossBefore != tt.loss || math.Abs(s.LossCurrent-tt.current) > 1e-10 {
			t.Fatalf("%+v", s)
		}
	}
	for _, raw := range []string{`null`, `{"x":[1]}`, `{"x":[],"y":[]}`, `{"x":[1],"y":[1,2]}`, `{"x":[1e308],"y":[1e308]}`, `{"lr":0}`, `{"lr":1.01}`, `{"epochs":0}`, `{"epochs":101}`, `{"lr":null}`, `{"x":null,"y":null}`, `{`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"lr":1,"epochs":1}`, `{"epochs":100}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Errorf("rejected %s: %v", raw, err)
		}
	}
}

func TestGradientDescentLessonReplay(t *testing.T) {
	x, y := []float64{1, 2}, []float64{3, 5}
	a := Trace(x, y, .05, 2)
	b := Trace(x, y, .05, 2)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("nondeterministic")
	}
	first := a[0].State.(renderers.GradientDescentLessonState)
	last := a[len(a)-1].State.(renderers.GradientDescentLessonState)
	x[0] = 99
	last.X[0] = 88
	last.Predictions[0] = 77
	last.History[0].Loss = 66
	if first.X[0] != 1 || first.Predictions[0] != 0 || len(first.History) != 0 {
		t.Fatal("aliased snapshot")
	}
	for i, st := range a {
		s := st.State.(renderers.GradientDescentLessonState)
		if s.DataXMin != first.DataXMin || s.SliceLossMax != first.SliceLossMax {
			t.Fatalf("domain drift %d", i)
		}
	}
	for _, args := range []struct {
		x, y []float64
		lr   float64
		e    int
	}{{nil, nil, .1, 1}, {[]float64{math.NaN()}, []float64{1}, .1, 1}, {[]float64{1}, []float64{1}, math.Inf(1), 1}} {
		ss := Trace(args.x, args.y, args.lr, args.e)
		s := ss[len(ss)-1].State.(renderers.GradientDescentLessonState)
		if s.Phase != "invalid" || s.StopReason != "invalid-input" || len(ss) > 2 {
			t.Fatal("invalid guard")
		}
	}
}

func TestGradientDescentLessonSampleBounds(t *testing.T) {
	for _, n := range []int{64, 65} {
		raw, _ := json.Marshal(map[string]any{"x": make([]float64, n), "y": make([]float64, n), "epochs": 1})
		_, err := (Lesson{}).Build(raw)
		if (err == nil) != (n == 64) {
			t.Fatalf("sample boundary %d: %v", n, err)
		}
	}
}

func TestGradientDescentLessonRenderedTrace(t *testing.T) {
	for _, fixture := range []struct {
		x, y   []float64
		lr     float64
		epochs int
	}{
		{[]float64{1, 2, 3, 4}, []float64{3, 5, 7, 9}, .05, 2},
		{[]float64{0, 0}, []float64{2, 2}, .1, 1},
		{[]float64{1}, []float64{1}, 1, 100},
		{[]float64{1e308}, []float64{1e308}, .1, 1},
	} {
		steps := Trace(fixture.x, fixture.y, fixture.lr, fixture.epochs)
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		for _, st := range steps {
			for _, width := range []int{320, 390, 768, 1440} {
				svg := renderers.RenderGradientDescentLessonSVG(st, width, 640)
				if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
					t.Fatalf("nonfinite SVG at %s", st.LineID)
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
		}
	}
}

func TestGradientDescentLessonMotionEndpoints(t *testing.T) {
	steps := Trace([]float64{1, 2, 3, 4}, []float64{3, 5, 7, 9}, .05, 1)
	for _, width := range []int{320, 768} {
		var previous [2]float64
		for _, st := range steps {
			svg := renderers.RenderGradientDescentLessonSVG(st, width, 640)
			d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
			for {
				tok, err := d.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				el, ok := tok.(xml.StartElement)
				if !ok {
					continue
				}
				a := map[string]string{}
				for _, attr := range el.Attr {
					a[attr.Name.Local] = attr.Value
				}
				if a["data-entity"] != "parameter:w" {
					continue
				}
				x, _ := strconv.ParseFloat(a["data-x"], 64)
				y, _ := strconv.ParseFloat(a["data-y"], 64)
				if st.LineID == "L14" {
					var path [][2]float64
					if err := json.Unmarshal([]byte(a["data-path"]), &path); err != nil {
						t.Fatal(err)
					}
					if len(path) < 2 || len(path) > 65 {
						t.Fatal("path length")
					}
					for _, pair := range [][2][2]float64{{path[0], previous}, {path[len(path)-1], {x, y}}} {
						if math.Hypot(pair[0][0]-pair[1][0], pair[0][1]-pair[1][1]) > .01 {
							t.Fatalf("path endpoint %v", pair)
						}
					}
				}
				previous = [2]float64{x, y}
			}
		}
	}
}
