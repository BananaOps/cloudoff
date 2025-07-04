package clean

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	ec2 "github.com/bananaops/cloudoff/internal/aws"
)

var logger *slog.Logger

// CleanEC2Instance cleans up EC2 instances based on the provided schedule.
func CleanEC2Instance() {
	ec2List := ec2.DiscoverEC2Instances()

	for _, instance := range ec2List {
		for _, tag := range instance.Tags {
			if tag.Key == "cloudoff:ttl" {
				if DurationExceeded(instance) {
					// Perform cleanup action (e.g., terminate the instance)
					err := ec2.TerminateInstance(instance.InstanceId, instance.Region)
					if err != nil {
						logger.Error("error terminating instance", "instance", instance.ID, "region", instance.Region, "error", err)
						continue
					}
					// Log the termination action
					logger.Info("Instance terminated", "instance", instance.ID, "region", instance.Region, "AttachTime", instance.AttachTime, "ttl", tag.Value)
					continue
				}

			}
		}
	}
}

// Duration Exceeded Function
func DurationExceeded(instance ec2.Instance) bool {
	// Check if the instance has a "cloudoff:ttl" tag
	for _, tag := range instance.Tags {

		if tag.Key == "cloudoff:ttl" {
			if tag.Value == "infinity" {
				return false // If the tag value is "infinity", do not consider it for cleanup
			}
			// Parse the duration from the tag value
			duration, err := parseDuration(tag.Value)
			if err != nil {
				// Handle error (e.g., log it)
				return false
			}

			// Check if the instance's launch time exceeds the specified duration
			return isDurationExceeded(instance.AttachTime, duration)
		}
	}
	return false
}

// isDurationExceeded checks if the duration between a given time and the current time exceeds a specified duration.
func isDurationExceeded(t time.Time, d time.Duration) bool {
	// Calculate the elapsed time between the given time and now
	elapsed := time.Since(t)

	// Check if the elapsed time is greater than the specified duration
	return elapsed > d
}

func parseDuration(input string) (time.Duration, error) {
	// Check that the input string is not empty
	if len(input) < 2 {
		return 0, errors.New("invalid input: must contain a number followed by a unit (w, d, h)")
	}

	// Extract the numeric part and the unit
	numberPart := input[:len(input)-1]
	unit := input[len(input)-1:]

	// Convert the numeric part to an integer
	value, err := strconv.Atoi(numberPart)
	if err != nil {
		return 0, errors.New("invalid input: the numeric part is incorrect")
	}

	// Check if the value is negative
	if value < 0 {
		return 0, errors.New("invalid input: duration value cannot be negative")
	}

	// Calculate the duration based on the unit
	switch strings.ToLower(unit) {
	case "w": // weeks
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "d": // days
		return time.Duration(value) * 24 * time.Hour, nil
	case "h": // hours
		return time.Duration(value) * time.Hour, nil
	default:
		return 0, errors.New("invalid unit: must be 'w', 'd', or 'h'")
	}
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
