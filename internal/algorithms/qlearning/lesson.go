package qlearning

import (
	"encoding/json"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	Episodes int     `json:"episodes,omitempty"`
	Alpha    float64 `json:"alpha,omitempty"`
	Gamma    float64 `json:"gamma,omitempty"`
	Epsilon  float64 `json:"epsilon,omitempty"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "qlearning",
		Name:     "Q-Learning (Reinforcement Learning)",
		Category: catalog.CategoryAI,
		Renderer: "table",
		Documentation: catalog.Documentation{
			Summary:         "A model-free reinforcement learning algorithm that learns the optimal action-value function Q(s, a) using the Bellman equation and Temporal Difference updates.",
			TimeComplexity:  "O(episodes · steps)",
			SpaceComplexity: "O(|S| · |A|)",
			Difficulty:      "Hard",
			KeyConcepts: []string{
				"Temporal Difference",
				"Exploration vs Exploitation",
				"Bellman Optimality",
				"Policy Extraction",
			},
			HowItWorks: "Maintains a lookup table Q(s, a) for all state-action pairs. In each training episode, an agent traverses the environment choosing actions with an epsilon-greedy policy balancing exploration and exploitation. At each transition (s, a, r, s'), it calculates the TD target r + gamma * max Q(s', a') and updates Q(s, a) via Q(s, a) <- Q(s, a) + alpha * [TD Target - Q(s, a)]. Values propagate backwards from the goal to form an optimal navigation policy.",
		},
		DefaultInput: `{"episodes": 5, "alpha": 0.5, "gamma": 0.9, "epsilon": 0.3}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func QLearning(env Environment, episodes int) QTable {"},
		{ID: "L2", Number: 2, Code: "    Q := initTable(env.States, env.Actions)"},
		{ID: "L3", Number: 3, Code: "    for ep := 0; ep < episodes; ep++ {"},
		{ID: "L4", Number: 4, Code: "        state := env.Reset()"},
		{ID: "L5", Number: 5, Code: "        for !env.IsTerminal(state) {"},
		{ID: "L6", Number: 6, Code: "            action := chooseAction(state, Q, epsilon)"},
		{ID: "L7", Number: 7, Code: "            nextState, reward := env.Step(action)"},
		{ID: "L8", Number: 8, Code: "            bestNextQ := max(Q[nextState])"},
		{ID: "L9", Number: 9, Code: "            tdTarget := reward + gamma * bestNextQ"},
		{ID: "L10", Number: 10, Code: "            Q[state][action] += alpha * (tdTarget - Q[state][action])"},
		{ID: "L11", Number: 11, Code: "            state = nextState"},
		{ID: "L12", Number: 12, Code: "        }"},
		{ID: "L13", Number: 13, Code: "    }"},
		{ID: "L14", Number: 14, Code: "    return Q"},
		{ID: "L15", Number: 15, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	episodes := 5
	alpha := 0.5
	gamma := 0.9
	epsilon := 0.3

	if len(input) > 0 {
		var custom LessonInput
		if err := json.Unmarshal(input, &custom); err == nil {
			if custom.Episodes > 0 {
				episodes = custom.Episodes
			}
			if custom.Alpha > 0 {
				alpha = custom.Alpha
			}
			if custom.Gamma > 0 {
				gamma = custom.Gamma
			}
			if custom.Epsilon > 0 {
				epsilon = custom.Epsilon
			}
		}
	}

	env := DefaultGridWorld()
	return Trace(env, episodes, alpha, gamma, epsilon), nil
}

func init() {
	catalog.Register(Lesson{})
}
