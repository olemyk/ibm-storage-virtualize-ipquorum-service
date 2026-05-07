import React from 'react';
import { HelpCircle, Clock, Activity, Globe, Server, RefreshCw } from 'lucide-react';
import { Card, CardHeader, CardBody } from '../components/ui';

const Help: React.FC = () => {
  return (
    <div className="max-w-4xl mx-auto space-y-6">
      <div className="flex items-center gap-3 mb-6">
        <HelpCircle className="h-8 w-8 text-primary" />
        <h1 className="text-3xl font-bold">Help & Documentation</h1>
      </div>

      {/* Timezone Handling */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Clock className="h-5 w-5 text-primary" />
            <h2 className="text-xl font-semibold">Timezone Handling</h2>
          </div>
        </CardHeader>
        <CardBody>
          <div className="space-y-4 text-muted-foreground">
            <div>
              <h3 className="font-semibold text-foreground mb-2">How Timezones Work</h3>
              <p className="mb-2">
                The IPQuorum Management Platform uses a standardized approach to handle timezones:
              </p>
              <ul className="list-disc list-inside space-y-2 ml-4">
                <li>
                  <strong className="text-foreground">Backend Storage:</strong> All timestamps are stored in UTC (Coordinated Universal Time) in the database
                </li>
                <li>
                  <strong className="text-foreground">API Format:</strong> The API returns timestamps in ISO 8601 format with timezone information (e.g., "2026-05-06T20:24:30Z")
                </li>
                <li>
                  <strong className="text-foreground">Frontend Display:</strong> Your browser automatically converts UTC timestamps to your local timezone
                </li>
              </ul>
            </div>

            <div className="bg-muted p-4 rounded-lg">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <Globe className="h-4 w-4" />
                Example: UTC vs Local Time
              </h4>
              <div className="space-y-1 text-sm">
                <p><strong>Server (UTC):</strong> 2026-05-06 20:24:30 UTC</p>
                <p><strong>Oslo (CEST, UTC+2):</strong> 2026-05-06 22:24:30 CEST</p>
                <p><strong>New York (EDT, UTC-4):</strong> 2026-05-06 16:24:30 EDT</p>
                <p><strong>Tokyo (JST, UTC+9):</strong> 2026-05-07 05:24:30 JST</p>
              </div>
            </div>

            <div>
              <h4 className="font-semibold text-foreground mb-2">Why UTC?</h4>
              <p>
                Using UTC as the standard ensures consistency across different servers, locations, and timezones. 
                This is especially important when managing IPQuorum instances across multiple geographic locations 
                or when team members are in different timezones.
              </p>
            </div>

            <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 p-4 rounded-lg">
              <p className="text-sm">
                <strong className="text-foreground">Note:</strong> All times displayed in the web interface 
                (Last Health Check, Started At, etc.) are automatically converted to your browser's timezone. 
                You don't need to do any manual conversion.
              </p>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Health Check System */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Activity className="h-5 w-5 text-primary" />
            <h2 className="text-xl font-semibold">Health Check System</h2>
          </div>
        </CardHeader>
        <CardBody>
          <div className="space-y-4 text-muted-foreground">
            <div>
              <h3 className="font-semibold text-foreground mb-2">How Health Checks Work</h3>
              <p className="mb-2">
                The platform continuously monitors the health of all IPQuorum instances:
              </p>
              <ul className="list-disc list-inside space-y-2 ml-4">
                <li>
                  <strong className="text-foreground">Check Interval:</strong> Health checks run every 30 seconds automatically
                </li>
                <li>
                  <strong className="text-foreground">Local Instances:</strong> The manager directly checks the process status on the local server
                </li>
                <li>
                  <strong className="text-foreground">Remote Instances:</strong> The manager communicates with the agent on the remote server to check status
                </li>
                <li>
                  <strong className="text-foreground">Status Updates:</strong> Instance status, uptime, and last health check timestamp are updated in real-time
                </li>
              </ul>
            </div>

            <div className="bg-muted p-4 rounded-lg">
              <h4 className="font-semibold text-foreground mb-2 flex items-center gap-2">
                <RefreshCw className="h-4 w-4" />
                Health Check Flow
              </h4>
              <div className="space-y-2 text-sm">
                <div className="flex items-start gap-2">
                  <span className="font-mono bg-primary/10 text-primary px-2 py-1 rounded text-xs">1</span>
                  <p>Manager initiates health check every 30 seconds</p>
                </div>
                <div className="flex items-start gap-2">
                  <span className="font-mono bg-primary/10 text-primary px-2 py-1 rounded text-xs">2</span>
                  <p>For local instances: Check process status directly</p>
                </div>
                <div className="flex items-start gap-2">
                  <span className="font-mono bg-primary/10 text-primary px-2 py-1 rounded text-xs">3</span>
                  <p>For remote instances: Send request to agent on remote server</p>
                </div>
                <div className="flex items-start gap-2">
                  <span className="font-mono bg-primary/10 text-primary px-2 py-1 rounded text-xs">4</span>
                  <p>Update database with current status, uptime, and timestamp</p>
                </div>
                <div className="flex items-start gap-2">
                  <span className="font-mono bg-primary/10 text-primary px-2 py-1 rounded text-xs">5</span>
                  <p>Web UI automatically refreshes to show latest data</p>
                </div>
              </div>
            </div>

            <div>
              <h4 className="font-semibold text-foreground mb-2">Status Indicators</h4>
              <div className="space-y-2">
                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                    <span className="h-1.5 w-1.5 rounded-full bg-green-600 dark:bg-green-400"></span>
                    Running
                  </span>
                  <p className="text-sm">Instance is active and healthy</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200">
                    <span className="h-1.5 w-1.5 rounded-full bg-gray-600 dark:bg-gray-400"></span>
                    Stopped
                  </span>
                  <p className="text-sm">Instance is not running</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
                    <span className="h-1.5 w-1.5 rounded-full bg-red-600 dark:bg-red-400"></span>
                    Failed
                  </span>
                  <p className="text-sm">Instance encountered an error</p>
                </div>
              </div>
            </div>

            <div>
              <h4 className="font-semibold text-foreground mb-2">Displayed Information</h4>
              <ul className="list-disc list-inside space-y-2 ml-4">
                <li>
                  <strong className="text-foreground">Uptime:</strong> Shows how long the instance has been running (e.g., "6h 14m")
                </li>
                <li>
                  <strong className="text-foreground">Last Health Check:</strong> Timestamp of the most recent health check (in your local timezone)
                </li>
                <li>
                  <strong className="text-foreground">Started At:</strong> When the instance was last started (in your local timezone)
                </li>
              </ul>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* System Architecture */}
      <Card>
        <CardHeader>
          <div className="flex items-center gap-2">
            <Server className="h-5 w-5 text-primary" />
            <h2 className="text-xl font-semibold">System Architecture</h2>
          </div>
        </CardHeader>
        <CardBody>
          <div className="space-y-4 text-muted-foreground">
            <div>
              <h3 className="font-semibold text-foreground mb-2">Manager-Agent Architecture</h3>
              <p className="mb-2">
                The IPQuorum Management Platform uses a distributed architecture:
              </p>
              <ul className="list-disc list-inside space-y-2 ml-4">
                <li>
                  <strong className="text-foreground">Manager:</strong> Central control plane that manages all instances and provides the web UI
                </li>
                <li>
                  <strong className="text-foreground">Agent:</strong> Lightweight service running on each server that manages local IPQuorum instances
                </li>
                <li>
                  <strong className="text-foreground">Communication:</strong> Manager and agents communicate via REST API with API key authentication
                </li>
                <li>
                  <strong className="text-foreground">Database:</strong> SQLite database stores instance configurations, status, and metadata
                </li>
              </ul>
            </div>

            <div className="bg-muted p-4 rounded-lg">
              <h4 className="font-semibold text-foreground mb-2">Key Features</h4>
              <ul className="list-disc list-inside space-y-1 text-sm ml-4">
                <li>Centralized management of multiple IPQuorum instances</li>
                <li>Support for both local and remote instances</li>
                <li>Real-time health monitoring and status updates</li>
                <li>Secure API key-based authentication</li>
                <li>Role-based access control (Admin, Operator, Viewer)</li>
                <li>Automatic health checks every 30 seconds</li>
                <li>Web-based user interface for easy management</li>
              </ul>
            </div>
          </div>
        </CardBody>
      </Card>

      {/* Additional Resources */}
      <Card>
        <CardHeader>
          <h2 className="text-xl font-semibold">Additional Resources</h2>
        </CardHeader>
        <CardBody>
          <div className="space-y-2 text-muted-foreground">
            <p>
              For more detailed information about the IPQuorum Management Platform, please refer to:
            </p>
            <ul className="list-disc list-inside space-y-1 ml-4">
              <li>Architecture documentation in the project repository</li>
              <li>API documentation for integration</li>
              <li>Deployment guides for manager and agent setup</li>
            </ul>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

export default Help;

