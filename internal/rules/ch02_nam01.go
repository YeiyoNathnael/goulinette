package rules

import (
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam01Rule struct{}

const nam01Chapter = 2

// NewNAM01 returns the NAM01 rule implementation.
func NewNAM01() Rule {
	return nam01Rule{}
}

// ID returns the rule identifier.
func (nam01Rule) ID() string {
	return ruleNAM01
}

// Chapter returns the chapter number for this rule.
func (nam01Rule) Chapter() int {
	return nam01Chapter
}

// Run executes this rule against the provided context.
func (nam01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, ident := range collectDeclaredIdents(pf.File) {
			if ident.Name == "_" {
				continue
			}
			if strings.Contains(ident.Name, "_") {
				pos := pf.FSet.Position(ident.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleNAM01,
					Severity: diag.SeverityError,
					Message:  "identifier must use camelCase or PascalCase, snake_case is forbidden",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "rename identifier to camelCase/PascalCase",
				})
			}
		}
	}

	return diagnostics, nil
}
