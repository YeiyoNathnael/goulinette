package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type tst01Rule struct{}

const (
	tst01Chapter            = 19
	tst01MinGroupedTests    = 3
	tst01MessageUnrolled    = "multiple subtests are manually unrolled; use table-driven tests with a range loop"
	tst01HintUnrolled       = "create a tests table slice and iterate with for _, tc := range tests { t.Run(...) }"
	tst01HintBySubject      = "combine related Test* variants into one table-driven test with t.Run"
	tst01SubjectMessageBase = "test variations for "
	tst01SubjectMessageSfx  = " should use a table-driven pattern"
)

// NewTST01 returns the TST01 rule implementation.
func NewTST01() Rule {
	return tst01Rule{}
}

// ID returns the rule identifier.
func (tst01Rule) ID() string {
	return ruleTST01
}

// Chapter returns the chapter number for this rule.
func (tst01Rule) Chapter() int {
	return tst01Chapter
}

// Run executes this rule against the provided context.
func (tst01Rule) Run(ctx Context) ([]diag.Finding, error) {
	files, err := collectTestFiles(ctx)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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

			var topLevelTRuns int
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

			if topLevelTRuns >= tst01MinGroupedTests {
				pos := tf.FSet.Position(fn.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleTST01,
					Severity: diag.SeverityError,
					Message:  tst01MessageUnrolled,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     tst01HintUnrolled,
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
			if len(group) < tst01MinGroupedTests {
				continue
			}
			for _, ti := range group {
				pos := tf.FSet.Position(ti.Node.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleTST01,
					Severity: diag.SeverityError,
					Message:  tst01SubjectMessageBase + subject + tst01SubjectMessageSfx,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     tst01HintBySubject,
				})
			}
		}
	}

	return diagnostics, nil
}
