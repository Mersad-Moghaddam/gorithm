# Teaching Animations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make all 40 algorithms teach their decisions through understandable visual changes.

**Architecture:** Existing Go traces remain the execution source of truth. Shared SVG renderers and playback preserve visual identity; algorithm-specific state exposes worklists, dependencies, and models. One dedicated subagent owns each algorithm, with shared integration handled centrally.

**Tech Stack:** Go, Go/WASM, SVG, HTML, CSS; no new runtime dependencies.

**Spec:** `docs/superpowers/specs/2026-10-27-teaching-animations-design.md`

## Global Constraints

- Retain Go/WASM, deterministic trace snapshots, existing controls, source synchronization, and SVG; add no runtime dependencies.
- Review and improve all 40 registered algorithms, with one dedicated subagent per algorithm as requested.
- Trace/source correctness, custom-input validation, and immutable historical snapshots must survive changes.
- Preserve existing user work; stage only files owned by this task.
- Reduced motion, labels beyond color, keyboard access, and exact seek/reverse snapshots are required.

## Review Focus

1. Duplicate values must not exchange visual identity accidentally; test stable IDs during swaps.
2. Pause, seek, restart, and lesson changes during motion must not allow stale callbacks to advance state.
3. Empty/singleton, negative, repeated, and large accepted inputs must not produce invalid geometry or false results.
4. Graph disconnection, unreachable targets, and nonconvergent training must end with truthful explanations.
5. Custom text must be escaped in SVG/HTML; mobile and reduced-motion users must retain readable state.

## Task 1: Shared teaching and motion contracts

**Files:** `internal/trace/types.go`, `web/renderers/array.go`, `web/renderers/graph.go`, `web/renderers/tree.go`, `web/renderers/table.go`, `web/renderers/hanoi.go`, `web/renderers/renderers_test.go`, `web/wasm/player.go`, `web/static/styles.css`, `web/static/index.html`.

**Interfaces:** Preserve existing `RenderArraySVG`, `RenderGraphSVG`, `RenderTreeSVG`, `RenderTableSVG`, and `RenderHanoiSVG` signatures `(step trace.Step, width, height int) string`. Add `RenderStepSVG(step trace.Step, renderer string, width, height int) string` in `web/renderers/scene.go` as central dispatch. Specialized state implements `SceneState` with `RenderSVG(width, height int) string`; dispatch checks this first, then existing renderer names. Native Go tests can render every lesson without browser imports. Shared SVG objects use `data-key` for stable identity; algorithm agents supply stable identities where values can repeat.

- [ ] Inspect existing trace builders and snapshot copying; finalize smallest additive state fields before dispatching implementers.
- [ ] Add failing tests for dispatch fallback, specialized scenes, escaped labels, and distinct identities for equal values.
- [ ] Run `go test ./web/renderers ./internal/trace` and confirm intended failures.
- [ ] Implement dispatch and shared labeled visuals; retain compatible existing state fields. Use native DOM reconciliation by `data-key` for continuous playback; exact replacement for nonadjacent seek is acceptable.
- [ ] Fix playback lifecycle with cancellation/generation ownership; cancel pending callbacks on pause, seek, reset, or load. Keep code, narration, and visual snapshot aligned.
- [ ] Respect `prefers-reduced-motion`, bound transitions by playback interval, and keep variable display ordering deterministic.
- [ ] Run renderer/trace tests and compile with `make wasm`. Add runnable browser regression check for stale playback and duplicate identity when browser harness is available.

## Task 2: Forty dedicated algorithm tasks

**Ownership:** For each ID below, one fresh subagent modifies `internal/algorithms/<id>/trace.go`, `lesson.go`, and existing `algorithm_test.go` as needed. Verify actual filenames before dispatch. Specialized scenes live in `web/renderers/<id>_scene.go` and `<id>_scene_test.go`; no agent edits shared renderer/player files. Existing algorithm computation changes require a demonstrated correctness issue.

**Interfaces:** `Lesson.Build(json.RawMessage) ([]trace.Step, error)` and `Trace` retain current signatures. Use Task 1's `SceneState.RenderSVG(width, height int) string` for new specialized visuals. Each snapshot carries sufficient state to render independently; no renderer reruns the algorithm or depends on mutable globals.

| ID | Required teaching focus |
|---|---|
| bubble | Adjacent comparison, actual swap, settled suffix |
| selection | Candidate minimum, scan, selected placement |
| insertion | Held key, shifted values, growing sorted prefix |
| merge | Split ranges, two source runs, merge destination |
| quick | Pivot, partition boundaries, comparison and placement |
| heap | Heap parent/children, sift-down, extracted suffix |
| countingsort | Frequency counts, cumulative placement, output |
| radixsort | Active digit, buckets, stable collection per pass |
| bucketsort | Distribution, per-bucket sorting, collection |
| linearsearch | Target, checked values, found/not found |
| binarysearch | Low/mid/high, comparison, discarded half |
| kmp | Text/pattern alignment, matched prefix, fallback |
| bfs | FIFO queue, discovered nodes, visit order |
| dfs | Stack/recursive path, descent, return |
| dijkstra | Frontier, candidate distance, relaxation, final path |
| astar | Frontier, g/h/f costs, chosen node, goal path |
| prim | Growing tree, crossing candidate edge, accepted edge |
| kruskal | Ordered edges, components, cycle rejection |
| bellmanford | Passes, relaxation calculation, negative-cycle outcome |
| floydwarshall | Intermediate vertex, source distances, updated pair |
| toposort | Indegrees, ready queue, output ordering/cycle outcome |
| kosaraju | Finish order, reversed edges, component discovery |
| bst | Comparisons along path, insertion/search result |
| avl | Imbalance, rotation participants, restored structure |
| trie | Character path, shared prefixes, word termination |
| segmenttree | Covered intervals, selected nodes, combined answer |
| unionfind | Parent forest, root discovery, union/compression |
| nqueens | Chessboard, threats, attempted placement, backtrack |
| sudoku | Board, row/column/box conflict, backtracking |
| hanoitower | Disk identity, legal peg move, recursion context |
| fibonacci | Two source cells, addition, stored result |
| knapsack | Item/capacity, take/skip alternatives, chosen value |
| lcs | Character comparison, predecessor cells, subsequence |
| coinchange | Amount, coin candidate, source cell, impossible amounts |
| minimax | MAX/MIN turns, propagated score, pruned branches |
| kmeans | Point assignments, cluster identity, centroid movement |
| gradientdescent | Data plot, current line, residuals, parameter update |
| qlearning | Environment, agent/action/reward, Q update and policy |
| perceptron | Labeled points, prediction, boundary/weight update |
| genetic | Population, selected parents, crossover/mutation, fitness |

Each row runs the following checklist independently:

- [ ] Review algorithm, trace, renderer compatibility, source mapping, tests, and accepted inputs. Record concrete teaching gap.
- [ ] Write one focused regression covering critical decision state and final correctness; include relevant Review Focus cases. Assert historical snapshots stay unchanged and source IDs resolve.
- [ ] Run `go test ./internal/algorithms/<id>`; confirm new assertion fails for the missing teaching behavior.
- [ ] Implement meaningful visual states and value-specific narration. Add specialized scene only when shared renderer cannot express required concept. Return shared-contract requests to central owner.
- [ ] Run `go test ./internal/algorithms/<id> ./web/renderers`; report commands and outcomes.
- [ ] Fresh review checks algorithm meaning, visual truth, accessible labels, and ownership boundaries; resolve findings before marking row complete.

## Task 3: Integrate and verify all lessons

**Files:** `web/wasm/player.go`, shared renderers/styles as needed, `internal/catalog/catalog_test.go` or existing all-lesson test location, and a small browser check under `scripts/` if tooling supports it.

**Interfaces:** Player calls `renderers.RenderStepSVG`; existing renderer metadata remains compatible. New state render methods do not import browser packages.

- [ ] Integrate algorithm batches and update per-lesson legend/goal content to match actual state meanings.
- [ ] Add catalog-wide regression building all 40 defaults, checking source IDs and rendering every snapshot without invalid numeric geometry. Keep semantic assertions in each algorithm's tests.
- [ ] Run `go test ./...`, `go test -race ./internal/... ./web/renderers`, `go vet ./...`, and `make build`.
- [ ] Browser-review all 40 default lessons; verify each task's teaching focus. Check mobile, backward/forward seek, pause during motion, replay, speed changes, custom inputs, keyboard, and reduced motion.
- [ ] Compare failures against traces rather than merely adjusting snapshots to rendered output. Fix and rerun affected checks.

## Task 4: Review, commit, push

- [ ] Fresh whole-change review against spec and 40-row coverage; resolve correctness and teaching gaps.
- [ ] Run `git diff --check`, final required checks, and inspect `git status --short` for unrelated files.
- [ ] Commit task-owned changes with `feat: teach all algorithms with clearer visual animations` (split coherent commits if needed).
- [ ] Push current configured branch without force; report commit, branch, verification, and any remaining limitations.

## Progress

- [x] Teaching direction approved by user.
- [x] Source architecture and initial renderer gaps inspected.
- [x] Plan self-reviewed for scope, ownership, interfaces, and verification coverage.
- [ ] User confirms implementation plan; preserve requested one-subagent-per-algorithm method.
