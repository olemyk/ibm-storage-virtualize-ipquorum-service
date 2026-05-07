import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Save, X, ArrowLeft } from 'lucide-react';
import {
  Card,
  CardHeader,
  CardBody,
  CardFooter,
  Button,
  Input,
  Alert,
} from '../components/ui';
import { api } from '../services/api';
import type { Instance } from '../types/instance';
import type { Server } from '../types/server';

interface FormData {
  name: string;
  server_id: string;
  api_endpoint: string;
  username: string;
  password: string;
  partnersystem: string;
  storage_system: string;
  description: string;
  location: string;
  enable_download: boolean;
  enable_mkquorumapp: boolean;
  ipquorum_name: string;
  ip6: boolean;
  partnerip6: boolean;
  nometadata: boolean;
}

const InstanceForm: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isEditMode = id !== 'new' && !!id;

  const [formData, setFormData] = useState<FormData>({
    name: '',
    server_id: 'local',
    api_endpoint: '',
    username: '',
    password: '',
    partnersystem: '',
    storage_system: '',
    description: '',
    location: '',
    enable_download: true,
    enable_mkquorumapp: false,
    ipquorum_name: '',
    ip6: false,
    partnerip6: false,
    nometadata: false,
  });

  const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});
  const [submitError, setSubmitError] = useState('');

  // Fetch servers list
  const { data: servers } = useQuery<Server[]>({
    queryKey: ['servers'],
    queryFn: () => api.servers.list(),
  });

  // Fetch instance data if editing
  const { data: instance, isLoading } = useQuery<Instance>({
    queryKey: ['instance', id],
    queryFn: () => api.instances.get(id!),
    enabled: isEditMode,
  });

  // Populate form when editing
  useEffect(() => {
    if (instance && isEditMode) {
      setFormData({
        name: instance.name,
        server_id: instance.server_id,
        api_endpoint: instance.api_endpoint,
        username: instance.username,
        password: '', // Don't populate password for security
        partnersystem: instance.partnersystem || '',
        storage_system: instance.storage_system || '',
        description: instance.description || '',
        location: instance.location || '',
        enable_download: instance.enable_download,
        enable_mkquorumapp: instance.enable_mkquorumapp,
        ipquorum_name: instance.ipquorum_name || '',
        ip6: instance.ip6,
        partnerip6: instance.partnerip6,
        nometadata: instance.nometadata,
      });
    }
  }, [instance, isEditMode]);

  // Create mutation
  const createMutation = useMutation({
    mutationFn: (data: any) => api.instances.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      navigate('/instances');
    },
    onError: (error: any) => {
      setSubmitError(error.response?.data?.error || 'Failed to create instance');
    },
  });

  // Update mutation
  const updateMutation = useMutation({
    mutationFn: (data: any) => api.instances.update(id!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances'] });
      queryClient.invalidateQueries({ queryKey: ['instance', id] });
      navigate('/instances');
    },
    onError: (error: any) => {
      setSubmitError(error.response?.data?.error || 'Failed to update instance');
    },
  });

  const validateForm = (): boolean => {
    const newErrors: Partial<Record<keyof FormData, string>> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }

    if (!formData.server_id.trim()) {
      newErrors.server_id = 'Server is required';
    }

    if (!formData.api_endpoint.trim()) {
      newErrors.api_endpoint = 'API endpoint is required';
    }

    if (!formData.username.trim()) {
      newErrors.username = 'Username is required';
    }

    if (!isEditMode && !formData.password.trim()) {
      newErrors.password = 'Password is required';
    }

    if (formData.enable_mkquorumapp && !formData.partnersystem.trim()) {
      newErrors.partnersystem = 'Partner system is required when mkquorumapp is enabled';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitError('');

    if (!validateForm()) {
      return;
    }

    const submitData: any = {
      name: formData.name,
      server_id: formData.server_id,
      api_endpoint: formData.api_endpoint,
      username: formData.username,
      partnersystem: formData.partnersystem,
      storage_system: formData.storage_system,
      description: formData.description,
      location: formData.location,
      enable_download: formData.enable_download,
      enable_mkquorumapp: formData.enable_mkquorumapp,
      ipquorum_name: formData.ipquorum_name,
      ip6: formData.ip6,
      partnerip6: formData.partnerip6,
      nometadata: formData.nometadata,
    };

    // Only include password if provided
    if (formData.password) {
      submitData.password = formData.password;
    }

    if (isEditMode) {
      updateMutation.mutate(submitData);
    } else {
      createMutation.mutate(submitData);
    }
  };

  const handleChange = (field: keyof FormData, value: string | boolean) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }));
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  if (isEditMode && isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"></div>
          <p className="mt-2 text-sm text-gray-600">Loading instance...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
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
          <h1 className="text-3xl font-bold text-gray-900">
            {isEditMode ? 'Edit Instance' : 'Create New Instance'}
          </h1>
          <p className="mt-2 text-gray-600">
            {isEditMode
              ? 'Update instance configuration'
              : 'Configure a new IP Quorum instance'}
          </p>
        </div>
      </div>

      {/* Error Alert */}
      {submitError && (
        <Alert variant="destructive" dismissible onDismiss={() => setSubmitError('')}>
          {submitError}
        </Alert>
      )}

      {/* Form */}
      <form onSubmit={handleSubmit}>
        <Card variant="elevated">
          <CardHeader title="Instance Configuration" />
          <CardBody>
            <div className="space-y-6">
              {/* Basic Information */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-4">Basic Information</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <Input
                    label="Instance Name"
                    value={formData.name}
                    onChange={(e) => handleChange('name', e.target.value)}
                    error={errors.name}
                    placeholder="e.g., production-cluster-01"
                    required
                    fullWidth
                    disabled={isPending}
                  />

                  <div className="space-y-2">
                    <label className="block text-sm font-medium text-gray-700">
                      Target Server <span className="text-red-500">*</span>
                    </label>
                    <select
                      value={formData.server_id}
                      onChange={(e) => handleChange('server_id', e.target.value)}
                      disabled={isPending || isEditMode}
                      className={`w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                        errors.server_id ? 'border-red-500' : 'border-gray-300'
                      } ${isPending || isEditMode ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'}`}
                    >
                      <option value="local">Local (Manager Container)</option>
                      {servers?.filter(s => s.status === 'online').map((server) => (
                        <option key={server.id} value={server.id}>
                          {server.hostname} ({server.ip_address}:{server.agent_port})
                        </option>
                      ))}
                    </select>
                    {errors.server_id && (
                      <p className="text-sm text-red-600">{errors.server_id}</p>
                    )}
                    <p className="text-sm text-gray-500">
                      {isEditMode
                        ? 'Server cannot be changed after creation'
                        : 'Select where this instance will run'}
                    </p>
                  </div>

                  <Input
                    label="API Endpoint"
                    value={formData.api_endpoint}
                    onChange={(e) => handleChange('api_endpoint', e.target.value)}
                    error={errors.api_endpoint}
                    placeholder="e.g., 10.0.0.100"
                    helperText="IP address or hostname of the storage system"
                    required
                    fullWidth
                    disabled={isPending}
                  />

                  <Input
                    label="Storage System"
                    value={formData.storage_system}
                    onChange={(e) => handleChange('storage_system', e.target.value)}
                    placeholder="e.g., SVC-Cluster-01"
                    helperText="Optional: Storage system identifier"
                    fullWidth
                    disabled={isPending}
                  />

                  <Input
                    label="Location"
                    value={formData.location}
                    onChange={(e) => handleChange('location', e.target.value)}
                    placeholder="e.g., Datacenter A"
                    helperText="Optional: Physical location"
                    fullWidth
                    disabled={isPending}
                  />
                </div>

                <div className="mt-4">
                  <Input
                    label="Description"
                    value={formData.description}
                    onChange={(e) => handleChange('description', e.target.value)}
                    placeholder="Optional description"
                    fullWidth
                    disabled={isPending}
                  />
                </div>
              </div>

              {/* Authentication */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-4">Authentication</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <Input
                    label="Username"
                    value={formData.username}
                    onChange={(e) => handleChange('username', e.target.value)}
                    error={errors.username}
                    placeholder="superuser"
                    helperText="Storage system username"
                    required
                    fullWidth
                    disabled={isPending}
                    autoComplete="username"
                  />

                  <Input
                    label="Password"
                    type="password"
                    value={formData.password}
                    onChange={(e) => handleChange('password', e.target.value)}
                    error={errors.password}
                    placeholder={isEditMode ? 'Leave blank to keep current' : 'Enter password'}
                    helperText={isEditMode ? 'Only enter if changing password' : 'Storage system password'}
                    required={!isEditMode}
                    fullWidth
                    disabled={isPending}
                    autoComplete="new-password"
                  />
                </div>
              </div>

              {/* Quorum Configuration */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 mb-4">Quorum Configuration</h3>
                
                <div className="space-y-4">
                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="enable_mkquorumapp"
                      checked={formData.enable_mkquorumapp}
                      onChange={(e) => handleChange('enable_mkquorumapp', e.target.checked)}
                      disabled={isPending}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <label htmlFor="enable_mkquorumapp" className="text-sm font-medium text-gray-700">
                      Enable mkquorumapp (Create quorum application)
                    </label>
                  </div>

                  {formData.enable_mkquorumapp && (
                    <div className="space-y-4 pl-7 border-l-2 border-blue-200">
                      <Input
                        label="Partner System"
                        value={formData.partnersystem}
                        onChange={(e) => handleChange('partnersystem', e.target.value)}
                        error={errors.partnersystem}
                        placeholder="e.g., remote-cluster"
                        helperText="Remote system name for PBHA configuration (required)"
                        required
                        fullWidth
                        disabled={isPending}
                      />

                      <Input
                        label="IPQuorum Application Name"
                        value={formData.ipquorum_name}
                        onChange={(e) => handleChange('ipquorum_name', e.target.value)}
                        placeholder="Leave empty for auto-generated name"
                        helperText="Optional: Custom name for the IPQuorum application"
                        fullWidth
                        disabled={isPending}
                      />

                      <div className="space-y-2">
                        <p className="text-sm font-medium text-gray-700">IPv6 Configuration</p>
                        <div className="flex items-center gap-3">
                          <input
                            type="checkbox"
                            id="ip6"
                            checked={formData.ip6}
                            onChange={(e) => handleChange('ip6', e.target.checked)}
                            disabled={isPending}
                            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                          />
                          <label htmlFor="ip6" className="text-sm text-gray-600">
                            Enable IPv6 for local system
                          </label>
                        </div>
                        <div className="flex items-center gap-3">
                          <input
                            type="checkbox"
                            id="partnerip6"
                            checked={formData.partnerip6}
                            onChange={(e) => handleChange('partnerip6', e.target.checked)}
                            disabled={isPending}
                            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                          />
                          <label htmlFor="partnerip6" className="text-sm text-gray-600">
                            Enable IPv6 for partner system
                          </label>
                        </div>
                      </div>

                      <div className="flex items-center gap-3">
                        <input
                          type="checkbox"
                          id="nometadata"
                          checked={formData.nometadata}
                          onChange={(e) => handleChange('nometadata', e.target.checked)}
                          disabled={isPending}
                          className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                        />
                        <label htmlFor="nometadata" className="text-sm text-gray-600">
                          Disable metadata collection
                        </label>
                      </div>
                    </div>
                  )}

                  <div className="flex items-center gap-3">
                    <input
                      type="checkbox"
                      id="enable_download"
                      checked={formData.enable_download}
                      onChange={(e) => handleChange('enable_download', e.target.checked)}
                      disabled={isPending}
                      className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                    />
                    <label htmlFor="enable_download" className="text-sm font-medium text-gray-700">
                      Enable download (Download ip_quorum.jar)
                    </label>
                  </div>
                </div>
              </div>
            </div>
          </CardBody>

          <CardFooter align="right">
            <Button
              type="button"
              variant="ghost"
              onClick={() => navigate('/instances')}
              disabled={isPending}
              leftIcon={<X className="h-4 w-4" />}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="default"
              isLoading={isPending}
              leftIcon={<Save className="h-4 w-4" />}
            >
              {isEditMode ? 'Update Instance' : 'Create Instance'}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </div>
  );
};

export default InstanceForm;


