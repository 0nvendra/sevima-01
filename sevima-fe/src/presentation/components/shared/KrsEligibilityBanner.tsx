import React from 'react';
import { KrsStatus } from '@/domain/models/types';
import { CheckCircle2, AlertTriangle, ShieldCheck } from 'lucide-react';

interface Props {
  status: KrsStatus;
}

export const KrsEligibilityBanner: React.FC<Props> = ({ status }) => {
  if (status.is_eligible_krs) {
    return (
      <div className="bg-emerald-50/80 border border-emerald-200 rounded-xl p-5 mb-6 shadow-sm">
        <div className="flex items-start gap-4">
          <div className="p-2.5 bg-emerald-100 text-emerald-700 rounded-lg shrink-0">
            <CheckCircle2 className="w-6 h-6" />
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <h3 className="text-base font-bold text-emerald-950">
                Status KRS: Memenuhi Syarat (Eligible)
              </h3>
              <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-semibold bg-emerald-200/70 text-emerald-800">
                <ShieldCheck className="w-3.5 h-3.5" />
                Siap KRS
              </span>
            </div>
            <p className="text-sm text-emerald-800 mt-1">
              Seluruh kewajiban pembayaran UKT semester <span className="font-semibold uppercase">{status.active_semester}</span> tahun akademik <span className="font-semibold">{status.active_academic_year}</span> telah lunas. Anda dapat langsung melanjutkan ke pengisian KRS pada SIAKAD kampus.
            </p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-amber-50/80 border border-amber-200 rounded-xl p-5 mb-6 shadow-sm">
      <div className="flex items-start gap-4">
        <div className="p-2.5 bg-amber-100 text-amber-700 rounded-lg shrink-0">
          <AlertTriangle className="w-6 h-6" />
        </div>
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h3 className="text-base font-bold text-amber-950">
              Status KRS: Belum Memenuhi Syarat (Terkunci)
            </h3>
            <span className="px-2 py-0.5 rounded text-xs font-semibold bg-amber-200/70 text-amber-900">
              {status.unpaid_invoices_count} Tagihan Tertunda
            </span>
          </div>
          <p className="text-sm text-amber-800 mt-1">
            Pengisian KRS semester <span className="font-semibold uppercase">{status.active_semester}</span> ditangguhkan karena Anda masih memiliki tagihan UKT aktif yang belum diselesaikan. Harap segera lakukan pelunasan untuk membuka kunci KRS.
          </p>
        </div>
      </div>
    </div>
  );
};
