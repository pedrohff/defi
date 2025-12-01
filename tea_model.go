package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
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

// Footer status messages displayed in the UI.
const (
	statusIdle                = "Idle"
	statusReadyToRun          = "Ready to run"
	statusListeningForChanges = "Listening for changes..."
	statusListeningForFiles   = "Listening for file changes..."
	statusWaitingForFile      = "Waiting for file..."
	statusWaitingForFiles     = "Waiting for matching files..."
	statusNoTestCases         = "No test cases found"
	statusRunningTests        = "Running tests..."
	statusPreparingRun        = "Preparing run..."
)

type model struct {
	cfg appConfig

	spinner        spinner.Model
	runnerUpdates  <-chan tea.Msg
	watcherUpdates <-chan tea.Msg

	runnerActive bool

	activePath string

	pendingPath string
	hasPending  bool

	watchHasFile bool
	watcherErr   error

	width  int
	height int
	ready  bool

	summaryPassed int
	summaryTotal  int
	summaryErr    error

	testCases            []view.TestCaseData
	selectedIndex        int // -1 means no selection
	footerStatus         string
	footerLanguage       string
	footerFilename       string
	ignoreInitialWatcher bool
}

func newModel(cfg appConfig, initialPath string) model {
	m := model{
		cfg:           cfg,
		spinner:       components.NewSpinner(),
		selectedIndex: -1,
	}

	if initialPath != "" {
		m.activePath = initialPath
		m.watchHasFile = true
		m.footerLanguage = languageLabelForPath(initialPath)
		m.footerFilename = footerFilename(initialPath)
		if cfg.once {
			m.footerStatus = statusReadyToRun
		} else {
			m.footerStatus = statusListeningForChanges
			m.ignoreInitialWatcher = true
		}
	}

	if m.footerStatus == "" {
		if cfg.once {
			m.footerStatus = statusWaitingForFile
		} else {
			m.footerStatus = statusWaitingForFiles
		}
	}

	if m.footerLanguage == "" {
		m.footerLanguage = "-"
	}

	if m.footerFilename == "" {
		m.footerFilename = "-"
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
			m.footerStatus = v.Name

		case testsInitMsg:
			m.testCases = make([]view.TestCaseData, v.Total)
			for i := range m.testCases {
				m.testCases[i] = view.TestCaseData{
					Name:             fmt.Sprintf("Case %d", i+1),
					Status:           components.TestCasePending,
					CompileSuccess:   false,
					AssertionSuccess: false,
				}
			}
			if v.Total == 0 {
				m.footerStatus = statusNoTestCases
			} else {
				m.footerStatus = statusRunningTests
			}

		case testStatusMsg:
			if idx := v.Current - 1; idx >= 0 && idx < len(m.testCases) {
				tc := &m.testCases[idx]
				tc.Inputs = v.Inputs
				tc.ExpectedOutput = v.ExpectedOutput
				tc.ActualOutput = v.ActualOutput
				switch v.Status {
				case testStatusRunning:
					tc.Status = components.TestCaseRunning
					tc.CompileSuccess = false
					tc.AssertionSuccess = false
					m.footerStatus = fmt.Sprintf("Case %d/%d running", v.Current, v.Total)
				case testStatusPassed:
					tc.Status = components.TestCaseFinished
					tc.CompileSuccess = v.CompileSuccess
					tc.AssertionSuccess = v.AssertionSuccess
					m.footerStatus = fmt.Sprintf("Case %d/%d passed", v.Current, v.Total)
				case testStatusFailed:
					tc.Status = components.TestCaseFinished
					tc.CompileSuccess = v.CompileSuccess
					tc.AssertionSuccess = v.AssertionSuccess
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
			m.runnerActive = false
			if v.Err != nil {
				m.footerStatus = shortenString(v.Err.Error(), 60)
			} else if !m.cfg.once {
				m.footerStatus = statusListeningForFiles
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
			m.footerLanguage = languageLabelForPath(v.Path)
			m.footerFilename = footerFilename(v.Path)
			if !m.runnerActive {
				if m.cfg.once {
					m.footerStatus = statusReadyToRun
				} else {
					m.footerStatus = statusListeningForChanges
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
				m.footerFilename = "-"
				m.footerLanguage = "-"
				if m.cfg.once {
					m.footerStatus = statusWaitingForFile
				} else {
					m.footerStatus = statusWaitingForFiles
				}
			}
		case watchErrMsg:
			m.watcherErr = v.Err
			m.footerStatus = fmt.Sprintf("Watcher error: %s", shortenString(v.Err.Error(), 40))
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

		m.resetForNewRun(msg.path)
		return m, startRunnerCmd(msg.path, m.cfg.compileFlags)
	}

	return m, nil
}

// resetForNewRun clears all test state and prepares the model for a fresh run.
func (m *model) resetForNewRun(path string) {
	// Runner state
	m.runnerActive = true
	m.hasPending = false
	m.pendingPath = ""

	// Previous results
	m.summaryErr = nil
	m.summaryPassed = 0
	m.summaryTotal = 0

	// File info
	m.activePath = path
	m.footerLanguage = languageLabelForPath(path)
	m.footerFilename = footerFilename(path)
	m.footerStatus = statusPreparingRun
	m.watchHasFile = true
}

func (m model) View() string {
	if !m.ready {
		return "🚀 Starting Défi...\n"
	}

	if m.watcherErr != nil {
		return components.RenderError(m.watcherErr.Error())
	}

	statusText := m.footerStatus
	if statusText == "" {
		statusText = statusIdle
	}

	mainView := view.NewMainView(m.width, m.height, m.testCases,
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
