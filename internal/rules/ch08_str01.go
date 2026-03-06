package rules

import (
	"go/ast"
	"strings"
	"unicode"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type str01Rule struct{}

func NewSTR01() Rule {
	return str01Rule{}
}

func (str01Rule) ID() string {
	return "STR-01"
}

func (str01Rule) Chapter() int {
	return 8
}

func (str01Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	parsed, err := parseFiles(ctx.Files)
	if err != nil {
		return nil, err
	}

	type methodRec struct {
		filePath     string
		line         int
		col          int
		receiverName string
		receiverType string
		functionName string
	}

	byType := map[string][]methodRec{}
	for _, pf := range parsed {
		for _, decl := range pf.File.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
				continue
			}
			recvField := fn.Recv.List[0]
			if len(recvField.Names) == 0 || recvField.Names[0] == nil {
				continue
			}
			recvName := recvField.Names[0].Name
			typeName := receiverBaseTypeName(recvField.Type)
			if typeName == "" {
				continue
			}

			pos := pf.FSet.Position(recvField.Names[0].Pos())
			byType[typeName] = append(byType[typeName], methodRec{
				filePath:     pos.Filename,
				line:         pos.Line,
				col:          pos.Column,
				receiverName: recvName,
				receiverType: typeName,
				functionName: fn.Name.Name,
			})
		}
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for typeName, methods := range byType {
		names := map[string]int{}
		for _, m := range methods {
			count, ok := names[m.receiverName]
			if !ok {
				names[m.receiverName] = 1
				continue
			}
			names[m.receiverName] = count + 1
		}
		inconsistent := len(names) > 1

		for _, m := range methods {
			if inconsistent {
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "STR-01",
					Severity: diag.SeverityError,
					Message:  "receiver names must be consistent across all methods of the same type",
					Pos:      diag.Position{File: m.filePath, Line: m.line, Col: m.col},
					Hint:     "use one consistent abbreviation for type " + typeName,
				})
				continue
			}

			if isShortReceiverAbbreviation(m.receiverName, typeName) {
				continue
			}

			diagnostics = append(diagnostics, diag.Diagnostic{
				RuleID:   "STR-01",
				Severity: diag.SeverityError,
				Message:  "receiver name must be a short abbreviation of the type name",
				Pos:      diag.Position{File: m.filePath, Line: m.line, Col: m.col},
				Hint:     "use receiver like " + expectedReceiverSuggestion(typeName),
			})
		}
	}

	return diagnostics, nil
}

func receiverBaseTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverBaseTypeName(t.X)
	case *ast.IndexExpr:
		return receiverBaseTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverBaseTypeName(t.X)
	default:
		return ""
	}
}

func isShortReceiverAbbreviation(name, typeName string) bool {
	if name == "" || typeName == "" || len(name) > 3 {
		return false
	}
	if strings.ToLower(name) != name {
		return false
	}

	initial := strings.ToLower(string([]rune(typeName)[0]))
	if name == initial {
		return true
	}

	initials := typeInitials(typeName)
	if len(initials) > 1 && name == initials {
		return true
	}

	return strings.HasPrefix(strings.ToLower(typeName), name)
}

func typeInitials(typeName string) string {
	if typeName == "" {
		return ""
	}
	runes := []rune(typeName)
	out := []rune{unicode.ToLower(runes[0])}
	for i := 1; i < len(runes); i++ {
		if unicode.IsUpper(runes[i]) {
			out = append(out, unicode.ToLower(runes[i]))
		}
	}
	if len(out) > 3 {
		out = out[:3]
	}
	return string(out)
}

func expectedReceiverSuggestion(typeName string) string {
	initials := typeInitials(typeName)
	if initials != "" {
		return initials
	}
	return strings.ToLower(string(typeName[0]))
}
