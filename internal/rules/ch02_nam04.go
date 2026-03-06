package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam04Rule struct{}

func NewNAM04() Rule {
	return nam04Rule{}
}

func (nam04Rule) ID() string {
	return "NAM-04"
}

func (nam04Rule) Chapter() int {
	return 2
}

func (nam04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || (gd.Tok != token.VAR && gd.Tok != token.CONST) {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, n := range vs.Names {
					if n == nil || n.Name == "_" {
						continue
					}
					if len([]rune(n.Name)) > 2 {
						continue
					}
					pos := pf.FSet.Position(n.Pos())
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "NAM-04",
						Severity: diag.SeverityWarning,
						Message:  "package-level identifier name is too short to be descriptive",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "use full descriptive names for package-level identifiers",
					})
				}
			}
		}
	}

	return diagnostics, nil
}
