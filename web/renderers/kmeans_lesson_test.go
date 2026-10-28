package renderers

import (
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"io"
	"strings"
	"testing"
)

func TestKMeansLessonSVG(t *testing.T) {
	s := KMeansLessonState{Points: [][2]float64{{0, 0}, {2, 0}}, Centroids: [][2]float64{{1, 0}, {0, 0}}, PreviousCentroids: [][2]float64{{0, 0}, {0, 0}}, Sums: [][2]float64{{2, 0}, {0, 0}}, Counts: []int{2, 0}, Assignments: []int{0, 0}, ActivePoint: 1, Closest: 0, DistanceSquared: []float64{4, 4}, Phase: "means", XMin: -1, XMax: 3, YMin: -1, YMax: 1}
	for _, w := range []int{0, 320, 390, 768, 1440} {
		out := RenderKMeansLessonSVG(trace.Step{State: s}, w, 600)
		for _, want := range []string{`data-lesson="kmeans"`, `data-entity="point:0"`, `data-entity="centroid:0"`, `sum (2, 0) / 2`, `retained: no members`, `data-scene-height=`} {
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
