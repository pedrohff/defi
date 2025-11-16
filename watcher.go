package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type watchMode int

const (
	watchModeFile watchMode = iota
	watchModeDirectory
)

type watchSpec struct {
	original string
	mode     watchMode
	filePath string
	dir      string
	pattern  string
}

var errNoMatchingFiles = errors.New("no matching files found")

func parseWatchSpec(input string) (watchSpec, error) {
	if input == "" {
		input = "."
	}

	cleaned := filepath.Clean(input)
	spec := watchSpec{original: input}

	if hasGlob(cleaned) {
		dir := filepath.Dir(cleaned)
		if dir == "" {
			dir = "."
		}
		spec.mode = watchModeDirectory
		spec.dir = dir
		spec.pattern = filepath.Base(cleaned)
		return spec, nil
	}

	info, err := os.Stat(cleaned)
	if err != nil {
		return spec, err
	}

	if info.IsDir() {
		spec.mode = watchModeDirectory
		spec.dir = cleaned
		return spec, nil
	}

	spec.mode = watchModeFile
	spec.filePath = cleaned
	spec.dir = filepath.Dir(cleaned)
	return spec, nil
}

func hasGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func resolveLatestTarget(spec watchSpec) (string, time.Time, error) {
	switch spec.mode {
	case watchModeFile:
		info, err := os.Stat(spec.filePath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return "", time.Time{}, errNoMatchingFiles
			}
			return "", time.Time{}, err
		}
		return spec.filePath, info.ModTime(), nil

	case watchModeDirectory:
		entries, err := os.ReadDir(spec.dir)
		if err != nil {
			return "", time.Time{}, err
		}

		var (
			latestPath string
			latestTime time.Time
		)

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if spec.pattern != "" {
				matched, err := filepath.Match(spec.pattern, name)
				if err != nil {
					return "", time.Time{}, err
				}
				if !matched {
					continue
				}
			} else {
				if _, ok := supportedLanguages[filepath.Ext(name)]; !ok {
					continue
				}
			}

			path := filepath.Join(spec.dir, name)
			info, err := entry.Info()
			if err != nil {
				return "", time.Time{}, err
			}

			modTime := info.ModTime()
			if latestPath == "" || modTime.After(latestTime) {
				latestPath = path
				latestTime = modTime
			}
		}

		if latestPath == "" {
			return "", time.Time{}, errNoMatchingFiles
		}

		return latestPath, latestTime, nil
	default:
		return "", time.Time{}, errors.New("unknown watch mode")
	}
}

type watchEventMsg struct {
	Path    string
	ModTime time.Time
	Initial bool
}

type watchIdleMsg struct{}

type watchErrMsg struct {
	Err error
}

type watcherStartedMsg struct {
	ch <-chan tea.Msg
}

type watcherUpdateMsg struct {
	msg  tea.Msg
	done bool
}

func startWatcherCmd(spec watchSpec, interval time.Duration) tea.Cmd {
	if interval <= 0 {
		interval = time.Second
	}

	return func() tea.Msg {
		ch := make(chan tea.Msg, 16)
		go watchLoop(spec, interval, ch)
		return watcherStartedMsg{ch: ch}
	}
}

func watchLoop(spec watchSpec, interval time.Duration, ch chan<- tea.Msg) {
	defer close(ch)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	first := true
	var (
		lastPath    string
		lastTime    time.Time
		lastHadFile bool
	)

	for {
		path, modTime, err := resolveLatestTarget(spec)
		if err != nil {
			if errors.Is(err, errNoMatchingFiles) {
				if lastHadFile || first {
					ch <- watchIdleMsg{}
				}
				lastHadFile = false
			} else {
				ch <- watchErrMsg{Err: err}
			}
		} else {
			lastHadFile = true
			if first || path != lastPath || modTime.After(lastTime) {
				ch <- watchEventMsg{Path: path, ModTime: modTime, Initial: first}
				lastPath = path
				lastTime = modTime
			}
		}

		first = false

		if _, ok := <-ticker.C; !ok {
			return
		}
	}
}

func readWatcherUpdateCmd(ch <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return watcherUpdateMsg{done: true}
		}
		return watcherUpdateMsg{msg: msg}
	}
}

func (spec watchSpec) DisplayBase() string {
	if spec.mode == watchModeFile {
		return spec.filePath
	}
	if spec.pattern != "" {
		return filepath.Join(spec.dir, spec.pattern)
	}
	return spec.dir
}
