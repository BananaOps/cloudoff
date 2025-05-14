package scheduler

import (
	"fmt"
	"strings"
)

// Structure pour représenter les jours et les heures
type Schedule struct {
	Days  []string
	Start string
	End   string
}

var allDays = []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}

// Fonction pour parser une chaîne de type "Mon-Fri 09:00-17:00"
func ParseSchedule(input string) (Schedule, error) {
	// Diviser la chaîne en deux parties : jours et heures
	parts := strings.Split(input, " ")
	if len(parts) != 2 {
		return Schedule{}, fmt.Errorf("format invalide, attendu 'Days Start-End'")
	}

	// Extraire les jours
	daysRange := parts[0]
	days := []string{}
	if strings.Contains(daysRange, "-") {
		// Gestion des plages de jours (ex: Mon-Fri)
		daysSplit := strings.Split(daysRange, "-")

		if len(daysSplit) != 2 {
			return Schedule{}, fmt.Errorf("format des jours invalide")
		}
		startDay, endDay := daysSplit[0], daysSplit[1]
		startIndex, endIndex := -1, -1
		for i, day := range allDays {
			if day == startDay {
				startIndex = i
			}
			if day == endDay {
				endIndex = i
			}
		}
		if startIndex == -1 || endIndex == -1 || startIndex > endIndex {
			return Schedule{}, fmt.Errorf("jours invalides ou hors ordre")
		}
		days = allDays[startIndex : endIndex+1]
	} else {
		// Gestion d'un seul jour (ex: Mon)
		for _, day := range allDays {
			if day == daysRange {
				days = append(days, daysRange)
				break
			} else {
				return Schedule{}, fmt.Errorf("jour invalide")
			}
		}
	}

	// Extraire les heures
	hoursRange := parts[1]
	hoursSplit := strings.Split(hoursRange, "-")
	if len(hoursSplit) != 2 {
		return Schedule{}, fmt.Errorf("format des heures invalide")
	}
	startTime, endTime := hoursSplit[0], hoursSplit[1]

	// Retourner la structure Schedule
	return Schedule{
		Days:  days,
		Start: startTime,
		End:   endTime,
	}, nil
}
