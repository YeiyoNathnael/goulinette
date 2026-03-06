package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

type err08Rule struct{}

const (
	err08Chapter     = 6
	err08PanicFnName = "panic"
	err08RecoverName = "recover"
)

// NewERR08 returns the ERR08 rule implementation.
func NewERR08() Rule {
	return err08Rule{}
}

// ID returns the rule identifier.
func (err08Rule) ID() string {
	return ruleERR08
}

// Chapter returns the chapter number for this rule.
func (err08Rule) Chapter() int {
	return err08Chapter
}

// Run executes this rule against the provided context.
func (err08Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			panicCalls := collectCalls(syntaxFile, err08PanicFnName)
			recoverCalls := collectCalls(syntaxFile, err08RecoverName)
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
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleERR08,
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
