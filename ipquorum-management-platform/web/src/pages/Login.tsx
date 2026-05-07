import React, { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { LogIn } from 'lucide-react';
import { useAuthStore } from '../stores/authStore';
import { Button, Input, Alert, Card, CardBody } from '../components/ui';

const Login: React.FC = () => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const navigate = useNavigate();
  const location = useLocation();
  const login = useAuthStore((state) => state.login);

  const from = (location.state as any)?.from?.pathname || '/dashboard';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      await login(username, password);
      navigate(from, { replace: true });
    } catch (err: any) {
      setError(err.message || 'Login failed. Please check your credentials.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-background to-muted px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-primary to-primary/60 bg-clip-text text-transparent mb-2">IPQuorum</h1>
          <p className="text-muted-foreground">Management Platform</p>
        </div>

        <Card variant="elevated">
          <CardBody>
            <h2 className="text-2xl font-semibold mb-6">Sign In</h2>

            {error && (
              <Alert variant="destructive" dismissible onDismiss={() => setError('')} className="mb-4">
                {error}
              </Alert>
            )}

            <form onSubmit={handleSubmit} className="space-y-4">
              <Input
                label="Username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter your username"
                required
                fullWidth
                autoComplete="username"
                disabled={isLoading}
              />

              <Input
                label="Password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Enter your password"
                required
                fullWidth
                autoComplete="current-password"
                disabled={isLoading}
              />

              <Button
                type="submit"
                variant="default"
                size="lg"
                fullWidth
                isLoading={isLoading}
                leftIcon={<LogIn className="h-5 w-5" />}
              >
                Sign In
              </Button>
            </form>

            <div className="mt-6 text-center text-sm text-gray-600">
              <p>Default credentials:</p>
              <p className="font-mono text-xs mt-1">admin / admin123</p>
            </div>
          </CardBody>
        </Card>

        <p className="mt-4 text-center text-sm text-gray-600">
          IBM Storage Virtualize IP Quorum Service
        </p>
      </div>
    </div>
  );
};

export default Login;


