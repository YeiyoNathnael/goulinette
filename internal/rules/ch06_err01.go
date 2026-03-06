package rules

import (
	"strings"
	"unicode"

	"goulinette/internal/diag"
)

type err01Rule struct{}

func NewERR01() Rule {
	return err01Rule{}
}

func (err01Rule) ID() string {
	return "ERR-01"
}

func (err01Rule) Chapter() int {
	return 6
}

func (err01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
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
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "ERR-01",
				Severity: diag.SeverityError,
				Message:  "error message must not start with a capital letter",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "start error strings with lowercase",
			})
		}
	}

	return diagnostics, nil
}
