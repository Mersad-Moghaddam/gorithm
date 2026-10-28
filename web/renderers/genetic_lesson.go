package renderers

import (
	"fmt"
	"html"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type GeneticIndividualView struct {
	ID, Genes string
	Fitness   int
}

type GeneticLessonState struct {
	TableState
	Target                                           string
	Generation, Pair, Cut, BestIndex, MaxGenerations int
	MutationRate                                     float64
	Population, NextPopulation                       []GeneticIndividualView
	Parents, Crossed, Mutated                        [2]GeneticIndividualView
	ChangedLoci                                      [2][]int
	UsedOffspring                                    [2]bool
	EliteID                                          string
	BestFitnessHistory                               []int
	Phase, StopReason                                string
}

func geneticLessonColor(id string) string {
	colors := [5]string{"#c45d75", "#387faa", "#8763bb", "#b67b30", "#258b80"}
	gen, slot := 0, 0
	fmt.Sscanf(id, "g%d-i%d", &gen, &slot)
	return colors[(gen*24+slot)%len(colors)]
}

func geneticLessonText(b *strings.Builder, x, y int, text, color string) {
	fmt.Fprintf(b, `<text x="%d" y="%d" fill="%s" font-size="14">%s</text>`, x, y, color, html.EscapeString(text))
}

// RenderGeneticLessonSVG draws only the selected recorded state. Local scrolling
// preserves 14px labels rather than scaling long populations into unreadable rows.
func RenderGeneticLessonSVG(step trace.Step, width, height int) string {
	s, ok := step.State.(GeneticLessonState)
	if !ok {
		return `<g data-lesson="genetic" data-phase="invalid-input" data-scene-width="320" data-scene-height="100"><title>Genetic state unavailable</title></g>`
	}
	if width < 390 {
		width = 390
	}
	if width > 1440 {
		width = 1440
	}
	mobile := width < 640
	inset := 24
	if mobile {
		inset = 16
	}
	columns := 10
	if mobile {
		columns = (width - 2*inset) / 28
		if columns > 8 {
			columns = 8
		}
		if columns < 1 {
			columns = 1
		}
	}
	geneLen := len(s.Target)
	if geneLen < 1 {
		geneLen = 1
	}
	rows := (geneLen + columns - 1) / columns
	stripH := 44 + rows*34
	panelW := (width - 3*inset) / 2
	if mobile {
		panelW = width - 2*inset
	}
	// Each row wraps to explicit byte ranges. Gene boxes never shrink below 26px.
	if columns*28 > panelW-76 {
		columns = (panelW - 76) / 28
		if columns < 1 {
			columns = 1
		}
		rows = (geneLen + columns - 1) / columns
		stripH = 44 + rows*34
	}
	top := 100 + stripH
	popH := 34 + len(s.Population)*(stripH+10)
	benchH := 60 + 4*(stripH+12) + 80
	benchX, benchY := 2*inset+panelW, top
	nextY := top + max(popH, benchH) + 32
	if mobile {
		benchX = inset
		benchY = top + popH + 24
		nextY = benchY + benchH + 24
	}
	changes := len(s.ChangedLoci[0]) + len(s.ChangedLoci[1])
	if summaryH := 32 + 4*(stripH+12) + changes*20; summaryH > benchH {
		benchH = summaryH
		nextY = top + max(popH, benchH) + 32
		if mobile {
			nextY = benchY + benchH + 24
		}
	}
	nextH := 40 + max(1, len(s.NextPopulation))*(stripH+10)
	historyY := nextY + nextH + 40
	sceneH := historyY + 240
	if sceneH < height && height <= 32768 {
		sceneH = height
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<g data-lesson="genetic" data-phase="%s" data-scene-width="%d" data-scene-height="%d" data-min-width="%d" font-family="system-ui, sans-serif"><title>%s</title>`, html.EscapeString(s.Phase), width, sceneH, width, html.EscapeString(fmt.Sprintf("Genetic algorithm: %s, generation %d. Target %s. %s", s.Phase, s.Generation, s.Target, s.StopReason)))
	fmt.Fprintf(&b, `<defs><marker id="genetic-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto"><path d="M0 0L10 5L0 10Z" fill="#8bd5ff"/></marker></defs>`)
	geneticLessonText(&b, inset, 26, fmt.Sprintf("Generation %d/%d · %s", s.Generation, s.MaxGenerations, s.Phase), "#eff5ed")
	// All movable groups use root coordinates; none inherits a translated parent.
	strip := func(role string, p GeneticIndividualView, x, y int, alias string, origins [2]string, cut int, changed []int, before string, note string) {
		if p.ID == "" {
			return
		}
		fmt.Fprintf(&b, `<g data-entity="%s" data-x="%d" data-y="%d" transform="translate(%d %d)" data-motion="linear"`, html.EscapeString(role+":"+p.ID), x, y, x, y)
		if alias != "" {
			fmt.Fprintf(&b, ` data-from-entity="%s"`, html.EscapeString(alias))
		}
		fmt.Fprint(&b, `>`)
		geneticLessonText(&b, 0, 0, p.ID, "#eff5ed")
		geneticLessonText(&b, 0, 18, fmt.Sprintf("%d/%d %s", p.Fitness, len(s.Target), note), "#eff5ed")
		for k := 0; k < len(p.Genes); k++ {
			cx, cy := (k%columns)*28, 32+(k/columns)*34
			origin := p.ID
			if origins[0] != "" {
				origin = origins[0]
				if k >= cut {
					origin = origins[1]
				}
			}
			color := geneticLessonColor(origin)
			stroke := "#bbcab9"
			sw := 1
			for _, l := range changed {
				if l == k {
					stroke = "#ffe29a"
					sw = 3
				}
			}
			fmt.Fprintf(&b, `<rect data-origin="%s" x="%d" y="%d" width="26" height="26" rx="4" fill="%s" stroke="%s" stroke-width="%d"/>`, html.EscapeString(origin), cx, cy, color, stroke, sw)
			geneticLessonText(&b, cx+8, cy+18, string(p.Genes[k]), "#ffffff")
			if k < len(s.Target) && p.Genes[k] == s.Target[k] {
				fmt.Fprintf(&b, `<path d="M%d %dl3 3 6-6" fill="none" stroke="#9bf0bb" stroke-width="2"/>`, cx+16, cy+22)
			}
			if origins[0] != "" && k == cut && k%columns != 0 {
				fmt.Fprintf(&b, `<path d="M%d %dv30" stroke="#ffe29a" stroke-width="2" stroke-dasharray="3 3"/>`, cx-1, cy-2)
			}
			if k%columns == 0 && rows > 1 {
				fmt.Fprintf(&b, `<text x="%d" y="%d" fill="#bbcab9" font-size="14">[%d–%d]</text>`, columns*28+4, cy+18, k, min(k+columns-1, len(p.Genes)-1))
			}
		}
		fmt.Fprint(&b, `</g>`)
		_ = before
	}
	geneticLessonText(&b, inset, 48, "Target: "+s.Target, "#9bf0bb")
	strip("target", GeneticIndividualView{ID: "target", Genes: s.Target, Fitness: len(s.Target)}, inset, 76, "", [2]string{}, 0, nil, "", "")
	geneticLessonText(&b, inset, top, "Population · stable candidate IDs", "#ffe29a")
	for i, p := range s.Population {
		alias := ""
		if s.Phase == "replacement" {
			alias = "next:" + p.ID
		}
		note := ""
		if p.ID == s.EliteID {
			note = "elite ✓"
		}
		strip("population", p, inset, top+28+i*(stripH+10), alias, [2]string{}, 0, nil, "", note)
	}
	geneticLessonText(&b, benchX, benchY, fmt.Sprintf("Operator bench · pair %d", s.Pair), "#ffe29a")
	for j, p := range s.Parents {
		strip(fmt.Sprintf("parent%d", j), p, benchX, benchY+32+j*(stripH+12), "population:"+p.ID, [2]string{}, 0, nil, "", fmt.Sprintf("parent %d", j+1))
	}
	for j, p := range s.Crossed {
		if s.Phase == "mutation" {
			p = s.Mutated[j]
		}
		origin := [2]string{s.Parents[j].ID, s.Parents[1-j].ID}
		note := fmt.Sprintf("child %d · cut %d", j+1, s.Cut)
		if s.Phase == "mutation" && !s.UsedOffspring[j] {
			note = "unused: no slot"
		}
		strip("offspring", p, benchX, benchY+32+(j+2)*(stripH+12), fmt.Sprintf("parent%d:%s", j, s.Parents[j].ID), origin, s.Cut, s.ChangedLoci[j], s.Crossed[j].Genes, note)
	}
	reasonY := benchY + 32 + 4*(stripH+12)
	for j, loci := range s.ChangedLoci {
		for _, k := range loci {
			if k >= 0 && k < len(s.Crossed[j].Genes) && k < len(s.Mutated[j].Genes) {
				geneticLessonText(&b, benchX, reasonY, fmt.Sprintf("child %d · %d: %c → %c", j+1, k, s.Crossed[j].Genes[k], s.Mutated[j].Genes[k]), "#ffe29a")
				reasonY += 20
			}
		}
	}
	geneticLessonText(&b, inset, nextY, "Next generation · elite + offspring", "#8bd5ff")
	for i, p := range s.NextPopulation {
		alias := "offspring:" + p.ID
		if p.ID == s.EliteID {
			alias = "population:" + p.ID
		}
		strip("next", p, inset, nextY+28+i*(stripH+10), alias, [2]string{}, 0, nil, "", "")
	}
	if s.Phase == "selection" {
		for j, p := range s.Parents {
			for i, candidate := range s.Population {
				if p.ID != candidate.ID {
					continue
				}
				fromX, fromY := inset+panelW-8, top+28+i*(stripH+10)+14
				toX, toY := benchX, benchY+32+j*(stripH+12)+14
				fmt.Fprintf(&b, `<path d="M%d %dL%d %d" fill="none" stroke="#8bd5ff" stroke-dasharray="5 4" marker-end="url(#genetic-arrow)"/>`, fromX, fromY, toX, toY)
			}
		}
	}
	if s.Phase == "mutation" {
		for j, p := range s.Mutated {
			if !s.UsedOffspring[j] {
				continue
			}
			for i, dest := range s.NextPopulation {
				if dest.ID == p.ID {
					fmt.Fprintf(&b, `<path d="M%d %dL%d %d" fill="none" stroke="#8bd5ff" marker-end="url(#genetic-arrow)"/>`, benchX+panelW-8, benchY+46+(j+2)*(stripH+12), inset+panelW-8, nextY+42+i*(stripH+10))
				}
			}
		}
	}
	geneticLessonText(&b, inset, historyY, "Best fitness · elite never regresses", "#9bf0bb")
	fmt.Fprintf(&b, `<path d="M%d %dH%d" stroke="#bbcab9" fill="none"/>`, inset, historyY+90, width-inset)
	if len(s.BestFitnessHistory) > 0 {
		var points strings.Builder
		for i, v := range s.BestFitnessHistory {
			x := inset
			if len(s.BestFitnessHistory) > 1 {
				x += i * (width - 2*inset) / (len(s.BestFitnessHistory) - 1)
			}
			y := historyY + 90 - v*60/geneLen
			fmt.Fprintf(&points, "%d,%d ", x, y)
		}
		fmt.Fprintf(&b, `<polyline points="%s" fill="none" stroke="#aa92e7" stroke-width="3"/>`, points.String())
	}
	geneticLessonText(&b, inset, historyY+116, "fitness(g) = sum_i [g_i == target_i]", "#eff5ed")
	geneticLessonText(&b, inset, historyY+140, "c1 = p1[:cut] + p2[cut:]", "#8bd5ff")
	geneticLessonText(&b, inset, historyY+164, "c2 = p2[:cut] + p1[cut:]", "#8bd5ff")
	status := s.StopReason
	if status == "" {
		status = fmt.Sprintf("Mutation rate %.2f · check = target match", s.MutationRate)
	}
	geneticLessonText(&b, inset, historyY+188, status, "#ffe29a")
	fmt.Fprint(&b, `</g>`)
	return b.String()
}
