package genetic

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// cloneTableState returns an independent deep copy of TableState and its slice fields.
func cloneTableState(ts renderers.TableState) renderers.TableState {
	var rowHeaders []string
	if ts.RowHeaders != nil {
		rowHeaders = make([]string, len(ts.RowHeaders))
		copy(rowHeaders, ts.RowHeaders)
	}

	var colHeaders []string
	if ts.ColHeaders != nil {
		colHeaders = make([]string, len(ts.ColHeaders))
		copy(colHeaders, ts.ColHeaders)
	}

	var cells []renderers.TableCell
	if ts.Cells != nil {
		cells = make([]renderers.TableCell, len(ts.Cells))
		copy(cells, ts.Cells)
	}

	var deps []renderers.TableDependency
	if ts.Dependencies != nil {
		deps = make([]renderers.TableDependency, len(ts.Dependencies))
		copy(deps, ts.Dependencies)
	}

	return renderers.TableState{
		Caption:      ts.Caption,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
		Dependencies: deps,
	}
}

// buildPopulationTableState formats the population into a TableState for visualization.
func buildPopulationTableState(pop []Individual, target string, bestIdx, p1Idx, p2Idx int, caption string) renderers.TableState {
	targetLen := len(target)
	popSize := len(pop)

	rowHeaders := make([]string, popSize)
	for i := 0; i < popSize; i++ {
		if i == p1Idx && i == p2Idx {
			rowHeaders[i] = fmt.Sprintf("Ind %d (P1&P2)", i)
		} else if i == p1Idx {
			rowHeaders[i] = fmt.Sprintf("Ind %d (P1)", i)
		} else if i == p2Idx {
			rowHeaders[i] = fmt.Sprintf("Ind %d (P2)", i)
		} else {
			rowHeaders[i] = fmt.Sprintf("Ind %d", i)
		}
	}

	if targetLen <= 10 {
		colHeaders := make([]string, targetLen+1)
		for c := 0; c < targetLen; c++ {
			colHeaders[c] = fmt.Sprintf("G%d", c)
		}
		colHeaders[targetLen] = "Fitness"

		cells := make([]renderers.TableCell, 0, popSize*(targetLen+1))
		for r := 0; r < popSize; r++ {
			ind := pop[r]
			genes := ind.Genes

			for c := 0; c < targetLen; c++ {
				ch := ""
				if c < len(genes) {
					ch = string(genes[c])
				}
				cellState := "normal"
				if c < len(genes) && c < len(target) && genes[c] == target[c] {
					cellState = "computed"
				} else if r == p1Idx {
					cellState = "source"
				} else if r == p2Idx {
					cellState = "active"
				} else if r == bestIdx {
					cellState = "source"
				}

				cells = append(cells, renderers.TableCell{
					Row:   r,
					Col:   c,
					Value: ch,
					State: cellState,
				})
			}

			// Fitness column
			fitState := "normal"
			if ind.Fitness == targetLen {
				fitState = "computed"
			} else if r == p1Idx {
				fitState = "source"
			} else if r == p2Idx {
				fitState = "active"
			} else if r == bestIdx {
				fitState = "active"
			}
			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   targetLen,
				Value: strconv.Itoa(ind.Fitness),
				State: fitState,
			})
		}

		return renderers.TableState{
			Caption:    caption,
			RowHeaders: rowHeaders,
			ColHeaders: colHeaders,
			Cells:      cells,
		}
	}

	// For longer sequences (> 10 characters), use 2-column layout
	colHeaders := []string{"Chromosome", "Fitness"}
	cells := make([]renderers.TableCell, 0, popSize*2)
	for r := 0; r < popSize; r++ {
		ind := pop[r]
		chromState := "normal"
		if ind.Fitness == targetLen {
			chromState = "computed"
		} else if r == p1Idx {
			chromState = "source"
		} else if r == p2Idx {
			chromState = "active"
		} else if r == bestIdx {
			chromState = "source"
		}

		cells = append(cells, renderers.TableCell{
			Row:   r,
			Col:   0,
			Value: ind.Genes,
			State: chromState,
		})

		fitState := "normal"
		if ind.Fitness == targetLen {
			fitState = "computed"
		} else if r == p1Idx {
			fitState = "source"
		} else if r == p2Idx {
			fitState = "active"
		} else if r == bestIdx {
			fitState = "active"
		}
		cells = append(cells, renderers.TableCell{
			Row:   r,
			Col:   1,
			Value: strconv.Itoa(ind.Fitness),
			State: fitState,
		})
	}

	return renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// buildTableState is a backward-compatible wrapper for buildPopulationTableState.
func buildTableState(pop []Individual, target string, bestIdx int, highlightIdx int) renderers.TableState {
	return buildPopulationTableState(pop, target, bestIdx, highlightIdx, -1, "")
}

// buildCrossoverTableState formats parents and their crossover offspring.
func buildCrossoverTableState(p1, p2, c1, c2 Individual, p1Idx, p2Idx, cut int, target, caption string) renderers.TableState {
	targetLen := len(target)
	rowHeaders := []string{
		fmt.Sprintf("Parent 1 (Ind %d)", p1Idx),
		fmt.Sprintf("Parent 2 (Ind %d)", p2Idx),
		"Child 1 (P1+P2)",
		"Child 2 (P2+P1)",
	}

	if targetLen <= 10 {
		colHeaders := make([]string, targetLen+1)
		for c := 0; c < targetLen; c++ {
			colHeaders[c] = fmt.Sprintf("G%d", c)
		}
		colHeaders[targetLen] = "Fitness"

		cells := make([]renderers.TableCell, 0, 4*(targetLen+1))
		rowsData := []struct {
			genes   string
			fitness int
			isP1    bool
			isP2    bool
			isC1    bool
			isC2    bool
		}{
			{genes: p1.Genes, fitness: p1.Fitness, isP1: true},
			{genes: p2.Genes, fitness: p2.Fitness, isP2: true},
			{genes: c1.Genes, fitness: c1.Fitness, isC1: true},
			{genes: c2.Genes, fitness: c2.Fitness, isC2: true},
		}

		for r, row := range rowsData {
			for c := 0; c < targetLen; c++ {
				ch := ""
				if c < len(row.genes) {
					ch = string(row.genes[c])
				}
				cellState := "normal"
				if row.isP1 {
					if c < cut {
						cellState = "source" // Contributed to Child 1
					} else {
						cellState = "normal"
					}
				} else if row.isP2 {
					if c >= cut {
						cellState = "active" // Contributed to Child 1
					} else {
						cellState = "normal"
					}
				} else if row.isC1 {
					if c < cut {
						cellState = "source" // From P1
					} else {
						cellState = "active" // From P2
					}
				} else if row.isC2 {
					if c < cut {
						cellState = "active" // From P2
					} else {
						cellState = "source" // From P1
					}
				}
				if c < len(row.genes) && c < len(target) && row.genes[c] == target[c] && cellState == "normal" {
					cellState = "computed"
				}

				cells = append(cells, renderers.TableCell{
					Row:   r,
					Col:   c,
					Value: ch,
					State: cellState,
				})
			}

			fitState := "normal"
			if row.fitness == targetLen {
				fitState = "computed"
			} else if row.isP1 {
				fitState = "source"
			} else if row.isP2 {
				fitState = "active"
			}
			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   targetLen,
				Value: strconv.Itoa(row.fitness),
				State: fitState,
			})
		}

		return renderers.TableState{
			Caption:    caption,
			RowHeaders: rowHeaders,
			ColHeaders: colHeaders,
			Cells:      cells,
		}
	}

	colHeaders := []string{"Chromosome", "Fitness"}
	cells := []renderers.TableCell{
		{Row: 0, Col: 0, Value: p1.Genes, State: "source"},
		{Row: 0, Col: 1, Value: strconv.Itoa(p1.Fitness), State: "source"},
		{Row: 1, Col: 0, Value: p2.Genes, State: "active"},
		{Row: 1, Col: 1, Value: strconv.Itoa(p2.Fitness), State: "active"},
		{Row: 2, Col: 0, Value: c1.Genes, State: "source"},
		{Row: 2, Col: 1, Value: strconv.Itoa(c1.Fitness), State: "normal"},
		{Row: 3, Col: 0, Value: c2.Genes, State: "active"},
		{Row: 3, Col: 1, Value: strconv.Itoa(c2.Fitness), State: "normal"},
	}

	return renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// buildMutationTableState formats children before and after mutation with visual indicators.
func buildMutationTableState(preC1, postC1, preC2, postC2 Individual, mut1, mut2 []int, target, caption string) renderers.TableState {
	targetLen := len(target)
	rowHeaders := []string{
		"Child 1 (Pre)",
		"Child 1 (Mutated)",
		"Child 2 (Pre)",
		"Child 2 (Mutated)",
	}

	isMutated := func(idx int, list []int) bool {
		for _, m := range list {
			if m == idx {
				return true
			}
		}
		return false
	}

	if targetLen <= 10 {
		colHeaders := make([]string, targetLen+1)
		for c := 0; c < targetLen; c++ {
			colHeaders[c] = fmt.Sprintf("G%d", c)
		}
		colHeaders[targetLen] = "Fitness"

		cells := make([]renderers.TableCell, 0, 4*(targetLen+1))
		rowsData := []struct {
			genes   string
			fitness int
			mutated []int
			isPost  bool
		}{
			{genes: preC1.Genes, fitness: preC1.Fitness, isPost: false},
			{genes: postC1.Genes, fitness: postC1.Fitness, mutated: mut1, isPost: true},
			{genes: preC2.Genes, fitness: preC2.Fitness, isPost: false},
			{genes: postC2.Genes, fitness: postC2.Fitness, mutated: mut2, isPost: true},
		}

		for r, row := range rowsData {
			for c := 0; c < targetLen; c++ {
				ch := ""
				if c < len(row.genes) {
					ch = string(row.genes[c])
				}
				cellState := "normal"
				val := ch
				if row.isPost && isMutated(c, row.mutated) {
					cellState = "active" // Visual indicator: bright glowing cell
					val = ch + "*"       // Visual indicator: asterisk symbol
				} else if c < len(row.genes) && c < len(target) && row.genes[c] == target[c] {
					cellState = "computed"
				}

				cells = append(cells, renderers.TableCell{
					Row:   r,
					Col:   c,
					Value: val,
					State: cellState,
				})
			}

			fitState := "normal"
			if row.fitness == targetLen {
				fitState = "computed"
			}
			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   targetLen,
				Value: strconv.Itoa(row.fitness),
				State: fitState,
			})
		}

		return renderers.TableState{
			Caption:    caption,
			RowHeaders: rowHeaders,
			ColHeaders: colHeaders,
			Cells:      cells,
		}
	}

	colHeaders := []string{"Chromosome", "Fitness"}
	cells := []renderers.TableCell{
		{Row: 0, Col: 0, Value: preC1.Genes, State: "normal"},
		{Row: 0, Col: 1, Value: strconv.Itoa(preC1.Fitness), State: "normal"},
		{Row: 1, Col: 0, Value: postC1.Genes, State: func() string {
			if len(mut1) > 0 {
				return "active"
			}
			return "normal"
		}()},
		{Row: 1, Col: 1, Value: strconv.Itoa(postC1.Fitness), State: "normal"},
		{Row: 2, Col: 0, Value: preC2.Genes, State: "normal"},
		{Row: 2, Col: 1, Value: strconv.Itoa(preC2.Fitness), State: "normal"},
		{Row: 3, Col: 0, Value: postC2.Genes, State: func() string {
			if len(mut2) > 0 {
				return "active"
			}
			return "normal"
		}()},
		{Row: 3, Col: 1, Value: strconv.Itoa(postC2.Fitness), State: "normal"},
	}

	return renderers.TableState{
		Caption:    caption,
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace produces a step-by-step visual execution of the genetic algorithm.
func Trace(cfg Config) []trace.Step {
	if cfg.Target == "" {
		cfg.Target = "GENETIC"
	}
	if cfg.PopSize <= 0 {
		cfg.PopSize = 8
	}
	if cfg.MutationRate < 0 {
		cfg.MutationRate = 0.08
	}
	if cfg.MaxGenerations <= 0 {
		cfg.MaxGenerations = 15
	}
	if cfg.TournamentSize <= 0 {
		cfg.TournamentSize = 3
	}
	if cfg.Charset == "" {
		cfg.Charset = DefaultCharset(cfg.Target)
	}
	if cfg.RNG == nil {
		cfg.RNG = rand.New(rand.NewSource(42))
	}

	b := trace.NewBuilder()
	targetLen := len(cfg.Target)

	vars := func(gen, bestFitness int, bestGenes string) map[string]any {
		return map[string]any{
			"generation":  gen,
			"bestFitness": bestFitness,
			"bestGenes":   bestGenes,
			"target":      cfg.Target,
			"popSize":     cfg.PopSize,
		}
	}

	// L1: Start
	placeholderPop := make([]Individual, cfg.PopSize)
	for i := 0; i < cfg.PopSize; i++ {
		placeholderPop[i] = Individual{Genes: strings.Repeat("-", targetLen), Fitness: 0}
	}
	initCaption := fmt.Sprintf("Gen 0: Target '%s' (length %d). Initializing population of %d.", cfg.Target, targetLen, cfg.PopSize)
	emptyState := buildPopulationTableState(placeholderPop, cfg.Target, -1, -1, -1, initCaption)
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Starting Genetic Algorithm targeting %q (population: %d, max generations: %d).", cfg.Target, cfg.PopSize, cfg.MaxGenerations),
		cloneTableState(emptyState), vars(0, 0, ""))

	// L2: Initialize Population
	pop := make([]Individual, cfg.PopSize)
	for i := 0; i < cfg.PopSize; i++ {
		pop[i] = RandomIndividual(targetLen, cfg.Charset, cfg.RNG)
		b.IncOperations()
	}
	l2Caption := fmt.Sprintf("Gen 0: Initialized random population (%d individuals).", cfg.PopSize)
	initState := buildPopulationTableState(pop, cfg.Target, -1, -1, -1, l2Caption)
	b.AddStep("L2", trace.ActionInit,
		fmt.Sprintf("Initialized random population of %d individuals.", cfg.PopSize),
		cloneTableState(initState), vars(0, 0, ""))

	// L3: Initial Fitness Evaluation
	bestIdx := 0
	for i := 0; i < cfg.PopSize; i++ {
		pop[i].Fitness = CalculateFitness(pop[i].Genes, cfg.Target)
		b.IncComparisons()
		if pop[i].Fitness > pop[bestIdx].Fitness {
			bestIdx = i
		}
	}
	bestOverall := pop[bestIdx]
	l3Caption := fmt.Sprintf("Gen 0: Best candidate '%s' (fitness %d/%d). Initial evaluation.", bestOverall.Genes, bestOverall.Fitness, targetLen)
	evalState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, l3Caption)
	b.AddStep("L3", trace.ActionCompare,
		fmt.Sprintf("Evaluated initial population fitness against target %q. Best candidate #%d %q with fitness %d/%d.", cfg.Target, bestIdx, bestOverall.Genes, bestOverall.Fitness, targetLen),
		cloneTableState(evalState), vars(0, bestOverall.Fitness, bestOverall.Genes))

	// Check if already solved
	if bestOverall.Fitness == targetLen {
		doneCaption := fmt.Sprintf("Gen 0: Best candidate '%s' (fitness %d/%d). Target achieved!", bestOverall.Genes, bestOverall.Fitness, targetLen)
		doneState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, doneCaption)
		b.AddStep("L6", trace.ActionDone,
			fmt.Sprintf("Optimal solution %q found in initial generation!", bestOverall.Genes),
			cloneTableState(doneState), vars(0, bestOverall.Fitness, bestOverall.Genes))
		return b.Steps()
	}

	// Evolution loop
	for gen := 0; gen < cfg.MaxGenerations; gen++ {
		genCaption := fmt.Sprintf("Gen %d: Best candidate '%s' (fitness %d/%d). Crossover & mutation producing next gen.", gen+1, bestOverall.Genes, bestOverall.Fitness, targetLen)

		// L4: Generation start
		genState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, genCaption)
		b.AddStep("L4", trace.ActionVisit,
			fmt.Sprintf("Beginning Generation %d / %d. Best candidate so far: %q (%d/%d).", gen+1, cfg.MaxGenerations, bestOverall.Genes, bestOverall.Fitness, targetLen),
			cloneTableState(genState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L5: Find best candidate of current generation
		bestIdx = 0
		for i := 1; i < len(pop); i++ {
			b.IncComparisons()
			if pop[i].Fitness > pop[bestIdx].Fitness {
				bestIdx = i
			}
		}
		if pop[bestIdx].Fitness > bestOverall.Fitness {
			bestOverall = pop[bestIdx]
			genCaption = fmt.Sprintf("Gen %d: Best candidate '%s' (fitness %d/%d). Crossover & mutation producing next gen.", gen+1, bestOverall.Genes, bestOverall.Fitness, targetLen)
		}
		hlState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, genCaption)
		b.AddStep("L5", trace.ActionHighlight,
			fmt.Sprintf("Generation %d fittest candidate #%d: %q (fitness %d/%d).", gen+1, bestIdx, pop[bestIdx].Genes, pop[bestIdx].Fitness, targetLen),
			cloneTableState(hlState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L6: Check termination
		if bestOverall.Fitness == targetLen {
			doneCaption := fmt.Sprintf("Gen %d: Best candidate '%s' (fitness %d/%d). Target achieved!", gen+1, bestOverall.Genes, bestOverall.Fitness, targetLen)
			doneState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, doneCaption)
			b.AddStep("L6", trace.ActionDone,
				fmt.Sprintf("Target sequence %q achieved at Generation %d!", bestOverall.Genes, gen+1),
				cloneTableState(doneState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))
			return b.Steps()
		}

		// Next generation container with elitism
		nextGen := make([]Individual, 0, cfg.PopSize)
		nextGen = append(nextGen, bestOverall)

		// L9: Selection (tournament selection)
		p1Idx := SelectTournamentIndex(pop, cfg.TournamentSize, cfg.RNG)
		p2Idx := SelectTournamentIndex(pop, cfg.TournamentSize, cfg.RNG)
		if p1Idx == p2Idx && len(pop) > 1 {
			for attempts := 0; attempts < 5 && p2Idx == p1Idx; attempts++ {
				p2Idx = SelectTournamentIndex(pop, cfg.TournamentSize, cfg.RNG)
			}
		}
		p1 := pop[p1Idx]
		p2 := pop[p2Idx]
		b.IncComparisons()

		selState := buildPopulationTableState(pop, cfg.Target, bestIdx, p1Idx, p2Idx, genCaption)
		b.AddStep("L9", trace.ActionCompare,
			fmt.Sprintf("Tournament selection (k=%d): Selected Parent 1 (Ind %d: %q, fitness %d/%d) and Parent 2 (Ind %d: %q, fitness %d/%d).", cfg.TournamentSize, p1Idx, p1.Genes, p1.Fitness, targetLen, p2Idx, p2.Genes, p2.Fitness, targetLen),
			cloneTableState(selState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L10: Crossover
		cut := 1
		if targetLen > 1 {
			cut = cfg.RNG.Intn(targetLen-1) + 1
		}
		c1, c2 := Crossover(p1, p2, cut)
		c1.Fitness = CalculateFitness(c1.Genes, cfg.Target)
		c2.Fitness = CalculateFitness(c2.Genes, cfg.Target)
		b.IncSwaps()

		crossState := buildCrossoverTableState(p1, p2, c1, c2, p1Idx, p2Idx, cut, cfg.Target, genCaption)
		b.AddStep("L10", trace.ActionSwap,
			fmt.Sprintf("Single-point crossover at split %d: recombined genetic material into Child 1 (%q) and Child 2 (%q).", cut, c1.Genes, c2.Genes),
			cloneTableState(crossState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L11: Mutation
		preC1 := c1
		preC2 := c2
		postC1, mut1 := MutateTracked(c1, cfg.MutationRate, cfg.Charset, cfg.RNG)
		postC2, mut2 := MutateTracked(c2, cfg.MutationRate, cfg.Charset, cfg.RNG)
		if cfg.MutationRate > 0 && len(mut1) == 0 && len(mut2) == 0 && targetLen > 0 {
			mutIdx := cfg.RNG.Intn(targetLen)
			oldChar := postC1.Genes[mutIdx]
			newChar := cfg.Charset[cfg.RNG.Intn(len(cfg.Charset))]
			if len(cfg.Charset) > 1 && newChar == oldChar {
				newChar = cfg.Charset[(strings.IndexByte(cfg.Charset, oldChar)+1)%len(cfg.Charset)]
			}
			bBytes := []byte(postC1.Genes)
			bBytes[mutIdx] = newChar
			postC1.Genes = string(bBytes)
			mut1 = append(mut1, mutIdx)
		}
		postC1.Fitness = CalculateFitness(postC1.Genes, cfg.Target)
		postC2.Fitness = CalculateFitness(postC2.Genes, cfg.Target)
		b.IncOperations()

		mutState := buildMutationTableState(preC1, postC1, preC2, postC2, mut1, mut2, cfg.Target, genCaption)
		var mutDesc string
		if len(mut1) > 0 || len(mut2) > 0 {
			mutDesc = fmt.Sprintf("Mutation phase (rate %.2f): Gene mutation occurred with visual indicator (*). Offspring 1: %q, Offspring 2: %q.", cfg.MutationRate, postC1.Genes, postC2.Genes)
		} else {
			mutDesc = fmt.Sprintf("Mutation phase (rate %.2f): No genes mutated. Offspring 1: %q, Offspring 2: %q.", cfg.MutationRate, postC1.Genes, postC2.Genes)
		}
		b.AddStep("L11", trace.ActionSet,
			mutDesc,
			cloneTableState(mutState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		nextGen = append(nextGen, postC1)
		if len(nextGen) < cfg.PopSize {
			nextGen = append(nextGen, postC2)
		}

		for len(nextGen) < cfg.PopSize {
			paIdx := SelectTournamentIndex(pop, cfg.TournamentSize, cfg.RNG)
			pbIdx := SelectTournamentIndex(pop, cfg.TournamentSize, cfg.RNG)
			pa, pb := pop[paIdx], pop[pbIdx]
			split := 1
			if targetLen > 1 {
				split = cfg.RNG.Intn(targetLen-1) + 1
			}
			off1, off2 := Crossover(pa, pb, split)
			off1, _ = MutateTracked(off1, cfg.MutationRate, cfg.Charset, cfg.RNG)
			off1.Fitness = CalculateFitness(off1.Genes, cfg.Target)
			nextGen = append(nextGen, off1)

			if len(nextGen) < cfg.PopSize {
				off2, _ = MutateTracked(off2, cfg.MutationRate, cfg.Charset, cfg.RNG)
				off2.Fitness = CalculateFitness(off2.Genes, cfg.Target)
				nextGen = append(nextGen, off2)
			}
		}

		// L13: Update population
		pop = nextGen
		updState := buildPopulationTableState(pop, cfg.Target, -1, -1, -1, genCaption)
		b.AddStep("L13", trace.ActionSet,
			fmt.Sprintf("Updated population to Generation %d (%d offspring ready for evaluation).", gen+1, len(pop)),
			cloneTableState(updState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L14: Evaluate Fitness
		bestIdx = 0
		for i := 0; i < len(pop); i++ {
			b.IncComparisons()
			if pop[i].Fitness > pop[bestIdx].Fitness {
				bestIdx = i
			}
		}
		if pop[bestIdx].Fitness > bestOverall.Fitness {
			bestOverall = pop[bestIdx]
			genCaption = fmt.Sprintf("Gen %d: Best candidate '%s' (fitness %d/%d). Crossover & mutation producing next gen.", gen+1, bestOverall.Genes, bestOverall.Fitness, targetLen)
		}
		newEvalState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, genCaption)
		b.AddStep("L14", trace.ActionCompare,
			fmt.Sprintf("Evaluated Generation %d fitness. Best candidate #%d %q with fitness %d/%d.", gen+1, bestIdx, pop[bestIdx].Genes, pop[bestIdx].Fitness, targetLen),
			cloneTableState(newEvalState), vars(gen+1, bestOverall.Fitness, bestOverall.Genes))
	}

	// L16: Done
	finalCaption := fmt.Sprintf("Gen %d: Best candidate '%s' (fitness %d/%d). Crossover & mutation producing next gen.", cfg.MaxGenerations, bestOverall.Genes, bestOverall.Fitness, targetLen)
	finalState := buildPopulationTableState(pop, cfg.Target, bestIdx, -1, -1, finalCaption)
	b.AddStep("L16", trace.ActionDone,
		fmt.Sprintf("Evolution complete after %d generations. Best candidate: %q with fitness %d/%d.", cfg.MaxGenerations, bestOverall.Genes, bestOverall.Fitness, targetLen),
		cloneTableState(finalState), vars(cfg.MaxGenerations, bestOverall.Fitness, bestOverall.Genes))

	return b.Steps()
}
