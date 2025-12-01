package components

import "github.com/charmbracelet/lipgloss"

func Tag(label string, color lipgloss.Color, background lipgloss.Color) string {
	tagStyle := lipgloss.NewStyle().
		Foreground(color).
		Background(background)

	borderStyle := lipgloss.NewStyle().Foreground(background)

	return borderStyle.Render("") + tagStyle.Render(label) + borderStyle.Render("")
}
