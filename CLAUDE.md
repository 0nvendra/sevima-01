# Project: SEVIMA-PAY - Integrasi Pembayaran UKT via Webhook

## Deskripsi singkat
SEVIMA-PAY adalah platform SaaS multi-tenant untuk manajemen tagihan Uang Kuliah Tunggal (UKT) mahasiswa lintas institusi perguruan tinggi. Sistem menerima notifikasi pembayaran dari payment gateway pihak ketiga via webhook, memperbarui status tagihan secara real-time, mengunci kelayakan pengisian KRS mahasiswa, serta menjamin pemisahan data (data segregation) antar-institusi tanpa kebocoran.

## Stack dan cara menjalankan
- **Backend**: Go 1.22+ (Gin Web Framework, pgx/v5, golang-jwt/v5, crypto).
- **Frontend**: React 18, Vite, TypeScript, TailwindCSS, Zustand / TanStack Query, Lucide Icons.
- **Database**: PostgreSQL 16.
- **Cara Menjalankan**:
  - Database: `psql -U postgres -d sevima_pay -f sevima-docs/02_SQL_Schema_MultiTenant.sql`
  - Backend: Masuk ke `sevima-be/`, salin `.env.example` ke `.env`, jalankan `go run cmd/seed/main.go` lalu `go run cmd/api/main.go` (berjalan di port 8080).
  - Frontend: Masuk ke `sevima-fe/`, jalankan `npm install` lalu `npm run dev` (berjalan di port 5173).
  - Testing: Di `sevima-be/`, jalankan `go test -v ./...`.

## Arsitektur singkat
Sistem mengadopsi prinsip **Clean Architecture** (Onion Architecture):
1. **Domain Layer**: Entity model murni (`entity/models.go`) dan kontrak repository (`domain/repository/interfaces.go`) tanpa dependensi ke framework luar.
2. **Use Case Layer**: Orkestrator logika bisnis (`usecase/usecases.go`), mencakup pemrosesan webhook, validasi HMAC, state machine tagihan, dan validasi kelayakan KRS.
3. **Adapter Layer**: 
   - HTTP Handlers (`adapter/handler/handlers.go`)
   - Middlewares (`adapter/middleware/middleware.go` untuk Auth JWT, Tenant Scoping, dan RBAC)
   - Database Repository (`adapter/repository/pg_repos.go` menggunakan pgx connection pool)
4. **Mock Gateway**: Endpoint internal simulator webhook untuk pengujian siklus pembayaran tanpa dependensi payment gateway nyata.

## Aturan bisnis penting (termasuk multi-institusi)
1. **Multi-Tenancy & Isolasi Data**:
   - Kolom `institusi_id` wajib ada pada setiap entitas institusional (`users`, `students`, `invoices`, `transaction_histories`, `webhook_logs`).
   - Admin Keuangan dan Mahasiswa hanya memiliki akses ke data dengan `institusi_id` yang sesuai dengan token JWT mereka. Akses lintas institusi mutlak dilarang (zero-tolerance cross-tenant access).
   - Super Admin dapat melihat lintas institusi dan dapat menyaring institusi menggunakan header `X-Institution-ID` atau query parameter.
2. **KRS Eligibility (Kelayakan KRS)**:
   - Mahasiswa hanya berhak mengisi KRS (`is_eligible_krs = true`) jika tidak memiliki tagihan berstatus `unpaid` pada semester dan tahun akademik aktif.
3. **Webhook Verification & Idempotensi**:
   - Setiap payload webhook diverifikasi menggunakan signature `X-Signature` dengan algoritma HMAC-SHA256 berdasar secret institusi terkait.
   - Idempotensi: Notifikasi ganda untuk tagihan yang sudah `paid` harus menghasilkan respons sukses tanpa mengubah state atau menambah transaksi ganda.
   - Tagihan yang sudah `expired` tidak boleh dibayar. Nominal yang dibayarkan harus sama persis dengan nominal tagihan.
4. **Audit Trail**:
   - Setiap mutasi status tagihan (`unpaid` -> `paid`, `unpaid` -> `expired`) wajib mencatat riwayat transaksi di tabel `transaction_histories`.
   - Setiap webhook yang masuk (baik valid maupun invalid) wajib tercatat di tabel `webhook_logs`.

## Konvensi kode
1. **Go (Backend)**:
   - Mengikuti idiomatic Go (Effective Go).
   - Penanganan error eksplisit (`if err != nil`), tidak mengabaikan error kritis.
   - Nama fungsi public menggunakan PascalCase, private camelCase.
   - Struktur Clean Architecture yang ketat: Handler -> Use Case -> Repository.
2. **TypeScript/React (Frontend)**:
   - Komponen fungsional dengan TypeScript typing eksplisit.
   - Struktur Clean Architecture / Feature-Driven di `src/presentation`, `src/application`, `src/domain`, `src/core`.
   - Tidak menggunakan generic/placeholder UI; menggunakan styling modern, responsive, dan ramah pengguna.
3. **Keamanan & Kredensial**:
   - Tidak boleh melakukan hardcode secret, API key, atau token credential ke dalam kode repositori.
   - Selalu gunakan environment variable (`.env`).

## Hal yang tidak boleh dilakukan AI
1. Jangan membypass atau menghapus validasi `institusi_id` pada query database atau handler endpoint mana pun.
2. Jangan mengizinkan akses ke tagihan atau mahasiswa milik institusi lain tanpa otorisasi Super Admin.
3. Jangan mengizinkan perubahan status tagihan dari `paid` atau `expired` kembali ke status lainnya secara sembarangan.
4. Jangan menyimpan atau meng-commit password plain text, private key, JWT secret, atau HMAC secret ke dalam repository git.
5. Jangan mengubah response envelope standar API (`status`, `message`, `data`, `meta`) yang dapat merusak integrasi frontend.
