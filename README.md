# Gorithm

> **See Go algorithms think.**

Interactive educational algorithm visualizer written 100% in Go, running in the browser via WebAssembly with zero external runtime dependencies. Every visual frame maps 1:1 to an executable Go source line, animated synchronously via deterministic execution traces.

![Go Version](https://img.shields.io/badge/go-1.24%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Architecture](https://img.shields.io/badge/architecture-Single%20Binary%20Go%2FWASM-purple)

---

## Highlights

- **Deterministic Trace Engine**: Pure Go algorithms emit immutable execution steps containing variable snapshots, metrics, and line IDs.
- **Line-by-Line Code Synchronization**: The active source code line highlight advances synchronously with the SVG animation frame.
- **Pure SVG/DOM Visualizers**: Fast, accessible vector rendering with smooth CSS transitions orchestrated by the Go/WASM player.
- **Educational Explanations & In-Depth Docs**: Each algorithm features a dedicated "Learn" tab detailing how it works, time/space complexities, key invariants, and metrics (comparisons, swaps, operations).
- **Single Distributable Binary**: Native Go HTTP server embeds static HTML/CSS and compiled WebAssembly via `embed.FS`.

---

## 40 Essential Algorithms

| Category | Algorithm | Time Complexity | Space Complexity | Visual Renderer |
|---|---|---|---|---|
| **AI** | Minimax & Alpha-Beta Pruning | $O(b^{d/2})$ | $O(b \cdot d)$ | Tree |
| **AI** | K-Means 2D Clustering | $O(k \cdot n \cdot I)$ | $O(n + k)$ | Graph Clusters |
| **AI** | Linear Regression (Gradient Descent) | $O(\text{epochs} \cdot n)$ | $O(1)$ | Table |
| **AI** | Q-Learning (Reinforcement Learning) | $O(\text{episodes} \cdot \text{steps})$ | $O(\|S\| \cdot \|A\|)$ | Table |
| **AI** | Perceptron (Neural Unit) | $O(\text{epochs} \cdot n \cdot d)$ | $O(d)$ | Array Bars |
| **AI** | Genetic Algorithm | $O(\text{generations} \cdot P)$ | $O(P)$ | Table |
| **Sorting** | Bubble Sort | $O(n^2)$ | $O(1)$ | Array Bars |
| **Sorting** | Selection Sort | $O(n^2)$ | $O(1)$ | Array Bars |
| **Sorting** | Insertion Sort | $O(n^2)$ | $O(1)$ | Array Bars |
| **Sorting** | Merge Sort | $O(n \log n)$ | $O(n)$ | Array Bars |
| **Sorting** | Quick Sort | $O(n \log n)$ | $O(\log n)$ | Array Bars |
| **Sorting** | Heap Sort | $O(n \log n)$ | $O(1)$ | Array Bars |
| **Sorting** | Counting Sort | $O(n + k)$ | $O(k)$ | Array Bars |
| **Sorting** | Radix Sort (LSD) | $O(d \cdot (n + b))$ | $O(n + b)$ | Array Bars |
| **Sorting** | Bucket Sort | $O(n + k)$ | $O(n + k)$ | Array Bars |
| **Searching** | Linear Search | $O(n)$ | $O(1)$ | Array Elements |
| **Searching** | Binary Search | $O(\log n)$ | $O(1)$ | Array Elements |
| **Strings** | KMP String Matching | $O(n + m)$ | $O(m)$ | Table / Grid |
| **Graphs** | Breadth-First Search (BFS) | $O(V + E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Depth-First Search (DFS) | $O(V + E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Dijkstra's Shortest Path | $O((V + E) \log V)$ | $O(V)$ | Vector Graph |
| **Graphs** | A* Pathfinding | $O(E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Prim's MST | $O(E \log V)$ | $O(V)$ | Vector Graph |
| **Graphs** | Kruskal's MST | $O(E \log E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Bellman-Ford Shortest Path | $O(V \cdot E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Floyd-Warshall All-Pairs Path | $O(V^3)$ | $O(V^2)$ | Matrix Table |
| **Graphs** | Topological Sort (DAG) | $O(V + E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Kosaraju's SCC | $O(V + E)$ | $O(V)$ | Vector Graph |
| **Trees** | Binary Search Tree (BST) | $O(\log n)$ | $O(n)$ | Binary Tree |
| **Trees** | AVL Tree (Self-Balancing) | $O(\log n)$ | $O(n)$ | Balanced Tree |
| **Trees** | Trie (Prefix Tree) | $O(L)$ | $O(\Sigma \cdot L \cdot N)$ | Prefix Tree |
| **Trees** | Segment Tree (Range Queries) | $O(\log n)$ | $O(4n)$ | Interval Tree |
| **Trees** | Disjoint Set Union (Union-Find) | $O(\alpha(n))$ | $O(n)$ | Component Graph |
| **Backtracking** | N-Queens Puzzle | $O(N!)$ | $O(N)$ | Chessboard |
| **Backtracking** | Sudoku Solver | $O(9^{\text{cells}})$ | $O(1)$ | 9x9 Grid |
| **Recursion** | Tower of Hanoi | $O(2^n)$ | $O(n)$ | Pegs & Disks |
| **DP** | Fibonacci (Tabulation) | $O(n)$ | $O(n)$ | Decision Table |
| **DP** | 0/1 Knapsack | $O(n \cdot W)$ | $O(n \cdot W)$ | Decision Table |
| **DP** | Longest Common Subsequence | $O(m \cdot n)$ | $O(m \cdot n)$ | Decision Table |
| **DP** | Coin Change | $O(A \cdot \|C\|)$ | $O(A)$ | Decision Table |

---

## Architecture

```text
gorithm/
├── cmd/gorithm/main.go          # Single binary server entrypoint
├── internal/
│   ├── trace/                  # Deterministic Step, LineID, Metrics engine
│   ├── catalog/                # Lesson registry, Meta, and Documentation contracts
│   ├── server/                 # Embedded HTTP file server with security headers
│   └── algorithms/             # 20 isolated algorithm folders
│       ├── bubble/
│       ├── dijkstra/
│       ├── bst/
│       └── ...
├── web/
│   ├── wasm/                   # Go/WASM client: player, scrubber, keyboard controls
│   ├── renderers/              # Pure Go SVG renderers (array, graph, tree, table, hanoi)
│   ├── static/                 # HTML shell, dark editorial CSS, wasm_exec.js
│   └── embed.go                # go:embed static/*
├── scripts/build_wasm.sh       # WASM compilation script
├── Dockerfile                  # Multi-stage distroless build
├── fly.toml                    # Fly.io deployment config
└── Makefile
```

---

## Quickstart

### Local Development

Prerequisites: Go 1.24+

```bash
# 1. Clone repository
git clone https://github.com/Mersad-Moghaddam/gorithm.git
cd gorithm

# 2. Build WASM binary and launch server
make run
```

Open `http://localhost:8080` in your browser.

### Find an algorithm

Start in the Living Lab library. Search names, categories, summaries, and key
concepts; every search word must match. Combine search with a category filter,
then open a card to explore its visualization, source, and explanation.

Lesson links such as `/#algorithm=merge` can be shared directly. Returning to
the library preserves your search and category. Motion respects your system's
reduced-motion preference; playback always starts on your command.

Use **Theme** in the header to choose Living Lab, Midnight Blue, Warm Paper,
or Plum. Your choice is saved locally and restored on reload. Warm Paper keeps
the visualization canvas dark so existing algorithm state colors stay legible.

### Verification

```bash
go test ./...
make build
```

Optional browser regression check requires an external Playwright installation
and Chromium, with the app running. No Node dependency is needed by the app:

```bash
NODE_PATH=/path/to/playwright/node_modules node scripts/check_ui.mjs
```

Set `APP_URL` or `CHROMIUM` to override `http://localhost:8080` or
`/usr/bin/chromium`. The check covers discovery, history, pause cancellation,
dialog validation, tabs, mobile overflow, and runtime loading failure.

### Keyboard Shortcuts

- `/`: Focus library search
- `R`: Restart current lesson
- Arrow keys, Home, End: Move between Code & state and Learn tabs when focused
- `Escape`: Close custom dataset dialog
- `Space`: Play / Pause animation
- `ArrowRight`: Step forward
- `ArrowLeft`: Step backward
- Speed Selector: `0.25x` to `2.0x`
- Scrubber: Seek to any step in execution history

### Docker

```bash
docker build -t gorithm:latest .
docker run -p 8080:8080 gorithm:latest
```

### Deploy to Fly.io

```bash
fly launch
fly deploy
```

---

## License

MIT © [Mersad Moghaddam](https://github.com/Mersad-Moghaddam)
