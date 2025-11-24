package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var supportedLanguages = map[string]string{
	".cpp": "c++",
}

var defaultCompileFlags = map[string][]string{
	".cpp": {"-std=c++11"},
}

const (
	compiledBinary    = "defitestprogram"
	compiledBinaryWin = "defitestprogram.exe"
)

type phaseMsg struct {
	Name      string
	Index     int
	Total     int
	Completed bool
}

type testsInitMsg struct {
	Total int
}

type testStatus string

const (
	testStatusRunning testStatus = "running"
	testStatusPassed  testStatus = "passed"
	testStatusFailed  testStatus = "failed"
)

type testStatusMsg struct {
	Current          int
	Total            int
	Passed           int
	Status           testStatus
	Err              error
	CompileSuccess   bool
	AssertionSuccess bool
}

type testsDoneMsg struct {
	Passed int
	Total  int
	Err    error
}

func runWorkflow(sourcePath string, overrideFlags []string, send func(tea.Msg)) (int, int, error) {
	var (
		compiler     string
		cases        []PromptCase
		total        int
		defaultFlags []string
	)

	phases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "🔍 Validating source",
			fn: func() error {
				info, err := os.Stat(sourcePath)
				if err != nil {
					return fmt.Errorf("failed to access %q: %w", sourcePath, err)
				}

				if info.IsDir() {
					return fmt.Errorf("%q is a directory, expected a file", sourcePath)
				}

				ext := filepath.Ext(sourcePath)
				var ok bool
				compiler, ok = supportedLanguages[ext]
				if !ok {
					return fmt.Errorf("unsupported file extension %q", ext)
				}
				defaultFlags = defaultCompileFlags[ext]

				if _, err := exec.LookPath(compiler); err != nil {
					return fmt.Errorf("required compiler %q not found in PATH: %w", compiler, err)
				}

				return nil
			},
		},
		{
			name: "🧹 Cleaning previous build",
			fn:   removeExistingBinaries,
		},
		{
			name: "🛠️ Compiling",
			fn: func() error {
				flags := defaultFlags
				if len(overrideFlags) > 0 {
					flags = overrideFlags
				}
				return compileSource(sourcePath, compiler, flags)
			},
		},
		{
			name: "📝 Parsing prompts",
			fn: func() error {
				parsed, err := NewPromptParser(sourcePath).Parse()
				if err != nil {
					return err
				}
				cases = parsed
				total = len(parsed)
				return nil
			},
		},
	}

	for i, phase := range phases {
		send(phaseMsg{Name: phase.name, Index: i + 1, Total: len(phases)})
		time.Sleep(time.Millisecond * 100) // Simulate some delay for better UX
		if err := phase.fn(); err != nil {
			return 0, total, err
		}
		send(phaseMsg{Name: phase.name, Index: i + 1, Total: len(phases), Completed: true})
	}

	send(testsInitMsg{Total: total})

	passed := 0
	var firstErr error

	for idx, c := range cases {
		time.Sleep(time.Millisecond * 100) // Simulate some delay for better UX
		send(testStatusMsg{
			Current: idx + 1,
			Total:   total,
			Passed:  passed,
			Status:  testStatusRunning,
		})

		outputs, err := runSingleCase(idx, c)
		time.Sleep(time.Millisecond * 200) // Simulate some delay for better UX
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			send(testStatusMsg{
				Current:          idx + 1,
				Total:            total,
				Passed:           passed,
				Status:           testStatusFailed,
				CompileSuccess:   false,
				AssertionSuccess: false,
				Err:              err,
			})
			continue
		}

		if err := compareOutputs(c.Outputs, outputs); err != nil {
			wrapped := fmt.Errorf("case %d: %w", idx+1, err)
			if firstErr == nil {
				firstErr = wrapped
			}
			send(testStatusMsg{
				Current:          idx + 1,
				Total:            total,
				Passed:           passed,
				Status:           testStatusFailed,
				CompileSuccess:   true,
				AssertionSuccess: false,
				Err:              wrapped,
			})
			continue
		}

		passed++
		send(testStatusMsg{
			Current:          idx + 1,
			Total:            total,
			Passed:           passed,
			Status:           testStatusPassed,
			CompileSuccess:   true,
			AssertionSuccess: true,
		})
	}
	time.Sleep(time.Millisecond * 300) // Simulate some delay for better UX

	return passed, total, firstErr
}

func compileSource(sourcePath, compiler string, flags []string) error {
	args := append([]string{}, flags...)
	args = append(args, sourcePath, "-o", compiledBinary)
	cmd := exec.Command(compiler, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation failed: %w", err)
	}
	return nil
}

func removeExistingBinaries() error {
	if err := removeIfExists(compiledBinary); err != nil {
		return err
	}

	if err := removeIfExists(compiledBinaryWin); err != nil {
		return err
	}

	return nil
}

func removeIfExists(path string) error {
	err := os.Remove(path)
	if err == nil || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return fmt.Errorf("unable to delete %q: %w", path, err)
}

func runSingleCase(idx int, c PromptCase) ([]string, error) {
	cmd := exec.Command("./" + compiledBinary)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("case %d: failed to obtain stdin: %w", idx+1, err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("case %d: failed to obtain stdout: %w", idx+1, err)
	}

	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		stdin.Close()
		return nil, fmt.Errorf("case %d: start failed: %w", idx+1, err)
	}

	for _, line := range c.Inputs {
		if _, err := fmt.Fprintln(stdin, line); err != nil {
			stdin.Close()
			cmd.Wait()
			return nil, fmt.Errorf("case %d: failed to write input: %w", idx+1, err)
		}
	}
	stdin.Close()

	var outputs []string
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		outputs = append(outputs, strings.TrimRight(scanner.Text(), "\r"))
	}

	if err := scanner.Err(); err != nil {
		cmd.Wait()
		return nil, fmt.Errorf("case %d: failed to read stdout: %w", idx+1, err)
	}

	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("case %d: execution failed: %w", idx+1, err)
	}

	return outputs, nil
}

func compareOutputs(expected, actual []string) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("expected %d output lines, got %d", len(expected), len(actual))
	}

	for i := range expected {
		if strings.TrimSpace(expected[i]) != strings.TrimSpace(actual[i]) {
			return fmt.Errorf("expected output %q, got %q (line %d)", expected[i], actual[i], i+1)
		}
	}

	return nil
}
