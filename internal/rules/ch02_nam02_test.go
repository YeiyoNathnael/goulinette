package rules

import "testing"

// TestIsAllCapsStyle documents this exported function.
func TestIsAllCapsStyle(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "snake all caps", value: "MAX_RETRIES", want: true},
		{name: "acronym all caps", value: "URL", want: true},
		{name: "camel case", value: "maxRetries", want: false},
		{name: "pascal case", value: "MaxRetries", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Helper()
			got := isAllCapsStyle(tc.value)
			if got != tc.want {
				t.Fatalf("isAllCapsStyle(%q) = %v, want %v", tc.value, got, tc.want)
			}
		})
	}
}
