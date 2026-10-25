# Gorithm

> **See Go algorithms think.**

Interactive educational algorithm visualizer written 100% in Go, running in the browser via WebAssembly with zero external runtime dependencies. Every animation frame maps 1:1 to executable Go source line execution.

## Features
- **Deterministic Trace Engine**: Pure Go algorithms emit immutable execution steps.
- **Line-by-Line Synchronization**: Source code line highlight advances synchronously with visual state changes.
- **20 Essential Algorithms**: Sorting, Searching, Graphs, Trees, Recursion, and Dynamic Programming.
- **Pure SVG/DOM Animation**: Smooth CSS transitions orchestrated by Go/WASM player.
- **Single Distributable Binary**: Server embeds frontend static assets and compiled WebAssembly binary.

## Quickstart

```bash
# Build WASM binary
./scripts/build_wasm.sh

# Run local server
go run cmd/gorithm/main.go
```

Open `http://localhost:8080` in your browser.
