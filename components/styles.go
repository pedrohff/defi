package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

// Shared color palette used across Défi components and views.
const (
	ColorWhite          lipgloss.Color = "#FFFFFF"
	ColorSuccess        lipgloss.Color = "#a6e3a1"
	ColorAlert          lipgloss.Color = "#ff4e86"
	ColorMutedAccent    lipgloss.Color = "#b6bca8"
	ColorSurfaceDark    lipgloss.Color = "#313130"
	ColorLanguageBadge  lipgloss.Color = "#b149e8"
	ColorFilenameBadge  lipgloss.Color = "#6b21e8"
	ColorPendingLabel   lipgloss.Color = "#717171ff"
	ColorPendingBlockFG lipgloss.Color = "#808080"
	ColorPendingBlockBG lipgloss.Color = "#1c1c1c"
	ColorHeaderMutedFG  lipgloss.Color = "#585858ff"
	ColorHeaderMutedBG  lipgloss.Color = "#b4d3b1ff"
	ColorSpinnerAccent  lipgloss.Color = "#f9e2af"
	ColorTextPrimary    lipgloss.Color = "#cdd6f4"
	ColorAccentBlue     lipgloss.Color = "#89b4fa"
	ColorTextMuted      lipgloss.Color = "#9399b2"
	ColorFailure        lipgloss.Color = "#f38ba8"
	ColorHelpText       lipgloss.Color = "241"
	ColorDemoAccent     lipgloss.Color = "63"
	ColorDemoPurple     lipgloss.Color = "#7D56F4"
	ColorDemoWhite      lipgloss.Color = "#FAFAFA"
	ColorSelectedBg     lipgloss.Color = "#d9d9d9"
)

// NewSpinner returns a pre-styled spinner using the shared color palette.
func NewSpinner() spinner.Model {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(ColorSpinnerAccent)
	return sp
}

// RenderError formats an error message with failure styling.
func RenderError(msg string) string {
	style := lipgloss.NewStyle().Foreground(ColorFailure).Bold(true)
	return style.Render("⚠️ " + msg)
}
