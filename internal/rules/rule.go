package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

// Context documents this exported type.
type Context struct {
	Root  string
	Files []string

	StrictTools bool
}

// Rule documents this exported type.
type Rule interface {
	ID() string
	Chapter() int
	Run(ctx Context) ([]diag.Finding, error)
}
