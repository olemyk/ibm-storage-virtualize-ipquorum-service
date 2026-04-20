package utils

import (
	"strings"
)

// ToBool converts various string inputs to boolean
func ToBool(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))

	switch v {
	case "true", "1", "yes", "y", "on":
		return true
	case "false", "0", "no", "n", "off", "":
		return false
	default:
		return false
	}
}

//
