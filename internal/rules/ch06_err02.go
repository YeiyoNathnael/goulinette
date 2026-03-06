package rules

import (
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type err02Rule struct{}

const err02Chapter = 6

// NewERR02 returns the ERR02 rule implementation.
func NewERR02() Rule {
	return err02Rule{}
}

// ID returns the rule identifier.
func (err02Rule) ID() string {
	return ruleERR02
}

// Chapter returns the chapter number for this rule.
func (err02Rule) Chapter() int {
	return err02Chapter
}

// Run executes this rule against the provided context.
func (err02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, item := range collectErrorMessageLiterals(pf.File) {
			if !hasForbiddenErrorSuffix(item.text) {
				continue
			}

			pos := pf.FSet.Position(item.call.Lparen)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleERR02,
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
