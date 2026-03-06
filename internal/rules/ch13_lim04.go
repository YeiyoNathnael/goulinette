package rules

import (
	"os"
	"strconv"
	"strings"

	"goulinette/internal/diag"
)

type lim04Rule struct{}

func NewLIM04() Rule {
	return lim04Rule{}
}

func (lim04Rule) ID() string {
	return "LIM-04"
}

func (lim04Rule) Chapter() int {
	return 13
}

func (lim04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		filename := pf.Path
		if strings.HasSuffix(filename, "_test.go") {
			continue
		}
		if isGeneratedSourceFile(filename, pf.File) {
			continue
		}

		content, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		lineCount := normalizedFileLineCount(string(content))
		if lineCount <= 500 {
			continue
		}

		pos := pf.FSet.Position(pf.File.Package)
		diagnostics = append(diagnostics, diag.Diagnostic{
			RuleID:   "LIM-04",
			Severity: diag.SeverityWarning,
			Message:  "source files should not exceed 500 lines",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "file has " + strconv.Itoa(lineCount) + " lines; split responsibilities into smaller files",
		})
	}

	return diagnostics, nil
}

func normalizedFileLineCount(content string) int {
	lines := strings.Split(content, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return len(lines)
}
