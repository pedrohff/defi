package main

import (
	"fmt"

	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/pedrohff/defi/components"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string         // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
	width    int
	height   int
}

func initialModel() model {
	return model{
		// Our to-do list is a grocery list
		choices: []string{"Buy carrots", "Buy celery", "Buy kohlrabi"},

		// A map which indicates which choices are selected. We're using
		// the  map like a mathematical set. The keys refer to the indexes
		// of the `choices` slice, above.
		selected: make(map[int]struct{}),
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
			if m.cursor < len(m.choices)-1 {
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
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {

	var styleHeader = lipgloss.NewStyle().
		Foreground(components.ColorWhite).
		Background(components.ColorSuccess).
		Width(m.width).
		Align(lipgloss.Center).Bold(true).Italic(true)

	var styleStatusTitle = lipgloss.NewStyle().
		Foreground(components.ColorWhite).
		Background(components.ColorAlert).
		Align(lipgloss.Center)

	var styleStatusDesc = lipgloss.NewStyle().
		Foreground(components.ColorMutedAccent).
		Background(components.ColorSurfaceDark)

	var (
		headerContent = styleHeader.Render("Défi")
		mainContent   = ""
		footerContent = lipgloss.JoinHorizontal(lipgloss.Left,
			styleStatusTitle.Padding(0, 1).Render("STATUS"),
			styleStatusDesc.Padding(0, 1).Width(m.width-8).Render("All systems operational"))
	)
	_ = footerContent

	mainContent = lipgloss.JoinVertical(
		lipgloss.Top,
		components.TestCaseHeader(m.width),
		components.TestCase(m.width, "Sample Test Case 1", components.TestCasePending, false, false),
		components.TestCase(m.width, "Sample Test Case 2", components.TestCasePending, true, false),
		components.TestCase(m.width, "Sample Test Case 3", components.TestCasePending, true, true),
		components.TestCase(m.width, "Sample Test Case 4", components.TestCaseRunning, false, false),
		components.TestCase(m.width, "Sample Test Case 5", components.TestCaseRunning, true, false),
		components.TestCase(m.width, "Sample Test Case 6", components.TestCaseRunning, true, true),
		components.TestCase(m.width, "Sample Test Case 7", components.TestCaseFinished, false, false),
		components.TestCase(m.width, "Sample Test Case 8", components.TestCaseFinished, true, false),
		components.TestCase(m.width, "Sample Test Case 9", components.TestCaseFinished, true, true),
	)
	// Send the UI for rendering
	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.PlaceHorizontal(m.width, lipgloss.Center, headerContent),
		lipgloss.PlaceVertical(m.height-2, lipgloss.Left, mainContent),
		// lipgloss.PlaceHorizontal(m.width, lipgloss.Left, footerContent),
		components.Footer(m.width, "lets check", "cpp", "somebiglongfilenamefortesting.cpp"),
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func mai2n() {
	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(components.ColorDemoWhite).
		Background(components.ColorDemoPurple).
		PaddingTop(2).
		PaddingLeft(4).
		Width(22)
	_ = style
	fmt.Println(style.Render("Hello, kitty"))

	var stylenew = lipgloss.NewStyle().
		SetString("What’s for lunch?").
		Width(24).
		Height(3).
		Foreground(components.ColorDemoAccent)
	_ = stylenew
	// fmt.Println(stylenew.Render())

	var styleStatusTitle = lipgloss.NewStyle().
		Foreground(components.ColorWhite).
		Background(components.ColorAlert)
	_ = styleStatusTitle

	var styleStatusDesc = lipgloss.NewStyle().
		Foreground(components.ColorMutedAccent).
		Background(components.ColorSurfaceDark)

	output := lipgloss.JoinHorizontal(lipgloss.Top,
		styleStatusTitle.Render(" STATUS "),
		styleStatusDesc.Render(" All systems operational "),
	)
	fmt.Println(output)
}
