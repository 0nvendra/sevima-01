import React, { useState } from 'react';
import { useInstitutionsQuery, useCreateInstitutionMutation } from '@/application/api/queries';
import { Building2, Plus, ShieldCheck, Mail, Phone, MapPin, X } from 'lucide-react';

export const SuperAdminInstitutionsPage: React.FC = () => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [code, setCode] = useState('');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [address, setAddress] = useState('');

  const { data: institutions, isLoading, refetch } = useInstitutionsQuery();
  const { mutate: createInst, isPending } = useCreateInstitutionMutation();

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    createInst(
      { code, name, email, phone, address },
      {
        onSuccess: () => {
          setIsModalOpen(false);
          setCode('');
          setName('');
          setEmail('');
          setPhone('');
          setAddress('');
          refetch();
        },
      }
    );
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Manajemen Kampus & Multi-Tenant</h1>
          <p className="text-sm text-slate-500 mt-1">
            Kelola master data perguruan tinggi pengguna SEVIMA-PAY dan konfigurasi rahasia HMAC webhook.
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="inline-flex items-center gap-2 px-4 py-2.5 bg-blue-600 hover:bg-blue-700 text-white rounded-xl text-xs font-bold shadow-sm transition-all"
        >
          <Plus className="w-4 h-4" />
          Daftarkan Kampus Baru
        </button>
      </div>

      {/* Institutions Grid */}
      {isLoading ? (
        <div className="p-12 text-center text-sm text-slate-500">Memuat data institusi...</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
          {institutions?.map((inst) => (
            <div
              key={inst.id}
              className="bg-white rounded-2xl border border-slate-200 p-6 shadow-xs hover:border-blue-300 transition-all space-y-4"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-xl bg-blue-50 text-blue-700 flex items-center justify-center font-black text-sm border border-blue-100">
                    {inst.code}
                  </div>
                  <div>
                    <h3 className="font-bold text-slate-900 text-sm leading-snug">{inst.name}</h3>
                    <span className="font-mono text-[11px] text-slate-400">Kode: {inst.code}</span>
                  </div>
                </div>
                <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-[11px] font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
                  <ShieldCheck className="w-3 h-3" />
                  Aktif
                </span>
              </div>

              <div className="space-y-1.5 text-xs text-slate-600 border-t border-slate-100 pt-3">
                {inst.email && (
                  <div className="flex items-center gap-2">
                    <Mail className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span className="truncate">{inst.email}</span>
                  </div>
                )}
                {inst.phone && (
                  <div className="flex items-center gap-2">
                    <Phone className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span>{inst.phone}</span>
                  </div>
                )}
                {inst.address && (
                  <div className="flex items-center gap-2">
                    <MapPin className="w-3.5 h-3.5 text-slate-400 shrink-0" />
                    <span className="truncate">{inst.address}</span>
                  </div>
                )}
              </div>

              <div className="bg-slate-50 p-2.5 rounded-lg border border-slate-100 text-[11px] text-slate-500 font-mono flex items-center justify-between">
                <span>HMAC Secret:</span>
                <span className="text-slate-400 italic">Terenkripsi aman</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal Tambah Kampus */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-slate-900/50 backdrop-blur-xs flex items-center justify-center p-4">
          <div className="bg-white rounded-2xl max-w-md w-full shadow-2xl border border-slate-200 overflow-hidden">
            <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50">
              <h3 className="font-bold text-slate-900 text-base">Daftarkan Institusi Baru</h3>
              <button onClick={() => setIsModalOpen(false)} className="text-slate-400 hover:text-slate-600">
                <X className="w-5 h-5" />
              </button>
            </div>

            <form onSubmit={handleCreate} className="p-6 space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Kode Kampus (Unik)</label>
                <input
                  type="text"
                  required
                  placeholder="Contoh: UI, ITB, UNAIR"
                  value={code}
                  onChange={(e) => setCode(e.target.value.toUpperCase())}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs uppercase font-mono font-bold"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Nama Perguruan Tinggi</label>
                <input
                  type="text"
                  required
                  placeholder="Nama resmi kampus"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-medium"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Email Resmi</label>
                <input
                  type="email"
                  required
                  placeholder="info@kampus.ac.id"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Nomor Telepon</label>
                <input
                  type="text"
                  placeholder="021-xxxxxx"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-700 mb-1">Alamat Kampus</label>
                <input
                  type="text"
                  placeholder="Kota / Alamat lengkap"
                  value={address}
                  onChange={(e) => setAddress(e.target.value)}
                  className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
                />
              </div>

              <div className="pt-2 flex justify-end gap-2">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 text-xs text-slate-600 hover:bg-slate-100 rounded-lg"
                >
                  Batal
                </button>
                <button
                  type="submit"
                  disabled={isPending}
                  className="px-5 py-2 text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white rounded-lg disabled:opacity-50"
                >
                  {isPending ? 'Menyimpan...' : 'Daftarkan Kampus'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
