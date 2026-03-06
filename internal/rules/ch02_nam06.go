package rules

import (
	"strings"

	"goulinette/internal/diag"
)

type nam06Rule struct{}

func NewNAM06() Rule {
	return nam06Rule{}
}

func (nam06Rule) ID() string {
	return "NAM-06"
}

func (nam06Rule) Chapter() int {
	return 2
}

func (nam06Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	forbidden := map[string]struct{}{
		"util":    {},
		"helpers": {},
		"common":  {},
		"misc":    {},
		"shared":  {},
		"tools":   {},
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		pkgName := strings.ToLower(strings.TrimSpace(pf.File.Name.Name))
		if _, bad := forbidden[pkgName]; !bad {
			continue
		}

		pos := pf.FSet.Position(pf.File.Name.Pos())
		diagnostics = append(diagnostics, diag.Diagnostic{
			RuleID:   "NAM-06",
			Severity: diag.SeverityError,
			Message:  "package name is too generic and forbidden",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "rename package to a descriptive noun",
		})
	}

	return diagnostics, nil
}
