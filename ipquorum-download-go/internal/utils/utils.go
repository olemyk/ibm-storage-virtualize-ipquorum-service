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

// MaskPassword masks a password for logging purposes
// showChars: number of characters to show at the end (0 = show none)
func MaskPassword(password string, showChars int) string {
	if password == "" {
		return "<empty>"
	}

	passLen := len(password)

	if passLen <= showChars {
		return strings.Repeat("*", passLen)
	}

	if showChars > 0 {
		maskLen := passLen - showChars
		return strings.Repeat("*", maskLen) + password[maskLen:]
	}

	// Default: show 8 asterisks or password length, whichever is smaller
	maskLen := 8
	if passLen < maskLen {
		maskLen = passLen
	}
	return strings.Repeat("*", maskLen)
}

// Made with help from Bob
