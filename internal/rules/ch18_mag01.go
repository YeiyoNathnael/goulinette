package rules

import (
	"go/ast"
	"go/token"

	"goulinette/internal/diag"
)

type mag01Rule struct{}

func NewMAG01() Rule {
	return mag01Rule{}
}

func (mag01Rule) ID() string {
	return "MAG-01"
}

func (mag01Rule) Chapter() int {
	return 18
}

var mag01ExemptNumbers = map[string]struct{}{
	"0":  {},
	"1":  {},
	"2":  {},
	"-1": {},
}

func (mag01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
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
				RuleID:   "MAG-01",
				Severity: diag.SeverityError,
				Message:  "numeric literal " + literal + " is repeated; extract to a named constant",
				Pos:      diag.Position{File: occ.pos.Filename, Line: occ.pos.Line, Col: occ.pos.Column},
				Hint:     "declare a descriptive constant at the narrowest appropriate scope",
			})
		}
	}

	return diagnostics, nil
}
