package rules

import "github.com/YeiyoNathnael/goulinette/internal/diag"

// Context carries the inputs made available to every rule during an
// analysis run. Root is the absolute path of the module root directory;
// Files is the ordered list of .go source files to analyse;
// StrictTools causes tool-backed rules to treat a missing or failing
// external tool as an error rather than silently skipping the check.
type Context struct {
	Root  string
	Files []string

	StrictTools bool
}

// Rule is the interface that every goulinette linting rule must satisfy.
// ID returns the canonical rule identifier (e.g. "NAM-03");
// Chapter returns the numeric chapter the rule belongs to;
// Run analyses the files in ctx and returns zero or more findings.
type Rule interface {
	ID() string
	Chapter() int
	Run(ctx Context) ([]diag.Finding, error)
}
