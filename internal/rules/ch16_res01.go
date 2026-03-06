package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type res01Rule struct{}

const (
	res01Chapter            = 16
	res01BodyFieldName      = "Body"
	res01CloseMethodName    = "Close"
	res01HintCheckErrFirst  = "check err first, then defer close"
	res01HintDeferAfterOpen = "place defer <resource>.Close() immediately after successful acquisition"
	res01HintAddDefer       = "add defer <resource>.Close() immediately after acquisition"
)

// NewRES01 returns the RES01 rule implementation.
func NewRES01() Rule {
	return res01Rule{}
}

// ID returns the rule identifier.
func (res01Rule) ID() string {
	return ruleRES01
}

// Chapter returns the chapter number for this rule.
func (res01Rule) Chapter() int {
	return res01Chapter
}

type resourceKey struct {
	obj   *types.Var
	field string
}

type trackedResource struct {
	key          resourceKey
	errVar       *types.Var
	createdPos   token.Pos
	createdIndex int
	deferred     bool
	deferredPos  token.Pos
	reported     bool
}

// Run executes this rule against the provided context.
func (res01Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}
				diagnostics = append(diagnostics, analyzeRES01Block(fn.Body, pkg.TypesInfo, pkg.Fset)...)
			}

			ast.Inspect(file, func(n ast.Node) bool {
				lit, ok := n.(*ast.FuncLit)
				if !ok || lit.Body == nil {
					return true
				}
				diagnostics = append(diagnostics, analyzeRES01Block(lit.Body, pkg.TypesInfo, pkg.Fset)...)
				return true
			})
		}
	}

	return diagnostics, nil
}

func analyzeRES01Block(body *ast.BlockStmt, info *types.Info, fset *token.FileSet) []diag.Finding {
	if body == nil {
		return nil
	}

	tracked := make(map[resourceKey]*trackedResource)
	diagnostics := make([]diag.Finding, 0)

	for idx, stmt := range body.List {
		res01TrackAcquiredResources(tracked, stmt, idx, info)
		markDeferredResources(stmt, tracked, info)
		diagnostics = append(diagnostics, res01ErrCheckOrderDiagnostics(stmt, tracked, info, fset)...)
		diagnostics = append(diagnostics, res01EarlyReturnDiagnostics(stmt, tracked, info, fset)...)
	}

	for _, rs := range tracked {
		if rs.deferred || rs.reported {
			continue
		}
		pos := fset.Position(rs.createdPos)
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleRES01,
			Severity: diag.SeverityError,
			Message:  "closeable resource must be closed with defer",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     res01HintAddDefer,
		})
	}

	return diagnostics
}

func res01TrackAcquiredResources(tracked map[resourceKey]*trackedResource, stmt ast.Stmt, index int, info *types.Info) {
	for _, rs := range collectAcquiredResources(stmt, index, info) {
		tracked[rs.key] = rs
	}
}

func res01ErrCheckOrderDiagnostics(stmt ast.Stmt, tracked map[resourceKey]*trackedResource, info *types.Info, fset *token.FileSet) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, rs := range tracked {
		if rs.errVar == nil || rs.reported {
			continue
		}
		if !isErrCheckReturnStmt(stmt, rs.errVar, info) {
			continue
		}
		if rs.deferredPos == token.NoPos || rs.deferredPos >= stmt.Pos() {
			continue
		}
		pos := fset.Position(rs.deferredPos)
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleRES01,
			Severity: diag.SeverityError,
			Message:  "defer close must appear after acquisition error check",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     res01HintCheckErrFirst,
		})
		rs.reported = true
	}
	return diagnostics
}

func res01EarlyReturnDiagnostics(stmt ast.Stmt, tracked map[resourceKey]*trackedResource, info *types.Info, fset *token.FileSet) []diag.Finding {
	if !stmtHasReturnExcludingFuncLit(stmt) {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, rs := range tracked {
		if rs.deferred || rs.reported {
			continue
		}
		if isErrCheckReturnStmt(stmt, rs.errVar, info) {
			continue
		}
		pos := fset.Position(rs.createdPos)
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleRES01,
			Severity: diag.SeverityError,
			Message:  "closeable resource must be closed with defer before early return",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     res01HintDeferAfterOpen,
		})
		rs.reported = true
	}
	return diagnostics
}

func collectAcquiredResources(stmt ast.Stmt, index int, info *types.Info) []*trackedResource {
	out := make([]*trackedResource, 0)

	collectFromAssignment := func(lhs []ast.Expr, rhs []ast.Expr, tok token.Token) {
		for _, rhsExpr := range rhs {
			call, ok := rhsExpr.(*ast.CallExpr)
			if !ok {
				continue
			}

			sig, ok := info.TypeOf(call.Fun).(*types.Signature)
			if !ok || sig.Results() == nil {
				continue
			}

			errVar := assignedErrorVar(lhs, sig, tok, info)
			for i := 0; i < sig.Results().Len(); i++ {
				if i >= len(lhs) {
					continue
				}
				id, obj := assignedVar(lhs[i], tok, info)
				if id == nil || obj == nil || id.Name == "_" {
					continue
				}

				resType := sig.Results().At(i).Type()
				if isCloserType(resType) {
					out = append(out, &trackedResource{
						key:          resourceKey{obj: obj, field: ""},
						errVar:       errVar,
						createdPos:   id.Pos(),
						createdIndex: index,
					})
				}
				if hasCloserBodyField(resType) {
					out = append(out, &trackedResource{
						key:          resourceKey{obj: obj, field: res01BodyFieldName},
						errVar:       errVar,
						createdPos:   id.Pos(),
						createdIndex: index,
					})
				}
			}
		}
	}

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		collectFromAssignment(s.Lhs, s.Rhs, s.Tok)

	case *ast.DeclStmt:
		gd, ok := s.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			return out
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			lhs := make([]ast.Expr, 0, len(vs.Names))
			for _, n := range vs.Names {
				lhs = append(lhs, n)
			}
			collectFromAssignment(lhs, vs.Values, token.DEFINE)
		}
	default:
		// no-op
	}

	return out
}

func assignedErrorVar(lhs []ast.Expr, sig *types.Signature, tok token.Token, info *types.Info) *types.Var {
	if sig == nil || sig.Results() == nil {
		return nil
	}
	for i := 0; i < sig.Results().Len(); i++ {
		if i >= len(lhs) {
			continue
		}
		if !isErrorInterfaceType(sig.Results().At(i).Type()) {
			continue
		}
		_, obj := assignedVar(lhs[i], tok, info)
		return obj
	}
	return nil
}

func assignedVar(expr ast.Expr, tok token.Token, info *types.Info) (*ast.Ident, *types.Var) {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return nil, nil
	}
	if tok == token.DEFINE {
		obj, _ := info.Defs[id].(*types.Var)
		return id, obj
	}
	obj, _ := info.ObjectOf(id).(*types.Var)
	return id, obj
}

func markDeferredResources(stmt ast.Stmt, tracked map[resourceKey]*trackedResource, info *types.Info) {
	ds, ok := stmt.(*ast.DeferStmt)
	if !ok || ds.Call == nil {
		return
	}

	call := ds.Call
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if ok && sel.Sel != nil && sel.Sel.Name == res01CloseMethodName {
		switch x := sel.X.(type) {
		case *ast.Ident:
			obj, _ := info.ObjectOf(x).(*types.Var)
			if obj == nil {
				break
			}
			if rs, ok := tracked[resourceKey{obj: obj, field: ""}]; ok {
				rs.deferred = true
				rs.deferredPos = ds.Defer
			}

		case *ast.SelectorExpr:
			if x.Sel == nil || x.Sel.Name != res01BodyFieldName {
				break
			}
			id, ok := x.X.(*ast.Ident)
			if !ok {
				break
			}
			obj, _ := info.ObjectOf(id).(*types.Var)
			if obj == nil {
				break
			}
			if rs, ok := tracked[resourceKey{obj: obj, field: res01BodyFieldName}]; ok {
				rs.deferred = true
				rs.deferredPos = ds.Defer
			}
		default:
			// no-op
		}
	}
}

func isErrCheckReturnStmt(stmt ast.Stmt, errVar *types.Var, info *types.Info) bool {
	if errVar == nil {
		return false
	}
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok || ifStmt.Cond == nil || ifStmt.Body == nil {
		return false
	}

	if !conditionChecksVarNotNil(ifStmt.Cond, errVar, info) {
		return false
	}
	return stmtHasReturnExcludingFuncLit(ifStmt.Body)
}

func conditionChecksVarNotNil(expr ast.Expr, target *types.Var, info *types.Info) bool {
	b, ok := expr.(*ast.BinaryExpr)
	if !ok || b.Op != token.NEQ {
		return false
	}

	leftIsTarget := exprIsVar(b.X, target, info)
	rightIsTarget := exprIsVar(b.Y, target, info)
	leftNil := exprIsNilIdent(b.X)
	rightNil := exprIsNilIdent(b.Y)

	return (leftIsTarget && rightNil) || (rightIsTarget && leftNil)
}

func exprIsVar(expr ast.Expr, target *types.Var, info *types.Info) bool {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj, _ := info.ObjectOf(id).(*types.Var)
	return obj == target
}

func exprIsNilIdent(expr ast.Expr) bool {
	id, ok := expr.(*ast.Ident)
	return ok && id.Name == "nil"
}

func stmtHasReturnExcludingFuncLit(stmt ast.Stmt) bool {
	var found bool
	ast.Inspect(stmt, func(n ast.Node) bool {
		if found {
			return false
		}
		if n == nil {
			return true
		}
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		if _, ok := n.(*ast.ReturnStmt); ok {
			found = true
			return false
		}
		return true
	})
	return found
}
