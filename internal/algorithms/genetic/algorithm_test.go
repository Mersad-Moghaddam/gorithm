package genetic

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestGeneticAlgorithmImprovement(t *testing.T) {
	target := "GENETIC"
	cfg := Config{
		Target:          target,
		PopSize:         20,
		MutationRate:    0.05,
		MaxGenerations:  50,
		TournamentSize:  3,
		RNG:             rand.New(rand.NewSource(42)),
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
		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not renderers.TableState: %T", i, step.State)
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
		svg := renderers.RenderTableSVG(step, 800, 450)
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

	tableState := steps[len(steps)-1].State.(renderers.TableState)
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

func TestInputValidation(t *testing.T) {
	lesson := Lesson{}

	invalidInputs := []struct {
		name  string
		input string
	}{
		{"null literal", "null"},
		{"null with spaces", "   null   "},
		{"malformed json", "{\"target\": "},
		{"json array", "[1, 2, 3]"},
		{"json number", "42"},
		{"empty target", `{"target": ""}`},
		{"null target", `{"target": null}`},
		{"zero popSize", `{"popSize": 0}`},
		{"negative popSize", `{"popSize": -5}`},
		{"null popSize", `{"popSize": null}`},
		{"negative mutationRate", `{"mutationRate": -0.05}`},
		{"excessive mutationRate", `{"mutationRate": 1.2}`},
		{"null mutationRate", `{"mutationRate": null}`},
		{"zero maxGenerations", `{"maxGenerations": 0}`},
		{"negative tournamentSize", `{"tournamentSize": -1}`},
	}

	for _, tc := range invalidInputs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Errorf("expected error for input %q, got nil", tc.input)
			}
		})
	}

	validInputs := []struct {
		name  string
		input string
	}{
		{"empty input", ""},
		{"empty json object", "{}"},
		{"valid target", `{"target": "CODE"}`},
		{"valid full input", `{"target": "GO", "popSize": 10, "maxGenerations": 5, "mutationRate": 0.1, "tournamentSize": 2}`},
	}

	for _, tc := range validInputs {
		t.Run(tc.name, func(t *testing.T) {
			steps, err := lesson.Build(json.RawMessage(tc.input))
			if err != nil {
				t.Fatalf("expected success for input %q, got err: %v", tc.input, err)
			}
			if len(steps) == 0 {
				t.Fatalf("expected steps for input %q", tc.input)
			}
		})
	}
}

func TestFitnessTracking(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	prevFitness := -1
	generationStepFound := false

	for i, step := range steps {
		if rawFit, ok := step.Variables["bestFitness"]; ok {
			fit, ok := rawFit.(int)
			if !ok {
				t.Fatalf("step %d bestFitness is not an int: %T", i, rawFit)
			}
			if fit < prevFitness {
				t.Fatalf("step %d bestFitness decreased from %d to %d", i, prevFitness, fit)
			}
			prevFitness = fit
		}

		tableState, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is not TableState", i)
		}

		if step.LineID == "L4" || step.LineID == "L5" || step.LineID == "L9" || step.LineID == "L10" || step.LineID == "L11" || step.LineID == "L13" || step.LineID == "L14" {
			generationStepFound = true
			expectedPrefix := "Gen "
			expectedSuffix := "Crossover & mutation producing next gen."
			if !strings.HasPrefix(tableState.Caption, expectedPrefix) || !strings.HasSuffix(tableState.Caption, expectedSuffix) {
				t.Errorf("step %d (%s) caption %q does not match expected format", i, step.LineID, tableState.Caption)
			}
		}
	}

	if !generationStepFound {
		t.Fatal("no generation steps found in trace")
	}
}

func TestCrossoverAndMutationEvents(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}

	var crossoverSteps []trace.Step
	var mutationSteps []trace.Step
	var selectionSteps []trace.Step

	for _, step := range steps {
		switch step.LineID {
		case "L9":
			selectionSteps = append(selectionSteps, step)
		case "L10":
			crossoverSteps = append(crossoverSteps, step)
		case "L11":
			mutationSteps = append(mutationSteps, step)
		}
	}

	if len(selectionSteps) == 0 {
		t.Fatal("expected selection steps (L9)")
	}
	if len(crossoverSteps) == 0 {
		t.Fatal("expected crossover steps (L10)")
	}
	if len(mutationSteps) == 0 {
		t.Fatal("expected mutation steps (L11)")
	}

	// Verify selection highlights parents
	for i, selStep := range selectionSteps {
		ts := selStep.State.(renderers.TableState)
		foundP1 := false
		foundP2 := false
		for _, cell := range ts.Cells {
			if cell.State == "source" {
				foundP1 = true
			}
			if cell.State == "active" {
				foundP2 = true
			}
		}
		if !foundP1 || !foundP2 {
			t.Errorf("selection step %d missing highlighted parents (p1=%v, p2=%v)", i, foundP1, foundP2)
		}
	}

	// Verify crossover combines parents
	for i, crossStep := range crossoverSteps {
		ts := crossStep.State.(renderers.TableState)
		if len(ts.RowHeaders) != 4 {
			t.Errorf("crossover step %d expected 4 rows, got %d", i, len(ts.RowHeaders))
		}
		if crossStep.Action != trace.ActionSwap {
			t.Errorf("crossover step %d expected ActionSwap, got %s", i, crossStep.Action)
		}
		// Child 1 (row 2) should have both source (from P1) and active (from P2) cells
		hasSource := false
		hasActive := false
		for _, cell := range ts.Cells {
			if cell.Row == 2 {
				if cell.State == "source" {
					hasSource = true
				}
				if cell.State == "active" {
					hasActive = true
				}
			}
		}
		if !hasSource || !hasActive {
			t.Errorf("crossover step %d Child 1 missing blended genetic material (source=%v, active=%v)", i, hasSource, hasActive)
		}
	}

	// Verify mutation indicates flipped gene
	for i, mutStep := range mutationSteps {
		ts := mutStep.State.(renderers.TableState)
		if len(ts.RowHeaders) != 4 {
			t.Errorf("mutation step %d expected 4 rows, got %d", i, len(ts.RowHeaders))
		}
		if mutStep.Action != trace.ActionSet {
			t.Errorf("mutation step %d expected ActionSet, got %s", i, mutStep.Action)
		}
		hasIndicator := false
		for _, cell := range ts.Cells {
			if cell.State == "active" || strings.Contains(cell.Value, "*") {
				hasIndicator = true
				break
			}
		}
		if !hasIndicator {
			t.Errorf("mutation step %d missing visual mutation indicator", i)
		}
	}
}

func TestSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("failed to build trace: %v", err)
	}
	if len(steps) < 2 {
		t.Fatalf("expected at least 2 steps, got %d", len(steps))
	}

	for i := 0; i < len(steps)-1; i++ {
		stateA := steps[i].State.(renderers.TableState)
		stateB := steps[i+1].State.(renderers.TableState)

		if len(stateA.Cells) > 0 && len(stateB.Cells) > 0 {
			origVal := stateA.Cells[0].Value
			origState := stateA.Cells[0].State
			stateA.Cells[0].Value = "MUTATION_TEST_VAL"
			stateA.Cells[0].State = "MUTATION_TEST_STATE"

			if stateB.Cells[0].Value == "MUTATION_TEST_VAL" && origVal != "MUTATION_TEST_VAL" {
				t.Fatalf("steps %d and %d share Cells slice backing array", i, i+1)
			}
			stateA.Cells[0].Value = origVal
			stateA.Cells[0].State = origState
		}

		if len(stateA.RowHeaders) > 0 && len(stateB.RowHeaders) > 0 {
			origHeader := stateA.RowHeaders[0]
			stateA.RowHeaders[0] = "MODIFIED_HEADER"
			if stateB.RowHeaders[0] == "MODIFIED_HEADER" && origHeader != "MODIFIED_HEADER" {
				t.Fatalf("steps %d and %d share RowHeaders slice backing array", i, i+1)
			}
			stateA.RowHeaders[0] = origHeader
		}

		if len(stateA.ColHeaders) > 0 && len(stateB.ColHeaders) > 0 {
			origCol := stateA.ColHeaders[0]
			stateA.ColHeaders[0] = "MODIFIED_COL"
			if stateB.ColHeaders[0] == "MODIFIED_COL" && origCol != "MODIFIED_COL" {
				t.Fatalf("steps %d and %d share ColHeaders slice backing array", i, i+1)
			}
			stateA.ColHeaders[0] = origCol
		}
	}
}

