package rules

import (
	"go/ast"
	"go/token"

	"goulinette/internal/diag"
)

type imp01Rule struct{}

func NewIMP01() Rule {
	return imp01Rule{}
}

func (imp01Rule) ID() string {
	return "IMP-01"
}

func (imp01Rule) Chapter() int {
	return 15
}

func (imp01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	modulePath := readModulePath(ctx.Root)
	diagnostics := make([]diag.Diagnostic, 0)

	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.IMPORT || len(gd.Specs) == 0 {
				continue
			}

			groups := make([][]*ast.ImportSpec, 0)
			current := make([]*ast.ImportSpec, 0)
			var prevEndLine int
			for i, spec := range gd.Specs {
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

			if len(groups) > 3 {
				pos := pf.FSet.Position(gd.Pos())
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "IMP-01",
					Severity: diag.SeverityError,
					Message:  "imports must be organized into at most three groups (std, third-party, internal)",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "group imports as std, third-party, and module-internal with blank lines",
				})
			}

			lastClass := importClassStd
			for gi, group := range groups {
				if len(group) == 0 {
					continue
				}

				groupClass := classifyImportPath(group[0].Path.Value, modulePath)
				mixed := false
				for _, imp := range group[1:] {
					if classifyImportPath(imp.Path.Value, modulePath) != groupClass {
						mixed = true
						break
					}
				}

				if mixed {
					pos := pf.FSet.Position(group[0].Pos())
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "IMP-01",
						Severity: diag.SeverityError,
						Message:  "import group mixes standard, third-party, or internal packages",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "separate import classes with blank lines",
					})
				}

				if gi > 0 && groupClass < lastClass {
					pos := pf.FSet.Position(group[0].Pos())
					diagnostics = append(diagnostics, diag.Diagnostic{
						RuleID:   "IMP-01",
						Severity: diag.SeverityError,
						Message:  "import groups are out of order",
						Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
						Hint:     "order groups as standard library, third-party, then internal",
					})
				}

				lastClass = groupClass
			}
		}
	}

	return diagnostics, nil
}
