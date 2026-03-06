package rules

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam05Rule struct{}

const nam05Chapter = 2

var nam05AllowedInterfaceNames = map[string]struct{}{
	"Rule": {},
}

// NewNAM05 returns the NAM05 rule implementation.
func NewNAM05() Rule {
	return nam05Rule{}
}

// ID returns the rule identifier.
func (nam05Rule) ID() string {
	return ruleNAM05
}

// Chapter returns the chapter number for this rule.
func (nam05Rule) Chapter() int {
	return nam05Chapter
}

// Run executes this rule against the provided context.
func (nam05Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				_, ok = ts.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				if _, allowed := nam05AllowedInterfaceNames[ts.Name.Name]; allowed {
					continue
				}
				if strings.HasSuffix(ts.Name.Name, "er") {
					continue
				}
				pos := pf.FSet.Position(ts.Name.Pos())
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleNAM05,
					Severity: diag.SeverityWarning,
					Message:  "interface name should generally use the -er suffix",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer names like Reader, Writer, or other -er forms when natural",
				})
			}
		}
	}

	return diagnostics, nil
}
