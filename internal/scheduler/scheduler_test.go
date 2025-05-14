package scheduler

import (
	"reflect"
	"testing"
)

func TestParseSchedule(t *testing.T) {
	tests := []struct {
		input    string
		expected Schedule
		hasError bool
	}{
		{
			input: "Mon-Fri 09:00-17:00",
			expected: Schedule{
				Days:  []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
				Start: "09:00",
				End:   "17:00",
			},
			hasError: false,
		},
		{
			input: "Sat-Sun 10:00-18:00",
			expected: Schedule{
				Days:  []string{"Sat", "Sun"},
				Start: "10:00",
				End:   "18:00",
			},
			hasError: false,
		},
		{
			input: "Mon 08:00-12:00",
			expected: Schedule{
				Days:  []string{"Mon"},
				Start: "08:00",
				End:   "12:00",
			},
			hasError: false,
		},
		{
			input:    "Mon-Fri 09:00", // Format invalide (manque l'heure de fin)
			expected: Schedule{},
			hasError: true,
		},
		{
			input:    "Mon-Fri09:00-17:00", // Format invalide (manque l'espace entre jours et heures)
			expected: Schedule{},
			hasError: true,
		},
		{
			input:    "Fri-Mon 09:00-17:00", // Ordre des jours invalide
			expected: Schedule{},
			hasError: true,
		},
		{
			input:    "Invalid 09:00-17:00", // Jours invalides
			expected: Schedule{},
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := ParseSchedule(test.input)
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Did not expect error for input '%s', but got: %v", test.input, err)
			}
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("For input '%s', expected %+v but got %+v", test.input, test.expected, result)
			}
		}
	}
}
