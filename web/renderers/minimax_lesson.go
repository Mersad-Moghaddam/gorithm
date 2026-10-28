package renderers

import (
	"fmt"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"html"
	"math"
	"strings"
)

type MinimaxNodeView struct {
	ID, ParentID                       string
	Depth, Value                       int
	IsMax, Terminal, Evaluated, Pruned bool
}

type MinimaxLessonState struct {
	TreeState
	Nodes                                          []MinimaxNodeView
	ActiveID, ChildID, ReturnedFromID, BestChildID string
	Alpha, Beta, Eval, BestValue                   int
	EvalValid, BestValueValid, Cutoff              bool
	Stack                                          []string
	PrunedIDs                                      []string
	Phase                                          string
}

func minimaxBound(v int) string {
	if v == math.MinInt32 {
		return "−∞"
	}
	if v == math.MaxInt32 {
		return "+∞"
	}
	return fmt.Sprint(v)
}

func RenderMinimaxLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(MinimaxLessonState)
	if !ok {
		return `<g data-lesson="minimax" data-phase="unavailable"><title>Minimax state unavailable</title></g>`
	}
	if width < 320 {
		width = 320
	}
	if width > 1440 {
		width = 1440
	}
	if height < 640 {
		height = 640
	}
	if height > 2000 {
		height = 2000
	}
	// ponytail: full tree uses a local 1:1 viewport; add subtree navigator only for larger lesson limits.
	leaves, maxDepth := 0, 0
	children := map[string][]string{}
	nodes := map[string]MinimaxNodeView{}
	for _, n := range s.Nodes {
		nodes[n.ID] = n
		children[n.ParentID] = append(children[n.ParentID], n.ID)
		if n.Terminal {
			leaves++
		}
		if n.Depth > maxDepth {
			maxDepth = n.Depth
		}
	}
	sceneWidth := max(width, leaves*112+48, 640)
	sceneWidth = min(sceneWidth, 32768)
	treeBottom := max(360, 100+maxDepth*100)
	sceneHeight := max(height, treeBottom+300)
	positions := map[string][2]float64{}
	leaf := 0
	var layout func(string, int) float64
	layout = func(id string, depth int) float64 {
		if depth > 128 {
			return 24
		}
		cs := children[id]
		x := float64(sceneWidth) / 2
		if len(cs) == 0 {
			if leaves > 1 {
				x = 80 + float64(leaf)*float64(sceneWidth-160)/float64(leaves-1)
			}
			leaf++
		} else {
			first, last := 0.0, 0.0
			for i, c := range cs {
				cx := layout(c, depth+1)
				if i == 0 {
					first = cx
				}
				last = cx
			}
			x = (first + last) / 2
		}
		positions[id] = [2]float64{x, 70 + float64(depth)*100}
		return x
	}
	for _, id := range children[""] {
		layout(id, 0)
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="minimax" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif" font-size="14"><title>%s</title>`, html.EscapeString(s.Phase), sceneWidth, sceneHeight, sceneWidth, html.EscapeString("Minimax: "+step.Explanation))
	b.WriteString(`<defs><pattern id="minimax-pruned" width="8" height="8" patternUnits="userSpaceOnUse"><path d="M0 8L8 0" stroke="#bbcab9" stroke-width="1"/></pattern></defs>`)
	text := func(x, y float64, value, color string) {
		fmt.Fprintf(&b, `<text x="%.1f" y="%.1f" fill="%s">%s</text>`, x, y, color, html.EscapeString(value))
	}
	for _, n := range s.Nodes {
		p := positions[n.ID]
		if parent, exists := positions[n.ParentID]; exists {
			dash := ""
			if n.Pruned {
				dash = ` stroke-dasharray="6 5"`
			}
			fmt.Fprintf(&b, `<path d="M%.1f %.1fL%.1f %.1f" stroke="#bbcab9" fill="none"%s/>`, parent[0], parent[1]+24, p[0], p[1]-24, dash)
		}
	}
	palette := []string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	for i, n := range s.Nodes {
		p := positions[n.ID]
		stroke := "#bbcab9"
		if n.Evaluated {
			stroke = "#9bf0bb"
		}
		if n.ID == s.ActiveID {
			stroke = "#ffe29a"
		}
		if n.Pruned {
			stroke = "#ff9b9b"
		}
		fmt.Fprintf(&b, `<g data-entity="node:%s" data-x="%.1f" data-y="%.1f" transform="translate(%.1f %.1f)"><title>%s</title>`, html.EscapeString(n.ID), p[0], p[1], p[0], p[1], html.EscapeString(n.ID))
		shape := `<path d="M0 -28L44 28L-44 28Z"`
		role := "MAX"
		if !n.IsMax {
			shape = `<path d="M-44 -28L44 -28L0 28Z"`
			role = "MIN"
		}
		if n.Terminal {
			shape = `<rect x="-44" y="-26" width="88" height="52" rx="9"`
			role = "leaf"
		}
		fmt.Fprintf(&b, `%s fill="%s" stroke="%s" stroke-width="3"/>`, shape, palette[i%5], stroke)
		label := "?"
		if n.Terminal || n.Evaluated {
			label = fmt.Sprint(n.Value)
		}
		fmt.Fprintf(&b, `<text text-anchor="middle" y="5" fill="#ffffff" font-weight="700">%s</text>`, html.EscapeString(label))
		text(-42, -36, role, "#eff5ed")
		// Short node IDs remain above utility/status; full identity remains in accessible title.
		short := n.ID
		if len(short) > 14 {
			short = "…" + short[len(short)-13:]
		}
		text(-44, 47, short, "#eff5ed")
		if n.Pruned {
			b.WriteString(`<rect x="-44" y="-26" width="88" height="52" rx="9" fill="url(#minimax-pruned)"/><path d="M-32 -18L32 18M-32 18L32 -18" stroke="#ff9b9b" stroke-width="2"/>`)
			text(-48, 65, "not evaluated", "#ff9b9b")
		} else if n.Terminal && !n.Evaluated {
			text(-38, 65, "not visited", "#bbcab9")
		}
		b.WriteString(`</g>`)
	}
	if s.EvalValid && s.ReturnedFromID != "" {
		parent := positions[s.ActiveID]
		p := [2]float64{parent[0] + 80, parent[1]}
		// Find the nearest clear badge position without covering nodes or their labels.
		for radius := 80.0; ; radius += 40 {
			found := false
			for _, offset := range [][2]float64{{radius, 0}, {-radius, 0}, {0, radius}, {0, -radius}} {
				candidate := [2]float64{parent[0] + offset[0], parent[1] + offset[1]}
				if candidate[0] < 32 || candidate[0] > float64(sceneWidth-32) || candidate[1] < 15 || candidate[1] > float64(treeBottom+20) {
					continue
				}
				clear := true
				for _, n := range s.Nodes {
					q := positions[n.ID]
					dx, dy := math.Abs(candidate[0]-q[0]), candidate[1]-q[1]
					if dx < 76 && dy > -43 && dy < 43 || dx < 110 && (dy > -55 && dy < -20 || dy > 32 && dy < 80) {
						clear = false
						break
					}
				}
				if clear {
					p, found = candidate, true
					break
				}
			}
			if found || radius > float64(sceneWidth+sceneHeight) {
				break
			}
		}
		dx, dy := p[0]-parent[0], p[1]-parent[1]
		start := 1 / math.Max(math.Abs(dx)/44, math.Abs(dy)/28)
		end := 1 - 1/math.Max(math.Abs(dx)/32, math.Abs(dy)/15)
		fmt.Fprintf(&b, `<path data-return-to="node:%s" d="M%.1f %.1fL%.1f %.1f" fill="none" stroke="#8bd5ff" stroke-width="2" stroke-dasharray="5 4"/>`, html.EscapeString(s.ActiveID), parent[0]+dx*start, parent[1]+dy*start, parent[0]+dx*end, parent[1]+dy*end)
		fmt.Fprintf(&b, `<g data-entity="utility:%s" data-from-entity="node:%s" data-motion="linear" data-x="%.1f" data-y="%.1f" transform="translate(%.1f %.1f)"><rect x="-32" y="-15" width="64" height="30" rx="10" fill="#192923" stroke="#8bd5ff"/><text text-anchor="middle" y="5" fill="#8bd5ff">%d</text></g>`, html.EscapeString(s.ReturnedFromID), html.EscapeString(s.ReturnedFromID), p[0], p[1], p[0], p[1], s.Eval)
	}
	y := float64(treeBottom + 45)
	text(24, y, "α = "+minimaxBound(s.Alpha)+"    β = "+minimaxBound(s.Beta), "#ffe29a")
	if s.Cutoff {
		text(24, y+26, fmt.Sprintf("β = %s ≤ α = %s; cutoff bound satisfied", minimaxBound(s.Beta), minimaxBound(s.Alpha)), "#ff9b9b")
	}
	if s.EvalValid && s.ReturnedFromID != "" {
		op := "max"
		if !nodes[s.ActiveID].IsMax {
			op = "min"
		}
		text(24, y+52, fmt.Sprintf("%s(children seen, %d) = %d", op, s.Eval, s.BestValue), "#8bd5ff")
	} else if s.BestValueValid {
		text(24, y+52, fmt.Sprintf("Best returned utility = %d; chosen child %s", s.BestValue, s.BestChildID), "#9bf0bb")
	}
	// Wrap prose at spaces; no diagram scale-down needed for readable labels.
	prose := "Stack: " + strings.Join(s.Stack, " / ") + ". " + step.Explanation
	words := strings.Fields(prose)
	line := ""
	row := 0
	limit := (sceneWidth - 48) / 8
	for _, word := range words {
		if len(line)+len(word) > limit && line != "" {
			text(24, y+84+float64(row*22), line, "#eff5ed")
			row++
			line = ""
		}
		if line != "" {
			line += " "
		}
		line += word
	}
	text(24, y+84+float64(row*22), line, "#eff5ed")
	b.WriteString(`</g>`)
	return b.String()
}
