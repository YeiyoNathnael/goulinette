package rules

import (
	"strings"
	"unicode"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type err01Rule struct{}

const err01Chapter = 6

// NewERR01 returns the ERR01 rule implementation.
func NewERR01() Rule {
	return err01Rule{}
}

// ID returns the rule identifier.
func (err01Rule) ID() string {
	return ruleERR01
}

// Chapter returns the chapter number for this rule.
func (err01Rule) Chapter() int {
	return err01Chapter
}

// Run executes this rule against the provided context.
func (err01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, item := range collectErrorMessageLiterals(pf.File) {
			trimmed := strings.TrimSpace(item.text)
			if trimmed == "" {
				continue
			}
			first := []rune(trimmed)[0]
			if !unicode.IsUpper(first) {
				continue
			}

			pos := pf.FSet.Position(item.call.Lparen)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleERR01,
				Severity: diag.SeverityError,
				Message:  "error message must not start with a capital letter",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "start error strings with lowercase",
			})
		}
	}

	return diagnostics, nil
}
