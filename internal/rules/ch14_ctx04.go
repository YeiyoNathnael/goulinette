package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	ctx04RuleID                = "CTX-04"
	ctx04PkgContext            = "context"
	ctx04FnWithCancel          = "WithCancel"
	ctx04FnWithTimeout         = "WithTimeout"
	ctx04FnWithDeadline        = "WithDeadline"
	ctx04DefaultCancelName     = "cancel"
	ctx04MsgCancelMustHandled  = "cancel function from derived context must be handled"
	ctx04MsgCancelOnAllExits   = "derived context cancel function must be called or deferred on all exit paths"
	ctx04HintCaptureCancel     = "capture cancel function and defer/call it on all paths"
	ctx04HintDeferCancelPrefix = "defer "
	ctx04HintDeferCancelSuffix = "() immediately after context.WithCancel/WithTimeout/WithDeadline"
)

type ctx04Rule struct{}

// NewCTX04 returns the CTX04 rule implementation.
func NewCTX04() Rule {
	return ctx04Rule{}
}

// ID returns the rule identifier.
func (ctx04Rule) ID() string {
	return ctx04RuleID
}

// Chapter returns the chapter number for this rule.
func (ctx04Rule) Chapter() int {
	return 14
}

// Run executes this rule against the provided context.
func (ctx04Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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

func analyzeCTX04StmtList(stmts []ast.Stmt, in cancelState, info *types.Info, fset *token.FileSet) (cancelState, []diag.Finding) {
	state := cloneCancelState(in)
	diagnostics := make([]diag.Finding, 0)

	for _, stmt := range stmts {
		var ds []diag.Finding
		state, ds = analyzeCTX04Stmt(stmt, state, info, fset)
		diagnostics = append(diagnostics, ds...)
	}

	return state, diagnostics
}

func analyzeCTX04Stmt(stmt ast.Stmt, in cancelState, info *types.Info, fset *token.FileSet) (cancelState, []diag.Finding) {
	state := cloneCancelState(in)
	diagnostics := make([]diag.Finding, 0)

	switch stmtNode := stmt.(type) {
	case *ast.BlockStmt:
		return analyzeCTX04StmtList(stmtNode.List, state, info, fset)

	case *ast.AssignStmt:
		diagnostics = append(diagnostics, ctx04TrackCreationFromAssign(stmtNode, state, info, fset)...)
		if ctx04MarksCancelHandledByDirectCall(stmtNode.Rhs, state, info) {
			// no-op: state is updated in helper
		}
		return state, diagnostics

	case *ast.DeclStmt:
		return state, append(diagnostics, ctx04DeclStmtDiagnostics(stmtNode, state, info, fset)...)

	case *ast.ExprStmt:
		if call, ok := stmtNode.X.(*ast.CallExpr); ok {
			_ = ctx04MarkCancelHandledByCall(call, state, info, false)
		}
		return state, diagnostics

	case *ast.DeferStmt:
		_ = ctx04MarkCancelHandledByCall(stmtNode.Call, state, info, true)
		return state, diagnostics

	case *ast.GoStmt:
		return state, diagnostics

	case *ast.ReturnStmt:
		return state, append(diagnostics, ctx04ReturnStmtDiagnostics(stmtNode, state, fset)...)

	case *ast.IfStmt:
		merged, ds := analyzeCTX04IfStmt(stmtNode, state, info, fset)
		return merged, append(diagnostics, ds...)
	default:
		return state, diagnostics
	}
}

func ctx04DeclStmtDiagnostics(stmt *ast.DeclStmt, state cancelState, info *types.Info, fset *token.FileSet) []diag.Finding {
	gd, ok := stmt.Decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for i, rhs := range vs.Values {
			call, ok := rhs.(*ast.CallExpr)
			if !ok || !isContextDerivationCall(call, info) || i >= len(vs.Names)-1 {
				continue
			}
			cancelName := vs.Names[i+1]
			if cancelName == nil {
				continue
			}
			if cancelName.Name == "_" {
				pos := fset.Position(cancelName.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ctx04RuleID,
					Severity: diag.SeverityError,
					Message:  ctx04MsgCancelMustHandled,
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     ctx04HintCaptureCancel,
				})
				continue
			}
			if obj, ok := info.Defs[cancelName].(*types.Var); ok {
				state[obj] = false
			}
		}
	}
	return diagnostics
}

func ctx04ReturnStmtDiagnostics(stmt *ast.ReturnStmt, state cancelState, fset *token.FileSet) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for cancelVar, handled := range state {
		if handled {
			continue
		}
		pos := fset.Position(stmt.Return)
		name := cancelVar.Name()
		if name == "" {
			name = ctx04DefaultCancelName
		}
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ctx04RuleID,
			Severity: diag.SeverityError,
			Message:  ctx04MsgCancelOnAllExits,
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     ctx04HintDeferCancelPrefix + name + ctx04HintDeferCancelSuffix,
		})
	}
	return diagnostics
}

func analyzeCTX04IfStmt(stmt *ast.IfStmt, state cancelState, info *types.Info, fset *token.FileSet) (cancelState, []diag.Finding) {
	diagnostics := make([]diag.Finding, 0)
	currentState := cloneCancelState(state)

	if stmt.Init != nil {
		var initDiags []diag.Finding
		currentState, initDiags = analyzeCTX04Stmt(stmt.Init, currentState, info, fset)
		diagnostics = append(diagnostics, initDiags...)
	}

	thenState, thenDiags := analyzeCTX04StmtList(stmt.Body.List, cloneCancelState(currentState), info, fset)
	diagnostics = append(diagnostics, thenDiags...)

	elseState := cloneCancelState(currentState)
	if stmt.Else != nil {
		var elseDiags []diag.Finding
		elseState, elseDiags = analyzeCTX04Stmt(stmt.Else, elseState, info, fset)
		diagnostics = append(diagnostics, elseDiags...)
	}

	return mergeCancelState(thenState, elseState), diagnostics
}

func ctx04TrackCreationFromAssign(as *ast.AssignStmt, state cancelState, info *types.Info, fset *token.FileSet) []diag.Finding {
	if as == nil || info == nil {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
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
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ctx04RuleID,
				Severity: diag.SeverityError,
				Message:  ctx04MsgCancelMustHandled,
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     ctx04HintCaptureCancel,
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

func ctx04UnhandledAtExit(fset *token.FileSet, posToken token.Pos, state cancelState) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for cancelVar, handled := range state {
		if handled {
			continue
		}
		pos := fset.Position(posToken)
		name := cancelVar.Name()
		if name == "" {
			name = ctx04DefaultCancelName
		}
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ctx04RuleID,
			Severity: diag.SeverityError,
			Message:  ctx04MsgCancelOnAllExits,
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     ctx04HintDeferCancelPrefix + name + ctx04HintDeferCancelSuffix,
		})
	}
	return diagnostics
}

func ctx04MarksCancelHandledByDirectCall(exprs []ast.Expr, state cancelState, info *types.Info) bool {
	var updated bool
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

	var updated bool

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
	if obj.Pkg().Path() != ctx04PkgContext {
		return false
	}
	name := obj.Name()
	return name == ctx04FnWithCancel || name == ctx04FnWithTimeout || name == ctx04FnWithDeadline
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
	default:
		// no-op
	}

	return nil
}
