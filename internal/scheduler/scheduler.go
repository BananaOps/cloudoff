package scheduler

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	ec2 "github.com/bananaops/cloudoff/internal/aws"
)

var logger *slog.Logger

// Structure pour représenter les jours et les heures
type Schedule struct {
	Days  []string
	Start string
	End   string
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
				fmt.Printf("Error parsing schedule for instance %s: %v\n", instance.ID, err)
				continue
			}

			// Heure actuelle
			currentTime := time.Now()

			for _, schedule := range schedules {

				// Vérifier si l'heure actuelle et le jour sont dans l'horaire
				isInSchedule, err := IsTimeInSchedule(currentTime, schedule)
				if err != nil {
					fmt.Println("Erreur :", err)
				}
				if isInSchedule {
					logger.Info("Instance scheduled to stop", "instance", instance.ID, "schedule", schedule)
					if os.Getenv("DRY_RUN") == "false" {
						ec2.StopInstance(instance.ID, instance.Region)
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

			// Heure actuelle
			currentTime := time.Now()

			var uptime bool = false

			for _, schedule := range schedules {

				// Vérifier si l'heure actuelle et le jour sont dans l'horaire
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
				logger.Info("Instance scheduled to stop", "instance", instance.ID, "schedule", schedules)
				if os.Getenv("DRY_RUN") == "false" {
					ec2.StopInstance(instance.ID, instance.Region)
				}
			}

		}
	}

}

func UpscaleSchedule(instance ec2.Instance) {

	for _, tag := range instance.Tags {
		if tag.Key == "cloudoff:uptime" {
			schedules, err := ParseSchedule(tag.Value)
			if err != nil {
				fmt.Printf("Error parsing schedule for instance %s: %v\n", instance.ID, err)
				continue
			}

			// Heure actuelle
			currentTime := time.Now()

			for _, schedule := range schedules {

				// Vérifier si l'heure actuelle et le jour sont dans l'horaire
				isInSchedule, err := IsTimeInSchedule(currentTime, schedule)
				if err != nil {
					fmt.Println("Erreur :", err)
				}
				if isInSchedule {
					logger.Info("Instance scheduled to start", "instance", instance.ID, "schedule", schedule)
					if os.Getenv("DRY_RUN") == "false" {
						ec2.StartInstance(instance.ID, instance.Region)
					}
				}
			}
		}

		if tag.Key == "cloudoff:downtime" {
			schedules, err := ParseSchedule(tag.Value)
			if err != nil {
				fmt.Printf("Error parsing schedule for instance %s: %v\n", instance.ID, err)
				continue
			}

			// Heure actuelle
			currentTime := time.Now()

			var uptime bool = false

			for _, schedule := range schedules {

				// Vérifier si l'heure actuelle et le jour sont dans l'horaire
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
				logger.Info("Instance scheduled to start", "instance", instance.ID, "schedule", schedules)
				if os.Getenv("DRY_RUN") == "false" {
					ec2.StartInstance(instance.ID, instance.Region)
				}
			}

		}
	}

}

// ParseSchedule analyse une chaîne contenant plusieurs plages horaires
func ParseSchedule(input string) ([]Schedule, error) {
	var schedules []Schedule

	// Séparer les plages horaires par des virgules
	entries := strings.Split(input, ",")
	for _, entry := range entries {
		// Séparer les jours et les heures (ex. : "Mon-Fri 09:00-17:00")
		parts := strings.Fields(entry)
		if len(parts) != 2 {
			return nil, fmt.Errorf("format invalide pour l'entrée : %s", entry)
		}

		// Extraire les jours et les heures
		daysPart := parts[0]
		timePart := parts[1]

		// Gérer les plages de jours (ex. : "Mon-Fri")
		days, err := parseDays(daysPart)
		if err != nil {
			return nil, err
		}

		// Gérer les heures (ex. : "09:00-17:00")
		timeRange := strings.Split(timePart, "-")
		if len(timeRange) != 2 {
			return nil, fmt.Errorf("format invalide pour les heures : %s", timePart)
		}
		startTime := timeRange[0]
		endTime := timeRange[1]

		// Ajouter la plage horaire à la liste
		schedules = append(schedules, Schedule{
			Days:  days,
			Start: startTime,
			End:   endTime,
		})
	}

	return schedules, nil
}

// parseDays analyse une chaîne de jours (ex. : "Mon-Fri" ou "Sun")
func parseDays(daysPart string) ([]string, error) {
	var days []string
	dayMap := map[string]int{
		"Sun": 0, "Mon": 1, "Tue": 2, "Wed": 3, "Thu": 4, "Fri": 5, "Sat": 6,
	}

	// Gérer les plages de jours (ex. : "Mon-Fri")
	if strings.Contains(daysPart, "-") {
		rangeParts := strings.Split(daysPart, "-")
		if len(rangeParts) != 2 {
			return nil, fmt.Errorf("format invalide pour les jours : %s", daysPart)
		}
		startDay := rangeParts[0]
		endDay := rangeParts[1]

		// Vérifier que les jours sont valides
		startIndex, ok1 := dayMap[startDay]
		endIndex, ok2 := dayMap[endDay]
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("jour invalide dans la plage : %s", daysPart)
		}

		// Ajouter tous les jours dans la plage
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
		// Gérer un seul jour (ex. : "Sun")
		if _, ok := dayMap[daysPart]; !ok {
			return nil, fmt.Errorf("jour invalide : %s", daysPart)
		}
		days = append(days, daysPart)
	}

	return days, nil
}

// Fonction pour vérifier si une heure et un jour sont dans un horaire
func IsTimeInSchedule(currentTime time.Time, schedule Schedule) (bool, error) {
	// Vérifier si le jour actuel est dans la liste des jours
	currentDay := currentTime.Weekday().String()[:3] // Récupère les 3 premières lettres du jour (ex : "Mon")
	isDayValid := false
	for _, day := range schedule.Days {
		if strings.EqualFold(day, currentDay) {
			isDayValid = true
			break
		}
	}

	if !isDayValid {
		return false, nil // Le jour actuel n'est pas dans la liste
	}

	// Parse l'heure de début
	startTime, err := time.Parse("15:04", schedule.Start)
	if err != nil {
		return false, fmt.Errorf("heure de début invalide : %v", err)
	}

	// Parse l'heure de fin
	endTime, err := time.Parse("15:04", schedule.End)
	if err != nil {
		return false, fmt.Errorf("heure de fin invalide : %v", err)
	}

	// Extraire uniquement l'heure et les minutes de l'heure actuelle
	current := time.Date(0, 1, 1, currentTime.Hour(), currentTime.Minute(), 0, 0, time.UTC)

	// Vérifier si l'heure actuelle est dans la période
	if current.Equal(startTime) || current.Equal(endTime) || (current.After(startTime) && current.Before(endTime)) {
		return true, nil
	}

	return false, nil
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}
