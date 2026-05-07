import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';
import Login from './pages/Login';
import Dashboard from './pages/Dashboard';
import Instances from './pages/Instances';
import InstanceForm from './pages/InstanceForm';
import InstanceDetail from './pages/InstanceDetail';
import Servers from './pages/Servers';
import Users from './pages/Users';
import Settings from './pages/Settings';
import Help from './pages/Help';
import { ToastContainer } from './components/ui';
import { useAuthStore } from './stores/authStore';
import { useToastStore } from './stores/toastStore';

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5000,
    },
  },
});

const App: React.FC = () => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const toasts = useToastStore((state) => state.toasts);
  const removeToast = useToastStore((state) => state.removeToast);

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ToastContainer toasts={toasts} onClose={removeToast} />
        <Routes>
          {/* Public routes */}
          <Route
            path="/login"
            element={
              isAuthenticated ? <Navigate to="/dashboard" replace /> : <Login />
            }
          />

          {/* Protected routes */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <Layout>
                  <Dashboard />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/instances"
            element={
              <ProtectedRoute>
                <Layout>
                  <Instances />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/instances/new"
            element={
              <ProtectedRoute>
                <Layout>
                  <InstanceForm />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/instances/:id/edit"
            element={
              <ProtectedRoute>
                <Layout>
                  <InstanceForm />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/instances/:id"
            element={
              <ProtectedRoute>
                <Layout>
                  <InstanceDetail />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/servers"
            element={
              <ProtectedRoute>
                <Layout>
                  <Servers />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/users"
            element={
              <ProtectedRoute requiredRole="admin">
                <Layout>
                  <Users />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/settings"
            element={
              <ProtectedRoute>
                <Layout>
                  <Settings />
                </Layout>
              </ProtectedRoute>
            }
          />

          <Route
            path="/help"
            element={
              <ProtectedRoute>
                <Layout>
                  <Help />
                </Layout>
              </ProtectedRoute>
            }
          />

          {/* Default redirect */}
          <Route
            path="/"
            element={
              <Navigate to={isAuthenticated ? '/dashboard' : '/login'} replace />
            }
          />

          {/* 404 */}
          <Route
            path="*"
            element={
              <div className="min-h-screen flex items-center justify-center bg-gray-100">
                <div className="text-center">
                  <h1 className="text-6xl font-bold text-gray-900">404</h1>
                  <p className="mt-2 text-xl text-gray-600">Page not found</p>
                  <a
                    href="/"
                    className="mt-4 inline-block text-blue-600 hover:text-blue-800"
                  >
                    Go back home
                  </a>
                </div>
              </div>
            }
          />
        </Routes>
      </BrowserRouter>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
  );
};

export default App;


