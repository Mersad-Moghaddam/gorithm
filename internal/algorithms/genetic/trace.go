package genetic

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func geneticPrintable(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 32 || s[i] > 126 {
			return false
		}
	}
	return true
}

func geneticValidate(cfg Config) error {
	if len(cfg.Target) < 1 || len(cfg.Target) > 24 || !geneticPrintable(cfg.Target) {
		return fmt.Errorf("target must contain 1–24 printable ASCII bytes")
	}
	if cfg.PopSize < 1 || cfg.PopSize > 24 || cfg.MaxGenerations < 1 || cfg.MaxGenerations > 100 {
		return fmt.Errorf("population must be 1–24 and generations 1–100")
	}
	if math.IsNaN(cfg.MutationRate) || math.IsInf(cfg.MutationRate, 0) || cfg.MutationRate <= 0 || cfg.MutationRate > 1 {
		return fmt.Errorf("mutationRate must be finite and in (0,1]")
	}
	if cfg.TournamentSize < 1 || cfg.TournamentSize > 24 || !geneticPrintable(cfg.Charset) {
		return fmt.Errorf("invalid tournament or non-ASCII charset")
	}
	return nil
}

// geneticSelectIndex mirrors SelectTournament's draws and strict first-winner tie
// rule, retaining the actual winner's identity even for identical chromosomes.
func geneticSelectIndex(pop []renderers.GeneticIndividualView, k int, rng *rand.Rand) int {
	if k > len(pop) {
		k = len(pop)
	}
	best := rng.Intn(len(pop))
	for i := 1; i < k; i++ {
		j := rng.Intn(len(pop))
		if pop[j].Fitness > pop[best].Fitness {
			best = j
		}
	}
	return best
}

// Trace records real operators without additional RNG draws for presentation.
func Trace(cfg Config) []trace.Step {
	if cfg.Target == "" {
		cfg.Target = "GENETIC"
	}
	if cfg.PopSize == 0 {
		cfg.PopSize = 8
	}
	if cfg.MaxGenerations == 0 {
		cfg.MaxGenerations = 15
	}
	if cfg.MutationRate == 0 {
		cfg.MutationRate = .08
	}
	if cfg.TournamentSize == 0 {
		cfg.TournamentSize = 3
	}
	if cfg.Charset == "" {
		cfg.Charset = DefaultCharset(cfg.Target)
	}
	b := trace.NewBuilder()
	if err := geneticValidate(cfg); err != nil {
		b.AddStep("L1", trace.ActionDone, err.Error(), renderers.GeneticLessonState{Phase: "invalid-input", StopReason: "invalid-input"}, nil)
		return b.Steps()
	}
	if cfg.RNG == nil {
		cfg.RNG = rand.New(rand.NewSource(42))
	}
	s := renderers.GeneticLessonState{Target: cfg.Target, MaxGenerations: cfg.MaxGenerations, MutationRate: cfg.MutationRate, BestIndex: 0, Population: make([]renderers.GeneticIndividualView, cfg.PopSize)}
	best := renderers.GeneticIndividualView{}
	emit := func(line string, action trace.ActionType, phase, reason string) {
		s.Phase = phase
		snap := s
		snap.Population = append([]renderers.GeneticIndividualView(nil), s.Population...)
		snap.NextPopulation = append([]renderers.GeneticIndividualView(nil), s.NextPopulation...)
		snap.BestFitnessHistory = append([]int(nil), s.BestFitnessHistory...)
		for j := range s.ChangedLoci {
			snap.ChangedLoci[j] = append([]int(nil), s.ChangedLoci[j]...)
		}
		snap.TableState = buildTableState(snap.Population, cfg.Target, s.BestIndex)
		b.AddStep(line, action, reason, snap, map[string]any{"generation": s.Generation, "bestFitness": best.Fitness, "bestGenes": best.Genes, "target": cfg.Target, "popSize": cfg.PopSize})
	}
	for i := range s.Population {
		s.Population[i] = renderers.GeneticIndividualView{ID: fmt.Sprintf("g0-i%d", i), Genes: strings.Repeat("-", len(cfg.Target))}
	}
	emit("L1", trace.ActionInit, "start", "Each slot is a distinct candidate; fitness counts target-byte matches.")
	for i := range s.Population {
		s.Population[i].Genes = RandomIndividual(len(cfg.Target), cfg.Charset, cfg.RNG).Genes
		b.IncOperations()
	}
	emit("L2", trace.ActionInit, "initialization", "Random population created using the fixed input seed.")
	for i := range s.Population {
		s.Population[i].Fitness = CalculateFitness(s.Population[i].Genes, cfg.Target)
		b.IncComparisons()
		if s.Population[i].Fitness > s.Population[s.BestIndex].Fitness {
			s.BestIndex = i
		}
	}
	best = s.Population[s.BestIndex]
	s.BestFitnessHistory = append(s.BestFitnessHistory, best.Fitness)
	emit("L3", trace.ActionCompare, "fitness", fmt.Sprintf("Best candidate %s matches %d/%d target bytes.", best.ID, best.Fitness, len(cfg.Target)))
	finish := func() bool {
		if best.Fitness != len(cfg.Target) {
			return false
		}
		s.StopReason = "target-found"
		emit("L6", trace.ActionDone, "complete", "Exact target found; no further reproduction.")
		return true
	}
	if finish() {
		return b.Steps()
	}
	for gen := 1; gen <= cfg.MaxGenerations; gen++ {
		s.Generation = gen
		s.NextPopulation = nil
		s.Parents = [2]renderers.GeneticIndividualView{}
		s.Crossed = [2]renderers.GeneticIndividualView{}
		s.Mutated = [2]renderers.GeneticIndividualView{}
		s.ChangedLoci = [2][]int{}
		s.Pair = 0
		s.Cut = 0
		emit("L4", trace.ActionVisit, "generation", fmt.Sprintf("Generation %d of %d: preserve best, then breed remaining slots.", gen, cfg.MaxGenerations))
		for i := 1; i < len(s.Population); i++ {
			b.IncComparisons()
		}
		emit("L5", trace.ActionHighlight, "best", fmt.Sprintf("Elite %s has fitness %d/%d.", best.ID, best.Fitness, len(cfg.Target)))
		s.EliteID = best.ID
		s.NextPopulation = append(s.NextPopulation, best)
		emit("L7", trace.ActionSet, "elitism", "Copy the elite unchanged into next generation; its identity survives.")
		for len(s.NextPopulation) < cfg.PopSize {
			s.Pair++
			s.Cut = 0
			s.Crossed = [2]renderers.GeneticIndividualView{}
			s.Mutated = [2]renderers.GeneticIndividualView{}
			s.ChangedLoci = [2][]int{}
			s.UsedOffspring = [2]bool{}
			for j := range s.Parents {
				s.Parents[j] = s.Population[geneticSelectIndex(s.Population, cfg.TournamentSize, cfg.RNG)]
			}
			b.IncComparisons()
			emit("L9", trace.ActionCompare, "selection", fmt.Sprintf("Pair %d: tournament winners %s and %s (k=%d).", s.Pair, s.Parents[0].ID, s.Parents[1].ID, cfg.TournamentSize))
			s.Cut = 1
			if len(cfg.Target) > 1 {
				s.Cut = cfg.RNG.Intn(len(cfg.Target)-1) + 1
			}
			a, c := Crossover(Individual{Genes: s.Parents[0].Genes}, Individual{Genes: s.Parents[1].Genes}, s.Cut)
			slot := len(s.NextPopulation)
			for j, child := range []Individual{a, c} {
				s.Crossed[j] = renderers.GeneticIndividualView{ID: fmt.Sprintf("g%d-i%d", gen, slot+j), Genes: child.Genes, Fitness: CalculateFitness(child.Genes, cfg.Target)}
			}
			b.IncSwaps()
			emit("L10", trace.ActionSwap, "crossover", fmt.Sprintf("Cut at byte %d: exchange suffixes, preserving parent segment identity.", s.Cut))
			for j, child := range []Individual{a, c} {
				s.Mutated[j] = s.Crossed[j]
				// ponytail: preserve historical first-pair second-child draw even when full;
				// later pairs mutate second child only when capacity permits.
				if j == 0 || s.Pair == 1 || len(s.NextPopulation) < cfg.PopSize {
					changed := Mutate(child, cfg.MutationRate, cfg.Charset, cfg.RNG)
					s.Mutated[j].Genes = changed.Genes
					s.Mutated[j].Fitness = CalculateFitness(changed.Genes, cfg.Target)
					for k := range child.Genes {
						if child.Genes[k] != changed.Genes[k] {
							s.ChangedLoci[j] = append(s.ChangedLoci[j], k)
						}
					}
				}
				if len(s.NextPopulation) < cfg.PopSize {
					s.NextPopulation = append(s.NextPopulation, s.Mutated[j])
					s.UsedOffspring[j] = true
				}
			}
			b.IncOperations()
			emit("L11", trace.ActionSet, "mutation", fmt.Sprintf("Mutation rate %.2f; append only available slots. Gold marks observed byte changes, not unchanged replacements.", cfg.MutationRate))
		}
		s.Population = append([]renderers.GeneticIndividualView(nil), s.NextPopulation...)
		s.NextPopulation = nil
		s.BestIndex = -1
		emit("L13", trace.ActionSet, "replacement", "Next generation replaces population, including unchanged elite.")
		s.BestIndex = 0
		for i := range s.Population {
			b.IncComparisons()
			if s.Population[i].Fitness > s.Population[s.BestIndex].Fitness {
				s.BestIndex = i
			}
		}
		best = s.Population[s.BestIndex]
		s.BestFitnessHistory = append(s.BestFitnessHistory, best.Fitness)
		emit("L14", trace.ActionCompare, "fitness", fmt.Sprintf("Best fitness %d/%d. Elitism protects best-so-far, not every child.", best.Fitness, len(cfg.Target)))
		if finish() {
			return b.Steps()
		}
	}
	s.StopReason = "generation-limit"
	emit("L16", trace.ActionDone, "complete", fmt.Sprintf("Generation budget exhausted; best fitness %d/%d, not an exact solution.", best.Fitness, len(cfg.Target)))
	return b.Steps()
}

func buildTableState(pop []renderers.GeneticIndividualView, target string, best int) renderers.TableState {
	s := renderers.TableState{ColHeaders: []string{"Chromosome", "Fitness"}}
	for i, p := range pop {
		s.RowHeaders = append(s.RowHeaders, p.ID)
		state := "normal"
		if i == best {
			state = "active"
		}
		s.Cells = append(s.Cells, renderers.TableCell{Row: i, Col: 0, Value: p.Genes, State: state}, renderers.TableCell{Row: i, Col: 1, Value: fmt.Sprint(p.Fitness), State: state})
	}
	return s
}
