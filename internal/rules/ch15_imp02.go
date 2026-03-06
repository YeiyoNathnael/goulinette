package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type imp02Rule struct{}

const imp02Chapter = 15

// NewIMP02 returns the IMP02 rule implementation.
func NewIMP02() Rule {
	return imp02Rule{}
}

// ID returns the rule identifier.
func (imp02Rule) ID() string {
	return ruleIMP02
}

// Chapter returns the chapter number for this rule.
func (imp02Rule) Chapter() int {
	return imp02Chapter
}

// Run executes this rule against the provided context.
func (imp02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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
			var name string
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

			if used, ok := usedSelectors[selector]; ok && used {
				continue
			}

			pos := pf.FSet.Position(imp.Path.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleIMP02,
				Severity: diag.SeverityError,
				Message:  "unused import is forbidden",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "remove import " + imp.Path.Value + " or use it",
			})
		}
	}

	return diagnostics, nil
}
