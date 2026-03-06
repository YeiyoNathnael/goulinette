package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type var04Rule struct{}

func NewVAR04() Rule {
	return var04Rule{}
}

func (var04Rule) ID() string {
	return "VAR-04"
}

func (var04Rule) Chapter() int {
	return 3
}

func (var04Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	type packageVar struct {
		ident      *ast.Ident
		hasInit    bool
		filePath   string
		declaration token.Pos
	}

	varsByFile := map[string][]packageVar{}
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				hasInit := len(vs.Values) > 0
				for _, name := range vs.Names {
					if name == nil || name.Name == "_" {
						continue
					}
					varsByFile[pf.Path] = append(varsByFile[pf.Path], packageVar{
						ident:      name,
						hasInit:    hasInit,
						filePath:   pf.Path,
						declaration: name.Pos(),
					})
				}
			}
		}
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		vars := varsByFile[pf.Path]
		if len(vars) == 0 {
			continue
		}

		for _, variable := range vars {
			if !variable.hasInit {
				pos := pf.FSet.Position(variable.ident.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "VAR-04",
					Severity: diag.SeverityError,
					Message:  "mutable package-level variables are forbidden",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer constants or move state into explicit struct dependencies",
				})
				continue
			}

			if hasSameFileWriteToObj(pf.File, variable.ident.Obj) {
				pos := pf.FSet.Position(variable.ident.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "VAR-04",
					Severity: diag.SeverityError,
					Message:  "package-level variable is written after initialization",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "avoid mutable global state; encapsulate mutable state in structs",
				})
			}
		}
	}

	return diagnostics, nil
}

func hasSameFileWriteToObj(file *ast.File, obj *ast.Object) bool {
	if obj == nil {
		return false
	}

	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if stmt.Tok != token.ASSIGN && stmt.Tok != token.ADD_ASSIGN && stmt.Tok != token.SUB_ASSIGN && stmt.Tok != token.MUL_ASSIGN && stmt.Tok != token.QUO_ASSIGN && stmt.Tok != token.REM_ASSIGN && stmt.Tok != token.AND_ASSIGN && stmt.Tok != token.OR_ASSIGN && stmt.Tok != token.XOR_ASSIGN && stmt.Tok != token.SHL_ASSIGN && stmt.Tok != token.SHR_ASSIGN && stmt.Tok != token.AND_NOT_ASSIGN {
				return true
			}
			for _, lhs := range stmt.Lhs {
				ident, ok := lhs.(*ast.Ident)
				if ok && ident.Obj == obj {
					found = true
					return false
				}
			}
		case *ast.IncDecStmt:
			ident, ok := stmt.X.(*ast.Ident)
			if ok && ident.Obj == obj {
				found = true
				return false
			}
		}

		return true
	})

	return found
}
