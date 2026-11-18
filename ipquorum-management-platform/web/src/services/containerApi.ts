// Container API Client
// Handles all HTTP requests to the container management endpoints

import {
  ContainerInstance,
  CreateContainerRequest,
  ContainerListResponse,
  ContainerLogsOptions,
  ContainerLogsResponse,
  ContainerStatsResponse,
} from '../types/container';
import apiClient from './api';

const CONTAINER_BASE_URL = '/containers';

/**
 * Container API Service
 * Provides methods to interact with container management endpoints
 */
export const containerApi = {
  /**
   * List all containers
   * @returns Promise with list of containers
   */
  async listContainers(): Promise<ContainerListResponse> {
    const response = await apiClient.get<ContainerListResponse>(CONTAINER_BASE_URL);
    return response.data;
  },

  /**
   * Get container by ID
   * @param id - Container ID
   * @returns Promise with container details
   */
  async getContainer(id: string): Promise<ContainerInstance> {
    const response = await apiClient.get<ContainerInstance>(`${CONTAINER_BASE_URL}/${id}`);
    return response.data;
  },

  /**
   * Get container by instance ID
   * @param instanceId - Instance ID
   * @returns Promise with container details
   */
  async getContainerByInstanceId(instanceId: string): Promise<ContainerInstance> {
    const response = await apiClient.get<ContainerInstance>(
      `${CONTAINER_BASE_URL}/instance/${instanceId}`
    );
    return response.data;
  },

  /**
   * Create a new container
   * @param data - Container creation request
   * @returns Promise with created container
   */
  async createContainer(data: CreateContainerRequest): Promise<ContainerInstance> {
    const response = await apiClient.post<ContainerInstance>(CONTAINER_BASE_URL, data);
    return response.data;
  },

  /**
   * Start a container
   * @param id - Container ID
   * @returns Promise with success message
   */
  async startContainer(id: string): Promise<{ message: string }> {
    const response = await apiClient.post<{ message: string }>(
      `${CONTAINER_BASE_URL}/${id}/start`
    );
    return response.data;
  },

  /**
   * Stop a container
   * @param id - Container ID
   * @param timeout - Optional timeout in seconds
   * @returns Promise with success message
   */
  async stopContainer(id: string, timeout?: number): Promise<{ message: string }> {
    const response = await apiClient.post<{ message: string }>(
      `${CONTAINER_BASE_URL}/${id}/stop`,
      { timeout }
    );
    return response.data;
  },

  /**
   * Restart a container
   * @param id - Container ID
   * @param timeout - Optional timeout in seconds
   * @returns Promise with success message
   */
  async restartContainer(id: string, timeout?: number): Promise<{ message: string }> {
    const response = await apiClient.post<{ message: string }>(
      `${CONTAINER_BASE_URL}/${id}/restart`,
      { timeout }
    );
    return response.data;
  },

  /**
   * Get container logs
   * @param id - Container ID
   * @param options - Log options (tail, follow, timestamps)
   * @returns Promise with container logs
   */
  async getContainerLogs(
    id: string,
    options?: ContainerLogsOptions
  ): Promise<ContainerLogsResponse> {
    const params = new URLSearchParams();
    if (options?.tail) params.append('tail', options.tail.toString());
    if (options?.follow) params.append('follow', options.follow.toString());
    if (options?.timestamps) params.append('timestamps', options.timestamps.toString());

    const response = await apiClient.get<ContainerLogsResponse>(
      `${CONTAINER_BASE_URL}/${id}/logs?${params.toString()}`
    );
    return response.data;
  },

  /**
   * Get container stats
   * @param id - Container ID
   * @returns Promise with container statistics
   */
  async getContainerStats(id: string): Promise<ContainerStatsResponse> {
    const response = await apiClient.get<ContainerStatsResponse>(
      `${CONTAINER_BASE_URL}/${id}/stats`
    );
    return response.data;
  },

  /**
   * Delete a container
   * @param id - Container ID
   * @param force - Force delete (skip agent cleanup, database only)
   * @returns Promise with success message
   */
  async deleteContainer(id: string, force: boolean = false): Promise<{ message: string }> {
    const url = force
      ? `${CONTAINER_BASE_URL}/${id}?force=true`
      : `${CONTAINER_BASE_URL}/${id}`;
    const response = await apiClient.delete<{ message: string }>(url);
    return response.data;
  },
};

/**
 * Helper function to format container status for display
 */
export const formatContainerStatus = (status: string): string => {
  const statusMap: Record<string, string> = {
    created: 'Created',
    running: 'Running',
    paused: 'Paused',
    stopped: 'Stopped',
    error: 'Error',
  };
  return statusMap[status] || status;
};

/**
 * Helper function to get status color
 */
export const getStatusColor = (status: string): string => {
  const colorMap: Record<string, string> = {
    created: 'blue',
    running: 'green',
    paused: 'yellow',
    stopped: 'gray',
    error: 'red',
  };
  return colorMap[status] || 'gray';
};

/**
 * Helper function to format bytes to human-readable format
 */
export const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
};

/**
 * Helper function to format percentage
 */
export const formatPercent = (value: number): string => {
  return `${value.toFixed(2)}%`;
};

/**
 * Helper function to format uptime
 */
export const formatUptime = (startedAt?: string): string => {
  if (!startedAt) return 'N/A';
  
  const start = new Date(startedAt);
  const now = new Date();
  const diff = now.getTime() - start.getTime();
  
  const seconds = Math.floor(diff / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);
  
  if (days > 0) return `${days}d ${hours % 24}h`;
  if (hours > 0) return `${hours}h ${minutes % 60}m`;
  if (minutes > 0) return `${minutes}m ${seconds % 60}s`;
  return `${seconds}s`;
};

// Made with Bob
