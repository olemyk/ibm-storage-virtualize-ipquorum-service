import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, HardDrive, Trash2, CheckCircle, XCircle, AlertCircle } from 'lucide-react';
import {
  Card,
  CardHeader,
  CardBody,
  Button,
  Badge,
  Table,
  Modal,
  Alert,
  Spinner,
  Input,
} from '../components/ui';
import { api } from '../services/api';
import type { Server, ServerFormData } from '../types/server';
import type { Column } from '../components/ui/Table';

const Servers: React.FC = () => {
  const queryClient = useQueryClient();
  const [registerModalOpen, setRegisterModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedServer, setSelectedServer] = useState<Server | null>(null);
  const [actionError, setActionError] = useState('');
  const [formData, setFormData] = useState<ServerFormData>({
    hostname: '',
    ip_address: '',
    agent_port: 9090,
    api_key: '',
    tls_enabled: true,
    tls_verify: false,
  });

  // Fetch servers
  const {
    data: servers,
    isLoading,
    error,
  } = useQuery<Server[]>({
    queryKey: ['servers'],
    queryFn: api.servers.list,
    refetchInterval: 30000,
  });

  // Register server mutation
  const registerMutation = useMutation({
    mutationFn: (data: ServerFormData) => api.servers.register(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setRegisterModalOpen(false);
      resetForm();
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to register server');
    },
  });

  // Delete server mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.servers.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['servers'] });
      setDeleteModalOpen(false);
      setSelectedServer(null);
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to delete server');
    },
  });

  const resetForm = () => {
    setFormData({
      hostname: '',
      ip_address: '',
      agent_port: 9090,
      api_key: '',
      tls_enabled: true,
      tls_verify: false,
    });
    setActionError('');
  };

  const handleRegister = () => {
    setActionError('');
    registerMutation.mutate(formData);
  };

  const handleDelete = () => {
    if (selectedServer) {
      setActionError('');
      deleteMutation.mutate(selectedServer.id);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'online':
        return <Badge variant="success" dot>Online</Badge>;
      case 'offline':
        return <Badge variant="destructive" dot>Offline</Badge>;
      case 'error':
        return <Badge variant="destructive" dot>Error</Badge>;
      default:
        return <Badge variant="secondary" dot>Unknown</Badge>;
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'online':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'offline':
        return <XCircle className="h-5 w-5 text-red-500" />;
      default:
        return <AlertCircle className="h-5 w-5 text-gray-500" />;
    }
  };

  // Calculate statistics
  const stats = {
    total: servers?.length || 0,
    online: servers?.filter(s => s.status === 'online').length || 0,
    offline: servers?.filter(s => s.status === 'offline' || s.status === 'error').length || 0,
  };

  const columns: Column<Server>[] = [
    {
      key: 'hostname',
      header: 'Hostname',
      render: (server) => (
        <div className="flex items-center gap-2">
          {getStatusIcon(server.status)}
          <span className="font-medium">{server.hostname}</span>
        </div>
      ),
    },
    {
      key: 'ip_address',
      header: 'IP Address',
    },
    {
      key: 'agent_port',
      header: 'Port',
    },
    {
      key: 'status',
      header: 'Status',
      render: (server) => getStatusBadge(server.status),
    },
    {
      key: 'tls_enabled',
      header: 'TLS',
      render: (server) => (
        <Badge variant={server.tls_enabled ? 'success' : 'secondary'}>
          {server.tls_enabled ? 'Enabled' : 'Disabled'}
        </Badge>
      ),
    },
    {
      key: 'last_seen',
      header: 'Last Seen',
      render: (server) => {
        if (!server.last_seen) return <span className="text-muted-foreground">Never</span>;
        const date = new Date(server.last_seen);
        return (
          <span className="text-sm text-muted-foreground">
            {date.toLocaleString()}
          </span>
        );
      },
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (server) => (
        <div className="flex gap-2">
          <Button
            variant="ghost"
            size="sm"
            leftIcon={<Trash2 className="h-4 w-4" />}
            onClick={() => {
              setSelectedServer(server);
              setDeleteModalOpen(true);
            }}
          >
            Delete
          </Button>
        </div>
      ),
    },
  ];

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <Alert variant="destructive">
        Failed to load servers: {(error as Error).message}
      </Alert>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Servers</h1>
          <p className="text-muted-foreground mt-1">
            Manage remote agent servers
          </p>
        </div>
        <Button
          leftIcon={<Plus className="h-5 w-5" />}
          onClick={() => {
            resetForm();
            setRegisterModalOpen(true);
          }}
        >
          Register Server
        </Button>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Card>
          <CardBody>
            <div className="flex items-center gap-4">
              <div className="p-3 bg-primary/10 rounded-lg">
                <HardDrive className="h-6 w-6 text-primary" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total Servers</p>
                <p className="text-2xl font-bold">{stats.total}</p>
              </div>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="flex items-center gap-4">
              <div className="p-3 bg-green-100 rounded-lg">
                <CheckCircle className="h-6 w-6 text-green-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Online</p>
                <p className="text-2xl font-bold text-green-600">{stats.online}</p>
              </div>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="flex items-center gap-4">
              <div className="p-3 bg-red-100 rounded-lg">
                <XCircle className="h-6 w-6 text-red-600" />
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Offline</p>
                <p className="text-2xl font-bold text-red-600">{stats.offline}</p>
              </div>
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Error Alert */}
      {actionError && (
        <Alert variant="destructive">
          {actionError}
        </Alert>
      )}

      {/* Servers Table */}
      <Card>
        <CardHeader>
          <h2 className="text-xl font-semibold">Registered Servers</h2>
        </CardHeader>
        <CardBody>
          {servers && servers.length > 0 ? (
            <Table columns={columns} data={servers} keyExtractor={(server) => server.id} />
          ) : (
            <div className="text-center py-12">
              <HardDrive className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">No servers registered yet</p>
              <Button
                variant="outline"
                className="mt-4"
                onClick={() => setRegisterModalOpen(true)}
              >
                Register Your First Server
              </Button>
            </div>
          )}
        </CardBody>
      </Card>

      {/* Register Server Modal */}
      <Modal
        isOpen={registerModalOpen}
        onClose={() => {
          setRegisterModalOpen(false);
          resetForm();
        }}
        title="Register New Server"
      >
        <div className="space-y-4">
          {actionError && (
            <Alert variant="destructive">
              {actionError}
            </Alert>
          )}

          <Input
            label="Hostname"
            value={formData.hostname}
            onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
            placeholder="agent-server-01"
            required
          />

          <Input
            label="IP Address"
            value={formData.ip_address}
            onChange={(e) => setFormData({ ...formData, ip_address: e.target.value })}
            placeholder="10.33.3.215"
            required
          />

          <Input
            label="Agent Port"
            type="number"
            value={formData.agent_port}
            onChange={(e) => setFormData({ ...formData, agent_port: parseInt(e.target.value) || 9090 })}
            placeholder="9090"
            required
          />

          <Input
            label="API Key"
            type="password"
            value={formData.api_key}
            onChange={(e) => setFormData({ ...formData, api_key: e.target.value })}
            placeholder="Enter agent API key"
            required
          />

          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              id="tls_enabled"
              checked={formData.tls_enabled}
              onChange={(e) => setFormData({ ...formData, tls_enabled: e.target.checked })}
              className="h-4 w-4 rounded border-gray-300"
            />
            <label htmlFor="tls_enabled" className="text-sm font-medium">
              Enable TLS
            </label>
          </div>

          {formData.tls_enabled && (
            <div className="flex items-center gap-2 ml-6">
              <input
                type="checkbox"
                id="tls_verify"
                checked={formData.tls_verify}
                onChange={(e) => setFormData({ ...formData, tls_verify: e.target.checked })}
                className="h-4 w-4 rounded border-gray-300"
              />
              <label htmlFor="tls_verify" className="text-sm font-medium">
                Verify TLS Certificate
              </label>
            </div>
          )}

          <div className="flex gap-2 justify-end pt-4">
            <Button
              variant="outline"
              onClick={() => {
                setRegisterModalOpen(false);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleRegister}
              disabled={registerMutation.isPending}
            >
              {registerMutation.isPending ? 'Registering...' : 'Register'}
            </Button>
          </div>
        </div>
      </Modal>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => {
          setDeleteModalOpen(false);
          setSelectedServer(null);
        }}
        title="Delete Server"
      >
        <div className="space-y-4">
          {actionError && (
            <Alert variant="destructive">
              {actionError}
            </Alert>
          )}

          <p>
            Are you sure you want to delete server{' '}
            <strong>{selectedServer?.hostname}</strong>? This action cannot be undone.
          </p>

          <div className="flex gap-2 justify-end">
            <Button
              variant="outline"
              onClick={() => {
                setDeleteModalOpen(false);
                setSelectedServer(null);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default Servers;


