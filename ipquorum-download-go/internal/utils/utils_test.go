package utils

import (
	"testing"

	"github.com/olemyk/ipquorum-go/pkg/password"
)

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// True cases
		{"true lowercase", "true", true},
		{"true uppercase", "TRUE", true},
		{"true mixed case", "TrUe", true},
		{"1 string", "1", true},
		{"yes lowercase", "yes", true},
		{"yes uppercase", "YES", true},
		{"y lowercase", "y", true},
		{"y uppercase", "Y", true},
		{"on lowercase", "on", true},
		{"on uppercase", "ON", true},
		{"true with spaces", "  true  ", true},

		// False cases
		{"false lowercase", "false", false},
		{"false uppercase", "FALSE", false},
		{"0 string", "0", false},
		{"no lowercase", "no", false},
		{"no uppercase", "NO", false},
		{"n lowercase", "n", false},
		{"n uppercase", "N", false},
		{"off lowercase", "off", false},
		{"off uppercase", "OFF", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},

		// Invalid/default cases
		{"invalid string", "invalid", false},
		{"random text", "random", false},
		{"number 2", "2", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToBool(tt.input)
			if result != tt.expected {
				t.Errorf("ToBool(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		showChars int
		expected  string
	}{
		// Empty password
		{"empty password", "", 0, "<empty>"},
		{"empty password with showChars", "", 3, "<empty>"},

		// Short passwords
		{"short password no show", "abc", 0, "***"},
		{"short password show all", "abc", 3, "***"},
		{"short password show more than length", "abc", 5, "***"},

		// Normal passwords with showChars = 0
		{"normal password no show", "password123", 0, "********"},
		{"long password no show", "verylongpassword123", 0, "********"},

		// Normal passwords with showChars > 0
		{"show last 3 chars", "password123", 3, "********123"},
		{"show last 4 chars", "mypassword", 4, "******word"},
		{"show last 1 char", "secret", 1, "*****t"},

		// Edge cases
		{"password length equals showChars", "pass", 4, "****"},
		{"single char password", "x", 0, "*"},
		{"single char password show 1", "x", 1, "*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := password.MaskPassword(tt.password, tt.showChars)
			if result != tt.expected {
				t.Errorf("MaskPassword(%q, %d) = %q, want %q", tt.password, tt.showChars, result, tt.expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkToBool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ToBool("true")
	}
}

func BenchmarkMaskPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		password.MaskPassword("password123", 3)
	}
}

func BenchmarkMaskPasswordShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		password.MaskPassword("pass", 2)
	}
}

func BenchmarkMaskPasswordLong(b *testing.B) {
	for i := 0; i < b.N; i++ {
		password.MaskPassword("thisIsAVeryLongPasswordWithManyCharacters1234567890", 5)
	}
}

func BenchmarkMaskPasswordNoShow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		password.MaskPassword("password123", 0)
	}
}

//
