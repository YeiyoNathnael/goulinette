package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type typ01Rule struct{}

const typ01Chapter = 7

// NewTYP01 returns the TYP01 rule implementation.
func NewTYP01() Rule {
	return typ01Rule{}
}

// ID returns the rule identifier.
func (typ01Rule) ID() string {
	return ruleTYP01
}

// Chapter returns the chapter number for this rule.
func (typ01Rule) Chapter() int {
	return typ01Chapter
}

// Run executes this rule against the provided context.
func (typ01Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			diagnostics = append(diagnostics, typ01DiagnosticsForDecls(pkg.Fset, pkg.TypesInfo, pkg.Types, syntaxFile.Decls)...)
		}
	}

	return diagnostics, nil
}

func typ01DiagnosticsForDecls(fset *token.FileSet, info *types.Info, currentPkg *types.Package, decls []ast.Decl) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, decl := range decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Type == nil || fn.Type.Params == nil {
			continue
		}

		params := pointerParams(fn, info, currentPkg)
		for _, p := range params {
			usage := analyzePointerParamUsage(fn.Body, p)
			if !usage.used || usage.mutated || usage.escaped {
				continue
			}

			pos := fset.Position(p.Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleTYP01,
				Severity: diag.SeverityWarning,
				Message:  "pointer parameter appears read-only; prefer value parameter unless mutation is needed",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "change parameter to value type or document mutation intent",
			})
		}
	}

	return diagnostics
}

func pointerParams(fn *ast.FuncDecl, info *types.Info, currentPkg *types.Package) []*ast.Ident {
	out := make([]*ast.Ident, 0)
	if fn.Type == nil || fn.Type.Params == nil {
		return out
	}

	for _, field := range fn.Type.Params.List {
		for _, name := range field.Names {
			if name == nil || name.Name == "" {
				continue
			}
			t := info.TypeOf(name)
			if t == nil {
				continue
			}
			if isLocalStructPointerType(t, currentPkg) {
				out = append(out, name)
			}
		}
	}

	return out
}

func isLocalStructPointerType(t types.Type, currentPkg *types.Package) bool {
	if t == nil || currentPkg == nil {
		return false
	}

	ptr, ok := t.Underlying().(*types.Pointer)
	if !ok {
		return false
	}

	elem := ptr.Elem()
	named, ok := elem.(*types.Named)
	if !ok {
		return false
	}
	if named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	if named.Obj().Pkg().Path() != currentPkg.Path() {
		return false
	}

	_, isStruct := named.Underlying().(*types.Struct)
	return isStruct
}

type pointerUsage struct {
	used    bool
	mutated bool
	escaped bool
}

func analyzePointerParamUsage(body *ast.BlockStmt, param *ast.Ident) pointerUsage {
	usage := pointerUsage{}
	if body == nil || param == nil {
		return usage
	}

	ast.Inspect(body, func(n ast.Node) bool {
		if usage.mutated || usage.escaped {
			return true
		}

		switch x := n.(type) {
		case *ast.Ident:
			markPointerIdentUsage(&usage, x, param)
		case *ast.AssignStmt:
			checkPointerAssignmentMutation(&usage, x, param)
			if usage.mutated {
				return true
			}
		case *ast.IncDecStmt:
			if writesThroughPointer(x.X, param) {
				usage.mutated = true
				return true
			}
		case *ast.CallExpr:
			if callUsesPointerParam(x, param) {
				usage.escaped = true
				return true
			}
		default:
			// no-op
		}

		return true
	})

	return usage
}

func markPointerIdentUsage(usage *pointerUsage, id *ast.Ident, param *ast.Ident) {
	if usage == nil || id == nil || param == nil {
		return
	}
	if id.Obj == param.Obj {
		usage.used = true
	}
}

func checkPointerAssignmentMutation(usage *pointerUsage, stmt *ast.AssignStmt, param *ast.Ident) {
	if usage == nil || stmt == nil || param == nil {
		return
	}
	for _, lhs := range stmt.Lhs {
		if writesThroughPointer(lhs, param) {
			usage.mutated = true
			return
		}
	}
}

func writesThroughPointer(expr ast.Expr, param *ast.Ident) bool {
	switch e := expr.(type) {
	case *ast.StarExpr:
		id, ok := e.X.(*ast.Ident)
		return ok && id.Obj == param.Obj
	case *ast.SelectorExpr:
		id, ok := e.X.(*ast.Ident)
		return ok && id.Obj == param.Obj
	default:
		return false
	}
}

func callUsesPointerParam(call *ast.CallExpr, param *ast.Ident) bool {
	for _, arg := range call.Args {
		if id, ok := arg.(*ast.Ident); ok && id.Obj == param.Obj {
			return true
		}
	}

	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if id, ok := sel.X.(*ast.Ident); ok && id.Obj == param.Obj {
			return true
		}
	}

	return false
}
