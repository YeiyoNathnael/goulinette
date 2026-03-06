package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type res02Rule struct{}

func NewRES02() Rule {
	return res02Rule{}
}

func (res02Rule) ID() string {
	return "RES-02"
}

func (res02Rule) Chapter() int {
	return 16
}

func (res02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		loopDepth := 0
		funcLitDepth := 0
		stack := make([]ast.Node, 0, 64)

		ast.Inspect(pf.File, func(n ast.Node) bool {
			if n == nil {
				if len(stack) == 0 {
					return false
				}
				last := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				switch last.(type) {
				case *ast.ForStmt, *ast.RangeStmt:
					loopDepth--
				case *ast.FuncLit:
					funcLitDepth--
				}
				return false
			}

			stack = append(stack, n)
			switch n.(type) {
			case *ast.ForStmt, *ast.RangeStmt:
				loopDepth++

			case *ast.FuncLit:
				funcLitDepth++

			case *ast.DeferStmt:
				if loopDepth > 0 && funcLitDepth == 0 {
					pos := pf.FSet.Position(n.Pos())
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "RES-02",
						Severity: diag.SeverityWarning,
						Message:  "defer should not be used directly inside loops",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "close resources explicitly per iteration or extract loop body into helper function",
					})
				}
			}

			return true
		})
	}

	return diagnostics, nil
}
