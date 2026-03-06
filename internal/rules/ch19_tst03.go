package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type tst03Rule struct{}

const (
	tst03Chapter      = 19
	tst03SleepMessage = "time.Sleep in tests can create flaky synchronization"
	tst03SleepHint    = "prefer channels, sync.WaitGroup, or context deadlines over sleep-based synchronization"
)

// NewTST03 returns the TST03 rule implementation.
func NewTST03() Rule {
	return tst03Rule{}
}

// ID returns the rule identifier.
func (tst03Rule) ID() string {
	return ruleTST03
}

// Chapter returns the chapter number for this rule.
func (tst03Rule) Chapter() int {
	return tst03Chapter
}

// Run executes this rule against the provided context.
func (tst03Rule) Run(ctx Context) ([]diag.Finding, error) {
	diagnostics := make([]diag.Finding, 0)

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
					diagnostics = append(diagnostics, diag.Finding{
						RuleID:   ruleTST03,
						Severity: diag.SeverityError,
						Message:  tst03SleepMessage,
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     tst03SleepHint,
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
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleTST03,
				Severity: diag.SeverityError,
				Message:  tst03SleepMessage,
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     tst03SleepHint,
			})
			return true
		})
	}

	return diagnostics, nil
}
