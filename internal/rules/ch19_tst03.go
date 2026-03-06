package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type tst03Rule struct{}

func NewTST03() Rule {
	return tst03Rule{}
}

func (tst03Rule) ID() string {
	return "TST-03"
}

func (tst03Rule) Chapter() int {
	return 19
}

func (tst03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	diagnostics := make([]diag.Diagnostic, 0)

	if ctx.Root != "" {
		pkgs, err := loadTypedPackagesWithTests(ctx.Root)
		if err != nil {
			return nil, err
		}
		for _, pkg := range pkgs {
			if pkg == nil || pkg.TypesInfo == nil || pkg.Fset == nil {
				continue
			}
			for _, file := range pkg.Syntax {
				if file == nil {
					continue
				}
				filename := pkg.Fset.Position(file.Pos()).Filename
				if !isTestFile(filename) {
					continue
				}
				ast.Inspect(file, func(n ast.Node) bool {
					call, ok := n.(*ast.CallExpr)
					if !ok || !isTimeSleepCallTyped(call, pkg.TypesInfo) {
						return true
					}
					pos := pkg.Fset.Position(call.Pos())
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "TST-03",
						Severity: diag.SeverityError,
						Message:  "time.Sleep in tests can create flaky synchronization",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "prefer channels, sync.WaitGroup, or context deadlines over sleep-based synchronization",
					})
					return true
				})
			}
		}
		return diagnostics, nil
	}

	files, err := collectTestFiles(ctx)
	if err != nil {
		return nil, err
	}
	for _, tf := range files {
		ast.Inspect(tf.File, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !isTimeSleepCallAST(call, tf.ImportPath, tf.DotImports) {
				return true
			}
			pos := tf.FSet.Position(call.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "TST-03",
				Severity: diag.SeverityError,
				Message:  "time.Sleep in tests can create flaky synchronization",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "prefer channels, sync.WaitGroup, or context deadlines over sleep-based synchronization",
			})
			return true
		})
	}

	return diagnostics, nil
}
