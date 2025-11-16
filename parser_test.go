package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPromptParserParse(t *testing.T) {
	script := `#include <iostream>

/*defiprompt
INPUTS:
3
1
2
3
OUTPUT:
6
-*-
INPUTS:
4
10
2
3
5
OUTPUT:
20
-*-
INPUTS:
5
8
-3
4
0
11
OUTPUT:
20
*/

int main() {
    return 0;
}`

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "sample.cpp")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write temp script: %v", err)
	}

	parser := NewPromptParser(scriptPath)
	cases, err := parser.Parse()
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if len(cases) != 3 {
		t.Fatalf("expected 3 cases, got %d", len(cases))
	}

	expectedFirst := PromptCase{
		Inputs:  []string{"3", "1", "2", "3"},
		Outputs: []string{"6"},
	}

	if !reflect.DeepEqual(cases[0], expectedFirst) {
		t.Fatalf("unexpected first case: %#v", cases[0])
	}
}

func TestPromptParserParseNoPrompts(t *testing.T) {
	script := `int main() { return 0; }`

	dir := t.TempDir()
	scriptPath := filepath.Join(dir, "sample.cpp")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		t.Fatalf("failed to write temp script: %v", err)
	}

	parser := NewPromptParser(scriptPath)
	if _, err := parser.Parse(); err == nil {
		t.Fatalf("expected error when no prompts are present")
	}
}
