package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ06Rule struct{}

const typ06Chapter = 7

// NewTYP06 returns the TYP06 rule implementation.
func NewTYP06() Rule {
	return typ06Rule{}
}

// ID returns the rule identifier.
func (typ06Rule) ID() string {
	return ruleTYP06
}

// Chapter returns the chapter number for this rule.
func (typ06Rule) Chapter() int {
	return typ06Chapter
}

// Run executes this rule against the provided context.
func (typ06Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			iface, ok := n.(*ast.InterfaceType)
			if !ok || iface.Methods == nil || len(iface.Methods.List) != 0 {
				return true
			}

			pos := pf.FSet.Position(iface.Interface)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleTYP06,
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
