import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ArrowLeft,
  Edit,
  Save,
  X,
  Trash2,
  CheckCircle,
  XCircle,
  AlertCircle,
  Server as ServerIcon,
  Activity,
  Clock,
  HardDrive,
  FileText,
} from 'lucide-react';
import {
  Card,
  CardHeader,
  CardBody,
  Button,
  Badge,
  Alert,
  Spinner,
  Input,
  Modal,
  Table,
} from '../components/ui';
import { LogsViewer } from '../components/LogsViewer';
import { api } from '../services/api';
import type { Server } from '../types/server';
import type { Column } from '../components/ui/Table';

const ServerDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  
  const [isEditing, setIsEditing] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [logsModalOpen, setLogsModalOpen] = useState(false);
  const [selectedContainer, setSelectedContainer] = useState<{ id: string; name: string } | null>(null);
  const [actionError, setActionError] = useState('');
  const [testResult, setTestResult] = useState<{ status: string; error?: string } | null>(null);
  
  const [formData, setFormData] = useState({
    hostname: '',
    ip_address: '',
    agent_port: 8081,
    api_key: '',
    tls_enabled: true,
    tls_verify: false,
  });

  // Fetch server details
  const {
    data: server,
    isLoading,
    error,
  } = useQuery<Server>({
    queryKey: ['servers', id],
    queryFn: () => api.servers.get(id!),
    enabled: !!id,
    refetchInterval: 10000, // Refresh every 10 seconds
  });

  // Fetch server status (includes containers)
  const {
    data: serverStatus,
    isLoading: statusLoading,
  } = useQuery({
    queryKey: ['servers', id, 'status'],
    queryFn: () => api.servers.getStatus(id!),
    enabled: !!id,
    refetchInterval: 10000,
  });

  // Update server mutation
  const updateMutation = useMutation({
    mutationFn: (data: typeof formData) => api.servers.update(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers', id] });
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setIsEditing(false);
      setActionError('');
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to update server');
    },
  });

  // Delete server mutation
  const deleteMutation = useMutation({
    mutationFn: () => api.servers.delete(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      navigate('/servers');
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to delete server');
    },
  });

  // Test connection mutation
  const testMutation = useMutation({
    mutationFn: () => api.servers.test(id!),
    onSuccess: (data) => {
      setTestResult(data);
    },
    onError: (error: any) => {
      setTestResult({
        status: 'unhealthy',
        error: error.response?.data?.error || 'Connection test failed',
      });
    },
  });

  // Initialize form data when server loads
  React.useEffect(() => {
    if (server && !isEditing) {
      setFormData({
        hostname: server.hostname,
        ip_address: server.ip_address,
        agent_port: server.agent_port,
        api_key: '',
        tls_enabled: true,
        tls_verify: false,
      });
    }
  }, [server, isEditing]);

  const handleEdit = () => {
    setIsEditing(true);
    setActionError('');
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    setActionError('');
    if (server) {
      setFormData({
        hostname: server.hostname,
        ip_address: server.ip_address,
        agent_port: server.agent_port,
        api_key: '',
        tls_enabled: true,
        tls_verify: false,
      });
    }
  };

  const handleSave = () => {
    updateMutation.mutate(formData);
  };

  const handleDelete = () => {
    deleteMutation.mutate();
  };

  const handleTestConnection = () => {
    setTestResult(null);
    setTestModalOpen(true);
    testMutation.mutate();
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'online':
        return (
          <Badge variant="success">
            <CheckCircle className="w-3 h-3 mr-1" />
            Online
          </Badge>
        );
      case 'offline':
        return (
          <Badge variant="danger">
            <XCircle className="w-3 h-3 mr-1" />
            Offline
          </Badge>
        );
      default:
        return (
          <Badge variant="warning">
            <AlertCircle className="w-3 h-3 mr-1" />
            Unknown
          </Badge>
        );
    }
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return 'Never';
    return new Date(dateString).toLocaleString();
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error || !server) {
    return (
      <div className="space-y-4">
        <Button variant="ghost" onClick={() => navigate('/servers')}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          Back to Servers
        </Button>
        <Alert variant="danger">
          Failed to load server details. Server may not exist.
        </Alert>
      </div>
    );
  }

  // Handler for viewing container logs
  const handleViewLogs = (container: any) => {
    setSelectedContainer({
      id: container.id,
      name: container.name || container.id?.substring(0, 12) || 'Unknown',
    });
    setLogsModalOpen(true);
  };

  // Container columns for table
  const containerColumns: Column<any>[] = [
    {
      key: 'name',
      header: 'Container Name',
      render: (container) => (
        <div className="font-medium">{container.name || container.id?.substring(0, 12)}</div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      render: (container) => {
        const status = container.status?.toLowerCase() || 'unknown';
        if (status.includes('running')) {
          return <Badge variant="success">Running</Badge>;
        } else if (status.includes('exited')) {
          return <Badge variant="danger">Stopped</Badge>;
        }
        return <Badge variant="warning">{status}</Badge>;
      },
    },
    {
      key: 'image',
      header: 'Image',
      render: (container) => (
        <span className="text-sm text-gray-600">{container.image || 'N/A'}</span>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (container) => (
        <Button
          variant="outline"
          size="sm"
          onClick={() => handleViewLogs(container)}
        >
          <FileText className="w-4 h-4 mr-1" />
          Logs
        </Button>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Button variant="ghost" onClick={() => navigate('/servers')}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div>
            <h1 className="text-2xl font-bold text-gray-900 flex items-center">
              <ServerIcon className="w-6 h-6 mr-2" />
              {server.hostname}
            </h1>
            <p className="text-sm text-gray-500 mt-1">Server ID: {server.id}</p>
          </div>
        </div>
        <div className="flex items-center space-x-2">
          {!isEditing ? (
            <>
              <Button variant="outline" onClick={handleTestConnection}>
                <Activity className="w-4 h-4 mr-2" />
                Test Connection
              </Button>
              <Button variant="default" onClick={handleEdit}>
                <Edit className="w-4 h-4 mr-2" />
                Edit
              </Button>
              <Button
                variant="destructive"
                onClick={() => setDeleteModalOpen(true)}
              >
                <Trash2 className="w-4 h-4 mr-2" />
                Delete
              </Button>
            </>
          ) : (
            <>
              <Button variant="outline" onClick={handleCancelEdit}>
                <X className="w-4 h-4 mr-2" />
                Cancel
              </Button>
              <Button
                variant="default"
                onClick={handleSave}
                disabled={updateMutation.isPending}
              >
                <Save className="w-4 h-4 mr-2" />
                {updateMutation.isPending ? 'Saving...' : 'Save Changes'}
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Error Alert */}
      {actionError && (
        <Alert variant="danger">
          {actionError}
        </Alert>
      )}

      {/* Server Details Card */}
      <Card>
        <CardHeader>
          <h2 className="text-lg font-semibold">Server Configuration</h2>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {/* Status */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Status
              </label>
              <div>{getStatusBadge(server.status)}</div>
            </div>

            {/* Last Seen */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                <Clock className="w-4 h-4 inline mr-1" />
                Last Seen
              </label>
              <div className="text-sm text-gray-900">
                {formatDate(server.last_seen)}
              </div>
            </div>

            {/* Hostname */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Hostname
              </label>
              {isEditing ? (
                <Input
                  value={formData.hostname}
                  onChange={(e) =>
                    setFormData({ ...formData, hostname: e.target.value })
                  }
                  placeholder="server-hostname"
                />
              ) : (
                <div className="text-sm text-gray-900">{server.hostname}</div>
              )}
            </div>

            {/* IP Address */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                IP Address
              </label>
              {isEditing ? (
                <Input
                  value={formData.ip_address}
                  onChange={(e) =>
                    setFormData({ ...formData, ip_address: e.target.value })
                  }
                  placeholder="10.0.0.1"
                />
              ) : (
                <div className="text-sm text-gray-900">{server.ip_address}</div>
              )}
            </div>

            {/* Agent Port */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Agent Port
              </label>
              {isEditing ? (
                <Input
                  type="number"
                  value={formData.agent_port}
                  onChange={(e) =>
                    setFormData({
                      ...formData,
                      agent_port: parseInt(e.target.value) || 8081,
                    })
                  }
                />
              ) : (
                <div className="text-sm text-gray-900">{server.agent_port}</div>
              )}
            </div>

            {/* Version */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Agent Version
              </label>
              <div className="text-sm text-gray-900">
                {server.version || 'Unknown'}
              </div>
            </div>

            {/* Created At */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Created At
              </label>
              <div className="text-sm text-gray-900">
                {formatDate(server.created_at)}
              </div>
            </div>

            {/* API Key (only in edit mode) */}
            {isEditing && (
              <>
                <div className="md:col-span-2">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    API Key (leave empty to keep current)
                  </label>
                  <Input
                    type="password"
                    value={formData.api_key}
                    onChange={(e) =>
                      setFormData({ ...formData, api_key: e.target.value })
                    }
                    placeholder="Enter new API key (optional)"
                  />
                </div>

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={formData.tls_enabled}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          tls_enabled: e.target.checked,
                        })
                      }
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm font-medium text-gray-700">
                      TLS Enabled
                    </span>
                  </label>
                </div>

                <div>
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={formData.tls_verify}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          tls_verify: e.target.checked,
                        })
                      }
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm font-medium text-gray-700">
                      Verify TLS Certificate
                    </span>
                  </label>
                </div>
              </>
            )}
          </div>
        </CardBody>
      </Card>

      {/* Containers Card */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold flex items-center">
              <HardDrive className="w-5 h-5 mr-2" />
              Containers on this Server
            </h2>
            {serverStatus?.instances && (
              <Badge variant="info">
                {serverStatus.instances.length} container(s)
              </Badge>
            )}
          </div>
        </CardHeader>
        <CardBody>
          {statusLoading ? (
            <div className="flex justify-center py-8">
              <Spinner />
            </div>
          ) : serverStatus?.instances && serverStatus.instances.length > 0 ? (
            <Table
              columns={containerColumns}
              data={serverStatus.instances}
              keyExtractor={(container: any) => container.id || container.name}
            />
          ) : (
            <div className="text-center py-8 text-gray-500">
              No containers found on this server
            </div>
          )}
        </CardBody>
      </Card>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Delete Server"
      >
        <div className="space-y-4">
          <Alert variant="warning">
            Are you sure you want to delete this server? This action cannot be
            undone.
          </Alert>
          <div className="bg-gray-50 p-4 rounded">
            <p className="text-sm text-gray-700">
              <strong>Server:</strong> {server.hostname}
            </p>
            <p className="text-sm text-gray-700">
              <strong>IP:</strong> {server.ip_address}
            </p>
          </div>
          <div className="flex justify-end space-x-2">
            <Button
              variant="outline"
              onClick={() => setDeleteModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete Server'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Test Connection Modal */}
      <Modal
        isOpen={testModalOpen}
        onClose={() => setTestModalOpen(false)}
        title="Test Connection"
      >
        <div className="space-y-4">
          {testMutation.isPending ? (
            <div className="flex flex-col items-center justify-center py-8">
              <Spinner size="lg" />
              <p className="mt-4 text-sm text-gray-600">
                Testing connection to {server.ip_address}:{server.agent_port}...
              </p>
            </div>
          ) : testResult ? (
            <div>
              {testResult.status === 'healthy' ? (
                <Alert variant="success">
                  <CheckCircle className="w-5 h-5 mr-2" />
                  Connection successful! Agent is responding.
                </Alert>
              ) : (
                <Alert variant="danger">
                  <XCircle className="w-5 h-5 mr-2" />
                  Connection failed: {testResult.error || 'Unknown error'}
                </Alert>
              )}
            </div>
          ) : null}
          <div className="flex justify-end">
            <Button variant="outline" onClick={() => setTestModalOpen(false)}>
              Close
            </Button>
          </div>
        </div>
      </Modal>

      {/* Container Logs Modal */}
      <Modal
        isOpen={logsModalOpen}
        onClose={() => {
          setLogsModalOpen(false);
          setSelectedContainer(null);
        }}
        title="Container Logs"
        size="xl"
      >
        {selectedContainer && id && (
          <LogsViewer
            serverId={id}
            containerId={selectedContainer.id}
            containerName={selectedContainer.name}
            onClose={() => {
              setLogsModalOpen(false);
              setSelectedContainer(null);
            }}
          />
        )}
      </Modal>
    </div>
  );
};

export default ServerDetail;

// Made with Bob
