package genetic

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestGeneticAlgorithmImprovement(t *testing.T) {
	target := "GENETIC"
	cfg := Config{
		Target:         target,
		PopSize:        20,
		MutationRate:   0.05,
		MaxGenerations: 50,
		TournamentSize: 3,
		RNG:            rand.New(rand.NewSource(42)),
	}

	result := Evolve(cfg)
	if len(result.History) == 0 {
		t.Fatal("expected non-empty evolution history")
	}

	initialBest := result.History[0].BestFitness
	finalBest := result.Best.Fitness

	if finalBest <= initialBest && finalBest < len(target) {
		t.Errorf("expected fitness improvement: initial best %d, final best %d", initialBest, finalBest)
	}

	if finalBest != len(target) {
		t.Logf("final best %q with fitness %d/%d (target %q)", result.Best.Genes, finalBest, len(target), target)
	}
}

func TestBinarySequenceImprovement(t *testing.T) {
	target := "10110011"
	cfg := Config{
		Target:         target,
		PopSize:        16,
		MutationRate:   0.05,
		MaxGenerations: 40,
		TournamentSize: 3,
		RNG:            rand.New(rand.NewSource(99)),
	}

	result := Evolve(cfg)
	if result.Best.Fitness < len(target)/2 {
		t.Errorf("expected fitness >= %d, got %d", len(target)/2, result.Best.Fitness)
	}
}

func TestSelectionRouletteAndTournament(t *testing.T) {
	pop := []Individual{
		{Genes: "A", Fitness: 0},
		{Genes: "B", Fitness: 1},
		{Genes: "C", Fitness: 10},
	}
	rng := rand.New(rand.NewSource(1))

	// Tournament with k=3 must pick highest fitness (C)
	selectedTournament := SelectTournament(pop, 3, rng)
	if selectedTournament.Genes != "C" {
		t.Errorf("expected tournament to pick 'C', got %q", selectedTournament.Genes)
	}

	// Roulette selection over many trials should pick C most often
	counts := make(map[string]int)
	for i := 0; i < 300; i++ {
		ind := SelectRoulette(pop, rng)
		counts[ind.Genes]++
	}
	if counts["C"] <= counts["A"] || counts["C"] <= counts["B"] {
		t.Errorf("expected 'C' to be selected most frequently, got counts: %v", counts)
	}
}

func TestCrossover(t *testing.T) {
	p1 := Individual{Genes: "AAAAAA"}
	p2 := Individual{Genes: "BBBBBB"}

	c1, c2 := Crossover(p1, p2, 3)
	if c1.Genes != "AAABBB" {
		t.Errorf("expected c1 'AAABBB', got %q", c1.Genes)
	}
	if c2.Genes != "BBBAAA" {
		t.Errorf("expected c2 'BBBAAA', got %q", c2.Genes)
	}
}

func TestMutation(t *testing.T) {
	ind := Individual{Genes: "AAAA"}
	rng := rand.New(rand.NewSource(42))

	// With rate 1.0, all genes should mutate
	mutated := Mutate(ind, 1.0, "B", rng)
	if mutated.Genes != "BBBB" {
		t.Errorf("expected 'BBBB', got %q", mutated.Genes)
	}

	// With rate 0.0, genes should remain identical
	unmutated := Mutate(ind, 0.0, "B", rng)
	if unmutated.Genes != "AAAA" {
		t.Errorf("expected 'AAAA', got %q", unmutated.Genes)
	}
}

func TestGeneticLessonMeta(t *testing.T) {
	lesson := Lesson{}
	meta := lesson.Meta()

	if meta.ID != "genetic" {
		t.Errorf("expected ID 'genetic', got %q", meta.ID)
	}
	if meta.Name != "Genetic Algorithm (Evolutionary AI)" {
		t.Errorf("expected Name 'Genetic Algorithm (Evolutionary AI)', got %q", meta.Name)
	}
	if meta.Category != catalog.CategoryAI {
		t.Errorf("expected Category %q, got %q", catalog.CategoryAI, meta.Category)
	}
	if meta.Renderer != "table" {
		t.Errorf("expected Renderer 'table', got %q", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(generations · popSize · fitness)" {
		t.Errorf("expected TimeComplexity 'O(generations · popSize · fitness)', got %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(popSize)" {
		t.Errorf("expected SpaceComplexity 'O(popSize)', got %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("expected Difficulty 'Hard', got %q", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{"Natural Selection", "Chromosome Crossover", "Mutation Rate", "Fitness Landscape"}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("expected KeyConcepts %v, got %v", expectedConcepts, meta.Documentation.KeyConcepts)
	}
	if meta.Documentation.Summary == "" {
		t.Error("expected non-empty Summary")
	}
	if meta.Documentation.HowItWorks == "" {
		t.Error("expected non-empty HowItWorks")
	}

	l, ok := catalog.Get("genetic")
	if !ok {
		t.Fatal("lesson 'genetic' not registered in catalog")
	}
	if l.Meta().ID != "genetic" {
		t.Errorf("expected registered ID 'genetic', got %q", l.Meta().ID)
	}
}

func TestGeneticTraceValidation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
	if _, ok := lastStep.Variables["bestFitness"]; !ok {
		t.Fatalf("expected bestFitness in variables: %v", lastStep.Variables)
	}

	for i, step := range steps {
		tableState, ok := step.State.(renderers.GeneticLessonState)
		if !ok {
			t.Fatalf("step %d state is not renderers.GeneticLessonState: %T", i, step.State)
		}
		if len(tableState.RowHeaders) == 0 || len(tableState.ColHeaders) == 0 {
			t.Fatalf("step %d has empty headers: rows=%d cols=%d", i, len(tableState.RowHeaders), len(tableState.ColHeaders))
		}
	}
}

func TestGeneticTraceCustomInput(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"target": "EVO", "popSize": 6, "maxGenerations": 10, "mutationRate": 0.1}`)
	steps, err := lesson.Build(json.RawMessage(input))
	if err != nil {
		t.Fatalf("failed to build trace with custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("trace produced 0 steps")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	lastStep := steps[len(steps)-1]
	if lastStep.Action != trace.ActionDone {
		t.Fatalf("expected last action to be %s, got %s", trace.ActionDone, lastStep.Action)
	}
}

func TestGeneticRenderTableSVG(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	for i, step := range steps {
		svg := renderers.RenderGeneticLessonSVG(step, 800, 450)
		if svg == "" {
			t.Errorf("step %d produced empty SVG", i)
		}
	}
}

func TestLongSequenceTableState(t *testing.T) {
	lesson := Lesson{}
	input := []byte(`{"target": "EVOLUTIONARY_ALGORITHM", "popSize": 5, "maxGenerations": 3}`)
	steps, err := lesson.Build(json.RawMessage(input))
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps")
	}
	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("line validation failed: %v", err)
	}

	tableState := steps[len(steps)-1].State.(renderers.GeneticLessonState)
	if len(tableState.ColHeaders) != 2 || tableState.ColHeaders[0] != "Chromosome" {
		t.Errorf("expected 2-column layout for long target, got %v", tableState.ColHeaders)
	}
}

func TestEdgeCases(t *testing.T) {
	// Selection with empty population
	emptyInd := SelectTournament(nil, 3, rand.New(rand.NewSource(1)))
	if emptyInd.Genes != "" {
		t.Errorf("expected empty individual, got %v", emptyInd)
	}
	emptyRoulette := SelectRoulette(nil, rand.New(rand.NewSource(1)))
	if emptyRoulette.Genes != "" {
		t.Errorf("expected empty individual, got %v", emptyRoulette)
	}

	// Crossover with mismatched lengths
	p1 := Individual{Genes: "A"}
	p2 := Individual{Genes: "ABC"}
	c1, c2 := Crossover(p1, p2, 1)
	if c1.Genes != p1.Genes || c2.Genes != p2.Genes {
		t.Errorf("expected unchanged parents on length mismatch")
	}

	// CalculateFitness length handling
	if fit := CalculateFitness("ABC", "A"); fit != 1 {
		t.Errorf("expected fitness 1, got %d", fit)
	}

	// DefaultCharset
	if cs := DefaultCharset("0101"); cs != "01" {
		t.Errorf("expected '01', got %q", cs)
	}
	if cs := DefaultCharset(""); cs != "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		t.Errorf("expected uppercase letters, got %q", cs)
	}
}

func TestConvergenceToTarget(t *testing.T) {
	target := "HI"
	cfg := Config{
		Target:         target,
		PopSize:        30,
		MutationRate:   0.05,
		MaxGenerations: 50,
		TournamentSize: 3,
		RNG:            rand.New(rand.NewSource(42)),
	}

	res := Evolve(cfg)
	if res.Best.Fitness != len(target) {
		t.Errorf("expected full convergence to %q, got %q (fitness %d/%d)", target, res.Best.Genes, res.Best.Fitness, len(target))
	}
}
