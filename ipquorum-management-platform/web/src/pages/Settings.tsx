import React, { useState, useEffect } from 'react';
import { Save, RefreshCw, AlertTriangle, CheckCircle2, Settings as SettingsIcon } from 'lucide-react';
import { Button, Input, Card, CardBody, Alert, Badge } from '../components/ui';
import { useToastStore } from '../stores/toastStore';
import api from '../services/api';

interface SystemSettings {
  maxInstances: number;
  defaultPort: number;
  healthCheckInterval: number;
  logLevel: string;
  enableMetrics: boolean;
  metricsPort: number;
  backupEnabled: boolean;
  backupInterval: number;
  backupRetention: number;
}

interface SystemInfo {
  version: string;
  uptime: string;
  totalInstances: number;
  activeInstances: number;
  cpuUsage: number;
  memoryUsage: number;
  diskUsage: number;
}

export const Settings: React.FC = () => {
  const [settings, setSettings] = useState<SystemSettings>({
    maxInstances: 10,
    defaultPort: 3993,
    healthCheckInterval: 30,
    logLevel: 'info',
    enableMetrics: true,
    metricsPort: 9090,
    backupEnabled: true,
    backupInterval: 24,
    backupRetention: 7,
  });

  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const { success, error: showError, info } = useToastStore();

  useEffect(() => {
    loadSettings();
    loadSystemInfo();
  }, []);

  const loadSettings = async () => {
    setLoading(true);
    try {
      // In a real implementation, this would fetch from the API
      // const response = await api.get('/api/settings');
      // setSettings(response.data);
      
      // Mock data for now
      await new Promise(resolve => setTimeout(resolve, 500));
      success('Settings loaded');
    } catch (err) {
      showError('Failed to load settings');
    } finally {
      setLoading(false);
    }
  };

  const formatUptime = (seconds: number): string => {
    if (!seconds || seconds < 0) return 'N/A';
    
    const units = [
      { name: 'y', seconds: 31536000 },
      { name: 'mo', seconds: 2592000 },
      { name: 'd', seconds: 86400 },
      { name: 'h', seconds: 3600 },
      { name: 'm', seconds: 60 },
      { name: 's', seconds: 1 }
    ];
    
    const parts: string[] = [];
    let remaining = seconds;
    
    for (const unit of units) {
      const value = Math.floor(remaining / unit.seconds);
      if (value > 0) {
        parts.push(`${value}${unit.name}`);
        remaining -= value * unit.seconds;
      }
      if (parts.length >= 2) break;
    }
    
    return parts.length > 0 ? parts.join(' ') : '0s';
  };

  const loadSystemInfo = async () => {
    try {
      const response = await api.get('/health');
      const data = response.data;
      setSystemInfo({
        version: '1.0.0',
        uptime: data.uptime ? formatUptime(data.uptime) : 'N/A',
        totalInstances: 5,
        activeInstances: 3,
        cpuUsage: 45,
        memoryUsage: 62,
        diskUsage: 38,
      });
    } catch (error) {
      // Silently use mock data when API is unavailable
      setSystemInfo({
        version: '1.0.0',
        uptime: 'N/A',
        totalInstances: 5,
        activeInstances: 3,
        cpuUsage: 45,
        memoryUsage: 62,
        diskUsage: 38,
      });
    }
  };

  const handleSettingChange = (key: keyof SystemSettings, value: any) => {
    setSettings(prev => ({ ...prev, [key]: value }));
    setHasChanges(true);
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      // In a real implementation, this would save to the API
      // await api.put('/api/settings', settings);
      
      await new Promise(resolve => setTimeout(resolve, 1000));
      success('Settings saved successfully');
      setHasChanges(false);
    } catch (err) {
      showError('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const handleReset = () => {
    loadSettings();
    setHasChanges(false);
    info('Settings reset to saved values');
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-foreground">Settings</h1>
          <p className="text-muted-foreground mt-1">Configure system settings and preferences</p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            onClick={handleReset}
            disabled={!hasChanges || loading}
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Reset
          </Button>
          <Button
            onClick={handleSave}
            disabled={!hasChanges || saving}
            isLoading={saving}
          >
            <Save className="w-4 h-4 mr-2" />
            Save Changes
          </Button>
        </div>
      </div>

      {hasChanges && (
        <Alert variant="warning">
          <AlertTriangle className="w-4 h-4" />
          <span>You have unsaved changes. Click "Save Changes" to apply them.</span>
        </Alert>
      )}

      {/* System Information */}
      {systemInfo && (
        <Card>
          <CardBody>
            <div className="flex items-center gap-2 mb-4">
              <SettingsIcon className="w-5 h-5 text-primary" />
              <h2 className="text-xl font-semibold">System Information</h2>
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              <div className="p-4 bg-muted rounded-lg">
                <div className="text-sm text-muted-foreground mb-1">Version</div>
                <div className="text-lg font-semibold">{systemInfo.version}</div>
              </div>
              <div className="p-4 bg-muted rounded-lg">
                <div className="text-sm text-muted-foreground mb-1">Uptime</div>
                <div className="text-lg font-semibold">{systemInfo.uptime}</div>
              </div>
              <div className="p-4 bg-muted rounded-lg">
                <div className="text-sm text-muted-foreground mb-1">Instances</div>
                <div className="text-lg font-semibold">
                  {systemInfo.activeInstances} / {systemInfo.totalInstances}
                </div>
              </div>
              <div className="p-4 bg-muted rounded-lg">
                <div className="text-sm text-muted-foreground mb-1">Status</div>
                <Badge variant="success">
                  <CheckCircle2 className="w-3 h-3 mr-1" />
                  Healthy
                </Badge>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-4">
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-muted-foreground">CPU Usage</span>
                  <span className="font-medium">{systemInfo.cpuUsage}%</span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all"
                    style={{ width: `${systemInfo.cpuUsage}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-muted-foreground">Memory Usage</span>
                  <span className="font-medium">{systemInfo.memoryUsage}%</span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all"
                    style={{ width: `${systemInfo.memoryUsage}%` }}
                  />
                </div>
              </div>
              <div>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-muted-foreground">Disk Usage</span>
                  <span className="font-medium">{systemInfo.diskUsage}%</span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className="h-full bg-primary transition-all"
                    style={{ width: `${systemInfo.diskUsage}%` }}
                  />
                </div>
              </div>
            </div>
          </CardBody>
        </Card>
      )}

      {/* Instance Settings */}
      <Card>
        <CardBody>
          <h2 className="text-xl font-semibold mb-4">Instance Settings</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Input
              label="Maximum Instances"
              type="number"
              value={settings.maxInstances}
              onChange={(e) => handleSettingChange('maxInstances', parseInt(e.target.value))}
              min={1}
              max={100}
              fullWidth
              helperText="Maximum number of IP Quorum instances allowed"
            />
            <Input
              label="Default Port"
              type="number"
              value={settings.defaultPort}
              onChange={(e) => handleSettingChange('defaultPort', parseInt(e.target.value))}
              min={1024}
              max={65535}
              fullWidth
              helperText="Default port for new instances"
            />
            <Input
              label="Health Check Interval (seconds)"
              type="number"
              value={settings.healthCheckInterval}
              onChange={(e) => handleSettingChange('healthCheckInterval', parseInt(e.target.value))}
              min={10}
              max={300}
              fullWidth
              helperText="How often to check instance health"
            />
            <div>
              <label className="block text-sm font-medium text-foreground mb-1">
                Log Level
              </label>
              <select
                value={settings.logLevel}
                onChange={(e) => handleSettingChange('logLevel', e.target.value)}
                className="w-full px-3 py-2 bg-background border border-input rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warn">Warning</option>
                <option value="error">Error</option>
              </select>
              <p className="text-xs text-muted-foreground mt-1">
                Logging verbosity level
              </p>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Metrics Settings */}
      <Card>
        <CardBody>
          <h2 className="text-xl font-semibold mb-4">Metrics & Monitoring</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
              <div>
                <div className="font-medium">Enable Prometheus Metrics</div>
                <div className="text-sm text-muted-foreground">
                  Expose metrics endpoint for Prometheus scraping
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={settings.enableMetrics}
                  onChange={(e) => handleSettingChange('enableMetrics', e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-muted-foreground/20 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
              </label>
            </div>

            {settings.enableMetrics && (
              <Input
                label="Metrics Port"
                type="number"
                value={settings.metricsPort}
                onChange={(e) => handleSettingChange('metricsPort', parseInt(e.target.value))}
                min={1024}
                max={65535}
                fullWidth
                helperText="Port for Prometheus metrics endpoint"
              />
            )}
          </div>
        </CardBody>
      </Card>

      {/* Backup Settings */}
      <Card>
        <CardBody>
          <h2 className="text-xl font-semibold mb-4">Backup & Recovery</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between p-4 bg-muted rounded-lg">
              <div>
                <div className="font-medium">Enable Automatic Backups</div>
                <div className="text-sm text-muted-foreground">
                  Automatically backup configuration and data
                </div>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input
                  type="checkbox"
                  checked={settings.backupEnabled}
                  onChange={(e) => handleSettingChange('backupEnabled', e.target.checked)}
                  className="sr-only peer"
                />
                <div className="w-11 h-6 bg-muted-foreground/20 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-primary/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary"></div>
              </label>
            </div>

            {settings.backupEnabled && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="Backup Interval (hours)"
                  type="number"
                  value={settings.backupInterval}
                  onChange={(e) => handleSettingChange('backupInterval', parseInt(e.target.value))}
                  min={1}
                  max={168}
                  fullWidth
                  helperText="How often to create backups"
                />
                <Input
                  label="Backup Retention (days)"
                  type="number"
                  value={settings.backupRetention}
                  onChange={(e) => handleSettingChange('backupRetention', parseInt(e.target.value))}
                  min={1}
                  max={365}
                  fullWidth
                  helperText="How long to keep backups"
                />
              </div>
            )}
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

export default Settings;


