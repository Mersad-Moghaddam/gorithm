package qlearning

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func qlearningRateValid(x float64) bool {
	return !math.IsNaN(x) && !math.IsInf(x, 0) && x > 0 && x <= 1
}

func qlearningEnvValid(env *GridWorld) bool {
	if env.Rows < 1 || env.Cols < 1 || env.Rows > 64 || env.Cols > 64 || env.Rows*env.Cols > 64 {
		return false
	}
	inside := func(s Coord) bool { return s.Row >= 0 && s.Row < env.Rows && s.Col >= 0 && s.Col < env.Cols }
	if !inside(env.Start) || !inside(env.Goal) || env.Obstacles[env.Start] || env.Obstacles[env.Goal] {
		return false
	}
	for p := range env.Obstacles {
		if !inside(p) {
			return false
		}
	}
	return true
}

func Trace(env *GridWorld, episodes int, alpha, gamma, epsilon float64) []trace.Step {
	return TraceWithSeed(env, episodes, alpha, gamma, epsilon, 7)
}

func TraceWithSeed(env *GridWorld, episodes int, alpha, gamma, epsilon float64, seed int64) []trace.Step {
	if env == nil {
		env = DefaultGridWorld()
	}
	if episodes <= 0 {
		episodes = 4
	}
	if episodes > 20 {
		episodes = 20
	}
	// Preserve direct API's documented nonpositive defaults; Build rejects explicit zero.
	if alpha <= 0 && !math.IsInf(alpha, 0) {
		alpha = .5
	}
	if gamma <= 0 && !math.IsInf(gamma, 0) {
		gamma = .9
	}
	if epsilon <= 0 && !math.IsInf(epsilon, 0) {
		epsilon = .2
	}
	b := trace.NewBuilder()
	if !qlearningEnvValid(env) || !qlearningRateValid(alpha) || !qlearningRateValid(gamma) || !qlearningRateValid(epsilon) {
		b.AddStep("L2", trace.ActionDone, "Invalid input: grid must contain 1–64 cells with valid unblocked start/goal and rates in (0,1].", renderers.QLearningLessonState{Phase: "invalid-input", Action: -1, BestNextAction: -1}, map[string]any{"state": "invalid", "action": "-", "reward": 0.0})
		return b.Steps()
	}
	agent := NewQAgent(env, alpha, gamma, epsilon)
	rng := rand.New(rand.NewSource(seed))
	coord := func(p Coord) [2]int { return [2]int{p.Row, p.Col} }
	s := renderers.QLearningLessonState{Rows: env.Rows, Cols: env.Cols, Start: coord(env.Start), Goal: coord(env.Goal), State: coord(env.Start), NextState: coord(env.Start), Episodes: episodes, Alpha: alpha, Gamma: gamma, Epsilon: epsilon, Action: -1, BestNextAction: -1}
	// Row-major traversal fixes obstacle order without depending on map iteration.
	for r := 0; r < env.Rows; r++ {
		for c := 0; c < env.Cols; c++ {
			if env.Obstacles[Coord{r, c}] {
				s.Obstacles = append(s.Obstacles, [2]int{r, c})
			}
		}
	}
	add := func(line string, action trace.ActionType, phase, reason string) {
		s.Phase = phase
		snapshot := s
		snapshot.Obstacles = append([][2]int(nil), s.Obstacles...)
		snapshot.Q = make([][]float64, len(agent.Q))
		table := renderers.TableState{ColHeaders: []string{"Up", "Right", "Down", "Left"}}
		for i, row := range agent.Q {
			snapshot.Q[i] = append([]float64(nil), row...)
			p := env.IndexToState(i)
			label := fmt.Sprintf("(%d,%d)", p.Row, p.Col)
			if p == env.Start {
				label += " S"
			}
			if p == env.Goal {
				label += " G"
			}
			if env.Obstacles[p] {
				label += " X"
			}
			table.RowHeaders = append(table.RowHeaders, label)
			for a, v := range row {
				kind := "normal"
				if coord(p) == s.State && a == s.Action {
					kind = "active"
				}
				if phase == "update" && coord(p) == s.State && a == s.Action {
					kind = "computed"
				}
				table.Cells = append(table.Cells, renderers.TableCell{Row: i, Col: a, Value: fmt.Sprintf("%.2f", v), State: kind})
			}
		}
		snapshot.TableState = table
		actionName := "-"
		if s.Action >= 0 {
			actionName = ActionNames[s.Action]
		}
		b.AddStep(line, action, reason, snapshot, map[string]any{"state": fmt.Sprintf("(%d,%d)", s.State[0], s.State[1]), "action": actionName, "reward": s.Reward, "episode": s.Episode, "episodeStep": s.EpisodeStep, "nextState": s.NextState, "bestNextQ": s.BestNextQ, "tdTarget": s.TDTarget, "oldQ": s.OldQ, "newQ": s.NewQ})
	}
	clearTransition := func() {
		s.Action = -1
		s.BestNextAction = -1
		s.Reward = 0
		s.BestNextQ = 0
		s.TDTarget = 0
		s.OldQ = 0
		s.NewQ = 0
		s.TransitionValid = false
		s.UpdateValid = false
		s.Terminal = false
		s.NextState = s.State
	}
	add("L2", trace.ActionInit, "initialize", "Initialize all state-action estimates to zero; ties are not a learned route.")
	for ep := 1; ep <= episodes; ep++ {
		s.Episode = ep
		s.EpisodeStep = 0
		s.EpisodeStopReason = ""
		clearTransition()
		add("L3", trace.ActionHighlight, "episode", fmt.Sprintf("Start episode %d of %d. Keep learned Q estimates.", ep, episodes))
		state := env.Start
		s.State = coord(state)
		clearTransition()
		add("L4", trace.ActionSet, "reset", "Reset agent to S without resetting Q.")
		for n := 0; n < 12; n++ {
			clearTransition()
			s.EpisodeStep = n
			if env.IsTerminal(state) {
				s.Terminal = true
				s.EpisodeStopReason = "goal"
				s.CompletedEpisodes = ep
				add("L5", trace.ActionDone, "episode-end", "Agent is at goal; episode ends.")
				break
			}
			add("L5", trace.ActionCompare, "terminal-check", "Current state is not goal; select an action.")
			action := agent.ChooseAction(state, rng)
			s.Action = int(action)
			s.EpisodeStep = n + 1
			s.OldQ = agent.Q[env.StateIndex(state)][action]
			add("L6", trace.ActionHighlight, "select-action", fmt.Sprintf("Epsilon-greedy selected action %s; no claim of a unique optimal route.", ActionNames[action]))
			next, reward, done := env.Step(state, action)
			s.NextState = coord(next)
			s.Reward = reward
			s.Terminal = done
			s.TransitionValid = true
			b.IncOperations()
			add("L7", trace.ActionSet, "transition", fmt.Sprintf("%s reaches (%d,%d), reward %g. Hollow marker previews next position; agent stays until assignment.", ActionNames[action], next.Row, next.Col, reward))
			if !done {
				a, q := agent.MaxQ(next)
				s.BestNextAction = int(a)
				s.BestNextQ = q
			}
			add("L8", trace.ActionCompare, "future", fmt.Sprintf("Future maximum = %g; terminal states contribute zero.", s.BestNextQ))
			s.TDTarget = reward + gamma*s.BestNextQ
			add("L9", trace.ActionSet, "target", fmt.Sprintf("TD target = %g + %g × %g = %g.", reward, gamma, s.BestNextQ, s.TDTarget))
			s.NewQ, _ = agent.BellmanUpdate(state, action, reward, next, done)
			s.UpdateValid = true
			b.IncOperations()
			add("L10", trace.ActionSet, "update", fmt.Sprintf("Update one Q cell: %g + %g × (%g − %g) = %g.", s.OldQ, alpha, s.TDTarget, s.OldQ, s.NewQ))
			state = next
			s.State = coord(state)
			add("L11", trace.ActionSet, "assign-state", "Assign next state to current state. Agent now moves to recorded endpoint.")
			if done {
				s.EpisodeStopReason = "goal"
				s.CompletedEpisodes = ep
				add("L5", trace.ActionDone, "episode-end", fmt.Sprintf("Goal reached after %d transitions; episode ends.", n+1))
				break
			}
			if n == 11 {
				s.EpisodeStopReason = "step-limit"
				s.CompletedEpisodes = ep
				add("L12", trace.ActionHighlight, "episode-end", "12-transition budget exhausted; goal not reached. Keep Q estimates for next episode.")
			}
		}
	}
	add("L14", trace.ActionDone, "complete", fmt.Sprintf("%d episodes completed; estimates learned. Finite training does not prove convergence or optimality.", episodes))
	steps := b.Steps()
	for i := range steps {
		if steps[i].LineID == "L9" && i+1 < len(steps) && steps[i+1].LineID == "L10" {
			steps[i].HoldMS = 600
			steps[i+1].HoldMS = 1200
		}
	}
	return steps
}
