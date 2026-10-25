//go:build js && wasm

package main

import (
	"encoding/json"
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

	var activeCat catalog.Category
	if player.lesson != nil {
		activeCat = player.lesson.Meta().Category
	} else if len(metas) > 0 {
		activeCat = metas[0].Category
	}

	var sb strings.Builder
	for _, cat := range catOrder {
		items := categories[cat]
		if len(items) == 0 {
			continue
		}
		itemsDisplay := "none"
		if cat == activeCat {
			itemsDisplay = "flex"
		}
		sb.WriteString(fmt.Sprintf(
			`<div class="category-group" data-category="%s">
				<div class="category-header" style="cursor: pointer;">
					<span class="category-title">%s</span>
					<span class="category-count">%d</span>
				</div>
				<div class="category-items" style="display: %s;">`,
			string(cat), string(cat), len(items), itemsDisplay,
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

	// Attach click listeners to category headers for accordion toggling
	headers := container.Call("querySelectorAll", ".category-header")
	for i := 0; i < headers.Length(); i++ {
		headers.Index(i).Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			parent := this.Get("parentElement")
			if !parent.IsNull() && !parent.IsUndefined() {
				itemsEl := parent.Call("querySelector", ".category-items")
				if !itemsEl.IsNull() && !itemsEl.IsUndefined() {
					disp := itemsEl.Get("style").Get("display").String()
					if disp == "none" {
						itemsEl.Get("style").Set("display", "flex")
					} else {
						itemsEl.Get("style").Set("display", "none")
					}
				}
			}
			return nil
		}))
	}

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
					backdrop := doc.Call("getElementById", "drawer-backdrop")
					if !backdrop.IsNull() && !backdrop.IsUndefined() {
						backdrop.Get("style").Set("display", "none")
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
		parentGroup := activeBtn.Call("closest", ".category-group")
		if !parentGroup.IsNull() && !parentGroup.IsUndefined() {
			items := parentGroup.Call("querySelector", ".category-items")
			if !items.IsNull() && !items.IsUndefined() {
				items.Get("style").Set("display", "flex")
			}
		}
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

	// Scrubber Input & Contention Fix
	scrubber := doc.Call("getElementById", "scrubber")
	if !scrubber.IsNull() && !scrubber.IsUndefined() {
		scrubber.Call("addEventListener", "mousedown", js.FuncOf(func(this js.Value, args []js.Value) any {
			player.Pause()
			return nil
		}))
		scrubber.Call("addEventListener", "touchstart", js.FuncOf(func(this js.Value, args []js.Value) any {
			player.Pause()
			return nil
		}))
		scrubber.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
			val := this.Get("value").String()
			if idx, err := strconv.Atoi(val); err == nil {
				player.Seek(idx)
			}
			return nil
		}))
	}

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

	// Search filter with category visibility update
	searchInput := doc.Call("getElementById", "algo-search")
	if !searchInput.IsNull() && !searchInput.IsUndefined() {
		searchInput.Call("addEventListener", "input", js.FuncOf(func(this js.Value, args []js.Value) any {
			query := strings.ToLower(strings.TrimSpace(this.Get("value").String()))
			groups := doc.Call("querySelectorAll", ".category-group")
			for g := 0; g < groups.Length(); g++ {
				grp := groups.Index(g)
				btns := grp.Call("querySelectorAll", ".algo-btn")
				visibleCount := 0
				for b := 0; b < btns.Length(); b++ {
					btn := btns.Index(b)
					nameEl := btn.Call("querySelector", ".algo-name")
					name := ""
					if !nameEl.IsNull() && !nameEl.IsUndefined() {
						name = strings.ToLower(nameEl.Get("innerText").String())
					}
					id := strings.ToLower(btn.Get("dataset").Get("id").String())
					if query == "" || strings.Contains(name, query) || strings.Contains(id, query) {
						btn.Get("style").Set("display", "flex")
						visibleCount++
					} else {
						btn.Get("style").Set("display", "none")
					}
				}
				if visibleCount > 0 {
					grp.Get("style").Set("display", "block")
					if query != "" {
						itemsEl := grp.Call("querySelector", ".category-items")
						if !itemsEl.IsNull() && !itemsEl.IsUndefined() {
							itemsEl.Get("style").Set("display", "flex")
						}
					}
				} else {
					grp.Get("style").Set("display", "none")
				}
			}
			return nil
		}))
	}

	// Mobile Menu Button & Drawer Backdrop
	sidebar := doc.Call("getElementById", "sidebar")
	backdrop := doc.Call("getElementById", "drawer-backdrop")
	menuBtn := doc.Call("getElementById", "mobile-menu-btn")

	closeSidebar := func() {
		if !sidebar.IsNull() && !sidebar.IsUndefined() {
			sidebar.Get("classList").Call("remove", "open")
		}
		if !backdrop.IsNull() && !backdrop.IsUndefined() {
			backdrop.Get("style").Set("display", "none")
		}
	}

	if !backdrop.IsNull() && !backdrop.IsUndefined() {
		backdrop.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			closeSidebar()
			return nil
		}))
	}

	if !menuBtn.IsNull() && !menuBtn.IsUndefined() {
		menuBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			if !sidebar.IsNull() && !sidebar.IsUndefined() {
				isOpen := sidebar.Get("classList").Call("contains", "open").Bool()
				if isOpen {
					closeSidebar()
				} else {
					sidebar.Get("classList").Call("add", "open")
					if !backdrop.IsNull() && !backdrop.IsUndefined() {
						backdrop.Get("style").Set("display", "block")
					}
				}
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
	errBox := doc.Call("getElementById", "modal-error")

	closeModal := func() {
		if !modal.IsNull() && !modal.IsUndefined() {
			modal.Get("style").Set("display", "none")
		}
		if !errBox.IsNull() && !errBox.IsUndefined() {
			errBox.Get("style").Set("display", "none")
		}
	}

	if !customBtn.IsNull() && !customBtn.IsUndefined() {
		customBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			if player.lesson != nil {
				inputField.Set("value", player.lesson.Meta().DefaultInput)
			}
			if !errBox.IsNull() && !errBox.IsUndefined() {
				errBox.Get("style").Set("display", "none")
			}
			modal.Get("style").Set("display", "flex")
			inputField.Call("focus")
			return nil
		}))
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
	if !modal.IsNull() && !modal.IsUndefined() {
		modal.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) > 0 && args[0].Get("target").Equal(modal) {
				closeModal()
			}
			return nil
		}))
	}

	applyCustomInput := func() {
		val := strings.TrimSpace(inputField.Get("value").String())
		if val != "" && player.lesson != nil {
			var dummy any
			if err := json.Unmarshal([]byte(val), &dummy); err != nil {
				if !errBox.IsNull() && !errBox.IsUndefined() {
					errBox.Set("innerText", fmt.Sprintf("Invalid JSON: %v", err))
					errBox.Get("style").Set("display", "block")
				}
				return
			}
			closeModal()
			player.LoadLessonWithInput(player.lesson, []byte(val))
		} else {
			closeModal()
		}
	}

	if !applyBtn.IsNull() && !applyBtn.IsUndefined() {
		applyBtn.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			applyCustomInput()
			return nil
		}))
	}

	if !inputField.IsNull() && !inputField.IsUndefined() {
		inputField.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
			if len(args) > 0 {
				key := args[0].Get("key").String()
				if key == "Enter" {
					args[0].Call("preventDefault")
					applyCustomInput()
				} else if key == "Escape" {
					args[0].Call("preventDefault")
					closeModal()
				}
			}
			return nil
		}))
	}

	// Preset Chips
	presetChips := doc.Call("querySelectorAll", ".preset-chip")
	for i := 0; i < presetChips.Length(); i++ {
		chip := presetChips.Index(i)
		chip.Call("addEventListener", "click", js.FuncOf(func(this js.Value, args []js.Value) any {
			preset := this.Get("dataset").Get("preset").String()
			algoID := ""
			if player.lesson != nil {
				algoID = player.lesson.Meta().ID
			}
			val := getPresetInput(preset, algoID)
			if val != "" && !inputField.IsNull() && !inputField.IsUndefined() {
				inputField.Set("value", val)
			}
			return nil
		}))
	}

	// Keyboard controls
	doc.Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			key := args[0].Get("code").String()
			// Don't trigger shortcuts if typing inside an input
			activeTag := doc.Get("activeElement").Get("tagName").String()
			if activeTag == "INPUT" || activeTag == "TEXTAREA" || activeTag == "SELECT" {
				return nil
			}
			switch key {
			case "Escape":
				closeModal()
			case "Space":
				args[0].Call("preventDefault")
				player.TogglePlay()
			case "ArrowRight":
				args[0].Call("preventDefault")
				player.Next()
			case "ArrowLeft":
				args[0].Call("preventDefault")
				player.Prev()
			case "KeyR":
				args[0].Call("preventDefault")
				player.Reset()
			}
		}
		return nil
	}))
}

func getPresetInput(preset string, id string) string {
	switch preset {
	case "worst":
		switch id {
		case "linearsearch":
			return `{"array": [90, 80, 70, 60, 50, 40, 30], "target": 99}`
		case "binarysearch":
			return `{"array": [5, 10, 15, 20, 25, 30, 35, 40], "target": 99}`
		case "bst":
			return `{"initial": [10, 20, 30, 40, 50, 60], "ops": [{"type": "search", "value": 60}]}`
		case "hanoitower":
			return "5"
		case "fibonacci":
			return `{"n": 12}`
		case "nqueens":
			return `{"n": 8}`
		case "coinchange":
			return `{"coins": [1, 5, 10, 25], "amount": 37}`
		default:
			return "[95, 82, 70, 58, 44, 31, 18, 5]"
		}
	case "sorted":
		switch id {
		case "linearsearch":
			return `{"array": [10, 20, 30, 40, 50, 60, 70], "target": 10}`
		case "binarysearch":
			return `{"array": [10, 20, 30, 40, 50, 60, 70], "target": 40}`
		case "bst":
			return `{"initial": [40, 20, 60, 10, 30, 50, 70], "ops": [{"type": "search", "value": 40}]}`
		case "hanoitower":
			return "3"
		case "fibonacci":
			return `{"n": 6}`
		case "nqueens":
			return `{"n": 4}`
		case "coinchange":
			return `{"coins": [1, 2, 5], "amount": 6}`
		default:
			return "[10, 20, 25, 30, 40, 50, 65, 80]"
		}
	case "small":
		switch id {
		case "linearsearch":
			return `{"array": [42, 17, 88], "target": 17}`
		case "binarysearch":
			return `{"array": [17, 42, 88], "target": 42}`
		case "bst":
			return `{"initial": [20, 10, 30], "ops": [{"type": "search", "value": 10}]}`
		case "hanoitower":
			return "2"
		case "fibonacci":
			return `{"n": 3}`
		case "nqueens":
			return `{"n": 4}`
		case "coinchange":
			return `{"coins": [1, 2], "amount": 3}`
		default:
			return "[42, 17, 88]"
		}
	default:
		return ""
	}
}
