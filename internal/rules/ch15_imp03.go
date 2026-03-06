package rules

import (
	"go/ast"

	"goulinette/internal/diag"
)

type imp03Rule struct{}

func NewIMP03() Rule {
	return imp03Rule{}
}

func (imp03Rule) ID() string {
	return "IMP-03"
}

func (imp03Rule) Chapter() int {
	return 15
}

func (imp03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		defaultNameCount := make(map[string]int)
		for _, imp := range pf.File.Imports {
			defaultNameCount[defaultImportName(imp.Path.Value)]++
		}

		for _, imp := range pf.File.Imports {
			pos := pf.FSet.Position(imp.Path.Pos())
			defaultName := defaultImportName(imp.Path.Value)

			if imp.Name == nil {
				continue
			}

			alias := imp.Name.Name
			switch alias {
			case "_":
				if importSpecHasComment(imp) {
					continue
				}
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "IMP-03",
					Severity: diag.SeverityWarning,
					Message:  "blank imports should include a comment explaining side effects",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "add a comment documenting why this side-effect import is needed",
				})
				continue

			case ".":
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "IMP-03",
					Severity: diag.SeverityWarning,
					Message:  "dot imports should be avoided",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer explicit package-qualified usage",
				})
				continue
			}

			if len(alias) == 1 {
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "IMP-03",
					Severity: diag.SeverityWarning,
					Message:  "import alias should be descriptive",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace one-letter alias with a descriptive name",
				})
			}

			if alias == defaultName {
				continue
			}
			if defaultNameCount[defaultName] > 1 {
				continue
			}

			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "IMP-03",
				Severity: diag.SeverityWarning,
				Message:  "import alias appears unnecessary",
				Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
				Hint:     "remove alias unless resolving a naming conflict",
			})
		}
	}

	return diagnostics, nil
}

func importSpecHasComment(imp *ast.ImportSpec) bool {
	if imp == nil {
		return false
	}
	if imp.Doc != nil && len(imp.Doc.Text()) > 0 {
		return true
	}
	if imp.Comment != nil && len(imp.Comment.Text()) > 0 {
		return true
	}
	return false
}
