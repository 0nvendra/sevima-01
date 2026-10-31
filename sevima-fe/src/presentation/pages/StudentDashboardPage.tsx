import React, { useState } from 'react';
import { useInvoicesQuery, useKrsStatusQuery } from '@/application/api/queries';
import { useAuthStore } from '@/application/auth/useAuthStore';
import { KrsEligibilityBanner } from '@/presentation/components/shared/KrsEligibilityBanner';
import { StatusBadge } from '@/presentation/components/shared/StatusBadge';
import { PaymentModal } from '@/presentation/components/shared/PaymentModal';
import { Invoice } from '@/domain/models/types';
import { CreditCard, Calendar, FileText, CheckCircle2 } from 'lucide-react';

export const StudentDashboardPage: React.FC = () => {
  const { user, institutionCode } = useAuthStore();
  const [selectedInvoice, setSelectedInvoice] = useState<Invoice | null>(null);

  const { data: krsStatus, isLoading: isKrsLoading } = useKrsStatusQuery('ganjil', '2026/2027');
  const { data: invoicesData, isLoading: isInvoicesLoading } = useInvoicesQuery();

  const invoices = invoicesData?.data || [];

  const formatIDR = (val: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(val);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      {/* Student Welcome Header */}
      <div className="mb-6 flex flex-col md:flex-row md:items-center md:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Portal Tagihan Mahasiswa</h1>
          <p className="text-sm text-slate-500 mt-1">
            Selamat datang, <span className="font-semibold text-slate-800">{user?.full_name}</span>. Pantau kewajiban UKT dan kelayakan KRS semester Anda.
          </p>
        </div>
      </div>

      {/* Live KRS Eligibility Banner */}
      {isKrsLoading ? (
        <div className="h-24 bg-slate-200 animate-pulse rounded-xl mb-6" />
      ) : (
        krsStatus && <KrsEligibilityBanner status={krsStatus} />
      )}

      {/* Invoices List */}
      <div className="bg-white border border-slate-200 rounded-2xl shadow-xs overflow-hidden">
        <div className="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/50">
          <div className="flex items-center gap-2">
            <FileText className="w-5 h-5 text-blue-600" />
            <h2 className="font-bold text-slate-900 text-base">Riwayat & Daftar Tagihan UKT</h2>
          </div>
          <span className="text-xs font-semibold text-slate-500">
            Total {invoices.length} Tagihan Terdaftar
          </span>
        </div>

        {isInvoicesLoading ? (
          <div className="p-8 text-center text-sm text-slate-500">Memuat data tagihan...</div>
        ) : invoices.length === 0 ? (
          <div className="p-12 text-center text-slate-500">
            <CheckCircle2 className="w-12 h-12 mx-auto text-slate-300 mb-2" />
            <p className="text-sm font-semibold">Tidak ada tagihan yang tercatat.</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-slate-600 text-xs uppercase font-semibold border-b border-slate-200">
                <tr>
                  <th className="px-6 py-3.5">Nomor Tagihan</th>
                  <th className="px-6 py-3.5">Periode / Semester</th>
                  <th className="px-6 py-3.5">Jumlah Tagihan</th>
                  <th className="px-6 py-3.5">Jatuh Tempo</th>
                  <th className="px-6 py-3.5">Status Pembayaran</th>
                  <th className="px-6 py-3.5 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {invoices.map((inv) => (
                  <tr key={inv.id} className="hover:bg-slate-50/70 transition-colors">
                    <td className="px-6 py-4 font-mono text-xs font-bold text-blue-900">
                      {inv.invoice_number}
                    </td>
                    <td className="px-6 py-4">
                      <div className="font-medium text-slate-900 capitalize">Semester {inv.semester}</div>
                      <div className="text-xs text-slate-500">{inv.academic_year}</div>
                    </td>
                    <td className="px-6 py-4 font-semibold text-slate-900">
                      {formatIDR(inv.amount)}
                    </td>
                    <td className="px-6 py-4 text-slate-600 text-xs">
                      <div className="flex items-center gap-1.5">
                        <Calendar className="w-3.5 h-3.5 text-slate-400" />
                        {new Date(inv.due_date).toLocaleDateString('id-ID', {
                          day: 'numeric',
                          month: 'long',
                          year: 'numeric',
                        })}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={inv.status} />
                    </td>
                    <td className="px-6 py-4 text-right">
                      {inv.status === 'unpaid' ? (
                        <button
                          onClick={() => setSelectedInvoice(inv)}
                          className="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-xs font-bold transition-all shadow-xs flex items-center gap-1.5 ml-auto"
                        >
                          <CreditCard className="w-3.5 h-3.5" />
                          Bayar Sekarang
                        </button>
                      ) : (
                        <span className="text-xs text-slate-400 font-medium italic">
                          {inv.status === 'paid' ? 'Lunas Terverifikasi' : 'Kedaluwarsa'}
                        </span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Payment Simulator Modal */}
      {selectedInvoice && institutionCode && (
        <PaymentModal
          invoice={selectedInvoice}
          institutionCode={institutionCode}
          onClose={() => setSelectedInvoice(null)}
          onSuccess={() => {
            // Refreshes queries automatically via TanStack invalidation
          }}
        />
      )}
    </div>
  );
};
