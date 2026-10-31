import React, { useState } from 'react';
import { useMockPaymentMutation } from '@/application/api/queries';
import { CreditCard, ShieldCheck, ShieldAlert, RefreshCw, Send, CheckCircle2, AlertOctagon } from 'lucide-react';

export const MockGatewayPage: React.FC = () => {
  const [instCode, setInstCode] = useState('UGM');
  const [invoiceNumber, setInvoiceNumber] = useState('INV-2026-UGM-0001');
  const [method, setMethod] = useState('BCA_VA');
  const [eventType, setEventType] = useState('payment.success');
  const [tamper, setTamper] = useState(false);
  const [lastResponse, setLastResponse] = useState<any>(null);

  const { mutate: triggerPayment, isPending, error } = useMockPaymentMutation();

  const handleDispatch = (e: React.FormEvent) => {
    e.preventDefault();
    triggerPayment(
      {
        institution_code: instCode,
        invoice_number: invoiceNumber,
        payment_method: method,
        simulate_tamper: tamper,
        event_type: eventType,
      },
      {
        onSuccess: (data) => {
          setLastResponse(data);
        },
        onError: (err: any) => {
          setLastResponse({
            success: false,
            error: err.response?.data?.error || { message: err.message },
          });
        },
      }
    );
  };

  return (
    <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-6">
      <div>
        <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-purple-100 text-purple-800 text-xs font-bold mb-2">
          <CreditCard className="w-3.5 h-3.5" />
          Simulator Sandbox
        </div>
        <h1 className="text-2xl font-bold text-slate-900 tracking-tight">Payment Gateway & Webhook Simulator</h1>
        <p className="text-sm text-slate-500 mt-1">
          Simulator pengujian live coding untuk memverifikasi kalkulasi HMAC-SHA256, idempotensi duplikat webhook, dan transisi status transaksi.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left: Simulator Controls */}
        <div className="lg:col-span-5 bg-white rounded-2xl border border-slate-200 p-6 shadow-xs space-y-5">
          <h2 className="font-bold text-slate-900 text-sm flex items-center gap-2">
            <Send className="w-4 h-4 text-purple-600" />
            Konfigurasi Notifikasi Webhook
          </h2>

          <form onSubmit={handleDispatch} className="space-y-4">
            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">Institusi / Kampus</label>
              <select
                value={instCode}
                onChange={(e) => {
                  setInstCode(e.target.value);
                  if (e.target.value === 'UGM') setInvoiceNumber('INV-2026-UGM-0001');
                  else setInvoiceNumber('INV-2026-ITS-0001');
                }}
                className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-bold"
              >
                <option value="UGM">UGM · Universitas Gadjah Mada</option>
                <option value="ITS">ITS · Institut Teknologi Sepuluh Nopember</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">Nomor Invoice Tagihan</label>
              <input
                type="text"
                required
                value={invoiceNumber}
                onChange={(e) => setInvoiceNumber(e.target.value)}
                className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs font-mono font-semibold"
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">Tipe Event Webhook</label>
              <select
                value={eventType}
                onChange={(e) => setEventType(e.target.value)}
                className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
              >
                <option value="payment.success">payment.success (Pelunasan Berhasil)</option>
                <option value="payment.expired">payment.expired (Tagihan Kedaluwarsa)</option>
              </select>
            </div>

            <div>
              <label className="block text-xs font-semibold text-slate-700 mb-1">Metode Pembayaran</label>
              <select
                value={method}
                onChange={(e) => setMethod(e.target.value)}
                className="w-full px-3 py-2 bg-slate-50 border border-slate-200 rounded-xl text-xs"
              >
                <option value="BCA_VA">BCA Virtual Account</option>
                <option value="MANDIRI_VA">Mandiri Virtual Account</option>
                <option value="BNI_VA">BNI Virtual Account</option>
                <option value="QRIS">QRIS Standar</option>
              </select>
            </div>

            {/* Tamper Switch */}
            <div className="p-3 bg-amber-50/70 border border-amber-200 rounded-xl flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ShieldAlert className="w-4 h-4 text-amber-600" />
                <div>
                  <span className="text-xs font-bold text-amber-950">Simulasikan HMAC Rusak (Tamper)</span>
                  <p className="text-[11px] text-amber-800">Menyuntikkan signature acak</p>
                </div>
              </div>
              <input
                type="checkbox"
                checked={tamper}
                onChange={(e) => setTamper(e.target.checked)}
                className="w-4 h-4 text-purple-600 rounded border-amber-300"
              />
            </div>

            <button
              type="submit"
              disabled={isPending}
              className="w-full py-2.5 bg-purple-600 hover:bg-purple-700 text-white rounded-xl text-xs font-bold shadow-sm transition-all flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {isPending ? (
                <>
                  <RefreshCw className="w-4 h-4 animate-spin" />
                  Mengirimkan Webhook...
                </>
              ) : (
                <>
                  <Send className="w-4 h-4" />
                  Kirim Notifikasi Webhook Sekarang
                </>
              )}
            </button>
          </form>
        </div>

        {/* Right: Technical Inspector */}
        <div className="lg:col-span-7 space-y-4">
          <div className="bg-slate-900 rounded-2xl p-6 text-white shadow-xl space-y-4 border border-slate-800">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-xs font-mono font-bold uppercase tracking-wider text-purple-400">
                Log Inspector & Hasil Eksekusi Webhook
              </h3>
              <span className="text-[10px] text-slate-400 font-mono">POST /api/v1/webhook/payment</span>
            </div>

            {lastResponse ? (
              <div className="space-y-4 text-xs font-mono">
                <div>
                  <span className="text-slate-400 block mb-1 font-sans font-semibold">Status Respons HTTP:</span>
                  {lastResponse.success ? (
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-emerald-950 text-emerald-400 border border-emerald-800 font-bold">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      HTTP 200 OK — {lastResponse.message}
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded bg-rose-950 text-rose-400 border border-rose-800 font-bold">
                      <AlertOctagon className="w-3.5 h-3.5" />
                      HTTP 400/401 Ditolak — {lastResponse.error?.message}
                    </span>
                  )}
                </div>

                <div>
                  <span className="text-slate-400 block mb-1 font-sans font-semibold">HMAC-SHA256 Signature Dikirim:</span>
                  <div className="p-2.5 bg-slate-950 rounded-lg text-emerald-300 break-all select-all border border-slate-800">
                    {lastResponse.data?.calculated_signature || lastResponse.error?.details?.calculated_signature || 'N/A'}
                  </div>
                </div>

                <div>
                  <span className="text-slate-400 block mb-1 font-sans font-semibold">Payload JSON:</span>
                  <pre className="p-3 bg-slate-950 rounded-lg text-slate-300 overflow-x-auto text-[11px] border border-slate-800">
                    {JSON.stringify(lastResponse.data?.webhook_payload || lastResponse.error?.details?.webhook_payload, null, 2)}
                  </pre>
                </div>

                {lastResponse.data?.webhook_result && (
                  <div>
                    <span className="text-slate-400 block mb-1 font-sans font-semibold">Hasil Mesin FSM Status:</span>
                    <div className="p-2.5 bg-slate-950 rounded-lg text-purple-300 border border-slate-800 space-y-1">
                      <div>Hasil: <span className="font-bold text-white uppercase">{lastResponse.data.webhook_result.status}</span></div>
                      <div>Pesan: {lastResponse.data.webhook_result.message}</div>
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <div className="py-16 text-center text-slate-500 text-xs font-sans">
                Belum ada notifikasi webhook yang dieksekusi. Klik tombol kirim di samping untuk memulai simulasi.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};
