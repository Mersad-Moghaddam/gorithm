package renderers

import (
	"encoding/xml"
	"io"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

func TestKnapsackLessonSVG(t *testing.T) {
	s := KnapsackLessonState{Weights: []int{2}, Values: []int{3}, DP: [][]int{{0, 0, 0}, {0, 0, 3}}, Known: [][]bool{{true, true, true}, {true, true, true}}, Capacity: 2, Item: 1, Column: 2, Sources: [2][2]int{{0, 2}, {0, 0}}, Fits: true, Exclude: 0, Include: 3, Chosen: 3, Phase: "write"}
	for _, size := range []int{0, 1, 320, 390, 768, 1440} {
		svg := RenderKnapsackLessonSVG(trace.Step{State: s, Explanation: `<script>"&`}, size, size)
		for _, needle := range []string{`data-lesson="knapsack"`, `data-entity="item:0"`, `data-entity="cursor:target"`, `Exclude`, `Include`, `previous row`, `&lt;script&gt;`, `data-min-width=`} {
			if !strings.Contains(svg, needle) {
				t.Fatalf("missing %s", needle)
			}
		}
		if strings.Contains(svg, "NaN") || strings.Contains(svg, "Inf") || strings.Contains(svg, "<script>") {
			t.Fatal("unsafe geometry/text")
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
