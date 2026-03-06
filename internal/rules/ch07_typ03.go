package rules

import (
	"go/ast"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ03Rule struct{}

func NewTYP03() Rule {
	return typ03Rule{}
}

func (typ03Rule) ID() string {
	return "TYP-03"
}

func (typ03Rule) Chapter() int {
	return 7
}

func (typ03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, idxCtx := range collectMapReadsWithoutCommaOK(syntaxFile, pkg.TypesInfo) {
				if !hasMeaningfulZeroValue(idxCtx.valueType) {
					continue
				}

				pos := pkg.Fset.Position(idxCtx.indexExpr.Lbrack)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "TYP-03",
					Severity: diag.SeverityError,
					Message:  "map reads should use comma-ok form when zero value can be a meaningful result",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "use v, ok := m[key] and check ok",
				})
			}
		}
	}

	return diagnostics, nil
}

type mapReadContext struct {
	indexExpr *ast.IndexExpr
	valueType types.Type
}

func collectMapReadsWithoutCommaOK(file *ast.File, info *types.Info) []mapReadContext {
	out := make([]mapReadContext, 0)
	stack := make([]ast.Node, 0, 32)

	ast.Inspect(file, func(n ast.Node) bool {
		if n == nil {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			return false
		}

		idx, ok := n.(*ast.IndexExpr)
		if ok {
			mtyp, ok := info.TypeOf(idx.X).Underlying().(*types.Map)
			if ok {
				if isMapWriteContext(idx, stack) || isMapCommaOKRead(idx, stack) {
					stack = append(stack, n)
					return true
				}
				out = append(out, mapReadContext{indexExpr: idx, valueType: mtyp.Elem()})
			}
		}

		stack = append(stack, n)
		return true
	})

	return out
}

func isMapWriteContext(idx *ast.IndexExpr, ancestors []ast.Node) bool {
	for i := len(ancestors) - 1; i >= 0; i-- {
		as, ok := ancestors[i].(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, lhs := range as.Lhs {
			if lhs == idx {
				return true
			}
		}
		return false
	}
	return false
}

func isMapCommaOKRead(idx *ast.IndexExpr, ancestors []ast.Node) bool {
	for i := len(ancestors) - 1; i >= 0; i-- {
		as, ok := ancestors[i].(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, rhs := range as.Rhs {
			if rhs == idx && len(as.Lhs) >= 2 {
				return true
			}
		}
		return false
	}
	return false
}

func hasMeaningfulZeroValue(t types.Type) bool {
	if t == nil {
		return false
	}

	switch u := t.Underlying().(type) {
	case *types.Basic:
		return true
	case *types.Struct, *types.Array:
		return true
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Interface, *types.Signature:
		return false
	default:
		_ = u
		return true
	}
}
