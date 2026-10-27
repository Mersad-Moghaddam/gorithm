package genetic

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// buildTableState formats the population into a TableState for visualization.
func buildTableState(pop []Individual, target string, bestIdx int, highlightIdx int) renderers.TableState {
	targetLen := len(target)
	popSize := len(pop)

	rowHeaders := make([]string, popSize)
	for i := 0; i < popSize; i++ {
		rowHeaders[i] = fmt.Sprintf("Ind %d", i)
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
				} else if r == highlightIdx {
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
		if r == highlightIdx {
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
	if cfg.MutationRate <= 0 {
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
	emptyState := buildTableState(placeholderPop, cfg.Target, -1, -1)
	b.AddStep("L1", trace.ActionInit,
		fmt.Sprintf("Starting Genetic Algorithm targeting %q (population: %d, max generations: %d).", cfg.Target, cfg.PopSize, cfg.MaxGenerations),
		emptyState, vars(0, 0, ""))

	// L2: Initialize Population
	pop := make([]Individual, cfg.PopSize)
	for i := 0; i < cfg.PopSize; i++ {
		pop[i] = RandomIndividual(targetLen, cfg.Charset, cfg.RNG)
		b.IncOperations()
	}
	initState := buildTableState(pop, cfg.Target, -1, -1)
	b.AddStep("L2", trace.ActionInit,
		fmt.Sprintf("Initialized random population of %d individuals.", cfg.PopSize),
		initState, vars(0, 0, ""))

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
	evalState := buildTableState(pop, cfg.Target, bestIdx, -1)
	b.AddStep("L3", trace.ActionCompare,
		fmt.Sprintf("Evaluated initial population fitness. Best individual #%d %q with fitness %d/%d.", bestIdx, bestOverall.Genes, bestOverall.Fitness, targetLen),
		evalState, vars(0, bestOverall.Fitness, bestOverall.Genes))

	// Check if already solved
	if bestOverall.Fitness == targetLen {
		doneState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L6", trace.ActionDone,
			fmt.Sprintf("Optimal solution %q found in initial generation!", bestOverall.Genes),
			doneState, vars(0, bestOverall.Fitness, bestOverall.Genes))
		return b.Steps()
	}

	// Evolution loop
	for gen := 0; gen < cfg.MaxGenerations; gen++ {
		// L4: Generation start
		genState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L4", trace.ActionVisit,
			fmt.Sprintf("Beginning Generation %d / %d.", gen+1, cfg.MaxGenerations),
			genState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L5: Find best in current generation
		bestIdx = 0
		for i := 1; i < len(pop); i++ {
			b.IncComparisons()
			if pop[i].Fitness > pop[bestIdx].Fitness {
				bestIdx = i
			}
		}
		if pop[bestIdx].Fitness > bestOverall.Fitness {
			bestOverall = pop[bestIdx]
		}
		hlState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L5", trace.ActionHighlight,
			fmt.Sprintf("Generation %d fittest candidate #%d: %q (fitness %d/%d).", gen+1, bestIdx, pop[bestIdx].Genes, pop[bestIdx].Fitness, targetLen),
			hlState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L6: Check termination
		if bestOverall.Fitness == targetLen {
			doneState := buildTableState(pop, cfg.Target, bestIdx, -1)
			b.AddStep("L6", trace.ActionDone,
				fmt.Sprintf("Target sequence %q achieved at Generation %d!", bestOverall.Genes, gen+1),
				doneState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))
			return b.Steps()
		}

		// L7 & L8: Prepare next generation
		nextGen := make([]Individual, 0, cfg.PopSize)
		// Elitism
		nextGen = append(nextGen, bestOverall)

		// L9: Selection
		p1 := SelectTournament(pop, cfg.TournamentSize, cfg.RNG)
		p2 := SelectTournament(pop, cfg.TournamentSize, cfg.RNG)
		b.IncComparisons()
		selState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L9", trace.ActionCompare,
			fmt.Sprintf("Selected parent candidates %q and %q using tournament selection (k=%d).", p1.Genes, p2.Genes, cfg.TournamentSize),
			selState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L10: Crossover
		cut := 1
		if targetLen > 1 {
			cut = cfg.RNG.Intn(targetLen-1) + 1
		}
		c1, c2 := Crossover(p1, p2, cut)
		b.IncSwaps()
		crossState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L10", trace.ActionSwap,
			fmt.Sprintf("Single-point crossover at split point %d created offspring %q and %q.", cut, c1.Genes, c2.Genes),
			crossState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// L11: Mutation
		c1 = Mutate(c1, cfg.MutationRate, cfg.Charset, cfg.RNG)
		c2 = Mutate(c2, cfg.MutationRate, cfg.Charset, cfg.RNG)
		b.IncOperations()
		mutState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L11", trace.ActionSet,
			fmt.Sprintf("Mutated offspring with rate %.2f: %q and %q.", cfg.MutationRate, c1.Genes, c2.Genes),
			mutState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

		// Populate rest of next generation
		c1.Fitness = CalculateFitness(c1.Genes, cfg.Target)
		nextGen = append(nextGen, c1)
		if len(nextGen) < cfg.PopSize {
			c2.Fitness = CalculateFitness(c2.Genes, cfg.Target)
			nextGen = append(nextGen, c2)
		}

		for len(nextGen) < cfg.PopSize {
			pa := SelectTournament(pop, cfg.TournamentSize, cfg.RNG)
			pb := SelectTournament(pop, cfg.TournamentSize, cfg.RNG)
			split := 1
			if targetLen > 1 {
				split = cfg.RNG.Intn(targetLen-1) + 1
			}
			off1, off2 := Crossover(pa, pb, split)
			off1 = Mutate(off1, cfg.MutationRate, cfg.Charset, cfg.RNG)
			off1.Fitness = CalculateFitness(off1.Genes, cfg.Target)
			nextGen = append(nextGen, off1)

			if len(nextGen) < cfg.PopSize {
				off2 = Mutate(off2, cfg.MutationRate, cfg.Charset, cfg.RNG)
				off2.Fitness = CalculateFitness(off2.Genes, cfg.Target)
				nextGen = append(nextGen, off2)
			}
		}

		// L13: Update population
		pop = nextGen
		updState := buildTableState(pop, cfg.Target, -1, -1)
		b.AddStep("L13", trace.ActionSet,
			fmt.Sprintf("Updated population to Generation %d offspring.", gen+1),
			updState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))

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
		}
		newEvalState := buildTableState(pop, cfg.Target, bestIdx, -1)
		b.AddStep("L14", trace.ActionCompare,
			fmt.Sprintf("Re-evaluated Generation %d fitness. Best fitness now %d/%d.", gen+1, bestOverall.Fitness, targetLen),
			newEvalState, vars(gen+1, bestOverall.Fitness, bestOverall.Genes))
	}

	// L16: Done
	finalState := buildTableState(pop, cfg.Target, bestIdx, -1)
	b.AddStep("L16", trace.ActionDone,
		fmt.Sprintf("Evolution complete after %d generations. Best candidate: %q with fitness %d/%d.", cfg.MaxGenerations, bestOverall.Genes, bestOverall.Fitness, targetLen),
		finalState, vars(cfg.MaxGenerations, bestOverall.Fitness, bestOverall.Genes))

	return b.Steps()
}
