package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type mag02Rule struct{}

func NewMAG02() Rule {
	return mag02Rule{}
}

func (mag02Rule) ID() string {
	return "MAG-02"
}

func (mag02Rule) Chapter() int {
	return 18
}

func (mag02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	units, err := collectMagAstUnits(ctx)
	if err != nil {
		return nil, err
	}

	occByValue := make(map[string][]magLiteralOccurrence)
	for _, unit := range units {
		if unit.File == nil || unit.FSet == nil {
			continue
		}
		fileIsTest := isTestFile(unit.Filename)

		astInspectWithStack(unit.File, func(n ast.Node, stack []ast.Node) {
			lit, ok := n.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return
			}

			parent := astDirectParent(stack)
			if isInImportSpec(stack) || isStructTagLiteral(lit, parent) || isDirectErrorMessageLiteral(lit, parent) {
				return
			}
			if isInConstDecl(stack) {
				return
			}
			if shortStringLiteral(lit, 4) {
				return
			}

			pos := unit.FSet.Position(lit.Pos())
			occByValue[lit.Value] = append(occByValue[lit.Value], magLiteralOccurrence{
				value:  lit.Value,
				pos:    pos,
				isTest: fileIsTest,
			})
		})
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for literal, occs := range occByValue {
		nonTestCount := 0
		testCount := 0
		for _, occ := range occs {
			if occ.isTest {
				testCount++
			} else {
				nonTestCount++
			}
		}

		for _, occ := range occs {
			if occ.isTest {
				if testCount < 3 {
					continue
				}
			} else {
				if nonTestCount < 2 {
					continue
				}
			}

			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "MAG-02",
				Severity: diag.SeverityError,
				Message:  "string literal " + literal + " is repeated; extract to a named constant",
				Pos:      diag.Position{File: occ.pos.Filename, Line: occ.pos.Line, Col: occ.pos.Column},
				Hint:     "for repeated identifiers/keys/tokens, define a descriptive constant",
			})
		}
	}

	return diagnostics, nil
}
