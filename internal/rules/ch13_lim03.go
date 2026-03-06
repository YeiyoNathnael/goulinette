package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type lim03Rule struct{}

const (
	lim03Chapter  = 13
	lim03MaxDepth = 4
)

// NewLIM03 returns the LIM03 rule implementation.
func NewLIM03() Rule {
	return lim03Rule{}
}

// ID returns the rule identifier.
func (lim03Rule) ID() string {
	return ruleLIM03
}

// Chapter returns the chapter number for this rule.
func (lim03Rule) Chapter() int {
	return lim03Chapter
}

// Run executes this rule against the provided context.
func (lim03Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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

func collectNestingDiagnostics(fset *token.FileSet, body *ast.BlockStmt, depth int) []diag.Finding {
	out := make([]diag.Finding, 0)
	if body == nil {
		return out
	}

	for _, stmt := range body.List {
		out = append(out, collectNestingFromStmt(fset, stmt, depth)...)
	}

	return out
}

func collectNestingFromStmt(fset *token.FileSet, stmt ast.Stmt, depth int) []diag.Finding {
	diags := make([]diag.Finding, 0)

	emitIfTooDeep := func(pos token.Pos, d int) {
		if d <= lim03MaxDepth {
			return
		}
		p := fset.Position(pos)
		diags = append(diags, diag.Finding{
			RuleID:   ruleLIM03,
			Severity: diag.SeverityError,
			Message:  "nesting depth must not exceed 4 levels",
			Pos:      diag.Position{File: p.Filename, Line: p.Line, Col: p.Column},
			Hint:     "apply guard clauses or extract helper functions to flatten control flow",
		})
	}

	switch stmtNode := stmt.(type) {
	case *ast.BlockStmt:
		diags = append(diags, collectNestingDiagnostics(fset, stmtNode, depth)...)

	case *ast.IfStmt:
		diags = append(diags, collectNestingFromIfStmt(fset, stmtNode, depth, emitIfTooDeep)...)

	case *ast.ForStmt:
		diags = append(diags, collectNestingFromForStmt(fset, stmtNode, depth, emitIfTooDeep)...)

	case *ast.RangeStmt:
		diags = append(diags, collectNestingFromRangeStmt(fset, stmtNode, depth, emitIfTooDeep)...)

	case *ast.SwitchStmt:
		diags = append(diags, collectNestingFromSwitchStmt(fset, stmtNode, depth, emitIfTooDeep)...)

	case *ast.TypeSwitchStmt:
		diags = append(diags, collectNestingFromTypeSwitchStmt(fset, stmtNode, depth, emitIfTooDeep)...)

	case *ast.SelectStmt:
		diags = append(diags, collectNestingFromSelectStmt(fset, stmtNode, depth, emitIfTooDeep)...)
	default:
		// no-op
	}

	return diags
}

func collectNestingFromIfStmt(fset *token.FileSet, s *ast.IfStmt, depth int, emitIfTooDeep func(token.Pos, int)) []diag.Finding {
	diags := make([]diag.Finding, 0)
	if s.Init != nil {
		diags = append(diags, collectNestingFromStmt(fset, s.Init, depth)...)
	}

	thenDepth := depth + 1
	emitIfTooDeep(s.If, thenDepth)
	diags = append(diags, collectNestingDiagnostics(fset, s.Body, thenDepth)...)

	if s.Else != nil {
		diags = append(diags, collectNestingFromStmt(fset, s.Else, depth+1)...)
	}

	return diags
}

func collectNestingFromForStmt(fset *token.FileSet, s *ast.ForStmt, depth int, emitIfTooDeep func(token.Pos, int)) []diag.Finding {
	diags := make([]diag.Finding, 0)
	if s.Init != nil {
		diags = append(diags, collectNestingFromStmt(fset, s.Init, depth)...)
	}
	if s.Post != nil {
		diags = append(diags, collectNestingFromStmt(fset, s.Post, depth)...)
	}
	loopDepth := depth + 1
	emitIfTooDeep(s.For, loopDepth)
	diags = append(diags, collectNestingDiagnostics(fset, s.Body, loopDepth)...)
	return diags
}

func collectNestingFromRangeStmt(fset *token.FileSet, s *ast.RangeStmt, depth int, emitIfTooDeep func(token.Pos, int)) []diag.Finding {
	loopDepth := depth + 1
	emitIfTooDeep(s.For, loopDepth)
	return collectNestingDiagnostics(fset, s.Body, loopDepth)
}

func collectNestingFromSwitchStmt(fset *token.FileSet, s *ast.SwitchStmt, depth int, emitIfTooDeep func(token.Pos, int)) []diag.Finding {
	diags := make([]diag.Finding, 0)
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
	return diags
}

func collectNestingFromTypeSwitchStmt(fset *token.FileSet, s *ast.TypeSwitchStmt, depth int, emitIfTooDeep func(token.Pos, int)) []diag.Finding {
	diags := make([]diag.Finding, 0)
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
	return diags
}

func collectNestingFromSelectStmt(fset *token.FileSet, s *ast.SelectStmt, depth int, emitIfTooDeep func(token.Pos, int)) []diag.Finding {
	diags := make([]diag.Finding, 0)
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
	return diags
}
