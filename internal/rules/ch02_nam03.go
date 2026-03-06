package rules

import (
	"go/ast"
	"go/token"
	"unicode"

	"goulinette/internal/diag"
)

type nam03Rule struct{}

func NewNAM03() Rule {
	return nam03Rule{}
}

func (nam03Rule) ID() string {
	return "NAM-03"
}

func (nam03Rule) Chapter() int {
	return 2
}

type nam03VarInfo struct {
	name     string
	declLine int
	lastLine int
	pos      token.Pos
}

func (nam03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
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

			vars := collectNAM03Vars(fn.Body, pf.FSet)
			for _, info := range vars {
				scopeLen := info.lastLine - info.declLine + 1
				if scopeLen < 1 {
					scopeLen = 1
				}

				if isSingleLetterVar(info.name) && scopeLen > 20 {
					pos := pf.FSet.Position(info.pos)
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "NAM-03",
						Severity: diag.SeverityWarning,
						Message:  "single-letter variable name is too short for its scope",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "use a descriptive name for variables used across large blocks",
					})
				}

				if len(info.name) > 15 && scopeLen < 5 {
					pos := pf.FSet.Position(info.pos)
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "NAM-03",
						Severity: diag.SeverityWarning,
						Message:  "variable name is too long for a very short scope",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "use concise names for short-lived local variables",
					})
				}
			}
		}
	}

	return diagnostics, nil
}

func collectNAM03Vars(body *ast.BlockStmt, fset *token.FileSet) map[*ast.Object]*nam03VarInfo {
	vars := make(map[*ast.Object]*nam03VarInfo)
	if body == nil || fset == nil {
		return vars
	}

	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.AssignStmt:
			if s.Tok != token.DEFINE {
				break
			}
			for _, lhs := range s.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok || id.Name == "_" || id.Obj == nil {
					continue
				}
				line := fset.Position(id.Pos()).Line
				vars[id.Obj] = &nam03VarInfo{name: id.Name, declLine: line, lastLine: line, pos: id.Pos()}
			}

		case *ast.RangeStmt:
			if s.Tok != token.DEFINE {
				break
			}
			for _, expr := range []ast.Expr{s.Key, s.Value} {
				id, ok := expr.(*ast.Ident)
				if !ok || id.Name == "_" || id.Obj == nil {
					continue
				}
				line := fset.Position(id.Pos()).Line
				vars[id.Obj] = &nam03VarInfo{name: id.Name, declLine: line, lastLine: line, pos: id.Pos()}
			}

		case *ast.DeclStmt:
			gd, ok := s.Decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				break
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
					line := fset.Position(name.Pos()).Line
					vars[name.Obj] = &nam03VarInfo{name: name.Name, declLine: line, lastLine: line, pos: name.Pos()}
				}
			}

		case *ast.Ident:
			if s.Obj == nil {
				break
			}
			if info, ok := vars[s.Obj]; ok {
				line := fset.Position(s.Pos()).Line
				if line > info.lastLine {
					info.lastLine = line
				}
			}
		}
		return true
	})

	return vars
}

func isSingleLetterVar(name string) bool {
	runes := []rune(name)
	if len(runes) != 1 {
		return false
	}
	return unicode.IsLetter(runes[0])
}
