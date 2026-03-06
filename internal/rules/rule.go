package rules

import "goulinette/internal/diag"

type Context struct {
	Root  string
	Files []string

	StrictTools bool
}

type Rule interface {
	ID() string
	Chapter() int
	Run(ctx Context) ([]diag.Diagnostic, error)
}
