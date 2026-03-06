//lint:file-ignore SA1019 This rule intentionally relies on ast.Object identity for same-file package var mutation tracking.
package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type var04Rule struct{}

type packageVar struct {
	ident       *ast.Ident
	hasInit     bool
	filePath    string
	declaration token.Pos
}

const (
	var04Chapter = 3
)

// NewVAR04 returns the VAR04 rule implementation.
func NewVAR04() Rule {
	return var04Rule{}
}

// ID returns the rule identifier.
func (var04Rule) ID() string {
	return ruleVAR04
}

// Chapter returns the chapter number for this rule.
func (var04Rule) Chapter() int {
	return var04Chapter
}

// Run executes this rule against the provided context.
func (var04Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	varsByFile := map[string][]packageVar{}
	for _, pf := range parsed {
		varsByFile[pf.Path] = collectPackageVarsForFile(pf)
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		vars := varsByFile[pf.Path]
		if len(vars) == 0 {
			continue
		}

		for _, variable := range vars {
			if !variable.hasInit {
				pos := pf.FSet.Position(variable.ident.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleVAR04,
					Severity: diag.SeverityError,
					Message:  "mutable package-level variables are forbidden",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer constants or move state into explicit struct dependencies",
				})
				continue
			}

			if hasSameFileWriteToObj(pf.File, variable.ident.Obj) {
				pos := pf.FSet.Position(variable.ident.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleVAR04,
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

func collectPackageVarsForFile(pf parsedFile) []packageVar {
	vars := make([]packageVar, 0)
	for _, decl := range pf.File.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vars = append(vars, packageVarsFromSpec(pf, spec)...)
		}
	}
	return vars
}

func packageVarsFromSpec(pf parsedFile, spec ast.Spec) []packageVar {
	vs, ok := spec.(*ast.ValueSpec)
	if !ok {
		return nil
	}

	hasInit := len(vs.Values) > 0
	vars := make([]packageVar, 0)
	for _, name := range vs.Names {
		if name == nil || name.Name == "_" {
			continue
		}
		vars = append(vars, packageVar{
			ident:       name,
			hasInit:     hasInit,
			filePath:    pf.Path,
			declaration: name.Pos(),
		})
	}

	return vars
}

func hasSameFileWriteToObj(file *ast.File, obj *ast.Object) bool {
	if obj == nil {
		return false
	}

	var found bool
	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}

		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if !isWriteAssignmentToken(stmt.Tok) {
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
		default:
			// no-op
		}

		return true
	})

	return found
}

func isWriteAssignmentToken(tok token.Token) bool {
	switch tok {
	case token.ASSIGN,
		token.ADD_ASSIGN,
		token.SUB_ASSIGN,
		token.MUL_ASSIGN,
		token.QUO_ASSIGN,
		token.REM_ASSIGN,
		token.AND_ASSIGN,
		token.OR_ASSIGN,
		token.XOR_ASSIGN,
		token.SHL_ASSIGN,
		token.SHR_ASSIGN,
		token.AND_NOT_ASSIGN:
		return true
	default:
		return false
	}
}
