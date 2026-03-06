package rules

import (
	"go/ast"
	"strconv"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type lim02Rule struct{}

const (
	lim02Chapter            = 13
	lim02MaxParameterCount  = 5
	lim02MessageTooManyArgs = "functions must not have more than 5 parameters"
)

// NewLIM02 returns the LIM02 rule implementation.
func NewLIM02() Rule {
	return lim02Rule{}
}

// ID returns the rule identifier.
func (lim02Rule) ID() string {
	return ruleLIM02
}

// Chapter returns the chapter number for this rule.
func (lim02Rule) Chapter() int {
	return lim02Chapter
}

// Run executes this rule against the provided context.
func (lim02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				count := functionParamCount(fn.Type)
				if count <= lim02MaxParameterCount {
					return true
				}
				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleLIM02,
					Severity: diag.SeverityError,
					Message:  lim02MessageTooManyArgs,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "function has " + strconv.Itoa(count) + " parameters; group related inputs into a config struct",
				})

			case *ast.FuncLit:
				count := functionParamCount(fn.Type)
				if count <= lim02MaxParameterCount {
					return true
				}
				pos := pf.FSet.Position(fn.Type.Func)
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleLIM02,
					Severity: diag.SeverityError,
					Message:  lim02MessageTooManyArgs,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "anonymous function has " + strconv.Itoa(count) + " parameters; reduce arguments",
				})
			default:
				// no-op
			}

			return true
		})
	}

	return diagnostics, nil
}

func functionParamCount(ft *ast.FuncType) int {
	if ft == nil || ft.Params == nil {
		return 0
	}

	var count int
	for _, field := range ft.Params.List {
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		count += n
	}

	return count
}
