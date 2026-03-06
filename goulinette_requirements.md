# Go Moulinette — Requirements Specification

> An automated static analysis tool for enforcing idiomatic Go code style, structure, and safety conventions.
> Every rule below must be implemented as an enforceable check. Rules marked **Error** block submission.
> Rules marked **Warning** surface a diagnostic but do not block.

---

## Table of Contents

1. [Formatting & Tooling](#1-formatting--tooling)
2. [Naming Conventions](#2-naming-conventions)
3. [Variable Declarations](#3-variable-declarations)
4. [Control Structures](#4-control-structures)
5. [Functions & Return Values](#5-functions--return-values)
6. [Error Handling](#6-error-handling)
7. [Types, Interfaces & Pointers](#7-types-interfaces--pointers)
8. [Methods & Structs](#8-methods--structs)
9. [Packages & Documentation](#9-packages--documentation)
10. [Slices & Collections](#10-slices--collections)
11. [Concurrency](#11-concurrency)
12. [Custom Errors](#12-custom-errors)
13. [Structural Limits](#13-structural-limits)
14. [Context Handling](#14-context-handling)
15. [Import Organization](#15-import-organization)
16. [Resource Management](#16-resource-management)
17. [Concurrency Safety](#17-concurrency-safety)
18. [Magic Values](#18-magic-values)
19. [Testing](#19-testing)

---

## 1. Formatting & Tooling

---

#### FMT-01 — Code must be formatted with `go fmt` · **Error**

All submitted Go source files must be processed through `go fmt` prior to evaluation. The Go toolchain enforces a canonical, opinionated formatting style that eliminates all debate over brace placement, indentation width, and spacing. The moulinette must compare the submitted file byte-for-byte against the output of `gofmt -l` and fail immediately if any diff is produced. This is non-negotiable: unformatted code is considered unreadable by the Go community regardless of its logical correctness.

---

#### FMT-02 — Code must pass `go vet` with zero diagnostics · **Error**

`go vet` is the official Go static analysis tool. It catches a category of bugs that are syntactically valid but semantically wrong — things the compiler will accept but that represent clear programming mistakes, such as mismatched `Printf` format strings and argument counts, unreachable code after a `return`, suspicious composite literal usage, and incorrect uses of `sync` primitives. The moulinette must run `go vet ./...` and treat any single diagnostic as a fatal violation.

---

#### FMT-03 — Code should pass `staticcheck` or `golangci-lint` · **Warning**

Beyond `go vet`, the Go ecosystem has mature third-party linters that catch a much deeper class of issues. `staticcheck` performs data-flow analysis and catches things like deprecated API usage, redundant type conversions, unreachable branches, and incorrect uses of `context`. `golangci-lint` is a meta-runner that can orchestrate over 50 such tools in a single pass. The moulinette should run at minimum `staticcheck ./...` and surface its findings as warnings.

---

## 2. Naming Conventions

---

#### NAM-01 — All identifiers must use `camelCase`, never `snake_case` · **Error**

Go has a single, unified naming convention: `camelCase` for unexported identifiers, and `PascalCase` for exported ones. The underscore character has exactly two legal uses in Go: the blank identifier `_`, and the naming of test files (`_test.go`). Any variable, constant, function, type, or method name that contains an underscore as part of the name is a violation. Examples of illegal names: `user_id`, `parse_response`, `max_retries`.

---

#### NAM-02 — Constants must not be written in `ALL_CAPS` · **Error**

In most languages, constants are named in `ALL_CAPS_WITH_UNDERSCORES` (e.g., `MAX_SIZE`, `DEFAULT_TIMEOUT`). Go explicitly rejects this convention. Because Go uses the case of the *first letter* alone to determine export visibility, writing a constant in all caps unnecessarily exports it and signals something that Go itself does not distinguish. Constants follow the same `camelCase` / `PascalCase` rules as every other identifier. `maxRetries` and `MaxRetries` are correct; `MAX_RETRIES` is a violation.

---

#### NAM-03 — Within functions, variable names must be proportional to their scope · **Warning**

The shorter the scope of a variable, the shorter its name should be. Loop indices should be single letters (`i`, `j`, `k`). Map iteration variables should be (`k`, `v`) for key/value. Variables used across only a few lines may be abbreviated. Only variables that live across large blocks of logic or that carry significant semantic meaning warrant full, descriptive names. Naming a loop counter `loopIndex` is over-engineering; naming a function-scoped connection pool `p` is under-engineering. The moulinette should flag single-letter names used for variables with a scope exceeding 20 lines, and flag long names (over 15 characters) for variables whose scope is shorter than 5 lines.

---

#### NAM-04 — Package-level identifiers must use full, descriptive names · **Warning**

Package-level variables and constants are part of a package's public or internal contract. They persist for the lifetime of the program and may be referenced from many call sites. They must be named clearly enough that their purpose is self-evident without reading their declaration. Abbreviations that are obvious in a 5-line function body become confusing at the package level. `defaultTimeout` is acceptable; `dt` is not.

---

#### NAM-05 — Interface names must use the `-er` suffix · **Warning**

Go's convention for naming single-method interfaces is to name them after the method with an `-er` suffix appended: a type with a `Read` method becomes `Reader`, a type with a `Write` method becomes `Writer`, a type with a `Close` method becomes `Closer`. For interfaces with multiple methods, a descriptive noun is acceptable, but the `-er` pattern should still be preferred whenever it reads naturally. The moulinette should flag any interface declaration whose name does not end in `er` and surface it as a warning for the developer to evaluate.

---

#### NAM-06 — Package names must be descriptive nouns; generic names are forbidden · **Error**

A package name is part of every call site that uses it: `http.Get`, `json.Marshal`, `time.Now`. The name must make it obvious what the package creates, manages, or transforms. Generic catch-all names like `util`, `helpers`, `common`, `misc`, `shared`, or `tools` are forbidden — they say nothing about what the package actually does and tend to become dumping grounds for unrelated logic. If a package cannot be named without using one of these words, it is a signal that the package should be split or reorganized.

---

#### NAM-07 — Exported identifiers must not stutter the package name · **Error**

Because a package's exported identifiers are always qualified by the package name at the call site, repeating the package name inside the identifier name is redundant and reads awkwardly. A function named `ExtractNames` in a package named `names` is called as `names.ExtractNames` — the word "names" appears twice for no reason. The correct name is simply `Extract`, resulting in the clean call `names.Extract`. The moulinette must check all exported identifiers and flag any whose name contains the package name as a prefix or suffix (case-insensitively).

---

## 3. Variable Declarations

---

#### VAR-01 — Use `var` when declaring a variable at its zero value · **Error**

When a variable is being declared with the intent of starting at its type's zero value — `0` for integers, `""` for strings, `false` for booleans, `nil` for pointers and slices — the correct form is `var x T`. Using `:=` with an explicit zero literal (e.g., `x := 0`, `x := ""`, `x := false`) is misleading: it implies a deliberate assignment of a specific value, when in reality the developer simply wants an empty starting point. The `var` form communicates intent clearly and must be enforced.

---

#### VAR-02 — Use `var` when the default type of a literal is wrong · **Error**

When an untyped constant or a literal is assigned to a variable where you need a type other than the literal's default type, `var` with an explicit type must be used. For example, `var b byte = 20` declares `b` as a `byte`, whereas `b := 20` would infer `int`. This distinction matters for numeric types especially and prevents subtle type mismatch bugs at call sites. The moulinette should detect cases where `:=` is used with a literal and the inferred type would differ from what a typed `var` declaration would produce.

---

#### VAR-03 — Use `:=` for all other declarations inside functions · **Warning**

Within a function body, all variable declarations that are not zero-value initializations and not type-disambiguation cases (see VAR-01, VAR-02) must use the short declaration form `:=`. Writing `var x = someFunc()` inside a function is unnecessarily verbose and runs counter to Go's preference for conciseness within local scopes.

---

#### VAR-04 — Mutable package-level variables are forbidden · **Error**

Package-level variables that can be mutated after initialization represent hidden global state. They make data flow invisible, create implicit dependencies between parts of the program, and make concurrent access nearly impossible to reason about safely. Any `var` declaration at the package level that is not a constant and is not assigned once and never written to again is a violation. If configuration or shared state is needed, it must be passed explicitly through function parameters or encapsulated in a struct.

---

## 4. Control Structures

---

#### CTL-01 — Long `if/else if` chains must be replaced with a blank `switch` · **Warning**

When a sequence of `if/else if` conditions all compare against the same variable, or evaluate a set of related boolean expressions, a blank `switch` statement expresses the same logic more clearly and with less visual noise. The relationship between cases becomes explicit, and the structure signals to the reader that these conditions are meant to be evaluated as a mutually exclusive group. The moulinette should flag any `if/else if` chain of three or more branches where the conditions share a common subject and suggest replacing it with a blank `switch`.

---

#### CTL-02 — The `fallthrough` keyword is forbidden · **Error**

`fallthrough` in a Go `switch` case unconditionally transfers control to the next case's body, bypassing its condition check entirely. This is a relic of C-style switch semantics and almost never reflects actual intent in Go code. It creates hidden dependencies between cases, making the switch block behave in a way that is surprising to any reader who does not already know about the `fallthrough`. Logic that appears to require `fallthrough` must be restructured — either by combining the cases (`case a, b:`) or by extracting shared logic into a helper function.

---

#### CTL-03 — The `goto` statement is forbidden · **Error**

`goto` is an unconditional jump to an arbitrary label within the same function. It makes execution flow non-linear and unpredictable, forcing the reader to mentally simulate a state machine rather than following sequential logic. Go provides every structured control flow construct needed to express any algorithm cleanly: `for`, `break`, `continue`, `return`, labeled breaks for nested loops, and `defer`. There is no scenario in idiomatic Go where `goto` is the appropriate tool. Its use is a violation with no exceptions.

---

#### CTL-04 — `type switch` statements must always include a `default:` case · **Error**

A `type switch` inspects the dynamic type behind an interface value. If a new concrete type implementing that interface is added in the future, a `type switch` without a `default:` case will silently ignore it — producing no error, no warning, and no indication that the new type was not handled. This creates bugs that are extremely difficult to find. Every `type switch` must include a `default:` case that either handles the unknown type explicitly or panics with a descriptive message explaining that an unexpected type was encountered. Silent omission is not acceptable.

---

## 5. Functions & Return Values

---

#### FUN-01 — Naked `return` statements are forbidden · **Error**

A naked (or "bare") `return` statement is a `return` with no explicit values, used inside a function with named return values. While Go's spec allows this, it makes the function deeply confusing to read: the reader must scroll back to the function signature to find the named return variables, mentally track every assignment to those variables, and reconstruct what values are actually being returned at each exit point. Every `return` statement must explicitly list the values being returned, even when named returns are declared.

---

#### FUN-02 — `error` must be the last return value · **Error**

This is one of Go's strongest idioms and a near-universal convention in the entire Go ecosystem. When a function can fail, the `error` value must be the rightmost return value in the signature. This allows the ubiquitous Go error-checking pattern — `result, err := doSomething(); if err != nil { ... }` — to read naturally, and ensures that tooling, linters, and other developers can reason about error flow consistently. Any function signature where `error` appears in any position other than last is a violation.

---

#### FUN-03 — Functions should accept interfaces and return concrete types · **Warning**

Accepting an interface as a parameter decouples the function from any specific implementation, making it testable, composable, and easier to evolve. Returning a concrete struct gives the caller full access to the type's fields and methods without requiring them to perform type assertions. The exception is `error`, which is always returned as an interface. A function that accepts a concrete struct where an interface would serve the same purpose is overly coupled; a function that returns an interface where a concrete struct is available forces unnecessary type assertions on the caller.

---

#### FUN-04 — Unused return values must be explicitly discarded with `_` · **Error**

In Go, silently ignoring return values — especially error values — is a common and dangerous mistake. If a function returns multiple values and the caller genuinely does not need one of them, that intent must be made explicit by assigning the unused value to the blank identifier `_`. This communicates clearly to the reader: "I know this function returns something here, and I have made a deliberate decision not to use it." Implicitly ignoring a return value by not capturing it at all is a violation.

---

## 6. Error Handling

---

#### ERR-01 — Error messages must not start with a capital letter · **Error**

Error messages in Go are typically composed by wrapping or joining multiple errors together with context: `fmt.Errorf("parsing config: %w", err)` becomes part of a chain that might read `"loading app: parsing config: open file: no such file or directory"`. If any message in this chain begins with a capital letter, it looks broken and ungrammatical when embedded mid-sentence. Error messages created with `errors.New` or `fmt.Errorf` must always begin with a lowercase letter.

---

#### ERR-02 — Error messages must not end with punctuation or a newline · **Error**

For the same reason as ERR-01, error messages are intended to be composed together. A message that ends with a period (`"failed to connect."`) or a newline (`"failed to connect\n"`) breaks the composed chain. The moulinette must scan all string literals passed to `errors.New` and `fmt.Errorf` and flag any that end with `.`, `!`, `?`, `:`, or `\n`.

---

#### ERR-03 — When returning a non-nil error, all other return values must be zero values · **Warning**

When a function fails, its non-error return values are meaningless. Returning a partially populated struct or a non-zero integer alongside a non-nil error forces the caller to decide which of the two return values to trust — an impossible and dangerous situation. The Go convention is unambiguous: on error, return the zero value for every non-error return. `return nil, 0, fmt.Errorf("something went wrong")` is correct; `return result, count, fmt.Errorf("something went wrong")` where `result` and `count` are non-zero is a violation.

---

#### ERR-04 — Specific error values must be checked with `errors.Is`, not `==` · **Error**

A direct equality check (`err == ErrNotFound`) breaks as soon as error wrapping is introduced. The `fmt.Errorf("context: %w", ErrNotFound)` pattern wraps the original error inside a new one, making the outer error unequal to `ErrNotFound` even though it contains it. `errors.Is` traverses the entire error chain and correctly identifies wrapped errors. Any equality comparison between an error value and a sentinel error is a violation; `errors.Is` is always the correct approach.

---

#### ERR-05 — Specific error types must be checked with `errors.As`, not type assertions · **Error**

Performing a type assertion on an error (`err.(*PathError)`) breaks for the same reason as `==` checks: it cannot see through wrapped errors. `errors.As` unwraps the error chain recursively until it finds an error whose concrete type matches the target, then populates the target pointer. This is the only correct way to extract a specific error type from a potentially wrapped error chain. Any type assertion performed directly on an error value is a violation.

---

#### ERR-06 — `panic` must only be used for unrecoverable programmer errors · **Error**

`panic` is not an error handling mechanism — it is a signal that the program has entered a state that the programmer asserts is fundamentally impossible. Appropriate uses are extremely narrow: detecting a nil pointer that should by construction never be nil, indexing a slice out of bounds, or catching a violated invariant that represents a bug in the program itself, not a bad input or a failed I/O operation. Network errors, file not found, invalid user input, timeout — none of these are reasons to `panic`. They must all be handled by returning an `error`.

---

#### ERR-07 — `recover` must always be called from within a `defer` · **Error**

Once a panic is triggered, the only code that continues to execute is code registered with `defer`. A call to `recover` placed in ordinary function flow — outside of a `defer` — will never execute during a panic and will always return `nil` during normal execution, making it completely useless. The only correct pattern is: `defer func() { if r := recover(); r != nil { /* handle */ } }()`. Any call to `recover` that is not inside a deferred anonymous function is a violation. Additionally, public API boundaries must use this pattern to prevent panics from propagating to callers.

---

#### ERR-08 — `panic`/`recover` must never be used as a general control flow mechanism · **Error**

Using `panic` to signal a failure condition and `recover` to catch it — essentially replicating exception-based error handling from other languages — is a fundamental violation of Go's design philosophy. This pattern is hidden, non-obvious, difficult to compose, and defeats Go's explicit error propagation model. Code that raises a `panic` expecting a caller to `recover` it as part of normal operation is a violation. All expected failure conditions must be communicated via returned `error` values.

---

## 7. Types, Interfaces & Pointers

---

#### TYP-01 — Pointers must only be used to indicate mutability · **Warning**

In Go, passing a value copies it. Passing a pointer shares the original. Pointers should be used precisely and only when the function needs to modify the value in a way that is visible to the caller, or when the value is too large to copy efficiently. Using pointers for small structs or primitives "just in case" adds unnecessary indirection, increases garbage collector pressure, and signals mutability where none exists. If a function accepts a `*Config` but only reads from it, it should accept a `Config` by value instead.

---

#### TYP-02 — Functions must not accept a pointer to populate a struct · **Warning**

A common anti-pattern imported from C is: `func NewClient(c *Client) error { c.host = "..."; return nil }`. This forces the caller to allocate the struct first and then pass a pointer in, splitting construction across two steps and obscuring ownership. The idiomatic Go approach is: `func NewClient() (Client, error) { return Client{host: "..."}, nil }`. The function constructs and returns the value directly, making ownership clear and eliminating an unnecessary level of indirection.

---

#### TYP-03 — Map reads must use the comma-ok idiom · **Error**

Reading a value from a Go map with a single-value assignment (`v := m[key]`) returns the zero value of the value type when the key is absent — identical to the result of a key that exists but holds the zero value. This makes it impossible to distinguish "key not present" from "key present with zero value." Every map read that may need to distinguish these two cases must use the two-value form: `v, ok := m[key]`. The `ok` boolean must then be checked. The moulinette must flag any single-value map read where the zero value of the type could be a meaningful result.

---

#### TYP-04 — Channel reads must use the comma-ok idiom on closeable channels · **Error**

Reading from a closed channel in Go does not block and does not return an error — it returns the zero value of the channel's element type, just as if a sender had sent that value. This makes it impossible to distinguish a real value from the artifact of a closed channel without the two-value receive form: `v, ok := <-ch`. If `ok` is `false`, the channel is closed and the value must be discarded. Any channel receive that could be reading from a channel that might be closed must use this form.

---

#### TYP-05 — Type assertions must use the comma-ok idiom · **Error**

A single-value type assertion (`v := x.(ConcreteType)`) panics at runtime if the dynamic type of `x` is not `ConcreteType`. Unless the code has already proven the type through a `type switch`, this is a latent runtime crash waiting to happen. All type assertions must use the two-value form: `v, ok := x.(ConcreteType)`, followed by a check of `ok` before using `v`. The only exception is inside the body of a `type switch` case, where the type is already guaranteed by the language.

---

#### TYP-06 — `interface{}` must never appear in new code; use `any` · **Error**

Since Go 1.18, `any` is an official alias for `interface{}`. They are identical at the type level, but `any` is dramatically more readable and expresses the concept more cleanly. All new Go code must use `any`. Any occurrence of the literal string `interface{}` in a type position is a violation.

---

#### TYP-07 — Use of `any` must be minimized and well-justified · **Warning**

`any` carries no type information. A function that accepts or returns `any` surrenders all compile-time type safety, forces the caller to perform type assertions, and makes the code significantly harder to reason about, test, and refactor. Its use is only acceptable in situations where the type is genuinely unknowable at compile time — such as when decoding arbitrary JSON, implementing a generic data structure, or bridging between reflection-heavy code. Every use of `any` in a function signature or struct field should be accompanied by a comment explaining why a more specific type could not be used.

---

## 8. Methods & Structs

---

#### STR-01 — Method receiver names must be short abbreviations of the type name · **Error**

When defining a method on a type, the receiver is the first parameter and represents the instance the method operates on. By Go convention, the receiver name must be a short, consistent abbreviation of the type's name — typically its first letter or its initials. For a type named `Client`, the receiver should be `c`. For a type named `RequestHandler`, the receiver should be `rh`. The same abbreviation must be used consistently across all methods of a given type — mixing `c` and `cl` and `client` across methods of the same type is a violation.

---

#### STR-02 — Receiver names `this` and `self` are forbidden · **Error**

The names `this` and `self` are borrowed from object-oriented languages like Java, Python, and C++ where they are keywords or strong conventions. In Go, they carry no special meaning and actively mislead readers familiar with idiomatic Go. Their use signals that the author is thinking about the code in OOP terms rather than Go terms. The receiver is just another parameter — name it like one, using the type's abbreviation as described in STR-01.

---

#### STR-03 — Getter and setter methods must not be written without interface justification · **Warning**

Go is not Java. Encapsulating every field behind a `GetFoo()` and `SetFoo()` method is not idiomatic and adds unnecessary boilerplate that provides no safety benefit. If a field is meant to be read by consumers of a type, export it directly. If it is meant to be internal, unexport it. Methods should represent behaviour and business logic, not field proxies. The only acceptable reason to write a getter or setter is to satisfy an interface that requires those method signatures.

---

#### STR-04 — Struct literals must always use named fields · **Error**

A positional struct literal (`Point{10, 20}`) is valid Go but is deeply fragile: if anyone adds, removes, or reorders fields in the struct definition, every positional literal silently breaks — sometimes producing wrong behaviour rather than a compile error. Named field literals (`Point{X: 10, Y: 20}`) are immune to this: adding a new field simply means that field gets its zero value in all existing literals. The moulinette must flag every struct literal that does not use field names for all fields. The only exception is structs from external packages explicitly designed to be used positionally, and even then a comment justifying the exception is required.

---

## 9. Packages & Documentation

---

#### DOC-01 — Doc comments must be directly above the declaration with no blank line · **Error**

A Go doc comment is defined as a comment that appears *immediately* above a declaration with no blank lines between the comment and the declaration. A blank line separating the comment from the declaration causes Go's tooling — `go doc`, `godoc`, and `pkg.go.dev` — to treat the comment as a standalone comment rather than a documentation comment for the symbol. The moulinette must check all exported symbols and flag any whose preceding comment is separated from the declaration by one or more blank lines.

---

#### DOC-02 — Doc comments must use `//` line comments, not `/* */` block comments · **Error**

While Go supports both `//` and `/* */` comments syntactically, the entire Go ecosystem — including the standard library, all official tooling, and every major Go project — uses `//` line comments for documentation. Block comments are reserved for special use cases like package doc files or temporarily disabling code. Any documentation comment written in `/* */` style is a violation.

---

#### DOC-03 — Doc comments must begin with the exact name of the symbol being documented · **Error**

By convention, every Go doc comment must open with the name of the thing it is documenting: `// Client is a...`, `// NewHandler creates a...`, `// MaxRetries defines...`. This convention enables `go doc` to display the comment in a predictable, parseable format, and it mirrors how the entire standard library is documented. A comment that begins with anything other than the symbol's name — including "This", "The", "A", or a lowercase version of the name — is a violation.

---

#### DOC-04 — All exported symbols must have a doc comment · **Error**

Every exported function, type, constant, variable, and method must have a documentation comment. Exported symbols are the public API of a package — they are the contract between the package and its consumers. Undocumented exported symbols force every caller to read the implementation to understand behaviour, preconditions, and return values. The moulinette must enumerate all exported identifiers and fail for any that lack a doc comment.

---

#### DOC-05 — `init()` functions are forbidden except for immutable setup · **Warning**

`init()` functions run automatically before `main()`, in an order that depends on import graph and is not always obvious to the reader. They make it difficult to trace how state is established, impossible to test the initialization logic in isolation, and dangerous when they have side effects. The only acceptable use of `init()` is initializing package-level variables that are effectively constants — values set once at startup and never modified again. Any `init()` function that performs I/O, modifies external state, launches goroutines, or alters mutable variables is a violation.

---

## 10. Slices & Collections

---

#### SLC-01 — Empty slices must be declared as `nil` slices, not empty slice literals · **Warning**

A `nil` slice (`var s []T`) and an empty slice (`s := []T{}`) behave identically for `append`, `len`, and `range`. However, they differ in one critical way: a `nil` slice serializes to `null` in JSON, while an empty slice serializes to `[]`. When the intention is to start with no elements and grow the slice via `append`, the `nil` form is correct and preferred. It avoids an unnecessary allocation and communicates that the slice starts empty by design. The empty slice literal form should only be used when the calling code explicitly requires a non-nil, zero-length slice — for example, when JSON serialization demands `[]` over `null`.

---

## 11. Concurrency

---

#### CON-01 — Public APIs must not expose channels or mutexes in exported types or functions · **Error**

Concurrency is an implementation detail. When a public API exposes a `chan T` or a `sync.Mutex` in its type signatures or exported struct fields, it forces concurrency concerns onto the caller and leaks internal synchronization strategy into the public contract. If the implementation later needs to change from channels to mutexes or vice versa, the public API must change with it — a breaking change. Concurrency primitives must be hidden behind the abstraction. Expose behaviour, not mechanism.

---

#### CON-02 — Every launched goroutine must have a guaranteed exit path · **Error**

A goroutine that has no guaranteed way to terminate is a goroutine leak. Leaking goroutines accumulate silently: they hold memory, block channels, retain references that prevent garbage collection, and consume scheduler resources. Every goroutine launched by the program must be able to respond to a termination signal. The standard mechanism is `context.Context`: the goroutine must select on `ctx.Done()` and exit when the context is cancelled. The moulinette must flag any `go` statement where the goroutine's function does not reference a `context.Context` and does not have another demonstrable exit path.

---

#### CON-03 — Channel closing must be owned by the writer; multi-writer closing requires `sync.WaitGroup` · **Error**

Closing a channel signals to all receivers that no more values will be sent. Only the goroutine that owns the write side of a channel — the sender — should ever close it. A receiver closing a channel it does not own is a race condition: if the sender then attempts to send on the now-closed channel, the program panics. When multiple goroutines write to the same channel, no single goroutine can safely close it on behalf of the others. In this case, a `sync.WaitGroup` must be used: each writer calls `wg.Done()` when it finishes, and a dedicated coordinator goroutine calls `close(ch)` only after `wg.Wait()` returns, guaranteeing all writers have finished before the channel is closed.

---

## 12. Custom Errors

---

#### CER-01 — Functions must return the `error` interface, even when using custom error types · **Error**

If a function's return type is declared as a concrete custom error struct (e.g., `*ValidationError`) rather than the standard `error` interface, callers are forced to depend on that specific type and perform explicit type handling. Worse, it prevents proper error wrapping and breaks compatibility with `errors.Is` and `errors.As`. All functions that can fail must declare `error` as their return type, regardless of what concrete type they instantiate internally.

---

#### CER-02 — Custom error variables must never be declared as their concrete type · **Error**

Declaring a local variable as a concrete custom error type — `var err *ValidationError` — creates a critical and subtle bug. In Go, an interface value is nil only when both its type and value are nil. If `err` is declared as `*ValidationError` and is never assigned, it holds a nil pointer of type `*ValidationError`. When returned as an `error` interface, the interface value is *not* nil — it has a non-nil type component. Callers checking `if err != nil` will incorrectly conclude that an error occurred even when none did. Always declare: `var err error`.

---

#### CER-03 — Custom error variables must never be returned in an unassigned state · **Error**

Closely related to CER-02: any code path that can return a custom error value without explicitly assigning it risks returning a non-nil interface wrapping a nil pointer. Every success path in a function that uses a custom error variable must explicitly `return nil` — not `return err` where `err` might be an unassigned concrete type. The moulinette must track whether custom error variables have been explicitly assigned before they are returned.

---

## 13. Structural Limits

---

#### LIM-01 — Functions must not exceed 50 lines · **Error**

A function that exceeds 50 lines is almost always doing too many things. Long functions are harder to test (because they have more execution paths), harder to read (because the reader must hold more context in their head simultaneously), and harder to reuse (because logic is tangled together rather than composed from small, well-named pieces). The 50-line limit includes all lines within the function body: declarations, logic, blank lines, and inline comments. The function signature and closing brace are excluded from the count. When a function approaches this limit, the correct response is to extract cohesive subsets of logic into named helper functions with clear, descriptive names.

---

#### LIM-02 — Functions must not have more than 5 parameters · **Error**

A function with many parameters is a sign that it has too many dependencies or is doing too many things. Beyond 5 parameters, the function signature becomes difficult to call correctly (argument order errors become likely), difficult to read at call sites (the argument list becomes a wall of values), and difficult to evolve (adding a new parameter is a breaking change for all callers). When more than 5 inputs are genuinely needed, they must be grouped into a configuration or options struct that is passed as a single parameter. This also makes call sites self-documenting via named fields: `NewServer(ServerConfig{Port: 8080, Timeout: 30})` instead of `NewServer(8080, 30, true, nil, "")`.

---

#### LIM-03 — Nesting depth must not exceed 4 levels · **Error**

Every additional level of nesting — an `if` inside a `for` inside an `if` inside a function — compounds cognitive complexity exponentially. Code with 5 or more levels of indentation forces the reader to maintain a deeply nested mental stack, making it extremely easy to misunderstand which conditions apply at any given point. The correct solution is the application of guard clauses: return or `continue` early when a precondition fails, rather than wrapping the success path in ever-deeper nesting. A function whose happy path is buried inside 4 levels of conditionals should be restructured so the happy path is the straightforward top-level flow, and failure conditions are handled and returned as early as possible.

---

#### LIM-04 — Source files must not exceed 500 lines · **Warning**

A file that grows beyond 500 lines is a strong signal that it is accumulating too many responsibilities. Go packages are directories, not files — there is no reason to put everything in one file. Code should be organized into focused files where each file covers a coherent aspect of the package's responsibility: `client.go` for the client type and its methods, `errors.go` for error types, `config.go` for configuration. Files exceeding 500 lines should be reviewed for opportunities to extract related functionality into dedicated, descriptively named files.

---

## 14. Context Handling

---

#### CTX-01 — `context.Context` must always be the first parameter · **Error**

When a function accepts a `context.Context`, it must be the very first parameter in the function signature, by universal Go convention. This makes context propagation immediately visible to any reader scanning a function signature. It also enables consistent tooling — many linters, code generators, and documentation tools assume this position. A `context.Context` appearing anywhere other than the first parameter is a violation regardless of how many other parameters the function has.

---

#### CTX-02 — `context.Context` must never be stored in a struct · **Error**

Storing a `context.Context` in a struct field is a widely recognized anti-pattern explicitly warned against in the Go documentation. A context carries a deadline, cancellation signal, and request-scoped values — it is inherently tied to a single logical operation, not to a long-lived object. When stored in a struct, the context's lifetime becomes decoupled from the operation it is meant to control, making cancellation and deadline propagation unreliable and invisible. Contexts must always be passed explicitly as function parameters to every function that needs them, threading through the call chain from the outermost handler to the innermost operation.

---

#### CTX-03 — `nil` must never be passed as a `context.Context` · **Error**

Passing `nil` as a `context.Context` is explicitly forbidden by the Go standard library documentation. A nil context will cause a panic in any standard library function that dereferences it (such as `http.NewRequestWithContext`), and it signals to the reader that the developer has not thought about cancellation or deadline propagation at all. If no context is available or appropriate at the top of the call chain, `context.Background()` must be used as the root context. If the context is a placeholder pending a proper implementation, `context.TODO()` must be used and must be accompanied by a comment explaining what is missing and when it will be addressed.

---

#### CTX-04 — Cancel functions from derived contexts must always be deferred or called on all exit paths · **Error**

Any context created via `context.WithCancel`, `context.WithTimeout`, or `context.WithDeadline` returns a cancellation function that must be invoked to release resources. Failing to call this function leaks timers, references, and cancellation propagation resources. The required pattern is to place `defer cancel()` immediately after successful context creation, before any possible early return. The moulinette must flag any derived-context cancel function that is never handled, handled only conditionally, or handled too late to cover all returns.

---

## 15. Import Organization

---

#### IMP-01 — Imports must be organized into groups separated by blank lines · **Error**

Go imports must be organized into a maximum of three groups, in this exact order, each separated from the next by a single blank line:

1. **Standard library** packages (e.g., `"fmt"`, `"net/http"`, `"os"`)
2. **Third-party** packages (e.g., `"github.com/some/library"`)
3. **Internal** packages from the same module (e.g., `"mymodule/internal/config"`)

Mixing packages from different groups without a blank line separator, or ordering the groups differently, is a violation. The `goimports` tool enforces this automatically and must be integrated into the moulinette's formatting pass.

---

#### IMP-02 — Unused imports are forbidden · **Error**

The Go compiler already enforces this at the language level — code with an unused import will not compile. However, the moulinette should surface this violation with a clear, actionable diagnostic message that identifies the specific unused import by name and line number, so that the developer can fix it before reaching the compilation stage. This is especially valuable in pre-compilation static analysis pipelines where fast feedback is critical.

---

#### IMP-03 — Import aliases must be used only when necessary and must be descriptive · **Warning**

Import aliases (`import foo "some/long/package/path"`) should be avoided when the package's own name is clear and unambiguous. They are acceptable when two imported packages share the same package name — in which case one must be aliased to avoid a compile-time conflict. When aliasing is necessary, the alias must be descriptive and must not be a single letter or a cryptic abbreviation. The blank import (`import _ "package"`) is permitted only when a package's `init()` side effects are intentionally required (e.g., registering a database driver), and it must always be accompanied by a comment that explains exactly what side effect is being triggered and why.

---

## 16. Resource Management

---

#### RES-01 — Resources that require closing must be closed with `defer` · **Error**

Any value that implements `io.Closer` — including files, HTTP response bodies, database connections, and network connections — must be closed after it is acquired. The closing must be performed using `defer` placed immediately after the successful acquisition of the resource, before any other logic that might cause an early return. Placing the `Close()` call at the end of the function body without `defer` is fragile: any early return, panic, or new code path introduced in the future could bypass the close, silently leaking the resource. The correct and mandatory pattern is always: `f, err := os.Open(...); if err != nil { return err }; defer f.Close()`.

---

#### RES-02 — `defer` must not be used inside a loop · **Warning**

`defer` schedules a call to be executed when the *surrounding function* returns — not when the current loop iteration ends. Using `defer` inside a loop means all deferred calls accumulate across every iteration and are only flushed when the entire function returns. On a loop that runs thousands of times, this means thousands of open file handles or database connections are held simultaneously until the function exits, which is almost certainly a resource leak. Resources acquired inside a loop must be closed within the same iteration using an explicit call. If the logic is complex enough to warrant `defer`, the loop body must be extracted into its own helper function where `defer` operates correctly and closes the resource at the end of each call.

---

## 17. Concurrency Safety

---

#### SAF-01 — Structs containing a `sync.Mutex` must only use pointer receivers · **Error**

A `sync.Mutex` must never be copied after first use. If a struct containing a `sync.Mutex` is passed by value, Go copies the entire struct — including the mutex state — producing two completely independent mutex instances instead of one shared one. Any method defined on such a struct with a value receiver will operate on a copy of the struct and therefore a copy of the mutex, silently breaking the synchronization guarantee entirely. Every method on any struct that contains a `sync.Mutex`, `sync.RWMutex`, or any other synchronization primitive from the `sync` package must use a pointer receiver to ensure all methods operate on the same shared instance.

---

#### SAF-02 — `sync.WaitGroup` must never be copied after first use · **Error**

Like `sync.Mutex`, `sync.WaitGroup` carries internal state that must not be copied once the WaitGroup has been used. Passing a `WaitGroup` by value to a goroutine is a race condition: the goroutine holds an independent copy of the counter, and calling `Done()` on that copy does not affect the original WaitGroup that `Wait()` is blocking on. This causes `Wait()` to either hang indefinitely or return prematurely, depending on timing. A `sync.WaitGroup` must always be passed to goroutines by pointer: `go worker(&wg)`. The moulinette must flag any goroutine launch that passes a `sync.WaitGroup` by value.

---

## 18. Magic Values

---

#### MAG-01 — Numeric literals used more than once must be extracted into named constants · **Error**

A bare number in the middle of logic — a "magic number" — tells the reader nothing about its meaning or intent. `time.Sleep(30 * time.Second)` in one place and `deadline := time.Now().Add(30 * time.Second)` in another creates an invisible, untracked dependency: if the timeout needs to change, there is no way to find all of its occurrences without searching the entire codebase for the literal `30`, risking a missed instance. Any numeric literal that appears in two or more locations, or whose meaning is not immediately and completely self-evident from its surrounding context alone, must be extracted into a named constant declared at the tightest appropriate scope with a name that explains both the value and its purpose.

---

#### MAG-02 — String literals used as identifiers or keys must be named constants · **Error**

Repeating a string literal in multiple places — particularly strings used as map keys, HTTP header names, environment variable names, SQL column names, or configuration field names — creates the same fragility as magic numbers, with the added danger of typo-induced silent bugs. A misspelled map key does not produce a compile error; it simply returns a zero value at runtime, potentially far from the source of the mistake. Any string that serves as an identifier, a key, or a protocol-level token and appears in two or more locations across the package must be declared as a named constant. The moulinette must flag any string literal that appears two or more times in the same package.

---

## 19. Testing

---

#### TST-01 — Multiple test cases must use the table-driven pattern · **Error**

When testing a function across multiple input and output combinations, each case must not have its own isolated `Test_` function. Instead, all cases must be expressed as a slice of anonymous structs — the table — where each entry defines the test name, inputs, and expected outputs. The test function then iterates over the table with `for _, tc := range tests` and calls `t.Run(tc.name, func(t *testing.T) { ... })` for each case. This pattern eliminates code duplication, produces clear and individually named test output, makes it trivial to add new cases without touching existing logic, and allows `go test -run TestFoo/case_name` to target individual sub-cases. Any test file that contains three or more `Test_` functions testing variations of the same function under different inputs without using the table-driven pattern is a violation.

---

#### TST-02 — Test helper functions must call `t.Helper()` as their first statement · **Error**

When a test helper function is called from a test and triggers a failure via `t.Errorf` or `t.Fatalf`, Go's testing framework by default reports the failure at the line inside the helper — which is nearly useless for the developer who needs to know which invocation of the helper failed. Calling `t.Helper()` at the very start of the helper function instructs the framework to attribute any test failures to the call site of the helper rather than to the helper's internal line number. This makes test output immediately actionable. Every function in a `_test.go` file that calls any of `t.Error`, `t.Errorf`, `t.Fatal`, or `t.Fatalf` and is itself called by other test functions must call `t.Helper()` as its absolute first statement, before any other logic.

---

#### TST-03 — Tests must not use `time.Sleep` for synchronization · **Error**

Using `time.Sleep` in a test to wait for a goroutine to complete, a channel to receive a value, or an asynchronous event to occur is a fundamentally broken synchronization strategy. On a fast machine the sleep duration is unnecessarily long, slowing the entire test suite; on a slow machine, a loaded CI server, or under any unusual scheduling pressure, the sleep may be too short, causing the test to fail intermittently and unpredictably — a "flaky test" that erodes confidence in the entire suite. Synchronization in tests must use the same primitives used in production code: channels to signal completion, `sync.WaitGroup` to await multiple goroutines, or `context.Context` with a deadline to enforce a timeout. If a test must poll for an asynchronous condition, it must do so with an explicit timeout loop using `time.After` or a dedicated testing assertion library, never with a bare `time.Sleep`.

---

## Severity Reference

| Level | Meaning |
|-------|---------|
| **Error** | Violation fails the moulinette immediately; the code must be corrected before it can be submitted or evaluated |
| **Warning** | Violation is surfaced as a diagnostic and logged, but does not block submission; it is strongly recommended to address all warnings |