package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type fun04Rule struct{}

const fun04Chapter = 5

// NewFUN04 returns the FUN04 rule implementation.
func NewFUN04() Rule {
	return fun04Rule{}
}

// ID returns the rule identifier.
func (fun04Rule) ID() string {
	return ruleFUN04
}

// Chapter returns the chapter number for this rule.
func (fun04Rule) Chapter() int {
	return fun04Chapter
}

// Run executes this rule against the provided context.
func (fun04Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	returnCounts := map[string]int{}
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name == nil {
				continue
			}
			returnCounts[fn.Name.Name] = countFuncReturns(fn.Type)
		}
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			exprStmt, ok := n.(*ast.ExprStmt)
			if !ok {
				return true
			}

			call, ok := exprStmt.X.(*ast.CallExpr)
			if !ok {
				return true
			}

			ident, ok := call.Fun.(*ast.Ident)
			if !ok {
				return true
			}

			returns, ok := returnCounts[ident.Name]
			if !ok || returns <= 0 {
				return true
			}

			pos := pf.FSet.Position(call.Lparen)
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleFUN04,
				Severity: diag.SeverityError,
				Message:  "function return values are ignored; explicitly discard unused values with _",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "capture returns (e.g., _, err := call())",
			})

			return true
		})
	}

	return diagnostics, nil
}

func countFuncReturns(ft *ast.FuncType) int {
	if ft == nil || ft.Results == nil {
		return 0
	}

	var count int
	for _, field := range ft.Results.List {
		if len(field.Names) == 0 {
			count++
			continue
		}
		count += len(field.Names)
	}

	return count
}
