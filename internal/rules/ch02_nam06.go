package rules

import (
	"strings"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type nam06Rule struct{}

const nam06Chapter = 2
const nam06TestFileSuffix = "_test.go"

// NewNAM06 returns the NAM06 rule implementation.
func NewNAM06() Rule {
	return nam06Rule{}
}

// ID returns the rule identifier.
func (nam06Rule) ID() string {
	return ruleNAM06
}

// Chapter returns the chapter number for this rule.
func (nam06Rule) Chapter() int {
	return nam06Chapter
}

// Run executes this rule against the provided context.
func (nam06Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	forbidden := map[string]struct{}{
		"util":    {},
		"helpers": {},
		"common":  {},
		"misc":    {},
		"shared":  {},
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		if strings.HasSuffix(pf.Path, nam06TestFileSuffix) {
			continue
		}
		pkgName := strings.ToLower(strings.TrimSpace(pf.File.Name.Name))
		if _, bad := forbidden[pkgName]; !bad {
			continue
		}

		pos := pf.FSet.Position(pf.File.Name.Pos())
		diagnostics = append(diagnostics, diag.Finding{
			RuleID:   ruleNAM06,
			Severity: diag.SeverityError,
			Message:  "package name is too generic and forbidden",
			Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
			Hint:     "rename package to a descriptive noun",
		})
	}

	return diagnostics, nil
}
