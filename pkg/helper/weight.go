package helper

import "fmt"

// Weight conversion constants
const (
	// Grams to other units
	GramsPerOunce = 28.3495
	GramsPerPound = 453.592
)

// ConvertToGrams converts a weight from any unit to grams
func ConvertToGrams(weight float64, fromUnit string) float64 {
	switch fromUnit {
	case "oz":
		return weight * GramsPerOunce
	case "lb":
		return weight * GramsPerPound
	case "g":
		return weight
	default:
		return weight // Default to assuming grams
	}
}

// ConvertFromGrams converts a weight from grams to the specified unit
func ConvertFromGrams(weight float64, toUnit string) float64 {
	switch toUnit {
	case "oz":
		return weight / GramsPerOunce
	case "lb":
		return weight / GramsPerPound
	case "g":
		return weight
	default:
		return weight // Default to grams
	}
}

// FormatWeight formats a weight in grams to a human-readable string
func FormatWeight(weightInGrams float64, displayUnit string) string {
	convertedWeight := ConvertFromGrams(weightInGrams, displayUnit)

	switch displayUnit {
	case "oz":
		return fmt.Sprintf("%.1f oz", convertedWeight)
	case "lb":
		return fmt.Sprintf("%.2f lb", convertedWeight)
	case "g":
		return fmt.Sprintf("%d g", int(weightInGrams))
	default:
		return fmt.Sprintf("%d g", int(weightInGrams))
	}
}
