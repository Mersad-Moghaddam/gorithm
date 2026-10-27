package qlearning

import (
	"fmt"
	"math/rand"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

func formatStateHeader(env *GridWorld, s Coord) string {
	if s == env.Start {
		return fmt.Sprintf("(%d,%d) S", s.Row, s.Col)
	}
	if s == env.Goal {
		return fmt.Sprintf("(%d,%d) G", s.Row, s.Col)
	}
	if env.Obstacles[s] {
		return fmt.Sprintf("(%d,%d) X", s.Row, s.Col)
	}
	return fmt.Sprintf("(%d,%d)", s.Row, s.Col)
}

// buildTableState formats the Q-table into a renderers.TableState.
// ponytail: dense matrix allocated for states x actions; migrate to sparse map if state space exceeds 1000 cells.
func buildTableState(
	env *GridWorld,
	q [][]float64,
	activeRow, activeCol int,
	sourceRow, sourceCol int,
	updatedRow, updatedCol int,
) renderers.TableState {
	rowHeaders := make([]string, env.NumStates())
	for i := 0; i < env.NumStates(); i++ {
		s := env.IndexToState(i)
		rowHeaders[i] = formatStateHeader(env, s)
	}
	colHeaders := []string{"Up", "Right", "Down", "Left"}

	cells := make([]renderers.TableCell, 0, env.NumStates()*4)
	for r := 0; r < env.NumStates(); r++ {
		for c := 0; c < 4; c++ {
			st := "normal"
			if r == updatedRow && c == updatedCol {
				st = "computed"
			} else if r == activeRow && c == activeCol {
				st = "active"
			} else if r == sourceRow && c == sourceCol {
				st = "source"
			}
			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: fmt.Sprintf("%.2f", q[r][c]),
				State: st,
			})
		}
	}

	return renderers.TableState{
		RowHeaders: rowHeaders,
		ColHeaders: colHeaders,
		Cells:      cells,
	}
}

// Trace executes Q-learning with a default deterministic seed and returns step-by-step trace data.
func Trace(env *GridWorld, episodes int, alpha, gamma, epsilon float64) []trace.Step {
	return TraceWithSeed(env, episodes, alpha, gamma, epsilon, 7)
}

// TraceWithSeed runs Q-learning with a specified random seed.
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
	if alpha <= 0 {
		alpha = 0.5
	}
	if gamma <= 0 {
		gamma = 0.9
	}
	if epsilon <= 0 {
		epsilon = 0.2
	}

	b := trace.NewBuilder()
	rng := rand.New(rand.NewSource(seed))
	agent := NewQAgent(env, alpha, gamma, epsilon)

	// L2: Initialize Q-table
	initTable := buildTableState(env, agent.Q, -1, -1, -1, -1, -1, -1)
	b.AddStep("L2", trace.ActionInit, "Initialize Q-table with zeros for all state-action pairs.", initTable, map[string]any{
		"state":   fmt.Sprintf("(%d,%d)", env.Start.Row, env.Start.Col),
		"action":  "-",
		"reward":  0.0,
		"episode": 0,
	})

	maxStepsPerEpisode := 12

	for ep := 0; ep < episodes; ep++ {
		// L3: Outer episode loop
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Starting episode %d of %d.", ep+1, episodes),
			buildTableState(env, agent.Q, -1, -1, -1, -1, -1, -1),
			map[string]any{
				"state":   fmt.Sprintf("(%d,%d)", env.Start.Row, env.Start.Col),
				"action":  "-",
				"reward":  0.0,
				"episode": ep + 1,
			},
		)

		state := env.Start
		sIdx := env.StateIndex(state)

		// L4: Reset environment
		b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Reset agent to start state (%d,%d).", state.Row, state.Col),
			buildTableState(env, agent.Q, sIdx, -1, -1, -1, -1, -1),
			map[string]any{
				"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
				"action":  "-",
				"reward":  0.0,
				"episode": ep + 1,
			},
		)

		for step := 0; step < maxStepsPerEpisode; step++ {
			// L5: Terminal check
			if env.IsTerminal(state) {
				b.AddStep("L5", trace.ActionDone, fmt.Sprintf("State (%d,%d) is goal. Episode %d finished in %d steps.", state.Row, state.Col, ep+1, step),
					buildTableState(env, agent.Q, env.StateIndex(state), -1, -1, -1, -1, -1),
					map[string]any{
						"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
						"action":  "-",
						"reward":  0.0,
						"episode": ep + 1,
					},
				)
				break
			}

			b.AddStep("L5", trace.ActionCompare, fmt.Sprintf("State (%d,%d) is not goal. Proceed with step %d.", state.Row, state.Col, step+1),
				buildTableState(env, agent.Q, env.StateIndex(state), -1, -1, -1, -1, -1),
				map[string]any{
					"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":  "-",
					"reward":  0.0,
					"episode": ep + 1,
				},
			)

			// L6: Choose action
			action := agent.ChooseAction(state, rng)
			sIdx = env.StateIndex(state)
			b.AddStep("L6", trace.ActionHighlight, fmt.Sprintf("State (%d,%d): selected action %s (epsilon-greedy).", state.Row, state.Col, ActionNames[action]),
				buildTableState(env, agent.Q, sIdx, int(action), -1, -1, -1, -1),
				map[string]any{
					"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":  ActionNames[action],
					"reward":  0.0,
					"episode": ep + 1,
				},
			)

			// L7: Environment step
			nextState, reward, done := env.Step(state, action)
			b.IncOperations()
			b.AddStep("L7", trace.ActionSet, fmt.Sprintf("Executed action %s -> next state (%d,%d), reward = %.1f.", ActionNames[action], nextState.Row, nextState.Col, reward),
				buildTableState(env, agent.Q, sIdx, int(action), -1, -1, -1, -1),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"reward":    reward,
					"nextState": fmt.Sprintf("(%d,%d)", nextState.Row, nextState.Col),
					"episode":   ep + 1,
				},
			)

			// L8: Max next Q
			bestNextA, bestNextQ := agent.MaxQ(nextState)
			if done {
				bestNextQ = 0.0
			}
			nextIdx := env.StateIndex(nextState)
			srcA := int(bestNextA)
			if done {
				srcA = -1
			}
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Next state (%d,%d): max future Q(s', a') = %.2f.", nextState.Row, nextState.Col, bestNextQ),
				buildTableState(env, agent.Q, sIdx, int(action), nextIdx, srcA, -1, -1),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"reward":    reward,
					"bestNextQ": bestNextQ,
					"episode":   ep + 1,
				},
			)

			// L9: TD target
			tdTarget := reward + agent.Gamma*bestNextQ
			b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Calculated TD target: %.1f + %.2f * %.2f = %.2f.", reward, agent.Gamma, bestNextQ, tdTarget),
				buildTableState(env, agent.Q, sIdx, int(action), nextIdx, srcA, -1, -1),
				map[string]any{
					"state":    fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":   ActionNames[action],
					"reward":   reward,
					"tdTarget": tdTarget,
					"episode":  ep + 1,
				},
			)

			// L10: Bellman update
			oldQ := agent.Q[sIdx][action]
			newQ, _ := agent.BellmanUpdate(state, action, reward, nextState, done)
			b.IncOperations()
			b.AddStep("L10", trace.ActionSet, fmt.Sprintf("Bellman update Q((%d,%d), %s): %.2f -> %.2f.", state.Row, state.Col, ActionNames[action], oldQ, newQ),
				buildTableState(env, agent.Q, -1, -1, -1, -1, sIdx, int(action)),
				map[string]any{
					"state":    fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":   ActionNames[action],
					"reward":   reward,
					"oldQ":     oldQ,
					"newQ":     newQ,
					"tdTarget": tdTarget,
					"episode":  ep + 1,
				},
			)

			// L11: Update state
			state = nextState
			b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Agent advanced to state (%d,%d).", state.Row, state.Col),
				buildTableState(env, agent.Q, env.StateIndex(state), -1, -1, -1, -1, -1),
				map[string]any{
					"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":  ActionNames[action],
					"reward":  reward,
					"episode": ep + 1,
				},
			)

			if done {
				b.AddStep("L5", trace.ActionDone, fmt.Sprintf("Goal reached at (%d,%d)! Episode %d finished.", state.Row, state.Col, ep+1),
					buildTableState(env, agent.Q, env.StateIndex(state), -1, -1, -1, -1, -1),
					map[string]any{
						"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
						"action":  ActionNames[action],
						"reward":  reward,
						"episode": ep + 1,
					},
				)
				break
			}
		}
	}

	// L14: Complete
	b.AddStep("L14", trace.ActionDone, fmt.Sprintf("Q-learning completed across %d episodes. Q-table values learned.", episodes),
		buildTableState(env, agent.Q, -1, -1, -1, -1, -1, -1),
		map[string]any{
			"state":    fmt.Sprintf("(%d,%d)", env.Start.Row, env.Start.Col),
			"action":   "-",
			"reward":   0.0,
			"episodes": episodes,
		},
	)

	return b.Steps()
}
