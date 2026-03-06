package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"goulinette/internal/diag"
)

type cer03Rule struct{}

func NewCER03() Rule {
	return cer03Rule{}
}

func (cer03Rule) ID() string {
	return "CER-03"
}

func (cer03Rule) Chapter() int {
	return 12
}

func (cer03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}

				tracked := collectTrackedConcreteErrorVars(fn.Body, pkg.TypesInfo)
				if len(tracked) == 0 {
					continue
				}

				state := make(assignmentState, len(tracked))
				for v := range tracked {
					state[v] = false
				}

				_, diags := analyzeStmtListForCER03(fn.Body.List, state, tracked, pkg.TypesInfo, pkg.Fset)
				diagnostics = append(diagnostics, diags...)
			}
		}
	}

	return diagnostics, nil
}

type assignmentState map[*types.Var]bool

func cloneState(in assignmentState) assignmentState {
	out := make(assignmentState, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func intersectState(a assignmentState, b assignmentState) assignmentState {
	out := make(assignmentState, len(a))
	for k, va := range a {
		vb := b[k]
		out[k] = va && vb
	}
	return out
}

func collectTrackedConcreteErrorVars(body *ast.BlockStmt, info *types.Info) map[*types.Var]struct{} {
	tracked := make(map[*types.Var]struct{})
	if body == nil || info == nil {
		return tracked
	}

	ast.Inspect(body, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			return true
		}

		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name == nil || name.Name == "_" {
					continue
				}
				obj, ok := info.Defs[name].(*types.Var)
				if !ok {
					continue
				}
				if isConcreteErrorType(obj.Type()) {
					tracked[obj] = struct{}{}
				}
			}
		}

		return true
	})

	return tracked
}

func analyzeStmtListForCER03(stmts []ast.Stmt, in assignmentState, tracked map[*types.Var]struct{}, info *types.Info, fset *token.FileSet) (assignmentState, []diag.Diagnostic) {
	state := cloneState(in)
	diagnostics := make([]diag.Diagnostic, 0)

	for _, stmt := range stmts {
		var ds []diag.Diagnostic
		state, ds = analyzeStmtForCER03(stmt, state, tracked, info, fset)
		diagnostics = append(diagnostics, ds...)
	}

	return state, diagnostics
}

func analyzeStmtForCER03(stmt ast.Stmt, in assignmentState, tracked map[*types.Var]struct{}, info *types.Info, fset *token.FileSet) (assignmentState, []diag.Diagnostic) {
	state := cloneState(in)
	diagnostics := make([]diag.Diagnostic, 0)

	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return analyzeStmtListForCER03(s.List, state, tracked, info, fset)

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
				if name == nil {
					continue
				}
				obj, ok := info.Defs[name].(*types.Var)
				if !ok {
					continue
				}
				if _, ok := tracked[obj]; !ok {
					continue
				}
				state[obj] = i < len(vs.Values)
			}
		}
		return state, diagnostics

	case *ast.AssignStmt:
		for _, lhs := range s.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok {
				continue
			}
			obj, ok := info.ObjectOf(id).(*types.Var)
			if !ok {
				continue
			}
			if _, ok := tracked[obj]; ok {
				state[obj] = true
			}
		}
		return state, diagnostics

	case *ast.ReturnStmt:
		for _, result := range s.Results {
			id, ok := result.(*ast.Ident)
			if !ok {
				continue
			}
			obj, ok := info.ObjectOf(id).(*types.Var)
			if !ok {
				continue
			}
			if _, ok := tracked[obj]; !ok {
				continue
			}
			if assigned := state[obj]; assigned {
				continue
			}
			pos := fset.Position(id.Pos())
			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "CER-03",
				Severity: diag.SeverityError,
				Message:  "custom error variable may be returned in unassigned state",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "ensure variable is assigned on all paths before return, or declare as error and return nil on success",
			})
		}
		return state, diagnostics

	case *ast.IfStmt:
		if s.Init != nil {
			var initDiags []diag.Diagnostic
			state, initDiags = analyzeStmtForCER03(s.Init, state, tracked, info, fset)
			diagnostics = append(diagnostics, initDiags...)
		}

		thenOut, thenDiags := analyzeStmtListForCER03(s.Body.List, state, tracked, info, fset)
		diagnostics = append(diagnostics, thenDiags...)

		elseIn := cloneState(state)
		elseOut := elseIn
		if s.Else != nil {
			var elseDiags []diag.Diagnostic
			elseOut, elseDiags = analyzeStmtForCER03(s.Else, elseIn, tracked, info, fset)
			diagnostics = append(diagnostics, elseDiags...)
		}

		return intersectState(thenOut, elseOut), diagnostics

	case *ast.ForStmt:
		if s.Init != nil {
			var initDiags []diag.Diagnostic
			state, initDiags = analyzeStmtForCER03(s.Init, state, tracked, info, fset)
			diagnostics = append(diagnostics, initDiags...)
		}
		if s.Body != nil {
			_, bodyDiags := analyzeStmtListForCER03(s.Body.List, cloneState(state), tracked, info, fset)
			diagnostics = append(diagnostics, bodyDiags...)
		}
		if s.Post != nil {
			_, postDiags := analyzeStmtForCER03(s.Post, cloneState(state), tracked, info, fset)
			diagnostics = append(diagnostics, postDiags...)
		}
		return state, diagnostics

	case *ast.RangeStmt:
		if s.Body != nil {
			_, bodyDiags := analyzeStmtListForCER03(s.Body.List, cloneState(state), tracked, info, fset)
			diagnostics = append(diagnostics, bodyDiags...)
		}
		return state, diagnostics

	case *ast.SwitchStmt:
		if s.Init != nil {
			var initDiags []diag.Diagnostic
			state, initDiags = analyzeStmtForCER03(s.Init, state, tracked, info, fset)
			diagnostics = append(diagnostics, initDiags...)
		}
		var caseOut assignmentState
		hasDefault := false
		for _, ccStmt := range s.Body.List {
			cc, ok := ccStmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			if cc.List == nil {
				hasDefault = true
			}
			out, ds := analyzeStmtListForCER03(cc.Body, cloneState(state), tracked, info, fset)
			diagnostics = append(diagnostics, ds...)
			if caseOut == nil {
				caseOut = out
			} else {
				caseOut = intersectState(caseOut, out)
			}
		}
		if caseOut == nil {
			return state, diagnostics
		}
		if !hasDefault {
			caseOut = intersectState(caseOut, state)
		}
		return caseOut, diagnostics
	}

	return state, diagnostics
}
