import React, { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';
import { Button, Card, CardHeader, CardBody, Alert } from './ui';
import { Copy, Download, RefreshCw, Terminal, X } from 'lucide-react';

interface LogsViewerProps {
  serverId: string;
  containerId: string;
  containerName?: string;
  onClose?: () => void;
}

export const LogsViewer: React.FC<LogsViewerProps> = ({
  serverId,
  containerId,
  containerName,
  onClose,
}) => {
  const [tail, setTail] = useState(100);
  const [autoScroll, setAutoScroll] = useState(true);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const logsContainerRef = useRef<HTMLDivElement>(null);

  // Fetch logs
  const { data, isLoading, error, refetch } = useQuery({
    queryKey: ['container-logs', serverId, containerId, tail],
    queryFn: () => api.servers.getContainerLogs(serverId, containerId, tail),
    refetchInterval: 5000, // Auto-refresh every 5 seconds
  });

  // Auto-scroll to bottom when logs update
  useEffect(() => {
    if (autoScroll && logsEndRef.current) {
      logsEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [data, autoScroll]);

  // Handle manual scroll - disable auto-scroll if user scrolls up
  const handleScroll = () => {
    if (!logsContainerRef.current) return;
    
    const { scrollTop, scrollHeight, clientHeight } = logsContainerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;
    
    if (!isAtBottom && autoScroll) {
      setAutoScroll(false);
    } else if (isAtBottom && !autoScroll) {
      setAutoScroll(true);
    }
  };

  // Copy logs to clipboard
  const handleCopy = async () => {
    if (data?.logs) {
      try {
        await navigator.clipboard.writeText(data.logs);
        alert('Logs copied to clipboard!');
      } catch (err) {
        console.error('Failed to copy logs:', err);
        alert('Failed to copy logs to clipboard');
      }
    }
  };

  // Download logs as file
  const handleDownload = () => {
    if (data?.logs) {
      const blob = new Blob([data.logs], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${containerName || containerId}-logs-${new Date().toISOString()}.txt`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  };

  // Handle tail change
  const handleTailChange = (newTail: number) => {
    setTail(newTail);
  };

  return (
    <Card className="w-full">
      <CardHeader
        title="Container Logs"
        subtitle={`${containerName || containerId} - Last ${tail} lines`}
        action={
          onClose ? (
            <Button variant="ghost" size="sm" onClick={onClose}>
              <X className="h-4 w-4" />
            </Button>
          ) : undefined
        }
      />
      <CardBody>
        {/* Controls */}
        <div className="flex flex-wrap items-center gap-2 mb-4">
          <div className="flex items-center gap-2">
            <span className="text-sm text-gray-600">Lines:</span>
            {[50, 100, 200, 500, 1000].map((num) => (
              <Button
                key={num}
                variant={tail === num ? 'default' : 'outline'}
                size="sm"
                onClick={() => handleTailChange(num)}
              >
                {num}
              </Button>
            ))}
          </div>
          
          <div className="flex-1" />
          
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoScroll(!autoScroll)}
            className={autoScroll ? 'bg-green-50' : ''}
          >
            {autoScroll ? '✓ Auto-scroll' : 'Auto-scroll'}
          </Button>
          
          <Button
            variant="outline"
            size="sm"
            onClick={() => refetch()}
            disabled={isLoading}
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
          </Button>
          
          <Button
            variant="outline"
            size="sm"
            onClick={handleCopy}
            disabled={!data?.logs}
          >
            <Copy className="h-4 w-4" />
          </Button>
          
          <Button
            variant="outline"
            size="sm"
            onClick={handleDownload}
            disabled={!data?.logs}
          >
            <Download className="h-4 w-4" />
          </Button>
        </div>

        {/* Error display */}
        {error && (
          <Alert variant="destructive" className="mb-4">
            Failed to load logs: {error instanceof Error ? error.message : 'Unknown error'}
          </Alert>
        )}

        {/* Loading state */}
        {isLoading && !data && (
          <div className="flex items-center justify-center py-8">
            <RefreshCw className="h-6 w-6 animate-spin text-gray-400" />
            <span className="ml-2 text-gray-600">Loading logs...</span>
          </div>
        )}

        {/* Logs display */}
        {data?.logs && (
          <div
            ref={logsContainerRef}
            onScroll={handleScroll}
            className="bg-gray-900 text-green-400 p-4 rounded-lg font-mono text-sm overflow-auto"
            style={{ maxHeight: '500px', minHeight: '300px' }}
          >
            <pre className="whitespace-pre-wrap break-words">
              {data.logs}
            </pre>
            <div ref={logsEndRef} />
          </div>
        )}

        {/* Empty state */}
        {data && !data.logs && (
          <div className="text-center py-8 text-gray-500">
            No logs available for this container
          </div>
        )}
      </CardBody>
    </Card>
  );
};

// Made with Bob
