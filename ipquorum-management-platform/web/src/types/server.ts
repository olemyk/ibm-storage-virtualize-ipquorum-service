export interface Server {
  id: string;
  hostname: string;
  ip_address: string;
  agent_port: number;
  tls_enabled: boolean;
  tls_verify: boolean;
  connection_status: 'online' | 'offline' | 'unknown';
  status: 'online' | 'offline' | 'error';
  version?: string;
  last_seen?: string;
  created_at: string;
  updated_at: string;
}

export interface ServerFormData {
  hostname: string;
  ip_address: string;
  agent_port: number;
  api_key: string;
  tls_enabled: boolean;
  tls_verify: boolean;
}

export interface ServerWithInstances extends Server {
  instances: any[];
}


