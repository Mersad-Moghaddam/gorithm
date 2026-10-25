//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

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
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/heap"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/hanoitower"
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
	"github.com/Mersad-Moghaddam/gorithm/internal/catalog"
)

func main() {
	doc := js.Global().Get("document")
	player := NewPlayer(doc)

	// Render sidebar category list
	renderSidebar(doc, player)

	// Load first registered algorithm
	metas := catalog.List()
	if len(metas) > 0 {
		if firstLesson, ok := catalog.Get(metas[0].ID); ok {
			player.LoadLesson(firstLesson)
			markActiveButton(doc, metas[0].ID)
		}
	}

	// Bind controls
	bindControls(doc, player)

	// Keep WASM alive
	select {}
}

func renderSidebar(doc js.Value, player *Player) {
	metas := catalog.List()
	categories := make(map[catalog.Category][]catalog.Meta)
	catOrder := []catalog.Category{
		catalog.CategoryAI,
		catalog.CategorySorting,
		catalog.CategorySearching,
		catalog.CategoryGraphs,
		catalog.CategoryTrees,
		catalog.CategoryBacktracking,
		catalog.CategoryDP,
		catalog.CategoryRecursion,
	}

	for _, m := range metas {
		categories[m.Category] = append(categories[m.Category], m)
	}

	var sb strings.Builder
	for _, cat := range catOrder {
		items := categories[cat]
		if len(items) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="category-group">
				<div class="category-header">
					<span class="category-title">%s</span>
					<span class="category-count">%d</span>
				</div>
				<div class="category-items">`,
			string(cat), len(items),
		))
		for _, m := range items {
			sb.WriteString(fmt.Sprintf(
				`<button class="algo-btn" id="btn-algo-%s" data-id="%s">
					<span class="algo-name">%s</span>
					<span class="algo-complexity-tag">%s</span>
				</button>`,
				m.ID, m.ID, m.Name, m.Documentation.TimeComplexity,
			))
		}
		sb.WriteString(`</div></div>`)
	}

	container := doc.Call("getElementById", "algo-categories")
	container.Set("innerHTML", sb.String())

	// Attach click listeners to algorithm buttons
	for _, m := range metas {
		id := m.ID
		btn := doc.Call("getElementById", "btn-algo-"+id)
		if !btn.IsNull() && !btn.IsUndefined() {
			btn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
				if l, ok := catalog.Get(id); ok {
					player.LoadLesson(l)
					markActiveButton(doc, id)
					sidebar := doc.Call("getElementById", "sidebar")
					if !sidebar.IsNull() && !sidebar.IsUndefined() {
						sidebar.Get("classList").Call("remove", "open")
					}
				}
				return nil
			}))
		}
	}
}

func markActiveButton(doc js.Value, activeID string) {
	btns := doc.Call("querySelectorAll", ".algo-btn")
	for i := 0; i < btns.Length(); i++ {
		btns.Index(i).Get("classList").Call("remove", "active")
	}
	activeBtn := doc.Call("getElementById", "btn-algo-"+activeID)
	if !activeBtn.IsNull() && !activeBtn.IsUndefined() {
		activeBtn.Get("classList").Call("add", "active")
	}
}

func bindControls(doc js.Value, player *Player) {
	// Play / Pause
	doc.Call("getElementById", "btn-play").Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		player.TogglePlay()
		return nil
	}))

	// Next
	doc.Call("getElementById", "btn-next").Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		player.Next()
		return nil
	}))

	// Prev
	doc.Call("getElementById", "btn-prev").Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		player.Prev()
		return nil
	}))

	// Reset
	doc.Call("getElementById", "btn-reset").Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		player.Reset()
		return nil
	}))

	// Speed Select
	doc.Call("getElementById", "speed-select").Call("addEventListener", "change", js.FuncOf(func(this js.Value, args []js.Value) any {
		val := this.Get("value").String()
		if sp, err := strconv.ParseFloat(val, 64); err == nil {
			player.SetSpeed(sp)
		}
		return nil
	}))

	// Scrubber Input
	doc.Call("getElementById", "scrubber").Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
		val := this.Get("value").String()
		if idx, err := strconv.Atoi(val); err == nil {
			player.Seek(idx)
		}
		return nil
	}))

	// Tabs: Code vs Learn
	codeTabBtn := doc.Call("getElementById", "tab-code-btn")
	docTabBtn := doc.Call("getElementById", "tab-doc-btn")
	codeTab := doc.Call("getElementById", "tab-code")
	docTab := doc.Call("getElementById", "tab-doc")

	codeTabBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		codeTabBtn.Get("classList").Call("add", "active")
		docTabBtn.Get("classList").Call("remove", "active")
		codeTab.Get("style").Set("display", "block")
		docTab.Get("style").Set("display", "none")
		return nil
	}))

	docTabBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
		docTabBtn.Get("classList").Call("add", "active")
		codeTabBtn.Get("classList").Call("remove", "active")
		docTab.Get("style").Set("display", "block")
		codeTab.Get("style").Set("display", "none")
		return nil
	}))

	// Search filter
	searchInput := doc.Call("getElementById", "algo-search")
	if !searchInput.IsNull() && !searchInput.IsUndefined() {
		searchInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
			query := strings.ToLower(strings.TrimSpace(this.Get("value").String()))
			btns := doc.Call("querySelectorAll", ".algo-btn")
			for i := 0; i < btns.Length(); i++ {
				btn := btns.Index(i)
				nameEl := btn.Call("querySelector", ".algo-name")
				name := ""
				if !nameEl.IsNull() && !nameEl.IsUndefined() {
					name = strings.ToLower(nameEl.Get("innerText").String())
				}
				id := strings.ToLower(btn.Get("dataset").Get("id").String())
				if query == "" || strings.Contains(name, query) || strings.Contains(id, query) {
					btn.Get("style").Set("display", "flex")
				} else {
					btn.Get("style").Set("display", "none")
				}
			}
			return nil
		}))
	}

	// Mobile Menu Button
	menuBtn := doc.Call("getElementById", "mobile-menu-btn")
	if !menuBtn.IsNull() && !menuBtn.IsUndefined() {
		menuBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			sidebar := doc.Call("getElementById", "sidebar")
			if !sidebar.IsNull() && !sidebar.IsUndefined() {
				sidebar.Get("classList").Call("toggle", "open")
			}
			return nil
		}))
	}

	// User Interaction: Shuffle
	shuffleBtn := doc.Call("getElementById", "btn-shuffle")
	if !shuffleBtn.IsNull() && !shuffleBtn.IsUndefined() {
		shuffleBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			player.Shuffle()
			return nil
		}))
	}

	// User Interaction: Custom Input Modal
	modal := doc.Call("getElementById", "custom-modal")
	customBtn := doc.Call("getElementById", "btn-custom-input")
	closeBtn := doc.Call("getElementById", "btn-modal-close")
	cancelBtn := doc.Call("getElementById", "btn-modal-cancel")
	applyBtn := doc.Call("getElementById", "btn-modal-apply")
	inputField := doc.Call("getElementById", "custom-input-val")

	if !customBtn.IsNull() && !customBtn.IsUndefined() {
		customBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			modal.Get("style").Set("display", "flex")
			return nil
		}))
	}
	closeModal := func() {
		if !modal.IsNull() && !modal.IsUndefined() {
			modal.Get("style").Set("display", "none")
		}
	}
	if !closeBtn.IsNull() && !closeBtn.IsUndefined() {
		closeBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			closeModal()
			return nil
		}))
	}
	if !cancelBtn.IsNull() && !cancelBtn.IsUndefined() {
		cancelBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			closeModal()
			return nil
		}))
	}
	if !applyBtn.IsNull() && !applyBtn.IsUndefined() {
		applyBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			val := inputField.Get("value").String()
			closeModal()
			if strings.TrimSpace(val) != "" && player.lesson != nil {
				player.LoadLessonWithInput(player.lesson, []byte(val))
			}
			return nil
		}))
	}

	// Keyboard controls
	doc.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			key := args[0].Get("code").String()
			switch key {
			case "Space":
				args[0].Call("preventDefault")
				player.TogglePlay()
			case "ArrowRight":
				args[0].Call("preventDefault")
				player.Next()
			case "ArrowLeft":
				args[0].Call("preventDefault")
				player.Prev()
			}
		}
		return nil
	}))
}
