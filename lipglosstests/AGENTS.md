# lipglosstests Agent Guide

This package serves as a **playground** for experimenting with Lip Gloss styles and Bubble Tea UI patterns before integrating them into the main Défi application.

## Purpose

- Prototype new components or view compositions in isolation.
- Validate styling tweaks without impacting the production TUI.
- Exercise the `view` package to ensure composite layouts render correctly.

## Usage

Run the playground binary to preview UI experiments:

```bash
go run ./lipglosstests
```

Use keyboard navigation (arrows, `j`/`k`) to interact with sample test cases and observe styling behaviour.

## Guidelines for Copilots

1. **Test new views here first.** When developing features for the `view` package, import and exercise them inside `lipglosstests/main.go` before wiring them into `tea_model.go`.
2. **Keep sample data representative.** The `testCaseOption` slice should cover edge cases (pending, running, finished; pass/fail combinations).
3. **Do not import production runtime.** Avoid pulling in workflow, watcher, or config logic—this package is purely visual.
4. **Respect component boundaries.** Reuse widgets from `components/` and composites from `view/`. Do not duplicate styling here.
