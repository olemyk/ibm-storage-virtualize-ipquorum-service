// IPQuorum Container Types

export interface IPQuorumConfig {
  api_endpoint: string;
  username: string;
  password: string;
  partner_system: string;
  ip6: boolean;
  partnerip6: boolean;
  nometadata: boolean;
  insecure: boolean;
  download_jar: boolean;
  create_quorum: boolean;
  port: number;
  java_opts?: string;
  force_download?: boolean;
}

export interface CreateIPQuorumContainerRequest {
  name: string;
  agent_id: string;
  ipquorum_config: IPQuorumConfig;
  restart_policy?: string;
  image_name?: string;
  image_tag?: string;
}

export interface IPQuorumContainer {
  container: ContainerInstance;
  instance?: Instance;
  ipquorum_config: IPQuorumConfig;
}

export interface ContainerInstance {
  id: string;
  instance_id: string;
  agent_id: string;
  container_id: string;
  container_name: string;
  image_name: string;
  image_tag: string;
  runtime: string;
  status: string;
  state: string;
  restart_policy: string;
  restart_count: number;
  exit_code?: number;
  error?: string;
  created_at: string;
  started_at?: string;
  finished_at?: string;
  updated_at: string;
}

export interface Instance {
  id: string;
  name: string;
  server_id: string;
  deployment_type: 'systemd' | 'container';
  api_endpoint: string;
  username: string;
  storage_system: string;
  description?: string;
  location?: string;
  status: string;
  health: string;
  started_at?: string;
  uptime?: number;
  last_health_check?: string;
  enable_download: boolean;
  enable_mkquorumapp: boolean;
  partnersystem: string;
  ipquorum_name: string;
  ip6: boolean;
  partnerip6: boolean;
  nometadata: boolean;
  created_at: string;
  updated_at: string;
}

export interface ContainerLogs {
  container_id: string;
  logs: string[];
  timestamp: string;
}

export interface ContainerStats {
  container_id: string;
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

// Form validation types
export interface IPQuorumFormErrors {
  name?: string;
  agent_id?: string;
  api_endpoint?: string;
  username?: string;
  password?: string;
  partner_system?: string;
  port?: string;
}

// Default values
export const DEFAULT_IPQUORUM_CONFIG: Partial<IPQuorumConfig> = {
  ip6: false,
  partnerip6: false,
  nometadata: false,
  insecure: true,
  download_jar: true,
  create_quorum: false,
  port: 1260,
  java_opts: '-Xmx512m -Xms256m',
  force_download: false,
};

export const DEFAULT_RESTART_POLICY = 'unless-stopped';
export const DEFAULT_IMAGE_NAME = 'localhost/ipquorum-runner';
export const DEFAULT_IMAGE_TAG = 'latest';

// Status badge colors
export const CONTAINER_STATUS_COLORS: Record<string, string> = {
  created: 'bg-blue-100 text-blue-800',
  running: 'bg-green-100 text-green-800',
  paused: 'bg-yellow-100 text-yellow-800',
  stopped: 'bg-gray-100 text-gray-800',
  error: 'bg-red-100 text-red-800',
};

// Health status colors
export const HEALTH_STATUS_COLORS: Record<string, string> = {
  healthy: 'bg-green-100 text-green-800',
  degraded: 'bg-yellow-100 text-yellow-800',
  unhealthy: 'bg-red-100 text-red-800',
  unknown: 'bg-gray-100 text-gray-800',
};

// Made with Bob
