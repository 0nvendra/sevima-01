import React, { useState } from 'react';
import { useInvoicesQuery, useCreateInvoiceMutation, useInvoiceDetailQuery } from '@/application/api/queries';
import { useAuthStore } from '@/application/auth/useAuthStore';
import { StatusBadge } from '@/presentation/components/shared/StatusBadge';
import { PaymentModal } from '@/presentation/components/shared/PaymentModal';
import { Invoice } from '@/domain/models/types';
import {
  FileText,
  Plus,
  Search,
  Filter,
  CheckCircle,
  Clock,
  XCircle,
  Calendar,
  X,
  CreditCard,
  RefreshCw,
  History,
} from 'lucide-react';

export const AdminInvoicesPage: React.FC = () => {
  const { institutionCode } = useAuthStore();
  const [statusFilter, setStatusFilter] = useState('');
  const [search, setSearch] = useState('');
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [selectedDetailId, setSelectedDetailId] = useState<string | null>(null);
  const [selectedPayInvoice, setSelectedPayInvoice] = useState<Invoice | null>(null);

  // Form state
  const [studentId, setStudentId] = useState('88888888-8888-8888-8888-888888888888');
  const [amount, setAmount] = useState(7500000);
  const [semester, setSemester] = useState('ganjil');
  const [academicYear, setAcademicYear] = useState('2026/2027');
  const [dueDate, setDueDate] = useState('2026-12-31');
  const [notes, setNotes] = useState('Tagihan UKT Reguler');

  const { data: invoicesData, isLoading, refetch } = useInvoicesQuery({
    status: statusFilter || undefined,
    search: search || undefined,
  });

  const { data: detailData, isLoading: isDetailLoading } = useInvoiceDetailQuery(selectedDetailId || '');
  const { mutate: createInvoice, isPending: isCreating } = useCreateInvoiceMutation();

  const invoices = invoicesData?.data || [];
  const meta = invoicesData?.meta;

  const formatIDR = (val: number) => {
    return new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      maximumFractionDigits: 0,
    }).format(val);
  };

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    createInvoice(
      {
        student_id: studentId,
        amount: Number(amount),
        semester,
        academic_year: academicYear,
        due_date: dueDate,
        notes,
      },
      {
        onSuccess: () => {
          setIsCreateOpen(false);
          refetch();
        },
      }
    );
  };

  // Metric counts
  const totalInvoices = meta?.total_records || invoices.length;
  const paidCount = invoices.filter((i) => i.status === 'paid').length;
  const unpaidCount = invoices.filter((i) => i.status === 'unpaid').length;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      {/* Title & Actions */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Manajemen Tagihan UKT</h1>
          <p className="text-sm text-slate-500 mt-1">
            Kelola penerbitan invoice, pantau status pembayaran, dan riwayat audit webhook kampus Anda.
          </p>
        </div>
        <button
          onClick={() => setIsCreateOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-bold shadow-sm transition-all self-start sm:self-auto"
        >
          <Plus className="w-4 h-4" />
          Terbitkan Tagihan Baru
        </button>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
          <div className="flex items-center justify-between text-slate-500 text-xs font-bold uppercase tracking-wider">
            <span>Total Tagihan</span>
            <FileText className="w-4 h-4 text-slate-400" />
          </div>
          <div className="text-2xl font-black text-slate-900 mt-2">{totalInvoices}</div>
          <div className="text-xs text-slate-400 mt-1">Seluruh periode akademik</div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
          <div className="flex items-center justify-between text-emerald-700 text-xs font-bold uppercase tracking-wider">
            <span>Lunas Terbayar</span>
            <CheckCircle className="w-4 h-4 text-emerald-500" />
          </div>
          <div className="text-2xl font-black text-emerald-600 mt-2">{paidCount}</div>
          <div className="text-xs text-emerald-700/70 mt-1">Eligible KRS aktif</div>
        </div>

        <div className="bg-white p-5 rounded-2xl border border-slate-200 shadow-xs">
          <div className="flex items-center justify-between text-amber-700 text-xs font-bold uppercase tracking-wider">
            <span>Menunggu Pembayaran</span>
            <Clock className="w-4 h-4 text-amber-500" />
          </div>
          <div className="text-2xl font-black text-amber-600 mt-2">{unpaidCount}</div>
          <div className="text-xs text-amber-700/70 mt-1">KRS ditangguhkan</div>
        </div>
      </div>

      {/* Filters & Search */}
      <div className="bg-white p-4 rounded-2xl border border-slate-200 shadow-xs flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="relative w-full md:w-96">
          <Search className="w-4 h-4 text-slate-400 absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari NIM, Nama Mahasiswa, atau No. Invoice..."
            className="w-full pl-10 pr-4 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-800 focus:bg-white focus:outline-hidden focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all"
          />
        </div>

        <div className="flex items-center gap-2 w-full md:w-auto">
          <Filter className="w-4 h-4 text-slate-400" />
          <span className="text-xs font-bold text-slate-600">Status:</span>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            className="px-3 py-1.5 bg-slate-50 border border-slate-200 rounded-xl text-xs text-slate-700 font-medium focus:outline-hidden focus:ring-2 focus:ring-blue-500/20"
          >
            <option value="">Semua Status</option>
            <option value="unpaid">Belum Bayar (Unpaid)</option>
            <option value="paid">Lunas (Paid)</option>
            <option value="expired">Kedaluwarsa (Expired)</option>
          </select>
        </div>
      </div>

      {/* Invoices Table */}
      <div className="bg-white border border-slate-200 rounded-2xl shadow-xs overflow-hidden">
        {isLoading ? (
          <div className="p-12 text-center text-sm text-slate-500">Memuat data tagihan institusi...</div>
        ) : invoices.length === 0 ? (
          <div className="p-12 text-center text-slate-400">Tidak ada tagihan yang sesuai dengan filter.</div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-slate-600 text-xs uppercase font-semibold border-b border-slate-200">
                <tr>
                  <th className="px-6 py-3.5">Invoice</th>
                  <th className="px-6 py-3.5">Mahasiswa</th>
                  <th className="px-6 py-3.5">Nominal UKT</th>
                  <th className="px-6 py-3.5">Periode</th>
                  <th className="px-6 py-3.5">Jatuh Tempo</th>
                  <th className="px-6 py-3.5">Status</th>
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
                      <div className="font-semibold text-slate-900">{inv.student?.user?.full_name || 'Mahasiswa'}</div>
                      <div className="text-xs text-slate-500 font-mono">{inv.student?.nim} · {inv.student?.study_program}</div>
                    </td>
                    <td className="px-6 py-4 font-semibold text-slate-900">
                      {formatIDR(inv.amount)}
                    </td>
                    <td className="px-6 py-4 text-xs">
                      <span className="font-medium text-slate-700 capitalize">{inv.semester}</span>
                      <span className="text-slate-400"> ({inv.academic_year})</span>
                    </td>
                    <td className="px-6 py-4 text-slate-600 text-xs">
                      {new Date(inv.due_date).toLocaleDateString('id-ID')}
                    </td>
                    <td className="px-6 py-4">
                      <StatusBadge status={inv.status} />
                    </td>
                    <td className="px-6 py-4 text-right space-x-2">
                      <button
                        onClick={() => setSelectedDetailId(inv.id)}
                        className="px-2.5 py-1 text-xs font-medium text-slate-600 bg-slate-100 hover:bg-slate-200 rounded-md transition-colors"
                      >
                        Detail & Audit
                      </button>
                      {inv.status === 'unpaid' && institutionCode && (
                        <button
                          onClick={() => setSelectedPayInvoice(inv)}
                          className="px-2.5 py-1 text-xs font-semibold text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-md transition-colors"
                        >
                          Simulator
                        </button>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Modal Terbitkan Tagihan Baru */}
      {isCreateOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl max-w-md w-full shadow-2xl border border-slate-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50">
              <h3 className="font-bold text-slate-900 text-base">Terbitkan Tagihan UKT Baru</h3>
              <button onClick={() => setIsCreateOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreate} className="p-6 space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">
                  Pilih Mahasiswa
                </label>
                <select
                  value={studentId}
                  onChange={(e) => setStudentId(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                >
                  <option value="88888888-8888-8888-8888-888888888888">
                    Ahmad Fauzi (21/475123/TK/52111 - Teknik Informatika)
                  </option>
                  <option value="99999999-9999-9999-9999-999999999999">
                    Bima Sakti (5025211001 - Sistem Informasi)
                  </option>
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">
                  Nominal Tagihan (IDR)
                </label>
                <input
                  type="number"
                  required
                  min={100000}
                  step={50000}
                  value={amount}
                  onChange={(e) => setAmount(Number(e.target.value))}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-semibold"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">Semester</label>
                  <select
                    value={semester}
                    onChange={(e) => setSemester(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                  >
                    <option value="ganjil">Ganjil</option>
                    <option value="genap">Genap</option>
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-700 mb-1">Tahun Akademik</label>
                  <input
                    type="text"
                    value={academicYear}
                    onChange={(e) => setAcademicYear(e.target.value)}
                    className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-mono"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Batas Waktu (Jatuh Tempo)</label>
                <input
                  type="date"
                  required
                  value={dueDate}
                  onChange={(e) => setDueDate(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Catatan / Keterangan</label>
                <input
                  type="text"
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                />
              </div>

              <div className="pt-2 flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsCreateOpen(false)}
                  className="px-4 py-2 text-xs text-slate-600 hover:bg-slate-100 rounded-lg"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={isCreating}
                  className="px-5 py-2 text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50"
                >
                  {isCreating ? 'Menerbitkan...' : 'Terbitkan Sekarang'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Modal Detail & Transaction Audit Timeline */}
      {selectedDetailId && (
        <div className="fixed inset-0 z-50 bg-slate-900/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl max-w-lg w-full shadow-2xl border border-slate-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50">
              <div className="flex items-center gap-2">
                <History className="w-5 h-5 text-blue-600" />
                <h3 className="font-bold text-slate-900 text-base">Detail Tagihan & Riwayat Audit</h3>
              </div>
              <button onClick={() => setSelectedDetailId(null)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <div className="p-6 space-y-4 max-h-[80vh] overflow-y-auto">
              {isDetailLoading || !detailData ? (
                <div className="p-8 text-center text-sm text-slate-500">Memuat rincian transaksi...</div>
              ) : (
                <>
                  <div className="bg-slate-50 p-4 rounded-xl space-y-2 border border-slate-100 text-xs">
                    <div className="flex justify-between">
                      <span className="text-slate-500">Nomor Invoice:</span>
                      <span className="font-mono font-bold text-slate-900">{detailData.invoice.invoice_number}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Mahasiswa:</span>
                      <span className="font-semibold text-slate-900">{detailData.invoice.student?.user?.full_name}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-slate-500">Nominal:</span>
                      <span className="font-bold text-slate-900">{formatIDR(detailData.invoice.amount)}</span>
                    </div>
                    <div className="flex justify-between items-center">
                      <span className="text-slate-500">Status Saat Ini:</span>
                      <StatusBadge status={detailData.invoice.status} />
                    </div>
                    {detailData.invoice.payment_method && (
                      <div className="flex justify-between">
                        <span className="text-slate-500">Kanal Pembayaran:</span>
                        <span className="font-semibold text-blue-800">{detailData.invoice.payment_method}</span>
                      </div>
                    )}
                  </div>

                  {/* Audit Timeline */}
                  <div>
                    <h4 className="text-xs font-bold text-slate-900 uppercase tracking-wider mb-3">
                      Audit Trail Perubahan Status
                    </h4>
                    <div className="space-y-3 relative before:absolute before:inset-0 before:left-3.5 before:w-0.5 before:bg-slate-200">
                      {detailData.history.map((h) => (
                        <div key={h.id} className="flex items-start gap-3 relative">
                          <div className="w-7 h-7 rounded-full bg-blue-100 text-blue-700 flex items-center justify-center shrink-0 border-2 border-white shadow-xs">
                            <Clock className="w-3.5 h-3.5" />
                          </div>
                          <div className="flex-1 bg-white border border-slate-100 rounded-lg p-2.5 text-xs shadow-2xs">
                            <div className="flex justify-between items-center mb-1">
                              <span className="font-bold text-slate-800 capitalize">
                                {h.previous_status} → {h.new_status}
                              </span>
                              <span className="text-[10px] text-slate-400">
                                {new Date(h.created_at).toLocaleString('id-ID')}
                              </span>
                            </div>
                            <p className="text-[11px] text-slate-600">{h.notes || 'Status diperbarui'}</p>
                            <span className="inline-block mt-1 font-mono text-[10px] text-slate-400 bg-slate-50 px-1 rounded">
                              Source: {h.source}
                            </span>
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {/* Simulator Trigger Modal */}
      {selectedPayInvoice && institutionCode && (
        <PaymentModal
          invoice={selectedPayInvoice}
          institutionCode={institutionCode}
          onClose={() => setSelectedPayInvoice(null)}
          onSuccess={() => refetch()}
        />
      )}
    </div>
  );
};
