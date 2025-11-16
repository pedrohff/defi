# Agent Guide

This document orients automation agents and copilots working on Défi.

## Development Workflow

- **Formatting:** run `gofmt` on any touched Go files before delivering changes.
- **Validation:** execute `go test ./...` after modifications to ensure the workflow and parser stay healthy.
- **Terminal UI:** the Bubble Tea model lives in `tea_model.go`. Keep UI logic there and prefer helper functions over inline string manipulation inside `main.go`.
- **Watch mode:** `watcher.go` emits an initial event. The model uses the `ignoreInitialWatcher` flag to avoid duplicate runs—preserve this behaviour when editing watcher or update logic.

## Styling Guidelines

- Use the shared Lip Gloss palette in `components/styles.go`. If a new color is required, add a named constant there and reference it elsewhere.
- Reuse existing components (`Header`, `Footer`, `TestCase`) instead of re-implementing layouts. Extend the component files only when the adjustment benefits the entire UI.

## Prompt & Workflow Notes

- Test inputs come from `/*defiprompt … */` blocks. Update `parser.go` with care; it’s covered by `parser_test.go`.
- Compilation and execution helpers reside in `workflow.go`. When adding languages, wire them through the `supportedLanguages` map and update documentation in `README.md`.

## Communication

- Document any UX-facing change in `README.md` and drop screenshots under `docs/assets/` (see the “UI overview & screenshots” section for placement).
- Surface any non-trivial trade-offs in pull request descriptions so human reviewers can follow the reasoning quickly.
