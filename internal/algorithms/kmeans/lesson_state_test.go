package kmeans

import (
	"encoding/json"
	"encoding/xml"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
	"io"
	"math"
	"reflect"
	"strings"
	"testing"
)

func TestKMeansLessonMeansAndStop(t *testing.T) {
	points := []Point{{0, 0}, {2, 0}, {10, 0}, {12, 0}}
	for _, limit := range []int{1, 20} {
		steps := Trace(points, 2, limit)
		s, ok := steps[len(steps)-1].State.(renderers.KMeansLessonState)
		if !ok {
			t.Fatal("missing typed lesson state")
		}
		if !reflect.DeepEqual(s.Centroids, [][2]float64{{1, 0}, {11, 0}}) {
			t.Fatal(s.Centroids)
		}
		if s.Converged != (limit == 20) || (limit == 1 && s.StopReason != "iteration-limit") || (limit == 20 && (s.StopReason != "stable" || s.Iteration != 2)) {
			t.Fatalf("wrong termination: %+v", s)
		}
		if steps[len(steps)-1].Variables["converged"] != s.Converged {
			t.Fatal("variables disagree")
		}
		ids := map[string]bool{}
		for _, line := range (Lesson{}).Source() {
			ids[line.ID] = true
		}
		for _, step := range steps {
			if !ids[step.LineID] {
				t.Fatal(step.LineID)
			}
		}
		if !reflect.DeepEqual(steps, Trace(points, 2, limit)) {
			t.Fatal("nondeterministic")
		}
		for _, step := range steps {
			if step.LineID == "L7" && step.State.(renderers.KMeansLessonState).ActivePoint == 1 {
				s := step.State.(renderers.KMeansLessonState)
				if !reflect.DeepEqual(s.DistanceSquared, []float64{4, 64}) {
					t.Fatal(s.DistanceSquared)
				}
				break
			}
		}
	}
}

func TestKMeansLessonTiesAndEmpty(t *testing.T) {
	steps := Trace([]Point{{0, 0}, {0, 0}, {2, 0}}, 3, 20)
	s := steps[len(steps)-1].State.(renderers.KMeansLessonState)
	if s.Assignments[1] != 0 || s.Counts[1] != 0 || s.Centroids[1] != [2]float64{0, 0} {
		t.Fatal(s)
	}
	steps = Trace([]Point{{1, 1}}, 4, 20)
	s = steps[len(steps)-1].State.(renderers.KMeansLessonState)
	if s.RequestedK != 4 || len(s.Centroids) != 1 {
		t.Fatal(s)
	}
	for _, raw := range []string{`null`, `{"points":[]}`, `{"points":null}`, `{"points":[null]}`, `{"points":[{"x":null}]}`, `{"k":0}`, `{"k":65}`, `{"k":null}`, `{"points":[{"x":1e309}]}`, `broken`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
	pts := make([]Point, 65)
	raw, _ := json.Marshal(LessonInput{pts, 1})
	if _, err := (Lesson{}).Build(raw); err == nil {
		t.Fatal("accepted 65 points")
	}
	pts = pts[:64]
	raw, _ = json.Marshal(LessonInput{pts, 64})
	if _, err := (Lesson{}).Build(raw); err != nil {
		t.Fatal(err)
	}
	for _, pts := range [][]Point{nil, {{math.NaN(), 0}}, {{math.MaxFloat64, 0}}, {{math.MaxFloat64, 0}, {-math.MaxFloat64, 0}}} {
		steps := Trace(pts, 1, 1)
		if steps[len(steps)-1].State.(renderers.KMeansLessonState).StopReason != "invalid-input" {
			t.Fatal("invalid direct trace")
		}
	}
	if !strings.Contains((Lesson{}).Source()[13].Code, "!moved") {
		t.Fatal("source omits movement")
	}
}

func TestKMeansLessonSnapshotIsolation(t *testing.T) {
	steps := Trace([]Point{{0, 0}, {2, 0}, {10, 0}, {12, 0}}, 2, 20)
	first := steps[1].State.(renderers.KMeansLessonState)
	last := steps[len(steps)-1].State.(renderers.KMeansLessonState)
	last.Points[0][0] = 99
	last.Centroids[0][0] = 99
	last.Assignments[0] = 99
	if first.Points[0][0] != 0 || first.Centroids[0][0] != 0 || first.Assignments[0] != -1 {
		t.Fatal("aliased snapshots")
	}
	for _, step := range steps {
		s := step.State.(renderers.KMeansLessonState)
		if s.XMin != first.XMin || s.XMax != first.XMax || s.YMin != first.YMin || s.YMax != first.YMax {
			t.Fatal("domain drift")
		}
	}
}

func TestKMeansLessonMotionAndFiniteSnapshots(t *testing.T) {
	steps := Trace([]Point{{0, 0}, {2, 0}, {10, 0}, {12, 0}}, 2, 20)
	pair := false
	for i, step := range steps {
		if step.HoldMS != 0 && step.HoldMS != 600 && step.HoldMS != 1200 {
			t.Fatal("invalid hold")
		}
		if step.LineID == "L8" && i+1 < len(steps) && steps[i+1].LineID == "L9" {
			if step.HoldMS+steps[i+1].HoldMS != 1800 {
				t.Fatal("pair timing")
			}
			pair = true
		}
		out := renderers.RenderKMeansLessonSVG(step, 320, 600)
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
		if step.LineID == "L13" && step.State.(renderers.KMeansLessonState).Iteration == 1 {
			s := step.State.(renderers.KMeansLessonState)
			if !reflect.DeepEqual(s.PreviousCentroids, [][2]float64{{0, 0}, {10, 0}}) || !reflect.DeepEqual(s.Sums, [][2]float64{{2, 0}, {22, 0}}) || !reflect.DeepEqual(s.Counts, []int{2, 2}) {
				t.Fatal(s)
			}
		}
	}
	if !pair {
		t.Fatal("missing compare/commit")
	}
}
