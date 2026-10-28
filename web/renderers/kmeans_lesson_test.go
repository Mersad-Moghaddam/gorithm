package renderers

import (
	"encoding/xml"
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestKMeansLessonSVG(t *testing.T) {
	s := KMeansLessonState{Points: [][2]float64{{0, 0}, {2, 0}}, Centroids: [][2]float64{{1, 0}, {0, 0}}, PreviousCentroids: [][2]float64{{0, 0}, {0, 0}}, Sums: [][2]float64{{2, 0}, {0, 0}}, Counts: []int{2, 0}, Assignments: []int{0, 0}, ActivePoint: 1, Closest: 0, DistanceSquared: []float64{4, 4}, Phase: "means", XMin: -1, XMax: 3, YMin: -1, YMax: 1}
	for _, w := range []int{0, 320, 390, 768, 1440} {
		out := RenderKMeansLessonSVG(trace.Step{State: s}, w, 600)
		for _, want := range []string{`data-lesson="kmeans"`, `data-entity="point:0"`, `data-entity="centroid:0"`, `sum x: 2`, `retained: no members`, `data-scene-height=`} {
			if !strings.Contains(out, want) {
				t.Fatal(want)
			}
		}
		if strings.Contains(out, "NaN") || strings.Contains(out, "Inf") {
			t.Fatal(out)
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + out + "</svg>"))
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
	a := RenderKMeansLessonSVG(trace.Step{State: s}, 320, 600)
	s.Assignments[0] = 1
	b := RenderKMeansLessonSVG(trace.Step{State: s}, 320, 600)
	entity := func(out string) string {
		start := strings.Index(out, `data-entity="point:0"`)
		end := strings.Index(out[start:], ">")
		return out[start : start+end]
	}
	if entity(a) != entity(b) {
		t.Fatal("membership moved point")
	}
}

func TestKMeansMarkerIdentityKey(t *testing.T) {
	s := KMeansLessonState{Points: [][2]float64{{0, 0}, {0, 0}}, Centroids: [][2]float64{{0, 0}, {0, 0}}, XMin: -1, XMax: 1, YMin: -1, YMax: 1, ActivePoint: -1}
	for _, width := range []int{320, 1440} {
		out := RenderKMeansLessonSVG(trace.Step{State: s}, width, 600)
		for _, kind := range []string{"point", "centroid"} {
			for i := 0; i < 2; i++ {
				for _, want := range []string{fmt.Sprintf(`data-entity="%s:%d"`, kind, i), fmt.Sprintf(`<title>%s %d (0, 0)</title>`, kind, i), fmt.Sprintf(`data-marker-key="%s:%d"`, kind, i), fmt.Sprintf(`data-marker-link="%s:%d"`, kind, i)} {
					if !strings.Contains(out, want) {
						t.Errorf("width %d missing %s", width, want)
					}
				}
			}
		}
	}
}

func TestKMeansLessonCoincidentPointsAndLongNumbers(t *testing.T) {
	s := KMeansLessonState{Points: [][2]float64{{1e150, -1e150}, {1e150, -1e150}}, Centroids: [][2]float64{{1e150, -1e150}, {1e150, -1e150}}, Sums: [][2]float64{{2e150, -2e150}, {0, 0}}, Counts: []int{2, 0}, Assignments: []int{0, 0}, ActivePoint: 1, RequestedK: 2, Phase: "means", XMin: -2e150, XMax: 2e150, YMin: -2e150, YMax: 2e150}
	for _, width := range []int{320, 1440} {
		out := RenderKMeansLessonSVG(trace.Step{State: s}, width, 600)
		for _, want := range []string{`data-point-row="0"`, `data-point-row="1"`, `P0 · C0`, `P1 · C0`, `x: 1e+150`, `y: -1e+150`, `sum x: 2e+150`, `sum y: -2e+150`, `count: 2`} {
			if !strings.Contains(out, want) {
				t.Errorf("width %d missing %q", width, want)
			}
		}
		if strings.Contains(out, `<text x="8" y="-10">`) || strings.Contains(out, `<text x="12" y="20">`) {
			t.Errorf("width %d has overlapping plot labels", width)
		}
	}
}

func TestKMeansLessonEmptyClusterAndEffectiveK(t *testing.T) {
	s := KMeansLessonState{Points: [][2]float64{{0, 0}, {1, 1}}, RequestedK: 3, Phase: "data"}
	out := RenderKMeansLessonSVG(trace.Step{State: s}, 320, 600)
	if !strings.Contains(out, "k requested 3 · effective 2") {
		t.Error("initial effective k must reflect point count")
	}
	s.Centroids = [][2]float64{{0, 0}, {1, 1}}
	s.Counts = []int{2, 0}
	s.Sums = [][2]float64{{1, 1}, {0, 0}}
	out = RenderKMeansLessonSVG(trace.Step{State: s}, 320, 600)
	if strings.Contains(out, "sum (0, 0) / 0") || !strings.Contains(out, "retained: no members") {
		t.Fatal("empty centroid must not show division by zero")
	}
}
