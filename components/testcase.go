package components

import "github.com/charmbracelet/lipgloss"

const (
	// TestCasePending indicates that a case has not yet started.
	TestCasePending = "PENDING"
	// TestCaseRunning indicates that the case is currently executing.
	TestCaseRunning = "RUNNING"
	// TestCaseFinished indicates that both compilation and assertions completed.
	TestCaseFinished = "FINISHED"
	// TestCaseBlockStatusPending renders the block as pending.
	TestCaseBlockStatusPending = "-"
	// TestCaseBlockStatusPass renders the block as a successful pass.
	TestCaseBlockStatusPass = "PASS"
	// TestCaseBlockStatusFail renders the block as a failure.
	TestCaseBlockStatusFail = "FAIL"
	// TestCaseBlockSize defines the width reserved for each result block.
	TestCaseBlockSize = 9
)

var (
	// TestCaseNameStyle styles the test case label when a result is available.
	TestCaseNameStyle = lipgloss.NewStyle().Padding(0, 1)

	// TestCaseNameStylePending styles the test case label while pending.
	TestCaseNameStylePending = TestCaseNameStyle.Foreground(ColorPendingLabel)

	// TestCaseResultBlockPendingStyle styles a block representing a pending result.
	TestCaseResultBlockPendingStyle = lipgloss.NewStyle().
					Width(TestCaseBlockSize).Align(lipgloss.Center).
					Italic(true).
					Foreground(ColorPendingBlockFG).
					Background(ColorPendingBlockBG).Padding(0, 1)

	// TestCaseResultBlockRunningStyle styles a block while the case is in flight.
	TestCaseResultBlockRunningStyle = lipgloss.NewStyle().
					Width(TestCaseBlockSize).Align(lipgloss.Center).
					Foreground(ColorSuccess).
					Background(ColorPendingBlockBG).Padding(0, 1)

	// TestCaseResultBlockPassedStyle styles a block when the case succeeds.
	TestCaseResultBlockPassedStyle = lipgloss.NewStyle().
					Width(TestCaseBlockSize).Align(lipgloss.Center).
					Bold(true).
					Foreground(ColorWhite).
					Background(ColorSuccess).Padding(0, 1)

	// TestCaseResultBlockFailedStyle styles a block when the case fails.
	TestCaseResultBlockFailedStyle = lipgloss.NewStyle().
					Width(TestCaseBlockSize).Align(lipgloss.Center).
					Bold(true).
					Foreground(ColorSurfaceDark).
					Background(ColorAlert).Padding(0, 1)
)

// TestCaseHeader renders the column headers used by TestCase rows.
func TestCaseHeader(width int) string {
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorHeaderMutedFG).
		Background(ColorHeaderMutedBG).
		Width(width).
		Align(lipgloss.Left).Bold(true).Italic(true)

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		headerStyle.Width(width-(2*TestCaseBlockSize)).Render("TEST CASE"),
		headerStyle.Width(TestCaseBlockSize).Render("COMPILE"),
		headerStyle.Width(TestCaseBlockSize).Render("ASSERT"),
	)
}

// TestCase renders a single test case row with compilation and assertion result blocks.
func TestCase(width int, name string, status string, compileSuccess bool, assertionSuccess bool, isSelected bool) string {
	testCaseNameStyle := TestCaseNameStylePending
	compileStyle := TestCaseResultBlockPendingStyle
	assertionSuccessStyle := TestCaseResultBlockPendingStyle
	compileStatus := TestCaseBlockStatusPending
	assertionStatus := TestCaseBlockStatusPending

	switch status {
	case TestCaseFinished:
		testCaseNameStyle = TestCaseNameStyle
		if compileSuccess {
			compileStyle = TestCaseResultBlockPassedStyle
			compileStatus = TestCaseBlockStatusPass
		} else {
			compileStyle = TestCaseResultBlockFailedStyle
			compileStatus = TestCaseBlockStatusFail
		}

		if assertionSuccess {
			assertionSuccessStyle = TestCaseResultBlockPassedStyle
			assertionStatus = TestCaseBlockStatusPass
		} else {
			assertionSuccessStyle = TestCaseResultBlockFailedStyle
			assertionStatus = TestCaseBlockStatusFail
		}
	case TestCaseRunning:
		testCaseNameStyle = TestCaseNameStyle
		compileStyle = TestCaseResultBlockRunningStyle
		assertionSuccessStyle = TestCaseResultBlockRunningStyle
	}

	testCaseNameColumn := testCaseNameStyle.Width(width - (2 * TestCaseBlockSize))

	if isSelected {
		testCaseNameColumn = testCaseNameColumn.Background(ColorSelectedBg)
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		testCaseNameColumn.Render(name),
		lipgloss.PlaceHorizontal(TestCaseBlockSize, lipgloss.Center, compileStyle.Render(compileStatus)),
		lipgloss.PlaceHorizontal(TestCaseBlockSize, lipgloss.Center, assertionSuccessStyle.Render(assertionStatus)),
	)
}
