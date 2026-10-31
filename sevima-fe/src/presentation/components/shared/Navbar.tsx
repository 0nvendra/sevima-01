import React from 'react';
import { useAuthStore } from '@/application/auth/useAuthStore';
import { useNavigate, useLocation } from 'react-router-dom';
import { LogOut, Building2, User, CreditCard, ShieldCheck } from 'lucide-react';

export const Navbar: React.FC = () => {
  const { user, institutionCode, institutionName, logout } = useAuthStore();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  const getRoleLabel = () => {
    if (!user) return '';
    switch (user.role) {
      case 'super_admin':
        return 'Super Administrator';
      case 'admin_keuangan':
        return 'Admin Keuangan';
      case 'mahasiswa':
        return 'Mahasiswa';
      default:
        return user.role;
    }
  };

  return (
    <header className="bg-white border-b border-slate-200 sticky top-0 z-40 shadow-xs">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          {/* Brand & Institution Info */}
          <div className="flex items-center gap-6">
            <div className="flex items-center gap-2 cursor-pointer" onClick={() => navigate('/')}>
              <div className="w-8 h-8 rounded-lg bg-blue-600 flex items-center justify-center text-white font-black text-base shadow-sm">
                S
              </div>
              <span className="font-extrabold text-lg tracking-tight text-slate-900">
                SEVIMA<span className="text-blue-600">-PAY</span>
              </span>
            </div>

            {institutionCode && (
              <div className="hidden md:flex items-center gap-2 px-2.5 py-1 rounded-md bg-slate-100 border border-slate-200 text-xs font-medium text-slate-700">
                <Building2 className="w-3.5 h-3.5 text-slate-500" />
                <span className="font-bold text-slate-900">{institutionCode}</span>
                <span className="text-slate-400">|</span>
                <span className="truncate max-w-[200px]">{institutionName}</span>
              </div>
            )}
          </div>

          {/* Navigation Links based on Role */}
          <nav className="hidden lg:flex items-center gap-1">
            {user?.role === 'mahasiswa' && (
              <>
                <button
                  onClick={() => navigate('/student')}
                  className={`px-3 py-1.5 text-xs font-semibold rounded-lg transition-colors ${
                    location.pathname === '/student'
                      ? 'bg-blue-50 text-blue-700 font-bold'
                      : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                  }`}
                >
                  Tagihan Saya
                </button>
              </>
            )}

            {user?.role === 'admin_keuangan' && (
              <>
                <button
                  onClick={() => navigate('/admin/invoices')}
                  className={`px-3 py-1.5 text-xs font-semibold rounded-lg transition-colors ${
                    location.pathname === '/admin/invoices'
                      ? 'bg-blue-50 text-blue-700 font-bold'
                      : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                  }`}
                >
                  Daftar Tagihan
                </button>
              </>
            )}

            {user?.role === 'super_admin' && (
              <>
                <button
                  onClick={() => navigate('/superadmin/institutions')}
                  className={`px-3 py-1.5 text-xs font-semibold rounded-lg transition-colors ${
                    location.pathname === '/superadmin/institutions'
                      ? 'bg-blue-50 text-blue-700 font-bold'
                      : 'text-slate-600 hover:text-slate-900 hover:bg-slate-50'
                  }`}
                >
                  Kelola Institusi
                </button>
              </>
            )}

            <button
              onClick={() => navigate('/mock-gateway')}
              className={`px-3 py-1.5 text-xs font-semibold rounded-lg transition-colors flex items-center gap-1.5 ${
                location.pathname === '/mock-gateway'
                  ? 'bg-purple-50 text-purple-700 font-bold'
                  : 'text-purple-700 bg-purple-50/50 hover:bg-purple-50'
              }`}
            >
              <CreditCard className="w-3.5 h-3.5" />
              Simulator Gateway
            </button>
          </nav>

          {/* User Profile & Logout */}
          <div className="flex items-center gap-3">
            <div className="text-right hidden sm:block">
              <p className="text-xs font-bold text-slate-900">{user?.full_name}</p>
              <div className="flex items-center justify-end gap-1 text-[11px] text-slate-500">
                <span className="capitalize">{getRoleLabel()}</span>
              </div>
            </div>

            <button
              onClick={handleLogout}
              title="Keluar"
              className="p-2 text-slate-500 hover:text-rose-600 hover:bg-rose-50 rounded-lg transition-colors"
            >
              <LogOut className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </header>
  );
};
