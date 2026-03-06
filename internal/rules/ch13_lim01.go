package rules

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type lim01Rule struct{}

const (
	lim01Chapter          = 13
	lim01MaxFunctionLines = 50
	lim01Message          = "functions must not exceed 50 lines"
	lim01TestFileSuffix   = "_test.go"
)

// NewLIM01 returns the LIM01 rule implementation.
func NewLIM01() Rule {
	return lim01Rule{}
}

// ID returns the rule identifier.
func (lim01Rule) ID() string {
	return ruleLIM01
}

// Chapter returns the chapter number for this rule.
func (lim01Rule) Chapter() int {
	return lim01Chapter
}

// Run executes this rule against the provided context.
func (lim01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		isTestFile := strings.HasSuffix(pf.Path, lim01TestFileSuffix)
		ast.Inspect(pf.File, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				if shouldSkipLIM01FuncDecl(fn, isTestFile) {
					return true
				}
				if fn.Body == nil {
					return true
				}
				count := functionBodyLineCount(pf.FSet, fn.Body)
				if count <= lim01MaxFunctionLines {
					return true
				}

				pos := pf.FSet.Position(fn.Name.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleLIM01,
					Severity: diag.SeverityError,
					Message:  lim01Message,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "function body has " + strconv.Itoa(count) + " lines; extract helper functions",
				})

			case *ast.FuncLit:
				if isTestFile {
					return true
				}
				if fn.Body == nil {
					return true
				}
				count := functionBodyLineCount(pf.FSet, fn.Body)
				if count <= lim01MaxFunctionLines {
					return true
				}

				pos := pf.FSet.Position(fn.Type.Func)
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleLIM01,
					Severity: diag.SeverityError,
					Message:  lim01Message,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "anonymous function body has " + strconv.Itoa(count) + " lines; extract helper function",
				})
			default:
				// no-op
			}

			return true
		})
	}

	return diagnostics, nil
}

func shouldSkipLIM01FuncDecl(fn *ast.FuncDecl, isTestFile bool) bool {
	if fn == nil {
		return true
	}
	if isTestFile {
		return true
	}
	if fn.Recv != nil && fn.Name != nil && fn.Name.Name == "Run" {
		return true
	}
	return false
}

func functionBodyLineCount(fset *token.FileSet, body *ast.BlockStmt) int {
	if body == nil {
		return 0
	}
	count := rawBodyLineCount(fset, body)

	var nestedFuncLitBodyLines int
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
