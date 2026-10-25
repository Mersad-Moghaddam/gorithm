package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type TableCell struct {
	Row     int    `json:"row"`
	Col     int    `json:"col"`
	Value   string `json:"value"`
	State   string `json:"state"` // "normal", "active", "computed", "source"
}

type TableState struct {
	RowHeaders []string    `json:"rowHeaders"`
	ColHeaders []string    `json:"colHeaders"`
	Cells      []TableCell `json:"cells"`
}

func RenderTableSVG(step trace.Step, width, height int) string {
	state, ok := step.State.(TableState)
	if !ok {
		if m, isMap := step.State.(map[string]any); isMap {
			state = parseTableState(m)
		} else {
			return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">No table state</text>`
		}
	}

	numRows := len(state.RowHeaders)
	numCols := len(state.ColHeaders)
	if numRows == 0 || numCols == 0 {
		return `<text x="50%" y="50%" fill="#8b949e" text-anchor="middle">Empty table</text>`
	}

	cellW := 52.0
	cellH := 36.0
	startX := 90.0
	startY := 50.0

	if avail := (float64(width) - startX - 20.0) / float64(numCols); avail < cellW {
		cellW = avail
	}

	var sb strings.Builder

	// Column Headers
	for c, header := range state.ColHeaders {
		x := startX + float64(c)*cellW
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="12" font-family="ui-monospace, monospace" font-weight="600" text-anchor="middle">%s</text>`,
			x+cellW/2, startY-12, header,
		))
	}

	// Row Headers
	for r, header := range state.RowHeaders {
		y := startY + float64(r)*cellH
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="12" font-family="ui-monospace, monospace" font-weight="600" text-anchor="end">%s</text>`,
			startX-12, y+cellH/2+4, header,
		))
	}

	// Cells
	for _, cell := range state.Cells {
		x := startX + float64(cell.Col)*cellW
		y := startY + float64(cell.Row)*cellH

		fillColor := "#161b22"
		strokeColor := "#2d333b"
		textColor := "#f0f6fc"

		switch cell.State {
		case "active":
			fillColor = "#1e3a5f"
			strokeColor = "#38bdf8"
			textColor = "#7dd3fc"
		case "computed":
			fillColor = "#14532d"
			strokeColor = "#22c55e"
			textColor = "#86efac"
		case "source":
			fillColor = "#581c87"
			strokeColor = "#a855f7"
			textColor = "#d8b4fe"
		}

		sb.WriteString(fmt.Sprintf(
			`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="3" fill="%s" stroke="%s" stroke-width="1.5" />`,
			x, y, cellW-2, cellH-2, fillColor, strokeColor,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="%s" font-size="12" font-family="ui-monospace, monospace" font-weight="600" text-anchor="middle">%s</text>`,
			x+(cellW-2)/2, y+(cellH-2)/2+4, textColor, cell.Value,
		))
	}

	return sb.String()
}

func parseTableState(m map[string]any) TableState {
	var state TableState
	if rowHeaders, ok := m["rowHeaders"].([]any); ok {
		for _, rh := range rowHeaders {
			if s, ok := rh.(string); ok {
				state.RowHeaders = append(state.RowHeaders, s)
			}
		}
	}
	if colHeaders, ok := m["colHeaders"].([]any); ok {
		for _, ch := range colHeaders {
			if s, ok := ch.(string); ok {
				state.ColHeaders = append(state.ColHeaders, s)
			}
		}
	}
	if rawCells, ok := m["cells"].([]any); ok {
		for _, item := range rawCells {
			if cm, ok := item.(map[string]any); ok {
				var cell TableCell
				if r, ok := cm["row"].(float64); ok {
					cell.Row = int(r)
				}
				if c, ok := cm["col"].(float64); ok {
					cell.Col = int(c)
				}
				if v, ok := cm["value"].(string); ok {
					cell.Value = v
				}
				if st, ok := cm["state"].(string); ok {
					cell.State = st
				}
				state.Cells = append(state.Cells, cell)
			}
		}
	}
	return state
}
