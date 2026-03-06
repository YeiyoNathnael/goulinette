package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type err07Rule struct{}

func NewERR07() Rule {
	return err07Rule{}
}

func (err07Rule) ID() string {
	return "ERR-07"
}

func (err07Rule) Chapter() int {
	return 6
}

func (err07Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, recoverCall := range collectCalls(syntaxFile, "recover") {
				if isRecoverInDeferredAnonymousFunc(recoverCall) {
					continue
				}

				pos := pkg.Fset.Position(recoverCall.call.Lparen)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "ERR-07",
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
