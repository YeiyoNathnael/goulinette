package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"goulinette/internal/diag"
)

type nam05Rule struct{}

func NewNAM05() Rule {
	return nam05Rule{}
}

func (nam05Rule) ID() string {
	return "NAM-05"
}

func (nam05Rule) Chapter() int {
	return 2
}

func (nam05Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				_, ok = ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				if strings.HasSuffix(ts.Name.Name, "er") {
					continue
				}
				pos := pf.FSet.Position(ts.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "NAM-05",
					Severity: diag.SeverityWarning,
					Message:  "interface name should generally use the -er suffix",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer names like Reader, Writer, or other -er forms when natural",
				})
			}
		}
	}

	return diagnostics, nil
}
