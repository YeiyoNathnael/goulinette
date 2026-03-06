package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type imp01Rule struct{}

const (
	imp01Chapter   = 15
	imp01MaxGroups = 3
)

// NewIMP01 returns the IMP01 rule implementation.
func NewIMP01() Rule {
	return imp01Rule{}
}

// ID returns the rule identifier.
func (imp01Rule) ID() string {
	return ruleIMP01
}

// Chapter returns the chapter number for this rule.
func (imp01Rule) Chapter() int {
	return imp01Chapter
}

// Run executes this rule against the provided context.
func (imp01Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	modulePath := readModulePath(ctx.Root)
	diagnostics := make([]diag.Finding, 0)

	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.IMPORT || len(gd.Specs) == 0 {
				continue
			}

			groups := importSpecGroups(pf, gd.Specs)
			diagnostics = append(diagnostics, imp01MaxGroupDiagnostics(pf, gd, groups)...)
			diagnostics = append(diagnostics, imp01GroupOrderDiagnostics(pf, groups, modulePath)...)
		}
	}

	return diagnostics, nil
}

func importSpecGroups(pf parsedFile, specs []ast.Spec) [][]*ast.ImportSpec {
	groups := make([][]*ast.ImportSpec, 0)
	current := make([]*ast.ImportSpec, 0)
	var prevEndLine int

	for i, spec := range specs {
		imp, ok := spec.(*ast.ImportSpec)
		if !ok {
			continue
		}

		line := pf.FSet.Position(imp.Pos()).Line
		if i == 0 {
			current = append(current, imp)
			prevEndLine = pf.FSet.Position(imp.End()).Line
			continue
		}

		if line-prevEndLine > 1 {
			groups = append(groups, current)
			current = []*ast.ImportSpec{imp}
		} else {
			current = append(current, imp)
		}
		prevEndLine = pf.FSet.Position(imp.End()).Line
	}

	if len(current) > 0 {
		groups = append(groups, current)
	}

	return groups
}

func imp01MaxGroupDiagnostics(pf parsedFile, gd *ast.GenDecl, groups [][]*ast.ImportSpec) []diag.Finding {
	if len(groups) <= imp01MaxGroups {
		return nil
	}

	pos := pf.FSet.Position(gd.Pos())
	return []diag.Finding{{
		RuleID:   ruleIMP01,
		Severity: diag.SeverityError,
		Message:  "imports must be organized into at most three groups (std, third-party, internal)",
		Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
		Hint:     "group imports as std, third-party, and module-internal with blank lines",
	}}
}

func imp01GroupOrderDiagnostics(pf parsedFile, groups [][]*ast.ImportSpec, modulePath string) []diag.Finding {
	diagnostics := make([]diag.Finding, 0)
	lastClass := importClassStd

	for gi, group := range groups {
		if len(group) == 0 {
			continue
		}

		groupClass := classifyImportPath(group[0].Path.Value, modulePath)
		if imp01IsMixedGroup(group, modulePath, groupClass) {
			pos := pf.FSet.Position(group[0].Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleIMP01,
				Severity: diag.SeverityError,
				Message:  "import group mixes standard, third-party, or internal packages",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "separate import classes with blank lines",
			})
		}

		if gi > 0 && groupClass < lastClass {
			pos := pf.FSet.Position(group[0].Pos())
			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleIMP01,
				Severity: diag.SeverityError,
				Message:  "import groups are out of order",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "order groups as standard library, third-party, then internal",
			})
		}

		lastClass = groupClass
	}

	return diagnostics
}

func imp01IsMixedGroup(group []*ast.ImportSpec, modulePath string, groupClass importClass) bool {
	for _, imp := range group[1:] {
		if classifyImportPath(imp.Path.Value, modulePath) != groupClass {
			return true
		}
	}
	return false
}
