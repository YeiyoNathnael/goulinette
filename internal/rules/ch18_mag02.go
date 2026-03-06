package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type mag02Rule struct{}

const (
	mag02Chapter                = 18
	mag02ShortStringThreshold   = 4
	mag02TestRepeatThreshold    = 3
	mag02NonTestRepeatThreshold = 2
	mag02MessageRepeatedLiteral = " is repeated; extract to a named constant"
	mag02HintRepeatedLiteral    = "for repeated identifiers/keys/tokens, define a descriptive constant"
)

// NewMAG02 returns the MAG02 rule implementation.
func NewMAG02() Rule {
	return mag02Rule{}
}

// ID returns the rule identifier.
func (mag02Rule) ID() string {
	return ruleMAG02
}

// Chapter returns the chapter number for this rule.
func (mag02Rule) Chapter() int {
	return mag02Chapter
}

// Run executes this rule against the provided context.
func (mag02Rule) Run(ctx Context) ([]diag.Finding, error) {
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
			if shortStringLiteral(lit, mag02ShortStringThreshold) {
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

	diagnostics := make([]diag.Finding, 0)
	for literal, occs := range occByValue {
		var nonTestCount int
		var testCount int
		for _, occ := range occs {
			if occ.isTest {
				testCount++
			} else {
				nonTestCount++
			}
		}

		for _, occ := range occs {
			if occ.isTest {
				if testCount < mag02TestRepeatThreshold {
					continue
				}
			} else {
				if nonTestCount < mag02NonTestRepeatThreshold {
					continue
				}
			}

			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleMAG02,
				Severity: diag.SeverityError,
				Message:  "string literal " + literal + mag02MessageRepeatedLiteral,
				Pos:      diag.Position{File: occ.pos.Filename, Line: occ.pos.Line, Col: occ.pos.Column},
				Hint:     mag02HintRepeatedLiteral,
			})
		}
	}

	return diagnostics, nil
}
