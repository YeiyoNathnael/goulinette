package rules

import (
	"go/ast"
	"go/token"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	builtinTypeBool       = "bool"
	builtinTypeString     = "string"
	builtinTypeByte       = "byte"
	builtinTypeRune       = "rune"
	builtinTypeInt        = "int"
	builtinTypeInt8       = "int8"
	builtinTypeInt16      = "int16"
	builtinTypeInt32      = "int32"
	builtinTypeInt64      = "int64"
	builtinTypeUint       = "uint"
	builtinTypeUint8      = "uint8"
	builtinTypeUint16     = "uint16"
	builtinTypeUint32     = "uint32"
	builtinTypeUint64     = "uint64"
	builtinTypeUintptr    = "uintptr"
	builtinTypeFloat32    = "float32"
	builtinTypeFloat64    = "float64"
	builtinTypeComplex64  = "complex64"
	builtinTypeComplex128 = "complex128"

	boolLiteralTrue  = "true"
	boolLiteralFalse = "false"
)

var builtinTypeNames = map[string]struct{}{
	builtinTypeBool:       {},
	builtinTypeString:     {},
	builtinTypeByte:       {},
	builtinTypeRune:       {},
	builtinTypeInt:        {},
	builtinTypeInt8:       {},
	builtinTypeInt16:      {},
	builtinTypeInt32:      {},
	builtinTypeInt64:      {},
	builtinTypeUint:       {},
	builtinTypeUint8:      {},
	builtinTypeUint16:     {},
	builtinTypeUint32:     {},
	builtinTypeUint64:     {},
	builtinTypeUintptr:    {},
	builtinTypeFloat32:    {},
	builtinTypeFloat64:    {},
	builtinTypeComplex64:  {},
	builtinTypeComplex128: {},
}

type var02Rule struct{}

const var02Chapter = 3

// NewVAR02 returns the VAR02 rule implementation.
func NewVAR02() Rule {
	return var02Rule{}
}

// ID returns the rule identifier.
func (var02Rule) ID() string {
	return ruleVAR02
}

// Chapter returns the chapter number for this rule.
func (var02Rule) Chapter() int {
	return var02Chapter
}

// Run executes this rule against the provided context.
func (var02Rule) Run(ctx Context) ([]diag.Finding, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Finding, 0)
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

					defaultType, isLiteral := defaultLiteralType(rhs)
					if !isLiteral {
						continue
					}

					if conversionType, found := findPostDeclConversionNeed(body, assign.End(), lhs, defaultType); found {
						pos := pf.FSet.Position(lhs.Pos())
						diagnostics = append(diagnostics, diag.Finding{
							RuleID:   ruleVAR02,
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
			return builtinTypeInt, true
		case token.FLOAT:
			return builtinTypeFloat64, true
		case token.IMAG:
			return builtinTypeComplex128, true
		case token.CHAR:
			return builtinTypeRune, true
		case token.STRING:
			return builtinTypeString, true
		}
	case *ast.Ident:
		if e.Name == boolLiteralTrue || e.Name == boolLiteralFalse {
			return builtinTypeBool, true
		}
	default:
		return "", false
	}

	return "", false
}

func findPostDeclConversionNeed(body *ast.BlockStmt, after token.Pos, ident *ast.Ident, defaultType string) (string, bool) {
	var conversionType string
	var found bool
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
	_, ok := builtinTypeNames[name]
	return ok
}
