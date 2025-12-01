// Package view provides composite layouts built from components.
package view

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/pedrohff/defi/components"
)

// TestCaseData holds the information needed to render a single test case row
// and its optional detail pane.
type TestCaseData struct {
	Name             string
	Status           string
	CompileSuccess   bool
	AssertionSuccess bool
	Inputs           []string
	ExpectedOutput   string
	ActualOutput     string
}

// MainView encapsulates everything required to render the primary Défi screen.
type MainView struct {
	Width         int
	Height        int
	TestCases     []TestCaseData
	SelectedIndex int // -1 means nothing selected
	Filename      string
	Language      string
	Status        string
}

// MainViewOption defines a functional option for configuring MainView.
type MainViewOption func(*MainView)

// WithSelectedIndex sets the currently selected test case index.
func WithSelectedIndex(index int) MainViewOption {
	return func(v *MainView) {
		v.SelectedIndex = index
	}
}

// WithFilename sets the filename displayed in the footer.
func WithFilename(filename string) MainViewOption {
	return func(v *MainView) {
		v.Filename = filename
	}
}

// WithLanguage sets the language label displayed in the footer.
func WithLanguage(language string) MainViewOption {
	return func(v *MainView) {
		v.Language = language
	}
}

// WithStatus sets the status message displayed in the footer.
func WithStatus(status string) MainViewOption {
	return func(v *MainView) {
		v.Status = status
	}
}

// NewMainView constructs a MainView with required parameters and optional configuration.
// Required: width, height, testCases. Optional fields can be set via functional options.
func NewMainView(width, height int, testCases []TestCaseData, opts ...MainViewOption) *MainView {
	v := &MainView{
		Width:         width,
		Height:        height,
		TestCases:     testCases,
		SelectedIndex: -1,
		Language:      "-",
		Filename:      "-",
		Status:        "Idle",
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

// Render composes the header, test case list, optional details pane, and footer.
func (v *MainView) Render() string {
	header := components.Header(v.Width, " Défi")

	// Build test case rows
	rows := []string{components.TestCaseHeader(v.Width)}
	for i, tc := range v.TestCases {
		focused := i == v.SelectedIndex
		row := components.TestCase(
			v.Width,
			tc.Name,
			tc.Status,
			tc.CompileSuccess,
			tc.AssertionSuccess,
			focused,
		)
		rows = append(rows, row)
	}
	testList := lipgloss.JoinVertical(lipgloss.Top, rows...)

	footer := components.Footer(v.Width, v.Status, v.Language, v.Filename)

	// Compute remaining vertical space for details pane
	headerHeight := 2
	listHeight := len(v.TestCases) + 1 // +1 for header row
	footerHeight := 1
	leftover := v.Height - headerHeight - listHeight - footerHeight
	if leftover < 0 {
		leftover = 0
	}

	// Build details pane if a test is selected
	var details string
	testCaseDetailsWidth := v.Width / 2
	if v.Width < 80 {
		testCaseDetailsWidth = v.Width - 10
	}
	if v.SelectedIndex >= 0 && v.SelectedIndex < len(v.TestCases) {
		tc := v.TestCases[v.SelectedIndex]
		details = components.TestCaseDetails(
			testCaseDetailsWidth,
			leftover-4,
			tc.Name,
			tc.Inputs,
			tc.ExpectedOutput,
			tc.ActualOutput,
		)
	}
	detailsPane := lipgloss.PlaceVertical(leftover, lipgloss.Center, details)

	return lipgloss.JoinVertical(
		lipgloss.Center,
		header,
		testList,
		detailsPane,
		footer,
	)
}
