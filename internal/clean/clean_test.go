package clean

import (
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"2w", 336 * time.Hour, false},  // 2 weeks = 336 hours
		{"3d", 72 * time.Hour, false},   // 3 days = 72 hours
		{"15d", 360 * time.Hour, false}, // 3 days = 72 hours
		{"5h", 5 * time.Hour, false},    // 5 hours
		{"10x", 0, true},                // Invalid unit
		{"7", 0, true},                  // Missing unit
		{"", 0, true},                   // Empty input
		{"-3d", 0, true},                // Negative duration
		{"0w", 0, false},                // Zero weeks
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := parseDuration(test.input)
			if test.hasError {
				if err == nil {
					t.Errorf("expected an error for input '%s', but got none", test.input)
				}
			} else {
				if err != nil {
					t.Errorf("did not expect an error for input '%s', but got: %v", test.input, err)
				}
				if result != test.expected {
					t.Errorf("for input '%s', expected %v, but got %v", test.input, test.expected, result)
				}
			}
		})
	}
}
