package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// FooterStatusTabStyle styles the static "STATUS" label in the footer.
	FooterStatusTabStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorAlert).
				Align(lipgloss.Center).Padding(0, 1)

	// FooterStatusDescStyle styles the dynamic status description area.
	FooterStatusDescStyle = lipgloss.NewStyle().
				Foreground(ColorMutedAccent).
				Background(ColorSurfaceDark).Padding(0, 1)

	FooterLanguageStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorLanguageBadge).
				Align(lipgloss.Center)

	FooterFilenameStyle = lipgloss.NewStyle().
				Foreground(ColorWhite).
				Background(ColorFilenameBadge).
				Align(lipgloss.Center).Padding(0, 1)
)
// Footer composes the footer layout with the current status message.
func Footer(width int, status string, language string, filename string) string {
	if len(filename) > 18 {
		filename = strings.TrimSpace(filename[:17]) + "…"
	}
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		FooterStatusTabStyle.Render("STATUS"),
		FooterStatusDescStyle.Width(width-8-10-20).Render(status),
		FooterLanguageStyle.Width(10).Render(language),
		FooterFilenameStyle.Width(20).Render(filename),
	)
}
