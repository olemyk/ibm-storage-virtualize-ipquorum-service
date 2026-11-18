import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const authDuration = new Trend('auth_duration');
const containerListDuration = new Trend('container_list_duration');
const containerCreateDuration = new Trend('container_create_duration');
const requestCounter = new Counter('requests_total');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },  // Ramp up to 10 users
    { duration: '1m', target: 50 },   // Ramp up to 50 users
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '2m', target: 100 },  // Stay at 100 users
    { duration: '1m', target: 50 },   // Ramp down to 50 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% of requests should be below 500ms
    http_req_failed: ['rate<0.05'],                  // Error rate should be below 5%
    errors: ['rate<0.05'],
  },
};

// Configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const USERNAME = __ENV.USERNAME || 'admin';
const PASSWORD = __ENV.PASSWORD || 'admin123';

// Get authentication token
function getAuthToken() {
  const loginPayload = JSON.stringify({
    username: USERNAME,
    password: PASSWORD,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };

  const startTime = new Date();
  const response = http.post(`${BASE_URL}/api/v1/auth/login`, loginPayload, params);
  authDuration.add(new Date() - startTime);
  requestCounter.add(1);

  const success = check(response, {
    'login status is 200': (r) => r.status === 200,
    'login has token': (r) => r.json('token') !== undefined,
  });

  if (!success) {
    errorRate.add(1);
    console.error(`Login failed: ${response.status} ${response.body}`);
    return null;
  }

  errorRate.add(0);
  return response.json('token');
}

// Test scenarios
export default function () {
  // Get auth token
  const token = getAuthToken();
  if (!token) {
    return;
  }

  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  // Scenario 1: List containers (most common operation)
  const listStartTime = new Date();
  const listResponse = http.get(`${BASE_URL}/api/v1/containers`, params);
  containerListDuration.add(new Date() - listStartTime);
  requestCounter.add(1);

  const listSuccess = check(listResponse, {
    'list containers status is 200': (r) => r.status === 200,
    'list containers returns array': (r) => Array.isArray(r.json('containers')),
  });

  errorRate.add(listSuccess ? 0 : 1);

  sleep(1);

  // Scenario 2: Get runtime info
  const infoResponse = http.get(`${BASE_URL}/api/v1/runtime/info`, params);
  requestCounter.add(1);

  const infoSuccess = check(infoResponse, {
    'runtime info status is 200': (r) => r.status === 200,
    'runtime info has type': (r) => r.json('type') !== undefined,
  });

  errorRate.add(infoSuccess ? 0 : 1);

  sleep(1);

  // Scenario 3: Get runtime version
  const versionResponse = http.get(`${BASE_URL}/api/v1/runtime/version`, params);
  requestCounter.add(1);

  const versionSuccess = check(versionResponse, {
    'version status is 200': (r) => r.status === 200,
    'version has version field': (r) => r.json('version') !== undefined,
  });

  errorRate.add(versionSuccess ? 0 : 1);

  sleep(1);

  // Scenario 4: List images
  const imagesResponse = http.get(`${BASE_URL}/api/v1/images`, params);
  requestCounter.add(1);

  const imagesSuccess = check(imagesResponse, {
    'list images status is 200': (r) => r.status === 200,
  });

  errorRate.add(imagesSuccess ? 0 : 1);

  sleep(2);

  // Scenario 5: Create container (less frequent, more resource-intensive)
  if (Math.random() < 0.1) { // Only 10% of iterations create containers
    const containerName = `load-test-${Date.now()}-${__VU}-${__ITER}`;
    const createPayload = JSON.stringify({
      name: containerName,
      image: 'alpine:latest',
      cmd: ['sleep', '300'],
      labels: {
        'test': 'k6-load-test',
        'vu': `${__VU}`,
        'iter': `${__ITER}`,
      },
    });

    const createStartTime = new Date();
    const createResponse = http.post(`${BASE_URL}/api/v1/containers`, createPayload, params);
    containerCreateDuration.add(new Date() - createStartTime);
    requestCounter.add(1);

    const createSuccess = check(createResponse, {
      'create container status is 201': (r) => r.status === 201,
      'create container returns id': (r) => r.json('id') !== undefined,
    });

    errorRate.add(createSuccess ? 0 : 1);

    if (createSuccess) {
      const containerId = createResponse.json('id');

      // Start the container
      const startResponse = http.post(`${BASE_URL}/api/v1/containers/${containerId}/start`, null, params);
      requestCounter.add(1);

      const startSuccess = check(startResponse, {
        'start container status is 200': (r) => r.status === 200,
      });

      errorRate.add(startSuccess ? 0 : 1);

      sleep(2);

      // Stop the container
      const stopResponse = http.post(`${BASE_URL}/api/v1/containers/${containerId}/stop`, null, params);
      requestCounter.add(1);

      const stopSuccess = check(stopResponse, {
        'stop container status is 200': (r) => r.status === 200,
      });

      errorRate.add(stopSuccess ? 0 : 1);

      sleep(1);

      // Remove the container
      const removeResponse = http.del(`${BASE_URL}/api/v1/containers/${containerId}?force=true`, null, params);
      requestCounter.add(1);

      const removeSuccess = check(removeResponse, {
        'remove container status is 200': (r) => r.status === 200,
      });

      errorRate.add(removeSuccess ? 0 : 1);
    }
  }

  sleep(1);
}

// Teardown function to clean up test containers
export function teardown(data) {
  const token = getAuthToken();
  if (!token) {
    return;
  }

  const params = {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  };

  // List all containers
  const listResponse = http.get(`${BASE_URL}/api/v1/containers?all=true`, params);
  if (listResponse.status !== 200) {
    return;
  }

  const containers = listResponse.json('containers') || [];

  // Remove test containers
  containers.forEach((container) => {
    if (container.labels && container.labels.test === 'k6-load-test') {
      console.log(`Cleaning up test container: ${container.id}`);
      http.del(`${BASE_URL}/api/v1/containers/${container.id}?force=true`, null, params);
    }
  });
}

// Made with Bob
