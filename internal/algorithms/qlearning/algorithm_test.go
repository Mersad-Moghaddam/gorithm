package qlearning

import (
	"encoding/json"
	"math/rand"
	"reflect"
	"regexp"
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
		tbl, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d state is %T, want renderers.TableState", i, step.State)
		}
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
		svg := renderers.RenderTableSVG(step, 800, 600)
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
	lastTbl := lastStep.State.(renderers.TableState)
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

func TestInputValidationStrict(t *testing.T) {
	lesson := Lesson{}

	// Valid nil or empty
	if _, err := lesson.Build(nil); err != nil {
		t.Fatalf("expected nil input to succeed, got %v", err)
	}
	if _, err := lesson.Build(json.RawMessage("")); err != nil {
		t.Fatalf("expected empty input to succeed, got %v", err)
	}

	invalidCases := []struct {
		name  string
		input string
	}{
		{"null", "null"},
		{"malformed json", "{episodes: 5"},
		{"negative episodes", `{"episodes": -1}`},
		{"negative alpha", `{"alpha": -0.1}`},
		{"alpha greater than 1", `{"alpha": 1.05}`},
		{"negative gamma", `{"gamma": -0.5}`},
		{"gamma greater than 1", `{"gamma": 1.2}`},
		{"negative epsilon", `{"epsilon": -0.01}`},
		{"epsilon greater than 1", `{"epsilon": 2.0}`},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := lesson.Build(json.RawMessage(tc.input))
			if err == nil {
				t.Fatalf("expected error for input %q, got nil", tc.input)
			}
		})
	}
}

func TestSnapshotImmutability(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	for i := 0; i < len(steps)-1; i++ {
		t1, ok1 := steps[i].State.(renderers.TableState)
		t2, ok2 := steps[i+1].State.(renderers.TableState)
		if !ok1 || !ok2 {
			t.Fatalf("step state is not TableState")
		}

		origVal := t2.Cells[0].Value
		origState := t2.Cells[0].State

		t1.Cells[0].Value = "MUTATED"
		t1.Cells[0].State = "MUTATED_STATE"

		if t2.Cells[0].Value == "MUTATED" || t2.Cells[0].State == "MUTATED_STATE" {
			t.Fatalf("snapshot mutation leaked from step %d to step %d", i, i+1)
		}

		t2.Cells[0].Value = origVal
		t2.Cells[0].State = origState
	}
}

func TestBellmanUpdateVisualExplanation(t *testing.T) {
	lesson := Lesson{}
	steps, err := lesson.Build(nil)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	captionRegex := regexp.MustCompile(`^Agent took action (Up|Right|Down|Left) from state \(\d+,\d+\) -> got reward \d+\. TD error = -?\d+\.\d{2}; updated Q\(\(\d+,\d+\), (Up|Right|Down|Left)\) = -?\d+\.\d{2}$`)

	l10Count := 0
	for i, step := range steps {
		if step.LineID != "L10" {
			continue
		}
		l10Count++

		tbl, ok := step.State.(renderers.TableState)
		if !ok {
			t.Fatalf("step %d (L10) state is not TableState", i)
		}

		if !captionRegex.MatchString(tbl.Caption) {
			t.Errorf("step %d (L10) caption %q does not match expected format", i, tbl.Caption)
		}

		hasActive := false
		var activeRow, activeCol int
		for _, cell := range tbl.Cells {
			if cell.State == "active" {
				hasActive = true
				activeRow = cell.Row
				activeCol = cell.Col
				break
			}
		}
		if !hasActive {
			t.Errorf("step %d (L10) missing active cell for updated (s, a)", i)
		}

		if len(tbl.Dependencies) > 0 {
			dep := tbl.Dependencies[0]
			if dep.ToRow != activeRow || dep.ToCol != activeCol {
				t.Errorf("step %d dependency target (%d,%d) does not match active cell (%d,%d)", i, dep.ToRow, dep.ToCol, activeRow, activeCol)
			}
			hasSource := false
			for _, cell := range tbl.Cells {
				if cell.Row == dep.FromRow && cell.Col == dep.FromCol && cell.State == "source" {
					hasSource = true
					break
				}
			}
			if !hasSource {
				t.Errorf("step %d dependency source (%d,%d) is not marked as source cell", i, dep.FromRow, dep.FromCol)
			}
		}

		if _, ok := step.Variables["tdError"]; !ok {
			t.Errorf("step %d missing 'tdError' variable", i)
		}
		if _, ok := step.Variables["tdTarget"]; !ok {
			t.Errorf("step %d missing 'tdTarget' variable", i)
		}
	}

	if l10Count == 0 {
		t.Fatal("expected at least one L10 Bellman update step in trace")
	}
}

func TestPolicyConvergence(t *testing.T) {
	env := DefaultGridWorld()
	agent := NewQAgent(env, 0.5, 0.9, 0.1)
	rng := rand.New(rand.NewSource(42))

	agent.Train(300, 25, rng)

	policy := agent.ExtractPolicy()
	if len(policy) != env.NumStates() {
		t.Fatalf("expected policy for %d states, got %d", env.NumStates(), len(policy))
	}

	startPolicy, ok := policy["(0,0)"]
	if !ok {
		t.Fatal("missing policy for start state (0,0)")
	}
	if startPolicy != "Right" && startPolicy != "Down" {
		t.Fatalf("expected start policy to be Right or Down, got %q", startPolicy)
	}

	f := agent.FormatPolicy()
	if !strings.Contains(f, "(0,0):") {
		t.Errorf("FormatPolicy output %q missing start state", f)
	}
}
