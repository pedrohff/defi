package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	cfg, initialPath, err := parseAppConfig(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, usageMessage)
		os.Exit(1)
	}

	if cfg.once && initialPath == "" {
		fmt.Fprintln(os.Stderr, "no file to run")
		os.Exit(1)
	}

	m := newModel(cfg, initialPath)
	program := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := program.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fm, ok := finalModel.(model)
	if !ok {
		fmt.Fprintln(os.Stderr, "unexpected program state")
		os.Exit(1)
	}

	if fm.summaryErr != nil {
		fmt.Printf("🚨 Tests passed: %d/%d\n", fm.summaryPassed, fm.summaryTotal)
		fmt.Fprintln(os.Stderr, fm.summaryErr)
		os.Exit(1)
	}

	if fm.summaryTotal > 0 {
		fmt.Printf("🎉 Tests passed: %d/%d\n", fm.summaryPassed, fm.summaryTotal)
	}
}
