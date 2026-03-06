package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type err06Rule struct{}

const err06Chapter = 6

// NewERR06 returns the ERR06 rule implementation.
func NewERR06() Rule {
	return err06Rule{}
}

// ID returns the rule identifier.
func (err06Rule) ID() string {
	return ruleERR06
}

// Chapter returns the chapter number for this rule.
func (err06Rule) Chapter() int {
	return err06Chapter
}

// Run executes this rule against the provided context.
func (err06Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, panicCall := range collectCalls(syntaxFile, "panic") {
				if len(panicCall.call.Args) == 0 || !isOperationalPanicArg(panicCall.call.Args[0], pkg.TypesInfo) {
					continue
				}

				pos := pkg.Fset.Position(panicCall.call.Lparen)
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleERR06,
					Severity: diag.SeverityError,
					Message:  "panic should not be used for operational errors; return error instead",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace panic path with explicit error return",
				})
			}
		}
	}

	return diagnostics, nil
}
