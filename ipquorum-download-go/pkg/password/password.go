package password

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/olemyk/ipquorum-go/internal/errors"
	"golang.org/x/term"
)

// GetPassword retrieves password based on the specified method
// method: "prompt", "file", "env", or "direct"
// source: file path for "file" method, or direct password for "direct" method
// username: username for prompt display
func GetPassword(method, source, username string) (string, error) {
	switch method {
	case "prompt":
		return PromptPassword(username)
	case "file":
		return ReadPasswordFromFile(source)
	case "env":
		return os.Getenv("VIRTUALIZE_PASSWORD"), nil
	case "direct":
		return source, nil
	default:
		return "", errors.NewValidationError(fmt.Sprintf("unknown password method: %s", method))
	}
}

// ReadPasswordFromFile reads password from a file
func ReadPasswordFromFile(filepath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return "", errors.NewValidationError(fmt.Sprintf("password file not found: %s", filepath))
	}

	// Check file permissions (warn if too permissive on Unix-like systems)
	if info, err := os.Stat(filepath); err == nil {
		mode := info.Mode()
		// On Unix-like systems, warn if file is readable by others (world-readable)
		// 440 (owner+group read) is acceptable for systemd services
		// 400 (owner read only) is more secure
		if mode&0007 != 0 {
			fmt.Fprintf(os.Stderr, "Warning: Password file is world-readable (%o). Recommend: chmod 400 or 440 %s\n", mode.Perm(), filepath)
		}
	}

	// Read file content
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", errors.NewValidationError(fmt.Sprintf("error reading password file: %v", err))
	}

	// Trim whitespace and newlines
	password := strings.TrimSpace(string(data))

	if password == "" {
		return "", errors.NewValidationError("password file is empty")
	}

	return password, nil
}

// PromptPassword prompts the user for password with hidden input
func PromptPassword(username string) (string, error) {
	fmt.Fprintf(os.Stderr, "Enter password for %s: ", username)

	// Read password with hidden input
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr) // Print newline after hidden input

	if err != nil {
		return "", errors.NewValidationError(fmt.Sprintf("error reading password: %v", err))
	}

	password := string(bytePassword)

	if password == "" {
		return "", errors.NewValidationError("password cannot be empty")
	}

	return password, nil
}

// MaskPassword masks a password for logging
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
