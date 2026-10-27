package qlearning

import (
	"fmt"
	"math/rand"
	"strings"
)

// Action represents an agent movement in the grid.
type Action int

const (
	ActionUp Action = iota
	ActionRight
	ActionDown
	ActionLeft
)

var ActionNames = [4]string{"Up", "Right", "Down", "Left"}
var ActionArrows = [4]string{"↑", "→", "↓", "←"}

// Coord represents a 2D coordinate (row, col) in the grid world.
type Coord struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// GridWorld models a discrete 2D grid environment with start, goal, and obstacles.
type GridWorld struct {
	Rows      int            `json:"rows"`
	Cols      int            `json:"cols"`
	Start     Coord          `json:"start"`
	Goal      Coord          `json:"goal"`
	Obstacles map[Coord]bool `json:"obstacles"`
}

// NewGridWorld constructs a GridWorld environment.
func NewGridWorld(rows, cols int, start, goal Coord, obstacles []Coord) *GridWorld {
	obsMap := make(map[Coord]bool, len(obstacles))
	for _, o := range obstacles {
		obsMap[o] = true
	}
	return &GridWorld{
		Rows:      rows,
		Cols:      cols,
		Start:     start,
		Goal:      goal,
		Obstacles: obsMap,
	}
}

// DefaultGridWorld returns standard 3x3 grid world:
// Start (0,0), Goal (2,2), Obstacle at (1,1).
func DefaultGridWorld() *GridWorld {
	return NewGridWorld(3, 3, Coord{Row: 0, Col: 0}, Coord{Row: 2, Col: 2}, []Coord{{Row: 1, Col: 1}})
}

// NumStates returns total state count.
func (g *GridWorld) NumStates() int {
	return g.Rows * g.Cols
}

// StateIndex maps (Row, Col) to a flattened integer state index.
func (g *GridWorld) StateIndex(s Coord) int {
	return s.Row*g.Cols + s.Col
}

// IndexToState maps flattened index back to Coord.
func (g *GridWorld) IndexToState(idx int) Coord {
	return Coord{Row: idx / g.Cols, Col: idx % g.Cols}
}

// IsTerminal checks if state is the goal.
func (g *GridWorld) IsTerminal(s Coord) bool {
	return s == g.Goal
}

// Step applies an action from state s, returning nextState, reward, and terminal status.
func (g *GridWorld) Step(s Coord, a Action) (Coord, float64, bool) {
	if g.IsTerminal(s) {
		return s, 0.0, true
	}

	next := s
	switch a {
	case ActionUp:
		next.Row--
	case ActionRight:
		next.Col++
	case ActionDown:
		next.Row++
	case ActionLeft:
		next.Col--
	}

	// Boundary or obstacle collision: bounce back to original state
	if next.Row < 0 || next.Row >= g.Rows || next.Col < 0 || next.Col >= g.Cols || g.Obstacles[next] {
		next = s
	}

	if g.IsTerminal(next) {
		return next, 10.0, true
	}
	return next, 0.0, false
}

// QAgent implements tabular Q-learning with Bellman equation updates.
// ponytail: tabular lookup used for discrete grid states; swap with deep Q-network (DQN) if state space explodes.
type QAgent struct {
	Env     *GridWorld
	Q       [][]float64 // [stateIndex][action]
	Alpha   float64     // Learning rate
	Gamma   float64     // Discount factor
	Epsilon float64     // Exploration rate
}

// NewQAgent initializes Q-table with zeros and sets hyperparameters.
func NewQAgent(env *GridWorld, alpha, gamma, epsilon float64) *QAgent {
	numStates := env.NumStates()
	q := make([][]float64, numStates)
	for i := range q {
		q[i] = make([]float64, 4)
	}
	return &QAgent{
		Env:     env,
		Q:       q,
		Alpha:   alpha,
		Gamma:   gamma,
		Epsilon: epsilon,
	}
}

// BestActions returns all actions achieving maximum Q-value for state s.
func (ag *QAgent) BestActions(s Coord) ([]Action, float64) {
	sIdx := ag.Env.StateIndex(s)
	maxVal := ag.Q[sIdx][0]
	for a := 1; a < 4; a++ {
		if ag.Q[sIdx][a] > maxVal {
			maxVal = ag.Q[sIdx][a]
		}
	}
	var best []Action
	for a := 0; a < 4; a++ {
		if ag.Q[sIdx][a] == maxVal {
			best = append(best, Action(a))
		}
	}
	return best, maxVal
}

// MaxQ returns best action and its maximum Q-value for state s.
func (ag *QAgent) MaxQ(s Coord) (Action, float64) {
	best, maxVal := ag.BestActions(s)
	return best[0], maxVal
}

// ChooseActionWithMode selects action using epsilon-greedy exploration,
// returning both chosen action and strategy name ("exploration" or "exploitation").
func (ag *QAgent) ChooseActionWithMode(s Coord, rng *rand.Rand) (Action, string) {
	if rng != nil && rng.Float64() < ag.Epsilon {
		return Action(rng.Intn(4)), "exploration"
	}
	best, _ := ag.BestActions(s)
	if rng != nil && len(best) > 1 {
		return best[rng.Intn(len(best))], "exploitation"
	}
	return best[0], "exploitation"
}

// ChooseAction selects action using epsilon-greedy exploration with random tie-breaking.
func (ag *QAgent) ChooseAction(s Coord, rng *rand.Rand) Action {
	a, _ := ag.ChooseActionWithMode(s, rng)
	return a
}

// ExtractPolicy returns the best action name for each reachable state based on the current Q-table.
func (ag *QAgent) ExtractPolicy() map[string]string {
	policy := make(map[string]string, ag.Env.NumStates())
	for r := 0; r < ag.Env.Rows; r++ {
		for c := 0; c < ag.Env.Cols; c++ {
			coord := Coord{Row: r, Col: c}
			key := fmt.Sprintf("(%d,%d)", r, c)
			if coord == ag.Env.Goal {
				policy[key] = "Goal"
				continue
			}
			if ag.Env.Obstacles[coord] {
				policy[key] = "Obstacle"
				continue
			}
			bestA, maxVal := ag.MaxQ(coord)
			if maxVal > 0 {
				policy[key] = ActionNames[bestA]
			} else {
				policy[key] = "none"
			}
		}
	}
	return policy
}

// FormatPolicy returns a compact string of policy arrows for active states.
func (ag *QAgent) FormatPolicy() string {
	var parts []string
	for r := 0; r < ag.Env.Rows; r++ {
		for c := 0; c < ag.Env.Cols; c++ {
			coord := Coord{Row: r, Col: c}
			if coord == ag.Env.Goal || ag.Env.Obstacles[coord] {
				continue
			}
			bestA, maxVal := ag.MaxQ(coord)
			arrow := "·"
			if maxVal > 0 {
				arrow = ActionArrows[bestA]
			}
			parts = append(parts, fmt.Sprintf("(%d,%d):%s", r, c, arrow))
		}
	}
	return strings.Join(parts, " ")
}

// BellmanUpdate applies the Bellman optimality update:
// Q(s, a) <- Q(s, a) + alpha * [reward + gamma * max_a' Q(s', a') - Q(s, a)]
func (ag *QAgent) BellmanUpdate(s Coord, a Action, reward float64, next Coord, done bool) (newQ float64, tdTarget float64) {
	sIdx := ag.Env.StateIndex(s)
	maxNext := 0.0
	if !done {
		_, maxNext = ag.MaxQ(next)
	}
	tdTarget = reward + ag.Gamma*maxNext
	tdError := tdTarget - ag.Q[sIdx][a]
	ag.Q[sIdx][a] += ag.Alpha * tdError
	return ag.Q[sIdx][a], tdTarget
}

// Train runs Q-learning for a given number of episodes and max steps per episode.
func (ag *QAgent) Train(episodes, maxSteps int, rng *rand.Rand) {
	if rng == nil {
		rng = rand.New(rand.NewSource(42))
	}
	for ep := 0; ep < episodes; ep++ {
		state := ag.Env.Start
		for step := 0; step < maxSteps; step++ {
			if ag.Env.IsTerminal(state) {
				break
			}
			action := ag.ChooseAction(state, rng)
			nextState, reward, done := ag.Env.Step(state, action)
			ag.BellmanUpdate(state, action, reward, nextState, done)
			state = nextState
			if done {
				break
			}
		}
	}
}
