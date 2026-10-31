import React, { useState } from 'react';
import { useLoginMutation } from '@/application/api/queries';
import { useNavigate } from 'react-router-dom';
import { LogIn, Building2, User, Shield, AlertCircle, RefreshCw } from 'lucide-react';

export const LoginPage: React.FC = () => {
  const [email, setEmail] = useState('fauzi@mail.ugm.ac.id');
  const [password, setPassword] = useState('Password123!');
  const navigate = useNavigate();

  const { mutate: login, isPending, error } = useLoginMutation();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    login(
      { email, password },
      {
        onSuccess: (data) => {
          if (data.user.role === 'mahasiswa') {
            navigate('/student');
          } else if (data.user.role === 'admin_keuangan') {
            navigate('/admin/invoices');
          } else if (data.user.role === 'super_admin') {
            navigate('/superadmin/institutions');
          } else {
            navigate('/');
          }
        },
      }
    );
  };

  const setDemoAccount = (roleEmail: string) => {
    setEmail(roleEmail);
    setPassword('Password123!');
  };

  return (
    <div className="min-h-screen bg-slate-100/70 flex flex-col justify-center items-center p-4">
      <div className="max-w-md w-full bg-white rounded-2xl shadow-xl border border-slate-200/80 overflow-hidden">
        {/* Brand Header */}
        <div className="bg-linear-to-b from-blue-900 to-blue-950 p-6 text-center text-white">
          <div className="w-12 h-12 bg-blue-600 rounded-xl mx-auto flex items-center justify-center font-black text-2xl text-white shadow-lg mb-3">
            S
          </div>
          <h1 className="text-xl font-bold tracking-tight">SEVIMA-PAY</h1>
          <p className="text-xs text-blue-200 mt-1">
            Platform Penagihan UKT & Webhook Payment Multi-Kampus
          </p>
        </div>

        {/* Demo Account Quick Switcher */}
        <div className="bg-slate-50 border-b border-slate-200 px-6 py-3">
          <span className="text-[11px] font-bold text-slate-500 uppercase tracking-wider block mb-2">
            Pilih Akun Demo Pengujian:
          </span>
          <div className="flex flex-wrap gap-1.5">
            <button
              type="button"
              onClick={() => setDemoAccount('fauzi@mail.ugm.ac.id')}
              className={`text-xs px-2.5 py-1 rounded-md font-medium transition-colors ${
                email === 'fauzi@mail.ugm.ac.id'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white border border-slate-300 text-slate-700 hover:bg-slate-100'
              }`}
            >
              Mahasiswa UGM
            </button>
            <button
              type="button"
              onClick={() => setDemoAccount('admin@ugm.ac.id')}
              className={`text-xs px-2.5 py-1 rounded-md font-medium transition-colors ${
                email === 'admin@ugm.ac.id'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white border border-slate-300 text-slate-700 hover:bg-slate-100'
              }`}
            >
              Admin UGM
            </button>
            <button
              type="button"
              onClick={() => setDemoAccount('admin@its.ac.id')}
              className={`text-xs px-2.5 py-1 rounded-md font-medium transition-colors ${
                email === 'admin@its.ac.id'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white border border-slate-300 text-slate-700 hover:bg-slate-100'
              }`}
            >
              Admin ITS
            </button>
            <button
              type="button"
              onClick={() => setDemoAccount('superadmin@sevima.com')}
              className={`text-xs px-2.5 py-1 rounded-md font-medium transition-colors ${
                email === 'superadmin@sevima.com'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white border border-slate-300 text-slate-700 hover:bg-slate-100'
              }`}
            >
              Super Admin
            </button>
          </div>
        </div>

        {/* Form Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 bg-rose-50 border border-rose-200 rounded-xl text-xs text-rose-700 flex items-center gap-2">
              <AlertCircle className="w-4 h-4 shrink-0" />
              <span>{(error as any)?.response?.data?.error?.message || error.message}</span>
            </div>
          )}

          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">
              Alamat Email
            </label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="nama@kampus.ac.id"
              className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-sm text-slate-900 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-blue-500/20 focus:border-blue-600 transition-all"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-700 mb-1">
              Kata Sandi
            </label>
            <input
              type="password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              className="w-full px-3.5 py-2.5 bg-slate-50 border border-slate-300 rounded-xl text-sm text-slate-900 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-blue-500/20 focus:border-blue-600 transition-all"
            />
          </div>

          <button
            type="submit"
            disabled={isPending}
            className="w-full mt-2 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-sm font-bold shadow-sm transition-all flex items-center justify-center gap-2 disabled:opacity-50"
          >
            {isPending ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" />
                Memverifikasi...
              </>
            ) : (
              <>
                <LogIn className="w-4 h-4" />
                Masuk ke Portal
              </>
            )}
          </button>
        </form>

        <div className="px-6 py-4 bg-slate-50 border-t border-slate-100 text-center">
          <p className="text-[11px] text-slate-500">
            Tes Teknis Software Engineer SEVIMA · Project M2 (Pembayaran UKT via Webhook)
          </p>
        </div>
      </div>
    </div>
  );
};
