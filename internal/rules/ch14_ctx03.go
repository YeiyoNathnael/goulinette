package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type ctx03Rule struct{}

const (
	ctx03Chapter = 14
)

// NewCTX03 returns the CTX03 rule implementation.
func NewCTX03() Rule {
	return ctx03Rule{}
}

// ID returns the rule identifier.
func (ctx03Rule) ID() string {
	return ruleCTX03
}

// Chapter returns the chapter number for this rule.
func (ctx03Rule) Chapter() int {
	return ctx03Chapter
}

// Run executes this rule against the provided context.
func (ctx03Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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

func analyzeCTX03StmtList(stmts []ast.Stmt, in ctxAssignState, info *types.Info, fset *token.FileSet) (ctxAssignState, []diag.Finding) {
	state := cloneCtxAssignState(in)
	diagnostics := make([]diag.Finding, 0)

	for _, stmt := range stmts {
		var ds []diag.Finding
		state, ds = analyzeCTX03Stmt(stmt, state, info, fset)
		diagnostics = append(diagnostics, ds...)
	}

	return state, diagnostics
}

func analyzeCTX03Stmt(stmt ast.Stmt, in ctxAssignState, info *types.Info, fset *token.FileSet) (ctxAssignState, []diag.Finding) {
	state := cloneCtxAssignState(in)
	diagnostics := make([]diag.Finding, 0)

	switch stmtNode := stmt.(type) {
	case *ast.BlockStmt:
		return analyzeCTX03StmtList(stmtNode.List, state, info, fset)

	case *ast.DeclStmt:
		applyCTX03DeclState(state, stmtNode, info)
		return state, diagnostics

	case *ast.AssignStmt:
		applyCTX03AssignmentState(state, stmtNode, info)
		return state, diagnostics

	case *ast.ExprStmt:
		return state, append(diagnostics, ctx03ExprStmtDiagnostics(stmtNode, state, info, fset)...)

	case *ast.GoStmt:
		return state, append(diagnostics, ctx03GoStmtDiagnostics(stmtNode, state, info, fset)...)

	case *ast.DeferStmt:
		return state, append(diagnostics, ctx03DeferStmtDiagnostics(stmtNode, state, info, fset)...)

	case *ast.IfStmt:
		return analyzeCTX03IfStmt(stmtNode, state, diagnostics, info, fset)
	default:
		return state, diagnostics
	}
}

func applyCTX03DeclState(state ctxAssignState, stmt *ast.DeclStmt, info *types.Info) {
	gd, ok := stmt.Decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
		return
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
				continue
			}
			state[obj] = false
		}
	}
}

func ctx03ExprStmtDiagnostics(stmt *ast.ExprStmt, state ctxAssignState, info *types.Info, fset *token.FileSet) []diag.Finding {
	call, ok := stmt.X.(*ast.CallExpr)
	if !ok {
		return nil
	}
	return ctx03CallDiagnostics(call, state, info, fset)
}

func ctx03GoStmtDiagnostics(stmt *ast.GoStmt, state ctxAssignState, info *types.Info, fset *token.FileSet) []diag.Finding {
	if stmt.Call == nil {
		return nil
	}
	return ctx03CallDiagnostics(stmt.Call, state, info, fset)
}

func ctx03DeferStmtDiagnostics(stmt *ast.DeferStmt, state ctxAssignState, info *types.Info, fset *token.FileSet) []diag.Finding {
	if stmt.Call == nil {
		return nil
	}
	return ctx03CallDiagnostics(stmt.Call, state, info, fset)
}

func applyCTX03AssignmentState(state ctxAssignState, assign *ast.AssignStmt, info *types.Info) {
	for i, lhs := range assign.Lhs {
		obj, ok := ctx03ContextVarForLHS(lhs, info)
		if !ok {
			continue
		}

		if assign.Tok == token.DEFINE {
			applyCTX03DefineState(state, obj, i, assign.Rhs)
			continue
		}

		if i < len(assign.Rhs) {
			state[obj] = !isNilExpr(assign.Rhs[i])
		}
	}
}

func ctx03ContextVarForLHS(lhs ast.Expr, info *types.Info) (*types.Var, bool) {
	id, ok := lhs.(*ast.Ident)
	if !ok {
		return nil, false
	}
	obj, ok := info.ObjectOf(id).(*types.Var)
	if !ok || !isStrictContextType(obj.Type()) {
		return nil, false
	}
	return obj, true
}

func applyCTX03DefineState(state ctxAssignState, obj *types.Var, index int, rhs []ast.Expr) {
	if _, exists := state[obj]; exists {
		return
	}
	if index < len(rhs) {
		state[obj] = !isNilExpr(rhs[index])
		return
	}
	state[obj] = false
}

func analyzeCTX03IfStmt(s *ast.IfStmt, state ctxAssignState, diagnostics []diag.Finding, info *types.Info, fset *token.FileSet) (ctxAssignState, []diag.Finding) {
	if s.Init != nil {
		var initDiags []diag.Finding
		state, initDiags = analyzeCTX03Stmt(s.Init, state, info, fset)
		diagnostics = append(diagnostics, initDiags...)
	}

	thenOut, thenDiags := analyzeCTX03StmtList(s.Body.List, cloneCtxAssignState(state), info, fset)
	diagnostics = append(diagnostics, thenDiags...)

	elseOut := cloneCtxAssignState(state)
	if s.Else != nil {
		var elseDiags []diag.Finding
		elseOut, elseDiags = analyzeCTX03Stmt(s.Else, elseOut, info, fset)
		diagnostics = append(diagnostics, elseDiags...)
	}

	return intersectCtxAssignState(thenOut, elseOut), diagnostics
}

func ctx03CallDiagnostics(call *ast.CallExpr, state ctxAssignState, info *types.Info, fset *token.FileSet) []diag.Finding {
	if call == nil || info == nil {
		return nil
	}

	sig, ok := info.TypeOf(call.Fun).(*types.Signature)
	if !ok || sig == nil || sig.Params() == nil {
		return nil
	}

	argCount := len(call.Args)
	paramCount := sig.Params().Len()
	limit := argCount
	if paramCount < limit {
		limit = paramCount
	}

	diags := make([]diag.Finding, 0)
	for i := 0; i < limit; i++ {
		param := sig.Params().At(i)
		if !isStrictContextType(param.Type()) {
			continue
		}
		diags = append(diags, ctx03DiagnosticForArg(call.Args[i], state, info, fset)...)
	}

	return diags
}

func ctx03DiagnosticForArg(arg ast.Expr, state ctxAssignState, info *types.Info, fset *token.FileSet) []diag.Finding {
	if isNilExpr(arg) {
		pos := fset.Position(arg.Pos())
		return []diag.Finding{{
			RuleID:   ruleCTX03,
			Severity: diag.SeverityError,
			Message:  "nil must not be passed as context.Context",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "use context.Background() or context.TODO()",
		}}
	}

	id, ok := arg.(*ast.Ident)
	if !ok {
		return nil
	}
	obj, ok := info.ObjectOf(id).(*types.Var)
	if !ok || !isStrictContextType(obj.Type()) {
		return nil
	}
	assigned, tracked := state[obj]
	if !tracked || assigned {
		return nil
	}

	pos := fset.Position(id.Pos())
	return []diag.Finding{{
		RuleID:   ruleCTX03,
		Severity: diag.SeverityError,
		Message:  "context.Context variable may be nil when passed to call",
		Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
		Hint:     "ensure context is assigned (e.g., context.Background()) before passing",
	}}
}

func isNilExpr(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}
