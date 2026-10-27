package renderers

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

var arrayRect = regexp.MustCompile(`<rect x="([^"]+)" y="([^"]+)" width="([^"]+)" height="([^"]+)"`)
var arrayText = regexp.MustCompile(`<text x="([^"]+)" y="([^"]+)"[^>]*>([^<]+)</text>`)

func arrayNumber(t *testing.T, s string) float64 {
	t.Helper()
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		t.Fatalf("invalid coordinate %q: %v", s, err)
	}
	return v
}

func TestArraySignedGeometry(t *testing.T) {
	got := RenderArraySVG(trace.Step{State: ArrayState{Values: []int{-2, 0, 2}, IDs: []string{"neg", "zero", "pos"}, Caption: "Signed", Current: -1}}, 800, 450)
	rects := arrayRect.FindAllStringSubmatch(got, -1)
	if len(rects) != 3 {
		t.Fatalf("want 3 bars, got %d: %s", len(rects), got)
	}
	// Range [-2, 2] spans 310 pixels: zero is halfway at y=225.
	for i, want := range [][2]float64{{225, 155}, {225, 0}, {70, 155}} {
		if y, h := arrayNumber(t, rects[i][2]), arrayNumber(t, rects[i][4]); y != want[0] || h != want[1] {
			t.Fatalf("bar %d y/height = %v/%v, want %v", i, y, h, want)
		}
	}
	if !strings.Contains(got, `y1="225.0"`) || !strings.Contains(got, `data-key="array-neg"`) || !strings.Contains(got, `data-key="array-zero"`) || !strings.Contains(got, `data-key="array-pos"`) || !strings.Contains(got, "Signed") {
		t.Fatalf("missing baseline, keys, or caption: %s", got)
	}
	labels := arrayText.FindAllStringSubmatch(got, -1)
	for _, label := range labels {
		if label[3] == "-2" && arrayNumber(t, label[2]) <= 380 {
			t.Fatalf("negative label must sit below bar: %v", label)
		}
		if label[3] == "[0]" && arrayNumber(t, label[2]) <= 380 {
			t.Fatalf("index must sit below signed bars: %v", label)
		}
	}
}

func TestArrayExtremeAndDenseGeometry(t *testing.T) {
	max := int(^uint(0) >> 1)
	for _, values := range [][]int{{-max - 1, 0, max}, {-max - 1, -1}, {0, 0}, {1, 2}, make([]int, 100)} {
		got := RenderArraySVG(trace.Step{State: ArrayState{Values: values, Current: -1}}, 125, 450)
		rects := arrayRect.FindAllStringSubmatch(got, -1)
		if len(rects) != len(values) {
			t.Fatalf("want %d bars, got %d", len(values), len(rects))
		}
		for _, rect := range rects {
			for _, coord := range rect[1:] {
				arrayNumber(t, coord)
			}
			if arrayNumber(t, rect[3]) < 0 || arrayNumber(t, rect[4]) < 0 {
				t.Fatalf("negative size: %v", rect)
			}
		}
		if strings.Contains(got, "NaN") || strings.Contains(got, "Inf") {
			t.Fatalf("nonfinite SVG: %s", got)
		}
		for _, label := range arrayText.FindAllStringSubmatch(got, -1) {
			if x, y := arrayNumber(t, label[1]), arrayNumber(t, label[2]); x < 0 || x > 125 || y < 0 || y > 450 {
				t.Fatalf("label outside scene: %v", label)
			}
		}
		if values[0] == -max-1 && values[1] == -1 && arrayNumber(t, rects[0][4]) != 310 {
			t.Fatalf("all-negative range should use whole plot height: %v", rects[0])
		}
	}
}

func TestArrayAllNegativeUsesTopZeroBaseline(t *testing.T) {
	got := RenderArraySVG(trace.Step{State: ArrayState{Values: []int{-2, -1}, Current: -1}}, 800, 450)
	rects := arrayRect.FindAllStringSubmatch(got, -1)
	if len(rects) != 2 || arrayNumber(t, rects[0][2]) != 70 || arrayNumber(t, rects[0][4]) != 310 || arrayNumber(t, rects[1][4]) != 155 {
		t.Fatalf("negative bars must descend from top zero line: %v", rects)
	}
}

func TestArrayPositiveBarsKeepFloorGeometry(t *testing.T) {
	got := RenderArraySVG(trace.Step{State: ArrayState{Values: []int{2, 4}, Current: -1}}, 800, 450)
	rects := arrayRect.FindAllStringSubmatch(got, -1)
	if len(rects) != 2 || arrayNumber(t, rects[0][2]) != 225 || arrayNumber(t, rects[0][4]) != 155 || arrayNumber(t, rects[1][2]) != 70 || arrayNumber(t, rects[1][4]) != 310 {
		t.Fatalf("positive bars changed floor placement: %v", rects)
	}
}
