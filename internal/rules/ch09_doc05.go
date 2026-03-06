package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type doc05Rule struct{}

func NewDOC05() Rule {
	return doc05Rule{}
}

func (doc05Rule) ID() string {
	return "DOC-05"
}

func (doc05Rule) Chapter() int {
	return 9
}

func (doc05Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	packageVars := collectPackageVarNames(parsed)
	diagnostics := make([]diag.Diagnostic, 0)

	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Name.Name != "init" {
				continue
			}

			if isImmutableInitBody(fn.Body, packageVars) {
				continue
			}

			pos := pf.FSet.Position(fn.Name.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "DOC-05",
				Severity: diag.SeverityWarning,
				Message:  "init() should be avoided except for immutable setup",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "move side effects to explicit initialization code",
			})
		}
	}

	return diagnostics, nil
}
