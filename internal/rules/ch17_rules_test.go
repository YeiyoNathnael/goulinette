package rules

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSAFFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func TestSAF01_PointerReceiverForMutexStructs(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "value receiver with mutex fails",
			source: `package sample
import "sync"
type Counter struct { mu sync.Mutex; n int }
func (c Counter) Inc() { c.mu.Lock(); c.n++; c.mu.Unlock() }
`,
			wantCount: 1,
		},
		{
			name: "pointer receiver with mutex passes",
			source: `package sample
import "sync"
type Counter struct { mu sync.Mutex; n int }
func (c *Counter) Inc() { c.mu.Lock(); c.n++; c.mu.Unlock() }
`,
			wantCount: 0,
		},
		{
			name: "nested rwmutex in embedded struct fails",
			source: `package sample
import "sync"
type inner struct { rw sync.RWMutex }
type S struct { inner }
func (s S) Read() { s.rw.RLock(); s.rw.RUnlock() }
`,
			wantCount: 1,
		},
		{
			name: "value receiver with sync.Once fails",
			source: `package sample
import "sync"
type Runner struct { once sync.Once }
func (r Runner) Run(f func()) { r.once.Do(f) }
`,
			wantCount: 1,
		},
		{
			name: "embedded noCopy sentinel with value receiver fails",
			source: `package sample
type noCopy struct{}
func (*noCopy) Lock() {}
func (*noCopy) Unlock() {}

type Buffer struct {
	noCopy
	data []byte
}

func (b Buffer) Len() int { return len(b.data) }
`,
			wantCount: 1,
		},
		{
			name: "pointer to mutex field is allowed",
			source: `package sample
import "sync"
type S struct { mu *sync.Mutex }
func (s S) Touch() {}
`,
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeSAFFile(t, dir, "go.mod", "module example.com/saf01\n\ngo 1.22\n")
			writeSAFFile(t, dir, "sample.go", tc.source)

			diags, err := NewSAF01().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run SAF-01: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d SAF-01 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}

func TestSAF02_WaitGroupCopyPatterns(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantCount int
	}{
		{
			name: "goroutine passing waitgroup by value fails",
			source: `package sample
import "sync"
func worker(wg sync.WaitGroup) { wg.Done() }
func f() {
	var wg sync.WaitGroup
	wg.Add(1)
	go worker(wg)
	wg.Wait()
}
`,
			wantCount: 1,
		},
		{
			name: "goroutine passing sync.Once by value fails",
			source: `package sample
import "sync"
func worker(o sync.Once) {}
func f() {
	var once sync.Once
	go worker(once)
}
`,
			wantCount: 1,
		},
		{
			name: "goroutine passing pointer waitgroup passes",
			source: `package sample
import "sync"
func worker(wg *sync.WaitGroup) { wg.Done() }
func f() {
	var wg sync.WaitGroup
	wg.Add(1)
	go worker(&wg)
	wg.Wait()
}
`,
			wantCount: 0,
		},
		{
			name: "assignment copy fails",
			source: `package sample
import "sync"
func f() {
	var a sync.WaitGroup
	var b sync.WaitGroup
	b = a
	_ = b
}
`,
			wantCount: 1,
		},
		{
			name: "returning waitgroup by value fails",
			source: `package sample
import "sync"
func f() sync.WaitGroup {
	var wg sync.WaitGroup
	return wg
}
`,
			wantCount: 1,
		},
		{
			name: "struct containing waitgroup by value to goroutine fails",
			source: `package sample
import "sync"
type Job struct { wg sync.WaitGroup }
func worker(j Job) { j.wg.Done() }
func f() {
	var j Job
	j.wg.Add(1)
	go worker(j)
	j.wg.Wait()
}
`,
			wantCount: 1,
		},
		{
			name: "sync.Map assignment copy fails",
			source: `package sample
import "sync"
func f() {
	var a sync.Map
	var b sync.Map
	b = a
	_ = b
}
`,
			wantCount: 1,
		},
		{
			name: "embedded noCopy assignment copy fails",
			source: `package sample
type noCopy struct{}
func (*noCopy) Lock() {}
func (*noCopy) Unlock() {}

type Item struct {
	noCopy
	v int
}

func f() {
	var a Item
	var b Item
	b = a
	_ = b
}
`,
			wantCount: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			writeSAFFile(t, dir, "go.mod", "module example.com/saf02\n\ngo 1.22\n")
			writeSAFFile(t, dir, "sample.go", tc.source)

			diags, err := NewSAF02().Run(Context{Root: dir})
			if err != nil {
				t.Fatalf("run SAF-02: %v", err)
			}
			if len(diags) != tc.wantCount {
				t.Fatalf("expected %d SAF-02 diagnostics, got %d", tc.wantCount, len(diags))
			}
		})
	}
}
