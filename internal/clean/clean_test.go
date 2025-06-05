package clean

import (
	"testing"
	"time"

	ec2 "github.com/bananaops/cloudoff/internal/aws"
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

func TestDurationExceeded(t *testing.T) {

	tests := []struct {
		name     string
		instance ec2.Instance
		expected bool
	}{
		{
			name: "TTL tag with infinity value",
			instance: ec2.Instance{
				Tags: []ec2.Tag{
					{Key: "cloudoff:ttl", Value: "infinity"},
				},
			},
			expected: false,
		},
		{
			name: "TTL tag with valid duration",
			instance: ec2.Instance{
				Tags: []ec2.Tag{
					{Key: "cloudoff:ttl", Value: "2h"},
				},
				LaunchTime: time.Now().Add(-3 * time.Hour),
			},
			expected: true,
		},
		{
			name: "TTL tag with invalid duration",
			instance: ec2.Instance{
				Tags: []ec2.Tag{
					{Key: "cloudoff:ttl", Value: "invalid"},
				},
			},
			expected: false,
		},
		{
			name: "Infinity tag",
			instance: ec2.Instance{
				Tags: []ec2.Tag{
					{Key: "cloudoff:ttl", Value: "infinity"},
				},
			},
			expected: false,
		},
		{
			name: "No TTL tag",
			instance: ec2.Instance{
				Tags: []ec2.Tag{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DurationExceeded(tt.instance)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
