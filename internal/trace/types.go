package trace

type ActionType string

const (
	ActionInit      ActionType = "init"
	ActionCompare   ActionType = "compare"
	ActionSwap      ActionType = "swap"
	ActionSet       ActionType = "set"
	ActionVisit     ActionType = "visit"
	ActionHighlight ActionType = "highlight"
	ActionPush      ActionType = "push"
	ActionPop       ActionType = "pop"
	ActionDone      ActionType = "done"
)

type Metrics struct {
	Comparisons int `json:"comparisons"`
	Swaps       int `json:"swaps"`
	Operations  int `json:"operations"`
}

type Step struct {
	HoldMS      int            `json:"holdMs,omitempty"`
	Index       int            `json:"index"`
	LineID      string         `json:"lineId"`
	Action      ActionType     `json:"action"`
	Explanation string         `json:"explanation"`
	State       any            `json:"state"`
	Variables   map[string]any `json:"variables"`
	Metrics     Metrics        `json:"metrics"`
}

type SourceLine struct {
	ID     string `json:"id"`
	Number int    `json:"number"`
	Code   string `json:"code"`
}
