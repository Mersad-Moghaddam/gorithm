//go:build js && wasm

package main

import (
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/astar"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/avl"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bellmanford"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bfs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/binarysearch"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bst"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bubble"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bucketsort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/coinchange"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/countingsort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/dfs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/dijkstra"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/fibonacci"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/floydwarshall"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/genetic"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/gradientdescent"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/hanoitower"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/heap"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/insertion"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kmeans"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kmp"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/knapsack"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kosaraju"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kruskal"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/lcs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/linearsearch"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/merge"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/minimax"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/nqueens"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/perceptron"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/prim"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/qlearning"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/quick"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/radixsort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/segmenttree"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/selection"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/sudoku"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/toposort"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/trie"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/unionfind"
	"syscall/js"
)

func main() {
	doc := js.Global().Get("document")
	player := NewPlayer(doc)
	bindLabControls(doc, player)
	initLibrary(doc, player)
	select {}
}

func getPresetInput(preset, id string) string {
	index := map[string]int{"worst": 0, "sorted": 1, "small": 2}[preset]
	values := map[string][3]string{
		"linearsearch": {`{"array":[90,80,70,60,50,40,30],"target":99}`, `{"array":[10,20,30,40,50,60,70],"target":10}`, `{"array":[42,17,88],"target":17}`},
		"binarysearch": {`{"array":[5,10,15,20,25,30,35,40],"target":99}`, `{"array":[10,20,30,40,50,60,70],"target":40}`, `{"array":[17,42,88],"target":42}`},
		"bst":          {`{"initial":[10,20,30,40,50,60],"ops":[{"type":"search","value":60}]}`, `{"initial":[40,20,60,10,30,50,70],"ops":[{"type":"search","value":40}]}`, `{"initial":[20,10,30],"ops":[{"type":"search","value":10}]}`},
		"hanoitower":   {"5", "3", "2"}, "fibonacci": {`{"n":12}`, `{"n":6}`, `{"n":3}`}, "nqueens": {`{"n":8}`, `{"n":4}`, `{"n":4}`},
		"coinchange": {`{"coins":[1,5,10,25],"amount":37}`, `{"coins":[1,2,5],"amount":6}`, `{"coins":[1,2],"amount":3}`},
	}
	if v, ok := values[id]; ok {
		return v[index]
	}
	return [3]string{"[95,82,70,58,44,31,18,5]", "[10,20,25,30,40,50,65,80]", "[42,17,88]"}[index]
}
