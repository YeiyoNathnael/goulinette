package rules

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type err03Rule struct{}

const err03Chapter = 6

// NewERR03 returns the ERR03 rule implementation.
func NewERR03() Rule {
	return err03Rule{}
}

// ID returns the rule identifier.
func (err03Rule) ID() string {
	return ruleERR03
}

// Chapter returns the chapter number for this rule.
func (err03Rule) Chapter() int {
	return err03Chapter
}

// Run executes this rule against the provided context.
func (err03Rule) Run(ctx Context) ([]diag.Finding, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
	for _, pkg := range pkgs {
		for _, syntaxFile := range pkg.Syntax {
			for _, decl := range syntaxFile.Decls {
				diagnostics = append(diagnostics, err03DiagnosticsForDecl(pkg.Fset, pkg.TypesInfo, decl)...)
			}
		}
	}

	return diagnostics, nil
}

func err03DiagnosticsForDecl(fset *token.FileSet, info *types.Info, decl ast.Decl) []diag.Finding {
	fn, ok := decl.(*ast.FuncDecl)
	if !ok || fn.Body == nil || fn.Type == nil || fn.Name == nil {
		return nil
	}

	sig, errorIdx, ok := err03SignatureWithErrorLast(info, fn)
	if !ok {
		return nil
	}

	diagnostics := make([]diag.Finding, 0)
	for _, ret := range collectReturnsExcludingNestedFuncs(fn.Body) {
		finding, ok := err03FindingForReturn(fset, info, sig, errorIdx, ret)
		if ok {
			diagnostics = append(diagnostics, finding)
		}
	}

	return diagnostics
}

func err03SignatureWithErrorLast(info *types.Info, fn *ast.FuncDecl) (*types.Signature, int, bool) {
	obj, ok := info.Defs[fn.Name].(*types.Func)
	if !ok {
		return nil, 0, false
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return nil, 0, false
	}

	results := sig.Results()
	if results == nil || results.Len() == 0 {
		return nil, 0, false
	}

	errorIdx := results.Len() - 1
	if !isErrorType(results.At(errorIdx).Type()) {
		return nil, 0, false
	}

	return sig, errorIdx, true
}

func err03FindingForReturn(fset *token.FileSet, info *types.Info, sig *types.Signature, errorIdx int, ret *ast.ReturnStmt) (diag.Finding, bool) {
	results := sig.Results()
	if len(ret.Results) != results.Len() {
		return diag.Finding{}, false
	}

	errState := classifyReturnedErrorExpr(info, ret.Results[errorIdx])
	if errState == returnedErrorNil {
		return diag.Finding{}, false
	}

	var hasNonZero bool
	for i := 0; i < results.Len()-1; i++ {
		if !isZeroValueExprForType(info, ret.Results[i], results.At(i).Type()) {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		return diag.Finding{}, false
	}

	severity := diag.SeverityWarning
	message := "when returning a non-nil error, all other return values must be zero values"
	hint := "return zero values in all non-error positions on error paths"
	if errState == returnedErrorUnknown {
		message = "possible non-zero values returned alongside a potentially non-nil error"
		hint = "if error can be non-nil here, ensure all other return values are zero"
	}

	pos := fset.Position(ret.Return)
	return diag.Finding{
		RuleID:   ruleERR03,
		Severity: severity,
		Message:  message,
		Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
		Hint:     hint,
	}, true
}

type returnedErrorState int

const (
	returnedErrorNil returnedErrorState = iota
	returnedErrorNonNil
	returnedErrorUnknown
)

func classifyReturnedErrorExpr(info *types.Info, expr ast.Expr) returnedErrorState {
	if id, ok := expr.(*ast.Ident); ok && id.Name == "nil" {
		return returnedErrorNil
	}

	if call, ok := expr.(*ast.CallExpr); ok {
		if isErrorsNewCall(call) || isFmtErrorfCall(call) {
			return returnedErrorNonNil
		}
	}

	typ := info.TypeOf(expr)
	if isErrorType(typ) {
		return returnedErrorUnknown
	}

	return returnedErrorUnknown
}

func collectReturnsExcludingNestedFuncs(body *ast.BlockStmt) []*ast.ReturnStmt {
	out := make([]*ast.ReturnStmt, 0)
	if body == nil {
		return out
	}

	walk := func(n ast.Node) {
		if n == nil {
			return
		}
		switch x := n.(type) {
		case *ast.FuncLit:
			return
		case *ast.ReturnStmt:
			out = append(out, x)
			return
		default:
			// no-op
		}

		ast.Inspect(n, func(child ast.Node) bool {
			if child == n {
				return true
			}
			if _, isFuncLit := child.(*ast.FuncLit); isFuncLit {
				return false
			}
			if ret, ok := child.(*ast.ReturnStmt); ok {
				out = append(out, ret)
				return false
			}
			return true
		})
	}

	walk(body)
	return out
}

func isZeroValueExprForType(info *types.Info, expr ast.Expr, expected types.Type) bool {
	if id, ok := expr.(*ast.Ident); ok {
		if id.Name == "nil" {
			return isNilableType(expected)
		}
		if id.Name == "false" {
			if b, ok := expected.Underlying().(*types.Basic); ok {
				return (b.Info() & types.IsBoolean) != 0
			}
		}
	}

	if bl, ok := expr.(*ast.BasicLit); ok {
		switch bl.Kind {
		case token.INT, token.FLOAT, token.IMAG, token.CHAR:
			if b, ok := expected.Underlying().(*types.Basic); ok {
				if (b.Info() & (types.IsInteger | types.IsFloat | types.IsComplex)) == 0 {
					return false
				}
				if tv, ok := info.Types[expr]; ok && tv.Value != nil {
					return constant.Sign(tv.Value) == 0
				}
			}
		case token.STRING:
			if b, ok := expected.Underlying().(*types.Basic); ok {
				if (b.Info() & types.IsString) == 0 {
					return false
				}
				if tv, ok := info.Types[expr]; ok && tv.Value != nil {
					return constant.StringVal(tv.Value) == ""
				}
			}
		}
	}

	if cl, ok := expr.(*ast.CompositeLit); ok {
		if len(cl.Elts) != 0 {
			return false
		}
		switch expected.Underlying().(type) {
		case *types.Struct, *types.Array:
			return true
		default:
			// no-op
		}
	}

	return false
}

func isNilableType(t types.Type) bool {
	if t == nil {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Signature, *types.Interface:
		return true
	default:
		return false
	}
}
