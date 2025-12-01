# Agent Guide

This document orients automation agents and copilots working on Défi.

## Development Workflow

- **Formatting:** run `gofmt` on any touched Go files before delivering changes.
- **Validation:** execute `go test ./...` after modifications to ensure the workflow and parser stay healthy.
- **Terminal UI:** the Bubble Tea model lives in `tea_model.go`. Keep UI logic there and prefer helper functions over inline string manipulation inside `main.go`.
- **Watch mode:** `watcher.go` emits an initial event. The model uses the `ignoreInitialWatcher` flag to avoid duplicate runs—preserve this behaviour when editing watcher or update logic.

## Pre-Commit Checklist

Before committing changes, agents should verify:

1. **Unused code:** Search for orphaned functions, types, or struct fields that are written but never read. Removing a field may leave helper functions unused.
2. **Architecture compliance:** Confirm changes respect the boundaries in the "Architecture Boundaries" section (e.g., no `lipgloss` in `tea_model.go`, no duplicated types).
3. **Formatting:** Run `gofmt -l .` and format any listed files.
4. **Static analysis:** Run `go vet ./...` to catch common issues.
5. **Tests:** Run `go test ./...` to ensure nothing is broken.
6. **Documentation:** Review the conversation for any new patterns, guidelines, guardrails, or architecture decisions discussed—ask if they should be added to this file.

## Styling Guidelines

- Use the shared Lip Gloss palette in `components/styles.go`. If a new color is required, add a named constant there and reference it elsewhere.
- Reuse existing components (`Header`, `Footer`, `TestCase`) instead of re-implementing layouts. Extend the component files only when the adjustment benefits the entire UI.

## Prompt & Workflow Notes

- Test inputs come from `/*defiprompt … */` blocks. Update `parser.go` with care; it’s covered by `parser_test.go`.
- Compilation and execution helpers reside in `workflow.go`. When adding languages, wire them through the `supportedLanguages` map and update documentation in `README.md`.

## Communication

- Document any UX-facing change in `README.md` and drop screenshots under `docs/assets/` (see the "UI overview & screenshots" section for placement).
- Surface any non-trivial trade-offs in pull request descriptions so human reviewers can follow the reasoning quickly.

## Architecture Boundaries

### tea_model.go
- Contains the Bubble Tea `Model` implementation: state, `Init()`, `Update()`, and `View()`.
- **Must not import `lipgloss`** — styling is a view concern. The model handles state transitions and message routing; visual rendering belongs in `components/` and `view/`.
- **Rationale:** Keeping styling out of the model makes it easier to test state logic in isolation and ensures styling changes don't require modifications to core application logic.
- Delegate all rendering to `view.MainView` and use helpers from `components/` (e.g., `NewSpinner()`, `RenderError()`) for any styled output.
- Avoid duplicating types that exist in `view/`; use `view.TestCaseData` directly for test case state.

### view/
- Composite layouts that combine multiple components.
- Responsible for layout, positioning, and calling into `components/`.

### components/
- Reusable UI primitives (`Header`, `Footer`, `TestCase`, etc.).
- All colors and shared styles live in `styles.go`; add new helpers here when the model needs styled output.
