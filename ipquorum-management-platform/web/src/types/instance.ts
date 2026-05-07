export interface Instance {
  id: string
  name: string
  server_id: string
  api_endpoint: string
  username: string
  storage_system?: string
  description?: string
  location?: string
  status: InstanceStatus
  health: HealthStatus
  enable_download: boolean
  enable_mkquorumapp: boolean
  partnersystem?: string
  ipquorum_name?: string
  ip6: boolean
  partnerip6: boolean
  nometadata: boolean
  uptime?: number  // Changed from string to number (seconds)
  started_at?: string
  last_health_check?: string
  created_at: string
  updated_at: string
}

export type InstanceStatus = 'running' | 'stopped' | 'failed' | 'unknown'
export type HealthStatus = 'healthy' | 'unhealthy' | 'degraded' | 'unknown'

export interface CreateInstanceRequest {
  name: string
  server_id: string
  api_endpoint: string
  username: string
  password: string
  storage_system?: string
  description?: string
  location?: string
  enable_download?: boolean
  enable_mkquorumapp?: boolean
  partnersystem?: string
  ipquorum_name?: string
  ip6?: boolean
  partnerip6?: boolean
  nometadata?: boolean
}

export interface UpdateInstanceRequest {
  description?: string
  location?: string
  storage_system?: string
}

export interface InstanceStatusResponse {
  id: string
  name: string
  status: InstanceStatus
  health: HealthStatus
  updated: string
}

export interface InstanceHealthResponse {
  instance: {
    id: string
    name: string
    status: InstanceStatus
    health: HealthStatus
    updated: string
  }
  history: HealthHistoryEntry[]
}

export interface HealthHistoryEntry {
  timestamp: string
  status: InstanceStatus
  health: HealthStatus
}

export interface InstanceMetrics {
  instance_id: string
  metrics: {
    uptime_seconds: number
    health_checks_total: number
    health_checks_failed: number
    operations_total: number
    operations_failed: number
  }
  period: {
    from: string
    to: string
  }
}

export interface InstanceOperation {
  type: 'start' | 'stop' | 'restart' | 'delete'
  instance_id: string
  timestamp: string
  status: 'success' | 'failed'
  message?: string
}

export interface SystemHealth {
  status: 'healthy' | 'unhealthy' | 'degraded'
  database?: string
  timestamp?: string
  version?: string
}


