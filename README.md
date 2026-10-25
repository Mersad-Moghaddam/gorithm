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

## 20 Essential Algorithms

| Category | Algorithm | Time Complexity | Space Complexity | Visual Renderer |
|---|---|---|---|---|
| **Sorting** | Bubble Sort | $O(n^2)$ | $O(1)$ | Array Bars |
| **Sorting** | Selection Sort | $O(n^2)$ | $O(1)$ | Array Bars |
| **Sorting** | Insertion Sort | $O(n^2)$ | $O(1)$ | Array Bars |
| **Sorting** | Merge Sort | $O(n \log n)$ | $O(n)$ | Array Bars |
| **Sorting** | Quick Sort | $O(n \log n)$ | $O(\log n)$ | Array Bars |
| **Sorting** | Heap Sort | $O(n \log n)$ | $O(1)$ | Array Bars |
| **Searching** | Linear Search | $O(n)$ | $O(1)$ | Array Elements |
| **Searching** | Binary Search | $O(\log n)$ | $O(1)$ | Array Elements |
| **Strings** | KMP String Matching | $O(n + m)$ | $O(m)$ | Table / Grid |
| **Graphs** | Breadth-First Search (BFS) | $O(V + E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Depth-First Search (DFS) | $O(V + E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Dijkstra's Shortest Path | $O((V + E) \log V)$ | $O(V)$ | Vector Graph |
| **Graphs** | A* Pathfinding | $O(E)$ | $O(V)$ | Vector Graph |
| **Graphs** | Prim's MST | $O(E \log V)$ | $O(V)$ | Vector Graph |
| **Graphs** | Kruskal's MST | $O(E \log E)$ | $O(V)$ | Vector Graph |
| **Trees** | Binary Search Tree (BST) | $O(\log n)$ | $O(n)$ | Binary Tree |
| **Recursion** | Tower of Hanoi | $O(2^n)$ | $O(n)$ | Pegs & Disks |
| **DP** | Fibonacci (Tabulation) | $O(n)$ | $O(n)$ | Decision Table |
| **DP** | 0/1 Knapsack | $O(n \cdot W)$ | $O(n \cdot W)$ | Decision Table |
| **DP** | Longest Common Subsequence | $O(m \cdot n)$ | $O(m \cdot n)$ | Decision Table |

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

### Keyboard Shortcuts

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
