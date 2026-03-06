package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type res02Rule struct{}

const (
	res02Chapter = 16
)

// NewRES02 returns the RES02 rule implementation.
func NewRES02() Rule {
	return res02Rule{}
}

// ID returns the rule identifier.
func (res02Rule) ID() string {
	return ruleRES02
}

// Chapter returns the chapter number for this rule.
func (res02Rule) Chapter() int {
	return res02Chapter
}

// Run executes this rule against the provided context.
func (res02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		var loopDepth int
		var funcLitDepth int
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
				default:
					// no-op
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
					diagnostics = append(diagnostics, diag.Finding{
						RuleID:   ruleRES02,
						Severity: diag.SeverityWarning,
						Message:  "defer should not be used directly inside loops",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "close resources explicitly per iteration or extract loop body into helper function",
					})
				}
			default:
				// no-op
			}

			return true
		})
	}

	return diagnostics, nil
}
