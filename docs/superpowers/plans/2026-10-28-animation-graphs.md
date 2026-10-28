# Graph lesson animations

Implementation handoff for [algorithm animations](2026-10-28-algorithm-animations.md). **Plan only; proposed states and renderers do not exist yet.** Scope: BFS, DFS, Dijkstra, Bellman–Ford, Floyd–Warshall, A*, Prim, Kruskal, topological sort, Kosaraju.

## Evidence and ownership

Read current `lesson.go`, `trace.go`, and `algorithm.go` in all ten algorithm directories, plus `web/renderers/graph.go` and `web/renderers/table.go`. Parent supplied Verify discovery at HEAD `49815a34b5cd2f77321f254c54de80d7bbe8e53e`: 226 symbol results, pagination complete. Project `home-mersad-Projects-gorithm`, root `/home/mersad/Projects/gorithm`, generation `2026-10-27T07:56:32Z`. This document's author confirmed project identity and checked coverage for all 32 source files: `no_recorded_issue`, `metadata_match`. Best-effort metadata, not proof of completeness. Current-source reads govern claims below. Initial worktree clean. No product code or runtime tests changed/run for this documentation task.

Parent owns shared motion contract, renderer dispatch, playback lifecycle, and integration checks. Each implementer owns only assigned algorithm trace/lesson changes, dedicated renderer, and tests. Do not edit shell, themes, generic graph/table renderers, or introduce dependencies. Coordinate any validation/source correction with parent rather than silently widening input semantics.

## Common requirements, not a new framework

- Put state types in `package renderers`; algorithm packages already import it. Never import algorithm packages from renderers. Each section gives exact additional fields; no `map[string]any` visualization model, text parsing, future-event inference, or renderer-side algorithm rerun.
- Required API is `func Render<ID>LessonSVG(step trace.Step, width, height int) string`, returning SVG fragment. Required files are `web/renderers/<id>_lesson.go` and `web/renderers/<id>_lesson_test.go`. Explicit exported names appear below. Algorithm tests belong in existing algorithm test files when possible.
- Preserve Trace signatures. Every recorded state is complete, immutable, and independent: clone embedded nodes/edges, maps, all slices, matrix rows, heap entries, call frames, and nested components. Initial and terminal steps use same typed state. Missing indices use `-1`, absent map entry means unknown, never implicit zero distance. Capture old/candidate values at computation time, before assignment.
- `step.LineID`, explanation, metrics, variables, state, and motion must describe one operation. Existing traces group some Source lines; this document names those groups rather than pretending every Source line already has a frame. Any split must use existing meaningful IDs and preserve algorithm operation counts. No decorative frame invents algorithm progress.
- Identity colors, in node input order modulo five: coral `#c45d75`, blue `#387faa`, violet `#8763bb`, amber `#b67b30`, teal `#258b80`. Keep identity fill across operations. Comparison ring/outline gold `#ffe29a`; completed invariant/result gets green `#9bf0bb` check plus text. Queue membership, current, candidate, tentative, finalized, rejected, and unreachable need labels/shapes, not color alone. Use dark canvas with high-contrast light text; identity colors are accents, not guaranteed text backgrounds.
- Normal operation 1800ms; movement 750ms inside operation, never an extra 750ms step. Native SVG `<animate>`, `<animateTransform>`, `<animateMotion>` or scoped native CSS only, driven by parent contract. Underlying SVG attributes always represent selected snapshot's settled result. Parent suppresses/restarts motion on seek, reverse, reset, pause, lesson change, and reduced motion. Never depend on DOM replacement preserving CSS transitions. Motion metadata describes actual previous/current endpoints, not future outcomes. Check exact parent contract before coding; no separate animation controller.
- Deterministic geometry: use supplied coordinates, fitted once from entire input extent into scene graph box with uniform scale and centered translation. Do not refit from changing visited subset. Degenerate extent uses center on that axis. For missing/coincident layout where necessary, deterministic input-order ring `angle=2π*i/n-π/2`, center graph box center, radii `(boxWidth/2-40, boxHeight/2-40)`; do not mutate algorithm coordinates, especially A* heuristic coordinates. No force layout/randomness.
- Unless overridden, logical canvas 960×640: title/status `(24,30)`, graph box `(24,64,592,348)`, auxiliary panel `(640,64,296,348)`, ordered strips `(24,440,912,72)` and equation/status `(24,536,912,80)`. Nodes radius 22; edges terminate at node boundary; arrowheads on directed algorithms. Edge identity is original input index, not endpoint pair; reciprocal edges use quadratic offsets ±18, self-loops radius 26 above node, parallel edges get indexed offsets. Weighted scenes show **zero** as well as positive/negative weights. Weight pills at curve midpoint offset 10 along normal. Long labels wrap in two text lines with full escaped `<title>`.
- At renderer width below 600, use logical width 360 and stacked graph `(16,56,328,240)`, auxiliary panel starting `(16,320)`, strips below panel, equation last. Compute content height from row counts; fit fragment to supplied viewport with parent-approved SVG viewport contract, not hidden data or unreadably compressed rows. Expose needed intrinsic height to parent during integration if its viewport remains fixed. Wide tables use parent-owned local SVG scrolling where available. No shell changes by per-lesson agent.
- In row panels, header 28 high, data rows 30 high. Lists use chips at least 48×32 with 8 gap; wrap by available width. For custom inputs exceeding one panel, extend logical scene height deterministically. Node/edge IDs in DOM use safe numeric indices plus lesson prefix; escape all labels, titles, values, attributes via `html.EscapeString`. Generic graph/table renderers currently interpolate labels/headers/value strings directly; do not copy that trust boundary.
- Current `GraphState` only contains `Nodes []GraphNode`, `Edges []GraphEdge`. Node fields: `ID, Label string`, `X,Y float64`, `State string`; edge fields: `From,To string`, `Weight int`, `Active,Directed bool`. Current `TableState` contains `RowHeaders,ColHeaders []string`, `Cells []TableCell`; cells have `Row,Col int`, `Value,State string`. Generic graph renderer uses raw coordinates and suppresses zero weights; generic table starts `(100,60)` with 54×38 cells and can shrink fonts. Dedicated scenes replace these presentation limits.

### Checks required for every implementation

- [ ] Trace steps reference valid `Source()` IDs; snapshot mutation tests alter final maps/slices and prove earlier snapshots unchanged.
- [ ] Render each state twice and assert byte-identical settled geometry; render first/middle/last independently. No renderer requires prior calls.
- [ ] Parse fragments inside `<svg>` with `encoding/xml`; test escaped `<`, `&`, quotes in node labels and table headers. Assert no raw injected element/event attribute.
- [ ] Assert operation-specific state and result below, not merely presence of `<svg>` or screenshot similarity. Compare final results with native algorithm on matching semantics.
- [ ] Run `go test ./...`, `go test -race ./...` for trace state changes, and `make build`; account for Docker-backed Go wrapper described in `AGENTS.md`.
- [ ] Parent browser integration: actual WASM load; 320px/phone/tablet/desktop; all themes; seek/reverse same state; reduced motion settled state; pause/reset/navigation cancel pending motion. Screenshot assertions supplement numeric state checks.

## 1. BFS — queue makes distance layers

**Sources:** `internal/algorithms/bfs/lesson.go`, `internal/algorithms/bfs/trace.go`, `internal/algorithms/bfs/algorithm.go`; renderer baseline `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step
```

**Current input/default:** JSON object `{nodes?: GraphNode[], edges?: GraphEdge[], start?: string}`; metadata `{"start":"A"}`. Each nonempty field replaces default independently; malformed JSON silently falls back. Default nodes A(150,225), B(320,130), C(320,320), D(500,130), E(500,320), F(670,225); directed traversal edges AB, AC, BD, BE, CE, DF, EF in that order. Trace builds only forward adjacency regardless of edge `Directed`. Empty nodes return nil; empty start chooses first node; unknown nonempty start currently not rejected. `Traverse(adj,start)` returns nil for empty start. Current edge highlight matches either orientation, so do not reuse `Active` as exact directed edge identity.

**Proposed file/API/state:** `bfs_lesson.go`, `bfs_lesson_test.go`, `RenderBFSLessonSVG`.

```go
type BFSLessonState struct {
    GraphState
    Start, Current, Neighbor string
    Queue, Order []string
    Distance map[string]int
    Parent map[string]string
    Discovered map[string]bool
    Finished map[string]bool
    EdgeIndex int
    Enqueued, Dequeued string
    Skip bool
}
```

Persist queue/order even on steps whose existing `Variables` omit them. Distances are trace instrumentation: start 0, first discovery `Distance[v]=Distance[u]+1`. No new shortest-path rerun.

**Scene:** common graph box; level badges at `(node.x+18,node.y-20)`, `?` before discovery. Right panel columns vertex/level/parent in input order. First bottom strip FIFO left/front to right/back, second strip visit order; equation sits below strips if needed. Default queue transitions `[A]`, `[B,C]`, `[C,D,E]`; ghost chip follows actual node-to-tail route, while front chip moves to current-node halo. Node positions stay fixed: layers are badges, not destructive relayout.

**Teaching/source beats:** L1 setup; L3 enqueue A and mark discovered (currently combines L3/L4); L5 empty order; L6 inspect FIFO; L7 pop (currently combines L7/L8); L9 append order; L10 inspect directed edge; L11 test discovered; L12 mark/enqueue neighbor (currently combines L12/L13); L15 finish adjacency; L17 return. Show `visited` in source means discovered, not necessarily fully explored. Gold at L11; only L15 gets finished check. Distance becomes fixed on discovery for unweighted BFS, separately labeled from finished adjacency.

**Example:** A→B, A→C, B→D gives distances 0,1,1,2. First edge count to D is `d[B]+1=2`; BFS is not minimum weighted cost.

**Motion:** enqueue/dequeue 750ms, edge probe gold pulse within 1800ms, stable distance badge appears on discovery. Already discovered neighbor gets “skip: queued/visited”; no second moving chip.

**Edges/tests:** default order `[A,B,C,D,E,F]`, distances `{A:0,B:1,C:1,D:2,E:2,F:3}`, E enqueued exactly once despite BE and CE. Cycle A→B→A terminates with two visits. Isolated C remains unknown when starting A. Empty Trace remains empty. Validate unknown start/dangling endpoint at Build boundary before richer instrumentation; do not invent visible node. Test directed reverse edge never highlighted by endpoint-symmetric match. Assert L7 queue excludes popped current and L9 order includes it; gold frame cannot already show undiscovered neighbor in queue.

**Implementation sequence**
- [ ] Add `BFSLessonState` and trace snapshots including distances/parents and queue mutations.
- [ ] Implement fixed graph, level ledger, FIFO, order, directed-edge identity, and escaped labels.
- [ ] Add default/cycle/disconnected/source-ID/immutability assertions; pass parent motion and dispatch handoff.

**Copy-ready agent prompt**
> Implement BFS section of `docs/superpowers/plans/2026-10-28-animation-graphs.md`, including common requirements/checks. Read BFS lesson/trace/algorithm and graph renderer first. Preserve exact Trace API; add `renderers.BFSLessonState`, `web/renderers/bfs_lesson.go`, `web/renderers/bfs_lesson_test.go`, `RenderBFSLessonSVG`. Snapshot real FIFO/discovery/level/parent/order at listed Source IDs; default order A,B,C,D,E,F and distances 0,1,1,2,2,3. Implement specified geometry, palette, 1800ms operations/750ms moves through parent's native motion contract. No shell/theme/shared dispatch edits, dependencies, random layout, or invented results. Test directed cycles, duplicate discovery, disconnected nodes, escaping, immutable snapshots, and seek-independent render. Coordinate input validation gaps with parent. Run required checks and report exact files/results/blockers; no commit.

## 2. DFS — call stack and return, not a queue

**Sources:** `internal/algorithms/dfs/lesson.go`, `internal/algorithms/dfs/trace.go`, `internal/algorithms/dfs/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step
```

**Current input/default:** JSON string `"A"` or object `{start?:string,nodes?:GraphNode[],edges?:GraphEdge[]}`. Metadata `DefaultInput` is raw `A`, not JSON-quoted; Build falls back when parsing it. Defaults A(400,80), B(240,200), C(560,200), D(160,340), E(360,340), F(640,340), directed edges AB, AC, BD, BE, CE, CF. Nonempty object arrays replace independently. Trace follows outgoing edges, skips unknown neighbor IDs, returns empty for unknown start or empty nodes; blank start picks first node. Native `Traverse` follows supplied adjacency and returns nil for blank start.

**Files/API/state:** `dfs_lesson.go`, `dfs_lesson_test.go`, `RenderDFSLessonSVG`.

```go
type DFSCallFrame struct { Node, Caller string; NextNeighbor int }
type DFSLessonState struct {
    GraphState
    Start, Current, Neighbor, ReturningFrom string
    Stack []DFSCallFrame
    Order, FinishOrder []string
    Parent map[string]string
    Discovered, Finished map[string]bool
    EdgeIndex int
    BackEdge, Skip bool
}
```

**Scene:** graph common left box preserves tree-like default positions; right panel call frames stacked bottom(root) to top(current), 44px per frame with caller and next neighbor index. Show explicit top arrow; never label stack as visitation order. Bottom visit strip and separate finished strip. Active recursion path gets identity-colored outlines; each call frame retains node identity.

**Source/teaching:** L1/L2/L3 initialize; L14 root call; L6 discovered/visiting; L7 append preorder; L8 inspect adjacency; L9 distinguish unseen, active ancestor/back-edge, finished vertex; L10 push child call; L11 return to parent; L12 finish current; L15 result. Capture child frame at L10/root at L14, update frame neighbor cursor as loop advances, pop on actual return before L11. Terminal root pop occurs before L15. Multiple active ancestors are correct, not several “current” nodes.

**Example:** default enters A,B,D, returns B, enters E, returns A, enters C, skips finished E, enters F. Preorder `[A,B,D,E,C,F]`; finish order `[D,E,B,F,C,A]`. Rule: explore newest unfinished call first. Edge C→E is not a back-edge because E is finished.

**Motion:** 750ms frame push and cursor along parent→child on L10; return cursor child→parent on L11; no node relocation. L9 gold ring plus “already active”/“already finished”; finished check on L12, not first entry.

**Edges/tests:** assert default orders, L10 child stack depth +1, L11 restored parent at top, L15 empty stack. A→B→A marks back-edge at L9, visits each once. DAG diamond detects finished-neighbor skip without cycle claim. Unreachable vertices stay untouched. Unknown start/empty graph preserve current no-step behavior or coordinate explicit validation. Check root-only graph produces preorder/finish `[A]`; snapshot frame slices independent.

**Implementation sequence**
- [ ] Instrument real recursion entry/return and finish order without changing adjacency order.
- [ ] Implement stack/graph/dual-order scene and direction-correct edge probes.
- [ ] Add stack-depth, back-edge versus cross-edge, singleton, and snapshot tests; parent handoff.

**Copy-ready agent prompt**
> Implement DFS section and common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read DFS lesson/trace/algorithm and graph renderer. Keep Trace signature; add `DFSLessonState`, `DFSCallFrame`, `dfs_lesson.go`, `dfs_lesson_test.go`, `RenderDFSLessonSVG`. Snapshot actual recursion stack at L14/L10/L11/L12/L15; preorder A,B,D,E,C,F and finish D,E,B,F,C,A. Render fixed graph, LIFO frames, visit/finish strips with specified palette and parent's 1800ms/750ms native motion contract. Test active-ancestor versus finished-neighbor edges, depth, final empty stack, escaping, immutability, direct seek. No dependencies, shell/themes/shared dispatch edits, fake results, or commit. Report checks and blockers.

## 3. Dijkstra — tentative distance, minimum extraction, relaxation

**Sources:** `internal/algorithms/dijkstra/lesson.go`, `internal/algorithms/dijkstra/trace.go`, `internal/algorithms/dijkstra/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start, target string) []trace.Step
```

**Current input/default:** only `{start?:string,target?:string}`, default A→E; custom nodes/edges are not Build input. Fixed A(120,225), B(320,120), C(320,330), D(520,120), E(680,225); undirected weights AB4, AC2, BC1, BD5, CD8, CE10, DE2. Trace examines incident edges in both directions irrespective of `Directed`. Native `ShortestPath` takes its own `WeightedEdge` with `Directed bool`, and runs until queue empty; Trace exits at target. Do not compare all non-target tentative distances with native final map after early exit. Missing start emits L1/L17; `math.MaxInt` means infinity.

**Files/API/state:** `dijkstra_lesson.go`, `dijkstra_lesson_test.go`, `RenderDijkstraLessonSVG`.

```go
type DijkstraQueueEntry struct { Node string; Distance, Serial int }
type DijkstraLessonState struct {
    GraphState
    Start, Target, Current, Neighbor string
    Distance map[string]int
    Previous map[string]string
    Settled map[string]bool
    Queue []DijkstraQueueEntry
    Path []string
    EdgeIndex, OldDistance, Candidate, PoppedDistance int
    HasCandidate, Improved, Stale, Found, InvalidStart bool
}
```

Queue is cloned actual heap array; serial identifies duplicate entries, not a new tie-break that changes execution. Render heap-array slots with root/min badge, not a fictitious sorted queue. Optional sorted display must be labeled “candidates by priority” and actual extraction remains recorded.

**Scene:** graph common left; right ledger vertex/distance/predecessor/status. Bottom heap slots, then `old d[v]` versus `d[u]+w`; final path replaces only result strip. Infinity is `∞`, never host/WASM-specific MaxInt numeral. Keep target bullseye persistent. Edge weights always visible including zero.

**Source/teaching:** L1/L2 distances; L4 actual initial push; L5 queue test; L6 extract and stale skip; L7 target check; L8 edge; L9 compute candidate; L10 compare; L11 change distance; L12 predecessor; L13 push new entry; L15 finish neighbors; L17 reconstruct actual predecessor path. At valid minimum extraction distance is final under nonnegative weights; mark settled explicitly without confusing existing visited flag, which Trace updates after neighbors and does not update for early-exit target. Do not preview L11 result at L10.

**Example:** A→C costs 2; C→B gives `2+1=3<4`, replace B's tentative 4. B→D gives 8, D→E gives 10. Default path `[A,C,B,D,E]`, cost 10. Old heap entry B:4 remains until skipped.

**Motion:** L6 chip→node 750ms; L9 edge-cost token→equation; L11 old/new distance crossfade, predecessor pointer changes only L12. Stale entry fades into “already settled” pocket. Final path stroke reveal at L17 uses recorded path only.

**Edges/tests:** default E10/path above; fixture A-B4,A-C1,C-B1 produces stale B4 and final B2 when target empty in direct Trace test. Unreachable target: empty path, infinity; start=target path `[A]` cost0. Invalid start terminal uses same typed state. Reject negative weights if graph input later exposed; no silent support claim. Test zero edge visible. Unknown target map lookup currently risks implicit zero during path reconstruction: add membership-aware state/result handling and parent-approved validation, never show zero cost for nonexistent node. Check gold compare retains old distance; L11 changes only intended entry.

**Implementation sequence**
- [ ] Snapshot heap entries, distances, predecessors, candidate operands and settled state at exact operations.
- [ ] Implement graph/ledger/heap/equation/path with infinity and stale-entry labels.
- [ ] Test early exit semantics, stale entries, missing target, zero weights, snapshot independence; parent handoff.

**Copy-ready agent prompt**
> Implement Dijkstra section/common contract from `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read Dijkstra lesson/trace/algorithm and graph renderer. Preserve Trace; add `DijkstraLessonState`, `DijkstraQueueEntry`, dedicated `dijkstra_lesson.go`/`dijkstra_lesson_test.go`, `RenderDijkstraLessonSVG`. Capture real heap duplicates, old/candidate values, predecessors and settled state; distinguish Trace target early exit from native full search. Default path A,C,B,D,E costs10. Use specified geometry/palette and parent's native 1800ms/750ms motion, no separate scheduler. Test stale duplicates, zero weights, absent/unreachable target, start=target, escaping, immutable snapshots. No shell/themes/shared dispatch edits, dependencies, random layouts or commit. Coordinate validation corrections and report checks/blockers.

## 4. Bellman–Ford — repeated passes and reachable negative-cycle proof

**Sources:** `internal/algorithms/bellmanford/lesson.go`, `internal/algorithms/bellmanford/trace.go`, `internal/algorithms/bellmanford/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge, start string) []trace.Step
```

**Current input/default:** `{start?:string,nodes?:GraphNode[],edges?:GraphEdge[]}`, metadata A. Nonempty fields replace independently; malformed input falls back. A(120,225), B(280,120), C(280,330), D(480,120), E(640,225); directed AB4, AC2, BD2, CD5, CE4, DE-3. Trace and native `ShortestPath(nodes, []DirectedEdge,start)` use forward edges. Invalid start emits L1/L18; distances use MaxInt. Relaxations are in-place within each pass, not a previous-pass-only DP copy.

**Files/API/state:** `bellmanford_lesson.go`, `bellmanford_lesson_test.go`, `RenderBellmanFordLessonSVG`.

```go
type BellmanFordLessonState struct {
    GraphState
    Start, From, To string
    Distance, PassStartDistance map[string]int
    Pass, MaxPasses, EdgeIndex, Weight, OldDistance, Candidate int
    HasCandidate, Improved, ChangedThisPass bool
    CheckingCycle, NegativeCycle, InvalidStart bool
    WitnessEdge int
}
```

`PassStartDistance` is comparison context only; algorithm continues reading current Distance. Witness is first relaxable edge in detection scan, not invented full cycle reconstruction.

**Scene:** graph common left with directed arrows and negative-weight pills; right vertex/start-of-pass/current ledger. Bottom original-input-order edge strip with cursor and pass counter `pass p / |V|-1`, then expression row. Final scan has separate header “Can any reachable edge still improve?”; never recolor all finite nodes as permanently finalized after one pass.

**Source/teaching:** L1/L2 initialize; L3 begin pass; L4 reset changed; L5 focus exact edge; L6 guard reachable source then compare; L7 assign; L8 changed=true; L11 early-stop decision; L13 detection scan; L14 test; L15 reachable negative-cycle result or L18 no cycle. A skip for ∞ source must not display `∞ + w` as finite arithmetic. Green check for successful assignment says “updated”, not “final shortest”; final result check only after detection scan.

**Example:** default D gets 6 via A→B→D; E first gets6 via C then `6+(-3)=3<6` via D. Pass1 distances A0 B4 C2 D6 E3; pass2 unchanged; detection reports false.

**Motion:** edge cursor slides 750ms between original-order slots, candidate token traverses directed edge and equation; repeated passes rewind cursor only. Negative-cycle witness remains gold/coral with explicit warning, no green optimal-path celebration.

**Edges/tests:** assert default map above and early-stop pass2. Fixture A→B1,B→C-2,C→B-2 reports true at L15; disconnected negative cycle X↔Y cannot trigger from A. Reversed edge order may need more passes; assertions allow in-place propagation, never claim exactly one extra reachable edge per pass. Empty nodes/invalid start show invalid termination rather than fake optimum. Large integer sums require overflow-safe validation/guard before accepting expanded numeric inputs. Assert witness index really satisfies finite-source inequality in stored detection snapshot; no post-detection mutation.

**Implementation sequence**
- [ ] Add complete pass/edge/operand snapshots and detection phase flags.
- [ ] Implement graph plus pass ledger, original edge strip and guarded equation.
- [ ] Test default/early stop/reachable versus unreachable negative cycles, source IDs and immutable pass snapshots; parent handoff.

**Copy-ready agent prompt**
> Implement Bellman–Ford section/common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read actual lesson/trace/algorithm and graph renderer. Keep Trace API; add `BellmanFordLessonState`, dedicated `bellmanford_lesson.go`/`bellmanford_lesson_test.go`, `RenderBellmanFordLessonSVG`. Expose real in-place pass state, changed flag, edge index, guarded operands and negative-cycle witness; no fabricated cycle path. Default A0 B4 C2 D6 E3, unchanged pass2, no negative cycle. Use specified geometry/colors and parent's native 1800ms/750ms motion. Test reachable/unreachable negative cycles, pass ordering, infinity, escaping, snapshots. No shell/themes/shared dispatch edits, dependencies or commit. Coordinate input guards; report checks/blockers.

## 5. Floyd–Warshall — two source cells, one destination

**Sources:** `internal/algorithms/floydwarshall/lesson.go`, `internal/algorithms/floydwarshall/trace.go`, `internal/algorithms/floydwarshall/algorithm.go`; `web/renderers/table.go`, `web/renderers/graph.go`.

```go
func Trace(nodes []string, edges []WeightedEdge) []trace.Step
```

**Current input/default:** `{nodes?:string[],edges?:WeightedEdge[],graphNodes?:GraphNode[],graphEdges?:GraphEdge[]}`. WeightedEdge `{from,to:string,weight:int,directed?:bool}`; omitted directed means undirected. Nonempty nodes precede graphNodes; nonempty edges precede graphEdges; graphEdges converted directed=true. Defaults A,B,C,D; directed AB3, AC8, BC2, BD5, CD1, DA2. Invalid/empty fields preserve defaults. Trace ignores unknown endpoints, takes minimum parallel weight, diagonal starts0 and negative self-loop can lower it. Native `AllPairsShortestPath` same matrix operation. Current negative-cycle branch ends on L16 `ActionHighlight`, not ActionDone.

**Files/API/state:** `floydwarshall_lesson.go`, `floydwarshall_lesson_test.go`, `RenderFloydWarshallLessonSVG`.

```go
type FloydWarshallLessonState struct {
    TableState
    Nodes []string
    Edges []GraphEdge
    Distance [][]int
    I, J, K int
    LeftDistance, RightDistance, OldDistance, Candidate int
    HasCandidate, Improved, CheckingDiagonal, NegativeCycle bool
    DiagonalIndex int
}
```

Edges preserve WeightedEdge direction. TableState remains available, but typed Distance and saved operands govern formulas; do not parse cell text. Capture operands before write, crucial when source/destination cells overlap.

**Scene:** desktop matrix `(72,100)` with 58×48 cells (default 4×4), labels row origin x56 and column y82. Center/right mini route nodes i at(420,176), k(590,96), j(760,176); arrows label saved `d[i][k]`, `d[k][j]`; direct comparison arc y264 labels old `d[i][j]`. This is distance-route schematic, not claim of direct input edges. k selector chips `(24,360)`; full equation `(24,440)`; diagonal/status below. Large n extends matrix size/scene, no tiny fonts. Mobile matrix origin `(48,80)`, cell width/height 48; route schematic stacks below, using local scroll integration for wider matrix.

**Source/teaching:** L1 already initialized matrix (L2 not separately emitted); L3 choose intermediate; L6 unreachable half-path skip; L7 compare saved operands; L8 commit only if smaller; L14 diagonal scan; L16 negative diagonal or L19 complete. Optional granular i/j focus uses L4/L5 only when real loop events are recorded. Formula `d[i][j]=min(d[i][j],d[i][k]+d[k][j])`; row/column index labels always visible. Gold outlines source cells and destination; source/destination overlap uses concentric outlines plus role labels, not precedence that hides one role.

**Example:** k=B, i=A, j=C: `min(8,3+2)=5`; later k=C gives A→D `5+1=6`. Expected final matrix rows A `[0,3,5,6]`, B `[5,0,2,3]`, C `[3,6,0,1]`, D `[2,5,7,0]`.

**Motion:** 750ms operand copies from source cells into equation; commit token to destination at L8. Keep authoritative matrix value settled per selected snapshot. No visual graph path reconstruction without explicit predecessor state. Negative diagonal ring plus “negative cycle; finite optimum not defined for affected pairs,” no fake matrix of final shortest paths.

**Edges/tests:** assert full matrix above against native function. Infinity+finite never produces Candidate; unreachable cells remain∞. Duplicate directed edges AB8/AB3 initialize3; undirected flag mirrors weight. A→A-1 triggers L16 and no successful L19. Empty direct Trace yields valid empty matrix completion. Test I=K/J=K alias cases with saved prewrite values. Decide with parent whether to change L16 to ActionDone for shared terminal treatment; do not silently append L19 false result.

**Implementation sequence**
- [ ] Snapshot numeric matrix, operands, indices, edge direction, diagonal phase.
- [ ] Implement matrix/route schematic/equation with overlapping-role outlines and infinity guard.
- [ ] Add exact matrix, negative diagonal, duplicate/undirected edge and mutation tests; coordinate terminal action/viewport with parent.

**Copy-ready agent prompt**
> Implement Floyd–Warshall section/common requirements from `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read lesson/trace/algorithm and graph/table renderers. Preserve Trace; add `FloydWarshallLessonState`, `floydwarshall_lesson.go`, `floydwarshall_lesson_test.go`, `RenderFloydWarshallLessonSVG`. Snapshot numeric matrix and prewrite operands, render exact matrix plus i-k-j distance schematic, no invented predecessor path. Verify final default rows [0,3,5,6],[5,0,2,3],[3,6,0,1],[2,5,7,0]. Use parent motion 1800ms/750ms and palette; test infinity, negative diagonals, overlap, duplicates, escaping, immutable rows. Coordinate current L16 Highlight terminal and wide-matrix viewport. No shell/themes/shared dispatch edits, dependencies or commit. Report checks/blockers.

## 6. A* — show g, h, f separately

**Sources:** `internal/algorithms/astar/lesson.go`, `internal/algorithms/astar/trace.go`, `internal/algorithms/astar/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []AStarNode, edges []AStarEdge, start, goal string) []trace.Step
```

**Current input/default:** `{nodes:AStarNode[],edges:AStarEdge[],start?:string,goal?:string}`; node `{id,label?,x,y}`, edge `{from,to,weight,undirected?:bool}`. Custom accepted only with nonempty nodes; then edges replaced even empty, endpoints default first/last node. Blank labels become IDs; coordinates both0 become deterministic ellipse `(400+180cosθ,225+130sinθ)`. Otherwise default A(140,220), B(300,120), C(300,320), D(500,120), E(500,320), F(660,220); undirected AB190, AC200, BD200, CE200, BE280, CD280, DF190, EF200; A→F. Trace/native FindPath use rounded Euclidean `int(math.Round(math.Hypot(dx,dy)))`; heap priority f then h. Closed nodes never reopen. Heuristic correctness is conditional, not guaranteed for arbitrary custom coordinates/weights.

**Files/API/state:** `astar_lesson.go`, `astar_lesson_test.go`, `RenderAStarLessonSVG`.

```go
type AStarOpenEntry struct { Node string; G, H, F, HeapIndex int }
type AStarLessonState struct {
    GraphState
    Start, Goal, Current, Neighbor string
    G, H, F map[string]int
    CameFrom map[string]string
    Closed map[string]bool
    Open []AStarOpenEntry
    Path []string
    EdgeUndirected []bool
    EdgeIndex, OldG, TentativeG, Cost int
    KnownOldG, HasTentative, Improved, Found, InvalidEndpoints bool
    HeuristicConsistent bool
}
```

`HeuristicConsistent` is computed from actual rounded h for all traversable edges (`h(u)≤w+h(v)`, h(goal)=0, nonnegative weights); this supports this closed-set implementation's guarantee. Label custom violations truthfully; no claim that admissibility alone fixes no-reopen behavior. H uses original input coordinates, never fitted display positions.

**Scene:** common graph left, right ledger node/g/h/f/status; target bullseye, current→goal dashed heuristic segment labeled rounded distance. Open heap chips below, f main value with g+h underneath, stable input identity. Equation strip displays `tentative g`, previous known g or∞, then `f=g+h`. Dashed heuristic is estimate, not traversable edge or accumulated path.

**Source/teaching:** L1/L2 setup; L3 push start; L4 open check; L5 min extraction; L6 goal; L9 edge and closed skip; L10 tentative g; L11 compare; L12 parent; L13 g; L14 f; L15 push/update same open entry; L7 actual path/cost; L19 no path. Missing g must display∞, despite current explanation formatting absent map value as0. Closed check shown before arithmetic.

**Example:** start h(A)=520. B gets g190,h374,f564; D via B gets g390,h189,f579; F via D gets g580,h0,f580. Default result `[A,B,D,F]`, cost580. h is heuristic, not remaining true path cost; input coordinates are algorithm data.

**Motion:** 750ms pop chip to current halo; dashed estimate drawn once per active node; g/h tokens join f on L14; L15 existing chip changes priority in actual heap slot snapshots, not a duplicated node. Final path follows recorded CameFrom only.

**Edges/tests:** default result and h values above; start=goal returns `[start]`,0; unreachable/invalid endpoints no path with cost-1; empty nodes zero steps. Undirected=false allows only forward travel. Equal f uses lower h; decrease-key leaves one open entry per node. Test display coordinate normalization cannot change H. Custom overestimating/inconsistent heuristic must show conditional result rather than “guaranteed shortest”; coordinate/weight validation needs parent agreement. Assert missing g is∞ and closed neighbor cannot animate relaxation.

**Implementation sequence**
- [ ] Snapshot maps, real open heap, direction flags, operands, and heuristic condition.
- [ ] Implement fixed graph, heuristic segment, g/h/f ledger and heap/equation.
- [ ] Test result, tie/decrease-key, custom heuristic limits, original-coordinate invariant and snapshots; parent handoff.

**Copy-ready agent prompt**
> Implement A* section/common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read A* lesson/trace/algorithm and graph renderer. Keep Trace signature; add `AStarLessonState`, `AStarOpenEntry`, `astar_lesson.go`, `astar_lesson_test.go`, `RenderAStarLessonSVG`. Snapshot actual g/h/f/open/closed/parents, preserve rounded Euclidean input-coordinate semantics and f-then-h heap. Default A,B,D,F cost580; never assert arbitrary custom heuristic guarantees optimality. Render specified graph/ledger/estimate/equation using parent native 1800ms/750ms motion. Test decrease-key, direction, missing g as infinity, invalid endpoints, start=goal, coordinate normalization, escaping and snapshots. No shared dispatch/shell/themes edits, dependencies or commit. Report checks/blockers.

## 7. Prim — growing tree and its cut

**Sources:** `internal/algorithms/prim/lesson.go`, `internal/algorithms/prim/trace.go`, `internal/algorithms/prim/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []WeightedEdge, start string) []trace.Step
```

**Current input/default:** `{nodes?:GraphNode[],edges?:WeightedEdge[],start?:string}`; WeightedEdge `{from,to,weight}` undirected. Custom applied only if edges nonempty; absent nodes derived from first appearance of edge endpoints, on rounded ellipse center400,225 radii250,140. Start override only in accepted custom branch. Invalid/empty start falls back first node. `Build(nil)` uses default coordinates A(150,225), B(300,100), C(300,350), D(500,100), E(500,350), F(650,225); metadata JSON supplies edges without nodes, so Build(metadata) produces ellipse instead. Default weights AB4, AC2, BC1, BD5, CD8, CE10, DE2, DF6, EF3. Trace/native MST grow only start's connected component; current terminal prose incorrectly calls disconnected partial tree complete MST.

**Files/API/state:** `prim_lesson.go`, `prim_lesson_test.go`, `RenderPrimLessonSVG`.

```go
type PrimCandidate struct { From, To string; Weight, EdgeIndex, Serial int }
type PrimLessonState struct {
    GraphState
    Start, Current string
    InTree map[string]bool
    Candidates []PrimCandidate
    SelectedEdges []int
    ActiveEdge, TotalWeight, CandidateWeight int
    CutEdges []int
    Discarded, Connected bool
}
```

Preserve original edge identity when building adjacency; candidate carries traversal orientation and original index. `CutEdges` are edges with exactly one endpoint InTree. Actual heap may still contain obsolete internal edges; distinguish from current mathematical cut.

**Scene:** graph common left with small “in tree” check badges, no enclosing polygon that misleadingly includes unrelated nodes. Cross-cut edges dashed gold, selected edges solid identity accent plus check, obsolete heap entries labeled internal. Right panel membership list and `selected count / V−1`, total. Bottom candidate heap and accepted edge strip; current candidate equation `total + weight` only on L13.

**Source/teaching:** L1/L2 setup; L3 mark start; L4 heap; L5 push start incident edges; L6 loop; L7 min edge; L8 test destination; L9 discard internal edge; L11 add vertex; L12 append tree edge; L13 sum; L14/L15 inspect/push new cut candidates; L18 result. Do not add edge or total at pop, because cycle check still pending. First vertex joins for free; MST has V−1 edges only if connected.

**Example:** default picks AC2, CB1, BD5, DE2, EF3, total13. AB4 becomes obsolete after C joins B via weight1; explain cheaper candidate can reveal later, so chosen weights are not globally sorted (2 then1).

**Motion:** 750ms candidate from heap to probe overlay; accepted edge stroke draws at L12, sum token moves at L13; discarded edge to labeled reject pocket at L9. Node positions fixed while membership/cut changes.

**Edges/tests:** assert default selected indices `[1,2,3,6,8]`, total13, 5 edges; CB orientation still references original BC index2. Empty Trace one L1 done empty state; singleton L18 zero weight/edges. Disconnected A-B1 plus C reports partial start-component tree, Connected=false, no “spanning all vertices” check. Negative/zero weights valid MST inputs. Parallel edges must mark chosen original index, not all endpoint matches (existing `isEdgeInMST` cannot distinguish). Equal-weight heap ties need deterministic repeat behavior without falsely promising lexical order. Validate dangling endpoints before accepting custom graph.

**Implementation sequence**
- [ ] Carry input edge indices through actual heap and snapshot cut/tree/total separately.
- [ ] Implement cut/heap/accepted-strip scene and truthful disconnected result.
- [ ] Test selected indices, staged total updates, obsolete candidates, singleton/empty/parallel edges and snapshots; parent handoff.

**Copy-ready agent prompt**
> Implement Prim section/common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read Prim lesson/trace/algorithm and graph renderer. Preserve Trace; add `PrimLessonState`, `PrimCandidate`, `prim_lesson.go`, `prim_lesson_test.go`, `RenderPrimLessonSVG`. Capture original edge identity through heap, current cut, tree membership, accepted edges and staged total. Default indices [1,2,3,6,8], cost13. Label disconnected result start-component tree, not full MST. Use specified geometry/palette and parent native 1800ms/750ms motion. Test obsolete candidates, negative/zero/parallel edges, singleton/empty, escaping, snapshots. No shell/themes/shared dispatch changes, dependencies or commit. Coordinate validation and report checks/blockers.

## 8. Kruskal — sorted edges and component union

**Sources:** `internal/algorithms/kruskal/lesson.go`, `internal/algorithms/kruskal/trace.go`, `internal/algorithms/kruskal/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []WeightedEdge) []trace.Step
```

**Current input/default:** `{nodes:GraphNode[],edges:WeightedEdge[]}`; custom accepted only when both nonempty. Missing labels/state filled; zero-zero coordinates get input-order ellipse center400,225 radii180,130. WeightedEdge `{from,to,weight}` undirected. Defaults A(160,140), B(400,100), C(640,140), D(160,320), E(400,350), F(640,320); AB4, AD2, BC5, BD1, BE3, CE6, CF2, DE7, EF4. Trace sorts cloned edges by weight with `slices.SortFunc`, uses map DSU path compression/rank, stops at V−1. Native MST returns forest if disconnected; no explicit connected flag. Empty Trace zero steps.

**Files/API/state:** `kruskal_lesson.go`, `kruskal_lesson_test.go`, `RenderKruskalLessonSVG`.

```go
type KruskalLessonState struct {
    GraphState
    SortedEdgeIndices, SelectedEdges, RejectedEdges []int
    Cursor, ActiveEdge, TotalWeight int
    Parent map[string]string
    Rank map[string]int
    RootFrom, RootTo string
    Components [][]string
    Cycle, Connected bool
}
```

Capture DSU roots at L5 after actual Find; copying map must not call Find and silently compress paths during rendering/snapshot. Derive Components by read-only root walks, member and component order based on input order. Edge indices retained through sorting; ties preserve whatever defined comparator produces unless parent approves explicit original-index tie-break.

**Scene:** graph common left; right DSU table node/parent/rank with component bracket labels. Bottom sorted edge strip (original edges before L2), cursor, accepted-cost strip. Component membership indicated by named group badge; keep individual identity fill. If component rows wrap, group brackets span explicit rows rather than guessed spatial convex hull.

**Source/teaching:** L1 initial graph; L2 sorted queue; L3 singleton DSU; L4 focus edge; L5 compare roots, reject same root; L6 actual union; L7 append edge and increment total; L10 forest/tree result. Green union check at L6 means components merged; edge becomes selected only L7. Path compression is real DSU mutation during Find; show changed parent rows, not fabricated search path animation without recorded parents.

**Example:** BD1 joins B,D; AD2 joins A; CF2 joins C,F; BE3 joins E. AB4 would close A-B-D cycle and is rejected; EF4 joins remaining components. Expected MST edges BD,AD,CF,BE,EF, total12; AD and CF same weight need not be claimed as uniquely ordered mathematical choice.

**Motion:** L2 chips move to recorded sorted positions in 750ms; L5 gold root badges; L6 component bracket merges and parent pointers change; L7 edge draws/weight token adds. Rejected edge stays dim with “same component” cross, no removal from original graph.

**Edges/tests:** default selected edge set indices `{3,1,6,4,8}`, total12, 5 edges, acyclic; AB rejected when inspected. Equal weights test invariant/minimum cost and repeated determinism, not unique MST. Disconnected A-B1,C-D2 returns forest weight3/two edges/Connected=false. Self-loop rejected; negative/zero valid. Parallel edges distinct; light one accepted, heavy one cycle-rejected if inspected before early termination. Dangling nodes require boundary validation to avoid DSU empty-root behavior. Assert at L6 components merged but selected count unchanged until L7, and parent-map mutation does not leak backward.

**Implementation sequence**
- [ ] Carry original indices through sorted edges; snapshot real DSU maps/read-only components.
- [ ] Implement graph/sorted-strip/DSU/union scene and connected versus forest result.
- [ ] Add minimum-cost/acyclicity, union staging, ties, self-loop, disconnected and snapshot tests; parent handoff.

**Copy-ready agent prompt**
> Implement Kruskal section/common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read Kruskal lesson/trace/algorithm and graph renderer. Preserve Trace; add `KruskalLessonState`, `kruskal_lesson.go`, `kruskal_lesson_test.go`, `RenderKruskalLessonSVG`. Snapshot sorted original edge indices and actual DSU parent/rank/components without mutating DSU during render. Default selected set {3,1,6,4,8}, cost12. Animate real sort/root compare/union/accept using parent native 1800ms/750ms and palette; label disconnected result forest. Test union before acceptance, cycles/ties/parallel edges, immutable maps and escaping. No shell/themes/shared dispatch edits, dependencies or commit. Report checks/blockers.

## 9. Topological sort — remaining prerequisites and ready queue

**Sources:** `internal/algorithms/toposort/lesson.go`, `internal/algorithms/toposort/trace.go`, `internal/algorithms/toposort/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step
```

**Current input/default:** `{nodes?:GraphNode[],edges?:GraphEdge[]}`, metadata `{}`; nonempty arrays independently replace defaults. Defaults clean(100,225), lint(280,130), compile(280,320), test(460,130), docs(460,320), package(640,225), deploy(800,225). Directed clean→lint, clean→compile, compile→test, compile→docs, lint→package, test→package, docs→package, package→deploy. Forward semantics independent of GraphEdge.Directed. Native `Sort(nodes,[][2]string)` returns order,bool. Trace empty graph zero steps; native empty result is valid true. Current metadata implies every unresolved node participates in a cycle; more precise result is blocked nodes may include downstream non-cycle vertices.

**Files/API/state:** `toposort_lesson.go`, `toposort_lesson_test.go`, `RenderToposortLessonSVG`.

```go
type ToposortLessonState struct {
    GraphState
    InDegree map[string]int
    Queue, Order, Blocked []string
    Processed map[string]bool
    ResolvedEdges []int
    Current, Neighbor string
    EdgeIndex, OldInDegree, NewInDegree int
    Decremented, Ready, Valid, Terminal bool
}
```

Carry all fields on every step, unlike current variable maps that omit queue in terminal state. `Blocked` is remaining unprocessed node IDs in input order; never label all as cycle members without separate proof.

**Scene:** default graph mapped left-to-right into `(24,88,912,220)` so dependency chain remains readable; bottom left ready FIFO `(24,344,432,80)`, bottom right order `(480,344,456,80)`; indegree ledger `(24,448)` rows vertex/remaining prerequisites/status. Mobile graph stacks and ledger follows FIFO/order. Node rectangular pill 92×44 for task names, count badge top-right. Resolved edges dashed/dim, original topology retained; no physical node deletion hiding prerequisites.

**Source/teaching:** L1/L2 compute indegrees; L4 enqueue all initial zero-degree nodes in input order; L5 empty order; L6 queue test; L7 pop; L8 append processed; L9 focus dependency; L10 decrement; L11 nonzero check or L12 enqueue when zero; L16 compare processed count with total. At L8 node enters order before its outgoing edges resolve, so don't mark all edges resolved prematurely.

**Example:** package initially3 prerequisites. After lint it is2, after test1, after docs0 and enters queue. Default order `[clean,lint,compile,test,docs,package,deploy]`; recurrence `inDegree[v] = inDegree[v] - 1` per resolved incoming edge.

**Motion:** L7 chip slides FIFO→current task; L8 duplicate label moves to order (graph task remains); L10 one prerequisite token leaves neighbor badge toward resolved edge in 750ms. L12 queue append only on actual zero. Cycle result holds unresolved edges plus text, not fake completion check.

**Edges/tests:** assert default order and each edge position(u)<position(v); package counts3,2,1,0 exactly on respective L10 frames, enqueued once at0. Multiple roots maintain input order. A→B,B→A,B→C yields Valid=false, Blocked A,B,C; C must be “blocked”, not proven cycle member. Self-loop invalid; duplicate edges increment/decrement twice without premature zero. Disconnected acyclic nodes all processed. Validate dangling endpoints/duplicate node IDs before custom build. Empty Trace renderer accepts empty typed fixture without crash. Snapshot indegrees/queue/order remain independent.

**Implementation sequence**
- [ ] Snapshot indegrees/old-new operands, resolved edge indices, queue/order/blocked result.
- [ ] Implement task pills, prerequisite badges, queue/order lanes and blocked status.
- [ ] Test ordering invariant, package countdown, cycles with downstream nodes, multiple roots and immutable snapshots; parent handoff.

**Copy-ready agent prompt**
> Implement topological-sort section/common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read toposort lesson/trace/algorithm and graph renderer. Preserve Trace; add `ToposortLessonState`, `toposort_lesson.go`, `toposort_lesson_test.go`, `RenderToposortLessonSVG`. Capture actual Kahn indegrees/ready queue/order/resolved edges at listed IDs. Default clean,lint,compile,test,docs,package,deploy; package countdown3,2,1,0. Render specified task graph/FIFO/order/ledger with parent 1800ms/750ms native motion and palette. Blocked nodes are not all necessarily cycle members. Test edge ordering, multiple roots, cycle downstream, duplicate edges, snapshots and escaping. No shell/themes/shared dispatch edits, dependencies or commit. Report checks/blockers.

## 10. Kosaraju — finish order, transpose, components

**Sources:** `internal/algorithms/kosaraju/lesson.go`, `internal/algorithms/kosaraju/trace.go`, `internal/algorithms/kosaraju/algorithm.go`; `web/renderers/graph.go`.

```go
func Trace(nodes []renderers.GraphNode, edges []renderers.GraphEdge) []trace.Step
```

**Current input/default:** `{nodes:GraphNode[],edges:GraphEdge[]}`; custom accepted when nodes nonempty, edges replaced even empty. Blank labels/state filled; zero-zero positions ellipse center400,225 radii180,130. Defaults A(140,150), B(260,150), C(200,300), D(380,220), E(480,300), F(600,150), G(700,260); directed AB, BC, CA, CD, DE, ED, EF, FG, GF. Trace forward DFS ignores unknown neighbor IDs; second pass reverses displayed edge endpoints at L7, restores original direction at final L16. `FindSCCs(nodes,[][2]string)` is native recursive oracle; empty graph nil/zero trace steps.

**Files/API/state:** `kosaraju_lesson.go`, `kosaraju_lesson_test.go`, `RenderKosarajuLessonSVG`.

```go
type KosarajuLessonState struct {
    GraphState
    Pass int
    Transposed bool
    Current, Neighbor string
    Stack, FinishOrder, CurrentComponent []string
    Components [][]string
    ComponentOf map[string]int
    VisitedForward, VisitedReverse map[string]bool
    OrderIndex, EdgeIndex int
    OriginalEdges []GraphEdge
    FinishedVertex, ReturningFrom string
}
```

Component indices assigned in discovery order, -1/absent before assignment. Keep real original edge indices across transpose. Explicit call stack instrumentation inside dfs1/dfs2; helper operations currently share L5 or L12 Source IDs, distinguish by Action and typed fields, not new pseudo-code lines.

**Scene:** graph common left; right panel pass badge plus call stack; bottom finish stack with latest finish at right/top, and component cards below. Default group cards expected `{A,C,B}`, `{D,E}`, `{F,G}` in discovered order. Brackets can outline group memberships but node identity fill remains fixed. Component badge `SCC 1`, etc., stable palette accent and label. Original/transpose label prominent; final original graph can show component badges and intercomponent edges, no new condensation graph unless explicitly stored.

**Source/teaching:** L1/L2 build both adjacency maps; L3 visited/order; L4 root check; L5 entry, edge probe, recursive return/finish push; L6 forward pass complete; L7 reset/reverse displayed graph and begin pass2; L8 inspect finish-order index descending; L11 empty component; L12 reverse DFS and add members; L13 commit component; L16 restore original direction/result. Transpose was logically available at L2; L7 visual switch does not claim Source L7 computes a new algorithmic adjacency structure. Finish order differs from discovery order.

**Example:** default first-pass finish `[G,F,E,D,C,B,A]`; process A first in reverse finish order. In transpose A reaches C then B, SCC `{A,B,C}`; then `{D,E}`, `{F,G}`. Definition: u and v same SCC iff paths exist both ways. One-way C→D connects SCCs but does not merge them.

**Motion:** L5 finished chip travels to finish stack in750ms; L7 arrowheads reverse at fixed node positions with original/reverse overlay, never node shuffle. L8 cursor walks stack right→left without deleting stored finish order. L12 node badge joins current card; L13 card gets check; final L16 reverses arrows back with explicit original-graph label. No component-colored claim before actual reverse visit.

**Edges/tests:** assert default finish order above and component sets `{A,B,C}`, `{D,E}`, `{F,G}`; verify every original edge reversed in L7 by same input index, restored at L16. ComponentOf assigns each vertex once, current component grows only L12, committed list changes only L13. Single node/self-loop yields one SCC; edgeless custom nodes each yield singleton; disconnected chains yield singleton SCCs, not weak components. Direction-only bridge must not merge groups. Reciprocal/parallel edge marker identity must stay distinct. Deep-copy nested Components; mutate final inner slice and prove prior card unchanged.

**Implementation sequence**
- [ ] Snapshot pass, real stacks/finish order, original edge indices/orientations, current/committed components.
- [ ] Implement fixed graph, transpose switch, finish stack and SCC cards with explicit labels.
- [ ] Add finish order, reversal/restoration, component partition, singleton/bridge and nested-copy tests; parent handoff.

**Copy-ready agent prompt**
> Implement Kosaraju section/common requirements in `docs/superpowers/plans/2026-10-28-animation-graphs.md`. Read Kosaraju lesson/trace/algorithm and graph renderer. Preserve Trace; add `KosarajuLessonState`, `kosaraju_lesson.go`, `kosaraju_lesson_test.go`, `RenderKosarajuLessonSVG`. Snapshot real forward/reverse calls, finish order, transpose orientation, current and committed components; L5/L12 helper events retain existing Source IDs. Default finish G,F,E,D,C,B,A and SCC sets ABC,DE,FG. Implement specified fixed geometry, labeled transpose/stack/cards and parent 1800ms/750ms native motion/palette. Test exact edge reversal/restoration, one-way bridge, edgeless/singleton, nested snapshots, escaping. No shell/themes/shared dispatch edits, dependencies or commit. Report checks/blockers.

## Integration dependencies and unresolved product decisions

- Parent must provide final motion/viewport contract and dispatch names exactly as listed; these renderer functions alone cannot safely control reduced-motion, seeking, pause, or SVG timeline lifetime.
- Several existing Build methods silently accept malformed input by falling back, cannot express empty edge arrays, or accept dangling endpoints. Sections document current behavior, not endorsement. Implementers must coordinate minimal validation needed for truthful typed states; no silent schema expansion.
- Dijkstra Trace is undirected and target-early-exit; native algorithm supports directed/full search. A* heuristic uses original coordinates and closes permanently. Prim disconnected result covers one component, Kruskal yields forest. These differences must survive teaching copy and oracle test design.
- Existing generic edge flags cannot distinguish parallel edges reliably; explicit input indices in new states are required. Directed traces must draw actual traversal direction even when current default GraphEdge.Directed is false.
- Floyd–Warshall negative-cycle terminal action and narrow-view intrinsic height need parent resolution. They do not block writing dedicated numeric states/tests, but block claiming complete playback/browser acceptance.

Documentation verification: source reads and 32-path coverage check performed. Planned numeric expectations above are source-derived acceptance targets, not claims of tests already run. Before implementation completion, record actual commands, outputs, browser evidence, and remaining limits in parent report.
