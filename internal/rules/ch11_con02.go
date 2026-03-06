package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type con02Rule struct{}

const (
	con02Chapter          = 11
	con02DoneMethodName   = "Done"
	con02ContextPkgPath   = "context"
	con02ContextTypeName  = "Context"
	con02CancelSignalMsg  = "goroutine has no obvious cancellation or exit path"
	con02CancelSignalHint = "pass context.Context and/or use select with ctx.Done() or an explicit done channel"
)

// NewCON02 returns the CON02 rule implementation.
func NewCON02() Rule {
	return con02Rule{}
}

// ID returns the rule identifier.
func (con02Rule) ID() string {
	return ruleCON02
}

// Chapter returns the chapter number for this rule.
func (con02Rule) Chapter() int {
	return con02Chapter
}

// Run executes this rule against the provided context.
func (con02Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleCON02,
					Severity: diag.SeverityError,
					Message:  con02CancelSignalMsg,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     con02CancelSignalHint,
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
	var uses bool
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
	var found bool
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
	default:
		// no-op
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
	if !ok || sel.Sel == nil || sel.Sel.Name != con02DoneMethodName {
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
		if obj != nil && obj.Pkg() != nil && obj.Pkg().Path() == con02ContextPkgPath && obj.Name() == con02ContextTypeName {
			return true
		}
	}
	if iface, ok := t.Underlying().(*types.Interface); ok {
		for i := 0; i < iface.NumMethods(); i++ {
			m := iface.Method(i)
			if m.Name() != con02DoneMethodName {
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
