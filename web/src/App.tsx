import type { ReactNode } from "react";
import { BrowserRouter, Routes, Route, Navigate, useLocation } from "react-router-dom";
import { Layout } from "./components/layout/Layout";
import { Dashboard } from "./pages/Dashboard";
import { Endpoints } from "./pages/Endpoints";
import { EndpointDetails } from "./pages/EndpointDetails";
import { Config } from "./pages/Config";
import { EndpointForm } from "./pages/EndpointForm";
import { Login } from "./pages/Login";
import { AuthProvider, useAuth } from "./context/AuthContext";

function RequireAuth({ children }: { children: ReactNode }) {
  const { isAuthenticated, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <div className="min-h-screen flex items-center justify-center bg-gray-900 text-white">Loading...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/" element={
            <RequireAuth>
              <Layout />
            </RequireAuth>
          }>
            <Route index element={<Dashboard />} />
            <Route path="endpoints" element={<Endpoints />} />
            <Route path="endpoints/new" element={<EndpointForm />} />
            <Route path="endpoints/:id" element={<EndpointDetails />} />
            <Route path="endpoints/:id/edit" element={<EndpointForm />} />
            <Route path="config" element={<Config />} />
            <Route path="settings" element={<div>Settings Page (Coming Soon)</div>} />
          </Route>
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
