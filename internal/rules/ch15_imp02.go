package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type imp02Rule struct{}

func NewIMP02() Rule {
	return imp02Rule{}
}

func (imp02Rule) ID() string {
	return "IMP-02"
}

func (imp02Rule) Chapter() int {
	return 15
}

func (imp02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		usedSelectors := make(map[string]bool)
		ast.Inspect(pf.File, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			usedSelectors[id.Name] = true
			return true
		})

		for _, imp := range pf.File.Imports {
			name := ""
			if imp.Name != nil {
				name = imp.Name.Name
			}

			if name == "_" || name == "." {
				continue
			}

			selector := name
			if selector == "" {
				selector = defaultImportName(imp.Path.Value)
			}

			if usedSelectors[selector] {
				continue
			}

			pos := pf.FSet.Position(imp.Path.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "IMP-02",
				Severity: diag.SeverityError,
				Message:  "unused import is forbidden",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "remove import " + imp.Path.Value + " or use it",
			})
		}
	}

	return diagnostics, nil
}
