package rules

import (
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type err02Rule struct{}

func NewERR02() Rule {
	return err02Rule{}
}

func (err02Rule) ID() string {
	return "ERR-02"
}

func (err02Rule) Chapter() int {
	return 6
}

func (err02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, item := range collectErrorMessageLiterals(pf.File) {
			if !hasForbiddenErrorSuffix(item.text) {
				continue
			}

			pos := pf.FSet.Position(item.call.Lparen)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "ERR-02",
				Severity: diag.SeverityError,
				Message:  "error message must not end with punctuation or newline",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "remove trailing punctuation/newline in error text",
			})
		}
	}

	return diagnostics, nil
}

func hasForbiddenErrorSuffix(msg string) bool {
	if strings.HasSuffix(msg, "\n") {
		return true
	}
	trimmed := strings.TrimSpace(msg)
	if trimmed == "" {
		return false
	}
	last := trimmed[len(trimmed)-1]
	switch last {
	case '.', '!', '?', ':':
		return true
	default:
		return false
	}
}
