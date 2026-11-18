// Container Types for IPQuorum Management Platform
// Matches backend API models

export type ContainerStatus = 'created' | 'running' | 'paused' | 'stopped' | 'error';
export type ContainerRuntime = 'docker' | 'podman';
export type RestartPolicy = 'no' | 'always' | 'on-failure' | 'unless-stopped';

export interface PortMapping {
  host_port: number;
  container_port: number;
  protocol: string; // 'tcp' | 'udp'
}

export interface VolumeMount {
  source: string;
  destination: string;
  read_only: boolean;
}

export interface ContainerConfig {
  ports?: PortMapping[];
  volumes?: VolumeMount[];
  environment?: Record<string, string>;
  labels?: Record<string, string>;
}

export interface ContainerInstance {
  id: string;
  instance_id: string;
  agent_id: string;
  container_id: string;
  container_name: string;
  image_name: string;
  image_tag: string;
  runtime: ContainerRuntime;
  status: ContainerStatus;
  state?: string;
  restart_policy: RestartPolicy;
  restart_count: number;
  exit_code?: number;
  error?: string;
  created_at: string;
  started_at?: string;
  finished_at?: string;
  updated_at: string;
  config?: ContainerConfig;
}

export interface CreateContainerRequest {
  instance_id: string;
  agent_id: string;
  image_name: string;
  image_tag?: string;
  restart_policy?: RestartPolicy;
  config?: ContainerConfig;
}

export interface ContainerListResponse {
  containers: ContainerInstance[];
  total: number;
}

export interface ContainerLogsOptions {
  tail?: number;
  follow?: boolean;
  timestamps?: boolean;
}

export interface ContainerLogsResponse {
  logs: string;
}

export interface ContainerStats {
  cpu_percent: number;
  memory_usage: number;
  memory_limit: number;
  memory_percent: number;
  network_rx_bytes: number;
  network_tx_bytes: number;
  block_read: number;
  block_write: number;
  pids: number;
  timestamp: string;
}

export interface ContainerStatsResponse {
  stats: ContainerStats;
}

// Filter and sort options for container list
export interface ContainerFilters {
  status?: ContainerStatus;
  runtime?: ContainerRuntime;
  agent_id?: string;
  instance_id?: string;
  search?: string;
}

export interface ContainerSortOptions {
  field: 'created_at' | 'updated_at' | 'status' | 'container_name';
  order: 'asc' | 'desc';
}

// UI State types
export interface ContainerState {
  containers: ContainerInstance[];
  selectedContainer: ContainerInstance | null;
  loading: boolean;
  error: string | null;
  filters: ContainerFilters;
  sortOptions: ContainerSortOptions;
  total: number;
}

// Action types for Redux
export interface ContainerAction {
  type: string;
  payload?: any;
}

// Made with Bob
