package rules

import (
	"fmt"
	"sort"
	"strings"
)

// explainText maps each rule ID to a short rationale that can be printed
// by the --explain flag. Entries cover what the rule checks, why it
// matters, and how to fix the most common violation.
var explainText = map[string]string{
	// ── Formatting ────────────────────────────────────────────────────────
	"FMT-01": "All Go source files must be formatted with gofmt. " +
		"Inconsistent formatting creates noisy diffs and signals tooling gaps. " +
		"Fix: run `gofmt -w .` or enable format-on-save in your editor.",

	"FMT-02": "go vet must report no issues. go vet catches real bugs — " +
		"unreachable code, misused Printf verbs, mutex copies — that the " +
		"compiler allows but that always indicate a mistake. Fix: run `go vet ./...`.",

	"FMT-03": "staticcheck must report no issues. staticcheck finds deprecated " +
		"API usage, dead code, and subtle correctness issues beyond what go vet " +
		"catches. Fix: run `staticcheck ./...` and address each finding.",

	// ── Naming ────────────────────────────────────────────────────────────
	"NAM-01": "Package names must be lowercase, single words with no underscores " +
		"or mixedCase. This follows the Effective Go convention and keeps import " +
		"paths readable. Fix: rename the package, e.g. `mypackage` not `my_package`.",

	"NAM-02": "Exported constant names must not use ALL_CAPS_WITH_UNDERSCORES. " +
		"Go style uses MixedCaps for constants. ALL_CAPS is a C convention that " +
		"has no place in idiomatic Go. Fix: rename `MAX_SIZE` → `MaxSize`.",

	"NAM-03": "Variable names must be proportional to their scope. Short-lived " +
		"loop variables should use brief names (i, j, x); variables used across " +
		"many lines need descriptive names. Fix: rename to match scope length.",

	"NAM-04": "Getter methods must not be prefixed with 'Get'. Go convention " +
		"uses the field name directly as the method name (e.g. `Name()`, not " +
		"`GetName()`). Fix: remove the 'Get' prefix from the method name.",

	"NAM-05": "Interface names should end in '-er' when they describe a single " +
		"behaviour (Reader, Writer, Stringer). This is a strong Go convention. " +
		"Fix: rename the interface to end in 'er', or add it to the allowlist.",

	"NAM-06": "Package names must not be generic stubs like 'util', 'common', " +
		"'helpers', 'misc', or 'shared'. These names reveal nothing about what " +
		"the package does. Fix: rename to something that describes the domain.",

	"NAM-07": "Avoid stuttering names where the package name is repeated in an " +
		"exported identifier (e.g. `http.HTTPClient`). The package qualifier " +
		"already provides context. Fix: rename to `http.Client`.",

	// ── Variables ────────────────────────────────────────────────────────
	"VAR-01": "Variables must not be initialised to their zero value explicitly. " +
		"`var n int = 0` is redundant; use `var n int` or let the zero value " +
		"appear implicitly. Fix: remove the explicit zero initialiser.",

	"VAR-02": "Avoid package-level variables that are mutated after init. Mutable " +
		"globals make code hard to test and reason about concurrently. " +
		"Fix: pass state explicitly or use sync primitives.",

	"VAR-03": "Short variable declarations (:=) must not shadow an outer variable " +
		"of the same name. Shadowing is a common source of subtle bugs where " +
		"the inner variable is modified but the outer one is never updated.",

	"VAR-04": "Boolean variables and parameters must not have negated names like " +
		"'notReady' or 'notFound'. Negated booleans invert readability. " +
		"Fix: rename to the positive form and flip the logic.",

	// ── Control flow ──────────────────────────────────────────────────────
	"CTL-01": "Functions must have a single return point where possible, avoiding " +
		"multiple early returns that make control flow hard to follow. " +
		"Complex functions may use early returns for guard clauses only.",

	"CTL-02": "goto is forbidden. It makes control flow non-local and produces " +
		"code that is nearly impossible to reason about. Fix: restructure " +
		"using loops, functions, or labelled breaks.",

	"CTL-03": "Labels (for goto/break/continue) without a corresponding goto are " +
		"forbidden. Unused labels are code smell; named break/continue targets " +
		"should only exist when they meaningfully clarify nested loop exits.",

	"CTL-04": "Type switches must include a default case. Without a default branch, " +
		"unrecognised types are silently ignored, which can mask bugs when new " +
		"types are added. Fix: add `default:` with an explicit panic or no-op.",

	// ── Functions ────────────────────────────────────────────────────────
	"FUN-01": "Functions must not exceed the maximum line count (default 60 lines). " +
		"Long functions are hard to test, review, and reason about. " +
		"Fix: extract sub-functions with descriptive names.",

	"FUN-02": "The error return value must be the last in the return list. " +
		"This is a strong Go convention enforced by many tools and expected by " +
		"callers. Fix: reorder return values so error is rightmost.",

	"FUN-03": "Functions must not have too many parameters. High parameter counts " +
		"signal missing abstraction. Fix: group related parameters into a struct " +
		"or split the function into smaller pieces.",

	"FUN-04": "Named return values should be used only when they meaningfully " +
		"document the return. Naked returns (bare `return`) in non-trivial " +
		"functions hide what is actually returned. Fix: use explicit returns.",

	// ── Errors ────────────────────────────────────────────────────────────
	"ERR-01": "Errors must not be discarded silently. Ignoring an error with `_` " +
		"or no assignment means failures pass undetected. Fix: handle the error " +
		"or wrap and propagate it with fmt.Errorf.",

	"ERR-02": "Errors must be checked immediately after the call that produces " +
		"them. Delaying the check allows the program to continue in an invalid " +
		"state. Fix: move the error check to immediately follow the call.",

	"ERR-03": "Non-error return values must be zero values when an error is " +
		"returned. Returning a non-zero value alongside an error is misleading — " +
		"callers checking the error still see a non-zero result. Fix: return `\"\"`, `0`, or `nil`.",

	"ERR-04": "Error variables (sentinel errors) must be declared with errors.New " +
		"or fmt.Errorf at package level, not inline. Inline sentinel errors are " +
		"not comparable with errors.Is. Fix: use `var ErrFoo = errors.New(...)`.",

	"ERR-05": "Error strings must not be capitalised or end with punctuation. " +
		"Go error strings are concatenated into larger messages; capitalisation " +
		"and trailing periods break the chain. Fix: use lowercase, no period.",

	"ERR-06": "Custom error types must implement the error interface correctly. " +
		"If Error() returns different content on each call the error cannot be " +
		"tested reliably. Fix: make Error() deterministic.",

	"ERR-07": "errors.Is / errors.As must be used for error comparison, not `==`. " +
		"Direct equality misses wrapped errors. Fix: use `errors.Is(err, target)`.",

	"ERR-08": "panic must not be used in library code except at init time to " +
		"signal programming errors (e.g. nil pointer in constructor). " +
		"Fix: return an error instead, or document the panic invariant.",

	// ── Types ────────────────────────────────────────────────────────────
	"TYP-01": "Function parameters that are pointers only for performance must be " +
		"value types when the struct is small enough. Unnecessary pointers hide " +
		"ownership semantics. Fix: pass by value and benchmark if concerned.",

	"TYP-02": "Avoid returning concrete types from constructors — return interfaces " +
		"or the unexported type to allow future extension without breaking callers. " +
		"Fix: return the interface the caller actually needs.",

	"TYP-03": "Map reads must use the comma-ok form when the key's zero value is " +
		"a meaningful result. `v := m[k]` cannot distinguish a missing key from " +
		"a key explicitly set to zero. Fix: use `v, ok := m[k]; if !ok { ... }`.",

	"TYP-04": "Channel receives must use the comma-ok form when the channel may " +
		"be closed. A receive from a closed channel returns the zero value, which " +
		"is indistinguishable without ok. Fix: use `v, ok := <-ch; if !ok { ... }`.",

	"TYP-05": "Type assertions must use the comma-ok form. A bare `x.(T)` panics " +
		"at runtime if the underlying type is wrong. Fix: use `v, ok := x.(T); if !ok { ... }`.",

	"TYP-06": "Avoid embedding types in structs purely for field promotion when " +
		"the embedded type's full API is not intended to be part of the struct's " +
		"public surface. Fix: use a named field instead of embedding.",

	"TYP-07": "The any / interface{} type must include a justification comment " +
		"explaining why a concrete type cannot be used. Undocumented any is a " +
		"sign of deferred thinking. Fix: add an inline comment or use generics.",

	// ── Strings ──────────────────────────────────────────────────────────
	"STR-01": "Use strings.Builder (or bytes.Buffer) instead of += for string " +
		"concatenation in loops. Repeated += is O(n²) due to allocations. " +
		"Fix: allocate a strings.Builder before the loop and call WriteString.",

	"STR-02": "Use fmt.Sprintf or string conversion instead of strconv.FormatX " +
		"where the intent is clearer. The rule targets systematic misuse patterns " +
		"that obscure intent. Fix: choose the variant that reads most naturally.",

	"STR-03": "String formatting verbs must match the type of the argument. " +
		"Mismatched verbs (%d for a string, %s for an int) produce unexpected " +
		"output. Fix: use the correct verb for the argument type.",

	"STR-04": "Avoid splitting multiline string literals with explicit `\\n` when " +
		"a raw string literal (backtick) would be clearer. Fix: use `` `...` `` " +
		"for strings that span multiple lines or contain backslashes.",

	// ── Documentation ────────────────────────────────────────────────────
	"DOC-01": "Every exported type must have a doc comment. The comment must be " +
		"a full sentence starting with the type name. Without it, godoc and " +
		"IDE hovers are empty. Fix: add `// TypeName does/represents ...`.",

	"DOC-02": "Every exported function and method must have a doc comment starting " +
		"with the function name. Fix: add `// FuncName returns/does ...`.",

	"DOC-03": "Every exported variable and constant must have a doc comment. " +
		"Undocumented exported constants are especially problematic for API " +
		"consumers. Fix: add a short description above the declaration.",

	"DOC-04": "Doc comments must be full sentences (start with a capital letter " +
		"and end with a period). This is required for godoc to render them " +
		"correctly. Fix: capitalise the first word and add a trailing period.",

	"DOC-05": "init functions must have a doc comment explaining what they " +
		"initialise and why at package load time. Side-effectful init is " +
		"surprising; at minimum it must be documented. Fix: add a comment.",

	// ── Slices ────────────────────────────────────────────────────────────
	"SLC-01": "Use nil instead of an empty slice literal (`[]T{}` or `make([]T, 0)`) " +
		"when the caller treats nil and empty slices equivalently. nil is the " +
		"idiomatic zero value for a slice. Fix: replace with `nil` or `var s []T`.",

	// ── Concurrency ──────────────────────────────────────────────────────
	"CON-01": "Exported functions must not accept or return bare channel or mutex " +
		"types. Exposing concurrency primitives in the public API leaks " +
		"implementation details. Fix: wrap them in a higher-level abstraction.",

	"CON-02": "Goroutines must have an obvious cancellation or exit path. A " +
		"goroutine without a way to be stopped leaks forever. Fix: accept a " +
		"context.Context and exit on ctx.Done(), or use an explicit done channel.",

	"CON-03": "A channel must be closed by the same goroutine that writes to it. " +
		"Having a separate goroutine close a channel it did not own leads to " +
		"double-close panics. Fix: let the writer own close, coordinate with WaitGroup.",

	// ── Custom error returns ──────────────────────────────────────────────
	"CER-01": "Functions returning a custom error type must return nil (not a " +
		"nil-typed interface) on success. A non-nil interface wrapping a nil " +
		"pointer is not equal to nil. Fix: return typed nil, not `(*MyError)(nil)`.",

	"CER-02": "Custom error variables must be assigned before first use and not " +
		"left as their zero value at any return point. An unassigned custom error " +
		"signals incomplete error handling logic. Fix: assign the error on every path.",

	"CER-03": "Custom error variables returned from a function must be assigned " +
		"on all code paths. If a variable may reach a return statement while " +
		"still holding its zero value the caller receives a meaningless error.",

	// ── Limits ───────────────────────────────────────────────────────────
	"LIM-01": "Functions must not exceed a maximum line count. Very long functions " +
		"are hard to review and test in isolation. Fix: extract cohesive blocks " +
		"into well-named helper functions.",

	"LIM-02": "Functions must not have more than a fixed number of parameters " +
		"(default 4). High parameter counts are a sign that the function is doing " +
		"too much. Fix: introduce a config/options struct.",

	"LIM-03": "Functions must not have more than a fixed number of results. " +
		"Many return values make call sites cluttered. Fix: return a struct, " +
		"or split into multiple focused functions.",

	"LIM-04": "Nesting depth must not exceed a maximum (default 4 levels). Deep " +
		"nesting makes control flow hard to follow. Fix: extract inner blocks " +
		"into separate functions, or invert conditions to reduce depth.",

	// ── Context ──────────────────────────────────────────────────────────
	"CTX-01": "context.Context must be the first parameter of any function that " +
		"accepts one. This is a hard Go convention. Fix: move ctx to position 0.",

	"CTX-02": "context.Background() or context.TODO() must not be used inside " +
		"functions that already receive a context. Pass the received context " +
		"downstream. Fix: replace context.Background() with the parameter ctx.",

	"CTX-03": "context.Background() or context.TODO() must not be passed to " +
		"functions that accept context.Context as an argument when a non-background " +
		"context is available. Fix: pass the received context instead.",

	"CTX-04": "The cancel function returned by context.WithCancel/WithTimeout/" +
		"WithDeadline must be called or deferred on all code paths. Leaking a " +
		"cancel func leaks the child context's resources. Fix: defer cancel() " +
		"immediately after creation.",

	// ── Imports ──────────────────────────────────────────────────────────
	"IMP-01": "Imports must be grouped into stdlib, external, and internal blocks " +
		"separated by blank lines. Mixed import groups make dependencies hard " +
		"to scan. Fix: reorganise with goimports or your editor's organise-imports.",

	"IMP-02": "Blank imports (`import _ \"pkg\"`) must be in a file whose sole " +
		"purpose is side-effect registration, not scattered across production files. " +
		"Fix: centralise blank imports in a dedicated file (e.g. imports.go).",

	"IMP-03": "Dot imports (`import . \"pkg\"`) are forbidden in non-test code. " +
		"They pollute the package namespace and make it impossible to tell where " +
		"an identifier is defined. Fix: use a regular import and the package qualifier.",

	// ── Resources ────────────────────────────────────────────────────────
	"RES-01": "Resources that implement io.Closer (files, HTTP bodies, DB rows) " +
		"must be closed, typically with defer. An unclosed resource leaks file " +
		"descriptors or connections. Fix: add `defer rc.Close()` after checking open error.",

	"RES-02": "HTTP response bodies must be drained before closing to allow " +
		"connection reuse. A body closed without reading prevents Keep-Alive. " +
		"Fix: add `io.Copy(io.Discard, resp.Body)` before close.",

	// ── Safety ────────────────────────────────────────────────────────────
	"SAF-01": "sync.Mutex, sync.WaitGroup, and other sync types must not be copied " +
		"after first use. Copying a mutex copies its internal state, breaking the lock. " +
		"Fix: pass a pointer, or embed the type in a struct accessed by pointer.",

	"SAF-02": "Structs containing copy-sensitive sync values must not be returned " +
		"by value. Returning by value implicitly copies the mutex/WaitGroup. " +
		"Fix: return a pointer to the struct.",

	// ── Magic values ──────────────────────────────────────────────────────
	"MAG-01": "Numeric literals other than 0 and 1 must be named constants. A " +
		"bare number like `86400` has no self-documenting value. " +
		"Fix: `const secondsPerDay = 86400`.",

	"MAG-02": "String literals used more than once must be extracted to a named " +
		"constant. Repeated string literals are fragile: a typo in one copy is " +
		"not caught at compile time. Fix: `const fooKey = \"foo\"` and use the const.",

	// ── Tests ────────────────────────────────────────────────────────────
	"TST-01": "Test helper functions must call t.Helper() as their first statement. " +
		"Without t.Helper(), failure messages point to the helper instead of the " +
		"call site. Fix: add `t.Helper()` at the top of every test helper.",

	"TST-02": "Tests must not use t.Parallel() without also accepting a *testing.T " +
		"parameter in table-driven subtests. Parallel subtests that close over " +
		"loop variables cause data races. Fix: capture the loop variable explicitly.",

	"TST-03": "Tests must not use time.Sleep for synchronisation. Sleep-based tests " +
		"are flaky and slow. Fix: use channels, sync.WaitGroup, or context deadlines " +
		"to wait for the condition you actually need.",
}

// Explain returns the rationale for the given rule ID and whether the rule
// is known. The id comparison is case-insensitive.
func Explain(id string) (string, bool) {
	text, ok := explainText[strings.ToUpper(id)]
	return text, ok
}

// ExplainAll returns all known rule IDs in sorted order alongside their
// rationale text. It is used when --explain is invoked without an argument.
func ExplainAll() []ExplainEntry {
	out := make([]ExplainEntry, 0, len(explainText))
	for id, text := range explainText {
		out = append(out, ExplainEntry{ID: id, Text: text})
	}
	sortExplainEntries(out)
	return out
}

// ExplainEntry pairs a rule ID with its explanation text.
type ExplainEntry struct {
	ID   string
	Text string
}

func sortExplainEntries(entries []ExplainEntry) {
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].ID < entries[j].ID
	})
}

// PrintExplain writes a formatted explanation for the given rule ID to w.
// It returns an error message (non-empty) when the ID is not recognised.
func PrintExplain(id string) string {
	text, ok := Explain(id)
	if !ok {
		known := make([]string, 0, len(explainText))
		for k := range explainText {
			known = append(known, k)
		}
		sort.Strings(known)
		return fmt.Sprintf("unknown rule %q\n\nKnown rules:\n  %s", id, strings.Join(known, "\n  "))
	}
	return fmt.Sprintf("%s\n\n%s", strings.ToUpper(id), text)
}
