package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type tst02Rule struct{}

func NewTST02() Rule {
	return tst02Rule{}
}

func (tst02Rule) ID() string {
	return "TST-02"
}

func (tst02Rule) Chapter() int {
	return 19
}

func (tst02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	files, err := collectTestFiles(ctx)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, tf := range files {
		helperCandidates := make(map[string]tstFuncInfo)
		topLevelCallers := make([]*ast.FuncDecl, 0)

		for _, d := range tf.File.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Name == nil || fn.Body == nil {
				continue
			}
			params := testingParamNames(fn.Type, tf.ImportPath)
			if len(params) == 0 {
				continue
			}

			if isTopLevelTestLike(fn.Name.Name) {
				topLevelCallers = append(topLevelCallers, fn)
			} else if callsTestingMethods(fn.Body, params) {
				helperCandidates[fn.Name.Name] = tstFuncInfo{
					Node:         fn,
					Name:         fn.Name.Name,
					Path:         tf.Path,
					ImportPath:   tf.ImportPath,
					TestingParam: params,
				}
			}

			tst02CheckSubtestLiterals(fn.Body, params, tf, &diagnostics)
		}

		helperUsedByTests := make(map[string]bool)
		for _, caller := range topLevelCallers {
			ast.Inspect(caller.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				id, ok := call.Fun.(*ast.Ident)
				if !ok {
					return true
				}
				if _, ok := helperCandidates[id.Name]; ok {
					helperUsedByTests[id.Name] = true
				}
				return true
			})
		}

		for name, info := range helperCandidates {
			if !helperUsedByTests[name] {
				continue
			}
			if firstStmtIsHelper(info.Node.Body, info.TestingParam) {
				continue
			}
			pos := tf.FSet.Position(info.Node.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "TST-02",
				Severity: diag.SeverityError,
				Message:  "test helper must call t.Helper() as its first statement",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "add t.Helper() as the first statement in helper functions using testing diagnostics",
			})
		}
	}

	return diagnostics, nil
}

func tst02CheckSubtestLiterals(body *ast.BlockStmt, params map[string]bool, tf tstFile, diagnostics *[]diag.Diagnostic) {
	if body == nil || len(params) == 0 {
		return
	}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isTRunCall(call, params) || len(call.Args) < 2 {
			return true
		}
		lit, ok := call.Args[1].(*ast.FuncLit)
		if !ok || lit.Type == nil || lit.Body == nil {
			return true
		}
		litParams := testingParamNames(lit.Type, tf.ImportPath)
		if len(litParams) == 0 || !callsTestingMethods(lit.Body, litParams) {
			return true
		}
		if firstStmtIsHelper(lit.Body, litParams) {
			return true
		}
		pos := tf.FSet.Position(lit.Pos())
		*diagnostics = append(*diagnostics, diag.Diagnostic{
			RuleID:   "TST-02",
			Severity: diag.SeverityError,
			Message:  "subtest function must call t.Helper() as its first statement",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "place t.Helper() first inside func(t *testing.T) passed to t.Run",
		})
		return true
	})
}
