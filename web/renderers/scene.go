package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

// SceneState returns SVG children, not an outer <svg>. Implementations escape
// labels and give moving objects unique, stable data-key attributes.
type SceneState interface {
	RenderSVG(width, height int) string
}

func RenderStepSVG(step trace.Step, renderer string, width, height int) string {
	if scene, ok := step.State.(SceneState); ok {
		return scene.RenderSVG(width, height)
	}
	switch renderer {
	case "graph":
		return RenderGraphSVG(step, width, height)
	case "tree":
		return RenderTreeSVG(step, width, height)
	case "table":
		return RenderTableSVG(step, width, height)
	case "hanoi":
		return RenderHanoiSVG(step, width, height)
	default:
		return RenderArraySVG(step, width, height)
	}
}

type WorklistItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// RenderWorklistSVG draws an ordered queue/stack/frontier. IDs identify entries,
// including repeated values. Caller reserves 62px below y for this strip.
func RenderWorklistSVG(label string, items []WorklistItem, x, y, width float64) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" fill="#cbd5e1" font-size="13">%s</text>`, x, y, html.EscapeString(label))
	if len(items) == 0 {
		fmt.Fprintf(&sb, `<text x="%.1f" y="%.1f" fill="#94a3b8" font-size="13">Empty</text>`, x, y+28)
		return sb.String()
	}
	w := min(90.0, width/float64(len(items)))
	for i, item := range items {
		fmt.Fprintf(&sb, `<g data-key="worklist-%s"><title>Position %d: %s</title><rect x="%.1f" y="%.1f" width="%.1f" height="30" rx="5" fill="#1e293b" stroke="#fde047"/><text x="%.1f" y="%.1f" fill="#f8fafc" font-size="13" text-anchor="middle">%s</text></g>`, html.EscapeString(item.ID), i+1, html.EscapeString(item.Label), x+float64(i)*w, y+10, w-5, x+float64(i)*w+(w-5)/2, y+30, html.EscapeString(item.Label))
	}
	return sb.String()
}

func renderCaption(caption string, width int) string {
	if caption == "" {
		return ""
	}
	return fmt.Sprintf(`<text data-key="caption" x="%d" y="24" fill="#e2e8f0" font-size="14" text-anchor="middle">%s</text>`, width/2, html.EscapeString(caption))
}
