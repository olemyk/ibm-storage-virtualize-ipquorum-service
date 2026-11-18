import React, { useState, useEffect } from 'react';
import { ipquorumApi, ipquorumHelpers } from '../api/ipquorum';
import { api } from '../services/api';
import type {
  CreateIPQuorumContainerRequest,
  IPQuorumConfig,
  IPQuorumFormErrors,
} from '../types/ipquorum';
import type { Server } from '../types/server';
import {
  DEFAULT_IPQUORUM_CONFIG,
  DEFAULT_RESTART_POLICY,
  DEFAULT_IMAGE_NAME,
  DEFAULT_IMAGE_TAG,
} from '../types/ipquorum';

interface Props {
  agentId?: string;
  onSuccess?: (containerId: string) => void;
  onCancel?: () => void;
}

export const IPQuorumContainerForm: React.FC<Props> = ({
  agentId: initialAgentId,
  onSuccess,
  onCancel,
}) => {
  // Form state
  const [name, setName] = useState('');
  const [agentId, setAgentId] = useState(initialAgentId || '');
  const [config, setConfig] = useState<IPQuorumConfig>({
    ...DEFAULT_IPQUORUM_CONFIG,
    api_endpoint: '',
    username: '',
    password: '',
    partner_system: '',
  } as IPQuorumConfig);

  // Advanced options
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [restartPolicy, setRestartPolicy] = useState(DEFAULT_RESTART_POLICY);
  const [imageName, setImageName] = useState(DEFAULT_IMAGE_NAME);
  const [imageTag, setImageTag] = useState(DEFAULT_IMAGE_TAG);

  // UI state
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<IPQuorumFormErrors>({});
  const [generalError, setGeneralError] = useState<string | null>(null);
  const [showPassword, setShowPassword] = useState(false);

  // Agent selection state
  const [availableAgents, setAvailableAgents] = useState<Server[]>([]);
  const [loadingAgents, setLoadingAgents] = useState(false);

  // Fetch available agents on mount
  useEffect(() => {
    const fetchAgents = async () => {
      setLoadingAgents(true);
      try {
        const servers = await api.servers.list();
        setAvailableAgents(servers || []);
        
        // If no agent is pre-selected and we have agents, select the first online one
        if (!initialAgentId && servers && servers.length > 0) {
          const onlineAgent = servers.find((s: Server) => s.status === 'online');
          if (onlineAgent) {
            setAgentId(onlineAgent.id);
          }
        }
      } catch (err) {
        console.error('Failed to fetch agents:', err);
        setGeneralError('Failed to load available agents. Please refresh the page.');
      } finally {
        setLoadingAgents(false);
      }
    };

    fetchAgents();
  }, [initialAgentId]);

  // Validate form
  const validateForm = (): boolean => {
    const newErrors: IPQuorumFormErrors = {};

    if (!name.trim()) {
      newErrors.name = 'Container name is required';
    }

    if (!agentId.trim()) {
      newErrors.agent_id = 'Agent ID is required';
    }

    // Validate IPQuorum config
    const configErrors = ipquorumHelpers.validateConfig(config);
    if (configErrors.length > 0) {
      configErrors.forEach((error) => {
        if (error.includes('endpoint')) newErrors.api_endpoint = error;
        if (error.includes('username')) newErrors.username = error;
        if (error.includes('password')) newErrors.password = error;
        if (error.includes('partner')) newErrors.partner_system = error;
        if (error.includes('port')) newErrors.port = error;
      });
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // Handle form submission
  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setGeneralError(null);

    if (!validateForm()) {
      return;
    }

    setLoading(true);

    try {
      const request: CreateIPQuorumContainerRequest = {
        name,
        agent_id: agentId,
        ipquorum_config: config,
        restart_policy: restartPolicy,
        image_name: imageName,
        image_tag: imageTag,
      };

      const response = await ipquorumApi.createContainer(request);
      
      if (onSuccess && response.container) {
        onSuccess(response.container.id);
      }
    } catch (err: any) {
      setGeneralError(
        err.response?.data?.message || 
        err.message || 
        'Failed to create IPQuorum container'
      );
    } finally {
      setLoading(false);
    }
  };

  // Update config field
  const updateConfig = (field: keyof IPQuorumConfig, value: any) => {
    setConfig((prev) => ({ ...prev, [field]: value }));
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Create IPQuorum Container
        </h2>
        <p className="text-sm text-gray-600">
          Deploy an IBM Storage Virtualize IP Quorum application in a container
        </p>
      </div>

      {/* General Error */}
      {generalError && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-red-400"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-red-800">{generalError}</p>
            </div>
          </div>
        </div>
      )}

      {/* Basic Configuration */}
      <div className="bg-white shadow rounded-lg p-6 space-y-4">
        <h3 className="text-lg font-medium text-gray-900">Basic Configuration</h3>

        {/* Container Name */}
        <div>
          <label className="block text-sm font-medium text-gray-700">
            Container Name *
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="ipquorum-prod-01"
            className={`mt-1 block w-full rounded-md shadow-sm ${
              errors.name
                ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
            }`}
          />
          {errors.name && (
            <p className="mt-1 text-sm text-red-600">{errors.name}</p>
          )}
        </div>

        {/* Agent Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-700">
            Agent Server *
          </label>
          {loadingAgents ? (
            <div className="mt-1 block w-full rounded-md border border-gray-300 bg-gray-50 px-3 py-2 text-sm text-gray-500">
              Loading available agents...
            </div>
          ) : availableAgents.length === 0 ? (
            <div className="mt-1">
              <div className="block w-full rounded-md border border-yellow-300 bg-yellow-50 px-3 py-2 text-sm text-yellow-800">
                No agents available. Please register an agent server first.
              </div>
            </div>
          ) : (
            <select
              value={agentId}
              onChange={(e) => setAgentId(e.target.value)}
              disabled={!!initialAgentId}
              className={`mt-1 block w-full rounded-md shadow-sm ${
                errors.agent_id
                  ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                  : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
              } ${initialAgentId ? 'bg-gray-100' : ''}`}
            >
              <option value="">Select an agent server...</option>
              {availableAgents.map((agent) => (
                <option key={agent.id} value={agent.id}>
                  {agent.hostname} ({agent.ip_address}) - {agent.status}
                </option>
              ))}
            </select>
          )}
          <p className="mt-1 text-sm text-gray-500">
            Select the agent server where the container will be deployed
          </p>
          {errors.agent_id && (
            <p className="mt-1 text-sm text-red-600">{errors.agent_id}</p>
          )}
        </div>
      </div>

      {/* Storage Virtualize Connection */}
      <div className="bg-white shadow rounded-lg p-6 space-y-4">
        <h3 className="text-lg font-medium text-gray-900">
          Storage Virtualize Connection
        </h3>

        {/* API Endpoint */}
        <div>
          <label className="block text-sm font-medium text-gray-700">
            API Endpoint *
          </label>
          <input
            type="text"
            value={config.api_endpoint}
            onChange={(e) => updateConfig('api_endpoint', e.target.value)}
            placeholder="10.33.7.80"
            className={`mt-1 block w-full rounded-md shadow-sm ${
              errors.api_endpoint
                ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
            }`}
          />
          <p className="mt-1 text-sm text-gray-500">
            IP address or hostname of Storage Virtualize system
          </p>
          {errors.api_endpoint && (
            <p className="mt-1 text-sm text-red-600">{errors.api_endpoint}</p>
          )}
        </div>

        {/* Username */}
        <div>
          <label className="block text-sm font-medium text-gray-700">
            Username *
          </label>
          <input
            type="text"
            value={config.username}
            onChange={(e) => updateConfig('username', e.target.value)}
            placeholder="superuser"
            className={`mt-1 block w-full rounded-md shadow-sm ${
              errors.username
                ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
            }`}
          />
          <p className="mt-1 text-sm text-gray-500">
            Monitor role for download or Restricted Administrator for creating quorum app
          </p>
          {errors.username && (
            <p className="mt-1 text-sm text-red-600">{errors.username}</p>
          )}
        </div>

        {/* Password */}
        <div>
          <label className="block text-sm font-medium text-gray-700">
            Password *
          </label>
          <div className="mt-1 relative">
            <input
              type={showPassword ? 'text' : 'password'}
              value={config.password}
              onChange={(e) => updateConfig('password', e.target.value)}
              className={`block w-full rounded-md shadow-sm pr-10 ${
                errors.password
                  ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                  : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
              }`}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute inset-y-0 right-0 pr-3 flex items-center"
            >
              {showPassword ? (
                <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
                </svg>
              ) : (
                <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                </svg>
              )}
            </button>
          </div>
          {errors.password && (
            <p className="mt-1 text-sm text-red-600">{errors.password}</p>
          )}
        </div>

        {/* Insecure TLS */}
        <div className="flex items-center">
          <input
            type="checkbox"
            checked={config.insecure}
            onChange={(e) => updateConfig('insecure', e.target.checked)}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <label className="ml-2 block text-sm text-gray-700">
            Allow insecure TLS (skip certificate verification)
          </label>
        </div>
      </div>

      {/* Quorum App Configuration */}
      <div className="bg-white shadow rounded-lg p-6 space-y-4">
        <h3 className="text-lg font-medium text-gray-900">
          Quorum App Configuration
        </h3>

        {/* Create Quorum App */}
        <div className="flex items-center">
          <input
            type="checkbox"
            checked={config.create_quorum}
            onChange={(e) => updateConfig('create_quorum', e.target.checked)}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <label className="ml-2 block text-sm text-gray-700">
            Create new Quorum App (requires Restricted Administrator role)
          </label>
        </div>

        {/* Partner System (conditional) */}
        {config.create_quorum && (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Partner System *
              </label>
              <input
                type="text"
                value={config.partner_system}
                onChange={(e) => updateConfig('partner_system', e.target.value)}
                placeholder="svc_cluster02"
                className={`mt-1 block w-full rounded-md shadow-sm ${
                  errors.partner_system
                    ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                    : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
                }`}
              />
              <p className="mt-1 text-sm text-gray-500">
                Remote system name in PBHA configuration
              </p>
              {errors.partner_system && (
                <p className="mt-1 text-sm text-red-600">{errors.partner_system}</p>
              )}
            </div>

            {/* IPv6 Options */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-center">
                <input
                  type="checkbox"
                  checked={config.ip6}
                  onChange={(e) => updateConfig('ip6', e.target.checked)}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label className="ml-2 block text-sm text-gray-700">
                  IPv6 for local system
                </label>
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  checked={config.partnerip6}
                  onChange={(e) => updateConfig('partnerip6', e.target.checked)}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label className="ml-2 block text-sm text-gray-700">
                  IPv6 for partner system
                </label>
              </div>

              <div className="flex items-center">
                <input
                  type="checkbox"
                  checked={config.nometadata}
                  onChange={(e) => updateConfig('nometadata', e.target.checked)}
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label className="ml-2 block text-sm text-gray-700">
                  No metadata
                </label>
              </div>
            </div>
          </>
        )}
      </div>

      {/* Network Configuration */}
      <div className="bg-white shadow rounded-lg p-6 space-y-4">
        <h3 className="text-lg font-medium text-gray-900">
          Network Configuration
        </h3>

        {/* Port */}
        <div>
          <label className="block text-sm font-medium text-gray-700">
            IPQuorum Port
          </label>
          <input
            type="number"
            value={config.port}
            onChange={(e) => updateConfig('port', parseInt(e.target.value))}
            min="1024"
            max="65535"
            className={`mt-1 block w-full rounded-md shadow-sm ${
              errors.port
                ? 'border-red-300 focus:border-red-500 focus:ring-red-500'
                : 'border-gray-300 focus:border-blue-500 focus:ring-blue-500'
            }`}
          />
          <p className="mt-1 text-sm text-gray-500">
            Default: 1260
          </p>
          {errors.port && (
            <p className="mt-1 text-sm text-red-600">{errors.port}</p>
          )}
        </div>
      </div>

      {/* Download Options */}
      <div className="bg-white shadow rounded-lg p-6 space-y-4">
        <h3 className="text-lg font-medium text-gray-900">
          Download Options
        </h3>

        <div className="flex items-center">
          <input
            type="checkbox"
            checked={config.download_jar}
            onChange={(e) => updateConfig('download_jar', e.target.checked)}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <label className="ml-2 block text-sm text-gray-700">
            Download ip_quorum.jar on container start (if not present)
          </label>
        </div>

        <div className="flex items-center">
          <input
            type="checkbox"
            checked={config.force_download || false}
            onChange={(e) => updateConfig('force_download', e.target.checked)}
            className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
          />
          <label className="ml-2 block text-sm text-gray-700">
            Force re-download even if JAR exists
          </label>
        </div>
      </div>

      {/* Advanced Options (collapsible) */}
      <div className="bg-white shadow rounded-lg p-6">
        <button
          type="button"
          onClick={() => setShowAdvanced(!showAdvanced)}
          className="flex items-center justify-between w-full text-left"
        >
          <h3 className="text-lg font-medium text-gray-900">
            Advanced Options
          </h3>
          <svg
            className={`h-5 w-5 text-gray-500 transform transition-transform ${
              showAdvanced ? 'rotate-180' : ''
            }`}
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 9l-7 7-7-7"
            />
          </svg>
        </button>

        {showAdvanced && (
          <div className="mt-4 space-y-4">
            {/* Restart Policy */}
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Restart Policy
              </label>
              <select
                value={restartPolicy}
                onChange={(e) => setRestartPolicy(e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              >
                <option value="no">No</option>
                <option value="always">Always</option>
                <option value="on-failure">On Failure</option>
                <option value="unless-stopped">Unless Stopped</option>
              </select>
            </div>

            {/* Image Name */}
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Image Name
              </label>
              <input
                type="text"
                value={imageName}
                onChange={(e) => setImageName(e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              />
            </div>

            {/* Image Tag */}
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Image Tag
              </label>
              <input
                type="text"
                value={imageTag}
                onChange={(e) => setImageTag(e.target.value)}
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              />
            </div>

            {/* Java Options */}
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Java Options
              </label>
              <input
                type="text"
                value={config.java_opts || ''}
                onChange={(e) => updateConfig('java_opts', e.target.value)}
                placeholder="-Xmx512m -Xms256m"
                className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
              />
              <p className="mt-1 text-sm text-gray-500">
                JVM memory and performance options
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Actions */}
      <div className="flex justify-end space-x-3">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={loading}
            className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            Cancel
          </button>
        )}
        <button
          type="submit"
          disabled={loading}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 flex items-center"
        >
          {loading ? (
            <>
              <svg
                className="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
                fill="none"
                viewBox="0 0 24 24"
              >
                <circle
                  className="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  strokeWidth="4"
                />
                <path
                  className="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                />
              </svg>
              Creating...
            </>
          ) : (
            'Create Container'
          )}
        </button>
      </div>
    </form>
  );
};

export default IPQuorumContainerForm;

// Made with Bob
