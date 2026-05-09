package collector

import "testing"

func TestValidateTCPTarget(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		target    string
		wantError bool
	}{
		{name: "accepts host port", target: "example.com:443"},
		{name: "accepts ipv6 host port", target: "[2001:4860:4860::8888]:443"},
		{name: "rejects missing port", target: "157.148.58.29", wantError: true},
		{name: "rejects invalid port", target: "example.com:70000", wantError: true},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := validateTCPTarget(tc.target)
			if tc.wantError && err == nil {
				t.Fatalf("expected validation error")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})
	}
}
