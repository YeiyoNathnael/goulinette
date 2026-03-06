package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type con02Rule struct{}

func NewCON02() Rule {
	return con02Rule{}
}

func (con02Rule) ID() string {
	return "CON-02"
}

func (con02Rule) Chapter() int {
	return 11
}

func (con02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			ast.Inspect(file, func(n ast.Node) bool {
				goStmt, ok := n.(*ast.GoStmt)
				if !ok {
					return true
				}

				if hasVisibleGoroutineExitSignal(goStmt.Call, pkg.TypesInfo) {
					return true
				}

				pos := pkg.Fset.Position(goStmt.Go)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "CON-02",
					Severity: diag.SeverityError,
					Message:  "goroutine has no obvious cancellation or exit path",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "pass context.Context and/or use select with ctx.Done() or an explicit done channel",
				})

				return true
			})
		}
	}

	return diagnostics, nil
}

func hasVisibleGoroutineExitSignal(call *ast.CallExpr, info *types.Info) bool {
	if call == nil || info == nil {
		return false
	}

	if callHasContextParam(call, info) {
		return true
	}
	if callHasContextArgument(call, info) {
		return true
	}

	fnLit, ok := call.Fun.(*ast.FuncLit)
	if !ok || fnLit.Body == nil {
		return false
	}

	if functionBodyUsesContext(fnLit.Body, info) {
		return true
	}
	if functionBodyHasDoneSelect(fnLit.Body, info) {
		return true
	}

	return false
}

func callHasContextParam(call *ast.CallExpr, info *types.Info) bool {
	t := info.TypeOf(call.Fun)
	sig, ok := t.(*types.Signature)
	if !ok || sig.Params() == nil {
		return false
	}

	for i := 0; i < sig.Params().Len(); i++ {
		if isContextType(sig.Params().At(i).Type()) {
			return true
		}
	}

	return false
}

func callHasContextArgument(call *ast.CallExpr, info *types.Info) bool {
	for _, arg := range call.Args {
		if isContextType(info.TypeOf(arg)) {
			return true
		}
	}
	return false
}

func functionBodyUsesContext(body *ast.BlockStmt, info *types.Info) bool {
	uses := false
	ast.Inspect(body, func(n ast.Node) bool {
		if uses {
			return false
		}
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if isContextType(info.TypeOf(id)) {
			uses = true
			return false
		}
		return true
	})
	return uses
}

func functionBodyHasDoneSelect(body *ast.BlockStmt, info *types.Info) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectStmt)
		if !ok || sel.Body == nil {
			return true
		}

		for _, stmt := range sel.Body.List {
			cc, ok := stmt.(*ast.CommClause)
			if !ok || cc.Comm == nil {
				continue
			}
			if commUsesContextDone(cc.Comm, info) {
				found = true
				return false
			}
		}

		return true
	})
	return found
}

func commUsesContextDone(comm ast.Stmt, info *types.Info) bool {
	var recv ast.Expr

	switch c := comm.(type) {
	case *ast.AssignStmt:
		if len(c.Rhs) == 1 {
			recv = c.Rhs[0]
		}
	case *ast.ExprStmt:
		recv = c.X
	}

	un, ok := recv.(*ast.UnaryExpr)
	if !ok || un.Op != token.ARROW {
		return false
	}

	call, ok := un.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Done" {
		return false
	}

	return isContextType(info.TypeOf(sel.X))
}

func isContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == "context" && obj.Name() == "Context" {
			return true
		}
	}
	if iface, ok := t.Underlying().(*types.Interface); ok {
		for i := 0; i < iface.NumMethods(); i++ {
			m := iface.Method(i)
			if m.Name() != "Done" {
				continue
			}
			sig, ok := m.Type().(*types.Signature)
			if !ok || sig.Results() == nil || sig.Results().Len() != 1 {
				continue
			}
			if _, ok := sig.Results().At(0).Type().(*types.Chan); ok {
				return true
			}
		}
	}
	return false
}
