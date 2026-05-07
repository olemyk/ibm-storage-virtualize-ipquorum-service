import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ArrowLeft,
  Play,
  Square,
  RotateCw,
  Edit,
  Trash2,
  RefreshCw,
  Activity,
  Server,
  Clock,
  MapPin,
  User,
  Settings as SettingsIcon,
  FileText,
  Download,
  HardDrive,
} from 'lucide-react';
import {
  Card,
  CardHeader,
  CardBody,
  Button,
  Badge,
  Alert,
  Spinner,
  Modal,
} from '../components/ui';
import { api } from '../services/api';
import { useToastStore } from '../stores/toastStore';
import type { Instance } from '../types/instance';

const InstanceDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastStore();
  
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [showLogs, setShowLogs] = useState(false);
  const [logLines, setLogLines] = useState(50);

  // Fetch instance details
  const {
    data: instance,
    isLoading,
    error,
    refetch,
  } = useQuery<Instance>({
    queryKey: ['instance', id],
    queryFn: () => api.instances.get(id!),
    refetchInterval: 10000, // Refresh every 10 seconds
    enabled: !!id,
  });

  // Fetch servers list to get server name
  const { data: servers } = useQuery({
    queryKey: ['servers'],
    queryFn: () => api.servers.list(),
  });

  // Get server name from server_id
  const getServerName = (serverId: string) => {
    if (serverId === 'local') {
      return 'Local (Manager Container)';
    }
    const server = servers?.find((s: any) => s.id === serverId);
    return server ? server.hostname : serverId;
  };

  // Fetch logs
  const {
    data: logs,
    isLoading: logsLoading,
    refetch: refetchLogs,
  } = useQuery<string>({
    queryKey: ['instance-logs', id, logLines],
    queryFn: () => api.instances.logs(id!, logLines),
    enabled: showLogs && !!id,
    refetchInterval: showLogs ? 5000 : false, // Auto-refresh every 5 seconds when visible
  });

  // Start mutation
  const startMutation = useMutation({
    mutationFn: () => api.instances.start(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instance', id] });
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      toast.success('Instance started successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to start instance');
    },
  });

  // Stop mutation
  const stopMutation = useMutation({
    mutationFn: () => api.instances.stop(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instance', id] });
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      toast.success('Instance stopped successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to stop instance');
    },
  });

  // Restart mutation
  const restartMutation = useMutation({
    mutationFn: () => api.instances.restart(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instance', id] });
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      toast.success('Instance restarted successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to restart instance');
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: () => api.instances.delete(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      toast.success('Instance deleted successfully');
      navigate('/instances');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete instance');
      setDeleteModalOpen(false);
    },
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return <Badge variant="success">Running</Badge>;
      case 'stopped':
        return <Badge variant="default">Stopped</Badge>;
      case 'starting':
        return <Badge variant="info">Starting</Badge>;
      case 'stopping':
        return <Badge variant="warning">Stopping</Badge>;
      case 'error':
        return <Badge variant="destructive">Error</Badge>;
      default:
        return <Badge variant="default">{status}</Badge>;
    }
  };

  const getHealthBadge = (health: string) => {
    switch (health) {
      case 'healthy':
        return <Badge variant="success">Healthy</Badge>;
      case 'unhealthy':
        return <Badge variant="destructive">Unhealthy</Badge>;
      case 'degraded':
        return <Badge variant="warning">Degraded</Badge>;
      default:
        return <Badge variant="default">{health}</Badge>;
    }
  };

  const formatDate = (dateString: string) => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleString();
  };

  const formatUptime = (seconds: number | undefined) => {
    if (!seconds || seconds === 0) return 'N/A';
    
    const years = Math.floor(seconds / (365 * 24 * 3600));
    const months = Math.floor((seconds % (365 * 24 * 3600)) / (30 * 24 * 3600));
    const days = Math.floor((seconds % (30 * 24 * 3600)) / (24 * 3600));
    const hours = Math.floor((seconds % (24 * 3600)) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    const parts: string[] = [];
    if (years > 0) parts.push(`${years}y`);
    if (months > 0) parts.push(`${months}mo`);
    if (days > 0) parts.push(`${days}d`);
    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    
    // If less than a minute, show seconds
    if (parts.length === 0) {
      return `${seconds}s`;
    }
    
    // Return first 2-3 most significant units
    return parts.slice(0, 3).join(' ');
  };

  const isPending =
    startMutation.isPending ||
    stopMutation.isPending ||
    restartMutation.isPending ||
    deleteMutation.isPending;

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error || !instance) {
    return (
      <div className="space-y-6">
        <Button
          variant="ghost"
          size="sm"
          leftIcon={<ArrowLeft className="h-4 w-4" />}
          onClick={() => navigate('/instances')}
        >
          Back to Instances
        </Button>
        <Alert variant="destructive" title="Error Loading Instance">
          {error instanceof Error ? error.message : 'Instance not found'}
        </Alert>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Button
            variant="ghost"
            size="sm"
            leftIcon={<ArrowLeft className="h-4 w-4" />}
            onClick={() => navigate('/instances')}
          >
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{instance.name}</h1>
            <p className="mt-1 text-sm text-gray-600">{instance.api_endpoint}</p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            leftIcon={<RefreshCw className="h-4 w-4" />}
            onClick={() => refetch()}
            disabled={isPending}
          >
            Refresh
          </Button>
          
          <Button
            variant="secondary"
            size="sm"
            leftIcon={<Edit className="h-4 w-4" />}
            onClick={() => navigate(`/instances/${id}/edit`)}
            disabled={isPending}
          >
            Edit
          </Button>

          <Button
            variant="default"
            size="sm"
            leftIcon={<Play className="h-4 w-4" />}
            onClick={() => startMutation.mutate()}
            isLoading={startMutation.isPending}
            disabled={isPending || instance.status === 'running'}
          >
            Start
          </Button>

          <Button
            variant="secondary"
            size="sm"
            leftIcon={<Square className="h-4 w-4" />}
            onClick={() => stopMutation.mutate()}
            isLoading={stopMutation.isPending}
            disabled={isPending || instance.status === 'stopped'}
          >
            Stop
          </Button>

          <Button
            variant="secondary"
            size="sm"
            leftIcon={<RotateCw className="h-4 w-4" />}
            onClick={() => restartMutation.mutate()}
            isLoading={restartMutation.isPending}
            disabled={isPending}
          >
            Restart
          </Button>

          <Button
            variant="destructive"
            size="sm"
            leftIcon={<Trash2 className="h-4 w-4" />}
            onClick={() => setDeleteModalOpen(true)}
            disabled={isPending}
          >
            Delete
          </Button>
        </div>
      </div>

      {/* Status Overview */}
      <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
        <Card variant="elevated">
          <CardBody>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 rounded-lg">
                <Activity className="h-6 w-6 text-blue-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600">Status</p>
                <div className="mt-1">{getStatusBadge(instance.status)}</div>
              </div>
            </div>
          </CardBody>
        </Card>

        <Card variant="elevated">
          <CardBody>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 rounded-lg">
                <Server className="h-6 w-6 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600">Health</p>
                <div className="mt-1">{getHealthBadge(instance.health)}</div>
              </div>
            </div>
          </CardBody>
        </Card>

        <Card variant="elevated">
          <CardBody>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-indigo-100 rounded-lg">
                <HardDrive className="h-6 w-6 text-indigo-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600">Running On</p>
                <p className="mt-1 text-sm font-medium text-gray-900 truncate" title={getServerName(instance.server_id)}>
                  {getServerName(instance.server_id)}
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        <Card variant="elevated">
          <CardBody>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-purple-100 rounded-lg">
                <Clock className="h-6 w-6 text-purple-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600">Uptime</p>
                <p className="mt-1 text-sm font-medium text-gray-900">
                  {formatUptime(instance.uptime)}
                </p>
              </div>
            </div>
          </CardBody>
        </Card>

        <Card variant="elevated">
          <CardBody>
            <div className="flex items-center gap-3">
              <div className="p-2 bg-orange-100 rounded-lg">
                <MapPin className="h-6 w-6 text-orange-600" />
              </div>
              <div>
                <p className="text-sm text-gray-600">Location</p>
                <p className="mt-1 text-sm font-medium text-gray-900">
                  {instance.location || 'Not set'}
                </p>
              </div>
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Configuration Details */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Basic Information */}
        <Card variant="elevated">
          <CardHeader title="Basic Information" />
          <CardBody>
            <dl className="space-y-4">
              <div>
                <dt className="text-sm font-medium text-gray-500">Instance Name</dt>
                <dd className="mt-1 text-sm text-gray-900">{instance.name}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Running On Server</dt>
                <dd className="mt-1 text-sm text-gray-900 flex items-center gap-2">
                  <HardDrive className="h-4 w-4 text-gray-400" />
                  {getServerName(instance.server_id)}
                  {instance.server_id === 'local' && (
                    <Badge variant="info" size="sm">Local</Badge>
                  )}
                  {instance.server_id !== 'local' && (
                    <Badge variant="success" size="sm">Remote</Badge>
                  )}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">API Endpoint</dt>
                <dd className="mt-1 text-sm text-gray-900">{instance.api_endpoint}</dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Storage System</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {instance.storage_system || 'Not specified'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Description</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {instance.description || 'No description'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Location</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {instance.location || 'Not specified'}
                </dd>
              </div>
            </dl>
          </CardBody>
        </Card>

        {/* Configuration */}
        <Card variant="elevated">
          <CardHeader title="Configuration" />
          <CardBody>
            <dl className="space-y-4">
              <div>
                <dt className="text-sm font-medium text-gray-500">Username</dt>
                <dd className="mt-1 text-sm text-gray-900 flex items-center gap-2">
                  <User className="h-4 w-4 text-gray-400" />
                  {instance.username}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Partner System</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {instance.partnersystem || 'Not configured'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">IPQuorum Application Name</dt>
                <dd className="mt-1 text-sm text-gray-900">
                  {instance.ipquorum_name || 'Auto-generated'}
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Download Enabled</dt>
                <dd className="mt-1">
                  <Badge variant={instance.enable_download ? 'success' : 'default'}>
                    {instance.enable_download ? 'Yes' : 'No'}
                  </Badge>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">MkQuorumApp Enabled</dt>
                <dd className="mt-1">
                  <Badge variant={instance.enable_mkquorumapp ? 'success' : 'default'}>
                    {instance.enable_mkquorumapp ? 'Yes' : 'No'}
                  </Badge>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">IPv6 Configuration</dt>
                <dd className="mt-1 space-y-1">
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-600">Local System:</span>
                    <Badge variant={instance.ip6 ? 'info' : 'default'} size="sm">
                      {instance.ip6 ? 'IPv6 Enabled' : 'IPv4'}
                    </Badge>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-gray-600">Partner System:</span>
                    <Badge variant={instance.partnerip6 ? 'info' : 'default'} size="sm">
                      {instance.partnerip6 ? 'IPv6 Enabled' : 'IPv4'}
                    </Badge>
                  </div>
                </dd>
              </div>
              <div>
                <dt className="text-sm font-medium text-gray-500">Metadata Collection</dt>
                <dd className="mt-1">
                  <Badge variant={instance.nometadata ? 'warning' : 'success'}>
                    {instance.nometadata ? 'Disabled' : 'Enabled'}
                  </Badge>
                </dd>
              </div>
            </dl>
          </CardBody>
        </Card>
      </div>

      {/* Timestamps */}
      <Card variant="elevated">
        <CardHeader title="Timeline" />
        <CardBody>
          <dl className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">Created</dt>
              <dd className="mt-1 text-sm text-gray-900">{formatDate(instance.created_at)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Last Updated</dt>
              <dd className="mt-1 text-sm text-gray-900">{formatDate(instance.updated_at)}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Last Health Check</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {formatDate(instance.last_health_check || '')}
              </dd>
            </div>
          </dl>
        </CardBody>
      </Card>

      {/* Logs Section */}
      <Card variant="elevated">
        <CardHeader
          title="Instance Logs"
          action={
            <div className="flex items-center gap-2">
              <select
                value={logLines}
                onChange={(e) => setLogLines(Number(e.target.value))}
                className="px-3 py-1 text-sm border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                disabled={logsLoading}
              >
                <option value={50}>Last 50 lines</option>
                <option value={100}>Last 100 lines</option>
                <option value={200}>Last 200 lines</option>
                <option value={500}>Last 500 lines</option>
              </select>
              <Button
                variant="ghost"
                size="sm"
                leftIcon={<RefreshCw className="h-4 w-4" />}
                onClick={() => refetchLogs()}
                disabled={logsLoading || !showLogs}
              >
                Refresh
              </Button>
              <Button
                variant={showLogs ? 'secondary' : 'default'}
                size="sm"
                leftIcon={<FileText className="h-4 w-4" />}
                onClick={() => setShowLogs(!showLogs)}
              >
                {showLogs ? 'Hide Logs' : 'Show Logs'}
              </Button>
            </div>
          }
        />
        {showLogs && (
          <CardBody>
            {logsLoading ? (
              <div className="flex items-center justify-center py-8">
                <Spinner />
                <span className="ml-2 text-sm text-gray-600">Loading logs...</span>
              </div>
            ) : logs ? (
              <div className="relative">
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-xs font-mono max-h-96 overflow-y-auto">
                  {logs}
                </pre>
                <Button
                  variant="ghost"
                  size="sm"
                  leftIcon={<Download className="h-4 w-4" />}
                  onClick={() => {
                    const blob = new Blob([logs], { type: 'text/plain' });
                    const url = URL.createObjectURL(blob);
                    const a = document.createElement('a');
                    a.href = url;
                    a.download = `${instance.name}-logs-${new Date().toISOString()}.txt`;
                    document.body.appendChild(a);
                    a.click();
                    document.body.removeChild(a);
                    URL.revokeObjectURL(url);
                  }}
                  className="absolute top-2 right-2"
                >
                  Download
                </Button>
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                <FileText className="h-12 w-12 mx-auto mb-2 text-gray-400" />
                <p>No logs available</p>
              </div>
            )}
          </CardBody>
        )}
      </Card>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Delete Instance"
      >
        <div className="space-y-4">
          <p className="text-sm text-gray-600">
            Are you sure you want to delete <strong>{instance.name}</strong>? This action cannot be
            undone.
          </p>
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              onClick={() => setDeleteModalOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => deleteMutation.mutate()}
              isLoading={deleteMutation.isPending}
            >
              Delete Instance
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default InstanceDetail;


