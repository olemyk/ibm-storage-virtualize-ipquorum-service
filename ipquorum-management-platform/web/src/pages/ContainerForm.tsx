import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Save, X, ArrowLeft, Plus, Trash2 } from 'lucide-react';
import {
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  Button,
  Input,
  Alert,
} from '../components/ui';
import { useContainerStore } from '../stores/containerStore';
import { api } from '../services/api';
import type { 
  CreateContainerRequest, 
  PortMapping, 
  VolumeMount, 
  RestartPolicy 
} from '../types/container';
import type { Instance } from '../types/instance';

interface FormData {
  instance_id: string;
  agent_id: string;
  image_name: string;
  image_tag: string;
  restart_policy: RestartPolicy;
  ports: PortMapping[];
  volumes: VolumeMount[];
  environment: Array<{ key: string; value: string }>;
  labels: Array<{ key: string; value: string }>;
}

const ContainerForm: React.FC = () => {
  const navigate = useNavigate();
  const { createContainer, isLoading: isCreating, error: createError } = useContainerStore();

  const [formData, setFormData] = useState<FormData>({
    instance_id: '',
    agent_id: '',
    image_name: '',
    image_tag: 'latest',
    restart_policy: 'unless-stopped',
    ports: [],
    volumes: [],
    environment: [],
    labels: [],
  });

  const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});
  const [submitError, setSubmitError] = useState('');

  // Fetch instances list for dropdown
  const { data: instances } = useQuery<Instance[]>({
    queryKey: ['instances'],
    queryFn: () => api.instances.list(),
  });

  // Update agent_id when instance_id changes
  useEffect(() => {
    if (formData.instance_id && instances) {
      const selectedInstance = instances.find(i => i.id === formData.instance_id);
      if (selectedInstance) {
        setFormData(prev => ({ ...prev, agent_id: selectedInstance.server_id }));
      }
    }
  }, [formData.instance_id, instances]);

  const validateForm = (): boolean => {
    const newErrors: Partial<Record<keyof FormData, string>> = {};

    if (!formData.instance_id.trim()) {
      newErrors.instance_id = 'Instance is required';
    }

    if (!formData.agent_id.trim()) {
      newErrors.agent_id = 'Agent ID is required';
    }

    if (!formData.image_name.trim()) {
      newErrors.image_name = 'Image name is required';
    }

    // Validate ports
    const invalidPorts = formData.ports.some(
      p => !p.host_port || !p.container_port || p.host_port < 1 || p.container_port < 1
    );
    if (invalidPorts) {
      newErrors.ports = 'All port mappings must have valid host and container ports';
    }

    // Validate volumes
    const invalidVolumes = formData.volumes.some(
      v => !v.source.trim() || !v.destination.trim()
    );
    if (invalidVolumes) {
      newErrors.volumes = 'All volume mounts must have source and destination paths';
    }

    // Validate environment variables
    const invalidEnv = formData.environment.some(e => !e.key.trim());
    if (invalidEnv) {
      newErrors.environment = 'All environment variables must have a key';
    }

    // Validate labels
    const invalidLabels = formData.labels.some(l => !l.key.trim());
    if (invalidLabels) {
      newErrors.labels = 'All labels must have a key';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError('');

    if (!validateForm()) {
      return;
    }

    try {
      // Convert arrays to objects for API
      const envObj: Record<string, string> = {};
      formData.environment.forEach(e => {
        if (e.key.trim()) {
          envObj[e.key] = e.value;
        }
      });

      const labelsObj: Record<string, string> = {};
      formData.labels.forEach(l => {
        if (l.key.trim()) {
          labelsObj[l.key] = l.value;
        }
      });

      const requestData: CreateContainerRequest = {
        instance_id: formData.instance_id,
        agent_id: formData.agent_id,
        image_name: formData.image_name,
        image_tag: formData.image_tag || 'latest',
        restart_policy: formData.restart_policy,
        config: {
          ports: formData.ports.length > 0 ? formData.ports : undefined,
          volumes: formData.volumes.length > 0 ? formData.volumes : undefined,
          environment: Object.keys(envObj).length > 0 ? envObj : undefined,
          labels: Object.keys(labelsObj).length > 0 ? labelsObj : undefined,
        },
      };

      await createContainer(requestData);
      navigate('/containers');
    } catch (error: any) {
      setSubmitError(error.response?.data?.error || error.message || 'Failed to create container');
    }
  };

  const handleInputChange = (field: keyof FormData, value: any) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error for this field
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  // Port mapping handlers
  const addPort = () => {
    setFormData(prev => ({
      ...prev,
      ports: [...prev.ports, { host_port: 0, container_port: 0, protocol: 'tcp' }],
    }));
  };

  const updatePort = (index: number, field: keyof PortMapping, value: any) => {
    setFormData(prev => ({
      ...prev,
      ports: prev.ports.map((p, i) => (i === index ? { ...p, [field]: value } : p)),
    }));
  };

  const removePort = (index: number) => {
    setFormData(prev => ({
      ...prev,
      ports: prev.ports.filter((_, i) => i !== index),
    }));
  };

  // Volume mount handlers
  const addVolume = () => {
    setFormData(prev => ({
      ...prev,
      volumes: [...prev.volumes, { source: '', destination: '', read_only: false }],
    }));
  };

  const updateVolume = (index: number, field: keyof VolumeMount, value: any) => {
    setFormData(prev => ({
      ...prev,
      volumes: prev.volumes.map((v, i) => (i === index ? { ...v, [field]: value } : v)),
    }));
  };

  const removeVolume = (index: number) => {
    setFormData(prev => ({
      ...prev,
      volumes: prev.volumes.filter((_, i) => i !== index),
    }));
  };

  // Environment variable handlers
  const addEnvironment = () => {
    setFormData(prev => ({
      ...prev,
      environment: [...prev.environment, { key: '', value: '' }],
    }));
  };

  const updateEnvironment = (index: number, field: 'key' | 'value', value: string) => {
    setFormData(prev => ({
      ...prev,
      environment: prev.environment.map((e, i) => (i === index ? { ...e, [field]: value } : e)),
    }));
  };

  const removeEnvironment = (index: number) => {
    setFormData(prev => ({
      ...prev,
      environment: prev.environment.filter((_, i) => i !== index),
    }));
  };

  // Label handlers
  const addLabel = () => {
    setFormData(prev => ({
      ...prev,
      labels: [...prev.labels, { key: '', value: '' }],
    }));
  };

  const updateLabel = (index: number, field: 'key' | 'value', value: string) => {
    setFormData(prev => ({
      ...prev,
      labels: prev.labels.map((l, i) => (i === index ? { ...l, [field]: value } : l)),
    }));
  };

  const removeLabel = (index: number) => {
    setFormData(prev => ({
      ...prev,
      labels: prev.labels.filter((_, i) => i !== index),
    }));
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <Button
            variant="ghost"
            size="sm"
            onClick={() => navigate('/containers')}
            leftIcon={<ArrowLeft className="h-4 w-4" />}
            className="mb-2"
          >
            Back to Containers
          </Button>
          <h1 className="text-3xl font-bold text-gray-900">Create Container</h1>
          <p className="mt-2 text-gray-600">
            Deploy a new container instance
          </p>
        </div>
      </div>

      {/* Error Alert */}
      {(submitError || createError) && (
        <Alert
          variant="destructive"
          dismissible
          onDismiss={() => {
            setSubmitError('');
            useContainerStore.setState({ error: null });
          }}
        >
          {submitError || createError}
        </Alert>
      )}

      <form onSubmit={handleSubmit}>
        {/* Basic Configuration */}
        <Card className="mb-6">
          <CardHeader>
            <h2 className="text-xl font-semibold">Basic Configuration</h2>
          </CardHeader>
          <CardBody>
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              {/* Instance Selection */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Instance <span className="text-red-500">*</span>
                </label>
                <select
                  className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                  value={formData.instance_id}
                  onChange={(e) => handleInputChange('instance_id', e.target.value)}
                  required
                >
                  <option value="">Select an instance</option>
                  {instances?.map((instance) => (
                    <option key={instance.id} value={instance.id}>
                      {instance.name} ({instance.api_endpoint})
                    </option>
                  ))}
                </select>
                {errors.instance_id && (
                  <p className="mt-1 text-sm text-red-600">{errors.instance_id}</p>
                )}
              </div>

              {/* Agent ID (auto-filled) */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Agent ID <span className="text-red-500">*</span>
                </label>
                <Input
                  type="text"
                  value={formData.agent_id}
                  onChange={(e) => handleInputChange('agent_id', e.target.value)}
                  placeholder="Agent ID (auto-filled from instance)"
                  required
                  disabled
                />
                {errors.agent_id && (
                  <p className="mt-1 text-sm text-red-600">{errors.agent_id}</p>
                )}
              </div>

              {/* Image Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Image Name <span className="text-red-500">*</span>
                </label>
                <Input
                  type="text"
                  value={formData.image_name}
                  onChange={(e) => handleInputChange('image_name', e.target.value)}
                  placeholder="e.g., nginx, redis, postgres"
                  required
                />
                {errors.image_name && (
                  <p className="mt-1 text-sm text-red-600">{errors.image_name}</p>
                )}
              </div>

              {/* Image Tag */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Image Tag
                </label>
                <Input
                  type="text"
                  value={formData.image_tag}
                  onChange={(e) => handleInputChange('image_tag', e.target.value)}
                  placeholder="latest"
                />
              </div>

              {/* Restart Policy */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Restart Policy
                </label>
                <select
                  className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                  value={formData.restart_policy}
                  onChange={(e) => handleInputChange('restart_policy', e.target.value as RestartPolicy)}
                >
                  <option value="no">No</option>
                  <option value="always">Always</option>
                  <option value="on-failure">On Failure</option>
                  <option value="unless-stopped">Unless Stopped</option>
                </select>
              </div>
            </div>
          </CardBody>
        </Card>

        {/* Port Mappings */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">Port Mappings</h2>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addPort}
                leftIcon={<Plus className="h-4 w-4" />}
              >
                Add Port
              </Button>
            </div>
          </CardHeader>
          <CardBody>
            {formData.ports.length === 0 ? (
              <p className="text-sm text-gray-500">No port mappings configured</p>
            ) : (
              <div className="space-y-4">
                {formData.ports.map((port, index) => (
                  <div key={index} className="flex items-center gap-4">
                    <div className="flex-1">
                      <Input
                        type="number"
                        value={port.host_port || ''}
                        onChange={(e) => updatePort(index, 'host_port', parseInt(e.target.value) || 0)}
                        placeholder="Host Port"
                        min="1"
                        max="65535"
                      />
                    </div>
                    <span className="text-gray-500">→</span>
                    <div className="flex-1">
                      <Input
                        type="number"
                        value={port.container_port || ''}
                        onChange={(e) => updatePort(index, 'container_port', parseInt(e.target.value) || 0)}
                        placeholder="Container Port"
                        min="1"
                        max="65535"
                      />
                    </div>
                    <div className="w-24">
                      <select
                        className="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary"
                        value={port.protocol}
                        onChange={(e) => updatePort(index, 'protocol', e.target.value)}
                      >
                        <option value="tcp">TCP</option>
                        <option value="udp">UDP</option>
                      </select>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removePort(index)}
                      leftIcon={<Trash2 className="h-4 w-4" />}
                    />
                  </div>
                ))}
              </div>
            )}
            {errors.ports && (
              <p className="mt-2 text-sm text-red-600">{errors.ports}</p>
            )}
          </CardBody>
        </Card>

        {/* Volume Mounts */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">Volume Mounts</h2>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addVolume}
                leftIcon={<Plus className="h-4 w-4" />}
              >
                Add Volume
              </Button>
            </div>
          </CardHeader>
          <CardBody>
            {formData.volumes.length === 0 ? (
              <p className="text-sm text-gray-500">No volume mounts configured</p>
            ) : (
              <div className="space-y-4">
                {formData.volumes.map((volume, index) => (
                  <div key={index} className="flex items-center gap-4">
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={volume.source}
                        onChange={(e) => updateVolume(index, 'source', e.target.value)}
                        placeholder="Source path (host)"
                      />
                    </div>
                    <span className="text-gray-500">→</span>
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={volume.destination}
                        onChange={(e) => updateVolume(index, 'destination', e.target.value)}
                        placeholder="Destination path (container)"
                      />
                    </div>
                    <div className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={volume.read_only}
                        onChange={(e) => updateVolume(index, 'read_only', e.target.checked)}
                        className="h-4 w-4 rounded border-gray-300 text-primary focus:ring-primary"
                      />
                      <label className="text-sm text-gray-700">Read-only</label>
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeVolume(index)}
                      leftIcon={<Trash2 className="h-4 w-4" />}
                    />
                  </div>
                ))}
              </div>
            )}
            {errors.volumes && (
              <p className="mt-2 text-sm text-red-600">{errors.volumes}</p>
            )}
          </CardBody>
        </Card>

        {/* Environment Variables */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">Environment Variables</h2>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addEnvironment}
                leftIcon={<Plus className="h-4 w-4" />}
              >
                Add Variable
              </Button>
            </div>
          </CardHeader>
          <CardBody>
            {formData.environment.length === 0 ? (
              <p className="text-sm text-gray-500">No environment variables configured</p>
            ) : (
              <div className="space-y-4">
                {formData.environment.map((env, index) => (
                  <div key={index} className="flex items-center gap-4">
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={env.key}
                        onChange={(e) => updateEnvironment(index, 'key', e.target.value)}
                        placeholder="KEY"
                      />
                    </div>
                    <span className="text-gray-500">=</span>
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={env.value}
                        onChange={(e) => updateEnvironment(index, 'value', e.target.value)}
                        placeholder="value"
                      />
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeEnvironment(index)}
                      leftIcon={<Trash2 className="h-4 w-4" />}
                    />
                  </div>
                ))}
              </div>
            )}
            {errors.environment && (
              <p className="mt-2 text-sm text-red-600">{errors.environment}</p>
            )}
          </CardBody>
        </Card>

        {/* Labels */}
        <Card className="mb-6">
          <CardHeader>
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold">Labels</h2>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addLabel}
                leftIcon={<Plus className="h-4 w-4" />}
              >
                Add Label
              </Button>
            </div>
          </CardHeader>
          <CardBody>
            {formData.labels.length === 0 ? (
              <p className="text-sm text-gray-500">No labels configured</p>
            ) : (
              <div className="space-y-4">
                {formData.labels.map((label, index) => (
                  <div key={index} className="flex items-center gap-4">
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={label.key}
                        onChange={(e) => updateLabel(index, 'key', e.target.value)}
                        placeholder="key"
                      />
                    </div>
                    <span className="text-gray-500">=</span>
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={label.value}
                        onChange={(e) => updateLabel(index, 'value', e.target.value)}
                        placeholder="value"
                      />
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeLabel(index)}
                      leftIcon={<Trash2 className="h-4 w-4" />}
                    />
                  </div>
                ))}
              </div>
            )}
            {errors.labels && (
              <p className="mt-2 text-sm text-red-600">{errors.labels}</p>
            )}
          </CardBody>
        </Card>

        {/* Form Actions */}
        <Card>
          <CardFooter>
            <div className="flex items-center justify-end gap-4">
              <Button
                type="button"
                variant="ghost"
                onClick={() => navigate('/containers')}
                leftIcon={<X className="h-4 w-4" />}
              >
                Cancel
              </Button>
              <Button
                type="submit"
                variant="default"
                isLoading={isCreating}
                leftIcon={<Save className="h-4 w-4" />}
              >
                Create Container
              </Button>
            </div>
          </CardFooter>
        </Card>
      </form>
    </div>
  );
};

export default ContainerForm;

// Made with Bob
