package tools

import (
	"testing"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

func TestParseOutputDiagnostics_WithPosition(t *testing.T) {
	input := "a/b/c.go:12:8: something bad"
	ds := ParseOutputDiagnostics(input, "FMT-02", "go vet", diag.SeverityError)
	if len(ds) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(ds))
	}
	if ds[0].Pos.File != "a/b/c.go" || ds[0].Pos.Line != 12 || ds[0].Pos.Col != 8 {
		t.Fatalf("unexpected position: %+v", ds[0].Pos)
	}
}

func TestParseOutputDiagnostics_WithoutPosition(t *testing.T) {
	input := "plain message"
	ds := ParseOutputDiagnostics(input, "FMT-03", "staticcheck", diag.SeverityWarning)
	if len(ds) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(ds))
	}
	if ds[0].Message != "plain message" {
		t.Fatalf("unexpected message: %q", ds[0].Message)
	}
}
