import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Plus, Play, Square, RotateCw, Trash2, Eye, Edit } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
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
} from '../components/ui';
import { api } from '../services/api';
import type { Instance } from '../types/instance';

const Instances: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [selectedInstance, setSelectedInstance] = useState<Instance | null>(null);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [actionError, setActionError] = useState('');

  // Fetch instances
  const {
    data: instances,
    isLoading,
    error,
  } = useQuery<Instance[]>({
    queryKey: ['instances'],
    queryFn: api.instances.list,
    refetchInterval: 30000,
  });

  // Start instance mutation
  const startMutation = useMutation({
    mutationFn: (id: string) => api.instances.start(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to start instance');
    },
  });

  // Stop instance mutation
  const stopMutation = useMutation({
    mutationFn: (id: string) => api.instances.stop(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to stop instance');
    },
  });

  // Restart instance mutation
  const restartMutation = useMutation({
    mutationFn: (id: string) => api.instances.restart(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to restart instance');
    },
  });

  // Delete instance mutation
  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.instances.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      setDeleteModalOpen(false);
      setSelectedInstance(null);
    },
    onError: (error: any) => {
      setActionError(error.response?.data?.error || 'Failed to delete instance');
    },
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return <Badge variant="success" dot>Running</Badge>;
      case 'stopped':
        return <Badge variant="secondary" dot>Stopped</Badge>;
      case 'failed':
        return <Badge variant="destructive" dot>Failed</Badge>;
      default:
        return <Badge variant="default" dot>{status}</Badge>;
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

  const handleStart = (instance: Instance) => {
    setActionError('');
    startMutation.mutate(instance.id);
  };

  const handleStop = (instance: Instance) => {
    setActionError('');
    stopMutation.mutate(instance.id);
  };

  const handleRestart = (instance: Instance) => {
    setActionError('');
    restartMutation.mutate(instance.id);
  };

  const handleDelete = (instance: Instance) => {
    setSelectedInstance(instance);
    setDeleteModalOpen(true);
    setActionError('');
  };

  const confirmDelete = () => {
    if (selectedInstance) {
      deleteMutation.mutate(selectedInstance.id);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Name',
      sortable: true,
      render: (instance: Instance) => (
        <div>
          <div className="font-medium text-gray-900">{instance.name}</div>
          <div className="text-sm text-gray-500">{instance.api_endpoint}</div>
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (instance: Instance) => getStatusBadge(instance.status),
    },
    {
      key: 'health',
      header: 'Health',
      sortable: true,
      render: (instance: Instance) => getHealthBadge(instance.health),
    },
    {
      key: 'storage_system',
      header: 'Storage System',
      render: (instance: Instance) => (
        <span className="text-sm text-gray-600">{instance.storage_system || '-'}</span>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (instance: Instance) => (
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="ghost"
            onClick={() => navigate(`/instances/${instance.id}`)}
            leftIcon={<Eye className="h-4 w-4" />}
            title="View Instance Details"
          >
            View
          </Button>
          
          {instance.status === 'stopped' && (
            <Button
              size="sm"
              variant="default"
              onClick={() => handleStart(instance)}
              leftIcon={<Play className="h-4 w-4" />}
              isLoading={startMutation.isPending}
              title="Start Instance"
            />
          )}
          
          {instance.status === 'running' && (
            <>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => handleStop(instance)}
                leftIcon={<Square className="h-4 w-4" />}
                isLoading={stopMutation.isPending}
                title="Stop Instance"
              />
              <Button
                size="sm"
                variant="secondary"
                onClick={() => handleRestart(instance)}
                leftIcon={<RotateCw className="h-4 w-4" />}
                isLoading={restartMutation.isPending}
                title="Restart Instance"
              />
            </>
          )}
          
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleDelete(instance)}
            leftIcon={<Trash2 className="h-4 w-4" />}
            isLoading={deleteMutation.isPending}
            title="Delete Instance"
          />
        </div>
      ),
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">IP Quorum Instances</h1>
          <p className="mt-2 text-gray-600">
            Manage your IP Quorum instances
          </p>
        </div>
        <Button
          variant="default"
          size="lg"
          leftIcon={<Plus className="h-5 w-5" />}
          onClick={() => navigate('/instances/new')}
        >
          Create Instance
        </Button>
      </div>

      {/* Error Alert */}
      {actionError && (
        <Alert
          variant="destructive"
          dismissible
          onDismiss={() => setActionError('')}
        >
          {actionError}
        </Alert>
      )}

      {/* Error State */}
      {error && (
        <Alert variant="destructive" title="Error Loading Instances">
          Failed to load instances. Please try again later.
        </Alert>
      )}

      {/* Instances Table */}
      <Card variant="elevated">
        <div className="p-6 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">All Instances</h3>
          <p className="mt-1 text-sm text-gray-500">{instances?.length || 0} total instances</p>
        </div>
        <CardBody>
          <Table
            data={instances || []}
            columns={columns}
            keyExtractor={(instance) => instance.id}
            isLoading={isLoading}
            emptyMessage="No instances found. Create your first instance to get started."
          />
        </CardBody>
      </Card>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Delete Instance"
        footer={
          <>
            <Button
              variant="ghost"
              onClick={() => setDeleteModalOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              isLoading={deleteMutation.isPending}
              leftIcon={<Trash2 className="h-4 w-4" />}
            >
              Delete
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to delete the instance{' '}
            <strong>{selectedInstance?.name}</strong>?
          </p>
          <Alert variant="warning">
            This action cannot be undone. All data associated with this instance will be permanently deleted.
          </Alert>
        </div>
      </Modal>
    </div>
  );
};

export default Instances;


