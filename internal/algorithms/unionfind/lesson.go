package unionfind

import (
	"encoding/json"
	"bytes"
	"fmt"

	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
	"github.com/Mersad-Moghaddam/gorithm/internal/trace"
)

type Lesson struct{}

type LessonInput struct {
	N          int         `json:"n"`
	Operations []Operation `json:"operations"`
}

func (l Lesson) Meta() catalog.Meta {
	return catalog.Meta{
		ID:       "unionfind",
		Name:     "Disjoint Set Union (Union-Find)",
		Category: catalog.CategoryTrees,
		Renderer: "graph",
		Documentation: catalog.Documentation{
			Summary:         "Maintains a partition of elements into disjoint sets, supporting near-constant time union and find operations using path compression and union by rank.",
			TimeComplexity:  "O(α(n)) amortized inverse Ackermann",
			SpaceComplexity: "O(n)",
			Difficulty:      "Medium",
			KeyConcepts:     []string{"Equivalence Relations", "Representative Roots", "Path Compression", "Union by Rank"},
			HowItWorks:      "Each set is structured as a tree where every element points to its parent and the root represents the entire set. Find traverses parent links to locate the root representative, simultaneously repointing traversed nodes directly to the root (path compression). Union joins two sets by linking the root of the shallower tree under the deeper tree (union by rank), keeping overall tree depth strictly bounded.",
		},
		DefaultInput: `{"n": 5, "operations": [{"op": "union", "i": 0, "j": 1}, {"op": "union", "i": 1, "j": 2}, {"op": "union", "i": 3, "j": 4}, {"op": "union", "i": 0, "j": 4}, {"op": "find", "i": 2}, {"op": "find", "i": 4}, {"op": "connected", "i": 0, "j": 3}]}`,
	}
}

func (l Lesson) Source() []trace.SourceLine {
	return []trace.SourceLine{
		{ID: "L1", Number: 1, Code: "func New(n int) *DSU {"},
		{ID: "L2", Number: 2, Code: "    parent, rank = make([]int, n), make([]int, n)"},
		{ID: "L3", Number: 3, Code: "    for i := range parent { parent[i] = i }"},
		{ID: "L4", Number: 4, Code: "}"},
		{ID: "L5", Number: 5, Code: "func Find(i int) int {"},
		{ID: "L6", Number: 6, Code: "    if parent[i] != i {"},
		{ID: "L7", Number: 7, Code: "        parent[i] = Find(parent[i]) // Path compression"},
		{ID: "L8", Number: 8, Code: "    }"},
		{ID: "L9", Number: 9, Code: "    return parent[i]"},
		{ID: "L10", Number: 10, Code: "}"},
		{ID: "L11", Number: 11, Code: "func Union(i, j int) bool {"},
		{ID: "L12", Number: 12, Code: "    rootI, rootJ := Find(i), Find(j)"},
		{ID: "L13", Number: 13, Code: "    if rootI == rootJ { return false }"},
		{ID: "L14", Number: 14, Code: "    if rank[rootI] < rank[rootJ] { parent[rootI] = rootJ }"},
		{ID: "L15", Number: 15, Code: "    else if rank[rootI] > rank[rootJ] { parent[rootJ] = rootI }"},
		{ID: "L16", Number: 16, Code: "    else { parent[rootJ] = rootI; rank[rootI]++ }"},
		{ID: "L17", Number: 17, Code: "    return true"},
		{ID: "L18", Number: 18, Code: "}"},
		{ID: "L19", Number: 19, Code: "func Connected(i, j int) bool {"},
		{ID: "L20", Number: 20, Code: "    return Find(i) == Find(j)"},
		{ID: "L21", Number: 21, Code: "}"},
	}
}

func (l Lesson) Build(input json.RawMessage) ([]trace.Step, error) {
	n := 5
	ops := []Operation{
		{Op: "union", I: 0, J: 1},
		{Op: "union", I: 1, J: 2},
		{Op: "union", I: 3, J: 4},
		{Op: "union", I: 0, J: 4},
		{Op: "find", I: 2},
		{Op: "find", I: 4},
		{Op: "connected", I: 0, J: 3},
	}

	if len(input) > 0 {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(input, &fields); err != nil { return nil,fmt.Errorf("union-find input must be an object: %w",err) }
		if fields == nil { return nil,fmt.Errorf("union-find input must not be null") }
		if raw,ok := fields["n"]; ok {
			if bytes.Equal(bytes.TrimSpace(raw),[]byte("null")) { return nil,fmt.Errorf("n must be an integer in [1,64]") }
			if err := json.Unmarshal(raw,&n); err != nil { return nil,fmt.Errorf("invalid n: %w",err) }
		}
		if n < 1 || n > 64 { return nil,fmt.Errorf("n must be in [1,64]") }
		if raw,ok := fields["operations"]; ok {
			var entries []struct { Op *string `json:"op"`; I *int `json:"i"`; J *int `json:"j"` }
			if bytes.Equal(bytes.TrimSpace(raw),[]byte("null")) { return nil,fmt.Errorf("operations must be an array") }
			if err := json.Unmarshal(raw,&entries); err != nil { return nil,fmt.Errorf("invalid operations: %w",err) }
			if len(entries)>128 { return nil,fmt.Errorf("operations must contain at most 128 entries") }
			ops = make([]Operation, len(entries))
			for i, e := range entries {
				if e.Op == nil || e.I == nil { return nil, fmt.Errorf("operation %d requires op and i", i) }
				if *e.Op != "find" && *e.Op != "union" && *e.Op != "connected" { return nil, fmt.Errorf("unknown operation %q", *e.Op) }
				if *e.Op != "find" && e.J == nil { return nil, fmt.Errorf("operation %d requires j", i) }
				ops[i] = Operation{Op: *e.Op, I: *e.I}
				if e.J != nil {
					ops[i].J = *e.J
					if *e.J < 0 || *e.J >= n { return nil, fmt.Errorf("operation %d: j outside [0,n)", i) }
				}
			}
		}
	}
	for i,op := range ops { if op.I < 0 || op.I >= n || (op.Op != "find" && (op.J < 0 || op.J >= n)) { return nil,fmt.Errorf("operation %d: index outside [0,n)",i) } }

	return Trace(n, ops), nil
}

func init() {
	catalog.Register(Lesson{})
}
