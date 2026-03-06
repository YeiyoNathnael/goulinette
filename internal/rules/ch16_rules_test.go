package rules

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	resTestFilePerms = 0o644
	resWriteErrFmt   = "write %s: %v"
	resGoModFile     = "go.mod"
	resSampleGoFile  = "sample.go"
)

func writeRESFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), resTestFilePerms); err != nil {
		t.Fatalf(resWriteErrFmt, name, err)
	}
	return path
}

// TestRES01DeferClosePatterns verifies that RES-01 flags file opens,
// network dials, and other resource acquisitions that are not followed
// by a matching defer to close or release the resource.
func TestRES01DeferClosePatterns(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "missing defer close fails",
			source: `package sample
import "os"
func f(path string) error {
	f, err := os.Open(path)
	if err != nil { return err }
	_ = f
	return nil
}
`,
			wantCount: 1,
		},
		{
			name: "defer before error check fails",
			source: `package sample
import "os"
func f(path string) error {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil { return err }
	return nil
}
`,
			wantCount: 1,
		},
		{
			name: "proper open pattern passes",
			source: `package sample
import "os"
func f(path string) error {
	f, err := os.Open(path)
	if err != nil { return err }
	defer f.Close()
	return nil
}
`,
			wantCount: 0,
		},
		{
			name: "http response body close detected",
			source: `package sample
import (
	"net/http"
	"context"
)
func f(ctx context.Context, url string) error {
	resp, err := http.DefaultClient.Do((&http.Request{}).WithContext(ctx))
	if err != nil { return err }
	defer resp.Body.Close()
	return nil
}
`,
			wantCount: 0,
		},
		{
			name: "blank assignment exempted",
			source: `package sample
import "os"
func f(path string) error {
	_, err := os.Open(path)
	if err != nil { return err }
	return nil
}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			_ = writeRESFile(t, dir, resGoModFile, "module example.com/res01\n\ngo 1.22\n")
			_ = writeRESFile(t, dir, resSampleGoFile, tc.source)

			diags, err := NewRES01().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run RES-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d RES-01 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

// TestRES02DeferInLoop verifies that RES-02 flags defer statements that
// appear directly inside a for/range loop body, where the deferred call
// will not execute until the enclosing function returns.
func TestRES02DeferInLoop(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "direct defer in loop warns",
			source: `package sample
func f(items []int) {
	for range items {
		defer func() {}()
	}
}
`,
			wantCount: 1,
		},
		{
			name: "defer inside func literal in loop is allowed",
			source: `package sample
func f(items []int) {
	for range items {
		func() {
			defer func() {}()
		}()
	}
}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			dir := t.TempDir()
			file := writeRESFile(t, dir, resSampleGoFile, tc.source)

			diags, err := NewRES02().Run(Context{Files: []string{file}})
			if err != nil {
				t.Fatalf("run RES-02: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d RES-02 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
