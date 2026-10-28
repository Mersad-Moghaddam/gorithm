package renderers

import (
	"fmt"
	"html"
	"math"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type QLearningLessonState struct {
	TableState
	Rows, Cols, Episode, EpisodeStep, Episodes                     int
	Start, Goal, State, NextState                                  [2]int
	Obstacles                                                      [][2]int
	Q                                                              [][]float64
	Action, BestNextAction                                         int
	Alpha, Gamma, Epsilon, Reward, BestNextQ, TDTarget, OldQ, NewQ float64
	TransitionValid, UpdateValid, Terminal                         bool
	Phase, EpisodeStopReason                                       string
	CompletedEpisodes                                              int
}

func RenderQLearningLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(QLearningLessonState)
	if !ok {
		s.Phase = "invalid-input"
	}
	// ponytail: large grids keep 64px cells in a local scrollable scene; reflow only panels, never shrink cell labels.
	narrow := width < 640
	inset := 24.0
	if narrow {
		inset = 16
	}
	cols, rows := s.Cols, s.Rows
	valid := cols > 0 && rows > 0 && cols <= 64 && rows <= 64 && rows*cols <= 64
	if !valid {
		cols = 3
		rows = 3
	}
	cell := math.Max(64, math.Min(100, 300/float64(max(cols, rows))))
	gw, gh := float64(cols)*cell, float64(rows)*cell
	sw := math.Max(760, gw+400+2*inset)
	panelX, panelY := inset+gw+36, 64.0
	gridX, gridY := inset, 64.0
	equationY := math.Max(gh+112, 440)
	if narrow {
		sw = math.Max(390, gw+2*inset)
		panelX = inset
		panelY = gridY + gh + 56
		equationY = panelY + 330
	}
	sh := equationY + 220
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="qlearning" data-phase="%s" data-scene-width="%.0f" data-scene-height="%.0f" data-min-width="%.0f"><title>Q-learning: %s</title>`, html.EscapeString(s.Phase), sw, sh, sw, html.EscapeString(s.Phase))
	text := func(x, y float64, color, content string) {
		fmt.Fprintf(&b, `<text x="%g" y="%g" fill="%s" font-size="14">%s</text>`, x, y, color, html.EscapeString(content))
	}
	text(inset, 28, "#eff5ed", fmt.Sprintf("Episode %d / %d · transition %d / 12", s.Episode, s.Episodes, s.EpisodeStep))
	if !valid || !ok || s.Phase == "invalid-input" {
		text(inset, 64, "#ff9b9b", "Invalid input: no environment rendered.")
		b.WriteString(`</g>`)
		return b.String()
	}
	finite := func(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
	for _, v := range []float64{s.Alpha, s.Gamma, s.Epsilon, s.Reward, s.BestNextQ, s.TDTarget, s.OldQ, s.NewQ} {
		if !finite(v) {
			text(inset, 64, "#ff9b9b", "Invalid numeric snapshot; cannot draw Q estimates.")
			b.WriteString(`</g>`)
			return b.String()
		}
	}
	for _, row := range s.Q {
		for _, v := range row {
			if !finite(v) {
				text(inset, 64, "#ff9b9b", "Invalid numeric snapshot; cannot draw Q estimates.")
				b.WriteString(`</g>`)
				return b.String()
			}
		}
	}
	inside := func(p [2]int) bool { return p[0] >= 0 && p[0] < rows && p[1] >= 0 && p[1] < cols }
	if !inside(s.State) || !inside(s.Start) || !inside(s.Goal) || (s.TransitionValid && !inside(s.NextState)) {
		text(inset, 64, "#ff9b9b", "Invalid state coordinates.")
		b.WriteString(`</g>`)
		return b.String()
	}
	position := func(p [2]int) (float64, float64) {
		return gridX + (float64(p[1])+.5)*cell, gridY + (float64(p[0])+.5)*cell
	}
	b.WriteString(`<defs><pattern id="qlearning-blocked" width="8" height="8" patternUnits="userSpaceOnUse"><path d="M0 8L8 0" stroke="#bbcab9" stroke-width="1"/></pattern><marker id="qlearning-arrow" markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto"><path d="M0 0L8 4L0 8Z" fill="#8bd5ff"/></marker></defs>`)
	obstacles := map[[2]int]bool{}
	for _, p := range s.Obstacles {
		obstacles[p] = true
	}
	names := [4]string{"Up", "Right", "Down", "Left"}
	dx := [4]float64{0, 1, 0, -1}
	dy := [4]float64{-1, 0, 1, 0}
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			p := [2]int{r, c}
			x, y := position(p)
			fill := "#22372c"
			if obstacles[p] {
				fill = "url(#qlearning-blocked)"
			}
			fmt.Fprintf(&b, `<rect x="%g" y="%g" width="%g" height="%g" rx="5" fill="%s" stroke="#bbcab9"/>`, x-cell/2, y-cell/2, cell, cell, fill)
			text(x-cell/2+5, y-cell/2+17, "#bbcab9", fmt.Sprintf("(%d,%d)", r, c))
			switch {
			case obstacles[p]:
				text(x-cell/2+5, y+8, "#eff5ed", "blocked")
			case p == s.Goal:
				fmt.Fprintf(&b, `<path d="M%g %g l4 10 11 0 -9 7 4 11 -10 -7 -10 7 4 -11 -9 -7 11 0Z" fill="#9bf0bb"/>`, x, y-10)
				text(x+16, y+10, "#9bf0bb", "G")
			default:
				idx := r*cols + c
				if idx < len(s.Q) && len(s.Q[idx]) == 4 {
					maximum := s.Q[idx][0]
					for _, v := range s.Q[idx] {
						if v > maximum {
							maximum = v
						}
					}
					for a, v := range s.Q[idx] {
						if v == maximum {
							ax, ay := x+dx[a]*15, y+dy[a]*15
							fmt.Fprintf(&b, `<path aria-label="candidate %s at (%d,%d)" d="M%g %g l%g %g" stroke="#aa92e7" stroke-width="2"/><text x="%g" y="%g" fill="#eff5ed" font-size="14">%d</text>`, names[a], r, c, ax, ay, dx[a]*7, dy[a]*7, ax-4, ay+5, a)
						}
					}
				}
			}
			if p == s.Start {
				fmt.Fprintf(&b, `<rect x="%g" y="%g" width="%g" height="%g" fill="none" stroke="#9bf0bb" stroke-width="2"/>`, x-cell/2+3, y-cell/2+3, cell-6, cell-6)
				text(x-cell/2+5, y+cell/2-5, "#9bf0bb", "S")
			}
		}
	}
	x, y := position(s.State)
	showAction := s.Phase != "assign-state" && s.Phase != "episode-end" && s.Phase != "complete"
	if showAction && s.Action >= 0 && s.Action < 4 {
		fmt.Fprintf(&b, `<path d="M%g %g l%g %g" stroke="#ffe29a" stroke-width="4"/>`, x, y, dx[s.Action]*(cell/2-5), dy[s.Action]*(cell/2-5))
	}
	if s.TransitionValid && s.Phase != "assign-state" && s.Phase != "episode-end" && s.Phase != "complete" {
		nx, ny := position(s.NextState)
		if s.NextState == s.State {
			text(x-cell/2+5, y+cell/2-5, "#ff9b9b", "× collision")
		} else {
			fmt.Fprintf(&b, `<path d="M%g %gL%g %g" stroke="#8bd5ff" stroke-width="3" marker-end="url(#qlearning-arrow)"/><circle cx="%g" cy="%g" r="16" fill="none" stroke="#8bd5ff" stroke-dasharray="4 3"/>`, x, y, nx, ny, nx, ny)
		}
	}
	agentID := "agent:0"
	if s.Phase == "reset" {
		agentID = "agent:reset"
	}
	fmt.Fprintf(&b, `<g data-entity="%s" data-x="%g" data-y="%g" data-motion="linear" transform="translate(%g %g)"><circle r="11" fill="#192923" stroke="#ee8b9b" stroke-width="4"/><title>Agent at (%d,%d)</title></g>`, agentID, x, y, x, y, s.State[0], s.State[1])
	text(inset, gridY+gh+25, "#bbcab9", "current Q estimates")
	text(inset, gridY+gh+45, "#bbcab9", "0 Up · 1 Right · 2 Down · 3 Left")
	number := func(v float64) string { return fmt.Sprintf("%.5g", v) }
	panel := func(px, py float64, label string, p [2]int) {
		text(px, py, "#eff5ed", label+fmt.Sprintf(" (%d,%d)", p[0], p[1]))
		idx := p[0]*cols + p[1]
		if !inside(p) || idx >= len(s.Q) || len(s.Q[idx]) != 4 {
			return
		}
		for a, v := range s.Q[idx] {
			color := "#eff5ed"
			if showAction && a == s.Action && label == "Current row" {
				color = "#ffe29a"
			}
			text(px+8, py+24+float64(a)*23, color, names[a]+": "+number(v))
		}
	}
	panel(panelX, panelY, "Current row", s.State)
	if s.TransitionValid {
		panel(panelX, panelY+140, "Next row", s.NextState)
		text(panelX, panelY+263, "#8bd5ff", fmt.Sprintf("reward %s · old %s", number(s.Reward), number(s.OldQ)))
		if s.UpdateValid {
			text(panelX, panelY+310, "#9bf0bb", "new Q "+number(s.NewQ))
		}
	}
	text(panelX, panelY+287, "#bbcab9", fmt.Sprintf("α %s · γ %s · ε %s", number(s.Alpha), number(s.Gamma), number(s.Epsilon)))
	text(inset, equationY, "#eff5ed", "target = r + γ max Q(next, a)")
	if s.Phase == "target" || s.UpdateValid {
		text(inset, equationY+25, "#8bd5ff", fmt.Sprintf("target = %s + %s × %s = %s", number(s.Reward), number(s.Gamma), number(s.BestNextQ), number(s.TDTarget)))
	} else {
		text(inset, equationY+25, "#bbcab9", "Target not calculated yet.")
	}
	text(inset, equationY+55, "#eff5ed", "Qnew = Qold + α(target − Qold)")
	if s.UpdateValid {
		text(inset, equationY+80, "#9bf0bb", fmt.Sprintf("new Q = %s + %s ×", number(s.OldQ), number(s.Alpha)))
		text(inset+16, equationY+103, "#9bf0bb", fmt.Sprintf("(%s − %s) = %s", number(s.TDTarget), number(s.OldQ), number(s.NewQ)))
	} else {
		text(inset, equationY+80, "#bbcab9", "Q update not applied yet.")
	}
	if s.Terminal {
		text(inset, equationY+128, "#9bf0bb", "Goal: future contribution forced to 0.")
	}
	status := fmt.Sprintf("Completed episodes: %d", s.CompletedEpisodes)
	if s.EpisodeStopReason != "" {
		status += " · " + s.EpisodeStopReason
	}
	text(inset, equationY+154, "#eff5ed", status)
	text(inset, equationY+180, "#bbcab9", "Finite budget; no optimal-policy guarantee.")
	b.WriteString(`</g>`)
	return b.String()
}
