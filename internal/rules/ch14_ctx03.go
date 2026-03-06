package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"goulinette/internal/diag"
)

type ctx03Rule struct{}

func NewCTX03() Rule {
	return ctx03Rule{}
}

func (ctx03Rule) ID() string {
	return "CTX-03"
}

func (ctx03Rule) Chapter() int {
	return 14
}

func (ctx03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok && fn.Body != nil {
					state := make(ctxAssignState)
					_, ds := analyzeCTX03StmtList(fn.Body.List, state, pkg.TypesInfo, pkg.Fset)
					diagnostics = append(diagnostics, ds...)
				}
			}

			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.FuncLit)
				if !ok || lit.Body == nil {
					return true
				}
				state := make(ctxAssignState)
				_, ds := analyzeCTX03StmtList(lit.Body.List, state, pkg.TypesInfo, pkg.Fset)
				diagnostics = append(diagnostics, ds...)
				return true
			})
		}
	}

	return diagnostics, nil
}

type ctxAssignState map[*types.Var]bool

func cloneCtxAssignState(in ctxAssignState) ctxAssignState {
	out := make(ctxAssignState, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func intersectCtxAssignState(a, b ctxAssignState) ctxAssignState {
	out := make(ctxAssignState, len(a))
	for k, va := range a {
		vb, ok := b[k]
		out[k] = va && ok && vb
	}
	return out
}

func analyzeCTX03StmtList(stmts []ast.Stmt, in ctxAssignState, info *types.Info, fset *token.FileSet) (ctxAssignState, []diag.Diagnostic) {
	state := cloneCtxAssignState(in)
	diagnostics := make([]diag.Diagnostic, 0)

	for _, stmt := range stmts {
		var ds []diag.Diagnostic
		state, ds = analyzeCTX03Stmt(stmt, state, info, fset)
		diagnostics = append(diagnostics, ds...)
	}

	return state, diagnostics
}

func analyzeCTX03Stmt(stmt ast.Stmt, in ctxAssignState, info *types.Info, fset *token.FileSet) (ctxAssignState, []diag.Diagnostic) {
	state := cloneCtxAssignState(in)
	diagnostics := make([]diag.Diagnostic, 0)

	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return analyzeCTX03StmtList(s.List, state, info, fset)

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
			for i, name := range vs.Names {
				if name == nil || name.Name == "_" {
					continue
				}
				obj, ok := info.Defs[name].(*types.Var)
				if !ok || !isStrictContextType(obj.Type()) {
					continue
				}
				if i < len(vs.Values) {
					state[obj] = !isNilExpr(vs.Values[i])
				} else {
					state[obj] = false
				}
			}
		}
		return state, diagnostics

	case *ast.AssignStmt:
		for i, lhs := range s.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok {
				continue
			}
			obj, ok := info.ObjectOf(id).(*types.Var)
			if !ok || !isStrictContextType(obj.Type()) {
				continue
			}

			if s.Tok == token.DEFINE {
				if _, exists := state[obj]; !exists {
					if i < len(s.Rhs) {
						state[obj] = !isNilExpr(s.Rhs[i])
					} else {
						state[obj] = false
					}
				}
				continue
			}

			if i < len(s.Rhs) {
				state[obj] = !isNilExpr(s.Rhs[i])
			}
		}
		return state, diagnostics

	case *ast.ExprStmt:
		if call, ok := s.X.(*ast.CallExpr); ok {
			diagnostics = append(diagnostics, ctx03CallDiagnostics(call, state, info, fset)...)
		}
		return state, diagnostics

	case *ast.GoStmt:
		if s.Call != nil {
			diagnostics = append(diagnostics, ctx03CallDiagnostics(s.Call, state, info, fset)...)
		}
		return state, diagnostics

	case *ast.DeferStmt:
		if s.Call != nil {
			diagnostics = append(diagnostics, ctx03CallDiagnostics(s.Call, state, info, fset)...)
		}
		return state, diagnostics

	case *ast.IfStmt:
		if s.Init != nil {
			var initDiags []diag.Diagnostic
			state, initDiags = analyzeCTX03Stmt(s.Init, state, info, fset)
			diagnostics = append(diagnostics, initDiags...)
		}

		thenOut, thenDiags := analyzeCTX03StmtList(s.Body.List, cloneCtxAssignState(state), info, fset)
		diagnostics = append(diagnostics, thenDiags...)

		elseOut := cloneCtxAssignState(state)
		if s.Else != nil {
			var elseDiags []diag.Diagnostic
			elseOut, elseDiags = analyzeCTX03Stmt(s.Else, elseOut, info, fset)
			diagnostics = append(diagnostics, elseDiags...)
		}

		return intersectCtxAssignState(thenOut, elseOut), diagnostics
	}

	return state, diagnostics
}

func ctx03CallDiagnostics(call *ast.CallExpr, state ctxAssignState, info *types.Info, fset *token.FileSet) []diag.Diagnostic {
	if call == nil || info == nil {
		return nil
	}

	sig, ok := info.TypeOf(call.Fun).(*types.Signature)
	if !ok || sig.Params() == nil {
		return nil
	}

	diags := make([]diag.Diagnostic, 0)
	argCount := len(call.Args)
	paramCount := sig.Params().Len()
	limit := argCount
	if paramCount < limit {
		limit = paramCount
	}

	for i := 0; i < limit; i++ {
		param := sig.Params().At(i)
		if !isStrictContextType(param.Type()) {
			continue
		}

		arg := call.Args[i]
		if isNilExpr(arg) {
			pos := fset.Position(arg.Pos())
			diags = append(diags, diag.Diagnostic{
				RuleID:   "CTX-03",
				Severity: diag.SeverityError,
				Message:  "nil must not be passed as context.Context",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use context.Background() or context.TODO()",
			})
			continue
		}

		id, ok := arg.(*ast.Ident)
		if !ok {
			continue
		}
		obj, ok := info.ObjectOf(id).(*types.Var)
		if !ok || !isStrictContextType(obj.Type()) {
			continue
		}
		assigned, tracked := state[obj]
		if !tracked || assigned {
			continue
		}

		pos := fset.Position(id.Pos())
		diags = append(diags, diag.Diagnostic{
			RuleID:   "CTX-03",
			Severity: diag.SeverityError,
			Message:  "context.Context variable may be nil when passed to call",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "ensure context is assigned (e.g., context.Background()) before passing",
		})
	}

	return diags
}

func isNilExpr(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}
