package suppress

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/YeiyoNathnael/goulinette/internal/diag"
)

const (
	suppressTestLine = 1
	wantNone         = 0
	wantOne          = 1

	ruleNAM01 = "NAM-01"
	ruleVAR01 = "VAR-01"
	ruleERR01 = "ERR-01"
)

// TestFilter covers the core filtering behaviour: same-line directives,
// preceding-line directives, bare (all-rule) directives, multiple-rule
// directives, and the pass-through case when no directive is present.
func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		findings []diag.Finding
		want     int
	}{
		{
			name: "matching rule on same line",
			src:  "x := nil //goulinette:ignore NAM-01\n",
			findings: []diag.Finding{
				{RuleID: ruleNAM01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
			},
			want: wantNone,
		},
		{
			name: "matching rule on preceding line",
			src:  "//goulinette:ignore NAM-01\nx := nil\n",
			findings: []diag.Finding{
				{RuleID: ruleNAM01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine + 1}},
			},
			want: wantNone,
		},
		{
			name: "bare directive suppresses all rules",
			src:  "x := nil //goulinette:ignore\n",
			findings: []diag.Finding{
				{RuleID: ruleNAM01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
				{RuleID: ruleVAR01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
			},
			want: wantNone,
		},
		{
			name: "different rule is not suppressed",
			src:  "x := nil //goulinette:ignore NAM-01\n",
			findings: []diag.Finding{
				{RuleID: ruleVAR01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
			},
			want: wantOne,
		},
		{
			name: "no directive passes through unchanged",
			src:  "x := nil\n",
			findings: []diag.Finding{
				{RuleID: ruleNAM01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
			},
			want: wantOne,
		},
		{
			name: "multiple rules on one directive",
			src:  "x := nil //goulinette:ignore NAM-01 VAR-01\n",
			findings: []diag.Finding{
				{RuleID: ruleNAM01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
				{RuleID: ruleVAR01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
				{RuleID: ruleERR01, Severity: diag.SeverityError, Pos: diag.Position{Line: suppressTestLine}},
			},
			want: wantOne,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			file := writeTempFile(t, tc.src)
			for i := range tc.findings {
				tc.findings[i].Pos.File = file
			}
			got := Filter(tc.findings)
			if len(got) != tc.want {
				t.Fatalf("Filter: got %d findings, want %d", len(got), tc.want)
			}
		})
	}
}

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sample.go")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}
