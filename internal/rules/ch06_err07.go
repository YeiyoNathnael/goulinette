package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type err07Rule struct{}

const err07Chapter = 6

// NewERR07 returns the ERR07 rule implementation.
func NewERR07() Rule {
	return err07Rule{}
}

// ID returns the rule identifier.
func (err07Rule) ID() string {
	return ruleERR07
}

// Chapter returns the chapter number for this rule.
func (err07Rule) Chapter() int {
	return err07Chapter
}

// Run executes this rule against the provided context.
func (err07Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, recoverCall := range collectCalls(syntaxFile, "recover") {
				if isRecoverInDeferredAnonymousFunc(recoverCall) {
					continue
				}

				pos := pkg.Fset.Position(recoverCall.call.Lparen)
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleERR07,
					Severity: diag.SeverityError,
					Message:  "recover must be called within a deferred anonymous function",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "use defer func(){ if r := recover(); r != nil { ... } }()",
				})
			}
		}
	}

	return diagnostics, nil
}
