package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type var02Rule struct{}

func NewVAR02() Rule {
	return var02Rule{}
}

func (var02Rule) ID() string {
	return "VAR-02"
}

func (var02Rule) Chapter() int {
	return 3
}

func (var02Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pf := range parsed {
		ast.Inspect(pf.File, func(n ast.Node) bool {
			body, ok := n.(*ast.BlockStmt)
			if !ok {
				return true
			}

			for _, stmt := range body.List {
				assign, ok := stmt.(*ast.AssignStmt)
				if !ok || assign.Tok != token.DEFINE {
					continue
				}

				for i, rhs := range assign.Rhs {
					if i >= len(assign.Lhs) {
						continue
					}
					lhs, ok := assign.Lhs[i].(*ast.Ident)
					if !ok || lhs.Name == "_" {
						continue
					}

					defaultType, isUntypedLiteral := defaultLiteralType(rhs)
					if !isUntypedLiteral {
						continue
					}

					if conversionType, found := findPostDeclConversionNeed(body, assign.End(), lhs, defaultType); found {
						pos := pf.FSet.Position(lhs.Pos())
						diagnostics = append(diagnostics, diag.Diagnostic{
							RuleID:   "VAR-02",
							Severity: diag.SeverityError,
							Message:  "literal short declaration infers default type but later usage expects a different target type",
							Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
							Hint:     "use explicit typed var, e.g. var " + lhs.Name + " " + conversionType + " = <literal>",
						})
					}
				}
			}

			return true
		})
	}

	return diagnostics, nil
}

func defaultLiteralType(expr ast.Expr) (string, bool) {
	switch e := expr.(type) {
	case *ast.BasicLit:
		switch e.Kind {
		case token.INT:
			return "int", true
		case token.FLOAT:
			return "float64", true
		case token.IMAG:
			return "complex128", true
		case token.CHAR:
			return "rune", true
		case token.STRING:
			return "string", true
		}
	case *ast.Ident:
		if e.Name == "true" || e.Name == "false" {
			return "bool", true
		}
	}

	return "", false
}

func findPostDeclConversionNeed(body *ast.BlockStmt, after token.Pos, ident *ast.Ident, defaultType string) (string, bool) {
	var conversionType string
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found || n == nil {
			return !found
		}

		call, ok := n.(*ast.CallExpr)
		if !ok || call.Pos() <= after || len(call.Args) != 1 {
			return true
		}

		typeIdent, ok := call.Fun.(*ast.Ident)
		if !ok || !isBuiltinTypeName(typeIdent.Name) {
			return true
		}
		if typeIdent.Name == defaultType {
			return true
		}

		argIdent, ok := call.Args[0].(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Obj != nil && argIdent.Obj != nil {
			if argIdent.Obj != ident.Obj {
				return true
			}
		} else if argIdent.Name != ident.Name {
			return true
		}

		conversionType = typeIdent.Name
		found = true
		return false
	})

	return conversionType, found
}

func isBuiltinTypeName(name string) bool {
	switch name {
	case "bool", "string", "byte", "rune", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "float32", "float64", "complex64", "complex128":
		return true
	default:
		return false
	}
}
