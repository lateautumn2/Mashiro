package controllers

import "testing"

func TestBuildLatencyTaskModelValidatesTargets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		input     LatencyTaskInput
		wantError bool
	}{
		{
			name: "accepts tcp host port",
			input: LatencyTaskInput{
				Name:   "tcp",
				Type:   "tcp",
				Target: "example.com:443",
			},
		},
		{
			name: "rejects tcp without port",
			input: LatencyTaskInput{
				Name:   "tcp",
				Type:   "tcp",
				Target: "157.148.58.29",
			},
			wantError: true,
		},
		{
			name: "accepts icmp host",
			input: LatencyTaskInput{
				Name:   "icmp",
				Type:   "icmp",
				Target: "8.8.8.8",
			},
		},
		{
			name: "rejects invalid http url",
			input: LatencyTaskInput{
				Name:   "http",
				Type:   "http",
				Target: "example.com/health",
			},
			wantError: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := buildLatencyTaskModel(tc.input)
			if tc.wantError && err == nil {
				t.Fatalf("expected validation error")
			}
			if !tc.wantError && err != nil {
				t.Fatalf("expected success, got error: %v", err)
			}
		})
	}
}
