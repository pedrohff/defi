package main

import (
	"os"
	"path/filepath"
	"strings"
)

func formatDisplayPath(path string) string {
	if path == "" {
		return ""
	}
	if rel, err := filepath.Rel(".", path); err == nil {
		return rel
	}
	return path
}

func languageLabelForPath(path string) string {
	if path == "" {
		return "-"
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".cpp", ".cc", ".cxx", ".hpp", ".hh":
		return "C++"
	case ".c":
		return "C"
	case ".go":
		return "Go"
	case ".py":
		return "Python"
	case ".rs":
		return "Rust"
	case ".java":
		return "Java"
	case ".js":
		return "JavaScript"
	case ".ts":
		return "TypeScript"
	case ".kt":
		return "Kotlin"
	case ".swift":
		return "Swift"
	}
	if ext != "" {
		return strings.TrimPrefix(ext, ".")
	}
	return "-"
}

func footerFilename(path string) string {
	if path == "" {
		return "-"
	}
	name := filepath.Base(path)
	if name == "." || name == string(os.PathSeparator) {
		return "-"
	}
	return name
}

func shortenString(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= 3 {
		return string(runes[:max])
	}
	return string(runes[:max-3]) + "..."
}
