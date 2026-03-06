package rules

import (
	"go/ast"
	"go/token"

	"goulinette/internal/diag"
)

type lim03Rule struct{}

func NewLIM03() Rule {
	return lim03Rule{}
}

func (lim03Rule) ID() string {
	return "LIM-03"
}

func (lim03Rule) Chapter() int {
	return 13
}

func (lim03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			diagnostics = append(diagnostics, collectNestingDiagnostics(pf.FSet, fn.Body, 0)...) // reset at function boundary
		}

		ast.Inspect(pf.File, func(n ast.Node) bool {
			lit, ok := n.(*ast.FuncLit)
			if !ok || lit.Body == nil {
				return true
			}
			diagnostics = append(diagnostics, collectNestingDiagnostics(pf.FSet, lit.Body, 0)...) // reset at func literal boundary
			return true
		})
	}

	return diagnostics, nil
}

func collectNestingDiagnostics(fset *token.FileSet, body *ast.BlockStmt, depth int) []diag.Diagnostic {
	out := make([]diag.Diagnostic, 0)
	if body == nil {
		return out
	}

	for _, stmt := range body.List {
		out = append(out, collectNestingFromStmt(fset, stmt, depth)...)
	}

	return out
}

func collectNestingFromStmt(fset *token.FileSet, stmt ast.Stmt, depth int) []diag.Diagnostic {
	diags := make([]diag.Diagnostic, 0)

	emitIfTooDeep := func(pos token.Pos, d int) {
		if d <= 4 {
			return
		}
		p := fset.Position(pos)
		diags = append(diags, diag.Diagnostic{
			RuleID:   "LIM-03",
			Severity: diag.SeverityError,
			Message:  "nesting depth must not exceed 4 levels",
			Pos:      diag.Position{File: p.Filename, Line: p.Line, Col: p.Column},
			Hint:     "apply guard clauses or extract helper functions to flatten control flow",
		})
	}

	switch s := stmt.(type) {
	case *ast.BlockStmt:
		diags = append(diags, collectNestingDiagnostics(fset, s, depth)...)

	case *ast.IfStmt:
		if s.Init != nil {
			diags = append(diags, collectNestingFromStmt(fset, s.Init, depth)...)
		}

		thenDepth := depth + 1
		emitIfTooDeep(s.If, thenDepth)
		diags = append(diags, collectNestingDiagnostics(fset, s.Body, thenDepth)...)

		if s.Else != nil {
			// else-if is treated as semantic nesting (else { if ... })
			diags = append(diags, collectNestingFromStmt(fset, s.Else, depth+1)...)
		}

	case *ast.ForStmt:
		if s.Init != nil {
			diags = append(diags, collectNestingFromStmt(fset, s.Init, depth)...)
		}
		if s.Post != nil {
			diags = append(diags, collectNestingFromStmt(fset, s.Post, depth)...)
		}
		loopDepth := depth + 1
		emitIfTooDeep(s.For, loopDepth)
		diags = append(diags, collectNestingDiagnostics(fset, s.Body, loopDepth)...)

	case *ast.RangeStmt:
		loopDepth := depth + 1
		emitIfTooDeep(s.For, loopDepth)
		diags = append(diags, collectNestingDiagnostics(fset, s.Body, loopDepth)...)

	case *ast.SwitchStmt:
		if s.Init != nil {
			diags = append(diags, collectNestingFromStmt(fset, s.Init, depth)...)
		}
		switchDepth := depth + 1
		emitIfTooDeep(s.Switch, switchDepth)
		for _, ccStmt := range s.Body.List {
			cc, ok := ccStmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			// case clauses do not add an extra nesting level beyond switch/select
			for _, cs := range cc.Body {
				diags = append(diags, collectNestingFromStmt(fset, cs, switchDepth)...)
			}
		}

	case *ast.TypeSwitchStmt:
		if s.Init != nil {
			diags = append(diags, collectNestingFromStmt(fset, s.Init, depth)...)
		}
		switchDepth := depth + 1
		emitIfTooDeep(s.Switch, switchDepth)
		for _, ccStmt := range s.Body.List {
			cc, ok := ccStmt.(*ast.CaseClause)
			if !ok {
				continue
			}
			for _, cs := range cc.Body {
				diags = append(diags, collectNestingFromStmt(fset, cs, switchDepth)...)
			}
		}

	case *ast.SelectStmt:
		selectDepth := depth + 1
		emitIfTooDeep(s.Select, selectDepth)
		for _, ccStmt := range s.Body.List {
			cc, ok := ccStmt.(*ast.CommClause)
			if !ok {
				continue
			}
			for _, cs := range cc.Body {
				diags = append(diags, collectNestingFromStmt(fset, cs, selectDepth)...)
			}
		}
	}

	return diags
}
