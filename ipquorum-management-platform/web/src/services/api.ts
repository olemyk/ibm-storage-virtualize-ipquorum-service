import axios, { AxiosInstance, AxiosError } from 'axios';

// API Configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1';

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor - add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('auth_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor - handle errors
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      // Unauthorized - clear token and redirect to login
      localStorage.removeItem('auth_token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// API Error type
export interface ApiError {
  message: string;
  code?: string;
  details?: unknown;
}

// Generic API response wrapper
export interface ApiResponse<T> {
  data?: T;
  error?: ApiError;
  success: boolean;
}

// Auth API
export const authApi = {
  login: async (username: string, password: string) => {
    const response = await apiClient.post('/auth/login', { username, password });
    return response.data;
  },
  
  logout: async () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user');
  },
  
  getCurrentUser: () => {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  },
  
  isAuthenticated: () => {
    return !!localStorage.getItem('auth_token');
  },
};

// Instances API
export const instancesApi = {
  getAll: async () => {
    const response = await apiClient.get('/instances');
    return response.data;
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/instances/${id}`);
    return response.data;
  },
  
  create: async (data: {
    name: string;
    host: string;
    port: number;
    username: string;
    password: string;
    description?: string;
  }) => {
    const response = await apiClient.post('/instances', data);
    return response.data;
  },
  
  update: async (id: string, data: Partial<{
    name: string;
    host: string;
    port: number;
    username: string;
    password: string;
    description?: string;
  }>) => {
    const response = await apiClient.put(`/instances/${id}`, data);
    return response.data;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/instances/${id}`);
    return response.data;
  },
  
  start: async (id: string) => {
    const response = await apiClient.post(`/instances/${id}/start`);
    return response.data;
  },
  
  stop: async (id: string) => {
    const response = await apiClient.post(`/instances/${id}/stop`);
    return response.data;
  },
  
  restart: async (id: string) => {
    const response = await apiClient.post(`/instances/${id}/restart`);
    return response.data;
  },
  
  getHealth: async (id: string) => {
    const response = await apiClient.get(`/instances/${id}/health`);
    return response.data;
  },
  
  getLogs: async (id: string, lines: number = 100) => {
    const response = await apiClient.get(`/instances/${id}/logs`, {
      params: { lines },
    });
    return response.data;
  },
};

// Users API
export const usersApi = {
  getAll: async () => {
    const response = await apiClient.get('/users');
    return response.data;
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/users/${id}`);
    return response.data;
  },
  
  create: async (data: {
    username: string;
    password: string;
    email: string;
    role: string;
  }) => {
    const response = await apiClient.post('/users', data);
    return response.data;
  },
  
  update: async (id: string, data: Partial<{
    username: string;
    email: string;
    role: string;
  }>) => {
    const response = await apiClient.put(`/users/${id}`, data);
    return response.data;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/users/${id}`);
    return response.data;
  },
  
  changePassword: async (id: string, oldPassword: string, newPassword: string) => {
    const response = await apiClient.post(`/users/${id}/password`, {
      old_password: oldPassword,
      new_password: newPassword,
    });
    return response.data;
  },
};

// Health API
export const healthApi = {
  getSystemHealth: async () => {
    const response = await apiClient.get('/health');
    return response.data;
  },
};
// Servers API
export const serversApi = {
  getAll: async () => {
    const response = await apiClient.get('/servers');
    return response.data;
  },
  
  getById: async (id: string) => {
    const response = await apiClient.get(`/servers/${id}`);
    return response.data;
  },
  
  register: async (data: {
    hostname: string;
    ip_address: string;
    agent_port: number;
    api_key: string;
    tls_enabled: boolean;
    tls_verify: boolean;
  }) => {
    const response = await apiClient.post('/servers', data);
    return response.data;
  },
  
  delete: async (id: string) => {
    const response = await apiClient.delete(`/servers/${id}`);
    return response.data;
  },
  
  getInstances: async (id: string) => {
    const response = await apiClient.get(`/servers/${id}/instances`);
    return response.data;
  },
};


// Metrics API (Prometheus format)
export const metricsApi = {
  getMetrics: async () => {
    const response = await axios.get('/metrics');
    return response.data;
  },
};

// Unified API object for easier imports
export const api = {
  auth: authApi,
  instances: {
    list: instancesApi.getAll,
    get: instancesApi.getById,
    create: instancesApi.create,
    update: instancesApi.update,
    delete: instancesApi.delete,
    start: instancesApi.start,
    stop: instancesApi.stop,
    restart: instancesApi.restart,
    health: instancesApi.getHealth,
    logs: instancesApi.getLogs,
  },
  servers: {
    list: serversApi.getAll,
    get: serversApi.getById,
    register: serversApi.register,
    delete: serversApi.delete,
    getInstances: serversApi.getInstances,
  },
  users: {
    list: usersApi.getAll,
    get: usersApi.getById,
    create: usersApi.create,
    update: usersApi.update,
    delete: usersApi.delete,
    changePassword: usersApi.changePassword,
  },
  health: {
    check: healthApi.getSystemHealth,
  },
  metrics: metricsApi,
};

export default apiClient;


