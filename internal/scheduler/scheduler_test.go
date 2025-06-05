package scheduler

import (
	"reflect"
	"testing"
	"time"
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
					Days:     []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start:    "09:00",
					End:      "17:00",
					Timezone: "UTC",
				},
			},
			wantErr: false,
		},
		{
			name:  "Range of days with time range",
			input: "Mon-Fri 09:00-17:00",
			expected: []Schedule{
				{
					Days:     []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start:    "09:00",
					End:      "17:00",
					Timezone: "UTC",
				},
			},
			wantErr: false,
		},
		{
			name:  "Range of days with time range with timezone",
			input: "Mon-Fri_09:00-17:00_Europe/Paris",
			expected: []Schedule{
				{
					Days:     []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start:    "09:00",
					End:      "17:00",
					Timezone: "Europe/Paris",
				},
			},
			wantErr: false,
		},
		{
			name:  "Multiple schedules",
			input: "Mon-Fri 09:00-17:00,Sun 00:00-23:59",
			expected: []Schedule{
				{
					Days:     []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start:    "09:00",
					End:      "17:00",
					Timezone: "UTC",
				},
				{
					Days:     []string{"Sun"},
					Start:    "00:00",
					End:      "23:59",
					Timezone: "UTC",
				},
			},
			wantErr: false,
		},
		{
			name:  "Multiple schedules with timezone",
			input: "Mon-Fri 09:00-17:00_Europe/Paris,Sun 00:00-23:59 Europe/Paris",
			expected: []Schedule{
				{
					Days:     []string{"Mon", "Tue", "Wed", "Thu", "Fri"},
					Start:    "09:00",
					End:      "17:00",
					Timezone: "Europe/Paris",
				},
				{
					Days:     []string{"Sun"},
					Start:    "00:00",
					End:      "23:59",
					Timezone: "Europe/Paris",
				},
			},
			wantErr: false,
		},
		{
			name:  "Infinity schedule",
			input: "infinity",
			expected: []Schedule{
				{
					Days:     []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
					Start:    "00:00",
					End:      "23:59",
					Timezone: "UTC",
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
					Days:     []string{"Tue"},
					Start:    "10:00",
					End:      "12:00",
					Timezone: "UTC",
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
func TestIsTimeInSchedule(t *testing.T) {
	tests := []struct {
		name        string
		currentTime time.Time
		schedule    Schedule
		expected    bool
		wantErr     bool
	}{
		{
			name:        "Within schedule time",
			currentTime: time.Date(2023, 10, 2, 10, 0, 0, 0, time.UTC), // Mon 10:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    true,
			wantErr:     false,
		},
		{
			name:        "Outside schedule time",
			currentTime: time.Date(2023, 10, 2, 18, 0, 0, 0, time.UTC), // Mon 18:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    false,
			wantErr:     false,
		},
		{
			name:        "Invalid timezone",
			currentTime: time.Date(2023, 10, 2, 10, 0, 0, 0, time.UTC), // Mon 10:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "Invalid/Timezone"},
			expected:    false,
			wantErr:     true,
		},
		{
			name:        "Day not in schedule",
			currentTime: time.Date(2023, 10, 3, 10, 0, 0, 0, time.UTC), // Tue 10:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    false,
			wantErr:     false,
		},
		{
			name:        "Exact start time",
			currentTime: time.Date(2023, 10, 2, 9, 0, 0, 0, time.UTC), // Mon 09:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    true,
			wantErr:     false,
		},
		{
			name:        "Exact end time",
			currentTime: time.Date(2023, 10, 2, 17, 0, 0, 0, time.UTC), // Mon 17:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    true,
			wantErr:     false,
		},
		{
			name:        "Within schedule time in UTC",
			currentTime: time.Date(2023, 10, 2, 10, 0, 0, 0, time.UTC), // Mon 10:00 UTC
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    true,
			wantErr:     false,
		},
		{
			name:        "Outside schedule time in UTC",
			currentTime: time.Date(2023, 10, 2, 18, 0, 0, 0, time.UTC), // Mon 18:00 UTC
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "UTC"},
			expected:    false,
			wantErr:     false,
		},
		{
			name:        "Within schedule time in Europe/Paris timezone",
			currentTime: time.Date(2023, 10, 2, 10, 0, 0, 0, time.UTC), // Mon 10:00 UTC
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "Europe/Paris"},
			expected:    true,
			wantErr:     false,
		},
		{
			name:        "Outside schedule time in Europe/Paris timezone",
			currentTime: time.Date(2023, 10, 2, 18, 0, 0, 0, time.UTC), // Mon 18:00 UTC
			schedule:    Schedule{Days: []string{"Mon"}, Start: "09:00", End: "17:00", Timezone: "Europe/Paris"},
			expected:    false,
			wantErr:     false,
		},
		{name: "Schedule with infinity",
			currentTime: time.Date(2023, 10, 2, 10, 0, 0, 0, time.UTC), // Mon 10:00
			schedule:    Schedule{Days: []string{"Mon"}, Start: "00:00", End: "23:59", Timezone: "UTC"},
			expected:    true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsTimeInSchedule(tt.currentTime, tt.schedule)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsTimeInSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("IsTimeInSchedule() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
