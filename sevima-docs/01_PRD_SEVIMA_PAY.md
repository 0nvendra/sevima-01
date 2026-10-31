# Product Requirements Document (PRD)
# SEVIMA-PAY: Integrasi Pembayaran UKT via Webhook

> **Version**: 1.0.0
> **Status**: Draft
> **Author**: Engineering Team
> **Date**: 2026-10-30
> **Stakeholder**: SEVIMA Technical Assessment

---

## 1. Executive Summary

SEVIMA-PAY adalah platform multi-tenant untuk mengelola tagihan Uang Kuliah Tunggal (UKT) mahasiswa lintas institusi pendidikan tinggi. Sistem menerima notifikasi pembayaran dari payment gateway pihak ketiga melalui webhook, memperbarui status tagihan secara real-time, dan menyediakan dashboard terpisah untuk Admin Keuangan dan Mahasiswa.

### 1.1 Problem Statement

Banyak kampus menagih UKT setiap semester melalui platform yang sama, dengan pembayaran lewat payment gateway pihak ketiga. Status pembayaran harus akurat karena menentukan apakah mahasiswa dapat mengisi KRS. Diperlukan sistem terpusat yang:

- Mengelola tagihan UKT multi-institusi dalam satu database
- Menjamin isolasi data antar-institusi (data segregation)
- Menerima dan memverifikasi notifikasi pembayaran via webhook dengan HMAC-SHA256
- Menyediakan audit trail lengkap untuk setiap transaksi

### 1.2 Success Metrics

| Metric | Target |
|--------|--------|
| Webhook processing time | < 500ms p95 |
| Data isolation violation | 0 (zero tolerance) |
| Invoice status accuracy | 100% konsisten dengan notifikasi gateway |
| System availability | 99.9% uptime |
| Audit log completeness | 100% notifikasi tercatat |

---

## 2. User Personas

### 2.1 Super Admin (Platform)

- **Siapa**: Administrator platform SEVIMA-PAY
- **Kebutuhan**: Mengelola institusi, melihat statistik keseluruhan platform, mengelola HMAC secret per institusi
- **Akses**: Semua data lintas institusi
- **Frekuensi**: Harian

### 2.2 Admin Keuangan (Per Institusi)

- **Siapa**: Staff bagian keuangan di setiap kampus
- **Kebutuhan**: Membuat tagihan UKT, memantau status pembayaran, melihat riwayat transaksi institusinya
- **Akses**: Hanya data institusinya sendiri (scoped by `institusi_id`)
- **Frekuensi**: Harian, terutama saat awal semester

### 2.3 Mahasiswa

- **Siapa**: Mahasiswa terdaftar di institusi
- **Kebutuhan**: Melihat tagihan UKT miliknya, melihat status pembayaran, melihat riwayat transaksi
- **Akses**: Hanya data miliknya sendiri di institusinya
- **Frekuensi**: Setiap awal semester atau saat membayar

---

## 3. Functional Requirements

### 3.1 Authentication & Authorization (RBAC)

| ID | Requirement | Priority |
|----|-------------|----------|
| AUTH-01 | Login dengan email + password menghasilkan JWT (Access Token + Refresh Token) | Must |
| AUTH-02 | Access Token berlaku 15 menit, Refresh Token 7 hari | Must |
| AUTH-03 | Middleware RBAC memvalidasi role (super_admin, admin_keuangan, mahasiswa) | Must |
| AUTH-04 | Middleware tenant isolation memfilter data berdasarkan `institusi_id` dari token | Must |
| AUTH-05 | Refresh token endpoint untuk mendapatkan access token baru | Must |
| AUTH-06 | Logout endpoint yang invalidasi refresh token | Should |

### 3.2 Manajemen Institusi (Super Admin)

| ID | Requirement | Priority |
|----|-------------|----------|
| INST-01 | CRUD institusi (nama, kode, alamat, status aktif) | Must |
| INST-02 | Generate dan manage HMAC secret per institusi | Must |
| INST-03 | Lihat daftar semua institusi dengan statistik tagihan | Should |
| INST-04 | Kode institusi unik di seluruh platform | Must |

### 3.3 Manajemen Tagihan UKT (Admin Keuangan)

| ID | Requirement | Priority |
|----|-------------|----------|
| BILL-01 | Membuat tagihan UKT per mahasiswa per semester | Must |
| BILL-02 | Data tagihan: nomor tagihan, nominal, tanggal kedaluwarsa, semester, tahun akademik | Must |
| BILL-03 | Nomor tagihan unik di dalam satu institusi | Must |
| BILL-04 | Daftar tagihan dengan filter: status, semester, tahun akademik, nama mahasiswa | Must |
| BILL-05 | Status tagihan: `unpaid`, `paid`, `expired` | Must |
| BILL-06 | Admin hanya bisa melihat/membuat tagihan untuk institusinya | Must |
| BILL-07 | Bulk create tagihan (CSV import) | Could |
| BILL-08 | Export daftar tagihan ke CSV | Could |

### 3.4 Webhook Payment Gateway

| ID | Requirement | Priority |
|----|-------------|----------|
| WH-01 | Endpoint POST `/api/v1/webhook/payment` menerima notifikasi | Must |
| WH-02 | Verifikasi HMAC-SHA256 signature dari header `X-Signature` | Must |
| WH-03 | Support event: `payment.success`, `payment.expired` | Must |
| WH-04 | Payload membawa `institution_code` untuk identifikasi institusi | Must |
| WH-05 | Idempotency: notifikasi duplikat tidak mengubah state yang sudah benar | Must |
| WH-06 | Semua notifikasi dicatat di tabel `webhook_logs` untuk audit | Must |
| WH-07 | Response 200 OK setelah processing (gateway expects acknowledgment) | Must |
| WH-08 | Response 401 jika signature invalid | Must |
| WH-09 | Response 404 jika tagihan tidak ditemukan | Must |

### 3.5 Riwayat Transaksi

| ID | Requirement | Priority |
|----|-------------|----------|
| TXN-01 | Setiap perubahan status tagihan menghasilkan record di `transaction_histories` | Must |
| TXN-02 | Data riwayat: status sebelum, status sesudah, amount, timestamp, source | Must |
| TXN-03 | Endpoint untuk melihat riwayat transaksi per tagihan | Must |
| TXN-04 | Riwayat bersifat append-only (tidak bisa diubah/dihapus) | Must |

### 3.6 Dashboard Mahasiswa

| ID | Requirement | Priority |
|----|-------------|----------|
| STD-01 | Melihat semua tagihan UKT miliknya | Must |
| STD-02 | Melihat detail tagihan dan riwayat transaksi | Must |
| STD-03 | Filter tagihan per semester / tahun akademik | Should |
| STD-04 | Ringkasan: total tagihan, sudah bayar, belum bayar | Should |

### 3.7 Mock Payment Gateway

| ID | Requirement | Priority |
|----|-------------|----------|
| MOCK-01 | Service terpisah untuk simulasi payment gateway | Must |
| MOCK-02 | UI sederhana untuk trigger webhook manually | Must |
| MOCK-03 | Generate signature HMAC-SHA256 yang valid | Must |
| MOCK-04 | Support kirim event `payment.success` dan `payment.expired` | Must |
| MOCK-05 | Dokumentasi format payload dan cara penggunaan | Must |

---

## 4. Non-Functional Requirements

| Category | Requirement | Target |
|----------|-------------|--------|
| Security | Data isolation antar-institusi di setiap query | Zero cross-tenant leak |
| Security | HMAC-SHA256 signature verification pada webhook | Timing-safe comparison |
| Security | Password hashing dengan bcrypt (cost 12) | Industry standard |
| Security | Secrets tidak boleh ada di repository | Environment variables |
| Performance | API response time | < 200ms p95 |
| Performance | Webhook processing | < 500ms p95 |
| Reliability | Database transactions untuk state changes | ACID compliance |
| Reliability | Idempotent webhook processing | Duplicate-safe |
| Observability | Structured logging (JSON) | Semua endpoint |
| Observability | Webhook audit trail | 100% coverage |
| Deployment | Docker Compose untuk semua services | One-command setup |
| Testing | Unit + Integration tests | Core business logic |

---

## 5. Out of Scope

- Integrasi payment gateway sungguhan
- Halaman pembayaran untuk mahasiswa (hanya status dan riwayat)
- Email notification
- Mobile app
- Real-time websocket updates
- Multi-currency support
- Partial payment

---

## 6. Data Model Overview

### Core Entities

```
institutions        ─── 1:N ──→ users
institutions        ─── 1:N ──→ students
institutions        ─── 1:N ──→ invoices
invoices            ─── 1:N ──→ transaction_histories
institutions        ─── 1:N ──→ webhook_logs
```

### Tenant Isolation Strategy

Setiap tabel yang mengandung data spesifik institusi memiliki kolom `institusi_id`. Middleware backend memastikan:

1. **Token-based scoping**: JWT mengandung `institusi_id` yang di-inject ke setiap query
2. **Query-level filtering**: Semua query menyertakan `WHERE institusi_id = $1`
3. **Row-Level Security (RLS)**: PostgreSQL RLS sebagai defense-in-depth layer
4. **API-level validation**: Input yang mereferensikan entity dari institusi lain ditolak

---

## 7. Business Rules

### 7.1 Invoice Lifecycle

```
[Created] ──→ unpaid ──→ paid (via webhook payment.success)
                │
                └──→ expired (via webhook payment.expired ATAU
                              cron job saat tanggal kedaluwarsa terlewati)
```

### 7.2 State Transition Rules

| From | To | Trigger | Valid |
|------|----|---------|-------|
| unpaid | paid | payment.success webhook | ✅ |
| unpaid | expired | payment.expired webhook / scheduled job | ✅ |
| paid | paid | duplicate payment.success | ✅ (idempotent, no-op) |
| paid | expired | payment.expired setelah paid | ❌ (reject, log warning) |
| expired | paid | late payment.success | ⚠️ (configurable, log for review) |
| expired | expired | duplicate payment.expired | ✅ (idempotent, no-op) |

### 7.3 HMAC Verification Flow

1. Gateway mengirim POST ke webhook endpoint
2. Header `X-Signature` berisi HMAC-SHA256(body, institution_secret)
3. Backend extract `institution_code` dari body
4. Lookup secret institusi dari database
5. Compute HMAC-SHA256 dari raw body dengan secret
6. Compare dengan `X-Signature` menggunakan constant-time comparison
7. Jika tidak cocok → 401 Unauthorized, log attempt
8. Jika cocok → process event

### 7.4 Invoice Number Uniqueness

- Nomor tagihan unik **di dalam satu institusi**
- Boleh sama di institusi berbeda
- Format: configurable per institusi (default: `INV-{YYYY}{MM}-{SEQ}`)
- Enforced via composite unique index `(institusi_id, invoice_number)`

---

## 8. User Interface Requirements

### 8.1 Pages Overview

| Page | Role | Description |
|------|------|-------------|
| Login | All | Form login dengan email + password |
| Super Admin Dashboard | Super Admin | Statistik platform, daftar institusi |
| Manage Institutions | Super Admin | CRUD institusi |
| Admin Dashboard | Admin Keuangan | Statistik tagihan institusinya |
| Invoice List | Admin Keuangan | Tabel tagihan dengan filter + pagination |
| Create Invoice | Admin Keuangan | Form buat tagihan baru |
| Student Dashboard | Mahasiswa | Ringkasan tagihan |
| Student Invoices | Mahasiswa | Daftar tagihan + detail + riwayat |

### 8.2 Design Requirements

- Responsive design (mobile-first)
- Dark mode support
- Loading states dan skeleton screens
- Error handling dengan user-friendly messages
- Accessible (WCAG 2.1 AA minimum)
- Indonesian language as default

---

## 9. Seed Data

Minimal 3 institusi dengan data berikut:

| Institusi | Kode | Admin Keuangan | Mahasiswa |
|-----------|------|----------------|-----------|
| Universitas Nusantara | UNIV-NUS | admin@univ-nus.ac.id | 3 mahasiswa |
| Institut Teknologi Merdeka | ITM | admin@itm.ac.id | 3 mahasiswa |
| Politeknik Digital Indonesia | PDI | admin@pdi.ac.id | 3 mahasiswa |

- Setiap institusi memiliki HMAC secret sendiri
- Setiap mahasiswa memiliki 1-2 tagihan dengan status bervariasi
- Super Admin: superadmin@sevima-pay.com

---

## 10. Deliverables

1. Source code (Go backend + React frontend + Mock Gateway)
2. Docker Compose untuk menjalankan semua services
3. README.md (setup, run, akses, akun demo)
4. CLAUDE.md (konteks AI)
5. AI_NOTES.md (refleksi pairing AI)
6. DECISIONS.md (keputusan teknis)
7. API documentation (endpoint webhook)
8. Automated tests (unit + integration)
9. Database migration files
10. Seed data script

---

## 11. Timeline Estimation

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| 1. Foundation | 3h | Database schema, migrations, Docker Compose setup |
| 2. Backend Core | 4h | Auth, RBAC, invoice CRUD, webhook endpoint |
| 3. Frontend Core | 3h | Login, Admin dashboard, Student dashboard |
| 4. Integration | 2h | Mock gateway, webhook flow, testing |
| 5. Polish | 2h | Documentation, seed data, cleanup |
| **Total** | **14h** | Full working application |

---

## Appendix A: Glossary

| Term | Definition |
|------|-----------|
| UKT | Uang Kuliah Tunggal — biaya kuliah per semester |
| KRS | Kartu Rencana Studi — pendaftaran mata kuliah yang memerlukan status pembayaran lunas |
| HMAC | Hash-based Message Authentication Code |
| Webhook | HTTP callback yang dipicu oleh event dari payment gateway |
| Multi-tenant | Arsitektur di mana satu instance aplikasi melayani banyak organisasi dengan data terisolasi |
| RBAC | Role-Based Access Control |
| Idempotent | Operasi yang menghasilkan hasil sama meskipun dipanggil berulang kali |
