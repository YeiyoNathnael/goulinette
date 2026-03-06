package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type tst01Rule struct{}

func NewTST01() Rule {
	return tst01Rule{}
}

func (tst01Rule) ID() string {
	return "TST-01"
}

func (tst01Rule) Chapter() int {
	return 19
}

func (tst01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	files, err := collectTestFiles(ctx)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, tf := range files {
		tests := make([]tstFuncInfo, 0)
		for _, d := range tf.File.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Body == nil {
				continue
			}
			if !isRealTest(fn.Name.Name) {
				continue
			}
			params := testingParamNames(fn.Type, tf.ImportPath)
			if len(params) == 0 {
				continue
			}
			tests = append(tests, tstFuncInfo{
				Node:         fn,
				Name:         fn.Name.Name,
				Path:         tf.Path,
				ImportPath:   tf.ImportPath,
				TestingParam: params,
				Subject:      testSubject(fn.Name.Name),
			})

			topLevelTRuns := 0
			for _, stmt := range fn.Body.List {
				expr, ok := stmt.(*ast.ExprStmt)
				if !ok {
					continue
				}
				call, ok := expr.X.(*ast.CallExpr)
				if !ok || !isTRunCall(call, params) {
					continue
				}
				topLevelTRuns++
			}

			if topLevelTRuns >= 3 {
				pos := tf.FSet.Position(fn.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "TST-01",
					Severity: diag.SeverityError,
					Message:  "multiple subtests are manually unrolled; use table-driven tests with a range loop",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "create a tests table slice and iterate with for _, tc := range tests { t.Run(...) }",
				})
			}
		}

		bySubject := make(map[string][]tstFuncInfo)
		for _, ti := range tests {
			if ti.Subject == "" {
				continue
			}
			bySubject[ti.Subject] = append(bySubject[ti.Subject], ti)
		}

		for subject, group := range bySubject {
			if len(group) < 3 {
				continue
			}
			for _, ti := range group {
				pos := tf.FSet.Position(ti.Node.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "TST-01",
					Severity: diag.SeverityError,
					Message:  "test variations for " + subject + " should use a table-driven pattern",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "combine related Test* variants into one table-driven test with t.Run",
				})
			}
		}
	}

	return diagnostics, nil
}
