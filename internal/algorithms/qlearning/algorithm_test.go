package qlearning

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

func TestGridWorldTransitions(t *testing.T) {
	env := DefaultGridWorld()

	// Collision with boundary: Up from (0,0) should stay at (0,0)
	next, r, done := env.Step(Coord{Row: 0, Col: 0}, ActionUp)
	if next != (Coord{Row: 0, Col: 0}) || r != 0.0 || done {
		t.Fatalf("expected bounce at top boundary, got next=%v, r=%v, done=%v", next, r, done)
	}

	// Collision with obstacle: Right from (1,0) into obstacle (1,1)
	next, r, done = env.Step(Coord{Row: 1, Col: 0}, ActionRight)
	if next != (Coord{Row: 1, Col: 0}) || r != 0.0 || done {
		t.Fatalf("expected bounce at obstacle (1,1), got next=%v, r=%v, done=%v", next, r, done)
	}

	// Step into goal from (2,1) moving Right
	next, r, done = env.Step(Coord{Row: 2, Col: 1}, ActionRight)
	if next != (Coord{Row: 2, Col: 2}) || r != 10.0 || !done {
		t.Fatalf("expected goal reach, got next=%v, r=%v, done=%v", next, r, done)
	}
}

func TestQLearningValuePropagation(t *testing.T) {
	env := DefaultGridWorld()
	agent := NewQAgent(env, 0.5, 0.9, 0.2)
	rng := rand.New(rand.NewSource(42))

	// Train 200 episodes
	agent.Train(200, 20, rng)

	// Goal neighbors must have positive Q-values for action leading into goal
	s21Idx := env.StateIndex(Coord{Row: 2, Col: 1})
	if agent.Q[s21Idx][ActionRight] <= 0 {
		t.Fatalf("expected Q((2,1), Right) > 0, got %f", agent.Q[s21Idx][ActionRight])
	}

	s12Idx := env.StateIndex(Coord{Row: 1, Col: 2})
	if agent.Q[s12Idx][ActionDown] <= 0 {
		t.Fatalf("expected Q((1,2), Down) > 0, got %f", agent.Q[s12Idx][ActionDown])
	}

	// Value must have propagated back to start (0,0)
	s00Idx := env.StateIndex(Coord{Row: 0, Col: 0})
	_, startMaxQ := agent.MaxQ(Coord{Row: 0, Col: 0})
	if startMaxQ <= 0 {
		t.Fatalf("expected start state max Q > 0, got %f (table row: %v)", startMaxQ, agent.Q[s00Idx])
	}

	// Verify greedy trajectory from Start reaches Goal without hitting obstacles
	curr := env.Start
	reachedGoal := false
	for step := 0; step < 10; step++ {
		if env.IsTerminal(curr) {
			reachedGoal = true
			break
		}
		bestA, _ := agent.MaxQ(curr)
		next, _, done := env.Step(curr, bestA)
		if env.Obstacles[next] {
			t.Fatalf("greedy policy moved into obstacle at %v", next)
		}
		curr = next
		if done {
			reachedGoal = true
			break
		}
	}

	if !reachedGoal {
		t.Fatalf("greedy policy failed to reach goal from start, ended at %v", curr)
	}
}

func TestQLearningLessonMeta(t *testing.T) {
	lesson, ok := catalog.Get("qlearning")
	if !ok {
		t.Fatal("expected qlearning lesson to be registered in catalog")
	}

	meta := lesson.Meta()
	if meta.ID != "qlearning" {
		t.Errorf("ID = %q, want 'qlearning'", meta.ID)
	}
	if meta.Name != "Q-Learning (Reinforcement Learning)" {
		t.Errorf("Name = %q, want 'Q-Learning (Reinforcement Learning)'", meta.Name)
	}
	if meta.Category != catalog.CategoryAI {
		t.Errorf("Category = %q, want %q", meta.Category, catalog.CategoryAI)
	}
	if meta.Renderer != "table" {
		t.Errorf("Renderer = %q, want 'table'", meta.Renderer)
	}
	if meta.Documentation.TimeComplexity != "O(episodes · steps)" {
		t.Errorf("TimeComplexity = %q", meta.Documentation.TimeComplexity)
	}
	if meta.Documentation.SpaceComplexity != "O(|S| · |A|)" {
		t.Errorf("SpaceComplexity = %q", meta.Documentation.SpaceComplexity)
	}
	if meta.Documentation.Difficulty != "Hard" {
		t.Errorf("Difficulty = %q", meta.Documentation.Difficulty)
	}

	expectedConcepts := []string{
		"Temporal Difference",
		"Exploration vs Exploitation",
		"Bellman Optimality",
		"Policy Extraction",
	}
	if !reflect.DeepEqual(meta.Documentation.KeyConcepts, expectedConcepts) {
		t.Errorf("KeyConcepts = %v, want %v", meta.Documentation.KeyConcepts, expectedConcepts)
	}
}

func TestQLearningTraceAndSourceLineIDs(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected at least one step in trace")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("source line validation failed: %v", err)
	}

	seenActions := make(map[trace.ActionType]bool)
	seenLines := make(map[string]bool)

	for i, step := range steps {
		seenActions[step.Action] = true
		seenLines[step.LineID] = true

		// Check variables
		if step.Variables == nil {
			t.Fatalf("step %d missing variables", i)
		}
		if _, ok := step.Variables["state"]; !ok {
			t.Errorf("step %d missing 'state' variable", i)
		}
		if _, ok := step.Variables["action"]; !ok {
			t.Errorf("step %d missing 'action' variable", i)
		}
		if _, ok := step.Variables["reward"]; !ok {
			t.Errorf("step %d missing 'reward' variable", i)
		}

		// Check TableState
		typed, ok := step.State.(renderers.QLearningLessonState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.QLearningLessonState", i, step.State)
		}
		tbl := typed.TableState
		if len(tbl.RowHeaders) != 9 {
			t.Errorf("step %d table row count = %d, want 9", i, len(tbl.RowHeaders))
		}
		if len(tbl.ColHeaders) != 4 {
			t.Errorf("step %d table col count = %d, want 4", i, len(tbl.ColHeaders))
		}
		if len(tbl.Cells) != 36 {
			t.Errorf("step %d table cell count = %d, want 36", i, len(tbl.Cells))
		}

		// Verify SVG rendering
		svg := renderers.RenderQLearningLessonSVG(step, 800, 600)
		if !strings.Contains(svg, "<svg") && !strings.Contains(svg, "<defs>") && !strings.Contains(svg, "<text") {
			t.Errorf("step %d RenderTableSVG output missing expected SVG elements", i)
		}
	}

	if !seenActions[trace.ActionInit] {
		t.Errorf("trace missing ActionInit")
	}
	if !seenActions[trace.ActionSet] {
		t.Errorf("trace missing ActionSet")
	}
	if !seenActions[trace.ActionCompare] {
		t.Errorf("trace missing ActionCompare")
	}
	if !seenActions[trace.ActionDone] {
		t.Errorf("trace missing ActionDone")
	}
	if !seenLines["L10"] {
		t.Errorf("trace missing Bellman update line L10")
	}

	lastStep := steps[len(steps)-1]
	lastTbl := lastStep.State.(renderers.QLearningLessonState).TableState
	hasLearnedVal := false
	for _, cell := range lastTbl.Cells {
		if cell.Value != "0.00" {
			hasLearnedVal = true
			break
		}
	}
	if !hasLearnedVal {
		t.Errorf("expected final trace table to contain non-zero learned Q-values")
	}
}

func TestQLearningCustomInput(t *testing.T) {
	lesson := Lesson{}
	rawInput := json.RawMessage(`{"episodes": 2, "alpha": 0.6, "gamma": 0.8, "epsilon": 0.1}`)

	steps, err := lesson.Build(rawInput)
	if err != nil {
		t.Fatalf("Build failed on custom input: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected steps with custom input")
	}

	if err := catalog.ValidateSourceLineIDs(lesson, steps); err != nil {
		t.Fatalf("custom input source line validation failed: %v", err)
	}
}
