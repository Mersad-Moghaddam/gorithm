# Algorithm animation design: beginner-friendly motion

## Approved direction

User approved colorful Bubble Sort and gradient-descent interactive studies, requested equivalent algorithm-specific treatment for every lesson, and explicitly excluded redesigning the project shell. Earlier slow, four-stage-per-comparison concepts were rejected. Use normal, interesting motion with short explanations, not compulsory reading pauses.

Durable visual references (open these local HTML files in a browser):

- [Approved colorful sorting study](2026-10-28-animation-references/bubble-sort.html).
- [Approved gradient-descent study](2026-10-28-animation-references/gradient-descent.html).

These are throwaway demonstrations, not production implementations. Their independent JavaScript simulation, global styles, controls, fixed datasets, and example stopping rules must not enter the app. Match their visual grammar and rhythm inside the existing workspace. Algorithm-specific mathematics must come from current Go source. In particular, do not assume the gradient-descent lesson optimizes the prototype's `w²` function.

## Scope and ownership

All 40 registered lessons receive an individual implementation worker and review. One shared owner integrates playback, SVG movement, renderer dispatch, legend routing, and narrowly scoped stage styling. Algorithm workers own only their algorithm package and dedicated renderer/tests. Every task must explain the algorithm's distinctive operation; recoloring the old generic renderer is insufficient.

Keep existing navigation, themes, library, typography, source inspector, dataset dialog, and transport placement. New auxiliary diagrams belong inside the visualization. Use the current custom-input dialog for parameters. No framework, animation dependency, generic scene engine, or separate JavaScript algorithm implementation.

## Visual system

- Keep current dark visualization canvas, including Warm Paper. Reuse shell theme tokens outside SVG.
- Stable entity fills, in original identity order: coral `#c45d75`, blue `#387faa`, violet `#8763bb`, amber `#b67b30`, teal `#258b80`; repeat colors with visible IDs when more than five entities exist.
- State is independent of entity identity: comparison outline `#ffe29a`, update/reference arrow `#8bd5ff`, finalized outline and check `#9bf0bb`, conflict cross/dash `#ff9b9b`, neutral text `#eff5ed`, supporting text `#bbcab9`.
- AI curve grammar: curve `#aa92e7`, current dot `#ee8b9b`, tangent `#ffe29a`, proposed update `#8bd5ff`, actual minimum (only when known) `#9bf0bb`.
- No color-only meaning. Labels, numbered identities, arrows, boundary brackets, checked results, and conflict marks carry the same information.
- Entity colors do not change during swap, insert, rotation, or transfer. Equal values have different stable IDs. Cluster membership is an explicit exception: point fill denotes cluster, point ID still persists.
- Use 800×450 base geometry where practical, consistent margins of 32 SVG units and at least 16 units between independent labels. Custom heights are permitted when information cannot fit. At 320px viewport, preserve readable labels with stage-local scrolling or responsive alternate layout; do not shrink a large matrix into illegible text.
- Text must render at least 14 CSS px for essential labels and 12 CSS px for secondary indices. Actual rendered pixels, not merely SVG `font-size`, determine acceptance. Fit small diagrams; locally scroll large diagrams with accessible region name and keyboard focus.

## Rhythm and explanations

- Default meaningful operation: approximately 1800 ms at 1×. Normal movement 750 ms; compare-only changes 200 ms. Keep current speed multiplier options.
- Compare/commit pairs share a budget: compare 600 ms, commit/result 1200 ms. Single decision without mutation: 1800 ms. Setup/control-flow snapshots: 350 ms only when explanation is short and not introducing a new concept. Never skip stored snapshots or fabricate microsteps for cosmetic lifting.
- Motion completes before another step begins. Clamp movement to at most 60% of the current step's hold time. No autoplay on load.
- Previous, seek, reset, or navigation settles immediately to authoritative stored state. Pause settles the current motion to its selected snapshot and cancels all future advancement. Resume starts from that snapshot.
- Each explanation answers “what changed, why, what next” in one or two short sentences, with actual values. Formula panels use substituted operands. Do not narrate “execute line 12.”
- Reduced motion uses immediate state changes; explanations, path markers, colors, and controls remain available.

## Truth and determinism

All state, metrics, source highlighting, and explanations belong to the selected `trace.Step`. Motion is a view of the transition ending at that step; it cannot increment counters or mutate algorithm data. Source line means the operation producing the snapshot. Planned next values must be explicitly labeled proposals, never already-completed work.

Deep-copy mutable snapshot fields. Seeds, tie breaks, entity IDs, and ordering must be deterministic for fixed input. Preserve algorithm semantics, termination criteria, accepted schemas, and validation. Report discovered discrepancies rather than silently replacing algorithms with textbook variants.

For numerical plots, fix axes over the built trace, include zero where relevant, pad finite domains by 10%, handle constant ranges with a nonzero fallback, and reject/label non-finite data. Never clip diverging updates into apparent convergence. Renderers escape all dataset-derived text and attribute values.

## Acceptance

Each lesson needs its own trace invariant test, renderer evidence test, and browser scenario showing the distinctive operation. Full acceptance includes 320/390/768/1440px, all eight themes, reduced motion, controls during active motion, source synchronization, valid custom input, rejected invalid input, and completed/failure outcomes. Check SVG geometry and labels, not screenshots alone. Baseline: `go test ./...`, `go test -race ./...`, `make build`, current UI checks, accessibility audit, and the new animation regression script.

Implementation details and individual worker prompts live in [implementation plan](../plans/2026-10-28-algorithm-animations.md) and its five linked algorithm task books. These detailed adaptations are proposed implementation decisions extending the two approved studies; user reviews the completed plan before implementation.
