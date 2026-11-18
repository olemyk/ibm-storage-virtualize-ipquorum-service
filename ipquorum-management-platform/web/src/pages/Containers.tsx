import React, { useEffect, useState } from 'react';
import { Plus, Play, Square, RotateCw, Trash2, Eye, Search, Filter, X } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import {
  Card,
  CardBody,
  Button,
  Badge,
  Table,
  Modal,
  Alert,
  Input,
} from '../components/ui';
import {
  useContainerStore,
  selectFilteredContainers,
  selectContainerStats
} from '../stores/containerStore';
import type { ContainerInstance, ContainerStatus, ContainerRuntime } from '../types/container';
import { formatContainerStatus, getStatusColor } from '../services/containerApi';

const Containers: React.FC = () => {
  const navigate = useNavigate();
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [selectedContainer, setSelectedContainer] = useState<ContainerInstance | null>(null);
  const [forceDelete, setForceDelete] = useState(false);
  const [showFilters, setShowFilters] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // Zustand store
  const {
    containers,
    isLoading,
    error,
    filters,
    sortOptions,
    fetchContainers,
    startContainer,
    stopContainer,
    restartContainer,
    deleteContainer,
    setFilters,
    setSortOptions,
  } = useContainerStore();

  // Fetch containers on mount
  useEffect(() => {
    fetchContainers();
    // Refresh every 30 seconds
    const interval = setInterval(() => {
      fetchContainers();
    }, 30000);
    return () => clearInterval(interval);
  }, [fetchContainers]);

  // Get filtered containers using selectors
  const state = useContainerStore.getState();
  const filteredContainers = selectFilteredContainers(state);
  const stats = selectContainerStats(state);

  // Handle search
  const handleSearch = (value: string) => {
    setSearchTerm(value);
    setFilters({ search: value });
  };

  // Handle filter changes
  const handleStatusFilter = (status: ContainerStatus | 'all') => {
    setFilters({ status: status === 'all' ? undefined : status });
  };

  const handleRuntimeFilter = (runtime: ContainerRuntime | 'all') => {
    setFilters({ runtime: runtime === 'all' ? undefined : runtime });
  };

  const handleAgentFilter = (agentId: string) => {
    setFilters({ agent_id: agentId || undefined });
  };

  // Clear all filters
  const clearFilters = () => {
    setSearchTerm('');
    setFilters({ status: undefined, runtime: undefined, agent_id: undefined, search: undefined });
  };

  // Handle sorting
  const handleSort = (column: string) => {
    const newOrder =
      sortOptions.field === column && sortOptions.order === 'asc' ? 'desc' : 'asc';
    setSortOptions({ field: column as any, order: newOrder });
  };

  // Container actions
  const handleStart = async (container: ContainerInstance) => {
    try {
      await startContainer(container.id);
    } catch (err) {
      console.error('Failed to start container:', err);
    }
  };

  const handleStop = async (container: ContainerInstance) => {
    try {
      await stopContainer(container.id);
    } catch (err) {
      console.error('Failed to stop container:', err);
    }
  };

  const handleRestart = async (container: ContainerInstance) => {
    try {
      await restartContainer(container.id);
    } catch (err) {
      console.error('Failed to restart container:', err);
    }
  };

  const handleDelete = (container: ContainerInstance) => {
    setSelectedContainer(container);
    setForceDelete(false); // Reset force delete checkbox
    setDeleteModalOpen(true);
  };

  const confirmDelete = async () => {
    if (selectedContainer) {
      try {
        await deleteContainer(selectedContainer.id, forceDelete);
        setDeleteModalOpen(false);
        setSelectedContainer(null);
        setForceDelete(false);
      } catch (err) {
        console.error('Failed to delete container:', err);
      }
    }
  };

  // Status badge renderer
  const getStatusBadge = (status: ContainerStatus) => {
    const color = getStatusColor(status);
    const variantMap: Record<string, 'success' | 'warning' | 'danger' | 'secondary' | 'info'> = {
      green: 'success',
      yellow: 'warning',
      red: 'danger',
      gray: 'secondary',
      blue: 'info',
    };
    
    return (
      <Badge variant={variantMap[color] || 'secondary'} dot>
        {formatContainerStatus(status)}
      </Badge>
    );
  };

  // Runtime badge renderer
  const getRuntimeBadge = (runtime: ContainerRuntime) => {
    const variantMap: Record<ContainerRuntime, 'info' | 'secondary'> = {
      docker: 'info',
      podman: 'secondary',
    };
    
    return (
      <Badge variant={variantMap[runtime]} size="sm">
        {runtime.toUpperCase()}
      </Badge>
    );
  };

  // Table columns
  const columns = [
    {
      key: 'container_name',
      header: 'Container',
      sortable: true,
      render: (container: ContainerInstance) => (
        <div>
          <div className="font-medium text-gray-900">{container.container_name}</div>
          <div className="text-xs text-gray-500 font-mono">{container.container_id?.substring(0, 12) || 'N/A'}</div>
        </div>
      ),
    },
    {
      key: 'image',
      header: 'Image',
      sortable: true,
      render: (container: ContainerInstance) => (
        <div className="text-sm">
          <div className="text-gray-900">{container.image_name}</div>
          {container.image_tag && (
            <div className="text-xs text-gray-500">:{container.image_tag}</div>
          )}
        </div>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (container: ContainerInstance) => getStatusBadge(container.status),
    },
    {
      key: 'runtime',
      header: 'Runtime',
      sortable: true,
      render: (container: ContainerInstance) => getRuntimeBadge(container.runtime),
    },
    {
      key: 'agent_id',
      header: 'Agent',
      render: (container: ContainerInstance) => (
        <span className="text-sm text-gray-600">{container.agent_id || 'N/A'}</span>
      ),
    },
    {
      key: 'ports',
      header: 'Ports',
      render: (container: ContainerInstance) => {
        const portCount = container.config?.ports?.length || 0;
        return (
          <span className="text-sm text-gray-600">
            {portCount > 0 ? `${portCount} port${portCount > 1 ? 's' : ''}` : 'None'}
          </span>
        );
      },
    },
    {
      key: 'actions',
      header: 'Actions',
      render: (container: ContainerInstance) => (
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="ghost"
            onClick={() => navigate(`/containers/${container.id}`)}
            leftIcon={<Eye className="h-4 w-4" />}
            title="View Container Details"
          >
            View
          </Button>
          
          {(container.status === 'stopped' || container.status === 'created') && (
            <Button
              size="sm"
              variant="default"
              onClick={() => handleStart(container)}
              leftIcon={<Play className="h-4 w-4" />}
              title="Start Container"
            />
          )}
          
          {container.status === 'running' && (
            <>
              <Button
                size="sm"
                variant="secondary"
                onClick={() => handleStop(container)}
                leftIcon={<Square className="h-4 w-4" />}
                title="Stop Container"
              />
              <Button
                size="sm"
                variant="secondary"
                onClick={() => handleRestart(container)}
                leftIcon={<RotateCw className="h-4 w-4" />}
                title="Restart Container"
              />
            </>
          )}
          
          <Button
            size="sm"
            variant="destructive"
            onClick={() => handleDelete(container)}
            leftIcon={<Trash2 className="h-4 w-4" />}
            title="Delete Container"
          />
        </div>
      ),
    },
  ];

  // Get unique agent IDs for filter dropdown
  const uniqueAgents = Array.from(new Set(containers.map(c => c.agent_id).filter(Boolean)));

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Containers</h1>
          <p className="mt-2 text-gray-600">
            Manage container instances across your infrastructure
          </p>
        </div>
        <Button
          variant="default"
          size="lg"
          leftIcon={<Plus className="h-5 w-5" />}
          onClick={() => navigate('/containers/new')}
        >
          Create Container
        </Button>
      </div>

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-5">
        <Card>
          <CardBody>
            <div className="text-sm font-medium text-gray-500">Total</div>
            <div className="mt-1 text-3xl font-semibold text-gray-900">{stats.total}</div>
          </CardBody>
        </Card>
        <Card>
          <CardBody>
            <div className="text-sm font-medium text-gray-500">Running</div>
            <div className="mt-1 text-3xl font-semibold text-green-600">{stats.running}</div>
          </CardBody>
        </Card>
        <Card>
          <CardBody>
            <div className="text-sm font-medium text-gray-500">Stopped</div>
            <div className="mt-1 text-3xl font-semibold text-gray-600">{stats.stopped}</div>
          </CardBody>
        </Card>
        <Card>
          <CardBody>
            <div className="text-sm font-medium text-gray-500">Paused</div>
            <div className="mt-1 text-3xl font-semibold text-yellow-600">{stats.paused}</div>
          </CardBody>
        </Card>
        <Card>
          <CardBody>
            <div className="text-sm font-medium text-gray-500">Error</div>
            <div className="mt-1 text-3xl font-semibold text-red-600">{stats.error}</div>
          </CardBody>
        </Card>
      </div>

      {/* Error Alert */}
      {error && (
        <Alert
          variant="destructive"
          dismissible
          onDismiss={() => useContainerStore.setState({ error: null })}
        >
          {error}
        </Alert>
      )}

      {/* Search and Filters */}
      <Card>
        <CardBody>
          <div className="space-y-4">
            {/* Search Bar */}
            <div className="flex items-center gap-4">
              <div className="flex-1">
                <Input
                  type="text"
                  placeholder="Search containers by name, image, or ID..."
                  value={searchTerm}
                  onChange={(e) => handleSearch(e.target.value)}
                  leftIcon={<Search className="h-4 w-4" />}
                />
              </div>
              <Button
                variant="outline"
                onClick={() => setShowFilters(!showFilters)}
                leftIcon={<Filter className="h-4 w-4" />}
              >
                Filters
                {(filters.status || filters.runtime || filters.agent_id) && (
                  <Badge variant="info" size="sm" className="ml-2">
                    Active
                  </Badge>
                )}
              </Button>
              {(filters.status || filters.runtime || filters.agent_id || filters.search) && (
                <Button
                  variant="ghost"
                  onClick={clearFilters}
                  leftIcon={<X className="h-4 w-4" />}
                >
                  Clear
                </Button>
              )}
            </div>

            {/* Filter Panel */}
            {showFilters && (
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-3 pt-4 border-t">
                {/* Status Filter */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Status
                  </label>
                  <select
                    className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                    value={filters.status || 'all'}
                    onChange={(e) => handleStatusFilter(e.target.value as any)}
                  >
                    <option value="all">All Statuses</option>
                    <option value="running">Running</option>
                    <option value="stopped">Stopped</option>
                    <option value="paused">Paused</option>
                    <option value="error">Error</option>
                    <option value="created">Created</option>
                    <option value="restarting">Restarting</option>
                    <option value="removing">Removing</option>
                    <option value="dead">Dead</option>
                  </select>
                </div>

                {/* Runtime Filter */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Runtime
                  </label>
                  <select
                    className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                    value={filters.runtime || 'all'}
                    onChange={(e) => handleRuntimeFilter(e.target.value as any)}
                  >
                    <option value="all">All Runtimes</option>
                    <option value="docker">Docker</option>
                    <option value="podman">Podman</option>
                  </select>
                </div>

                {/* Agent Filter */}
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Agent
                  </label>
                  <select
                    className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                    value={filters.agent_id || ''}
                    onChange={(e) => handleAgentFilter(e.target.value)}
                  >
                    <option value="">All Agents</option>
                    {uniqueAgents.map((agentId) => (
                      <option key={agentId} value={agentId}>
                        {agentId}
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            )}
          </div>
        </CardBody>
      </Card>

      {/* Containers Table */}
      <Card variant="elevated">
        <div className="p-6 border-b border-gray-200">
          <h3 className="text-lg font-semibold text-gray-900">Container List</h3>
          <p className="mt-1 text-sm text-gray-500">
            {filteredContainers.length} of {containers.length} containers
          </p>
        </div>
        <CardBody>
          <Table
            data={filteredContainers}
            columns={columns}
            keyExtractor={(container) => container.id}
            sortColumn={sortOptions.field}
            sortDirection={sortOptions.order}
            onSort={handleSort}
            isLoading={isLoading}
            emptyMessage="No containers found. Create your first container to get started."
          />
        </CardBody>
      </Card>

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={deleteModalOpen}
        onClose={() => {
          setDeleteModalOpen(false);
          setForceDelete(false);
        }}
        title="Delete Container"
        footer={
          <>
            <Button
              variant="ghost"
              onClick={() => {
                setDeleteModalOpen(false);
                setForceDelete(false);
              }}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmDelete}
              leftIcon={<Trash2 className="h-4 w-4" />}
            >
              {forceDelete ? 'Force Delete' : 'Delete'}
            </Button>
          </>
        }
      >
        <div className="space-y-4">
          <p className="text-gray-700">
            Are you sure you want to delete the container{' '}
            <strong>{selectedContainer?.container_name}</strong>?
          </p>
          <Alert variant="warning">
            This action cannot be undone. The container will be permanently removed from the system.
          </Alert>
          
          {/* Force Delete Checkbox */}
          <div className="border-t pt-4">
            <label className="flex items-start space-x-3 cursor-pointer">
              <input
                type="checkbox"
                checked={forceDelete}
                onChange={(e) => setForceDelete(e.target.checked)}
                className="mt-1 h-4 w-4 rounded border-gray-300 text-red-600 focus:ring-red-500"
              />
              <div className="flex-1">
                <div className="font-medium text-gray-900">Force Delete</div>
                <div className="text-sm text-gray-600">
                  Skip agent cleanup and remove only from database. Use this when the agent is unavailable or the container no longer exists.
                </div>
              </div>
            </label>
          </div>
          
          {forceDelete && (
            <Alert variant="destructive">
              <strong>Warning:</strong> Force delete will skip all cleanup operations on the agent.
              The container may still exist on the host system and will need to be manually removed.
            </Alert>
          )}
        </div>
      </Modal>
    </div>
  );
};

export default Containers;

// Made with Bob
