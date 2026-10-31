# Refleksi Pairing dengan AI — SEVIMA-PAY

## Bagian yang dikerjakan bersama AI
- **Perancangan Skema Multi-Tenant & DDL PostgreSQL**: Berkolaborasi dengan AI dalam mendesain tabel database dengan `institusi_id` sebagai discriminator column, pembuatan indeks performa (`btree`, unique index komposit), serta foreign key constraint.
- **Implementasi Clean Architecture Backend (Go)**: Menyusun pembagian layer yang bersih antara Domain Model, Repository Interface, Use Case Orchestrator, dan Gin HTTP Handlers.
- **Kriptografi & Webhook Engine**: Membangun mekanisme verifikasi signature HMAC-SHA256 untuk webhook payment gateway, hashing password bcrypt, dan penanganan JWT access token + refresh token.
- **Mock Payment Gateway Simulator**: Merancang endpoint dan antarmuka simulator interaktif pada frontend yang menghitung tanda tangan digital secara real-time dan menembakkan notifikasi webhook tiruan untuk pengujian end-to-end tanpa memerlukan gateway fisik.
- **Penyusunan Test Suite Otomatis**: Menulis unit test untuk menguji isolasi data multi-tenant, middleware otorisasi RBAC, dan kepatuhan transisi status tagihan (state machine).

## Satu contoh saran AI yang saya tolak/koreksi, dan alasannya
- **Saran Awal AI**: AI awalnya menyarankan penggunaan pendekatan *Database-per-Tenant* atau *Schema-per-Tenant* di PostgreSQL dengan alasan isolasi fisik yang mutlak, serta menyarankan penggunaan ORM berat seperti GORM.
- **Alasan Penolakan/Koreksi**: 
  1. *Database/Schema per Tenant* akan meningkatkan kompleksitas operasional secara eksponensial (maintenance migrasi schema pada puluhan kampus, overhead koneksi pool database terfragmentasi, dan query agregasi laporan nasional Super Admin menjadi rumit).
  2. Sebagai gantinya, saya memilih pendekatan *Shared Database, Shared Schema with Column Discriminator (`institusi_id`)* yang diperkuat oleh barikade *Tenant Middleware* di sisi backend dan Row-Level isolation checks.
  3. Saya juga menolak penggunaan ORM berat dan memilih driver native berkecepatan tinggi `pgx/v5` dengan raw query yang teruji, aman dari SQL injection via parameterized queries, serta jauh lebih efisien dalam concurrency tinggi.

## Bagian yang saya kerjakan tanpa AI, dan alasannya
- **Perumusan Aturan Bisnis & Relasi Kelayakan KRS**: Menentukan korelasi antara status pelunasan tagihan UKT dengan hak akses mahasiswa ke pengisian Kartu Rencana Studi (KRS). Aturan ini memerlukan pemahaman konteks spesifik dari alur akademik kampus di Indonesia yang belum tentu dipahami secara utuh oleh AI tanpa arahan domain ahli.
- **Pemilihan Skema Idempotensi Webhook**: Menentukan secara mandiri bahwa idempotensi webhook harus berbasis pada transaksi unik dan status guard (`unpaid` -> `paid`) dengan pengembalian response no-op duplicate yang ramah bagi retry gateway, bukan sekadar menggunakan cache memori Redis sementara yang rentan hilang jika instance server restart.
- **Penetapan Hak Akses Role (RBAC Matrix)**: Menentukan pembagian wewenang yang tegas antara Super Admin (platform-level), Admin Keuangan (institution-scoped), dan Mahasiswa (individual student-scoped).

## Kalau mengulang dari awal, apa yang akan saya ubah
- **Penerapan Row-Level Security (RLS) PostgreSQL**: Menerapkan PostgreSQL RLS secara langsung di layer database engine sebagai lapisan pertahanan kedua (*defense-in-depth*), sehingga jika ada query yang lupa menyertakan klausul `WHERE institusi_id = ?`, database server secara otomatis tetap memblokir kebocoran data.
- **Event-Driven Architecture dengan Message Queue**: Menggunakan message broker seperti RabbitMQ atau Apache Kafka untuk menampung webhook masuk sebelum diproses ke database, sehingga mampu meredam lonjakan trafik masif (*traffic burst*) pada hari-hari terakhir periode pembayaran UKT nasional.
- **Contract Testing (OpenAPI / Swagger First)**: Mendefinisikan spesifikasi API OpenAPI 3.0 terlebih dahulu sebelum menulis kode, sehingga kontrak antara backend Go dan frontend React dapat divalidasi dan di-generate otomatis sedari awal.
