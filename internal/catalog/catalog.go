package catalog

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Category string

const (
	CategorySorting   Category = "Sorting"
	CategorySearching Category = "Search & Strings"
	CategoryGraphs    Category = "Graphs"
	CategoryTrees     Category = "Trees"
	CategoryRecursion Category = "Recursion"
	CategoryDP        Category = "Dynamic Programming"
)

type Documentation struct {
	Summary         string   `json:"summary"`
	TimeComplexity  string   `json:"timeComplexity"`
	SpaceComplexity string   `json:"spaceComplexity"`
	KeyConcepts     []string `json:"keyConcepts"`
	HowItWorks      string   `json:"howItWorks"`
}

type Meta struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Category      Category      `json:"category"`
	Renderer      string        `json:"renderer"`
	Documentation Documentation `json:"documentation"`
	DefaultInput  string        `json:"defaultInput"`
}

type Lesson interface {
	Meta() Meta
	Source() []trace.SourceLine
	Build(input json.RawMessage) ([]trace.Step, error)
}

var (
	registry = make(map[string]Lesson)
	order    []string
	mu       sync.RWMutex
)

func Register(l Lesson) {
	mu.Lock()
	defer mu.Unlock()
	id := l.Meta().ID
	if _, exists := registry[id]; !exists {
		order = append(order, id)
	}
	registry[id] = l
}

func Get(id string) (Lesson, bool) {
	mu.RLock()
	defer mu.RUnlock()
	l, ok := registry[id]
	return l, ok
}

func List() []Meta {
	mu.RLock()
	defer mu.RUnlock()
	metas := make([]Meta, 0, len(order))
	for _, id := range order {
		metas = append(metas, registry[id].Meta())
	}
	return metas
}

func ValidateSourceLineIDs(l Lesson, steps []trace.Step) error {
	lines := l.Source()
	lineIDs := make(map[string]bool, len(lines))
	for _, line := range lines {
		lineIDs[line.ID] = true
	}
	for i, step := range steps {
		if !lineIDs[step.LineID] {
			return fmt.Errorf("step %d references unknown LineID: %q", i, step.LineID)
		}
	}
	return nil
}
