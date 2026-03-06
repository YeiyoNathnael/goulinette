package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type mag01Rule struct{}

const (
	mag01Chapter                = 18
	mag01TestRepeatThreshold    = 3
	mag01NonTestRepeatThreshold = 2
)

// NewMAG01 returns the MAG01 rule implementation.
func NewMAG01() Rule {
	return mag01Rule{}
}

// ID returns the rule identifier.
func (mag01Rule) ID() string {
	return ruleMAG01
}

// Chapter returns the chapter number for this rule.
func (mag01Rule) Chapter() int {
	return mag01Chapter
}

var mag01ExemptNumbers = map[string]struct{}{
	"0":  {},
	"1":  {},
	"2":  {},
	"-1": {},
}

// Run executes this rule against the provided context.
func (mag01Rule) Run(ctx Context) ([]diag.Finding, error) {
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
			if !ok {
				return
			}
			if lit.Kind != token.INT && lit.Kind != token.FLOAT {
				return
			}

			parent := astDirectParent(stack)
			key := numericLiteralKey(lit, parent)
			if _, exempt := mag01ExemptNumbers[key]; exempt {
				return
			}
			if isInConstDecl(stack) {
				return
			}

			pos := unit.FSet.Position(lit.Pos())
			occByValue[key] = append(occByValue[key], magLiteralOccurrence{
				value:  key,
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
				if testCount < mag01TestRepeatThreshold {
					continue
				}
			} else {
				if nonTestCount < mag01NonTestRepeatThreshold {
					continue
				}
			}

			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleMAG01,
				Severity: diag.SeverityError,
				Message:  "numeric literal " + literal + " is repeated; extract to a named constant",
				Pos:      diag.Position{File: occ.pos.Filename, Line: occ.pos.Line, Col: occ.pos.Column},
				Hint:     "declare a descriptive constant at the narrowest appropriate scope",
			})
		}
	}

	return diagnostics, nil
}
