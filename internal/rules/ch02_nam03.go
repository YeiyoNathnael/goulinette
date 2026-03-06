//lint:file-ignore SA1019 This rule intentionally uses ast.Object links for local-scope tracking without type-checking.
package rules

import (
	"go/ast"
	"go/token"
	"unicode"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam03Rule struct{}

const (
	nam03MaxNameLen   = 15
	nam03ShortScope   = 5
	nam03LongScopeMin = 20
)

// NewNAM03 returns the NAM03 rule implementation.
func NewNAM03() Rule {
	return nam03Rule{}
}

// ID returns the rule identifier.
func (nam03Rule) ID() string {
	return ruleNAM03
}

// Chapter returns the chapter number for this rule.
func (nam03Rule) Chapter() int {
	return 2
}

type nam03VarInfo struct {
	name     string
	declLine int
	lastLine int
	pos      token.Pos
}

// Run executes this rule against the provided context.
func (nam03Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		diagnostics = append(diagnostics, nam03DiagnosticsForFile(pf)...)
	}

	return diagnostics, nil
}

func nam03DiagnosticsForFile(pf parsedFile) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	for _, decl := range pf.File.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		vars := collectNAM03Vars(fn.Body, pf.FSet)
		for _, info := range vars {
			if info == nil {
				continue
			}
			diagnostics = append(diagnostics, nam03DiagnosticsForVar(pf, *info)...)
		}
	}
	return diagnostics
}

func nam03DiagnosticsForVar(pf parsedFile, varInfo nam03VarInfo) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	scopeLen := varInfo.lastLine - varInfo.declLine + 1
	if scopeLen < 1 {
		scopeLen = 1
	}

	diagnostics = append(diagnostics, nam03SingleLetterDiag(pf, varInfo, scopeLen)...)
	diagnostics = append(diagnostics, nam03LongNameDiag(pf, varInfo, scopeLen)...)

	return diagnostics
}

func nam03SingleLetterDiag(pf parsedFile, info nam03VarInfo, scopeLen int) []diag.Finding {
	if !isSingleLetterVar(info.name) || scopeLen <= nam03LongScopeMin {
		return nil
	}

	pos := pf.FSet.Position(info.pos)
	return []diag.Finding{{
		RuleID:   ruleNAM03,
		Severity: diag.SeverityWarning,
		Message:  "single-letter variable name is too short for its scope",
		Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
		Hint:     "use a descriptive name for variables used across large blocks",
	}}
}

func nam03LongNameDiag(pf parsedFile, info nam03VarInfo, scopeLen int) []diag.Finding {
	if len(info.name) <= nam03MaxNameLen || scopeLen >= nam03ShortScope {
		return nil
	}

	pos := pf.FSet.Position(info.pos)
	return []diag.Finding{{
		RuleID:   ruleNAM03,
		Severity: diag.SeverityWarning,
		Message:  "variable name is too long for a very short scope",
		Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
		Hint:     "use concise names for short-lived local variables",
	}}
}

func collectNAM03Vars(body *ast.BlockStmt, fset *token.FileSet) map[*ast.Object]*nam03VarInfo {
	vars := make(map[*ast.Object]*nam03VarInfo)
	if body == nil || fset == nil {
		return vars
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			collectNAM03FromAssign(vars, s, fset)

		case *ast.RangeStmt:
			collectNAM03FromRange(vars, s, fset)

		case *ast.DeclStmt:
			collectNAM03FromDecl(vars, s, fset)

		case *ast.Ident:
			collectNAM03Usage(vars, s, fset)
		default:
			// no-op
		}
		return true
	})

	return vars
}

func collectNAM03FromAssign(vars map[*ast.Object]*nam03VarInfo, stmt *ast.AssignStmt, fset *token.FileSet) {
	if stmt.Tok != token.DEFINE {
		return
	}
	for _, lhs := range stmt.Lhs {
		id, ok := lhs.(*ast.Ident)
		if !ok || id.Name == "_" || id.Obj == nil {
			continue
		}
		collectNAM03VarDecl(vars, id, fset)
	}
}

func collectNAM03FromRange(vars map[*ast.Object]*nam03VarInfo, stmt *ast.RangeStmt, fset *token.FileSet) {
	if stmt.Tok != token.DEFINE {
		return
	}
	for _, expr := range []ast.Expr{stmt.Key, stmt.Value} {
		id, ok := expr.(*ast.Ident)
		if !ok || id.Name == "_" || id.Obj == nil {
			continue
		}
		collectNAM03VarDecl(vars, id, fset)
	}
}

func collectNAM03FromDecl(vars map[*ast.Object]*nam03VarInfo, stmt *ast.DeclStmt, fset *token.FileSet) {
	gd, ok := stmt.Decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR {
		return
	}
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if name == nil || name.Name == "_" || name.Obj == nil {
				continue
			}
			collectNAM03VarDecl(vars, name, fset)
		}
	}
}

func collectNAM03Usage(vars map[*ast.Object]*nam03VarInfo, ident *ast.Ident, fset *token.FileSet) {
	if ident.Obj == nil {
		return
	}
	info, ok := vars[ident.Obj]
	if !ok {
		return
	}
	line := fset.Position(ident.Pos()).Line
	if line > info.lastLine {
		info.lastLine = line
	}
}

func collectNAM03VarDecl(vars map[*ast.Object]*nam03VarInfo, id *ast.Ident, fset *token.FileSet) {
	line := fset.Position(id.Pos()).Line
	vars[id.Obj] = &nam03VarInfo{name: id.Name, declLine: line, lastLine: line, pos: id.Pos()}
}

func isSingleLetterVar(name string) bool {
	runes := []rune(name)
	if len(runes) != 1 {
		return false
	}
	return unicode.IsLetter(runes[0])
}
