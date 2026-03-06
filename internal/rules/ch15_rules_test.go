package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	impTestFilePerms = 0o644
	impWriteErrFmt   = "write %s: %v"
	impGoModFile     = "go.mod"
	impSampleGoFile  = "sample.go"
)

func writeIMPFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), impTestFilePerms); err != nil {
		t.Fatalf(impWriteErrFmt, name, err)
	}
	return path
}

// TestIMP01Grouping verifies that IMP-01 requires import declarations to
// follow the stdlib / external / internal three-group convention and flags
// any deviation from that ordering.
func TestIMP01Grouping(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "mixed classes in one group fails",
			source: `package sample
import (
	"fmt"
	"github.com/pkg/errors"
)
func f() { fmt.Println(errors.New("x")) }
`,
			wantCount: 1,
		},
		{
			name: "properly grouped passes",
			source: `package sample
import (
	"fmt"

	"github.com/pkg/errors"

	"example.com/mod/internal/x"
)
func f() { fmt.Println(errors.New(x.Name)) }
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			_ = writeIMPFile(t, dir, impGoModFile, "module example.com/mod\n\ngo 1.22\n")
			file := writeIMPFile(t, dir, impSampleGoFile, tc.source)

			diags, err := NewIMP01().Run(Context{Root: dir, Files: []string{file}})
			if err != nil {
				t.Fatalf("run IMP-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d IMP-01 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

// TestIMP02UnusedImports verifies that IMP-02 reports import paths that
// are listed in the import block but never referenced in the file body.
func TestIMP02UnusedImports(t *testing.T) {
	dir := t.TempDir()
	file := writeIMPFile(t, dir, impSampleGoFile, `package sample
import (
	"fmt"
	"strings"
)
func f() { fmt.Println("ok") }
`)

	diags, err := NewIMP02().Run(Context{Files: []string{file}})
	if err != nil {
		t.Fatalf("run IMP-02: %v", err)
	}
	if len(diags) != 1 {
		t.Fatalf("expected 1 IMP-02 diagnostic, got %d", len(diags))
	}
}

// TestIMP03AliasPolicy verifies that IMP-03 flags import aliases that
// shadow the last path segment, rename stdlib packages to single-letter
// names, or otherwise violate the alias-naming policy.
func TestIMP03AliasPolicy(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "single-letter alias warns",
			source: `package sample
import s "strings"
func f() string { return s.ToUpper("x") }
`,
			wantCount: 2,
		},
		{
			name: "blank import without comment warns",
			source: `package sample
import _ "github.com/lib/pq"
func f() {}
`,
			wantCount: 1,
		},
		{
			name: "blank import with comment passes",
			source: `package sample
import _ "github.com/lib/pq" // register postgres driver
func f() {}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeIMPFile(t, dir, impSampleGoFile, tc.source)

			diags, err := NewIMP03().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run IMP-03: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d IMP-03 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
