package services

import (
	"craftsbite-backend/internal/models"
	"fmt"
	"strings"
	"time"
)

// parseMealTypes parses a comma-separated string of meal types into a slice
func parseMealTypes(mealsStr string) []models.MealType {
	if mealsStr == "" {
		return []models.MealType{}
	}

	parts := strings.Split(mealsStr, ",")
	result := make([]models.MealType, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, models.MealType(trimmed))
		}
	}

	return result
}

// serializeMealTypes converts a slice of MealType to comma-separated string
func serializeMealTypes(mealTypes []models.MealType) string {
	if len(mealTypes) == 0 {
		return ""
	}

	parts := make([]string, len(mealTypes))
	for i, mt := range mealTypes {
		parts[i] = string(mt)
	}

	return strings.Join(parts, ",")
}

func validateDate(date string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD")
	}
	return nil
}
