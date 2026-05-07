# IPQuorum Management Platform - Testing Guide

This comprehensive guide covers all testing aspects of the IPQuorum Management Platform, including unit tests, integration tests, end-to-end tests, and performance testing.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Backend Testing](#backend-testing)
3. [Frontend Testing](#frontend-testing)
4. [Integration Testing](#integration-testing)
5. [End-to-End Testing](#end-to-end-testing)
6. [Performance Testing](#performance-testing)
7. [Security Testing](#security-testing)
8. [CI/CD Integration](#cicd-integration)

## Testing Philosophy

### Test Pyramid

```
        /\
       /  \      E2E Tests (Few)
      /____\
     /      \    Integration Tests (Some)
    /________\
   /          \  Unit Tests (Many)
  /____________\
```

- **Unit Tests (70%)**: Fast, isolated, test individual functions
- **Integration Tests (20%)**: Test component interactions
- **E2E Tests (10%)**: Test complete user workflows

### Coverage Goals

- **Backend**: Minimum 80% code coverage
- **Frontend**: Minimum 70% code coverage
- **Critical Paths**: 100% coverage (auth, instance management)

## Backend Testing

### Setup

```bash
cd server
go test ./... -v
go test ./... -cover
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Unit Tests

#### Testing Structure

```
server/
├── internal/
│   ├── auth/
│   │   ├── jwt.go
│   │   └── jwt_test.go
│   ├── executor/
│   │   ├── executor.go
│   │   └── executor_test.go
│   └── storage/
│       ├── storage.go
│       └── storage_test.go
└── pkg/
    ├── metrics/
    │   ├── metrics.go
    │   └── metrics_test.go
    └── logger/
        ├── logger.go
        └── logger_test.go
```

#### Example: JWT Tests

```go
// server/internal/auth/jwt_test.go
package auth

import (
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
    jwtManager := NewJWTManager("test-secret", 1*time.Hour)
    
    token, err := jwtManager.GenerateToken(1, "testuser", "admin")
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
}

func TestValidateToken(t *testing.T) {
    jwtManager := NewJWTManager("test-secret", 1*time.Hour)
    
    token, _ := jwtManager.GenerateToken(1, "testuser", "admin")
    claims, err := jwtManager.ValidateToken(token)
    
    assert.NoError(t, err)
    assert.Equal(t, int64(1), claims.UserID)
    assert.Equal(t, "testuser", claims.Username)
    assert.Equal(t, "admin", claims.Role)
}

func TestExpiredToken(t *testing.T) {
    jwtManager := NewJWTManager("test-secret", -1*time.Hour)
    
    token, _ := jwtManager.GenerateToken(1, "testuser", "admin")
    _, err := jwtManager.ValidateToken(token)
    
    assert.Error(t, err)
}
```

#### Example: Executor Tests

```go
// server/internal/executor/executor_test.go
package executor

import (
    "context"
    "testing"
    "time"
    "github.com/stretchr/testify/assert"
)

func TestExecuteCommand(t *testing.T) {
    exec := NewExecutor()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    output, err := exec.Execute(ctx, "echo", "hello")
    assert.NoError(t, err)
    assert.Contains(t, output, "hello")
}

func TestExecuteCommandTimeout(t *testing.T) {
    exec := NewExecutor()
    
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()
    
    _, err := exec.Execute(ctx, "sleep", "10")
    assert.Error(t, err)
}

func TestExecuteCommandInvalidCommand(t *testing.T) {
    exec := NewExecutor()
    
    ctx := context.Background()
    _, err := exec.Execute(ctx, "nonexistent-command")
    assert.Error(t, err)
}
```

#### Example: Storage Tests

```go
// server/internal/storage/storage_test.go
package storage

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *Storage {
    db, err := NewStorage(":memory:")
    assert.NoError(t, err)
    return db
}

func TestCreateInstance(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    instance := &Instance{
        Name: "test-instance",
        Host: "10.0.0.1",
        Port: 3993,
    }
    
    err := db.CreateInstance(instance)
    assert.NoError(t, err)
    assert.NotZero(t, instance.ID)
}

func TestGetInstance(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    instance := &Instance{Name: "test", Host: "10.0.0.1", Port: 3993}
    db.CreateInstance(instance)
    
    retrieved, err := db.GetInstance(instance.ID)
    assert.NoError(t, err)
    assert.Equal(t, instance.Name, retrieved.Name)
}

func TestListInstances(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    db.CreateInstance(&Instance{Name: "test1", Host: "10.0.0.1", Port: 3993})
    db.CreateInstance(&Instance{Name: "test2", Host: "10.0.0.2", Port: 3993})
    
    instances, err := db.ListInstances()
    assert.NoError(t, err)
    assert.Len(t, instances, 2)
}
```

### Integration Tests

#### API Integration Tests

```go
// server/internal/api/api_integration_test.go
package api

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
)

func setupTestServer(t *testing.T) *gin.Engine {
    gin.SetMode(gin.TestMode)
    // Setup test server with in-memory DB
    // ...
    return router
}

func TestLoginEndpoint(t *testing.T) {
    router := setupTestServer(t)
    
    loginData := map[string]string{
        "username": "admin",
        "password": "admin123",
    }
    body, _ := json.Marshal(loginData)
    
    req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.NotEmpty(t, response["token"])
}

func TestCreateInstanceEndpoint(t *testing.T) {
    router := setupTestServer(t)
    token := getTestToken(t, router)
    
    instanceData := map[string]interface{}{
        "name": "test-instance",
        "host": "10.0.0.1",
        "port": 3993,
    }
    body, _ := json.Marshal(instanceData)
    
    req, _ := http.NewRequest("POST", "/api/v1/instances", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+token)
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
}
```

### Running Backend Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/auth -v

# Run with race detection
go test ./... -race

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
go test ./... -bench=. -benchmem
```

## Frontend Testing

### Setup

```bash
cd web
npm install --save-dev @testing-library/react @testing-library/jest-dom @testing-library/user-event vitest jsdom
```

### Vitest Configuration

```typescript
// web/vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
      ],
    },
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
```

### Test Setup

```typescript
// web/src/test/setup.ts
import '@testing-library/jest-dom';
import { cleanup } from '@testing-library/react';
import { afterEach } from 'vitest';

afterEach(() => {
  cleanup();
});
```

### Component Tests

#### Example: Button Component

```typescript
// web/src/components/ui/Button.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from './Button';

describe('Button', () => {
  it('renders with text', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('handles click events', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);
    
    fireEvent.click(screen.getByText('Click me'));
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  it('applies variant styles', () => {
    render(<Button variant="destructive">Delete</Button>);
    const button = screen.getByText('Delete');
    expect(button).toHaveClass('bg-destructive');
  });

  it('disables when disabled prop is true', () => {
    render(<Button disabled>Disabled</Button>);
    expect(screen.getByText('Disabled')).toBeDisabled();
  });
});
```

#### Example: Login Page

```typescript
// web/src/pages/Login.test.tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import Login from './Login';
import * as api from '../services/api';

vi.mock('../services/api');

describe('Login Page', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form', () => {
    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );
    
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('handles successful login', async () => {
    const mockLogin = vi.spyOn(api.auth, 'login').mockResolvedValue({
      token: 'test-token',
      user: { id: 1, username: 'testuser', role: 'admin' },
    });

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    fireEvent.change(screen.getByLabelText(/username/i), {
      target: { value: 'testuser' },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: 'password123' },
    });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('testuser', 'password123');
    });
  });

  it('displays error on failed login', async () => {
    vi.spyOn(api.auth, 'login').mockRejectedValue(new Error('Invalid credentials'));

    render(
      <BrowserRouter>
        <Login />
      </BrowserRouter>
    );

    fireEvent.change(screen.getByLabelText(/username/i), {
      target: { value: 'wronguser' },
    });
    fireEvent.change(screen.getByLabelText(/password/i), {
      target: { value: 'wrongpass' },
    });
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument();
    });
  });
});
```

#### Example: API Service Tests

```typescript
// web/src/services/api.test.ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api } from './api';

global.fetch = vi.fn();

describe('API Service', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('fetches instances successfully', async () => {
    const mockInstances = [
      { id: 1, name: 'instance-1', status: 'running' },
      { id: 2, name: 'instance-2', status: 'stopped' },
    ];

    (global.fetch as any).mockResolvedValueOnce({
      ok: true,
      json: async () => mockInstances,
    });

    const instances = await api.instances.list();
    expect(instances).toEqual(mockInstances);
  });

  it('handles API errors', async () => {
    (global.fetch as any).mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
    });

    await expect(api.instances.list()).rejects.toThrow();
  });
});
```

### Running Frontend Tests

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run in watch mode
npm run test:watch

# Run specific test file
npm test -- Login.test.tsx

# Update snapshots
npm test -- -u
```

## Integration Testing

### Docker Compose Test Environment

```yaml
# docker-compose.test.yml
version: '3.8'

services:
  api-test:
    build:
      context: ./server
      dockerfile: Dockerfile
    environment:
      - DATABASE_PATH=/tmp/test.db
      - JWT_SECRET=test-secret
      - LOG_LEVEL=debug
    ports:
      - "8080:8080"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 5s
      timeout: 3s
      retries: 3

  web-test:
    build:
      context: ./web
      dockerfile: Dockerfile
    environment:
      - VITE_API_URL=http://api-test:8080
    ports:
      - "3000:80"
    depends_on:
      api-test:
        condition: service_healthy
```

### Integration Test Script

```bash
#!/bin/bash
# test-integration.sh

set -e

echo "Starting integration tests..."

# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Wait for services
echo "Waiting for services to be ready..."
sleep 10

# Run API tests
echo "Running API integration tests..."
curl -f http://localhost:8080/health || exit 1

# Run frontend tests
echo "Running frontend integration tests..."
curl -f http://localhost:3000 || exit 1

# Cleanup
docker-compose -f docker-compose.test.yml down

echo "Integration tests completed successfully!"
```

## End-to-End Testing

### Playwright Setup

```bash
cd web
npm install --save-dev @playwright/test
npx playwright install
```

### Playwright Configuration

```typescript
// web/playwright.config.ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
  ],
  webServer: {
    command: 'npm run dev',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
  },
});
```

### E2E Test Examples

```typescript
// web/e2e/login.spec.ts
import { test, expect } from '@playwright/test';

test.describe('Login Flow', () => {
  test('should login successfully', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('text=Dashboard')).toBeVisible();
  });

  test('should show error on invalid credentials', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('input[name="username"]', 'invalid');
    await page.fill('input[name="password"]', 'wrong');
    await page.click('button[type="submit"]');
    
    await expect(page.locator('text=Invalid credentials')).toBeVisible();
  });
});

// web/e2e/instances.spec.ts
test.describe('Instance Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login before each test
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'admin123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard');
  });

  test('should create new instance', async ({ page }) => {
    await page.goto('/instances');
    await page.click('text=Create Instance');
    
    await page.fill('input[name="name"]', 'test-instance');
    await page.fill('input[name="host"]', '10.0.0.1');
    await page.fill('input[name="port"]', '3993');
    await page.click('button[type="submit"]');
    
    await expect(page.locator('text=test-instance')).toBeVisible();
  });

  test('should start instance', async ({ page }) => {
    await page.goto('/instances');
    await page.click('text=test-instance');
    await page.click('button:has-text("Start")');
    
    await expect(page.locator('text=Running')).toBeVisible();
  });
});
```

## Performance Testing

### Load Testing with k6

```javascript
// load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '30s', target: 20 },
    { duration: '1m', target: 50 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};

export default function () {
  const res = http.get('http://localhost:8080/health');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

Run with:
```bash
k6 run load-test.js
```

## Security Testing

### OWASP ZAP Scan

```bash
# Run ZAP baseline scan
docker run -t owasp/zap2docker-stable zap-baseline.py \
  -t http://localhost:8080 \
  -r zap-report.html
```

### SQL Injection Tests

```go
func TestSQLInjectionPrevention(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Attempt SQL injection
    maliciousInput := "'; DROP TABLE instances; --"
    _, err := db.GetInstanceByName(maliciousInput)
    
    // Should not cause error or drop table
    assert.NoError(t, err)
    
    // Verify table still exists
    instances, err := db.ListInstances()
    assert.NoError(t, err)
    assert.NotNil(t, instances)
}
```

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run tests
        run: |
          cd server
          go test ./... -v -cover -coverprofile=coverage.out
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./server/coverage.out

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install dependencies
        run: |
          cd web
          npm ci
      - name: Run tests
        run: |
          cd web
          npm test -- --coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./web/coverage/coverage-final.json

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - name: Install Playwright
        run: |
          cd web
          npm ci
          npx playwright install --with-deps
      - name: Run E2E tests
        run: |
          cd web
          npm run test:e2e
      - uses: actions/upload-artifact@v3
        if: always()
        with:
          name: playwright-report
          path: web/playwright-report/
```

## Test Coverage Reports

### Generate Combined Report

```bash
#!/bin/bash
# generate-coverage.sh

echo "Generating backend coverage..."
cd server
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o ../coverage-backend.html

echo "Generating frontend coverage..."
cd ../web
npm run test:coverage
cp coverage/index.html ../coverage-frontend.html

echo "Coverage reports generated:"
echo "  - coverage-backend.html"
echo "  - coverage-frontend.html"
```

## Best Practices

1. **Write Tests First**: Follow TDD when possible
2. **Keep Tests Fast**: Unit tests should run in milliseconds
3. **Isolate Tests**: Each test should be independent
4. **Use Mocks**: Mock external dependencies
5. **Test Edge Cases**: Don't just test happy paths
6. **Maintain Tests**: Update tests when code changes
7. **Review Coverage**: Aim for high coverage but focus on critical paths
8. **Automate**: Run tests in CI/CD pipeline

#