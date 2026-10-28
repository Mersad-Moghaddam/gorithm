# AI Algorithm Animation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `executing-plans` to implement task-by-task, or `subagent-driven-development` when parent explicitly delegates. Steps use checkboxes. Do not commit without authorization.

**Goal:** Teach six real Go algorithms through operation-level SVG scenes, with numerical explanations tied to recorded source lines.

**Architecture:** Algorithms emit immutable typed lesson snapshots; pure Go renderers consume `trace.Step`. Parent owns renderer dispatch, shared motion, playback timing, cancellation, and browser integration. No replacement JavaScript simulation.

**Tech Stack:** Go 1.24+, Go/WASM, `syscall/js`, SVG, existing CSS; no dependencies.

**Spec:** Approved visual grammar: `docs/superpowers/specs/2026-10-28-animation-references/gradient-descent.html`. Root integration plan: `docs/superpowers/plans/2026-10-28-algorithm-animations.md`; design: `docs/superpowers/specs/2026-10-28-algorithm-animations-design.md`. Root shared contract takes precedence for timing, ownership, dispatch and animation attributes. Any prompt mentioning “v5 prototype” means this durable reference.

## Global constraints and parent handoff

- Prototype supplies composition, color, trail, tangent, target marker, equation, and explanation grammar. Its `L(w)=w²`, initial `w=4`, stopping threshold, rate presets, timers, HTML shell, and JavaScript simulation are **not** algorithm requirements. Actual gradient lesson is two-parameter linear regression.
- Keep existing project themes, page layout, controls, routing, source tabs, and dataset dialog. Changes below live inside visualization SVG and lesson trace/input code. Do not edit `web/static/themes.css` or redesign shell.
- Semantic colors: current parameter `#ee8b9b`, loss curve `#aa92e7`, tangent `#ffe29a`, update `#8bd5ff`, known minimum `#9bf0bb`. Entity identity colors, in order: coral `#c45d75`, blue `#387faa`, violet `#8763bb`, amber `#b67b30`, teal `#258b80`. Pair colors with IDs, shapes, and text; palette wraps after five IDs, labels do not.
- Meaningful normal-speed operation: **1800ms**, with movement normally **750ms**, remaining time for reading. Root compare/commit pair budget is 600ms + 1200ms; motion is capped at 60% of current hold, so shorter holds shorten movement. Use root `HoldMS` contract, never synthetic interpolation steps or duplicate filler frames. Real initialization, comparisons, mutation, return, and termination frames remain seekable; renderer does not start timers.
- Each new file declares explicit `<ID>LessonState` and `Render<ID>LessonSVG(step trace.Step, width, height int) string` in package `renderers`. Exact ID spellings below are binding. State types must not import algorithm packages: those packages already import renderers. Reuse existing state by embedding where useful, but do not parse formatted table cells or `Explanation` to recover numbers.
- Parent dispatch requirements: map lesson IDs to renderers below; accept new state variants; preserve source/variables/metrics synchronization; use stable SVG entity keys; snap to selected recorded endpoint for seek/back/reset/reduced motion. Forward interpolation may show transition only, never recompute training or invent extra optimizer evaluations. Cancel movement on pause/navigation/data replacement. For static direct calls, renderer returns complete selected snapshot without prior DOM.
- Geometry: renderer returns SVG **fragment**, never outer `<svg>`; parent owns viewBox/sizing. Root group uses `data-lesson` and `data-phase`; movable groups use stable `data-entity`, finite `data-x`/`data-y`, and `transform="translate(x y)"`. Follow root `data-motion`, `data-path`, and `data-from-entity` contracts. At width >=640, use 24px inset, scene occupying top 65%, equation/reason beneath; at narrower widths use 16px inset, vertically stacked panels, minimum 14px labels, wrapped equations. Parent must supply adequate height or local viewport, never shrink a 760px diagram to illegible 320px text. Scene-specific dimensions below are reference desktop coordinates, proportionally mapped into plot rectangles, not fixed page widths. Clamp/guard zero or negative dimensions. Clip plotted geometry, not labels; avoid document overflow.
- Snapshot all slices, nested slices, maps, and history before `AddStep`; `trace.Builder` does not deep-copy state. Store only prefix history available at selected step. Stable index IDs survive assignment changes and generation replacement where identity persists. Round for display only. Escape all dataset text with `html.EscapeString`, including attributes; use renderer-prefixed marker/clip IDs.
- Keep `Trace` signatures. `Lesson.Build` has error channel; reject malformed explicitly supplied input instead of silently substituting defaults. Defaults apply to omitted fields/empty raw input. Proposed limits below are browser lesson bounds, not claims about algorithm theory. Validate finite inputs and finite intermediate/new values before recording or drawing. Public direct `Trace` calls with invalid data must yield bounded explicit invalid-input states using existing source IDs, not panic/NaN SVG. Preserve old dataset when Build returns error.
- Required implementation verification: per-task tests, `gofmt` changed Go, `go test ./...`, `make build`, and parent browser checks. Docker-backed local Go wrapper may ignore host `GOOS`/`GOARCH`; use root `AGENTS.md` explicit-container build if needed. Browser checks must use freshly embedded assets.

## Evidence and review focus

Read all 18 current `lesson.go`, `trace.go`, `algorithm.go` files below, `internal/trace/types.go`, and approved prototype. Parent supplied Verify graph search with 226 matches, complete pagination, HEAD `49815a34b5cd2f77321f254c54de80d7bbe8e53e`. This worker confirmed project root using `list_projects` and checked all 18 paths: generation `2026-10-27T07:56:32Z`, `metadata_match`, `no_recorded_issue`, best-effort only. Prototype excluded from graph; read directly. Source, not historical docs, governs claims.

Five high-risk review cases, pinned to tests below:
1. Loss measured before update mislabeled as current loss: gradient descent epoch provenance.
2. Iteration limit mislabeled convergence: k-means, perceptron, genetic, Q-learning termination reasons.
3. Duplicate identities or tied values: k-means ties, genetic identical parents, minimax first best child, Q-action ties.
4. Degenerate geometry or malformed data: zero variance, zero weights, ragged vectors, overflow, unsafe target strings.
5. Replay mutation/RNG drift: snapshot isolation and identical fixed-seed traces; no extra random draws for decoration.

Common testing contract for each task: assert every emitted `LineID` exists in that lesson’s `Source()`, repeated trace generation is equal with fresh equal seeds, later state mutation cannot change earlier snapshots, arbitrary mid-trace renderer call works, and SVG parses with `encoding/xml` without `NaN`/`Inf` coordinates. Exercise widths 320, 390, 768, 1440 through parent browser checks; pure string snapshots alone do not prove accessibility.

---

## Task 1: Gradient descent — predictions, loss, two parameter updates

### Existing contract and files

`func Trace(x, y []float64, lr float64, epochs int) []trace.Step`

Read/modify `internal/algorithms/gradientdescent/lesson.go` and `trace.go`; read `algorithm.go` (`Train` is numerical reference). Add `internal/algorithms/gradientdescent/lesson_state_test.go`, `web/renderers/gradientdescent_lesson.go`, `web/renderers/gradientdescent_lesson_test.go`.

Existing JSON object: `{"x":[1,2,3,4],"y":[3,5,7,9],"lr":0.05,"epochs":10}`; arrays `[]float64`, optional `lr *float64`, `epochs *int`. Existing Build accepts paired nonempty equal-length arrays and positive overrides; otherwise silently falls back. Proposed validation: both arrays supplied together, 1–64 equal-length finite samples, `0 < lr <= 1`, 1–100 epochs. Finite large data may still overflow; reject nonfinite predictions, sums, gradients or new parameters with readable Build error. Do not prohibit finite increasing loss; report it.

### Typed state and interface

Define in renderer file:

```go
type GradientDescentLessonState struct {
    TableState
    X, Y, Predictions, Residuals []float64
    Epoch, Sample int // Sample=-1 outside per-sample work
    Phase string // init, predict, residual, accumulate, loss, weight, bias, done, invalid
    W, B, BeforeW, BeforeB, LR float64
    SumSquaredError, SumErrorX, SumError, GradientW, GradientB float64
    LossBefore, LossCurrent float64
    LossBeforeValid, LossCurrentValid bool
    History []GradientDescentEpoch
    StopReason string // epoch-limit, numerical-error, invalid-input
    DataXMin, DataXMax, DataYMin, DataYMax float64
    SliceWMin, SliceWMax, SliceLossMin, SliceLossMax float64
    HistoryLossMin, HistoryLossMax float64
}
type GradientDescentEpoch struct {
    Epoch int
    W, B, Loss float64 // parameters at measurement, not after update
}
func RenderGradientDescentLessonSVG(step trace.Step, width, height int) string
```

No automatic convergence threshold. Optional current loss is a clearly labeled diagnostic evaluation, not another optimizer iteration; cost history retains pre-update definition. Numerical stop must not claim training success.

### Scene and exact source mapping

Reference 760×640: data/model plot `(56,56)-(354,302)`; loss slice `(420,56)-(716,302)`; measured epoch-loss strip `(56,350)-(716,422)`; formula/reason area y=458 onward. Data bounds include samples and current fitted line, with 10% padding and ±1 span for constants. Points retain `P<i>` identity, model line pink, residuals dashed gold. Parameter labels show both `w` and `b`.

Right plot explicitly says **“MSE versus w, b held at [value]”**. Plot `L(u,b)=mean((u*x+b-y)²)` in violet. Tangent gold uses `∂L/∂w`, current point pink, weight target hollow blue. Curve is a slice, not full two-dimensional landscape. At L15 bias changes the whole slice; do not draw a weight-only trail across different slices as if same curve. Minimum green only for analytically known slice minimum `u*=sum(x*(y-b))/sum(x²)` when denominator positive, labeled “slice minimum; b fixed.” For all-zero x, show flat slice/no unique weight minimum. No global minimum promise. On mobile stack both plots and formula blocks.

| IDs | Meaningful phase |
| --- | --- |
| L2 | Parameters zero; loss not yet measured, never show placeholder zero as measured MSE. |
| L4, L5 | Epoch start, reset accumulators. Existing trace omits L5; add real reset snapshot if displayed. |
| L6, L7, L8 | Active sample, actual prediction, actual signed residual. Record within real loop. |
| L9, L10, L11 | Accumulate squared error, error×x, error; show numerical sums, no motion for bookkeeping-only frame. |
| L13 | Divide cost by n; show measured loss and both normalized gradients `(2/n)*sum`. |
| L14 | Weight-only update from old gradients; target/arrow moves 750ms. Bias remains old. |
| L15 | Bias-only update from same old gradients; model translates, slice changes. |
| L17 | Budget complete; show final parameters and separately last pre-update loss/current diagnostic loss. |

Updates are discrete computations. Any arc or continuous movement is explanatory interpolation, not optimizer path or extra samples. Parent should label transition accordingly.

### Formula, example, tests

`pred_i=w*x_i+b`, `e_i=pred_i-y_i`, `L=sum(e_i²)/n`, `g_w=2 sum(e_i*x_i)/n`, `g_b=2 sum(e_i)/n`, `w'=w-lr*g_w`, `b'=b-lr*g_b`.

For default samples at epoch 0: predictions `[0,0,0,0]`, errors `[-3,-5,-7,-9]`, SSE `164`, `L=41`, `sum(e*x)=-70`, `sum(e)=-24`, gradients `(-35,-12)`. At lr=.05: L14 `w=1.75,b=0`; L15 `w=1.75,b=.6`. After this first epoch (not after default ten epochs), predictions `[2.35,4.1,5.85,7.6]`, MSE `1.12875`. At b=0 slice minimum is w=`70/30=7/3`, not zero. First-epoch test must pass `epochs=1` explicitly.

After trace construction compute the domain fields over all recorded finite parameters, predictions, loss history and slice endpoints/minima, with spec padding; copy identical domains to all snapshots. Slice changes when b changes but its axis scales do not. At L14 both typed W and Variables w must contain the updated weight, B remains old, and LossBefore remains explicitly attributed to BeforeW/BeforeB. L15 updates B in state and variables together. Do not relabel old table loss as current loss.

- [ ] Add `TestGradientDescentLessonFirstEpoch`: assert values above within `1e-10`, history first record `(w=0,b=0,loss=41)`, current diagnostic `1.12875` separate from last measured `41`.
- [ ] Add `TestGradientDescentLessonDegenerateAndDivergent`: x `[0,0]`, y `[2,2]`, lr .1, one epoch gives w=0,b=.4, initial loss=4; no divide-by-zero slice minimum. x `[1]`, y `[1]`, lr=1 gives w=2,b=2, current loss=9 versus initial 1; label increase, not convergence. Build rejects mismatched arrays and overflow-producing input.
- [ ] Run `go test ./internal/algorithms/gradientdescent ./web/renderers`; new state/render tests must fail before implementation.
- [ ] Emit typed snapshots in actual loop, correct loss provenance and validation; implement pure SVG renderer with guarded slice geometry.
- [ ] Add renderer assertions for fixed-b caption, measured/current loss labels, target/tangent, no false global minimum, XML validity and immutable replay; rerun targeted tests.
- [ ] Hand parent `gradientdescent` dispatch and weight/bias transition requirements; verify seek at L14 keeps old bias, reduced-motion endpoints, and full baseline checks.

**Copy-ready implementation prompt:**
> Implement Task 1 of `docs/superpowers/plans/2026-10-28-animation-ai.md`, including global constraints. Read gradientdescent lesson.go, trace.go, algorithm.go and approved v5 prototype. Preserve `Trace(x, y []float64, lr float64, epochs int) []trace.Step` and actual two-parameter MSE regression. Add `GradientDescentLessonState` and `RenderGradientDescentLessonSVG` in `web/renderers/gradientdescent_lesson.go`; test epoch-0 loss 41, gradients -35/-12, final w=1.75,b=.6, diagnostic loss 1.12875. Record real source-line phases, never substitute w² simulation or mix pre/post-update loss. Parent owns dispatch/motion; coordinate interface only. Keep themes/layout; run task tests, go test ./..., make build. No commit without authorization.

## Task 2: K-means — distance, membership, centroid mean

### Existing contract and files

`func Trace(points []Point, k, maxIters int) []trace.Step`

Read/modify `internal/algorithms/kmeans/lesson.go`, `trace.go`; read `algorithm.go`. Add `internal/algorithms/kmeans/lesson_state_test.go`, `web/renderers/kmeans_lesson.go`, `web/renderers/kmeans_lesson_test.go`.

JSON: `{"points":[{"x":0,"y":0},{"x":2,"y":0},{"x":10,"y":0},{"x":12,"y":0}],"k":2}`. `LessonInput.Points []Point`, `K int`; no JSON max-iteration field. Existing default k=3 and 12 points from `defaultPoints()`, Build passes `maxIters=20`. Trace clamps k to n and defaults nonpositive maxIters to 100. Proposed Build bounds: 1–64 finite points; 1–64 k, retain effective `min(k,n)` and display requested/effective k. Omitted fields use defaults; explicit empty list/nonpositive k reject. Preserve stride initialization and first-index tie rule.

```go
type KMeansLessonState struct {
    GraphState
    Points, Centroids, PreviousCentroids, Sums [][2]float64
    Assignments, Counts []int
    DistanceSquared []float64 // active point to each current centroid
    Iteration, ActivePoint, Closest, PreviousAssignment, RequestedK int
    Changed, Moved, Converged bool
    Phase, StopReason string // stop: stable, iteration-limit, invalid-input
}
func RenderKMeansLessonSVG(step trace.Step, width, height int) string
```

### Scene and operational phases

Reference 760×600: scatter `(56,50)-(520,380)`, centroid cards `(550,50)-(724,380)`, distance/mean equations y=420 onward. Aspect-preserving scale from data bounds plus initial/current/previous centroids; at least unit extent for constant coordinates. Stable P IDs are small circles; centroids are crosshair diamonds labeled C IDs. Identity color means membership, gold ring means active comparison. Do not move data points when cluster changes. Draw dashed active-point candidate spokes; nearest gets blue outline and squared distance label. Assigned edges remain faint; avoid opaque hulls implying computed Voronoi regions. Centroid update uses old hollow cross and new solid cross connected by blue arrow; cards show sum/count. Mobile cards go below plot, not outside viewBox.

L1 data; L2 stride seeds `points[j*(n/k)]`; L3 assignments -1 (unassigned); L4/L5 iteration/reset; L6 focus point; L7 show all computed squared distances and chosen centroid; L8 compare previous assignment; L9 change membership; L10 mark changed; L13 update each nonempty mean and show empty centroid “retained: no members”; L14 test **both** `!changed && !moved`; L16 report stable or limit. Existing displayed Source L14 only tests changed, while implementation also tests moved; update Source code text at same ID to match. Existing L16 always sets `converged:true`; fix typed state and Variables to actual exit reason. Real centroid computation only; renderer does not rerun clustering.

Formula: `d²(P,C)=(Px-Cx)²+(Py-Cy)²`; `Cj=sum(member points)/count`. For example above, seeds `(0,0),(10,0)`, P1 distances `4,64`, assignments `[0,0,1,1]`, sums `(2,0),(22,0)`, counts `[2,2]`, new means `(1,0),(11,0)`. Movement squared=1 for each; second iteration assignments/centroids unchanged, then stable.

- [ ] Add `TestKMeansLessonMeansAndStop`: example gives means `(1,0),(11,0)`; `maxIters=1` must report limit, not converged; `maxIters=20` stable at iteration 2.
- [ ] Add `TestKMeansLessonTiesAndEmpty`: points `(0,0),(0,0),(2,0)`, k=3 seeds all three; duplicate-centroid tie selects C0, C1 count=0 and stays `(0,0)`. k>n displays effective n. Constant points produce finite scale. Build rejects nonfinite/empty explicit input and excessive counts.
- [ ] Run `go test ./internal/algorithms/kmeans ./web/renderers` to see new tests fail.
- [ ] Snapshot distances before assignment and centroids before update; implement typed renderer; align L14 text and truthful L16 status without changing math.
- [ ] Add renderer tests for P/C identities, unchanged point coordinates through membership changes, sum/count caption, empty-cluster retention, snapshot isolation; rerun targeted tests.
- [ ] Parent verifies centroid-only 750ms motion, assignment strokes, responsive layout, then baseline build/tests.

**Copy-ready implementation prompt:**
> Implement Task 2 and global constraints in `docs/superpowers/plans/2026-10-28-animation-ai.md`. Preserve kmeans `Trace(points []Point, k, maxIters int) []trace.Step`, stride seeds, squared distances, first-index ties, empty-centroid retention. Add `KMeansLessonState` and `RenderKMeansLessonSVG` in `web/renderers/kmeans_lesson.go`. Expose real distances/counts/sums and old/new centroids. Fix unconditional convergence and Source L14 mismatch. Pin example centroids (1,0),(11,0), stable iteration 2, limit at iteration 1. Parent owns dispatch/motion; no theme/layout edits, no dependencies or commit. Run specified checks.

## Task 3: Perceptron — weighted sum, threshold, correction

### Existing contract and files

`func Trace(inputs [][]float64, labels []int, lr float64, epochs int) []trace.Step`

Read/modify `internal/algorithms/perceptron/lesson.go`, `trace.go`; read `algorithm.go`. Add `internal/algorithms/perceptron/lesson_state_test.go`, `web/renderers/perceptron_lesson.go`, `web/renderers/perceptron_lesson_test.go`.

JSON: `{"inputs":[[0,0],[0,1],[1,0],[1,1]],"labels":[0,1,1,1],"lr":0.1,"epochs":10}` (`[][]float64`, `[]int`, `float64`, `int`). Existing accepts nonempty rows/label count match but does not validate rectangularity or binary labels; can panic for ragged rows. Proposed Build bounds: 1–64 samples, 1–8 features with equal row lengths, labels exactly 0/1 and count n, finite features, `0<lr<=1`, 1–100 epochs; reject nonfinite new sums/weights. Features >2 retain valid weighted-sum scene without pretending 2D plot represents full classifier.

```go
type PerceptronLessonState struct {
    ArrayState
    Inputs [][]float64
    Labels []int
    Weights, BeforeWeights, Products []float64
    Bias, BeforeBias, Linear, LR float64
    Epoch, Sample, Target, Prediction, Error, Mistakes int
    PredictionValid, Converged bool
    Phase, StopReason string // stop: stable, epoch-limit, invalid-input, numerical-error
}
func RenderPerceptronLessonSVG(step trace.Step, width, height int) string
```

### Scene and phases

Reference 760×620: left sample plane `(52,56)-(330,330)` when d=2; right neuron flow input chips x=400, sum circle `(550,170)`, threshold box `(630,145)`, output `(710,170)`. Bias enters from above. Product labels `wj*xj` ride links, current sample ring gold, prediction/target chips distinct labeled shapes. Parameter correction panel spans `(40,380)-(720,550)`, shows old+delta=new for each weight and bias; use rows for up to eight features. At d!=2 replace sample plane with selected feature vector and sample list. Mobile stack plane, neuron, correction rows; no narrow horizontal pipeline.

Decision boundary is `w0*x+w1*y+b=0`, clipped to data rectangle. Handle w1=0 as vertical line, both weights zero as “constant prediction”; do not divide by zero. Points use class 0 coral circle, class 1 blue square; gold outer ring focuses current sample, not new class. At L12 only weights change, at L13 bias changes; boundary must reflect exact intermediate state.

L1/L2/L3 init; L4/L5 epoch/reset convergence assumption (not demonstrated convergence); L6 select sample; L7 actual dot sum; L8 threshold `z>=0` predicts 1; L9 `err=target-pred`; L10 branch (correct or incorrect); L11 mark epoch mistake; L12 weight changes; L13 bias change; L16 only error-free epoch proves training-set stability; L18 actual executed epoch count and stop reason. Existing final Variables use requested epochs even after early break; correct reporting. Do not infer prediction before L8 from placeholder zero.

For defaults sample 0 `(0,0)`, initial z=0 predicts 1, target0 gives err=-1, Δweights=(0,0), Δbias=-.1. Sample1 `(0,1)` then z=-.1 predicts0, target1 gives err=1, weights=(0,.1), bias=0. Zero-weight update is meaningful; show zero delta rather than gratuitous movement. Formula `w'_j=w_j+lr*err*x_j`, `b'=b+lr*err`.

- [ ] Add `TestPerceptronLessonThresholdAndCorrection`: assert both sample sequences above, old prediction persists as “prediction before correction”, and L12 retains old bias. One epoch ends `(0,.1),0`; don't claim error-free epoch.
- [ ] Add `TestPerceptronLessonLimitsAndDimensions`: XOR labels `[0,1,1,0]` must report budget exhaustion, not convergence; all-positive one-row dataset at initialization reaches error-free epoch1; mismatched labels/ragged vectors rejected; vertical/zero-weight boundaries finite.
- [ ] Run `go test ./internal/algorithms/perceptron ./web/renderers` before implementation, expect added assertions to fail.
- [ ] Add validated typed snapshots using full float precision, not scaled rounded `ArrayState.Values`; implement neuron/conditional plane and actual epoch count.
- [ ] Test old/new parameter snapshots, XML, stable IDs, d=1/d=3 vector view, unsafe labels handled structurally, and independent mid-trace render; rerun tests.
- [ ] Parent verifies weight then bias transitions, no phantom movement on zero delta, mobile/reduced-motion behavior, and baseline checks.

**Copy-ready implementation prompt:**
> Implement Task 3 and global constraints in `docs/superpowers/plans/2026-10-28-animation-ai.md`. Preserve perceptron Trace signature and z>=0 threshold. Add `PerceptronLessonState` and `RenderPerceptronLessonSVG` in `web/renderers/perceptron_lesson.go`. Teach dot product, binary prediction, signed error and separate L12/L13 corrections; use 2D boundary only for two features. Validate rectangular data/binary labels. Pin first OR samples and XOR nonconvergence, correct executed epoch reporting. Parent owns dispatch/motion; preserve themes/layout. Run task/baseline checks; no commit.

## Task 4: Genetic algorithm — selected parents, actual offspring, elitism

### Existing contract and files

`func Trace(cfg Config) []trace.Step`

Read/modify `internal/algorithms/genetic/lesson.go`, `trace.go`; read `algorithm.go` and modify narrowly only if recording helper results requires it. Add `internal/algorithms/genetic/lesson_state_test.go`, `web/renderers/genetic_lesson.go`, `web/renderers/genetic_lesson_test.go`.

Existing JSON: `{"target":"GENETIC","popSize":8,"maxGenerations":15,"mutationRate":0.08,"tournamentSize":3}`. `target string`; other fields pointers to int/int/float64/int. `Config` additionally has `Charset string` and `RNG *rand.Rand`, **not exposed by LessonInput**. Trace defaults target GENETIC, pop8, mutation .08, generations15, tournament3, charset DefaultCharset, nil RNG seed42. `Evolve` has different defaults and time-based nil seed: never replace Trace with Evolve for playback.

Proposed Build limits: target 1–24 printable ASCII bytes (algorithm operates on bytes, not Unicode genes), pop2–24, generations1–100, mutation `(0,1]` under existing default semantics, tournament1–pop. Explicit empty/non-ASCII/control target rejects. Escape printable `<`, `>`, `&`, quotes. Keep charset implicit. Direct Trace popSize1 currently appends elite plus child and exceeds requested size: add regression and handle capacity before generating/appending offspring; do not claim size1 behavior correct today. For existing valid pop>=2 preserve RNG order.

```go
type GeneticIndividualView struct { ID, Genes string; Fitness int }
type GeneticLessonState struct {
    TableState
    Target string
    Generation, Pair, Cut, BestIndex, MaxGenerations int
    MutationRate float64
    Population, NextPopulation []GeneticIndividualView
    Parents, Crossed, Mutated [2]GeneticIndividualView
    ChangedLoci [2][]int // observed byte changes, not every mutation trial
    EliteID string
    BestFitnessHistory []int
    Phase, StopReason string // stop: target-found, generation-limit, invalid-input
}
func RenderGeneticLessonSVG(step trace.Step, width, height int) string
```

### Scene and phases

Reference 760×680: target gene strip x=48,y=48; population rows `(40,96)-(330,390)`; operation bench `(370,96)-(720,390)`; next-generation slots y=430; best-so-far sparkline and equations y=550. Gene cells width min(28, available/targetLength), wrap long targets into labeled index ranges instead of tiny text. Rows carry stable `g<generation>-i<slot>` identity; elite retains original identity. Parent identity color follows segment origin through crossover; target match uses check underline, not recoloring identity. Mutation observed changes get gold outline and before/after glyph; if sampled replacement equals old byte, do not call it a changed gene. Draw selection arrows to two labeled parent slots, crossover cut vertical dashed line, then blue offspring delivery arrows. Desktop lists may use bounded SVG-local scrolling via parent viewport; mobile stacks bench and population with all candidates reachable.

L1 placeholders, L2 random population, L3 fitness; L4 generation, L5 best, L6 exact-target stop; L7 next population and elite (add actual snapshot; Source text must mention elite); L9 actual parent selection; L10 crossover cut; L11 mutation/append; L13 population replacement; L14 fitness comparison; L16 budget completion. Existing trace shows only first mating pair while later pairs execute without snapshots. Instrument each actual pair at repeated L9/L10/L11 without consuming extra RNG calls. Preserve existing first-pair c2 mutation ordering, even when second child cannot fit; clearly mark unused offspring. Do not invent tournament competitors from winning parent string: only show winner unless actual draws captured by existing helper without added draws. Tied identical chromosomes remain separate IDs. Keep Source L7/L8/L11/L13 text aligned with elite insertion and capacity-limited append.

Formula: `fitness(g)=sum_i [g_i==target_i]`; one-point crossover `c1=p1[:cut]+p2[cut:]`, `c2=p2[:cut]+p1[cut:]`. Example target `1010`, parents `1000` (fitness3), `0111` (fitness1), cut2 gives `1011` (fitness3), `0100` (fitness1). Mutation changing c1 index3 from1 to0 produces target1010, fitness4. This is a labeled operator fixture, not a promised seed42 event. Elitism makes best-so-far nondecreasing, not every child or mean fitness.

- [ ] Add `TestGeneticLessonOperators`: exact crossover/fitness example; `Mutate(Individual{Genes:"1111"},1,"0",freshSeed)` gives `0000`; distinguish unchanged mutation replacement from changed locus.
- [ ] Add `TestGeneticLessonDeterministicLineage`: two fresh seed42 configs equal; every emitted offspring matches captured parent/cut/mutation state; population size equals configured size including odd3 and direct1; elite fitness nondecreasing; no target-found status unless fitness==target byte length.
- [ ] Add `TestGeneticLessonValidationAndEscaping`: unsafe printable target rendered as escaped text, Unicode rejected rather than split invalid bytes, empty explicit target/error, generation exhaustion shown accurately. Direct config Target="1", Charset="0", MaxGenerations=1 cannot solve and must report limit.
- [ ] Run `go test ./internal/algorithms/genetic ./web/renderers` before implementation; then instrument existing operations, capacity fix and typed renderer without new random sampling.
- [ ] Add immutable generation/offspring tests and renderer ID/segment provenance tests; rerun targeted suite.
- [ ] Parent verifies selection/crossover/mutation movement uses 750ms meaningful phases, no duplicate decorative frames, then baseline checks.

**Copy-ready implementation prompt:**
> Implement Task 4/global constraints in `docs/superpowers/plans/2026-10-28-animation-ai.md`. Preserve `genetic.Trace(cfg Config) []trace.Step`, fixed seed42 and existing valid-input random draw order; do not call time-seeded Evolve. Add `GeneticLessonState` and `RenderGeneticLessonSVG` in `web/renderers/genetic_lesson.go`. Capture real parents, cut, offspring, changed loci, elite and every mating pair using L9/L10/L11. Keep candidate IDs distinct even for identical genes. Pin crossover 1000/0111 cut2 =>1011/0100, fixed charset mutation, size1/odd capacity and truthful limit status. Validate ASCII byte model and escape target. Parent owns dispatch/motion; no theme/layout changes, dependencies or commit. Run checks.

## Task 5: Minimax — returned utility and bound-driven pruning

### Existing contract and files

`func Trace(scores []int, branching int) []trace.Step`

Read/modify `internal/algorithms/minimax/lesson.go`, `trace.go`; read `algorithm.go` (`Solve`/`Minimax` numerical references). Add `internal/algorithms/minimax/lesson_state_test.go`, `web/renderers/minimax_lesson.go`, `web/renderers/minimax_lesson_test.go`.

Existing JSON object `{"scores":[-1,4,2,6,-3,-5,0,7],"branching":2}` or bare integer array of scores. Build defaults above; positive branching accepted, `BuildTree` maps <=1 to2, uneven leaf counts split by ceil(len/branching). Proposed Build bounds 1–64 scores, each `[-999999,999999]` (keeps real utilities distinct from current `formatBound` infinity display cutoff), branching2–8. Missing branching=2; explicit1 errors rather than silent coercion. Direct empty Trace remains one terminal utility0 tree, clearly “empty input,” not game strategy.

```go
type MinimaxNodeView struct {
    ID, ParentID string
    Depth, Value int
    IsMax, Terminal, Evaluated, Pruned bool
}
type MinimaxLessonState struct {
    TreeState
    Nodes []MinimaxNodeView
    ActiveID, ChildID, ReturnedFromID, BestChildID string
    Alpha, Beta, Eval, BestValue int
    EvalValid, BestValueValid, Cutoff bool
    Stack []string
    PrunedIDs []string
    Phase string
}
func RenderMinimaxLessonSVG(step trace.Step, width, height int) string
```

Infinity sentinels remain `math.MinInt32`/`math.MaxInt32` from algorithm, but renderer recognizes exact sentinels rather than treating arbitrary large scores as infinity. Node display “?” until evaluated; known input utility leaves may display their provided number with “not visited” status. Internal zero is never an evaluated score by default.

### Scene and phases

Reference 760×640: tree `(36,60)-(724,420)`; active stack/bounds/equation y=450 onward. Leaves equally spaced in original traversal order; internal x midpoint of outer child centers; depth y=`60+depth*85` before fitting vertical space. MAX nodes upward triangles, MIN downward triangles, leaves rounded squares; role labels redundant with shapes. Node radius about20, utility text readable. Large trees need horizontal **local** viewport or subtree detail plus overview, not overlapping 64 leaves; parent coordinates viewport. Mobile show current branch detail with breadcrumb IDs and sibling utility cards, preserving full tree via local viewport.

L1 initial bounds; L2 terminal check; L3 return known leaf; L5/L6 MAX/−∞; L7/L8 selected child/descent; L9 child return and max; L10 alpha; L11 compare `β<=α` and actual prune phase (same ID, distinct Action); L13 MAX return. MIN uses L14/L15, L16/L17, L18 min, L19 beta, L20 compare/prune, L22 return. L24 final root result and chosen child. Animate utility badge along child-to-parent edge at L9/L18, then bound change; prune edges use dashed muted stroke and hatch/cross “not evaluated,” never remove from tree or fabricate their minimax values. Cutoff with zero remaining siblings reports bound satisfied but zero branches pruned.

**Correctness prerequisite:** trace stores internal results only in `tr.values`, but final best-child scan currently compares `child.Val`, which remains zero for internal nodes. Choose best move from actual returned child evaluations captured at root, with first-child tie rule and visited flag; do not compare unvisited placeholder values. Root score/selected move must agree with `Solve`.

Formula: MAX returns max(children), MIN returns min(children); `α=max(α,eval)`, `β=min(β,eval)`; prune remaining siblings only when β<=α. Example scores `[3,5,2,9]`, branching2: left MIN returns3, root α=3; right MIN sees2, β=2<=3, skips9; root result3, chosen child `n0-1`, pruned leaf `n0-2-2`. No claim pruned leaf utility is unknown input: its input score9 is known, its search evaluation was skipped.

- [ ] Add `TestMinimaxLessonReturnAndCutoff`: exact example result3, chosen n0-1, skipped n0-2-2; no L3 visit to skipped leaf; L20 cutoff α3 β2; compare result/move with independent `Solve(BuildTree(...))`.
- [ ] Add `TestMinimaxLessonDefaultMove`: defaults root4, best n0-1; catches current stale `child.Val` bug. Tied scores `[1,1,1,1]` choose first root child, not unknown node. Single score7 returns7 without fake child choice.
- [ ] Add invalid bound/input tests, uneven leaves `[3,5,2]`, deep/local-overflow layout, branch-prune snapshot isolation; run `go test ./internal/algorithms/minimax ./web/renderers` before fixing.
- [ ] Capture typed frame/stack/bounds and root return provenance; fix chosen-child scan; implement role-shaped tree renderer and source-tied badges.
- [ ] Verify XML, no evaluated value on unknown internal nodes, visited/pruned distinctions and exact rendered cutoff numbers; rerun task tests.
- [ ] Parent checks backward seek restores pruned branches and bound badges, local viewport accessibility, cancellation, baseline checks.

**Copy-ready implementation prompt:**
> Implement Task 5/global constraints in `docs/superpowers/plans/2026-10-28-animation-ai.md`. Preserve minimax Trace signature, deterministic tree shape and alpha-beta evaluation. Add `MinimaxLessonState` and `RenderMinimaxLessonSVG` in `web/renderers/minimax_lesson.go`. Capture node roles, evaluated status, stack, return values, bounds and pruned IDs. Fix final move selection using actual root child returns, not stale child.Val. Pin [3,5,2,9] =>3/n0-1 and prune n0-2-2; default root4/n0-1. Never compute skipped subtrees for decoration. Parent owns dispatch/motion; preserve themes/layout. Run task/baseline checks; no commit.

## Task 6: Q-learning — transition, TD target, table update

### Existing contract and files

```go
func Trace(env *GridWorld, episodes int, alpha, gamma, epsilon float64) []trace.Step
func TraceWithSeed(env *GridWorld, episodes int, alpha, gamma, epsilon float64, seed int64) []trace.Step
```

Read/modify `internal/algorithms/qlearning/lesson.go`, `trace.go`; read `algorithm.go`. Add `internal/algorithms/qlearning/lesson_state_test.go`, `web/renderers/qlearning_lesson.go`, `web/renderers/qlearning_lesson_test.go`. No need to change action selection just to label explore/exploit: use “epsilon-greedy selected action” unless actual branch captured without extra RNG.

JSON `{"episodes":5,"alpha":0.5,"gamma":0.9,"epsilon":0.3}`: int and three float64 fields, `omitempty`. No custom environment or seed in current lesson schema. Default grid3×3, start(0,0), goal(2,2), obstacle(1,1). Trace delegates seed7, clamps episodes20, defaults nonpositive episodes4/alpha.5/gamma.9/epsilon.2; Build defaults differ (episodes5, epsilon.3). Per-episode cap12 steps. Proposed Build validation episodes1–20, alpha `(0,1]`, gamma `(0,1]`, epsilon `(0,1]`, retaining existing zero-as-default semantics only for omitted fields; explicit zero errors with useful message rather than silently adopting different effective value. Supporting zero discount/exploration later requires deliberately changing Trace normalization and tests, not silently claiming zero works now. Validate direct env dimensions/start/goal/obstacles before indexing; max64 cells for lesson scene.

```go
type QLearningLessonState struct {
    TableState
    Rows, Cols, Episode, EpisodeStep, Episodes int
    Start, Goal, State, NextState [2]int
    Obstacles [][2]int // sorted row-major
    Q [][]float64
    Action, BestNextAction int // -1 when absent
    Alpha, Gamma, Epsilon, Reward, BestNextQ, TDTarget, OldQ, NewQ float64
    TransitionValid, UpdateValid, Terminal bool
    Phase, EpisodeStopReason string // goal, step-limit, or empty while running
    CompletedEpisodes int
}
func RenderQLearningLessonSVG(step trace.Step, width, height int) string
```

### Scene and phases

Reference 760×650: environment grid `(48,64)-(348,364)`, selected state/action values `(400,64)-(716,260)`, next-state values `(400,280)-(716,390)`, Bellman equation `(40,440)-(720,560)`, episode status below. Grid cell side=min(300/cols,300/rows); fixed row-major geometry, origin upper-left, action order Up0/Right1/Down2/Left3. Start outlined square “S”, goal green star “G”, obstacle hatched “blocked”, agent pink ring. Show all greedy tied arrows as small labeled candidate arrows; policy label “current Q estimates,” never guaranteed optimal route. Selected action gold, actual transition blue; blocked moves keep agent at same coordinate and show collision mark. Side panel keeps exact old/new Q and reward; full Q table can stay in embedded state but selected/next rows are primary teaching scene.

L2 zero table; L3 episode; L4 reset agent without resetting Q; L5 terminal/nonterminal; L6 selected action; L7 actual environment transition/reward; L8 future maximum, forced0 at terminal; L9 TD target; L10 single cell update; L11 state assignment; L5 goal reached; L14 training budget exhausted. L7 may show hollow next-position marker and transition arrow while current state remains old until L11. Avoid moving agent twice for same transition. Add capped-episode notice at existing loop-end L12 (actual 12-step limit, update displayed Source to include cap) if goal not reached. L14 says “episodes completed; estimates learned,” not convergence/optimality. Existing Source loop lacks cap and terminal target handling; align text at L5/L8 without renumbering IDs. No fabricated reward for collision: real reward is zero.

Formula: `target=r+γ max_a Q(s',a)`, terminal future=0; `Qnew=Qold+α(target-Qold)`. From state(2,1), actionRight reaches goal(2,2), reward10; alpha.5, Qold0 => target10, newQ5. Nonterminal example next maximum5, reward0, gamma.9 => target4.5; old0 => new2.25. These are direct agent fixtures, not claims that every seed7 trace reaches goal. A short run with all-zero Q is legitimate.

- [ ] Add `TestQLearningLessonBellmanNumbers`: `NewQAgent(DefaultGridWorld(),.5,.9,.3)` goal transition gives next(2,2), reward10, done=true; update gives5,target10 even if terminal row manually set high. Seed next-state maximum5 for nonterminal update and assert2.25,target4.5.
- [ ] Add `TestQLearningLessonCollisionsAndTies`: Up at start remains(0,0),reward0,done=false; Right from(1,0) hits obstacle and remains; zero row `BestActions` returns `[0,1,2,3]`, `MaxQ` uses Up. Renderer must not present arbitrary tie as learned unique route.
- [ ] Add `TestQLearningLessonSeedAndBudget`: two fresh seed7 traces equal; no episode has >12 L7 transitions; never-reaching-goal valid custom env shows step-limit; goal episodes show goal, not cap; Q persists across L4 resets. Nil env fallback valid; invalid dimensions/coordinates rejected safely. Build rejects invalid rates and episode bounds.
- [ ] Run `go test ./internal/algorithms/qlearning ./web/renderers` before implementation; then add typed snapshots, terminal/cap source accuracy and environment/Q SVG, preserving RNG draws.
- [ ] Test immutable nested Q rows, obstacle order, independent mid-trace rendering, terminal target0 future and XML; rerun task suite.
- [ ] Parent verifies L7 preview versus L11 actual state, L10 cell update, reduced-motion reset and mobile grid; run baseline checks.

**Copy-ready implementation prompt:**
> Implement Task 6/global constraints in `docs/superpowers/plans/2026-10-28-animation-ai.md`. Preserve Q-learning Trace/TraceWithSeed signatures, seed7, real grid reward0/goal10, terminal future0 and 12-step cap. Add `QLearningLessonState` and `RenderQLearningLessonSVG` in `web/renderers/qlearning_lesson.go`. Show actual transition, next-Q maximum, target and one-cell Bellman update; never invent rewards, optimal route or exploration mode. Pin terminal update5/target10 and nonterminal2.25/target4.5. Align Source with cap/terminal handling, validate inputs and deep-copy Q. Parent owns dispatch/motion; preserve themes/layout, no dependencies or commit. Run specified checks.

## Integration gates and blockers

- [ ] Parent freezes root shared SVG sizing/stable-key/transition contract before product integration. This companion consumes that contract; workers do not create competing motion APIs.
- [ ] Parent dispatch includes all six exact renderer names and new state types; ordinary metadata renderer strings alone will not distinguish these lesson scenes.
- [ ] Parent tests play/pause/play, pause before timer, navigation/data dialog during motion, seek/back/reset, end-of-trace, reduced motion, and current theme contrast. Ensure SVG has accessible name/description and formulas have text equivalents; no SVG-only interactive controls needed.
- [ ] Product executor records targeted tests, `go test ./...`, `make build`, browser checks and any wrapper/tool blockers; no runtime success claim from this document.

Known implementation prerequisites: existing Build methods silently accept/fallback on bad data; validation changes above are deliberate. Gradient loss provenance, k-means unconditional convergence, perceptron final epoch reporting, genetic size1/hidden mating operations, minimax stale best-move values, and Q cap/terminal-source discrepancies must be resolved or explicitly blocked before claiming faithful animations. No product fixes were made while drafting this plan.
