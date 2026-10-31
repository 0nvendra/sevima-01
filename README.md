# SEVIMA-PAY: Platform Multi-Tenant Pembayaran UKT via Webhook

Platform terintegrasi pengelolaan tagihan Uang Kuliah Tunggal (UKT) multi-institusi perguruan tinggi dengan penerimaan pembayaran berbasis Webhook (HMAC-SHA256), kontrol akses berbasis peran (RBAC), isolasi data multi-tenant ketat, dan simulasi Mock Payment Gateway.

---

## 1. Prasyarat Sistem & Dependensi

Pastikan perangkat Anda telah terinstall:
- **Go**: Versi 1.22 atau lebih baru
- **Node.js**: Versi 18+ & npm
- **PostgreSQL**: Versi 14+ (disarankan v16)
- **Git**

---

## 2. Panduan Setup & Instalasi

### 2.1 Konfigurasi Database PostgreSQL
1. Buat database baru bernama `sevima_pay`:
   ```sql
   CREATE DATABASE sevima_pay;
   ```
2. Jalankan skrip DDL/Schema PostgreSQL:
   ```bash
   psql -U postgres -d sevima_pay -f sevima-docs/02_SQL_Schema_MultiTenant.sql
   ```

### 2.2 Setup Backend (Go)
1. Buka direktori backend:
   ```bash
   cd sevima-be
   ```
2. Salin file environment:
   ```bash
   cp .env.example .env
   ```
   *Sesuaikan kredensial `DB_USER`, `DB_PASSWORD`, `DB_HOST`, dan `DB_PORT` bila berbeda.*
3. Download dependensi Go:
   ```bash
   go mod download
   ```
4. Jalankan Seeding Data Awal (Institusi, Mahasiswa, Akun Demo, Tagihan):
   ```bash
   go run cmd/seed/main.go
   ```
5. Jalankan server Backend:
   ```bash
   go run cmd/api/main.go
   ```
   *Backend akan aktif di `http://localhost:8080`.*

### 2.3 Setup Frontend (React Vite + TypeScript)
1. Buka direktori frontend:
   ```bash
   cd ../sevima-fe
   ```
2. Salin file environment:
   ```bash
   cp .env.example .env
   ```
3. Install package:
   ```bash
   npm install
   ```
4. Jalankan development server:
   ```bash
   npm run dev
   ```
   *Frontend dapat diakses di browser melalui `http://localhost:5173`.*

---

## 3. Akun Demo untuk Setiap Role

Semua akun demo di bawah ini menggunakan password default yang sama:
> **Password**: `Password123!`

| Role | Nama Pengguna | Email | Institusi | Cakupan Akses |
|---|---|---|---|---|
| **Super Admin** | Super Administrator SEVIMA | `superadmin@sevima.com` | Global (Platform) | Akses seluruh institusi, kelola institusi, generate HMAC secret |
| **Admin Keuangan** | Budi Santoso | `admin@ugm.ac.id` | UGM (Yogyakarta) | Kelola tagihan & transaksi khusus institusi UGM |
| **Admin Keuangan** | Dewi Lestari | `admin@its.ac.id` | ITS (Surabaya) | Kelola tagihan & transaksi khusus institusi ITS |
| **Mahasiswa** | Ahmad Fauzi (NIM: 21/475123/TK/52111) | `fauzi@mail.ugm.ac.id` | UGM | Riwayat tagihan sendiri & status kelayakan KRS UGM |
| **Mahasiswa** | Bima Sakti (NIM: 5025211001) | `bima@mhs.its.ac.id` | ITS | Riwayat tagihan sendiri & status kelayakan KRS ITS |

---

## 4. Cara Mengakses & Menguji Fitur Utama

### 4.1 Menguji Isolasi Data Antar-Institusi (Data Segregation)
1. Login sebagai `admin@ugm.ac.id` (Admin UGM). Masuk ke menu **Daftar Tagihan**. Anda hanya akan melihat data tagihan mahasiswa UGM (`INV-2026-UGM-0001`).
2. Buka tab incognito atau logout, lalu login sebagai `admin@its.ac.id` (Admin ITS). Anda hanya akan melihat tagihan mahasiswa ITS (`INV-2026-ITS-0001`). Data UGM tidak dapat diakses maupun dimanipulasi.
3. Login sebagai mahasiswa UGM `fauzi@mail.ugm.ac.id`. Tagihan yang tampil hanya milik dirinya sendiri.

### 4.2 Simulasi Pembayaran Webhook via Mock Gateway UI
1. Akses menu **Mock Payment Gateway** pada sidebar navigasi (`/mock-gateway`).
2. Masukkan nomor tagihan (misal: `INV-2026-UGM-0001`), pilih institusi `UGM`, nominal `7500000`, metode pembayaran (VA/QRIS), lalu klik **Simulasikan Pembayaran**.
3. Sistem akan menghasilkan HMAC-SHA256 signature secara otomatis menggunakan secret institusi terkait, lalu mengirimkan payload ke endpoint webhook.
4. Kembali ke menu Tagihan / KRS Mahasiswa UGM: status seketika berubah menjadi **Lunas (Paid)** dan status eligibilitas KRS mahasiswa berubah menjadi **Layak Mengisi KRS**.

### 4.3 Menjalankan Test Otomatis
Jalankan unit test backend untuk logika isolasi multi-tenant, transisi status tagihan, dan verifikasi kriptografi:
```bash
cd sevima-be
go test -v ./...
```

---

## 5. Asumsi & Batasan Sistem

1. **Model Multi-Tenancy**: Menggunakan pola *Shared Database, Shared Schema with Column Discriminator* (`institusi_id`). Setiap query DML/SELECT pada layer aplikasi diproteksi oleh middleware JWT dan klausul `WHERE institusi_id = ?`.
2. **KRS Eligibility Rule**: Mahasiswa dinyatakan eligible/berhak mengisi Kartu Rencana Studi (KRS) pada semester berjalan jika dan hanya jika tidak ada tagihan UKT aktif yang berstatus `unpaid`.
3. **Webhook Idempotency**: Webhook yang dikirim ulang dengan payload transaksi yang sama tidak akan menyebabkan tagihan diproses ulang (mengembalikan status idempotensi sukses `duplicate`).
4. **Keamanan Webhook**: Setiap institusi memiliki secret HMAC-SHA256 unik yang tidak dibagikan antar institusi.
5. **Background Auto-Expiry**: Background worker secara berkala memeriksa tagihan yang melewati batas tanggal jatuh tempo (`due_date`) dan memperbarui statusnya menjadi `expired`.
