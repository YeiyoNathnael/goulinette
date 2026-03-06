package rules

import (
	"go/ast"
	"strings"
	"unicode"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type str01Rule struct{}

const (
	str01Chapter             = 8
	str01ReceiverMaxLen      = 3
	str01MessageConsistency  = "receiver names must be consistent across all methods of the same type"
	str01MessageAbbreviation = "receiver name must be a short abbreviation of the type name"
)

// NewSTR01 returns the STR01 rule implementation.
func NewSTR01() Rule {
	return str01Rule{}
}

// ID returns the rule identifier.
func (str01Rule) ID() string {
	return ruleSTR01
}

// Chapter returns the chapter number for this rule.
func (str01Rule) Chapter() int {
	return str01Chapter
}

// Run executes this rule against the provided context.
func (str01Rule) Run(ctx Context) ([]diag.Finding, error) {
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

	diagnostics := make([]diag.Finding, 0)
	for typeName, methods := range byType {
		names := map[string]int{}
		for _, method := range methods {
			count, ok := names[method.receiverName]
			if !ok {
				names[method.receiverName] = 1
				continue
			}
			names[method.receiverName] = count + 1
		}
		inconsistent := len(names) > 1

		for _, method := range methods {
			if inconsistent {
				diagnostics = append(diagnostics, diag.Finding{
					RuleID:   ruleSTR01,
					Severity: diag.SeverityError,
					Message:  str01MessageConsistency,
					Pos:      diag.Position{File: method.filePath, Line: method.line, Col: method.col},
					Hint:     "use one consistent abbreviation for type " + typeName,
				})
				continue
			}

			if isShortReceiverAbbreviation(method.receiverName, typeName) {
				continue
			}

			diagnostics = append(diagnostics, diag.Finding{
				RuleID:   ruleSTR01,
				Severity: diag.SeverityError,
				Message:  str01MessageAbbreviation,
				Pos:      diag.Position{File: method.filePath, Line: method.line, Col: method.col},
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
	if name == "" || typeName == "" || len(name) > str01ReceiverMaxLen {
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
		out = out[:str01ReceiverMaxLen]
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
