import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus } from 'lucide-react';
import { Button, Card, CardBody } from '../components/ui';

const IPQuorumContainers: React.FC = () => {
  const navigate = useNavigate();

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">IPQuorum Containers</h1>
          <p className="mt-2 text-gray-600">
            Create and manage IPQuorum containers directly via agent API
          </p>
        </div>
        <Button
          variant="default"
          size="lg"
          leftIcon={<Plus className="h-5 w-5" />}
          onClick={() => navigate('/ipquorum-containers/new')}
        >
          Create IPQuorum Container
        </Button>
      </div>

      {/* Info Card */}
      <Card variant="elevated">
        <CardBody>
          <div className="space-y-4">
            <h2 className="text-xl font-semibold text-gray-900">
              Direct Container Creation
            </h2>
            <p className="text-gray-600">
              This page allows you to create IPQuorum containers directly without requiring
              a systemd instance. Perfect for containerized deployments where systemd
              management is not available.
            </p>
            
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <h3 className="font-medium text-blue-900 mb-2">How it works:</h3>
              <ul className="list-disc list-inside space-y-1 text-blue-800 text-sm">
                <li>Creates containers directly via the container runtime agent</li>
                <li>No systemd instance required</li>
                <li>IPQuorum configuration embedded in container environment</li>
                <li>Ideal for Podman/Docker deployments</li>
              </ul>
            </div>

            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <h3 className="font-medium text-yellow-900 mb-2">Note:</h3>
              <p className="text-yellow-800 text-sm">
                This bypasses the traditional Instances → Containers workflow. Use this when
                your management server is running in a container and cannot manage systemd services.
              </p>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Coming Soon */}
      <Card>
        <CardBody>
          <div className="text-center py-12">
            <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 mb-4">
              <Plus className="h-8 w-8 text-gray-400" />
            </div>
            <h3 className="text-lg font-medium text-gray-900 mb-2">
              Container List Coming Soon
            </h3>
            <p className="text-gray-600 mb-4">
              The container list view is under development. For now, use the Containers page
              to view all containers.
            </p>
            <Button
              variant="outline"
              onClick={() => navigate('/containers')}
            >
              View All Containers
            </Button>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

export default IPQuorumContainers;

// Made with Bob