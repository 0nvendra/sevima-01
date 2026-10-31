import React, { useState } from 'react';
import { Invoice } from '@/domain/models/types';
import { useMockPaymentMutation } from '@/application/api/queries';
import { X, CreditCard, ShieldAlert, CheckCircle2, RefreshCw } from 'lucide-react';

interface Props {
  invoice: Invoice;
  institutionCode: string;
  onClose: () => void;
  onSuccess?: () => void;
}

export const PaymentModal: React.FC<Props> = ({ invoice, institutionCode, onClose, onSuccess }) => {
  const [method, setMethod] = useState('BCA_VA');
  const [tamper, setTamper] = useState(false);
  const [resultData, setResultData] = useState<any>(null);

  const { mutate: pay, isPending, error } = useMockPaymentMutation();

  const handlePay = () => {
    pay(
      {
        institution_code: institutionCode,
        invoice_number: invoice.invoice_number,
        payment_method: method,
        simulate_tamper: tamper,
        event_type: 'payment.success',
      },
      {
        onSuccess: (data) => {
          setResultData(data);
          if (onSuccess) onSuccess();
        },
      }
    );
  };

  const formattedAmount = new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  }).format(invoice.amount);

  return (
    <div className="fixed inset-0 z-50 bg-slate-900/50 backdrop-blur-xs flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl max-w-lg w-full shadow-2xl border border-slate-200 overflow-hidden">
        {/* Header */}
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between bg-slate-50">
          <div className="flex items-center gap-2.5">
            <div className="p-2 bg-blue-100 text-blue-700 rounded-lg">
              <CreditCard className="w-5 h-5" />
            </div>
            <div>
              <h3 className="font-bold text-slate-900 text-base">Simulator Pembayaran UKT</h3>
              <p className="text-xs text-slate-500">SEVIMA Payment Gateway Mock Sandbox</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-slate-600 p-1.5 rounded-lg hover:bg-slate-200/50 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="p-6 space-y-5">
          {/* Invoice Summary */}
          <div className="bg-blue-50/60 border border-blue-100 rounded-xl p-4">
            <div className="flex justify-between items-center mb-2">
              <span className="text-xs font-semibold text-blue-800 uppercase tracking-wider">No. Tagihan</span>
              <span className="font-mono text-xs font-bold text-blue-900 bg-blue-100/70 px-2 py-0.5 rounded">
                {invoice.invoice_number}
              </span>
            </div>
            <div className="flex justify-between items-baseline">
              <span className="text-sm text-slate-600">Total Biaya UKT:</span>
              <span className="text-xl font-bold text-slate-900">{formattedAmount}</span>
            </div>
            <div className="mt-2 pt-2 border-t border-blue-100/80 flex justify-between text-xs text-slate-500">
              <span>Semester {invoice.semester.toUpperCase()} {invoice.academic_year}</span>
              <span>Jatuh Tempo: {new Date(invoice.due_date).toLocaleDateString('id-ID')}</span>
            </div>
          </div>

          {/* Payment Method Selector */}
          <div>
            <label className="block text-xs font-semibold text-slate-700 uppercase tracking-wider mb-2">
              Pilih Kanal Pembayaran
            </label>
            <div className="grid grid-cols-2 gap-2.5">
              {[
                { id: 'BCA_VA', label: 'BCA Virtual Account' },
                { id: 'MANDIRI_VA', label: 'Mandiri Virtual Account' },
                { id: 'BNI_VA', label: 'BNI Virtual Account' },
                { id: 'QRIS', label: 'QRIS Nasional' },
              ].map((item) => (
                <button
                  key={item.id}
                  type="button"
                  onClick={() => setMethod(item.id)}
                  className={`p-3 rounded-xl border text-left text-xs font-semibold transition-all ${
                    method === item.id
                      ? 'border-blue-600 bg-blue-50/50 text-blue-800 ring-2 ring-blue-500/20'
                      : 'border-slate-200 text-slate-700 hover:border-slate-300'
                  }`}
                >
                  {item.label}
                </button>
              ))}
            </div>
          </div>

          {/* Security Test Option: Tamper Signature */}
          <div className="border border-slate-200 rounded-xl p-3 bg-slate-50/50 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ShieldAlert className="w-4 h-4 text-amber-600" />
              <div>
                <span className="text-xs font-semibold text-slate-800">Simulasikan HMAC Tamper</span>
                <p className="text-[11px] text-slate-500">Merusak signature untuk menguji proteksi HTTP 401</p>
              </div>
            </div>
            <input
              type="checkbox"
              id="tamper"
              checked={tamper}
              onChange={(e) => setTamper(e.target.checked)}
              className="w-4 h-4 text-blue-600 rounded border-slate-300 focus:ring-blue-500"
            />
          </div>

          {/* Execution Result Preview */}
          {resultData && (
            <div className="border border-emerald-200 bg-emerald-50/60 rounded-xl p-3 text-xs text-emerald-900 space-y-1">
              <div className="flex items-center gap-1.5 font-bold text-emerald-800">
                <CheckCircle2 className="w-4 h-4" />
                {resultData.message}
              </div>
              <div className="font-mono text-[11px] text-slate-600 truncate">
                Signature: {resultData.data?.calculated_signature}
              </div>
              <div className="text-[11px] font-semibold text-slate-700">
                Status Hasil: <span className="uppercase">{resultData.data?.webhook_result?.status}</span>
              </div>
            </div>
          )}

          {error && (
            <div className="border border-rose-200 bg-rose-50/60 rounded-xl p-3 text-xs text-rose-800">
              {(error as any)?.response?.data?.error?.message || error.message}
            </div>
          )}
        </div>

        {/* Footer Actions */}
        <div className="px-6 py-4 bg-slate-50 border-t border-slate-100 flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2 text-xs font-semibold text-slate-600 hover:text-slate-800 hover:bg-slate-200/50 rounded-lg transition-colors"
          >
            Tutup
          </button>
          <button
            type="button"
            onClick={handlePay}
            disabled={isPending}
            className="px-5 py-2 text-xs font-semibold bg-blue-600 hover:bg-blue-700 text-white rounded-lg shadow-sm flex items-center gap-2 transition-all disabled:opacity-50"
          >
            {isPending && <RefreshCw className="w-3.5 h-3.5 animate-spin" />}
            {isPending ? 'Mengirim Webhook...' : 'Kirim Notifikasi Pembayaran'}
          </button>
        </div>
      </div>
    </div>
  );
};
