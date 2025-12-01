package main

import (
	"fmt"
	"github.com/pedrohff/defi/components"
	"github.com/pedrohff/defi/view"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type testCaseOption struct {
	name             string
	status           string
	compileSuccess   bool
	assertionSuccess bool
	inputs           []string
	expectedOutput   string
	actualOutput     string
}

type model struct {
	testCases []testCaseOption // the list of choices
	cursor    int              // which to-do list item our cursor is pointing at
	selected  map[int]struct{} // which to-do items are selected
	width     int
	height    int
}

func initialModel() model {
	return model{
		// Sample test cases covering various states
		testCases: []testCaseOption{
			{"Sample Test Case 1", components.TestCasePending, false, false, []string{"1", "2"}, "3", ""},
			{"Sample Test Case 2", components.TestCasePending, true, false, []string{"5"}, "10", ""},
			{"Sample Test Case 3", components.TestCasePending, true, true, []string{"3", "4"}, "7", ""},
			{"Sample Test Case 4", components.TestCaseRunning, false, false, []string{"10"}, "20", ""},
			{"Sample Test Case 5", components.TestCaseRunning, true, false, []string{"2", "3"}, "5", ""},
			{"Sample Test Case 6", components.TestCaseRunning, true, true, []string{"1"}, "1", ""},
			{"Sample Test Case 7", components.TestCaseFinished, false, false, []string{"8"}, "16", "error: compile failed"},
			{"Sample Test Case 8", components.TestCaseFinished, true, false, []string{"4", "5"}, "9", "10"},
			{"Sample Test Case 9", components.TestCaseFinished, true, true, []string{"6", "7"}, "13", "13"},
		},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
		cursor:   0,
	}
}
func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.testCases)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "esc":
			m.cursor = -1
		}

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// Convert local testCaseOption to view.TestCaseData
	testCases := make([]view.TestCaseData, len(m.testCases))
	for i, tc := range m.testCases {
		testCases[i] = view.TestCaseData{
			Name:             tc.name,
			Status:           tc.status,
			CompileSuccess:   tc.compileSuccess,
			AssertionSuccess: tc.assertionSuccess,
			Inputs:           tc.inputs,
			ExpectedOutput:   tc.expectedOutput,
			ActualOutput:     tc.actualOutput,
		}
	}

	// Build and render the main view
	mainView := view.NewMainView(m.width, m.height, testCases,
		view.WithSelectedIndex(m.cursor),
		view.WithFilename("somebiglongfilenamefortesting.cpp"),
		view.WithLanguage("cpp"),
		view.WithStatus("All systems operational"),
	)

	return mainView.Render()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
