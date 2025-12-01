package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// detailsSectionTitle styles the label above each detail section.
	detailsSectionTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorAccentBlue).
				MarginBottom(1)

	// detailsContent styles the content block of each section.
	detailsContent = lipgloss.NewStyle().
			Foreground(ColorTextPrimary).
			Background(ColorSurfaceDark).
			Padding(0, 2)

	// detailsContainer wraps the entire details pane.
	detailsContainer = lipgloss.NewStyle().
				Padding(1, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ColorTextMuted)
)

// TestCaseDetails renders a details pane showing test inputs, expected output,
// and actual execution output side by side (or stacked if width is limited).
// inspiration https://www.gh-dash.dev
func TestCaseDetails(width int, height int, name string, testInputs []string, expectedOutput string, executionOutput string) string {

	detailsContainer = detailsContainer.Height(height)
	// Section widths
	sectionWidth := (width - 8) / 3
	if sectionWidth < 20 {
		sectionWidth = 20
	}

	_ = sectionWidth
	halfSectionWidth := (width - 2) / 2

	borderLeft := lipgloss.NewStyle().BorderLeft(true).BorderForeground(ColorTextMuted)

	// Build inputs section
	inputsLabel := detailsSectionTitle.Render("󱋴 INPUTS")
	inputsBody := detailsContent.Width(halfSectionWidth).Height(len(testInputs)).Render(strings.Join(testInputs, "\n"))
	inputsSection := borderLeft.Render(lipgloss.JoinVertical(lipgloss.Left, inputsLabel, inputsBody))

	// Build expected output section
	expectedLabel := detailsSectionTitle.Render("󱋲 EXPECTED")
	expectedBody := detailsContent.Width(halfSectionWidth).Height(len(testInputs)).Render(expectedOutput)
	expectedSection := lipgloss.JoinVertical(lipgloss.Left, expectedLabel, expectedBody)

	// Build actual output section
	actualLabel := detailsSectionTitle.Render(" OUTPUT")
	actualBodyStyle := detailsContent.Width(width)
	actualBody := ""
	if executionOutput == "" {
		actualBody = actualBodyStyle.Foreground(ColorTextMuted).Italic(true).Render("empty")
	} else {
		actualBody = actualBodyStyle.Render(executionOutput)
	}
	actualSection := lipgloss.JoinVertical(lipgloss.Left, actualLabel, actualBody)

	// Combine sections horizontally
	inputSections := lipgloss.JoinHorizontal(
		lipgloss.Top,
		inputsSection,
		lipgloss.NewStyle().Width(2).Render(""), // spacer
		expectedSection,
		lipgloss.NewStyle().Width(2).Render(""), // spacer
	)
	sections := lipgloss.JoinVertical(
		lipgloss.Top,
		inputSections,
		lipgloss.NewStyle().Height(1).Render(""), // spacer
		actualSection,
	)

	// Wrap with name tag and container
	nameTag := Tag(" "+name, lipgloss.Color("#ffffff"), ColorAccentBlue)

	return detailsContainer.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			nameTag,
			lipgloss.NewStyle().MarginTop(1).Render(sections),
		),
	)
}
