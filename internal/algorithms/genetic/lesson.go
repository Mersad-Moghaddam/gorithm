package genetic

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	Target         string   `json:"target"`
	PopSize        *int     `json:"popSize"`
	MaxGenerations *int     `json:"maxGenerations"`
	MutationRate   *float64 `json:"mutationRate"`
	TournamentSize *int     `json:"tournamentSize"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "genetic",
		Name:     "Genetic Algorithm (Evolutionary AI)",
		Category: catalog.CategoryAI,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "An evolutionary search heuristic inspired by Charles Darwin's theory of natural selection that iteratively evolves a population of candidate chromosomes toward an optimal solution.",
			TimeComplexity:  "O(generations · popSize · fitness)",
			SpaceComplexity: "O(popSize)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"Natural Selection",
				"Chromosome Crossover",
				"Mutation Rate",
				"Fitness Landscape",
			},
			HowItWorks: "Begins by creating an initial population of random chromosomes and scoring each candidate using a fitness function that quantifies its proximity to the target. In each generational cycle, parent individuals are chosen via tournament selection where higher fitness yields a greater probability of reproduction. Crossover recombines segments of parental genes at designated split points to construct offspring chromosomes, while stochastic point mutations randomly alter genes to preserve population diversity and avoid local optima. The fittest chromosomes survive across generations until convergence or completion.",
		},
		DefaultInput: `{"target":"GENETIC","popSize":8,"maxGenerations":15,"mutationRate":0.08}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func Evolve(target string, popSize, maxGen int) Individual {"},
		{ID: "L2", Number: 2, Code: "    pop := InitPopulation(popSize, len(target))"},
		{ID: "L3", Number: 3, Code: "    EvaluateFitness(pop, target)"},
		{ID: "L4", Number: 4, Code: "    for gen := 0; gen < maxGen; gen++ {"},
		{ID: "L5", Number: 5, Code: "        best := FindBest(pop)"},
		{ID: "L6", Number: 6, Code: "        if best.Fitness == len(target) { return best }"},
		{ID: "L7", Number: 7, Code: "        nextGen := []Individual{best} // elite retains identity"},
		{ID: "L8", Number: 8, Code: "        for len(nextGen) < popSize {"},
		{ID: "L9", Number: 9, Code: "            p1, p2 := Select(pop), Select(pop)"},
		{ID: "L10", Number: 10, Code: "            c1, c2 := Crossover(p1, p2)"},
		{ID: "L11", Number: 11, Code: "            nextGen = AppendMutatedWithinCapacity(nextGen, c1, c2, popSize)"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "        pop = nextGen // already capacity-limited"},
		{ID: "L14", Number: 14, Code: "        EvaluateFitness(pop, target)"},
		{ID: "L15", Number: 15, Code: "    }"},
		{ID: "L16", Number: 16, Code: "    return FindBest(pop)"},
		{ID: "L17", Number: 17, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	cfg := Config{
		Target:         "GENETIC",
		PopSize:        8,
		MaxGenerations: 15,
		MutationRate:   0.08,
		TournamentSize: 3,
	}

	if len(bytes.TrimSpace(input)) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil || fields == nil {
			return nil, fmt.Errorf("genetic input must be a JSON object")
		}
		for _, key := range []string{"target", "popSize", "maxGenerations", "mutationRate", "tournamentSize"} {
			if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
				return nil, fmt.Errorf("%s must not be null", key)
			}
		}
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err != nil {
			return nil, err
		}
		if _, supplied := fields["target"]; supplied {
			cfg.Target = custom.Target
		}
		if custom.PopSize != nil {
			cfg.PopSize = *custom.PopSize
		}
		if custom.MaxGenerations != nil {
			cfg.MaxGenerations = *custom.MaxGenerations
		}
		if custom.MutationRate != nil {
			cfg.MutationRate = *custom.MutationRate
		}
		if custom.TournamentSize != nil {
			cfg.TournamentSize = *custom.TournamentSize
		}
	}
	if cfg.PopSize < 2 || cfg.TournamentSize < 1 || cfg.TournamentSize > cfg.PopSize {
		return nil, fmt.Errorf("population must be 2–24; tournament must be 1–population")
	}
	if err := geneticValidate(cfg); err != nil {
		return nil, err
	}
	return Trace(cfg), nil
}

func init() {
	catalog.Register(Lesson{})
}
