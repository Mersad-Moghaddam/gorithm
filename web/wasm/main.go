//go:build js && wasm

package main

import (
	"fmt"
	"strconv"
	"strings"
	"syscall/js"

	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/astar"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bfs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/binarysearch"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/bubble"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/dfs"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/dijkstra"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/heap"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/insertion"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kmp"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/kruskal"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/linearsearch"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/merge"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/prim"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/quick"
	_ "github.com/Mersad-Moghaddam/gorithm/internal/algorithms/selection"
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
		catalog.CategorySorting,
		catalog.CategorySearching,
		catalog.CategoryGraphs,
		catalog.CategoryTrees,
		catalog.CategoryRecursion,
		catalog.CategoryDP,
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
		sb.WriteString(fmt.Sprintf(`<div class="category-title">%s</div>`, string(cat)))
		for _, m := range items {
			sb.WriteString(fmt.Sprintf(
				`<button class="algo-btn" id="btn-algo-%s" data-id="%s">%s</button>`,
				m.ID, m.ID, m.Name,
			))
		}
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
