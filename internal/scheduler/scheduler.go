package scheduler

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ec2 "github.com/bananaops/cloudoff/internal/aws"
	"github.com/bananaops/cloudoff/internal/clean"
)

var logger *slog.Logger

// Structure for scheduling information
type Schedule struct {
	Days     []string
	Start    string
	End      string
	Timezone string
}

func ScheduleEC2Instance() {

	ec2List := ec2.DiscoverEC2Instances()

	for _, instance := range ec2List {
		if instance.State == "running" {
			DownscaleSchedule(instance)
		}

		if instance.State == "stopped" {
			UpscaleSchedule(instance)
		}

	}
}

func DownscaleSchedule(instance ec2.Instance) {

	for _, tag := range instance.Tags {
		if tag.Key == "cloudoff:downtime" {
			schedules, err := ParseSchedule(tag.Value)
			if err != nil {
				logger.Error("error parsing schedule for instance", "instance", instance.ID, "error", err)
				continue
			}

			// Current time
			currentTime := time.Now()

			for _, schedule := range schedules {

				// Checker if the current time and day are in the schedule
				isInSchedule, err := IsTimeInSchedule(currentTime, schedule)
				if err != nil {
					fmt.Println("Erreur :", err)
				}
				if isInSchedule {
					if os.Getenv("DRYRUN") != "true" {
						err := ec2.StopInstance(instance.ID, instance.Region)
						if err != nil {
							logger.Error("error stopping instance", "instance", instance.ID, "error", err)
						} else {
							logger.Info("instance stopped successfully", "instance", instance.ID)
						}
					}
				}
			}
		}

		if tag.Key == "cloudoff:uptime" {
			schedules, err := ParseSchedule(tag.Value)
			if err != nil {
				fmt.Printf("Error parsing schedule for instance %s: %v\n", instance.ID, err)
				continue
			}

			// Current time
			currentTime := time.Now()

			var uptime = false

			for _, schedule := range schedules {

				// Check if the current time and day are in the schedule
				isInSchedule, err := IsTimeInSchedule(currentTime, schedule)

				if err != nil {
					fmt.Println("Erreur :", err)
				}
				if isInSchedule {
					uptime = true
					break

				}

			}

			if !uptime {
				if os.Getenv("DRYRUN") != "true" {
					err := ec2.StopInstance(instance.ID, instance.Region)
					if err != nil {
						logger.Error("error stopping instance", "instance", instance.ID, "error", err)
					} else {
						logger.Info("instance stopped successfully", "instance", instance.ID)
					}
				}
			}

		}
	}

}

func UpscaleSchedule(instance ec2.Instance) {

	if !clean.DurationExceeded(instance) {

		for _, tag := range instance.Tags {
			if tag.Key == "cloudoff:uptime" {
				schedules, err := ParseSchedule(tag.Value)
				if err != nil {
					logger.Error("error parsing schedule for instance", "instance", instance.ID, "error", err)
					continue
				}

				// Current time
				currentTime := time.Now()

				for _, schedule := range schedules {

					// Checker if the current time and day are in the schedule
					isInSchedule, err := IsTimeInSchedule(currentTime, schedule)
					if err != nil {
						logger.Error("error checking schedule for instance", "instance", instance.ID, "error", err)
					}
					if isInSchedule {
						if os.Getenv("DRYRUN") != "true" {
							err := ec2.StartInstance(instance.ID, instance.Region)
							if err != nil {
								logger.Error("error starting instance", "instance", instance.ID, "error", err)
							} else {
								logger.Info("instance started successfully", "instance", instance.ID)
							}
						}
					}
				}
			}

			if tag.Key == "cloudoff:downtime" {
				schedules, err := ParseSchedule(tag.Value)
				if err != nil {
					logger.Error("error parsing schedule for instance", "instance", instance.ID, "error", err)
					continue
				}

				// Current time
				currentTime := time.Now()

				var uptime = false

				for _, schedule := range schedules {

					// Check if the current time and day are in the schedule
					isInSchedule, err := IsTimeInSchedule(currentTime, schedule)
					if err != nil {
						logger.Error("error checking schedule for instance", "instance", instance.ID, "error", err)
					}
					if isInSchedule {
						uptime = true
						break

					}

				}

				if !uptime {
					if os.Getenv("DRYRUN") != "true" {
						err := ec2.StartInstance(instance.ID, instance.Region)
						if err != nil {
							logger.Error("error starting instance", "instance", instance.ID, "error", err)
						} else {
							logger.Info("instance started successfully", "instance", instance.ID)
						}
					}
				}
			}
		}
	}
}

// ParseSchedule convert string into a slice of Schedule structs
func ParseSchedule(input string) ([]Schedule, error) {
	var schedules []Schedule

	// Split the input by commas to handle multiple entries
	entries := strings.Split(input, ",")
	for _, entry := range entries {

		var timezone string
		var days []string
		var startTime, endTime string
		var err error

		if entry != "infinity" {

			normalizedInput := strings.ReplaceAll(entry, "_", " ")
			// Split the entry into parts
			parts := strings.Fields(normalizedInput)
			if len(parts) < 2 {
				return nil, fmt.Errorf("invalid format for input : %s", entry)
			}

			// Extract days, time range, and timezone
			daysPart := parts[0]
			timePart := parts[1]

			if len(parts) > 2 {
				timezone = parts[2]
			} else {
				timezone = "UTC"
			}

			// Generate a list of days from the days part
			days, err = parseDays(daysPart)
			if err != nil {
				return nil, err
			}

			// Generate hour range from the time part (ex. : "09:00-17:00")
			timeRange := strings.Split(timePart, "-")
			if len(timeRange) != 2 {
				return nil, fmt.Errorf("invalid format for hours : %s", timePart)
			}
			startTime = timeRange[0]
			endTime = timeRange[1]
		} else {
			// If the entry is "infinity", we set default values
			days = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"} // All days of the week
			startTime = "00:00" // Start at midnight
			endTime = "23:59"   // End at the end of the day
			timezone = "UTC"    // Default timezone
		}

		// Add the schedule to the list
		schedules = append(schedules, Schedule{
			Days:     days,
			Start:    startTime,
			End:      endTime,
			Timezone: timezone,
		})
	}

	return schedules, nil
}

// parseDays analyse a slice of days and returns a slice of valid day strings
func parseDays(daysPart string) ([]string, error) {
	var days []string
	dayMap := map[string]int{
		"Sun": 0, "Mon": 1, "Tue": 2, "Wed": 3, "Thu": 4, "Fri": 5, "Sat": 6,
	}

	// generate a list of days from the days part
	if strings.Contains(daysPart, "-") {
		rangeParts := strings.Split(daysPart, "-")
		if len(rangeParts) != 2 {
			return nil, fmt.Errorf("invalid format for days : %s", daysPart)
		}
		startDay := rangeParts[0]
		endDay := rangeParts[1]

		// Check if startDay and endDay are valid days
		startIndex, ok1 := dayMap[startDay]
		endIndex, ok2 := dayMap[endDay]
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("invalid day in range : %s", daysPart)
		}

		// Add days in the range
		for i := startIndex; ; i = (i + 1) % 7 {
			for day, index := range dayMap {
				if index == i {
					days = append(days, day)
					break
				}
			}
			if i == endIndex {
				break
			}
		}
	} else {
		// Manage single days
		if _, ok := dayMap[daysPart]; !ok {
			return nil, fmt.Errorf("invalid day: %s", daysPart)
		}
		days = append(days, daysPart)
	}

	return days, nil
}

// Check if the current time is within the schedule
func IsTimeInSchedule(currentTime time.Time, schedule Schedule) (bool, error) {
	// Aply the timezone to the current time
	location, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		return false, fmt.Errorf("invalid timezone : %v", err)
	}
	currentTime = currentTime.In(location)

	// Check if the current day is in the schedule
	currentDay := currentTime.Weekday().String()[:3] // Get the first three letters of the weekday
	isDayValid := false
	for _, day := range schedule.Days {
		if strings.EqualFold(day, currentDay) {
			isDayValid = true
			break
		}
	}

	if !isDayValid {
		return false, nil // The current day is not in the schedule
	}

	// Parse hour of start
	startTime, err := time.ParseInLocation("15:04", schedule.Start, location)
	if err != nil {
		return false, fmt.Errorf("invalid start time : %v", err)
	}

	// Parse hour of end
	endTime, err := time.ParseInLocation("15:04", schedule.End, location)
	if err != nil {
		return false, fmt.Errorf("invalid end time: %v", err)
	}

	// Extract the current time without the date
	current := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), 0, 0, location)

	// Check if the current time is within the schedule
	if current.Equal(startTime) || current.Equal(endTime) || (current.After(startTime) && current.Before(endTime)) {
		return true, nil
	}

	return false, nil
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
