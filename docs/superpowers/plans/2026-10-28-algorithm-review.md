# Algorithm Scene Review Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development. One worker per algorithm; coordinator integrates shared changes.

**Goal:** Correct algorithm presentation and improve readable, accurate teaching animations across all 40 lessons.

**Architecture:** Workers review their assigned algorithm's trace and dedicated renderer independently, making narrowly scoped fixes with regression tests. Coordinator reviews findings, resolves shared UI and motion issues, and verifies the integrated application.

**Tech Stack:** Go, SVG, Go/WASM, native Web Animations API, Playwright.

**Spec:** Approved scope in conversation; existing `docs/superpowers/specs/2026-10-28-algorithm-animations-design.md` supplies existing motion and scene contracts.

## Global Constraints

- Preserve eight themes, shell, controls, navigation, accessibility, custom inputs, deterministic traces and source alignment.
- No new dependencies, frontend framework, or JavaScript algorithm implementation.
- Workers own only `internal/algorithms/<id>/` and `web/renderers/<id>_lesson*.go`; shared files and commits belong to coordinator.
- Do not invent trace states or animations unsupported by algorithm data. Preserve static and reduced-motion results.

## Review Focus

- Long and duplicate values: maintain identities and readable, non-overlapping labels.
- Boundary and invalid custom inputs: preserve validation and correct trace results.
- Narrow viewport and dense scenes: allow local scrolling without clipping essential information.
- Reverse, seek, pause, and reduced motion: display the selected snapshot without stale animation.
- AI numerical inputs and final outcomes: show actual values and decisions, not fabricated motion.

## Tasks: one per catalog algorithm

- [ ] Assign each catalog ID to one worker; supply exact owned paths and existing design contract.
- [ ] Review current trace, source mapping, renderer, motion geometry, and default/custom fixtures; record concrete findings.
- [ ] For each confirmed defect, add a focused regression test, verify failure, implement the smallest correction, and run focused algorithm and renderer tests.
- [ ] Inspect representative operation and final scenes at 320px and desktop; report browser limitations and shared-file issues instead of touching shared files.
- [ ] Report files, evidence, tests, remaining risks, and precise review findings. Workers do not commit or push.

## Integration and delivery

- [ ] Review every worker's diff and findings; integrate only verified changes, resolve shared issues and run `gofmt` on changed Go files.
- [ ] Run `go test ./...`, `go test -race ./...`, `make build` or the documented explicit-container equivalent, and `git diff --check`.
- [ ] Run freshly built application through `scripts/check_ui.mjs`, `scripts/check_animations.mjs`, and `scripts/audit_ui.mjs`; inspect representative scenes, including AI lessons, at 320/390/768/1440px and across eight themes. Record unverified cases honestly.
- [ ] Review staged paths; commit and push reviewed changes to existing `origin/feat/teaching-animations` without force. Report commit, PR, commands, and limitations.
