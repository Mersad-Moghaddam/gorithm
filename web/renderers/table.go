package renderers

import (
	"fmt"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type TableCell struct {
	Row   int    `json:"row"`
	Col   int    `json:"col"`
	Value string `json:"value"`
	State string `json:"state"` // "normal", "active", "computed", "source"
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
			return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">No table state available</text>`
		}
	}

	numRows := len(state.RowHeaders)
	numCols := len(state.ColHeaders)
	if numRows == 0 || numCols == 0 {
		return `<text x="50%" y="50%" fill="#64748b" font-family="system-ui, sans-serif" font-size="14" text-anchor="middle">Empty table</text>`
	}

	cellW := 54.0
	cellH := 38.0
	startX := 100.0
	startY := 60.0

	if avail := (float64(width) - startX - 30.0) / float64(numCols); avail < cellW {
		cellW = avail
	}
	if availH := (float64(height) - startY - 30.0) / float64(numRows); availH < cellH {
		cellH = availH
	}

	var sb strings.Builder

	// Defs for table styling
	sb.WriteString(`<defs>
		<filter id="cell-glow" x="-30%" y="-30%" width="160%" height="160%">
			<feGaussianBlur in="SourceGraphic" stdDeviation="4" result="blur" />
			<feMerge>
				<feMergeNode in="blur" />
				<feMergeNode in="SourceGraphic" />
			</feMerge>
		</filter>
	</defs>`)

	// Column Headers
	for c, header := range state.ColHeaders {
		x := startX + float64(c)*cellW
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#94a3b8" font-size="12" font-family="ui-monospace, Menlo, monospace" font-weight="700" text-anchor="middle">%s</text>`,
			x+cellW/2, startY-14, header,
		))
	}

	// Row Headers
	for r, header := range state.RowHeaders {
		y := startY + float64(r)*cellH
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#94a3b8" font-size="12" font-family="ui-monospace, Menlo, monospace" font-weight="700" text-anchor="end">%s</text>`,
			startX-14, y+cellH/2+4, header,
		))
	}

	// Cells
	for _, cell := range state.Cells {
		x := startX + float64(cell.Col)*cellW
		y := startY + float64(cell.Row)*cellH

		fillColor := "#0f172a"
		strokeColor := "#1e293b"
		textColor := "#cbd5e1"
		filterStr := ""

		switch cell.State {
		case "active":
			fillColor = "#0284c7"
			strokeColor = "#38bdf8"
			textColor = "#ffffff"
			filterStr = `filter="url(#cell-glow)"`
		case "computed":
			fillColor = "#15803d"
			strokeColor = "#4ade80"
			textColor = "#ffffff"
		case "source":
			fillColor = "#7e22ce"
			strokeColor = "#c084fc"
			textColor = "#ffffff"
			filterStr = `filter="url(#cell-glow)"`
		}

		sb.WriteString(fmt.Sprintf(
			`<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="6" fill="%s" stroke="%s" stroke-width="1.5" %s style="transition: fill 0.2s ease, stroke 0.2s ease;" />`,
			x, y, cellW-4, cellH-4, fillColor, strokeColor, filterStr,
		))
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="%s" font-size="12" font-family="ui-monospace, Menlo, monospace" font-weight="700" text-anchor="middle">%s</text>`,
			x+(cellW-4)/2, y+(cellH-4)/2+4, textColor, cell.Value,
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
