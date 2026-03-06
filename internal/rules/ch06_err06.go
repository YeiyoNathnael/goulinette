package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type err06Rule struct{}

func NewERR06() Rule {
	return err06Rule{}
}

func (err06Rule) ID() string {
	return "ERR-06"
}

func (err06Rule) Chapter() int {
	return 6
}

func (err06Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, panicCall := range collectCalls(syntaxFile, "panic") {
				if len(panicCall.call.Args) == 0 || !isOperationalPanicArg(panicCall.call.Args[0], pkg.TypesInfo) {
					continue
				}

				pos := pkg.Fset.Position(panicCall.call.Lparen)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "ERR-06",
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
