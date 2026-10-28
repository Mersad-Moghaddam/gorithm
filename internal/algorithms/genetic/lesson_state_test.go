package genetic

import (
	"encoding/json"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func TestGeneticMutationExplanationPreservesRate(t *testing.T) {
	steps := Trace(Config{Target: "1111", Charset: "0", PopSize: 2, MaxGenerations: 1, MutationRate: 0.001, TournamentSize: 1})
	for _, step := range steps {
		if step.State.(renderers.GeneticLessonState).Phase == "mutation" {
			if !strings.Contains(step.Explanation, "0.001") {
				t.Fatalf("mutation explanation lost rate: %q", step.Explanation)
			}
			return
		}
	}
	t.Fatal("missing mutation")
}

func TestGeneticLessonOperators(t *testing.T) {
	a, b := Crossover(Individual{Genes: "1000"}, Individual{Genes: "0111"}, 2)
	if a.Genes != "1011" || b.Genes != "0100" || CalculateFitness(a.Genes, "1010") != 3 || CalculateFitness(b.Genes, "1010") != 1 {
		t.Fatal(a, b)
	}
	if got := Mutate(Individual{Genes: "1111"}, 1, "0", rand.New(rand.NewSource(42))); got.Genes != "0000" {
		t.Fatal(got)
	}
}

func TestGeneticLessonDeterministicLineage(t *testing.T) {
	for _, n := range []int{1, 2, 3, 8} {
		cfg := Config{Target: "1111", Charset: "0", PopSize: n, MaxGenerations: 2, MutationRate: 1, TournamentSize: 1}
		steps := Trace(cfg)
		if !reflect.DeepEqual(steps, Trace(cfg)) {
			t.Fatal("RNG drift")
		}
		if err := catalog.ValidateSourceLineIDs(Lesson{}, steps); err != nil {
			t.Fatal(err)
		}
		pairs := 0
		for _, step := range steps {
			s, ok := step.State.(renderers.GeneticLessonState)
			if !ok {
				t.Fatalf("missing typed state: %T", step.State)
			}
			if len(s.Population) != n || len(s.NextPopulation) > n {
				t.Fatalf("size %d: %+v", n, s)
			}
			ids := map[string]bool{}
			for _, p := range s.Population {
				if ids[p.ID] {
					t.Fatal("duplicate identity")
				}
				ids[p.ID] = true
			}
			if s.Phase == "crossover" {
				a, b := Crossover(Individual{Genes: s.Parents[0].Genes}, Individual{Genes: s.Parents[1].Genes}, s.Cut)
				if a.Genes != s.Crossed[0].Genes || b.Genes != s.Crossed[1].Genes {
					t.Fatal("wrong crossover")
				}
			}
			if s.Phase == "mutation" {
				pairs++
				for j := range s.Mutated {
					for k := range s.Mutated[j].Genes {
						changed := s.Crossed[j].Genes[k] != s.Mutated[j].Genes[k]
						found := false
						for _, l := range s.ChangedLoci[j] {
							found = found || l == k
						}
						if changed != found {
							t.Fatal("false mutation")
						}
					}
				}
			}
			for i := 1; i < len(s.BestFitnessHistory); i++ {
				if s.BestFitnessHistory[i] < s.BestFitnessHistory[i-1] {
					t.Fatal("elite regressed")
				}
			}
			if s.StopReason == "target-found" && s.Population[s.BestIndex].Fitness != len(s.Target) {
				t.Fatal("false convergence")
			}
		}
		if n > 1 && pairs != 2*((n-1+1)/2) {
			t.Fatalf("missing pair snapshots: n=%d pairs=%d", n, pairs)
		}
		last := steps[len(steps)-1].State.(renderers.GeneticLessonState)
		if last.StopReason != "generation-limit" {
			t.Fatal(last.StopReason)
		}
		before := steps[0].State.(renderers.GeneticLessonState).Population[0].Genes
		last.Population[0].Genes = "mutated"
		if steps[0].State.(renderers.GeneticLessonState).Population[0].Genes != before {
			t.Fatal("shared population")
		}
	}
}

func TestGeneticLessonValidationAndEscaping(t *testing.T) {
	invalid := []string{`null`, `[]`, `{`, `{"target":""}`, `{"target":"é"}`, `{"target":"\n"}`, `{"target":null}`, `{"popSize":null}`, `{"popSize":1}`, `{"popSize":25}`, `{"maxGenerations":0}`, `{"maxGenerations":101}`, `{"mutationRate":0}`, `{"mutationRate":1.1}`, `{"tournamentSize":0}`, `{"tournamentSize":9}`, `{"target":"1234567890123456789012345"}`}
	for _, raw := range invalid {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err == nil {
			t.Errorf("accepted %s", raw)
		}
	}
	for _, raw := range []string{`{}`, `{"target":"<&\"'>","popSize":2,"tournamentSize":2,"maxGenerations":1,"mutationRate":1}`, `{"target":"123456789012345678901234","popSize":24,"tournamentSize":24,"maxGenerations":100,"mutationRate":1}`} {
		if _, err := (Lesson{}).Build(json.RawMessage(raw)); err != nil {
			t.Fatal(raw, err)
		}
	}
	for _, cfg := range []Config{{Target: "é"}, {Target: "X", PopSize: 25}, {MutationRate: math.NaN()}, {Target: "X", Charset: "\n"}} {
		steps := Trace(cfg)
		s, ok := steps[0].State.(renderers.GeneticLessonState)
		if len(steps) != 1 || !ok || s.StopReason != "invalid-input" {
			t.Fatal("unbounded invalid input")
		}
	}
}

func TestGeneticLessonPreservesRandomDrawOrder(t *testing.T) {
	for _, n := range []int{2, 3, 8} {
		cfg := Config{Target: "11111111", Charset: "01", PopSize: n, MaxGenerations: 1, MutationRate: .08, TournamentSize: 3}
		steps := Trace(cfg)
		rng := rand.New(rand.NewSource(42))
		pop := make([]Individual, n)
		best := 0
		for i := range pop {
			pop[i] = RandomIndividual(len(cfg.Target), cfg.Charset, rng)
			pop[i].Fitness = CalculateFitness(pop[i].Genes, cfg.Target)
			if pop[i].Fitness > pop[best].Fitness {
				best = i
			}
		}
		next := []Individual{pop[best]}
		pair := 0
		for len(next) < n {
			pair++
			p1, p2 := SelectTournament(pop, cfg.TournamentSize, rng), SelectTournament(pop, cfg.TournamentSize, rng)
			c1, c2 := Crossover(p1, p2, rng.Intn(len(cfg.Target)-1)+1)
			c1 = Mutate(c1, cfg.MutationRate, cfg.Charset, rng)
			if pair == 1 {
				c2 = Mutate(c2, cfg.MutationRate, cfg.Charset, rng)
			}
			c1.Fitness = CalculateFitness(c1.Genes, cfg.Target)
			next = append(next, c1)
			if len(next) < n {
				if pair > 1 {
					c2 = Mutate(c2, cfg.MutationRate, cfg.Charset, rng)
				}
				c2.Fitness = CalculateFitness(c2.Genes, cfg.Target)
				next = append(next, c2)
			}
		}
		found := false
		for _, step := range steps {
			if step.LineID != "L13" {
				continue
			}
			found = true
			s := step.State.(renderers.GeneticLessonState)
			for i, want := range next {
				if s.Population[i].Genes != want.Genes || s.Population[i].Fitness != want.Fitness {
					t.Fatalf("RNG order changed n=%d slot=%d", n, i)
				}
			}
		}
		if !found {
			t.Fatal("missing replacement")
		}
	}
}

func TestGeneticLessonSnapshotIsolation(t *testing.T) {
	steps := Trace(Config{Target: "1111", Charset: "0", PopSize: 3, MaxGenerations: 2, MutationRate: 1})
	before, _ := json.Marshal(steps[0 : len(steps)-1])
	s := steps[len(steps)-1].State.(renderers.GeneticLessonState)
	s.Population[0].Genes = "changed"
	s.BestFitnessHistory[0] = 999
	after, _ := json.Marshal(steps[0 : len(steps)-1])
	if string(before) != string(after) {
		t.Fatal("snapshot aliases earlier frame")
	}
	for i, step := range steps {
		s := step.State.(renderers.GeneticLessonState)
		if s.Phase != "mutation" {
			continue
		}
		if len(s.NextPopulation) > 0 {
			s.NextPopulation[0].Genes = "changed"
		}
		for j := range s.ChangedLoci {
			if len(s.ChangedLoci[j]) > 0 {
				s.ChangedLoci[j][0] = 999
			}
		}
		for k := i + 1; k < len(steps)-1; k++ {
			later := steps[k].State.(renderers.GeneticLessonState)
			for _, p := range later.Population {
				if p.Genes == "changed" {
					t.Fatal("generation alias")
				}
			}
			for _, loci := range later.ChangedLoci {
				for _, l := range loci {
					if l == 999 {
						t.Fatal("locus alias")
					}
				}
			}
		}
		break
	}
}
