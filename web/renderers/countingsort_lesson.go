package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type CountingsortLessonState struct {
	ArrayState
	Input, InputIDs, Counts, Output, OutputIDs []int
	OutputFilled                               []bool
	Max, Scan, CountIndex, Target              int
	Phase                                      string
}

func RenderCountingsortLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(CountingsortLessonState)
	if width < 320 || width > 32768 {
		width = 800
	}
	if height < 360 || height > 32768 {
		height = 540
	}
	n := len(s.Input)
	width = max(width, 400, 64+80*n)
	cols := (width - 64) / 72
	rows := max(1, (len(s.Counts)+cols-1)/cols)
	outputY := 294 + rows*64
	height = max(height, outputY+100)
	var out strings.Builder
	fmt.Fprintf(&out, `<g data-lesson="countingsort" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" fill="#eff5ed"><title>Counting Sort: immutable input, value-indexed counts, stable output copies</title>`, html.EscapeString(s.Phase), width, height, width)
	if !ok || len(s.InputIDs) != n {
		out.WriteString(`<text x="32" y="40" font-size="16">No Counting Sort state available</text></g>`)
		return out.String()
	}
	if n == 0 {
		out.WriteString(`<text x="32" y="40" font-size="16">Empty array — nothing to sort</text></g>`)
		return out.String()
	}
	countLabel := "Frequency: occurrences of each value"
	if s.Phase == "prefix" {
		countLabel = "Prefix: number &lt;= value (updated left to right)"
	}
	if s.Phase == "output" || s.Phase == "done" {
		countLabel = "Remaining insertion boundary (exclusive)"
	}
	outputLabel := "Output — stable copies, ✓ filled"
	if s.Phase == "done" && len(s.Counts) == 0 {
		countLabel = "Counts not needed for an already sorted input"
		outputLabel = "Result — input returned unchanged, ✓ filled"
	}
	fmt.Fprintf(&out, `<text x="32" y="30" font-size="16" font-weight="600">Input — original identities stay visible</text><text x="32" y="232" font-size="14">%s</text><text x="32" y="%d" font-size="16" font-weight="600">%s</text>`, countLabel, outputY-34, outputLabel)
	// Input never changes, so this domain is identical across every snapshot.
	high := 1
	for _, v := range s.Input {
		high = max(high, v)
	}
	pitch := float64(width-64) / float64(n)
	palette := [...]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	fmt.Fprintf(&out, `<path d="M32 185H%d" stroke="#bbcab9"/><text x="32" y="208" font-size="12">Zero baseline</text>`, width-32)
	for i, value := range s.Input {
		id := s.InputIDs[i]
		if id < 0 {
			continue
		}
		x := 32 + (float64(i)+0.5)*pitch
		h := max(2.0, float64(value)*100/float64(high))
		stroke := "#bbcab9"
		if i == s.Scan {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&out, `<g data-entity="input:%d" data-x="%.2f" data-y="185" transform="translate(%.2f 185)"><title>Input[%d]: value %d, original identity %d</title><rect x="-24" y="%.2f" width="48" height="%.2f" rx="4" fill="%s" stroke="%s" stroke-width="3"/><text x="0" y="%.2f" font-size="14" text-anchor="middle">%d#%d</text><text x="0" y="18" font-size="12" text-anchor="middle">[%d]</text></g>`, id, x, x, i, value, id, -h, h, palette[id%5], stroke, -h-12, value, id, i)
	}
	if len(s.Counts) == 0 {
		out.WriteString(`<text x="32" y="266" font-size="14" fill="#bbcab9">Counts not allocated</text>`)
	}
	for i, value := range s.Counts {
		x, y := 32+(i%cols)*72, 250+(i/cols)*64
		stroke := "#bbcab9"
		if i == s.CountIndex {
			stroke = "#ffe29a"
		}
		fmt.Fprintf(&out, `<g><title>Count for value %d: %d</title><rect x="%d" y="%d" width="60" height="32" rx="4" fill="#263934" stroke="%s" stroke-width="2"/><text x="%d" y="%d" font-size="14" text-anchor="middle">%d</text><text x="%d" y="%d" font-size="12" text-anchor="middle">[%d]</text></g>`, i, value, x, y, stroke, x+30, y+21, value, x+30, y+47, i)
	}
	if len(s.Output) == 0 {
		fmt.Fprintf(&out, `<text x="32" y="%d" font-size="14" fill="#bbcab9">Output not allocated</text>`, outputY+12)
	}
	for i, value := range s.Output {
		x := 32 + (float64(i)+0.5)*pitch
		filled := i < len(s.OutputFilled) && s.OutputFilled[i] && i < len(s.OutputIDs) && s.OutputIDs[i] >= 0
		if !filled {
			fmt.Fprintf(&out, `<rect x="%.2f" y="%d" width="60" height="38" rx="4" fill="none" stroke="#bbcab9" stroke-dasharray="4 4"/><text x="%.2f" y="%d" font-size="14" text-anchor="middle">empty</text>`, x-30, outputY-19, x, outputY+5)
		} else {
			id := s.OutputIDs[i]
			alias := ""
			if step.LineID == "L21" && i == s.Target {
				alias = fmt.Sprintf(` data-from-entity="input:%d" data-motion="linear"`, id)
			}
			fmt.Fprintf(&out, `<g data-entity="output:%d" data-x="%.2f" data-y="%d" transform="translate(%.2f %d)"%s><title>Output[%d]: value %d, original identity %d</title><rect x="-30" y="-19" width="60" height="38" rx="4" fill="%s" stroke="#9bf0bb" stroke-width="3"/><text x="0" y="5" font-size="14" text-anchor="middle">%d#%d</text><text x="0" y="40" font-size="14" fill="#9bf0bb" text-anchor="middle">✓</text></g>`, id, x, outputY, x, outputY, alias, i, value, id, palette[id%5], value, id)
		}
		fmt.Fprintf(&out, `<text x="%.2f" y="%d" font-size="12" text-anchor="middle">[%d]</text>`, x, outputY+62, i)
	}
	out.WriteString(`</g>`)
	return out.String()
}
