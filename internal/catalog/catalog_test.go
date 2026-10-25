package catalog

import (
	"encoding/json"
	"testing"

	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type dummyLesson struct{}

func (d dummyLesson) Meta() Meta {
	return Meta{
		ID:       "dummy",
		Name:     "Dummy Algo",
		Category: CategorySorting,
		Renderer: "array",
		Documentation: Documentation{
			Summary:         "Simple test algo",
			TimeComplexity:  "O(1)",
			SpaceComplexity: "O(1)",
			KeyConcepts:     []string{"Testing"},
			HowItWorks:      "Does nothing, only for unit testing.",
		},
		DefaultInput: "[1,2,3]",
	}
}

func (d dummyLesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func dummy() {}"},
	}
}

func (d dummyLesson) Build(input json.RawMessage) ([]trace.Step, error) {
	b := trace.NewBuilder()
	b.AddStep("L1", trace.ActionDone, "Finished", []int{1, 2, 3}, nil)
	return b.Steps(), nil
}

func TestCatalogRegisterAndGet(t *testing.T) {
	Register(dummyLesson{})
	l, ok := Get("dummy")
	if !ok {
		t.Fatal("expected dummy lesson in registry")
	}
	if l.Meta().Name != "Dummy Algo" {
		t.Errorf("expected 'Dummy Algo', got %s", l.Meta().Name)
	}

	steps, err := l.Build(nil)
	if err != nil {
		t.Fatalf("build error: %v", err)
	}
	if err := ValidateSourceLineIDs(l, steps); err != nil {
		t.Fatalf("validation failed: %v", err)
	}
}
