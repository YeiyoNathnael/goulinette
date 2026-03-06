package rules

import (
	"go/ast"
	"go/token"
	"strconv"

	"goulinette/internal/diag"
)

type lim01Rule struct{}

func NewLIM01() Rule {
	return lim01Rule{}
}

func (lim01Rule) ID() string {
	return "LIM-01"
}

func (lim01Rule) Chapter() int {
	return 13
}

func (lim01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				if fn.Body == nil {
					return true
				}
				count := functionBodyLineCount(pf.FSet, fn.Body)
				if count <= 50 {
					return true
				}

				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "LIM-01",
					Severity: diag.SeverityError,
					Message:  "functions must not exceed 50 lines",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "function body has " + strconv.Itoa(count) + " lines; extract helper functions",
				})

			case *ast.FuncLit:
				if fn.Body == nil {
					return true
				}
				count := functionBodyLineCount(pf.FSet, fn.Body)
				if count <= 50 {
					return true
				}

				pos := pf.FSet.Position(fn.Type.Func)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "LIM-01",
					Severity: diag.SeverityError,
					Message:  "functions must not exceed 50 lines",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "anonymous function body has " + strconv.Itoa(count) + " lines; extract helper function",
				})
			}

			return true
		})
	}

	return diagnostics, nil
}

func functionBodyLineCount(fset *token.FileSet, body *ast.BlockStmt) int {
	if body == nil {
		return 0
	}
	count := rawBodyLineCount(fset, body)

	nestedFuncLitBodyLines := 0
	ast.Inspect(body, func(n ast.Node) bool {
		lit, ok := n.(*ast.FuncLit)
		if !ok || lit.Body == nil {
			return true
		}
		nestedFuncLitBodyLines += rawBodyLineCount(fset, lit.Body)
		return true
	})

	count -= nestedFuncLitBodyLines
	if count < 0 {
		return 0
	}
	return count
}

func rawBodyLineCount(fset *token.FileSet, body *ast.BlockStmt) int {
	if body == nil {
		return 0
	}
	start := fset.Position(body.Lbrace).Line
	end := fset.Position(body.Rbrace).Line
	if end <= start {
		return 0
	}
	count := end - start - 1
	if count < 0 {
		return 0
	}
	return count
}
