# Software Requirements Specification (SRS)
# SEVIMA-PAY: Multi-Tenant UKT Payment System

> **Version**: 1.0.0  
> **Status**: Approved  
> **Standard**: IEEE 830-1998 Aligned

---

## 1. Introduction

### 1.1 Purpose
Dokumen ini mendefinisikan seluruh spesifikasi kebutuhan perangkat lunak (Software Requirements Specification) untuk **SEVIMA-PAY**, platform penagihan dan rekonsiliasi Uang Kuliah Tunggal (UKT) multi-kampus berbasis webhook payment gateway dengan arsitektur multi-tenant.

### 1.2 Scope of System
- Pengelolaan master data institusi/perguruan tinggi (Super Admin).
- Manajemen data tagihan UKT, pemantauan status real-time, rekonsiliasi manual, dan reporting (Admin Keuangan).
- Pengecekan tagihan mandiri, simulator pembayaran, dan verifikasi kelayakan KRS (Mahasiswa).
- Pemrosesan notifikasi webhook asynchronous dengan verifikasi signature HMAC-SHA256, proteksi idempotensi, dan anti-race condition (Payment Engine).
- Mock Payment Gateway terintegrasi untuk kebutuhan verifikasi end-to-end tanpa dependensi internet eksternal.

---

## 2. Overall Description

### 2.1 User Roles & Personas

| Role | Deskripsi Singkat | Ruang Lingkup Akses |
|------|-------------------|---------------------|
| `super_admin` | Administrator platform pusat SEVIMA | Mengelola institusi, memantau metrik agregat seluruh kampus, mengonfigurasi HMAC secret |
| `admin_keuangan` | Staf biro administrasi keuangan kampus | Mengelola tagihan UKT mahasiswa di kampusnya, melihat riwayat transaksi, export rekapitulasi |
| `mahasiswa` | Mahasiswa terdaftar di suatu institusi | Melihat riwayat tagihan pribadi, melakukan pembayaran, unduh kuitansi, melihat status eligibilitas KRS |
| `payment_gateway` | Sistem eksternal (Mock Gateway) | Mengirimkan callback notifikasi transaksi via webhook HTTP POST |

### 2.2 System Architecture Paradigm
- **Backend**: Go (Gin) dengan Clean Architecture (Domain, UseCase, Adapter, Infrastructure).
- **Frontend**: React 18/19 (Vite + TypeScript) dengan Tailwind CSS, shadcn/ui, dan Clean Architecture.
- **Data Persistence**: PostgreSQL 16+ dengan Row-Level Security (RLS) dan compound unique indexes untuk isolasi tenant.

---

## 3. Functional Requirements (FR)

### 3.1 Authentication & Multi-Tenancy (AUTH)

- **FR-AUTH-01**: Sistem harus menyediakan autentikasi berbasis email dan password dengan hash `bcrypt` (work factor 12).
- **FR-AUTH-02**: Sistem harus menerbitkan JSON Web Token (JWT) dengan access token berumur pendek (15 menit) dan refresh token (7 hari) yang disimpan secara aman dalam database.
- **FR-AUTH-03**: Token JWT harus memuat klaim identitas: `user_id`, `role`, dan `institusi_id` (kecuali `super_admin`).
- **FR-AUTH-04**: Sistem harus menerapkan isolasi tenant secara mutlak; pengguna dari Kampus A tidak boleh dapat membaca atau memanipulasi data Kampus B dalam kondisi apapun.

### 3.2 Super Admin Management (SAM)

- **FR-SAM-01**: Super Admin dapat mendaftarkan institusi perguruan tinggi baru (Kode Kampus, Nama, Kontak, Alamat).
- **FR-SAM-02**: Sistem harus meng-generate rahasia HMAC-SHA256 (32 byte secure random hex) secara otomatis untuk setiap institusi baru.
- **FR-SAM-03**: Super Admin dapat menonaktifkan atau mengaktifkan status institusi (`is_active`).
- **FR-SAM-04**: Super Admin dapat melihat audit log agregat dari semua aktivitas institusi.

### 3.3 Admin Keuangan (ADM)

- **FR-ADM-01**: Admin Keuangan dapat membuat tagihan UKT individual untuk mahasiswa di institusinya.
- **FR-ADM-02**: Admin Keuangan dapat mengimpor tagihan secara massal (bulk upload) dengan format CSV/Excel.
- **FR-ADM-03**: Nomor invoice harus unik dalam cakupan institusi (`UNIQUE(institusi_id, invoice_number)`).
- **FR-ADM-04**: Admin Keuangan dapat memfilter tagihan berdasarkan status (`unpaid`, `paid`, `expired`), semester (`ganjil`, `genap`), tahun akademik, dan program studi.
- **FR-ADM-05**: Admin Keuangan dapat melihat rincian riwayat perubahan status invoice dari tabel audit log.

### 3.4 Mahasiswa (MHS)

- **FR-MHS-01**: Mahasiswa hanya dapat melihat data tagihan yang terhubung dengan akun NIM miliknya.
- **FR-MHS-02**: Mahasiswa dapat melihat status eligibilitas KRS; status KRS dinyatakan "ELIGIBLE / SIAP KRS" hanya jika seluruh tagihan UKT semester aktif telah berstatus `paid`.
- **FR-MHS-03**: Mahasiswa dapat melihat rincian instruksi pembayaran (Nomor VA, Bank, Jumlah Tagihan, Batas Waktu).
- **FR-MHS-04**: Mahasiswa dapat mencetak atau mengunduh bukti tanda terima pembayaran resmi (Kuitansi PDF/Digital).

### 3.5 Webhook Processing & Payment Engine (WBH)

- **FR-WBH-01**: Endpoint webhook harus menerima request HTTP POST pada `/api/v1/webhook/payment`.
- **FR-WBH-02**: Sistem harus memverifikasi header `X-Signature` atau `X-Hub-Signature` menggunakan algoritma HMAC-SHA256 dengan secret key institusi terkait.
- **FR-WBH-03**: Jika verifikasi signature gagal, sistem harus menolak request dengan status HTTP `401 Unauthorized` dan mencatat insiden keamanan ke log webhook.
- **FR-WBH-04**: Sistem harus menjamin **idempotensi**; pengiriman webhook berulang (duplicate event) tidak boleh mengubah data lebih dari satu kali atau menyebabkan inkonsistensi saldo/status.
- **FR-WBH-05**: Sistem harus menerapkan penguncian baris database (`SELECT ... FOR UPDATE`) untuk mencegah race condition pada update transaksi bersamaan.
- **FR-WBH-06**: Transisi status tagihan harus tunduk pada Finite State Machine (FSM):
  - `unpaid` -> `paid` (Valid)
  - `unpaid` -> `expired` (Valid)
  - `paid` -> `paid` (Idempotent No-Op)
  - `expired` -> `paid` (Invalid / Rejected by default)

### 3.6 Mock Payment Gateway (MCK)

- **FR-MCK-01**: Sistem harus menyediakan simulator pembayaran terintegrasi untuk keperluan testing.
- **FR-MCK-02**: Simulator harus mampu membuat signature HMAC yang valid secara programatis dan mengirimkannya ke endpoint webhook lokal.
- **FR-MCK-03**: Simulator harus mendukung skenario pengujian: Pembayaran Sukses, Pembayaran Kedaluwarsa, Pengiriman Ulang (Duplicate), dan Simulasi Signature Rusak (Tampered).

---

## 4. Non-Functional Requirements (NFR)

### 4.1 Performance & Latency
- **NFR-PERF-01**: Waktu respons pemrosesan webhook (termasuk verifikasi signature dan commit database) harus kurang dari **200 ms** pada persentil 95 (p95).
- **NFR-PERF-02**: Query dashboard mahasiswa dan admin harus memiliki waktu respons HTTP di bawah **300 ms** untuk dataset hingga 100.000 tagihan.
- **NFR-PERF-03**: Database pooling harus dikonfigurasi secara optimal (`MaxOpenConns: 25`, `MaxIdleConns: 10`, `ConnMaxLifetime: 5m`).

### 4.2 Security & Data Privacy
- **NFR-SEC-01**: Seluruh password disimpan menggunakan `bcrypt` dengan cost factor 12.
- **NFR-SEC-02**: HMAC Secret Key untuk setiap institusi disimpan dengan enkripsi atau dilindungi dari paparan log/API response (diberi tag `json:"-"`).
- **NFR-SEC-03**: Sistem harus terlindung dari kerentanan OWASP Top 10 (SQL Injection melalui parameterized query `sqlc`, XSS melalui React JSX automatic encoding/escaping, CSRF melalui SameSite cookie & HTTP-only tokens).
- **NFR-SEC-04**: Row-Level Security (RLS) diaktifkan pada tingkat tabel database PostgreSQL sebagai proteksi sekunder (Defense-in-Depth).

### 4.3 Reliability & Availability
- **NFR-REL-01**: Sistem harus menangani webhook retry dari gateway tanpa menghasilkan duplikasi pencatatan uang.
- **NFR-REL-02**: Seluruh kegagalan webhook dicatat dalam tabel `webhook_logs` dengan payload mentah untuk memfasilitasi audit trail dan analisis forensik.
- **NFR-REL-03**: Background worker auto-expiry berjalan setiap 60 detik secara non-blocking dengan query batch `FOR UPDATE SKIP LOCKED`.

### 4.4 Usability & Aesthetics
- **NFR-UI-01**: Tampilan antarmuka mengikuti standar modern: responsif (mobile-friendly), skema warna berestetika tinggi (TailwindCSS Slate/Blue), hierarki visual jelas, dan micro-interactions yang halus.
- **NFR-UI-02**: Aksesibilitas memenuhi pedoman WCAG 2.1 level AA (kontras rasio minimum 4.5:1 untuk teks biasa).

---

## 5. Verification Matrix (Traceability)

| Requirement ID | Fitur | Kriteria Verifikasi |
|----------------|-------|---------------------|
| `FR-AUTH-04` | Tenant Isolation | Test case: Mahasiswa Kampus A mengakses `/api/v1/invoices/:id` milik Kampus B -> Wajib HTTP 403/404 |
| `FR-WBH-02` | HMAC Verification | Test case: Kirim payload dengan signature acak -> Wajib HTTP 401 |
| `FR-WBH-04` | Idempotency | Test case: Kirim webhook `payment.success` yang sama 5 kali -> Status invoice tetap `paid`, hanya 1 history dicatat |
| `FR-WBH-05` | Concurrency | Test case: 10 concurrent requests memicu pembayaran pada 1 invoice -> Tepat 1 request berhasil update, 9 idempotently handled |
