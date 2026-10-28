package renderers

import (
	"encoding/xml"
	"io"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestBucketsortLessonSVG(t *testing.T) {
	s := BucketsortLessonState{ArrayState: ArrayState{Values: []int{-3, 0, 10}, Current: -1}, IDs: []int{0, 1, 2}, Min: -3, Max: 10, BucketCount: 5, Buckets: [][]int{{-3}, {0}, nil, nil, {10}}, BucketIDs: [][]int{{0}, {1}, nil, nil, {2}}, Phase: "scatter", Scan: 2, ActiveBucket: 4, BucketOffset: 0, WriteIndex: -1}
	for _, size := range [][2]int{{320, 360}, {960, 540}} {
		svg := RenderBucketsortLessonSVG(trace.Step{State: s, LineID: "L9"}, size[0], size[1])
		for _, want := range []string{`data-lesson="bucketsort"`, `data-entity="bucket4:2"`, `data-from-entity="main2:2"`, `Bucket 4`, `(x-min)*4/(max-min)`, `#c45d75`, `Empty`} {
			if !strings.Contains(svg, want) {
				t.Fatalf("missing %s", want)
			}
		}
		if strings.Contains(svg, "swap-up") || strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") {
			t.Fatal("invalid motion/geometry")
		}
		d := xml.NewDecoder(strings.NewReader("<svg>" + svg + "</svg>"))
		entities := map[string]bool{}
		for {
			token, err := d.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			if e, ok := token.(xml.StartElement); ok {
				for _, a := range e.Attr {
					if a.Name.Local == "data-entity" {
						if entities[a.Value] {
							t.Fatalf("duplicate entity %s", a.Value)
						}
						entities[a.Value] = true
					}
					switch a.Name.Local {
					case "data-x", "data-y", "width", "height", "data-scene-width", "data-scene-height":
						v, err := strconv.ParseFloat(a.Value, 64)
						if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 {
							t.Fatalf("invalid geometry %s=%s", a.Name.Local, a.Value)
						}
					}
				}
			}
		}
	}
	if !strings.Contains(RenderBucketsortLessonSVG(trace.Step{State: BucketsortLessonState{}}, 320, 360), "Empty array") {
		t.Fatal("empty state")
	}
}
