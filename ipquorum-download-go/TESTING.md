# Testing Guide

## Overview

This project uses Go's built-in testing framework. Tests are organized alongside the code they test, following Go conventions.

## Test Structure

```
ipquorum-download-go/
├── internal/
│   ├── errors/
│   │   ├── errors.go
│   │   └── errors_test.go      # Tests for error types
│   └── utils/
│       ├── utils.go
│       └── utils_test.go        # Tests for utility functions
└── Makefile                     # Build and test commands
```

## Running Tests

### Run All Tests
```bash
make test
```
or
```bash
go test -v ./...
```

### Run Tests for Specific Package
```bash
# Test utils package
go test -v ./internal/utils

# Test errors package
go test -v ./internal/errors
```

### Run Tests with Coverage
```bash
make test-coverage
```
This generates:
- `coverage.out` - Coverage data
- `coverage.html` - HTML coverage report (open in browser)

### Run Specific Test
```bash
go test -v ./internal/utils -run TestToBool
go test -v ./internal/errors -run TestValidationError
```

## Test Coverage

### Current Test Coverage

#### `internal/utils` Package
- ✅ `ToBool()` - 25 test cases covering:
  - True values: "true", "1", "yes", "y", "on" (case-insensitive)
  - False values: "false", "0", "no", "n", "off", "" (case-insensitive)
  - Invalid inputs and edge cases
  - Whitespace handling

- ✅ `MaskPassword()` - 13 test cases covering:
  - Empty passwords
  - Short passwords
  - Normal passwords with various showChars values
  - Edge cases (password length equals showChars)

#### `internal/errors` Package
- ✅ `IPQuorumError` - Base error type tests
- ✅ `ValidationError` - Configuration validation errors
- ✅ `AuthenticationError` - Authentication failure errors
- ✅ `NetworkError` - Network connectivity errors
- ✅ `APIError` - API call failure errors
- ✅ Error type assertions and unwrapping

## Writing New Tests

### Test File Naming
- Test files must end with `_test.go`
- Place test files in the same directory as the code being tested
- Example: `utils.go` → `utils_test.go`

### Test Function Naming
```go
func TestFunctionName(t *testing.T) {
    // Test implementation
}
```

### Table-Driven Tests (Recommended)
```go
func TestToBool(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected bool
    }{
        {"true lowercase", "true", true},
        {"false lowercase", "false", false},
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
```

### Subtests
Use `t.Run()` to create subtests for better organization:
```go
func TestMyFunction(t *testing.T) {
    t.Run("valid input", func(t *testing.T) {
        // Test valid input
    })
    
    t.Run("invalid input", func(t *testing.T) {
        // Test invalid input
    })
}
```

## Benchmarks

Benchmark tests are included for performance-critical functions:

```bash
# Run benchmarks
go test -bench=. ./internal/utils

# Run specific benchmark
go test -bench=BenchmarkToBool ./internal/utils
```

Example benchmark output:
```
BenchmarkToBool-8         50000000    25.4 ns/op
BenchmarkMaskPassword-8   10000000    120 ns/op
```

## Continuous Integration

Tests are automatically run in CI/CD pipelines. Ensure all tests pass before submitting pull requests.

## Best Practices

1. **Write tests first** (TDD approach when possible)
2. **Test edge cases** - empty strings, nil values, boundary conditions
3. **Use descriptive test names** - clearly indicate what is being tested
4. **Keep tests independent** - tests should not depend on each other
5. **Use table-driven tests** - for testing multiple inputs/outputs
6. **Test error conditions** - verify error handling works correctly
7. **Maintain high coverage** - aim for >80% code coverage
8. **Run tests before committing** - ensure nothing is broken

## Future Test Additions

Consider adding tests for:
- [ ] `pkg/client` - HTTP client functionality
- [ ] `pkg/config` - Configuration validation
- [ ] `pkg/logger` - Logging functionality
- [ ] `pkg/password` - Password handling
- [ ] `main.go` - CLI flag parsing and command execution
- [ ] Integration tests - End-to-end testing with mock API

## Troubleshooting

### Tests Fail to Run
```bash
# Ensure dependencies are installed
go mod download

# Verify Go version (requires Go 1.19+)
go version

# Clean and rebuild
make clean
make build
```

### Coverage Report Not Generated
```bash
# Manually generate coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Coverage](https://go.dev/blog/cover)