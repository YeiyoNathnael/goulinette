package rules

import (
	"os"
	"path/filepath"
	"testing"

	"goulinette/internal/diag"
)

func writeModuleFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestCON01_ExportedAPIExposure(t *testing.T) {
	dir := t.TempDir()
	writeModuleFile(t, dir, "go.mod", "module example.com/con01\n\ngo 1.22\n")
	writeModuleFile(t, dir, "sample.go", `package sample

import "sync"

type inner struct {
	mu sync.Mutex
}

type Public struct {
	inner
}

func Exported(ch chan int) Public {
	return Public{}
}

func ReturnChan() chan int {
	return nil
}
`)

	diags, err := NewCON01().Run(Context{Root: dir})
	if err != nil {
		t.Fatalf("run CON-01: %v", err)
	}
	if len(diags) < 3 {
		t.Fatalf("expected at least 3 CON-01 diagnostics, got %d", len(diags))
	}
}

func TestCON02_HeuristicCancellationSignals(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "no context signal warns",
			source: `package sample
func f() {
	go func() {
		for {}
	}()
}
`,
			wantCount: 1,
		},
		{
			name: "context parameter call is accepted",
			source: `package sample
import "context"
func worker(ctx context.Context) {}
func f(ctx context.Context) {
	go worker(ctx)
}
`,
			wantCount: 0,
		},
		{
			name: "ctx done select is accepted",
			source: `package sample
import "context"
func f(ctx context.Context) {
	go func() {
		select {
		case <-ctx.Done():
			return
		}
	}()
}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeModuleFile(t, dir, "go.mod", "module example.com/con02\n\ngo 1.22\n")
			writeModuleFile(t, dir, "sample.go", tc.source)

			diags, err := NewCON02().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run CON-02: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d CON-02 diagnostics, got %d", tc.wantCount, len(diags))
			}
			if tc.wantCount > 0 && diags[0].Severity != diag.SeverityError {
				t.Fatalf("expected CON-02 severity error, got %s", diags[0].Severity)
			}
		})
	}
}

func TestCON03_ConservativeOwnershipWarnings(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantAtLeast int
	}{
		{
			name: "close from different goroutine warns",
			source: `package sample
func f(ch chan int) {
	go func() {
		ch <- 1
	}()
	close(ch)
}
`,
			wantAtLeast: 1,
		},
		{
			name: "multi writer without waitgroup warns",
			source: `package sample
func f(ch chan int) {
	go func() { ch <- 1 }()
	go func() { ch <- 2 }()
	close(ch)
}
`,
			wantAtLeast: 1,
		},
		{
			name: "single context write and close can pass",
			source: `package sample
func f(ch chan int) {
	ch <- 1
	close(ch)
}
`,
			wantAtLeast: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeModuleFile(t, dir, "go.mod", "module example.com/con03\n\ngo 1.22\n")
			writeModuleFile(t, dir, "sample.go", tc.source)

			diags, err := NewCON03().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run CON-03: %v", err)
			}
			if len(diags) < tc.wantAtLeast {
				t.Fatalf("expected at least %d CON-03 diagnostics, got %d", tc.wantAtLeast, len(diags))
			}
			if tc.wantAtLeast > 0 && len(diags) > 0 && diags[0].Severity != diag.SeverityError {
				t.Fatalf("expected CON-03 severity error, got %s", diags[0].Severity)
			}
		})
	}
}
