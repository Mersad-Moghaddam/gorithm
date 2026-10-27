package qlearning

import (
	"fmt"
	"math/rand"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
	"github.com/Mersad-Moghaddam/gorithm/web/renderers"
)

// cloneTableState creates an independent deep copy of a TableState snapshot.
func cloneTableState(s renderers.TableState) renderers.TableState {
	rowHeaders := make([]string, len(s.RowHeaders))
	copy(rowHeaders, s.RowHeaders)

	colHeaders := make([]string, len(s.ColHeaders))
	copy(colHeaders, s.ColHeaders)

	cells := make([]renderers.TableCell, len(s.Cells))
	copy(cells, s.Cells)

	var deps []renderers.TableDependency
	if len(s.Dependencies) > 0 {
		deps = make([]renderers.TableDependency, len(s.Dependencies))
		copy(deps, s.Dependencies)
	}

	return renderers.TableState{
		Caption:      s.Caption,
		Dependencies: deps,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
	}
}

func formatStateHeader(env *GridWorld, s Coord, agentPos Coord, bestA Action, hasPolicy bool) string {
	tag := ""
	if s == env.Start {
		tag = " S"
	} else if s == env.Goal {
		tag = " G"
	} else if env.Obstacles[s] {
		tag = " X"
	}
	if hasPolicy && s != env.Goal && !env.Obstacles[s] {
		tag += " " + ActionArrows[bestA]
	}
	if s == agentPos {
		tag += "*"
	}
	return fmt.Sprintf("(%d,%d)%s", s.Row, s.Col, tag)
}

// buildTableState formats the Q-table into a deep-copied renderers.TableState.
// ponytail: dense matrix allocated for states x actions; migrate to sparse map if state space exceeds 1000 cells.
func buildTableState(
	env *GridWorld,
	agent *QAgent,
	agentPos Coord,
	activeRow, activeCol int,
	sourceRow, sourceCol int,
	deps []renderers.TableDependency,
	caption string,
) renderers.TableState {
	rowHeaders := make([]string, env.NumStates())
	for i := 0; i < env.NumStates(); i++ {
		s := env.IndexToState(i)
		bestA, maxVal := agent.MaxQ(s)
		rowHeaders[i] = formatStateHeader(env, s, agentPos, bestA, maxVal > 0)
	}
	colHeaders := []string{"Up", "Right", "Down", "Left"}

	cells := make([]renderers.TableCell, 0, env.NumStates()*4)
	for r := 0; r < env.NumStates(); r++ {
		for c := 0; c < 4; c++ {
			st := "normal"
			if r == activeRow && c == activeCol {
				st = "active"
			} else if r == sourceRow && c == sourceCol {
				st = "source"
			}
			cells = append(cells, renderers.TableCell{
				Row:   r,
				Col:   c,
				Value: fmt.Sprintf("%.2f", agent.Q[r][c]),
				State: st,
			})
		}
	}

	return cloneTableState(renderers.TableState{
		Caption:      caption,
		Dependencies: deps,
		RowHeaders:   rowHeaders,
		ColHeaders:   colHeaders,
		Cells:        cells,
	})
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
	initTable := buildTableState(env, agent, env.Start, -1, -1, -1, -1, nil, "Initialize Q-table with zeros for all state-action pairs")
	b.AddStep("L2", trace.ActionInit, "Initialize Q-table with zeros for all state-action pairs.", initTable, map[string]any{
		"state":   fmt.Sprintf("(%d,%d)", env.Start.Row, env.Start.Col),
		"action":  "-",
		"reward":  0.0,
		"episode": 0,
		"policy":  agent.FormatPolicy(),
	})

	maxStepsPerEpisode := 12

	for ep := 0; ep < episodes; ep++ {
		// L3: Outer episode loop
		b.AddStep("L3", trace.ActionHighlight, fmt.Sprintf("Starting episode %d of %d.", ep+1, episodes),
			buildTableState(env, agent, env.Start, -1, -1, -1, -1, nil, fmt.Sprintf("Episode %d of %d", ep+1, episodes)),
			map[string]any{
				"state":   fmt.Sprintf("(%d,%d)", env.Start.Row, env.Start.Col),
				"action":  "-",
				"reward":  0.0,
				"episode": ep + 1,
				"policy":  agent.FormatPolicy(),
			},
		)

		state := env.Start
		sIdx := env.StateIndex(state)

		// L4: Reset environment
		b.AddStep("L4", trace.ActionSet, fmt.Sprintf("Reset agent to start state (%d,%d).", state.Row, state.Col),
			buildTableState(env, agent, state, sIdx, -1, -1, -1, nil, fmt.Sprintf("Reset agent to (%d,%d)", state.Row, state.Col)),
			map[string]any{
				"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
				"action":  "-",
				"reward":  0.0,
				"episode": ep + 1,
				"policy":  agent.FormatPolicy(),
			},
		)

		for step := 0; step < maxStepsPerEpisode; step++ {
			// L5: Terminal check
			if env.IsTerminal(state) {
				b.AddStep("L5", trace.ActionDone, fmt.Sprintf("State (%d,%d) is goal. Episode %d finished in %d steps.", state.Row, state.Col, ep+1, step),
					buildTableState(env, agent, state, env.StateIndex(state), -1, -1, -1, nil, fmt.Sprintf("Goal reached at (%d,%d)!", state.Row, state.Col)),
					map[string]any{
						"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
						"action":  "-",
						"reward":  0.0,
						"episode": ep + 1,
						"policy":  agent.FormatPolicy(),
					},
				)
				break
			}

			b.AddStep("L5", trace.ActionCompare, fmt.Sprintf("State (%d,%d) is not goal. Proceed with step %d.", state.Row, state.Col, step+1),
				buildTableState(env, agent, state, env.StateIndex(state), -1, -1, -1, nil, fmt.Sprintf("Agent at (%d,%d) - evaluating step %d", state.Row, state.Col, step+1)),
				map[string]any{
					"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":  "-",
					"reward":  0.0,
					"episode": ep + 1,
					"policy":  agent.FormatPolicy(),
				},
			)

			// L6: Choose action
			action, mode := agent.ChooseActionWithMode(state, rng)
			sIdx = env.StateIndex(state)
			chooseCaption := fmt.Sprintf("State (%d,%d): chose %s via %s (ε = %.2f)", state.Row, state.Col, ActionNames[action], mode, agent.Epsilon)
			b.AddStep("L6", trace.ActionHighlight, fmt.Sprintf("State (%d,%d): selected action %s via %s (epsilon-greedy, rate=%.2f).", state.Row, state.Col, ActionNames[action], mode, agent.Epsilon),
				buildTableState(env, agent, state, sIdx, int(action), -1, -1, nil, chooseCaption),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"selection": mode,
					"epsilon":   agent.Epsilon,
					"reward":    0.0,
					"episode":   ep + 1,
					"policy":    agent.FormatPolicy(),
				},
			)

			// L7: Environment step
			nextState, reward, done := env.Step(state, action)
			b.IncOperations()
			stepCaption := fmt.Sprintf("Step %s -> next state (%d,%d), reward = %d", ActionNames[action], nextState.Row, nextState.Col, int(reward))
			b.AddStep("L7", trace.ActionSet, fmt.Sprintf("Executed action %s -> next state (%d,%d), reward = %.1f.", ActionNames[action], nextState.Row, nextState.Col, reward),
				buildTableState(env, agent, state, sIdx, int(action), -1, -1, nil, stepCaption),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"reward":    reward,
					"nextState": fmt.Sprintf("(%d,%d)", nextState.Row, nextState.Col),
					"episode":   ep + 1,
					"policy":    agent.FormatPolicy(),
				},
			)

			// L8: Max next Q
			bestNextA, bestNextQ := agent.MaxQ(nextState)
			if done {
				bestNextQ = 0.0
			}
			nextIdx := env.StateIndex(nextState)
			srcA := int(bestNextA)
			srcRow := nextIdx
			srcCol := srcA
			var deps []renderers.TableDependency
			if done || (srcRow == sIdx && srcCol == int(action)) {
				srcRow = -1
				srcCol = -1
			} else {
				deps = []renderers.TableDependency{
					{
						FromRow: nextIdx,
						FromCol: srcA,
						ToRow:   sIdx,
						ToCol:   int(action),
					},
				}
			}

			maxQCaption := fmt.Sprintf("Next state (%d,%d): max future Q(s', a') = %.2f via %s", nextState.Row, nextState.Col, bestNextQ, ActionNames[bestNextA])
			if done {
				maxQCaption = fmt.Sprintf("Next state (%d,%d) is goal: future reward is 0.00", nextState.Row, nextState.Col)
			}
			b.AddStep("L8", trace.ActionCompare, fmt.Sprintf("Next state (%d,%d): max future Q(s', a') = %.2f.", nextState.Row, nextState.Col, bestNextQ),
				buildTableState(env, agent, state, sIdx, int(action), srcRow, srcCol, nil, maxQCaption),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"reward":    reward,
					"bestNextQ": bestNextQ,
					"episode":   ep + 1,
					"policy":    agent.FormatPolicy(),
				},
			)

			// L9: TD target
			tdTarget := reward + agent.Gamma*bestNextQ
			oldQ := agent.Q[sIdx][action]
			tdError := tdTarget - oldQ
			tdCaption := fmt.Sprintf("TD Target = r + γ·max Q(s',a') = %.1f + %.2f·%.2f = %.2f; TD Error = %.2f", reward, agent.Gamma, bestNextQ, tdTarget, tdError)
			b.AddStep("L9", trace.ActionSet, fmt.Sprintf("Calculated TD target: %.1f + %.2f * %.2f = %.2f (TD error: %.2f).", reward, agent.Gamma, bestNextQ, tdTarget, tdError),
				buildTableState(env, agent, state, sIdx, int(action), srcRow, srcCol, deps, tdCaption),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"reward":    reward,
					"tdTarget":  tdTarget,
					"tdError":   tdError,
					"bestNextQ": bestNextQ,
					"episode":   ep + 1,
					"policy":    agent.FormatPolicy(),
				},
			)

			// L10: Bellman update
			newQ, _ := agent.BellmanUpdate(state, action, reward, nextState, done)
			b.IncOperations()
			caption := fmt.Sprintf("Agent took action %s from state %s -> got reward %d. TD error = %.2f; updated Q(%s, %s) = %.2f",
				ActionNames[action],
				fmt.Sprintf("(%d,%d)", state.Row, state.Col),
				int(reward),
				tdError,
				fmt.Sprintf("(%d,%d)", state.Row, state.Col),
				ActionNames[action],
				newQ,
			)
			b.AddStep("L10", trace.ActionSet, fmt.Sprintf("Bellman update Q((%d,%d), %s): %.2f -> %.2f.", state.Row, state.Col, ActionNames[action], oldQ, newQ),
				buildTableState(env, agent, state, sIdx, int(action), srcRow, srcCol, deps, caption),
				map[string]any{
					"state":     fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":    ActionNames[action],
					"reward":    reward,
					"oldQ":      oldQ,
					"newQ":      newQ,
					"tdTarget":  tdTarget,
					"tdError":   tdError,
					"nextState": fmt.Sprintf("(%d,%d)", nextState.Row, nextState.Col),
					"episode":   ep + 1,
					"policy":    agent.FormatPolicy(),
				},
			)

			// L11: Update state
			state = nextState
			b.AddStep("L11", trace.ActionSet, fmt.Sprintf("Agent advanced to state (%d,%d).", state.Row, state.Col),
				buildTableState(env, agent, state, env.StateIndex(state), -1, -1, -1, nil, fmt.Sprintf("Agent transitioned to (%d,%d)", state.Row, state.Col)),
				map[string]any{
					"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
					"action":  ActionNames[action],
					"reward":  reward,
					"episode": ep + 1,
					"policy":  agent.FormatPolicy(),
				},
			)

			if done {
				b.AddStep("L5", trace.ActionDone, fmt.Sprintf("Goal reached at (%d,%d)! Episode %d finished.", state.Row, state.Col, ep+1),
					buildTableState(env, agent, state, env.StateIndex(state), -1, -1, -1, nil, fmt.Sprintf("Goal (%d,%d) reached! Episode %d ended", state.Row, state.Col, ep+1)),
					map[string]any{
						"state":   fmt.Sprintf("(%d,%d)", state.Row, state.Col),
						"action":  ActionNames[action],
						"reward":  reward,
						"episode": ep + 1,
						"policy":  agent.FormatPolicy(),
					},
				)
				break
			}
		}
	}

	// L14: Complete
	b.AddStep("L14", trace.ActionDone, fmt.Sprintf("Q-learning completed across %d episodes. Converged policy: %s", episodes, agent.FormatPolicy()),
		buildTableState(env, agent, Coord{Row: -1, Col: -1}, -1, -1, -1, -1, nil, fmt.Sprintf("Training complete. Converged policy: %s", agent.FormatPolicy())),
		map[string]any{
			"state":    fmt.Sprintf("(%d,%d)", env.Start.Row, env.Start.Col),
			"action":   "-",
			"reward":   0.0,
			"episodes": episodes,
			"policy":   agent.FormatPolicy(),
		},
	)

	return b.Steps()
}
