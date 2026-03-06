package rules

import (
	"os"
	"strconv"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type lim04Rule struct{}

const (
	lim04Chapter      = 13
	lim04MaxFileLines = 500
	lim04TestFileSfx  = "_test.go"
)

// NewLIM04 returns the LIM04 rule implementation.
func NewLIM04() Rule {
	return lim04Rule{}
}

// ID returns the rule identifier.
func (lim04Rule) ID() string {
	return ruleLIM04
}

// Chapter returns the chapter number for this rule.
func (lim04Rule) Chapter() int {
	return lim04Chapter
}

// Run executes this rule against the provided context.
func (lim04Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		filename := pf.Path
		if strings.HasSuffix(filename, lim04TestFileSfx) {
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
		if lineCount <= lim04MaxFileLines {
			continue
		}

		pos := pf.FSet.Position(pf.File.Package)
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleLIM04,
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
