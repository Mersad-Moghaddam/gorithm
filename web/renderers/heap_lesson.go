package renderers

import (
	"fmt"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// HeapLessonState keeps heap slots separate from stable item identities.
type HeapLessonState struct {
	ArrayState
	IDs                                                          []int
	HeapSize, Root, LeftChild, RightChild, Largest, ExtractIndex int
	Building                                                     bool
}

func heapLessonPoint(slot int, width float64) (float64, float64) {
	depth := int(math.Floor(math.Log2(float64(slot + 1))))
	first := (1 << depth) - 1
	return (float64(slot-first) + 0.5) * width / math.Exp2(float64(depth)), 116 + float64(depth)*88
}

// RenderHeapLessonSVG projects the same items onto heap slots and indexed storage.
func RenderHeapLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(HeapLessonState)
	if !ok {
		return `<g data-lesson="heap" data-phase="unavailable"><title>Heap Sort</title><text x="32" y="48" fill="#eff5ed" font-size="16">No heap state available</text></g>`
	}
	n := len(s.Values)
	sceneW := math.Max(float64(width), math.Max(640, float64(n)*112+64))
	// A complete binary tree needs room for every potential last-level slot.
	depth := 0
	if n > 0 {
		depth = int(math.Floor(math.Log2(float64(n))))
	}
	sceneW = math.Max(sceneW, math.Exp2(float64(depth))*112+64)
	sceneW = math.Min(sceneW, 32768)
	treeBottom := 116 + float64(depth)*88
	railY := treeBottom + 118
	sceneH := math.Max(float64(height), railY+92)
	size := min(max(0, s.HeapSize), n)
	phase := "extract"
	label := "Extract maximum; restore remaining heap"
	if s.Building {
		phase = "build"
		label = "Build max-heap: inspect one subtree"
	}
	if step.Action == trace.ActionDone {
		phase = "done"
		label = "Sorted array"
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="heap" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f" font-family="system-ui, sans-serif" text-anchor="middle"><title>Heap Sort: tree and array share item identities</title><text x="%.1f" y="28" fill="#eff5ed" font-size="16">%s</text><text x="%.1f" y="54" fill="#bbcab9" font-size="14">Active tree: %d items; sorted suffix stays on array rail</text>`, phase, sceneW, sceneH, sceneW, sceneW/2, label, sceneW/2, size)
	if n == 0 {
		b.WriteString(`<text x="160" y="110" fill="#eff5ed" font-size="16">Empty array</text></g>`)
		return b.String()
	}
	for child := 1; child < size; child++ {
		parent := (child - 1) / 2
		x1, y1 := heapLessonPoint(parent, sceneW)
		x2, y2 := heapLessonPoint(child, sceneW)
		fmt.Fprintf(&b, `<line data-parent="%d" data-child="%d" x1="%.2f" y1="%.2f" x2="%.2f" y2="%.2f" stroke="#bbcab9" stroke-width="2"/>`, parent, child, x1, y1+20, x2, y2-20)
	}
	contains := func(a []int, i int) bool {
		for _, v := range a {
			if v == i {
				return true
			}
		}
		return false
	}
	colors := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, value := range s.Values {
		if i >= len(s.IDs) {
			continue
		}
		id := s.IDs[i]
		color := colors[((id%5)+5)%5]
		stroke := "#bbcab9"
		strokeWidth := 1
		if contains(s.Sorted, i) {
			stroke = "#9bf0bb"
			strokeWidth = 3
		}
		if contains(s.Comparing, i) {
			stroke = "#ffe29a"
			strokeWidth = 4
		}
		motion := ""
		if len(s.Swapped) == 2 && contains(s.Swapped, i) {
			kind := "swap-down"
			if i == s.Swapped[0] {
				kind = "swap-up"
			}
			motion = fmt.Sprintf(` data-motion="%s"`, kind)
		}
		emphasis := ""
		if contains(s.Comparing, i) {
			emphasis = ` data-emphasis="compare"`
		}
		// Secondary rail uses compact tiles: value, original identity, and slot stay explicit.
		x := 32 + (float64(i)+0.5)*(sceneW-64)/float64(n)
		fmt.Fprintf(&b, `<g data-entity="array:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s%s><rect x="-50" y="-23" width="100" height="46" rx="8" fill="%s" stroke="%s" stroke-width="%d"/><text y="5" fill="#000000" font-size="16">%d</text><text y="43" fill="#bbcab9" font-size="12">#%d · [%d]</text>`, id, x, railY, x, railY, motion, emphasis, color, stroke, strokeWidth, value, id, i)
		if contains(s.Sorted, i) {
			b.WriteString(`<text y="66" fill="#9bf0bb" font-size="14">✓ final</text>`)
		}
		b.WriteString(`</g>`)
		if i >= size {
			continue
		}
		tx, ty := heapLessonPoint(i, sceneW)
		fmt.Fprintf(&b, `<g data-entity="tree:%d" data-x="%.2f" data-y="%.2f" transform="translate(%.2f %.2f)"%s%s><rect x="-50" y="-23" width="100" height="46" rx="20" fill="%s" stroke="%s" stroke-width="%d"/><text y="5" fill="#000000" font-size="16">%d</text><text y="40" fill="#bbcab9" font-size="12">#%d · [%d]</text>`, id, tx, ty, tx, ty, motion, emphasis, color, stroke, strokeWidth, value, id, i)
		if i == s.Largest {
			b.WriteString(`<text y="-31" fill="#ffe29a" font-size="14">candidate</text>`)
		} else if i == s.Root {
			b.WriteString(`<text y="-31" fill="#8bd5ff" font-size="14">root</text>`)
		}
		b.WriteString(`</g>`)
	}
	b.WriteString(`</g>`)
	return b.String()
}
