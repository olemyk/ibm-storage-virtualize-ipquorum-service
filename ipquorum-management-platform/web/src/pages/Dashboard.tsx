import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Server, Activity, AlertCircle, CheckCircle } from 'lucide-react';
import { Card, CardHeader, CardBody, Badge, Spinner, Alert } from '../components/ui';
import { api } from '../services/api';
import type { Instance, SystemHealth } from '../types/instance';

const Dashboard: React.FC = () => {
  // Fetch instances
  const {
    data: instances,
    isLoading: instancesLoading,
    error: instancesError,
  } = useQuery<Instance[]>({
    queryKey: ['instances'],
    queryFn: api.instances.list,
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  // Fetch health status
  const {
    data: health,
    isLoading: healthLoading,
    error: healthError,
  } = useQuery<SystemHealth>({
    queryKey: ['health'],
    queryFn: api.health.check,
    refetchInterval: 10000, // Refetch every 10 seconds
  });

  const stats = React.useMemo(() => {
    if (!instances) return { total: 0, running: 0, stopped: 0, error: 0 };

    return {
      total: instances.length,
      running: instances.filter((i) => i.status === 'running').length,
      stopped: instances.filter((i) => i.status === 'stopped').length,
      error: instances.filter((i) => i.status === 'failed').length,
    };
  }, [instances]);

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'running':
        return <Badge variant="success" dot>Running</Badge>;
      case 'stopped':
        return <Badge variant="secondary" dot>Stopped</Badge>;
      case 'error':
        return <Badge variant="destructive" dot>Error</Badge>;
      default:
        return <Badge variant="default" dot>{status}</Badge>;
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-2 text-gray-600">
          Overview of your IP Quorum instances and system health
        </p>
      </div>

      {/* System Health */}
      {healthError && (
        <Alert variant="destructive" title="Health Check Failed">
          Unable to fetch system health status
        </Alert>
      )}

      {health && (
        <Card variant="elevated">
          <CardHeader title="System Health" />
          <CardBody>
            <div className="flex items-center gap-3">
              {health.status === 'healthy' ? (
                <>
                  <CheckCircle className="h-8 w-8 text-green-500" />
                  <div>
                    <p className="text-lg font-semibold text-gray-900">System Healthy</p>
                    <p className="text-sm text-gray-600">All services operational</p>
                  </div>
                </>
              ) : (
                <>
                  <AlertCircle className="h-8 w-8 text-red-500" />
                  <div>
                    <p className="text-lg font-semibold text-gray-900">System Issues</p>
                    <p className="text-sm text-gray-600">Some services may be degraded</p>
                  </div>
                </>
              )}
            </div>
            {health.database && (
              <div className="mt-4 pt-4 border-t border-gray-200">
                <p className="text-sm text-gray-600">
                  Database: <span className="font-medium text-gray-900">{health.database}</span>
                </p>
              </div>
            )}
          </CardBody>
        </Card>
      )}

      {/* Statistics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card variant="bordered">
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Total Instances</p>
                <p className="mt-2 text-3xl font-bold text-gray-900">{stats.total}</p>
              </div>
              <Server className="h-12 w-12 text-blue-500" />
            </div>
          </CardBody>
        </Card>

        <Card variant="bordered">
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Running</p>
                <p className="mt-2 text-3xl font-bold text-green-600">{stats.running}</p>
              </div>
              <Activity className="h-12 w-12 text-green-500" />
            </div>
          </CardBody>
        </Card>

        <Card variant="bordered">
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Stopped</p>
                <p className="mt-2 text-3xl font-bold text-gray-600">{stats.stopped}</p>
              </div>
              <Server className="h-12 w-12 text-gray-400" />
            </div>
          </CardBody>
        </Card>

        <Card variant="bordered">
          <CardBody>
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Errors</p>
                <p className="mt-2 text-3xl font-bold text-red-600">{stats.error}</p>
              </div>
              <AlertCircle className="h-12 w-12 text-red-500" />
            </div>
          </CardBody>
        </Card>
      </div>

      {/* Recent Instances */}
      <Card variant="elevated">
        <CardHeader
          title="Recent Instances"
          subtitle="Latest IP Quorum instances"
        />
        <CardBody>
          {instancesLoading && (
            <div className="flex justify-center py-8">
              <Spinner size="lg" label="Loading instances..." />
            </div>
          )}

          {instancesError && (
            <Alert variant="destructive">
              Failed to load instances. Please try again later.
            </Alert>
          )}

          {instances && instances.length === 0 && (
            <div className="text-center py-8">
              <Server className="mx-auto h-12 w-12 text-gray-400" />
              <p className="mt-2 text-sm text-gray-600">No instances found</p>
              <p className="text-xs text-gray-500">Create your first instance to get started</p>
            </div>
          )}

          {instances && instances.length > 0 && (
            <div className="space-y-3">
              {instances.slice(0, 5).map((instance) => (
                <div
                  key={instance.id}
                  className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors"
                >
                  <div className="flex-1">
                    <h4 className="font-medium text-gray-900">{instance.name}</h4>
                    <p className="text-sm text-gray-600">{instance.api_endpoint}</p>
                  </div>
                  <div className="flex items-center gap-4">
                    {getStatusBadge(instance.status)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
};

export default Dashboard;


