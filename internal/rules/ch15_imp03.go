package rules

import (
	"go/ast"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type imp03Rule struct{}

const (
	imp03Chapter = 15
)

// NewIMP03 returns the IMP03 rule implementation.
func NewIMP03() Rule {
	return imp03Rule{}
}

// ID returns the rule identifier.
func (imp03Rule) ID() string {
	return ruleIMP03
}

// Chapter returns the chapter number for this rule.
func (imp03Rule) Chapter() int {
	return imp03Chapter
}

// Run executes this rule against the provided context.
func (imp03Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pf := range parsed {
		defaultNameCount := make(map[string]int)
		for _, imp := range pf.File.Imports {
			key := defaultImportName(imp.Path.Value)
			count, ok := defaultNameCount[key]
			if !ok {
				defaultNameCount[key] = 1
				continue
			}
			defaultNameCount[key] = count + 1
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
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleIMP03,
					Severity: diag.SeverityWarning,
					Message:  "blank imports should include a comment explaining side effects",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "add a comment documenting why this side-effect import is needed",
				})
				continue

			case ".":
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleIMP03,
					Severity: diag.SeverityWarning,
					Message:  "dot imports should be avoided",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer explicit package-qualified usage",
				})
				continue
			}

			if len(alias) == 1 {
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleIMP03,
					Severity: diag.SeverityWarning,
					Message:  "import alias should be descriptive",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace one-letter alias with a descriptive name",
				})
			}

			if alias == defaultName {
				continue
			}
			count, ok := defaultNameCount[defaultName]
			if ok && count > 1 {
				continue
			}

			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleIMP03,
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
