package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// PromptCase represents a single parsed prompt with its inputs and expected outputs.
type PromptCase struct {
	Inputs  []string
	Outputs []string
}

// PromptParser extracts prompt test cases from a source file.
type PromptParser struct {
	path string
}

// NewPromptParser returns a parser bound to the provided file path.
func NewPromptParser(path string) *PromptParser {
	return &PromptParser{path: path}
}

// Parse reads the file and returns every prompt case contained in defiprompt blocks.
func (p *PromptParser) Parse() ([]PromptCase, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", p.path, err)
	}

	cases, err := parsePromptContent(string(data))
	if err != nil {
		return nil, err
	}

	if len(cases) == 0 {
		return nil, fmt.Errorf("no defiprompt blocks found in %q", p.path)
	}

	return cases, nil
}

func parsePromptContent(content string) ([]PromptCase, error) {
	const marker = "/*defiprompt"
	var (
		cases    []PromptCase
		searchAt int
	)

	for {
		start := strings.Index(content[searchAt:], marker)
		if start == -1 {
			break
		}
		start += searchAt

		end := strings.Index(content[start+len(marker):], "*/")
		if end == -1 {
			return nil, fmt.Errorf("unterminated defiprompt block")
		}
		end += start + len(marker)

		block := content[start+len(marker) : end]
		blockCases, err := parsePromptBlock(block)
		if err != nil {
			return nil, err
		}
		cases = append(cases, blockCases...)

		searchAt = end + len("*/")
	}

	return cases, nil
}

// parsePromptBlock walks through a defiprompt comment, emitting the contained cases.
func parsePromptBlock(block string) ([]PromptCase, error) {
	scanner := bufio.NewScanner(strings.NewReader(block))
	var (
		cases   []PromptCase
		current *PromptCase
		state   string
	)

	flushCurrent := func() error {
		if current == nil {
			return nil
		}
		if len(current.Inputs) == 0 || len(current.Outputs) == 0 {
			return fmt.Errorf("incomplete prompt case detected")
		}
		cases = append(cases, *current)
		current = nil
		state = ""
		return nil
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch line {
		case "INPUTS":
			fallthrough
		case "INPUTS:":
			if err := flushCurrent(); err != nil {
				return nil, err
			}
			current = &PromptCase{}
			state = "input"
		case "OUTPUT":
			fallthrough
		case "OUTPUT:":
			if current == nil {
				return nil, fmt.Errorf("OUTPUT encountered before INPUTS")
			}
			state = "output"
		case "-*-":
			if err := flushCurrent(); err != nil {
				return nil, err
			}
		default:
			switch state {
			case "input":
				current.Inputs = append(current.Inputs, line)
			case "output":
				current.Outputs = append(current.Outputs, line)
			default:
				// Ignore stray lines outside a known section.
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if err := flushCurrent(); err != nil {
		return nil, err
	}

	return cases, nil
}
