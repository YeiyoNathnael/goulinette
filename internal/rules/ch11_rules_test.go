package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	testFilePerms      = 0o644
	writeFileFatalfFmt = "write %s: %v"
	goModFileName      = "go.mod"
	sampleGoFileName   = "sample.go"
	minCON01Findings   = 3
)

func writeModuleFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), testFilePerms); err != nil {
		t.Fatalf(writeFileFatalfFmt, name, err)
	}
	return path
}

// TestCON01ExportedAPIExposure verifies that CON-01 flags exported functions
// that accept or return bare channel types, which leaks concurrency
// implementation details through the public API.
func TestCON01ExportedAPIExposure(t *testing.T) {
	dir := t.TempDir()
	_ = writeModuleFile(t, dir, goModFileName, "module example.com/con01\n\ngo 1.22\n")
	_ = writeModuleFile(t, dir, sampleGoFileName, `package sample

import "sync"

type inner struct {
	mu sync.Mutex
}

// Public documents this exported type.
type Public struct {
	inner
}

// Exported documents this exported function.
func Exported(ch chan int) Public {
	return Public{}
}

// ReturnChan documents this exported function.
func ReturnChan() chan int {
	return nil
}
`)

	diags, err := NewCON01().Run(Context{Root: dir})
	if err != nil {
		t.Fatalf("run CON-01: %v", err)
	}
	if len(diags) < minCON01Findings {
		t.Fatalf("expected at least 3 CON-01 diagnostics, got %d", len(diags))
	}
}

// TestCON02HeuristicCancellationSignals verifies that CON-02 detects
// goroutine launch sites that lack a context or done-channel cancellation
// mechanism, and does not fire when a proper signal is present.
func TestCON02HeuristicCancellationSignals(t *testing.T) {
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
			t.Helper()
			dir := t.TempDir()
			_ = writeModuleFile(t, dir, goModFileName, "module example.com/con02\n\ngo 1.22\n")
			_ = writeModuleFile(t, dir, sampleGoFileName, tc.source)

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

// TestCON03ConservativeOwnershipWarnings verifies that CON-03 warns when a
// channel is closed or written by more than one goroutine without a
// sync.WaitGroup coordinating ownership.
func TestCON03ConservativeOwnershipWarnings(t *testing.T) {
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
			t.Helper()
			dir := t.TempDir()
			_ = writeModuleFile(t, dir, goModFileName, "module example.com/con03\n\ngo 1.22\n")
			_ = writeModuleFile(t, dir, sampleGoFileName, tc.source)

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
