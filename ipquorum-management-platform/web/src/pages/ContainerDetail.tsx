import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  ArrowLeft,
  Play,
  Square,
  RotateCw,
  Trash2,
  RefreshCw,
  Activity,
  Server,
  Settings as SettingsIcon,
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
import apiClient from '../services/api';
import { useToastStore } from '../stores/toastStore';
import type { ContainerInstance } from '../types/container';

const ContainerDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const toast = useToastStore();
  
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);

  // Fetch container details
  const {
    data: container,
    isLoading,
    error,
    refetch,
  } = useQuery<ContainerInstance>({
    queryKey: ['container', id],
    queryFn: async () => {
      const response = await apiClient.get(`/containers/${id}`);
      return response.data;
    },
    refetchInterval: 10000, // Refresh every 10 seconds
    enabled: !!id,
  });

  // Start mutation
  const startMutation = useMutation({
    mutationFn: async () => {
      await apiClient.post(`/containers/${id}/start`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['container', id] });
      queryClient.invalidateQueries({ queryKey: ['containers'] });
      toast.success('Container started successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to start container');
    },
  });

  // Stop mutation
  const stopMutation = useMutation({
    mutationFn: async () => {
      await apiClient.post(`/containers/${id}/stop`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['container', id] });
      queryClient.invalidateQueries({ queryKey: ['containers'] });
      toast.success('Container stopped successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to stop container');
    },
  });

  // Restart mutation
  const restartMutation = useMutation({
    mutationFn: async () => {
      await apiClient.post(`/containers/${id}/restart`);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['container', id] });
      queryClient.invalidateQueries({ queryKey: ['containers'] });
      toast.success('Container restarted successfully');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to restart container');
    },
  });

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: async () => {
      await apiClient.delete(`/containers/${id}`);
    },
    onSuccess: () => {
      toast.success('Container deleted successfully');
      navigate('/containers');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to delete container');
    },
  });

  const handleStart = () => startMutation.mutate();
  const handleStop = () => stopMutation.mutate();
  const handleRestart = () => restartMutation.mutate();
  const handleDelete = () => {
    deleteMutation.mutate();
    setDeleteModalOpen(false);
  };

  const getStatusBadge = (status: string) => {
    const statusMap: Record<string, { variant: 'success' | 'warning' | 'danger' | 'info'; label: string }> = {
      running: { variant: 'success', label: 'Running' },
      stopped: { variant: 'danger', label: 'Stopped' },
      created: { variant: 'info', label: 'Created' },
      paused: { variant: 'warning', label: 'Paused' },
      restarting: { variant: 'warning', label: 'Restarting' },
      removing: { variant: 'warning', label: 'Removing' },
      exited: { variant: 'danger', label: 'Exited' },
      dead: { variant: 'danger', label: 'Dead' },
    };
    const config = statusMap[status] || { variant: 'info' as const, label: status };
    return <Badge variant={config.variant}>{config.label}</Badge>;
  };

  const getRuntimeBadge = (runtime: string) => {
    const runtimeMap: Record<string, { variant: 'info' | 'success'; label: string }> = {
      docker: { variant: 'info', label: 'Docker' },
      podman: { variant: 'success', label: 'Podman' },
    };
    const config = runtimeMap[runtime] || { variant: 'info' as const, label: runtime };
    return <Badge variant={config.variant}>{config.label}</Badge>;
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Spinner size="lg" />
      </div>
    );
  }

  if (error || !container) {
    return (
      <div className="space-y-4">
        <Button
          variant="ghost"
          leftIcon={<ArrowLeft className="h-4 w-4" />}
          onClick={() => navigate('/containers')}
        >
          Back to Containers
        </Button>
        <Alert variant="danger">
          <p className="font-semibold">Error Loading Container</p>
          <p className="text-sm mt-1">
            {(error as any)?.response?.data?.error || 'Container not found'}
          </p>
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
            leftIcon={<ArrowLeft className="h-4 w-4" />}
            onClick={() => navigate('/containers')}
          >
            Back
          </Button>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{container.container_name}</h1>
            <p className="text-sm text-gray-600 mt-1">Container ID: {container.id}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            leftIcon={<RefreshCw className="h-4 w-4" />}
            onClick={() => refetch()}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Status and Actions */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Activity className="h-5 w-5 text-gray-600" />
              <h2 className="text-lg font-semibold">Status & Actions</h2>
            </div>
            {getStatusBadge(container.status)}
          </div>
        </CardHeader>
        <CardBody>
          <div className="flex items-center gap-2">
            {(container.status === 'stopped' || container.status === 'created') && (
              <Button
                variant="default"
                leftIcon={<Play className="h-4 w-4" />}
                onClick={handleStart}
                disabled={startMutation.isPending}
              >
                {startMutation.isPending ? 'Starting...' : 'Start'}
              </Button>
            )}
            
            {container.status === 'running' && (
              <>
                <Button
                  variant="secondary"
                  leftIcon={<Square className="h-4 w-4" />}
                  onClick={handleStop}
                  disabled={stopMutation.isPending}
                >
                  {stopMutation.isPending ? 'Stopping...' : 'Stop'}
                </Button>
                <Button
                  variant="secondary"
                  leftIcon={<RotateCw className="h-4 w-4" />}
                  onClick={handleRestart}
                  disabled={restartMutation.isPending}
                >
                  {restartMutation.isPending ? 'Restarting...' : 'Restart'}
                </Button>
              </>
            )}
            
            <Button
              variant="destructive"
              leftIcon={<Trash2 className="h-4 w-4" />}
              onClick={() => setDeleteModalOpen(true)}
            >
              Delete
            </Button>
          </div>
        </CardBody>
      </Card>

      {/* Container Information */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Server className="h-5 w-5 text-gray-600" />
            <h2 className="text-lg font-semibold">Container Information</h2>
          </div>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Container ID</p>
              <p className="text-sm text-gray-900 font-mono">{container.container_id || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Runtime</p>
              <div>{getRuntimeBadge(container.runtime)}</div>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Image</p>
              <p className="text-sm text-gray-900 font-mono">
                {container.image_name}:{container.image_tag || 'latest'}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Agent ID</p>
              <p className="text-sm text-gray-900">{container.agent_id || 'N/A'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Created At</p>
              <p className="text-sm text-gray-900">
                {container.created_at ? new Date(container.created_at).toLocaleString() : 'N/A'}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600 mb-1">Updated At</p>
              <p className="text-sm text-gray-900">
                {container.updated_at ? new Date(container.updated_at).toLocaleString() : 'N/A'}
              </p>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Configuration */}
      {container.config && (
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <SettingsIcon className="h-5 w-5 text-gray-600" />
              <h2 className="text-lg font-semibold">Configuration</h2>
            </div>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              {/* Ports */}
              {container.config.ports && container.config.ports.length > 0 && (
                <div>
                  <p className="text-sm font-medium text-gray-600 mb-2">Port Mappings</p>
                  <div className="space-y-1">
                    {container.config.ports.map((port, idx) => (
                      <p key={idx} className="text-sm text-gray-900 font-mono">
                        {port.host_port}:{port.container_port}/{port.protocol || 'tcp'}
                      </p>
                    ))}
                  </div>
                </div>
              )}

              {/* Environment Variables */}
              {container.config.environment && Object.keys(container.config.environment).length > 0 && (
                <div>
                  <p className="text-sm font-medium text-gray-600 mb-2">Environment Variables</p>
                  <div className="space-y-1">
                    {Object.entries(container.config.environment).map(([key, value]) => (
                      <p key={key} className="text-sm text-gray-900 font-mono">
                        {key}={String(value)}
                      </p>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </CardBody>
        </Card>
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        title="Delete Container"
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to delete container <strong>{container.container_name}</strong>?
            This action cannot be undone.
          </p>
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              onClick={() => setDeleteModalOpen(false)}
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

export default ContainerDetail;

// Made with Bob
