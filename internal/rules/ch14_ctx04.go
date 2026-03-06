package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctx04Rule struct{}

func NewCTX04() Rule {
	return ctx04Rule{}
}

func (ctx04Rule) ID() string {
	return "CTX-04"
}

func (ctx04Rule) Chapter() int {
	return 14
}

func (ctx04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
					state := make(cancelState)
					out, ds := analyzeCTX04StmtList(fn.Body.List, state, pkg.TypesInfo, pkg.Fset)
					diagnostics = append(diagnostics, ds...)
					diagnostics = append(diagnostics, ctx04UnhandledAtExit(pkg.Fset, fn.Body.Rbrace, out)...)
				}
			}

			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.FuncLit)
				if !ok || lit.Body == nil {
					return true
				}
				state := make(cancelState)
				out, ds := analyzeCTX04StmtList(lit.Body.List, state, pkg.TypesInfo, pkg.Fset)
				diagnostics = append(diagnostics, ds...)
				diagnostics = append(diagnostics, ctx04UnhandledAtExit(pkg.Fset, lit.Body.Rbrace, out)...)
				return true
			})
		}
	}

	return diagnostics, nil
}

type cancelState map[*types.Var]bool

func cloneCancelState(in cancelState) cancelState {
	out := make(cancelState, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mergeCancelState(thenState, elseState cancelState) cancelState {
	out := make(cancelState)
	for k, v := range thenState {
		if ev, ok := elseState[k]; ok {
			out[k] = v && ev
			continue
		}
		out[k] = v
	}
	for k, v := range elseState {
		if _, seen := out[k]; !seen {
			out[k] = v
		}
	}
	return out
}

func analyzeCTX04StmtList(stmts []ast.Stmt, in cancelState, info *types.Info, fset *token.FileSet) (cancelState, []diag.Diagnostic) {
	state := cloneCancelState(in)
	diagnostics := make([]diag.Diagnostic, 0)

	for _, stmt := range stmts {
		var ds []diag.Diagnostic
		state, ds = analyzeCTX04Stmt(stmt, state, info, fset)
		diagnostics = append(diagnostics, ds...)
	}

	return state, diagnostics
}

func analyzeCTX04Stmt(stmt ast.Stmt, in cancelState, info *types.Info, fset *token.FileSet) (cancelState, []diag.Diagnostic) {
	state := cloneCancelState(in)
	diagnostics := make([]diag.Diagnostic, 0)

	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return analyzeCTX04StmtList(s.List, state, info, fset)

	case *ast.AssignStmt:
		diagnostics = append(diagnostics, ctx04TrackCreationFromAssign(s, state, info, fset)...)
		if ctx04MarksCancelHandledByDirectCall(s.Rhs, state, info) {
			// no-op: state is updated in helper
		}
		return state, diagnostics

	case *ast.DeclStmt:
		gd, ok := s.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			return state, diagnostics
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, rhs := range vs.Values {
				call, ok := rhs.(*ast.CallExpr)
				if !ok || !isContextDerivationCall(call, info) {
					continue
				}
				if i >= len(vs.Names)-1 {
					continue
				}
				cancelName := vs.Names[i+1]
				if cancelName == nil {
					continue
				}
				if cancelName.Name == "_" {
					pos := fset.Position(cancelName.Pos())
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "CTX-04",
						Severity: diag.SeverityError,
						Message:  "cancel function from derived context must be handled",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "capture cancel function and defer/call it on all paths",
					})
					continue
				}
				if obj, ok := info.Defs[cancelName].(*types.Var); ok {
					state[obj] = false
				}
			}
		}
		return state, diagnostics

	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			ctx04MarkCancelHandledByCall(call, state, info, false)
		}
		return state, diagnostics

	case *ast.DeferStmt:
		ctx04MarkCancelHandledByCall(s.Call, state, info, true)
		return state, diagnostics

	case *ast.GoStmt:
		return state, diagnostics

	case *ast.ReturnStmt:
		for cancelVar, handled := range state {
			if handled {
				continue
			}
			pos := fset.Position(s.Return)
			name := cancelVar.Name()
			if name == "" {
				name = "cancel"
			}
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "CTX-04",
				Severity: diag.SeverityError,
				Message:  "derived context cancel function must be called or deferred on all exit paths",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "defer " + name + "() immediately after context.WithCancel/WithTimeout/WithDeadline",
			})
		}
		return state, diagnostics

	case *ast.IfStmt:
		if s.Init != nil {
			var initDiags []diag.Diagnostic
			state, initDiags = analyzeCTX04Stmt(s.Init, state, info, fset)
			diagnostics = append(diagnostics, initDiags...)
		}

		thenState, thenDiags := analyzeCTX04StmtList(s.Body.List, cloneCancelState(state), info, fset)
		diagnostics = append(diagnostics, thenDiags...)

		elseState := cloneCancelState(state)
		if s.Else != nil {
			var elseDiags []diag.Diagnostic
			elseState, elseDiags = analyzeCTX04Stmt(s.Else, elseState, info, fset)
			diagnostics = append(diagnostics, elseDiags...)
		}

		return mergeCancelState(thenState, elseState), diagnostics
	}

	return state, diagnostics
}

func ctx04TrackCreationFromAssign(as *ast.AssignStmt, state cancelState, info *types.Info, fset *token.FileSet) []diag.Diagnostic {
	if as == nil || info == nil {
		return nil
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for i, rhs := range as.Rhs {
		call, ok := rhs.(*ast.CallExpr)
		if !ok || !isContextDerivationCall(call, info) {
			continue
		}

		if i >= len(as.Lhs)-1 {
			continue
		}

		cancelExpr := as.Lhs[i+1]
		cancelIdent, ok := cancelExpr.(*ast.Ident)
		if !ok {
			continue
		}
		if cancelIdent.Name == "_" {
			pos := fset.Position(cancelIdent.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "CTX-04",
				Severity: diag.SeverityError,
				Message:  "cancel function from derived context must be handled",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "capture cancel function and defer/call it on all paths",
			})
			continue
		}

		var obj *types.Var
		if as.Tok == token.DEFINE {
			obj, _ = info.Defs[cancelIdent].(*types.Var)
		} else {
			obj, _ = info.ObjectOf(cancelIdent).(*types.Var)
		}
		if obj != nil {
			state[obj] = false
		}
	}

	return diagnostics
}

func ctx04UnhandledAtExit(fset *token.FileSet, posToken token.Pos, state cancelState) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)
	for cancelVar, handled := range state {
		if handled {
			continue
		}
		pos := fset.Position(posToken)
		name := cancelVar.Name()
		if name == "" {
			name = "cancel"
		}
		diagnostics = append(diagnostics, diag.Diagnostic{
			RuleID:   "CTX-04",
			Severity: diag.SeverityError,
			Message:  "derived context cancel function must be called or deferred on all exit paths",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "defer " + name + "() immediately after context.WithCancel/WithTimeout/WithDeadline",
		})
	}
	return diagnostics
}

func ctx04MarksCancelHandledByDirectCall(exprs []ast.Expr, state cancelState, info *types.Info) bool {
	updated := false
	for _, expr := range exprs {
		call, ok := expr.(*ast.CallExpr)
		if !ok {
			continue
		}
		if ctx04MarkCancelHandledByCall(call, state, info, false) {
			updated = true
		}
	}
	return updated
}

func ctx04MarkCancelHandledByCall(call *ast.CallExpr, state cancelState, info *types.Info, fromDefer bool) bool {
	if call == nil || info == nil {
		return false
	}

	updated := false

	if id, ok := call.Fun.(*ast.Ident); ok {
		if obj, ok := info.ObjectOf(id).(*types.Var); ok {
			if _, tracked := state[obj]; tracked {
				state[obj] = true
				updated = true
			}
		}
	}

	if fromDefer {
		for _, arg := range call.Args {
			id, ok := arg.(*ast.Ident)
			if !ok {
				continue
			}
			obj, ok := info.ObjectOf(id).(*types.Var)
			if !ok {
				continue
			}
			if _, tracked := state[obj]; tracked {
				state[obj] = true
				updated = true
			}
		}
	}

	return updated
}

func isContextDerivationCall(call *ast.CallExpr, info *types.Info) bool {
	if call == nil || info == nil {
		return false
	}

	obj := calledFunctionObject(call, info)
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	if obj.Pkg().Path() != "context" {
		return false
	}
	name := obj.Name()
	return name == "WithCancel" || name == "WithTimeout" || name == "WithDeadline"
}

func calledFunctionObject(call *ast.CallExpr, info *types.Info) *types.Func {
	if call == nil || info == nil {
		return nil
	}

	switch fn := call.Fun.(type) {
	case *ast.SelectorExpr:
		if obj, ok := info.Uses[fn.Sel].(*types.Func); ok {
			return obj
		}
	case *ast.Ident:
		if obj, ok := info.Uses[fn].(*types.Func); ok {
			return obj
		}
	}

	return nil
}
