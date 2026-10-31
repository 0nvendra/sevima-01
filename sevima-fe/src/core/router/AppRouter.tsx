import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useAuthStore } from '@/application/auth/useAuthStore';
import { Navbar } from '@/presentation/components/shared/Navbar';
import { LoginPage } from '@/presentation/pages/LoginPage';
import { StudentDashboardPage } from '@/presentation/pages/StudentDashboardPage';
import { AdminInvoicesPage } from '@/presentation/pages/AdminInvoicesPage';
import { SuperAdminInstitutionsPage } from '@/presentation/pages/SuperAdminInstitutionsPage';
import { MockGatewayPage } from '@/presentation/pages/MockGatewayPage';
import { UserRole } from '@/domain/models/types';

interface ProtectedRouteProps {
  children: React.ReactNode;
  allowedRoles?: UserRole[];
}

const ProtectedRoute: React.FC<ProtectedRouteProps> = ({ children, allowedRoles }) => {
  const { isAuthenticated, user } = useAuthStore();

  if (!isAuthenticated || !user) {
    return <Navigate to="/login" replace />;
  }

  if (allowedRoles && !allowedRoles.includes(user.role)) {
    if (user.role === 'mahasiswa') return <Navigate to="/student" replace />;
    if (user.role === 'admin_keuangan') return <Navigate to="/admin/invoices" replace />;
    if (user.role === 'super_admin') return <Navigate to="/superadmin/institutions" replace />;
  }

  return (
    <div className="min-h-screen bg-slate-50 flex flex-col">
      <Navbar />
      <main className="flex-1">{children}</main>
    </div>
  );
};

export const AppRouter: React.FC = () => {
  const { isAuthenticated, user, initialize } = useAuthStore();

  React.useEffect(() => {
    initialize();
  }, [initialize]);

  const getDefaultRedirect = () => {
    if (!isAuthenticated || !user) return '/login';
    if (user.role === 'mahasiswa') return '/student';
    if (user.role === 'admin_keuangan') return '/admin/invoices';
    if (user.role === 'super_admin') return '/superadmin/institutions';
    return '/login';
  };

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />

        {/* Mahasiswa View */}
        <Route
          path="/student"
          element={
            <ProtectedRoute allowedRoles={['mahasiswa']}>
              <StudentDashboardPage />
            </ProtectedRoute>
          }
        />

        {/* Admin Keuangan View */}
        <Route
          path="/admin/invoices"
          element={
            <ProtectedRoute allowedRoles={['admin_keuangan']}>
              <AdminInvoicesPage />
            </ProtectedRoute>
          }
        />

        {/* Super Admin View */}
        <Route
          path="/superadmin/institutions"
          element={
            <ProtectedRoute allowedRoles={['super_admin']}>
              <SuperAdminInstitutionsPage />
            </ProtectedRoute>
          }
        />

        {/* Simulator Gateway Sandbox (Accessible by all roles) */}
        <Route
          path="/mock-gateway"
          element={
            <ProtectedRoute>
              <MockGatewayPage />
            </ProtectedRoute>
          }
        />

        {/* Default Catch-all */}
        <Route path="*" element={<Navigate to={getDefaultRedirect()} replace />} />
      </Routes>
    </BrowserRouter>
  );
};
