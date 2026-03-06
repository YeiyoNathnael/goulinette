package rules

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type cer03Rule struct{}

const (
	cer03Chapter = 12
)

// NewCER03 returns the CER03 rule implementation.
func NewCER03() Rule {
	return cer03Rule{}
}

// ID returns the rule identifier.
func (cer03Rule) ID() string {
	return ruleCER03
}

// Chapter returns the chapter number for this rule.
func (cer03Rule) Chapter() int {
	return cer03Chapter
}

// Run executes this rule against the provided context.
func (cer03Rule) Run(ctx Context) ([]diag.Finding, error) {
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

				tracked := collectTrackedConcreteErrorVars(fn.Body, pkg.TypesInfo)
				if len(tracked) == 0 {
					continue
				}

				state := make(assignmentState, len(tracked))
				for v := range tracked {
					state[v] = false
				}

				env := cer03Env{tracked: tracked, info: pkg.TypesInfo, fset: pkg.Fset}
				_, diags := analyzeStmtListForCER03(fn.Body.List, state, env)
				diagnostics = append(diagnostics, diags...)
			}
		}
	}

	return diagnostics, nil
}

type assignmentState map[*types.Var]bool

type cer03Env struct {
	tracked map[*types.Var]struct{}
	info    *types.Info
	fset    *token.FileSet
}

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
		vb, ok := b[k]
		out[k] = va && ok && vb
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

func analyzeStmtListForCER03(stmts []ast.Stmt, in assignmentState, env cer03Env) (assignmentState, []diag.Finding) {
	state := cloneState(in)
	diagnostics := make([]diag.Finding, 0)

	for _, stmt := range stmts {
		var ds []diag.Finding
		state, ds = analyzeStmtForCER03(stmt, state, env)
		diagnostics = append(diagnostics, ds...)
	}

	return state, diagnostics
}

func analyzeStmtForCER03(stmt ast.Stmt, in assignmentState, env cer03Env) (assignmentState, []diag.Finding) {
	state := cloneState(in)
	diagnostics := make([]diag.Finding, 0)

	switch stmtNode := stmt.(type) {
	case *ast.BlockStmt:
		return analyzeStmtListForCER03(stmtNode.List, state, env)

	case *ast.DeclStmt:
		return analyzeCER03DeclStmt(stmtNode, state, env.tracked, env.info), diagnostics

	case *ast.AssignStmt:
		analyzeCER03AssignStmt(stmtNode, state, env.tracked, env.info)
		return state, diagnostics

	case *ast.ReturnStmt:
		return state, cer03ReturnDiagnostics(stmtNode, state, env)

	case *ast.IfStmt:
		return analyzeCER03IfStmt(stmtNode, state, env)

	case *ast.ForStmt:
		return analyzeCER03ForStmt(stmtNode, state, env)

	case *ast.RangeStmt:
		return analyzeCER03RangeStmt(stmtNode, state, env)

	case *ast.SwitchStmt:
		return analyzeCER03SwitchStmt(stmtNode, state, env)

	default:
		return state, diagnostics
	}
}

func analyzeCER03DeclStmt(s *ast.DeclStmt, state assignmentState, tracked map[*types.Var]struct{}, info *types.Info) assignmentState {
	gd, ok := s.Decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
		return state
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
	return state
}

func analyzeCER03AssignStmt(s *ast.AssignStmt, state assignmentState, tracked map[*types.Var]struct{}, info *types.Info) {
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
}

func cer03ReturnDiagnostics(s *ast.ReturnStmt, state assignmentState, env cer03Env) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, result := range s.Results {
		id, ok := result.(*ast.Ident)
		if !ok {
			continue
		}
		obj, ok := env.info.ObjectOf(id).(*types.Var)
		if !ok {
			continue
		}
		if _, ok := env.tracked[obj]; !ok {
			continue
		}
		assigned, ok := state[obj]
		if ok && assigned {
			continue
		}
		pos := env.fset.Position(id.Pos())
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleCER03,
			Severity: diag.SeverityError,
			Message:  "custom error variable may be returned in unassigned state",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "ensure variable is assigned on all paths before return, or declare as error and return nil on success",
		})
	}
	return diagnostics
}

func analyzeCER03IfStmt(s *ast.IfStmt, state assignmentState, env cer03Env) (assignmentState, []diag.Finding) {
	diagnostics := make([]diag.Finding, 0)
	if s.Init != nil {
		var initDiags []diag.Finding
		state, initDiags = analyzeStmtForCER03(s.Init, state, env)
		diagnostics = append(diagnostics, initDiags...)
	}

	thenOut, thenDiags := analyzeStmtListForCER03(s.Body.List, state, env)
	diagnostics = append(diagnostics, thenDiags...)

	elseIn := cloneState(state)
	elseOut := elseIn
	if s.Else != nil {
		var elseDiags []diag.Finding
		elseOut, elseDiags = analyzeStmtForCER03(s.Else, elseIn, env)
		diagnostics = append(diagnostics, elseDiags...)
	}

	return intersectState(thenOut, elseOut), diagnostics
}

func analyzeCER03ForStmt(s *ast.ForStmt, state assignmentState, env cer03Env) (assignmentState, []diag.Finding) {
	diagnostics := make([]diag.Finding, 0)
	if s.Init != nil {
		var initDiags []diag.Finding
		state, initDiags = analyzeStmtForCER03(s.Init, state, env)
		diagnostics = append(diagnostics, initDiags...)
	}
	if s.Body != nil {
		_, bodyDiags := analyzeStmtListForCER03(s.Body.List, cloneState(state), env)
		diagnostics = append(diagnostics, bodyDiags...)
	}
	if s.Post != nil {
		_, postDiags := analyzeStmtForCER03(s.Post, cloneState(state), env)
		diagnostics = append(diagnostics, postDiags...)
	}
	return state, diagnostics
}

func analyzeCER03RangeStmt(s *ast.RangeStmt, state assignmentState, env cer03Env) (assignmentState, []diag.Finding) {
	diagnostics := make([]diag.Finding, 0)
	if s.Body != nil {
		_, bodyDiags := analyzeStmtListForCER03(s.Body.List, cloneState(state), env)
		diagnostics = append(diagnostics, bodyDiags...)
	}
	return state, diagnostics
}

func analyzeCER03SwitchStmt(s *ast.SwitchStmt, state assignmentState, env cer03Env) (assignmentState, []diag.Finding) {
	diagnostics := make([]diag.Finding, 0)
	if s.Init != nil {
		var initDiags []diag.Finding
		state, initDiags = analyzeStmtForCER03(s.Init, state, env)
		diagnostics = append(diagnostics, initDiags...)
	}

	var caseOut assignmentState
	var hasDefault bool
	for _, ccStmt := range s.Body.List {
		cc, ok := ccStmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		if cc.List == nil {
			hasDefault = true
		}
		out, ds := analyzeStmtListForCER03(cc.Body, cloneState(state), env)
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
