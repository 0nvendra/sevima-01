import React from 'react';
import { InvoiceStatus } from '@/domain/models/types';
import { CheckCircle2, Clock, XCircle } from 'lucide-react';

interface Props {
  status: InvoiceStatus;
}

export const StatusBadge: React.FC<Props> = ({ status }) => {
  switch (status) {
    case 'paid':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200">
          <CheckCircle2 className="w-3.5 h-3.5" />
          Lunas
        </span>
      );
    case 'unpaid':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-700 border border-amber-200">
          <Clock className="w-3.5 h-3.5" />
          Belum Bayar
        </span>
      );
    case 'expired':
      return (
        <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200">
          <XCircle className="w-3.5 h-3.5" />
          Kedaluwarsa
        </span>
      );
    default:
      return <span className="px-2.5 py-1 text-xs bg-slate-100 rounded-full">{status}</span>;
  }
};
