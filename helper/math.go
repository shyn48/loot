package helper

import (
	"fmt"
	"math"
)

func IntToFloatString(value int) string {
	floatValue := float64(value)
	floatValue = floatValue / math.Pow(10, 6)

	return fmt.Sprintf("%f", floatValue)
}

// HumanBytes formats a raw byte count into a friendly, human-readable string
// such as "0 B", "812 KB", "12.3 MB" or "1.21 GB". It uses binary (1024) units.
func HumanBytes(bytes int) string {
	if bytes < 0 {
		return "-"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	units := []string{"KB", "MB", "GB", "TB", "PB"}

	i := -1
	for value >= unit && i < len(units)-1 {
		value /= unit
		i++
	}

	// Fewer decimals as the number grows larger, so it stays compact.
	switch {
	case value >= 100:
		return fmt.Sprintf("%.0f %s", value, units[i])
	case value >= 10:
		return fmt.Sprintf("%.1f %s", value, units[i])
	default:
		return fmt.Sprintf("%.2f %s", value, units[i])
	}
}
