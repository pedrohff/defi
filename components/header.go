// Package components contains reusable Bubble Tea view helpers.
package components

import "github.com/charmbracelet/lipgloss"

// HeaderStyle defines the default presentation for header titles.
var HeaderStyle = lipgloss.NewStyle().
	Foreground(ColorWhite).
	Background(ColorSuccess).
	Align(lipgloss.Center).Bold(true).Italic(true).MarginBottom(1)

// Header renders content centered within a styled header of the given width.
func Header(width int, content string) string {
	return HeaderStyle.
		Width(width).
		Render(content)
}
