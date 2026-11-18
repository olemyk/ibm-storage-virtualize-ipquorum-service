import axios from 'axios';
import type {
  IPQuorumContainer,
  CreateIPQuorumContainerRequest,
  ContainerLogs,
  ContainerStats,
  IPQuorumConfig,
} from '../types/ipquorum';

const API_BASE_URL = import.meta.env.VITE_API_URL || '';

// Create axios instance with default config
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('auth_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// IPQuorum Container API
export const ipquorumApi = {
  /**
   * Create a new IPQuorum container
   */
  createContainer: async (
    request: CreateIPQuorumContainerRequest
  ): Promise<IPQuorumContainer> => {
    const response = await apiClient.post<IPQuorumContainer>(
      '/api/v1/ipquorum/containers',
      request
    );
    return response.data;
  },

  /**
   * List all IPQuorum containers
   * @param agentId Optional agent ID filter
   */
  listContainers: async (agentId?: string): Promise<IPQuorumContainer[]> => {
    const params = agentId ? { agent_id: agentId } : {};
    const response = await apiClient.get<IPQuorumContainer[]>(
      '/api/v1/ipquorum/containers',
      { params }
    );
    return response.data;
  },

  /**
   * Get a specific IPQuorum container by ID
   */
  getContainer: async (containerId: string): Promise<IPQuorumContainer> => {
    const response = await apiClient.get<IPQuorumContainer>(
      `/api/v1/ipquorum/containers/${containerId}`
    );
    return response.data;
  },

  /**
   * Delete an IPQuorum container
   */
  deleteContainer: async (containerId: string): Promise<void> => {
    await apiClient.delete(`/api/v1/ipquorum/containers/${containerId}`);
  },

  /**
   * Start an IPQuorum container
   */
  startContainer: async (containerId: string): Promise<{ status: string }> => {
    const response = await apiClient.post<{ status: string }>(
      `/api/v1/ipquorum/containers/${containerId}/start`
    );
    return response.data;
  },

  /**
   * Stop an IPQuorum container
   */
  stopContainer: async (containerId: string): Promise<{ status: string }> => {
    const response = await apiClient.post<{ status: string }>(
      `/api/v1/ipquorum/containers/${containerId}/stop`
    );
    return response.data;
  },

  /**
   * Restart an IPQuorum container
   */
  restartContainer: async (containerId: string): Promise<{ status: string }> => {
    const response = await apiClient.post<{ status: string }>(
      `/api/v1/ipquorum/containers/${containerId}/restart`
    );
    return response.data;
  },

  /**
   * Get container logs
   * @param containerId Container ID
   * @param tail Number of lines to retrieve (default: 100)
   */
  getContainerLogs: async (
    containerId: string,
    tail: number = 100
  ): Promise<ContainerLogs> => {
    const response = await apiClient.get<ContainerLogs>(
      `/api/v1/ipquorum/containers/${containerId}/logs`,
      { params: { tail } }
    );
    return response.data;
  },

  /**
   * Get container stats
   */
  getContainerStats: async (containerId: string): Promise<ContainerStats> => {
    const response = await apiClient.get<ContainerStats>(
      `/api/v1/ipquorum/containers/${containerId}/stats`
    );
    return response.data;
  },

  /**
   * Export IPQuorum configuration as JSON
   */
  exportConfig: async (containerId: string): Promise<IPQuorumConfig> => {
    const response = await apiClient.get<IPQuorumConfig>(
      `/api/v1/ipquorum/containers/${containerId}/config/export`
    );
    return response.data;
  },

  /**
   * Download IPQuorum configuration as file
   */
  downloadConfig: async (containerId: string, filename?: string): Promise<void> => {
    const response = await apiClient.get(
      `/api/v1/ipquorum/containers/${containerId}/config/export`,
      { responseType: 'blob' }
    );

    // Create download link
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', filename || 'ipquorum-config.json');
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  },
};

// Helper functions for common operations
export const ipquorumHelpers = {
  /**
   * Format bytes to human-readable string
   */
  formatBytes: (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  },

  /**
   * Format uptime in seconds to human-readable string
   */
  formatUptime: (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    const parts: string[] = [];
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);

    return parts.length > 0 ? parts.join(' ') : '< 1m';
  },

  /**
   * Format timestamp to local date/time string
   */
  formatTimestamp: (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  },

  /**
   * Get status badge color class
   */
  getStatusColor: (status: string): string => {
    const colors: Record<string, string> = {
      created: 'bg-blue-100 text-blue-800',
      running: 'bg-green-100 text-green-800',
      paused: 'bg-yellow-100 text-yellow-800',
      stopped: 'bg-gray-100 text-gray-800',
      error: 'bg-red-100 text-red-800',
    };
    return colors[status.toLowerCase()] || 'bg-gray-100 text-gray-800';
  },

  /**
   * Get health badge color class
   */
  getHealthColor: (health: string): string => {
    const colors: Record<string, string> = {
      healthy: 'bg-green-100 text-green-800',
      degraded: 'bg-yellow-100 text-yellow-800',
      unhealthy: 'bg-red-100 text-red-800',
      unknown: 'bg-gray-100 text-gray-800',
    };
    return colors[health.toLowerCase()] || 'bg-gray-100 text-gray-800';
  },

  /**
   * Validate IPQuorum configuration
   */
  validateConfig: (config: Partial<IPQuorumConfig>): string[] => {
    const errors: string[] = [];

    if (!config.api_endpoint) {
      errors.push('API endpoint is required');
    }

    if (!config.username) {
      errors.push('Username is required');
    }

    if (!config.password) {
      errors.push('Password is required');
    }

    if (config.create_quorum && !config.partner_system) {
      errors.push('Partner system is required when creating quorum app');
    }

    if (config.port && (config.port < 1024 || config.port > 65535)) {
      errors.push('Port must be between 1024 and 65535');
    }

    return errors;
  },
};

export default ipquorumApi;

// Made with Bob
