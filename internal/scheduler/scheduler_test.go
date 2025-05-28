package scheduler

import (
	"reflect"
	"testing"
)

func TestParseSchedule(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Schedule
		wantErr  bool
	}{
		{
			name:  "Range of days with time range",
			input: "Mon-Fri_09:00-17:00",
			expected: []Schedule{
				{
					Days:  []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start: "09:00",
					End:   "17:00",
				},
			},
			wantErr: false,
		},
		{
			name:  "Range of days with time range",
			input: "Mon-Fri 09:00-17:00",
			expected: []Schedule{
				{
					Days:  []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start: "09:00",
					End:   "17:00",
				},
			},
			wantErr: false,
		},
		{
			name:  "Multiple schedules",
			input: "Mon-Fri 09:00-17:00,Sun 00:00-23:59",
			expected: []Schedule{
				{
					Days:  []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start: "09:00",
					End:   "17:00",
				},
				{
					Days:  []string{"Sun"},
					Start: "00:00",
					End:   "23:59",
				},
			},
			wantErr: false,
		},
		{
			name:     "Invalid format (missing time range)",
			input:    "Mon-Fri",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid day range",
			input:    "Mon-Funday 09:00-17:00",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Invalid time range",
			input:    "Mon-Fri 09:00",
			expected: nil,
			wantErr:  true,
		},
		{
			name:  "Single day without range",
			input: "Tue 10:00-12:00",
			expected: []Schedule{
				{
					Days:  []string{"Tue"},
					Start: "10:00",
					End:   "12:00",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSchedule(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseSchedule() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
