package rules

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

type con03Rule struct{}

func NewCON03() Rule {
	return con03Rule{}
}

func (con03Rule) ID() string {
	return "CON-03"
}

func (con03Rule) Chapter() int {
	return 11
}

func (con03Rule) Run(ctx Context) ([]diag.Diagnostic, error) {
	pkgs, err := loadTypedPackages(ctx.Root)
	if err != nil {
		return nil, err
	}

	diagnostics := make([]diag.Diagnostic, 0)
	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Body == nil {
					continue
				}

				analysis := analyzeChannelOwnership(fn.Body, pkg.TypesInfo)
				diagnostics = append(diagnostics, con03DiagnosticsForOwnership(pkg.Fset, analysis)...)
			}
		}
	}

	return diagnostics, nil
}

type closeEvent struct {
	contextID int
	pos       token.Pos
}

type funcChannelAnalysis struct {
	writesByChannel map[string]map[int]bool
	closesByChannel map[string][]closeEvent
	hasWaitCall     bool
}

func analyzeChannelOwnership(body *ast.BlockStmt, info *types.Info) funcChannelAnalysis {
	result := funcChannelAnalysis{
		writesByChannel: make(map[string]map[int]bool),
		closesByChannel: make(map[string][]closeEvent),
	}

	nextContextID := 1
	var walk func(node ast.Node, contextID int)
	walk = func(node ast.Node, contextID int) {
		if node == nil {
			return
		}

		switch n := node.(type) {
		case *ast.GoStmt:
			if lit, ok := n.Call.Fun.(*ast.FuncLit); ok {
				childID := nextContextID
				nextContextID++
				walk(lit.Body, childID)
			}
			return

		case *ast.SendStmt:
			if chID, ok := channelIdentity(n.Chan, info); ok {
				if _, exists := result.writesByChannel[chID]; !exists {
					result.writesByChannel[chID] = make(map[int]bool)
				}
				result.writesByChannel[chID][contextID] = true
			}

		case *ast.CallExpr:
			if isWaitGroupWaitCall(n, info) {
				result.hasWaitCall = true
			}
			if chExpr, ok := closeCallTarget(n); ok {
				if chID, ok := channelIdentity(chExpr, info); ok {
					result.closesByChannel[chID] = append(result.closesByChannel[chID], closeEvent{contextID: contextID, pos: n.Lparen})
				}
			}
		}

		ast.Inspect(node, func(child ast.Node) bool {
			if child == node {
				return true
			}
			if goStmt, ok := child.(*ast.GoStmt); ok {
				walk(goStmt, contextID)
				return false
			}
			if call, ok := child.(*ast.CallExpr); ok {
				if isWaitGroupWaitCall(call, info) {
					result.hasWaitCall = true
				}
				if chExpr, ok := closeCallTarget(call); ok {
					if chID, ok := channelIdentity(chExpr, info); ok {
						result.closesByChannel[chID] = append(result.closesByChannel[chID], closeEvent{contextID: contextID, pos: call.Lparen})
					}
				}
				return true
			}
			if send, ok := child.(*ast.SendStmt); ok {
				if chID, ok := channelIdentity(send.Chan, info); ok {
					if _, exists := result.writesByChannel[chID]; !exists {
						result.writesByChannel[chID] = make(map[int]bool)
					}
					result.writesByChannel[chID][contextID] = true
				}
			}
			return true
		})
	}

	walk(body, 0)
	return result
}

func con03DiagnosticsForOwnership(fset *token.FileSet, analysis funcChannelAnalysis) []diag.Diagnostic {
	diagnostics := make([]diag.Diagnostic, 0)

	for chID, closes := range analysis.closesByChannel {
		writers := analysis.writesByChannel[chID]
		if len(writers) == 0 {
			continue
		}

		multiWriter := len(writers) > 1
		for _, closeEv := range closes {
			wrote, ok := writers[closeEv.contextID]
			if !ok || !wrote {
				pos := fset.Position(closeEv.pos)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "CON-03",
					Severity: diag.SeverityError,
					Message:  "channel appears to be closed by a different goroutine than the writer",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "prefer explicit ownership where writers coordinate close",
				})
			}

			if multiWriter && !analysis.hasWaitCall {
				pos := fset.Position(closeEv.pos)
				diagnostics = append(diagnostics, diag.Diagnostic{
					RuleID:   "CON-03",
					Severity: diag.SeverityError,
					Message:  "multiple goroutines write to channel without obvious WaitGroup coordination before close",
					Pos:      diag.Position{File: pos.Filename, Line: pos.Line, Col: pos.Column},
					Hint:     "use sync.WaitGroup and close channel only after wg.Wait()",
				})
			}
		}
	}

	return diagnostics
}

func closeCallTarget(call *ast.CallExpr) (ast.Expr, bool) {
	if call == nil {
		return nil, false
	}
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "close" || len(call.Args) != 1 {
		return nil, false
	}
	return call.Args[0], true
}

func channelIdentity(expr ast.Expr, info *types.Info) (string, bool) {
	if expr == nil || info == nil {
		return "", false
	}
	if id, ok := expr.(*ast.Ident); ok {
		if obj := info.ObjectOf(id); obj != nil {
			if _, ok := obj.Type().Underlying().(*types.Chan); ok {
				return fmt.Sprintf("%p", obj), true
			}
		}
	}
	t := info.TypeOf(expr)
	if t == nil {
		return "", false
	}
	if _, ok := t.Underlying().(*types.Chan); !ok {
		return "", false
	}
	return fmt.Sprintf("expr:%d", expr.Pos()), true
}

func isWaitGroupWaitCall(call *ast.CallExpr, info *types.Info) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != "Wait" {
		return false
	}
	recvType := info.TypeOf(sel.X)
	if recvType == nil {
		return false
	}
	if ptr, ok := recvType.(*types.Pointer); ok {
		recvType = ptr.Elem()
	}
	named, ok := recvType.(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return false
	}
	return named.Obj().Pkg().Path() == "sync" && named.Obj().Name() == "WaitGroup"
}
