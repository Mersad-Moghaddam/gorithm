package genetic

import (
	"math/rand"
	"strings"
	"time"
)

// Individual represents a candidate solution in the population.
type Individual struct {
	Genes   string `json:"genes"`
	Fitness int    `json:"fitness"`
}

// Config controls parameters for the genetic algorithm.
type Config struct {
	Target         string     `json:"target"`
	PopSize        int        `json:"popSize"`
	MutationRate   float64    `json:"mutationRate"`
	MaxGenerations int        `json:"maxGenerations"`
	TournamentSize int        `json:"tournamentSize"`
	Charset        string     `json:"charset"`
	RNG            *rand.Rand `json:"-"`
}

// GenerationInfo stores population state at a given generation.
type GenerationInfo struct {
	Generation  int          `json:"generation"`
	Population  []Individual `json:"population"`
	Best        Individual   `json:"best"`
	BestFitness int          `json:"bestFitness"`
}

// Result holds the final outcome and generation history of evolution.
type Result struct {
	Best    Individual       `json:"best"`
	History []GenerationInfo `json:"history"`
}

// CalculateFitness returns the number of genes matching target characters.
func CalculateFitness(genes, target string) int {
	score := 0
	limit := len(genes)
	if len(target) < limit {
		limit = len(target)
	}
	for i := 0; i < limit; i++ {
		if genes[i] == target[i] {
			score++
		}
	}
	return score
}

// DefaultCharset returns a suitable character set for the target string.
func DefaultCharset(target string) string {
	if len(target) == 0 {
		return "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}

	isBinary := true
	for i := 0; i < len(target); i++ {
		if target[i] != '0' && target[i] != '1' {
			isBinary = false
			break
		}
	}
	if isBinary {
		return "01"
	}

	base := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for _, r := range target {
		ch := string(r)
		if !strings.Contains(base, ch) {
			base += ch
		}
	}
	return base
}

// RandomIndividual generates an individual with random genes from charset.
func RandomIndividual(length int, charset string, rng *rand.Rand) Individual {
	if len(charset) == 0 {
		charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return Individual{Genes: string(b)}
}

// SelectTournament selects the best individual from a random subset of size k.
func SelectTournament(pop []Individual, k int, rng *rand.Rand) Individual {
	if len(pop) == 0 {
		return Individual{}
	}
	if k <= 1 || len(pop) == 1 {
		return pop[rng.Intn(len(pop))]
	}
	if k > len(pop) {
		k = len(pop)
	}

	best := pop[rng.Intn(len(pop))]
	for i := 1; i < k; i++ {
		contender := pop[rng.Intn(len(pop))]
		if contender.Fitness > best.Fitness {
			best = contender
		}
	}
	return best
}

// SelectTournamentIndex selects the best individual from a random subset of size k and returns its index.
func SelectTournamentIndex(pop []Individual, k int, rng *rand.Rand) int {
	if len(pop) == 0 {
		return -1
	}
	if k <= 1 || len(pop) == 1 {
		return rng.Intn(len(pop))
	}
	if k > len(pop) {
		k = len(pop)
	}

	bestIdx := rng.Intn(len(pop))
	for i := 1; i < k; i++ {
		contenderIdx := rng.Intn(len(pop))
		if pop[contenderIdx].Fitness > pop[bestIdx].Fitness {
			bestIdx = contenderIdx
		}
	}
	return bestIdx
}

// SelectRoulette selects an individual with probability proportional to fitness.
func SelectRoulette(pop []Individual, rng *rand.Rand) Individual {
	if len(pop) == 0 {
		return Individual{}
	}
	totalFitness := 0
	for _, ind := range pop {
		if ind.Fitness > 0 {
			totalFitness += ind.Fitness
		}
	}
	if totalFitness == 0 {
		return pop[rng.Intn(len(pop))]
	}

	pick := rng.Intn(totalFitness)
	current := 0
	for _, ind := range pop {
		if ind.Fitness > 0 {
			current += ind.Fitness
			if current > pick {
				return ind
			}
		}
	}
	return pop[len(pop)-1]
}

// Crossover performs single-point crossover between two parent individuals.
func Crossover(p1, p2 Individual, point int) (Individual, Individual) {
	n := len(p1.Genes)
	if n == 0 || n != len(p2.Genes) {
		return p1, p2
	}
	if point < 0 {
		point = 0
	} else if point > n {
		point = n
	}

	c1 := p1.Genes[:point] + p2.Genes[point:]
	c2 := p2.Genes[:point] + p1.Genes[point:]
	return Individual{Genes: c1}, Individual{Genes: c2}
}

// Mutate applies point mutations with the specified probability per gene.
func Mutate(ind Individual, rate float64, charset string, rng *rand.Rand) Individual {
	if len(charset) == 0 || rate <= 0 {
		return ind
	}
	b := []byte(ind.Genes)
	for i := 0; i < len(b); i++ {
		if rate >= 1.0 || rng.Float64() < rate {
			b[i] = charset[rng.Intn(len(charset))]
		}
	}
	return Individual{Genes: string(b)}
}

// MutateTracked applies point mutations with the specified probability and returns mutated gene indices.
func MutateTracked(ind Individual, rate float64, charset string, rng *rand.Rand) (Individual, []int) {
	if len(charset) == 0 || rate <= 0 {
		return ind, nil
	}
	b := []byte(ind.Genes)
	var mutated []int
	for i := 0; i < len(b); i++ {
		if rate >= 1.0 || rng.Float64() < rate {
			old := b[i]
			choice := charset[rng.Intn(len(charset))]
			if len(charset) > 1 && choice == old {
				choice = charset[(rng.Intn(len(charset)-1)+1+strings.IndexByte(charset, old))%len(charset)]
			}
			b[i] = choice
			if choice != old {
				mutated = append(mutated, i)
			}
		}
	}
	return Individual{Genes: string(b)}, mutated
}

// Evolve runs the genetic algorithm with the provided configuration.
func Evolve(cfg Config) Result {
	if cfg.Target == "" {
		cfg.Target = "GENETIC"
	}
	if cfg.PopSize <= 0 {
		cfg.PopSize = 20
	}
	if cfg.MutationRate <= 0 {
		cfg.MutationRate = 0.05
	}
	if cfg.MaxGenerations <= 0 {
		cfg.MaxGenerations = 50
	}
	if cfg.TournamentSize <= 0 {
		cfg.TournamentSize = 3
	}
	if cfg.Charset == "" {
		cfg.Charset = DefaultCharset(cfg.Target)
	}
	if cfg.RNG == nil {
		cfg.RNG = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	targetLen := len(cfg.Target)
	pop := make([]Individual, cfg.PopSize)
	for i := 0; i < cfg.PopSize; i++ {
		ind := RandomIndividual(targetLen, cfg.Charset, cfg.RNG)
		ind.Fitness = CalculateFitness(ind.Genes, cfg.Target)
		pop[i] = ind
	}

	history := make([]GenerationInfo, 0, cfg.MaxGenerations)
	bestOverall := pop[0]
	for _, ind := range pop {
		if ind.Fitness > bestOverall.Fitness {
			bestOverall = ind
		}
	}

	for gen := 0; gen < cfg.MaxGenerations; gen++ {
		bestThisGen := pop[0]
		for _, ind := range pop {
			if ind.Fitness > bestThisGen.Fitness {
				bestThisGen = ind
			}
		}
		if bestThisGen.Fitness > bestOverall.Fitness {
			bestOverall = bestThisGen
		}

		genPopCopy := make([]Individual, len(pop))
		copy(genPopCopy, pop)
		history = append(history, GenerationInfo{
			Generation:  gen,
			Population:  genPopCopy,
			Best:        bestThisGen,
			BestFitness: bestThisGen.Fitness,
		})

		if bestOverall.Fitness == targetLen {
			break
		}

		nextGen := make([]Individual, 0, cfg.PopSize)
		// Elitism: carry over the best individual
		nextGen = append(nextGen, bestOverall)

		for len(nextGen) < cfg.PopSize {
			p1 := SelectTournament(pop, cfg.TournamentSize, cfg.RNG)
			p2 := SelectTournament(pop, cfg.TournamentSize, cfg.RNG)

			cut := 1
			if targetLen > 1 {
				cut = cfg.RNG.Intn(targetLen-1) + 1
			}
			c1, c2 := Crossover(p1, p2, cut)

			c1 = Mutate(c1, cfg.MutationRate, cfg.Charset, cfg.RNG)
			c1.Fitness = CalculateFitness(c1.Genes, cfg.Target)
			nextGen = append(nextGen, c1)

			if len(nextGen) < cfg.PopSize {
				c2 = Mutate(c2, cfg.MutationRate, cfg.Charset, cfg.RNG)
				c2.Fitness = CalculateFitness(c2.Genes, cfg.Target)
				nextGen = append(nextGen, c2)
			}
		}

		pop = nextGen
	}

	return Result{
		Best:    bestOverall,
		History: history,
	}
}
