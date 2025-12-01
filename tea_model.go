package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pedrohff/defi/components"
	"github.com/pedrohff/defi/view"
)

type runnerStartedMsg struct {
	ch <-chan tea.Msg
}

type runnerUpdateMsg struct {
	msg  tea.Msg
	done bool
}

type runRequestMsg struct {
	path string
}

type uiStyles struct {
	title     lipgloss.Style
	accent    lipgloss.Style
	subtle    lipgloss.Style
	success   lipgloss.Style
	failure   lipgloss.Style
	body      lipgloss.Style
	helpStyle lipgloss.Style
}

type testCaseView struct {
	Name           string
	Status         string
	CompileOK      bool
	AssertionOK    bool
	Inputs         []string
	ExpectedOutput string
	ActualOutput   string
}

type model struct {
	cfg appConfig

	spinner        spinner.Model
	styles         uiStyles
	runnerUpdates  <-chan tea.Msg
	watcherUpdates <-chan tea.Msg

	runnerActive bool
	runFinished  bool

	activePath  string
	displayPath string

	pendingPath string
	hasPending  bool

	watchHasFile bool
	watcherErr   error

	width  int
	height int
	ready  bool

	phaseName  string
	phaseIndex int
	phaseTotal int

	testsStarted bool
	testsTotal   int
	testsPassed  int
	currentTest  int
	currentState testStatus
	currentErr   error

	summaryPassed int
	summaryTotal  int
	summaryErr    error

	testCases            []testCaseView
	selectedIndex        int // -1 means no selection
	footerStatus         string
	footerLanguage       string
	footerFilename       string
	footerSpinning       bool
	ignoreInitialWatcher bool
}

func newModel(cfg appConfig, initialPath string) model {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(components.ColorSpinnerAccent)

	styles := uiStyles{
		title:     lipgloss.NewStyle().Foreground(components.ColorTextPrimary).Bold(true),
		accent:    lipgloss.NewStyle().Foreground(components.ColorAccentBlue).Bold(true),
		subtle:    lipgloss.NewStyle().Foreground(components.ColorTextMuted),
		success:   lipgloss.NewStyle().Foreground(components.ColorSuccess).Bold(true),
		failure:   lipgloss.NewStyle().Foreground(components.ColorFailure).Bold(true),
		body:      lipgloss.NewStyle().Foreground(components.ColorTextPrimary),
		helpStyle: lipgloss.NewStyle().Foreground(components.ColorHelpText),
	}

	m := model{
		cfg:           cfg,
		spinner:       sp,
		styles:        styles,
		selectedIndex: -1,
	}

	if initialPath != "" {
		m.activePath = initialPath
		m.displayPath = formatDisplayPath(initialPath)
		m.watchHasFile = true
		m.footerLanguage = languageLabelForPath(initialPath)
		m.footerFilename = footerFilename(initialPath)
		if cfg.once {
			m.footerStatus = "Ready to run"
		} else {
			m.footerStatus = "Listening for changes..."
			m.ignoreInitialWatcher = true
		}
	}

	if m.footerStatus == "" {
		if cfg.once {
			m.footerStatus = "Waiting for file..."
		} else {
			m.footerStatus = "Waiting for matching files..."
		}
	}

	if m.footerLanguage == "" {
		m.footerLanguage = "-"
	}

	if m.footerFilename == "" {
		m.footerFilename = "-"
	}

	if !cfg.once {
		m.footerSpinning = true
	}

	return m
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, m.spinner.Tick)

	if m.cfg.once {
		if m.activePath != "" {
			cmds = append(cmds, requestRunCmd(m.activePath))
		}
	} else {
		cmds = append(cmds, startWatcherCmd(m.cfg.spec, m.cfg.interval))
		if m.activePath != "" {
			cmds = append(cmds, requestRunCmd(m.activePath))
		}
	}

	return tea.Batch(cmds...)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case tea.KeyDown:
			if m.selectedIndex < len(m.testCases)-1 {
				m.selectedIndex++
			}
		case tea.KeyEsc:
			m.selectedIndex = -1
		}
		// Vim-style navigation
		switch msg.String() {
		case "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
		case "j":
			if m.selectedIndex < len(m.testCases)-1 {
				m.selectedIndex++
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case runnerStartedMsg:
		m.runnerUpdates = msg.ch
		return m, readRunnerUpdateCmd(m.runnerUpdates)

	case runnerUpdateMsg:
		if msg.done {
			m.runnerUpdates = nil
			return m, nil
		}

		switch v := msg.msg.(type) {
		case phaseMsg:
			m.phaseName = v.Name
			m.phaseIndex = v.Index
			m.phaseTotal = v.Total
			m.footerStatus = v.Name
			m.footerSpinning = false

		case testsInitMsg:
			m.testsStarted = false
			m.testsTotal = v.Total
			m.testsPassed = 0
			m.testCases = make([]testCaseView, v.Total)
			for i := range m.testCases {
				m.testCases[i] = testCaseView{
					Name:        fmt.Sprintf("Case %d", i+1),
					Status:      components.TestCasePending,
					CompileOK:   false,
					AssertionOK: false,
				}
			}
			if v.Total == 0 {
				m.footerStatus = "No test cases found"
			} else {
				m.footerStatus = "Running tests..."
			}
			m.footerSpinning = false

		case testStatusMsg:
			m.testsStarted = true
			m.currentTest = v.Current
			m.testsTotal = v.Total
			m.testsPassed = v.Passed
			m.currentState = v.Status
			m.currentErr = v.Err
			m.footerSpinning = false

			if idx := v.Current - 1; idx >= 0 && idx < len(m.testCases) {
				tc := &m.testCases[idx]
				tc.Inputs = v.Inputs
				tc.ExpectedOutput = v.ExpectedOutput
				tc.ActualOutput = v.ActualOutput
				switch v.Status {
				case testStatusRunning:
					tc.Status = components.TestCaseRunning
					tc.CompileOK = false
					tc.AssertionOK = false
					m.footerStatus = fmt.Sprintf("Case %d/%d running", v.Current, v.Total)
				case testStatusPassed:
					tc.Status = components.TestCaseFinished
					tc.CompileOK = v.CompileSuccess
					tc.AssertionOK = v.AssertionSuccess
					m.footerStatus = fmt.Sprintf("Case %d/%d passed", v.Current, v.Total)
				case testStatusFailed:
					tc.Status = components.TestCaseFinished
					tc.CompileOK = v.CompileSuccess
					tc.AssertionOK = v.AssertionSuccess
					status := "failed"
					if v.Err != nil {
						status = fmt.Sprintf("failed: %s", shortenString(v.Err.Error(), 60))
					}
					m.footerStatus = fmt.Sprintf("Case %d/%d %s", v.Current, v.Total, status)
				}
			}

		case testsDoneMsg:
			m.summaryPassed = v.Passed
			m.summaryTotal = v.Total
			m.summaryErr = v.Err
			m.runFinished = true
			m.runnerActive = false
			m.footerSpinning = false
			if v.Err != nil {
				m.footerStatus = shortenString(v.Err.Error(), 60)
			} else if !m.cfg.once {
				m.footerStatus = "Listening for file changes..."
				m.footerSpinning = true
			}

			var cmds []tea.Cmd
			if m.runnerUpdates != nil {
				cmds = append(cmds, readRunnerUpdateCmd(m.runnerUpdates))
			}

			if m.cfg.once {
				cmds = append(cmds, tea.Quit)
			} else if m.hasPending {
				path := m.pendingPath
				m.pendingPath = ""
				m.hasPending = false
				cmds = append(cmds, requestRunCmd(path))
			}

			return m, tea.Batch(cmds...)
		}

		if m.runnerUpdates != nil {
			return m, readRunnerUpdateCmd(m.runnerUpdates)
		}
		return m, nil

	case watcherStartedMsg:
		m.watcherUpdates = msg.ch
		return m, readWatcherUpdateCmd(m.watcherUpdates)

	case watcherUpdateMsg:
		if msg.done {
			m.watcherUpdates = nil
			return m, nil
		}

		var cmds []tea.Cmd
		switch v := msg.msg.(type) {
		case watchEventMsg:
			m.watchHasFile = true
			m.watcherErr = nil
			m.activePath = v.Path
			m.displayPath = formatDisplayPath(v.Path)
			m.footerLanguage = languageLabelForPath(v.Path)
			m.footerFilename = footerFilename(v.Path)
			if !m.runnerActive {
				if m.cfg.once {
					m.footerStatus = "Ready to run"
					m.footerSpinning = false
				} else {
					m.footerStatus = "Listening for changes..."
					m.footerSpinning = true
				}
			}
			triggerRun := true
			if v.Initial && m.ignoreInitialWatcher {
				triggerRun = false
				m.ignoreInitialWatcher = false
			}
			if triggerRun {
				if m.runnerActive {
					m.pendingPath = v.Path
					m.hasPending = true
				} else {
					cmds = append(cmds, requestRunCmd(v.Path))
				}
			}
		case watchIdleMsg:
			m.watchHasFile = false
			if !m.runnerActive {
				m.displayPath = ""
				m.footerFilename = "-"
				m.footerLanguage = "-"
				if m.cfg.once {
					m.footerStatus = "Waiting for file..."
					m.footerSpinning = false
				} else {
					m.footerStatus = "Waiting for matching files..."
					m.footerSpinning = true
				}
			}
		case watchErrMsg:
			m.watcherErr = v.Err
			m.footerStatus = fmt.Sprintf("Watcher error: %s", shortenString(v.Err.Error(), 40))
			m.footerSpinning = false
		}

		cmds = append(cmds, readWatcherUpdateCmd(m.watcherUpdates))
		return m, tea.Batch(cmds...)

	case runRequestMsg:
		if msg.path == "" {
			return m, nil
		}
		if m.runnerActive {
			if m.pendingPath != msg.path {
				m.pendingPath = msg.path
			}
			m.hasPending = true
			return m, nil
		}

		m.runnerActive = true
		m.runFinished = false
		m.hasPending = false
		m.pendingPath = ""
		m.summaryErr = nil
		m.summaryPassed = 0
		m.summaryTotal = 0
		m.testsStarted = false
		m.testsTotal = 0
		m.testsPassed = 0
		m.currentTest = 0
		m.currentState = ""
		m.currentErr = nil
		m.phaseName = ""
		m.phaseIndex = 0
		m.phaseTotal = 0
		m.activePath = msg.path
		m.displayPath = formatDisplayPath(msg.path)
		m.footerLanguage = languageLabelForPath(msg.path)
		m.footerFilename = footerFilename(msg.path)
		m.footerStatus = "Preparing run..."
		m.footerSpinning = false
		m.watchHasFile = true

		return m, startRunnerCmd(msg.path, m.cfg.compileFlags)
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "🚀 Starting Défi...\n"
	}

	if m.watcherErr != nil {
		return m.styles.failure.Render("⚠️ " + m.watcherErr.Error())
	}

	// Convert testCaseView to view.TestCaseData
	testCases := make([]view.TestCaseData, len(m.testCases))
	for i, tc := range m.testCases {
		testCases[i] = view.TestCaseData{
			Name:             tc.Name,
			Status:           tc.Status,
			CompileSuccess:   tc.CompileOK,
			AssertionSuccess: tc.AssertionOK,
			Inputs:           tc.Inputs,
			ExpectedOutput:   tc.ExpectedOutput,
			ActualOutput:     tc.ActualOutput,
		}
	}

	statusText := m.footerStatus
	if statusText == "" {
		statusText = "Idle"
	}

	mainView := view.NewMainView(m.width, m.height, testCases,
		view.WithSelectedIndex(m.selectedIndex),
		view.WithFilename(m.footerFilename),
		view.WithLanguage(m.footerLanguage),
		view.WithStatus(statusText),
	)

	return mainView.Render()
}

func requestRunCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return runRequestMsg{path: path}
	}
}

func startRunnerCmd(sourcePath string, compileFlags []string) tea.Cmd {
	return func() tea.Msg {
		ch := make(chan tea.Msg, 16)
		go func() {
			passed, total, err := runWorkflow(sourcePath, compileFlags, func(msg tea.Msg) {
				ch <- msg
			})
			ch <- testsDoneMsg{Passed: passed, Total: total, Err: err}
			close(ch)
		}()
		return runnerStartedMsg{ch: ch}
	}
}

func readRunnerUpdateCmd(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return runnerUpdateMsg{done: true}
		}
		return runnerUpdateMsg{msg: msg}
	}
}
