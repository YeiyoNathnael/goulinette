package rules

import (
	"go/ast"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type str02Rule struct{}

const str02Chapter = 8

// NewSTR02 returns the STR02 rule implementation.
func NewSTR02() Rule {
	return str02Rule{}
}

// ID returns the rule identifier.
func (str02Rule) ID() string {
	return ruleSTR02
}

// Chapter returns the chapter number for this rule.
func (str02Rule) Chapter() int {
	return str02Chapter
}

// Run executes this rule against the provided context.
func (str02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			recv := fn.Recv.List[0]
			if len(recv.Names) == 0 || recv.Names[0] == nil {
				continue
			}
			name := strings.ToLower(recv.Names[0].Name)
			if name != "this" && name != "self" {
				continue
			}

			pos := pf.FSet.Position(recv.Names[0].Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleSTR02,
				Severity: diag.SeverityError,
				Message:  "receiver names this and self are forbidden",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "use a short receiver abbreviation based on type name",
			})
		}
	}

	return diagnostics, nil
}
