package rules

import (
	"go/ast"
	"strconv"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type lim02Rule struct{}

func NewLIM02() Rule {
	return lim02Rule{}
}

func (lim02Rule) ID() string {
	return "LIM-02"
}

func (lim02Rule) Chapter() int {
	return 13
}

func (lim02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				count := functionParamCount(fn.Type)
				if count <= 5 {
					return true
				}
				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "LIM-02",
					Severity: diag.SeverityError,
					Message:  "functions must not have more than 5 parameters",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "function has " + strconv.Itoa(count) + " parameters; group related inputs into a config struct",
				})

			case *ast.FuncLit:
				count := functionParamCount(fn.Type)
				if count <= 5 {
					return true
				}
				pos := pf.FSet.Position(fn.Type.Func)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "LIM-02",
					Severity: diag.SeverityError,
					Message:  "functions must not have more than 5 parameters",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "anonymous function has " + strconv.Itoa(count) + " parameters; reduce arguments",
				})
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

	count := 0
	for _, field := range ft.Params.List {
		n := len(field.Names)
		if n == 0 {
			n = 1
		}
		count += n
	}

	return count
}
