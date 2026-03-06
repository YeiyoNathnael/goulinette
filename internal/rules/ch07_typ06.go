package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ06Rule struct{}

func NewTYP06() Rule {
	return typ06Rule{}
}

func (typ06Rule) ID() string {
	return "TYP-06"
}

func (typ06Rule) Chapter() int {
	return 7
}

func (typ06Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			iface, ok := n.(*ast.InterfaceType)
			if !ok || iface.Methods == nil || len(iface.Methods.List) != 0 {
				return true
			}

			pos := pf.FSet.Position(iface.Interface)
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "TYP-06",
				Severity: diag.SeverityError,
				Message:  "interface{} is forbidden; use any",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "replace interface{} with any",
			})

			return true
		})
	}

	return diagnostics, nil
}
