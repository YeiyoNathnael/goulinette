package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type err08Rule struct{}

func NewERR08() Rule {
	return err08Rule{}
}

func (err08Rule) ID() string {
	return "ERR-08"
}

func (err08Rule) Chapter() int {
	return 6
}

func (err08Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			panicCalls := collectCalls(syntaxFile, "panic")
			recoverCalls := collectCalls(syntaxFile, "recover")
			if len(panicCalls) == 0 || len(recoverCalls) == 0 {
				continue
			}

			recoverByFunc := map[string]bool{}
			for _, rc := range recoverCalls {
				fname := enclosingFuncName(rc.ancestors)
				if fname != "" {
					recoverByFunc[fname] = true
				}
			}

			for _, pc := range panicCalls {
				fname := enclosingFuncName(pc.ancestors)
				recovered, ok := recoverByFunc[fname]
				if fname == "" || !ok || !recovered {
					continue
				}
				if len(pc.call.Args) == 0 || !isOperationalPanicArg(pc.call.Args[0], pkg.TypesInfo) {
					continue
				}

				pos := pkg.Fset.Position(pc.call.Lparen)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "ERR-08",
					Severity: diag.SeverityError,
					Message:  "panic/recover must not be used as general control flow",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "replace panic/recover flow with explicit returned errors",
				})
			}
		}
	}

	return diagnostics, nil
}
