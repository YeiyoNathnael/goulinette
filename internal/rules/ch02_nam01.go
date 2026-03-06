package rules

import (
	"strings"

	"goulinette/internal/diag"
)

type nam01Rule struct{}

func NewNAM01() Rule {
	return nam01Rule{}
}

func (nam01Rule) ID() string {
	return "NAM-01"
}

func (nam01Rule) Chapter() int {
	return 2
}

func (nam01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, ident := range collectDeclaredIdents(pf.File) {
			if ident.Name == "_" {
				continue
			}
			if strings.Contains(ident.Name, "_") {
				pos := pf.FSet.Position(ident.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "NAM-01",
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
