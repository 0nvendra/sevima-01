# Technical Decisions Record (ADR) — SEVIMA-PAY

Dokumen ini mencatat 4 keputusan teknis penting yang diambil selama perancangan dan implementasi platform SEVIMA-PAY, beserta konteks, alternatif opsi yang dipertimbangkan, keputusan yang diambil, dan konsekuensinya.

---

## 1. Arsitektur Multi-Tenancy: Shared Database with Discriminator Column vs Database-per-Tenant

Konteks :
SEVIMA-PAY melayani ratusan institusi perguruan tinggi di Indonesia. Sistem harus menjamin bahwa data mahasiswa, tagihan, dan transaksi antar-institusi terisolasi secara mutlak (tidak ada kebocoran data lintas kampus) dengan biaya infrastruktur yang terukur dan skalabilitas operasional yang mudah dikelola.

Opsi :
1. **Database-per-Tenant**: Setiap institusi memiliki database PostgreSQL terpisah secara fisik.
2. **Schema-per-Tenant**: Satu database PostgreSQL bersama, namun setiap institusi memiliki skema terisolasi (`tenant_ugm`, `tenant_its`).
3. **Shared Database, Shared Schema with Discriminator Column (`institusi_id`)**: Seluruh tabel berada dalam skema yang sama dengan kolom penanda `institusi_id` pada setiap entitas, dilindungi oleh Tenant Middleware dan index komposit.

Keputusan :
Memilih **Opsi 3: Shared Database, Shared Schema with Discriminator Column (`institusi_id`)**.

Konsekuensi :
- **Positif**: Efisiensi biaya infrastruktur tinggi, proses migrasi database skema hanya perlu dijalankan satu kali untuk seluruh kampus, connection pooling database lebih optimal, dan memudahkan Super Admin dalam melakukan pelaporan analitik global lintas institusi.
- **Tantangan / Mitigasi**: Risiko kebocoran data jika developer melewatkan filter tenant pada query. Dimitigasi dengan mewajibkan `TenantMiddleware` pada semua rute terlindungi, validasi kepemilikan tenant di layer Use Case, serta unit testing ketat terhadap isolasi data antar-institusi.

---

## 2. Strategi Verifikasi Webhook & Penanganan Idempotensi

Konteks :
Notifikasi pelunasan dari Payment Gateway pihak ketiga dikirimkan melalui HTTP Webhook secara asinkron. Webhook rentan terhadap serangan peniruan payload (tampering/spoofing) dan pengiriman berulang (at-least-once delivery / retry mechanism) yang berisiko memicu pencatatan pelunasan ganda atau inkonsistensi status tagihan.

Opsi :
1. **Verifikasi Token Statis (Bearer Token / API Key)** dan deduplikasi menggunakan Redis in-memory cache dengan TTL.
2. **Verifikasi HMAC-SHA256 Secret Per Institusi** dengan idempotensi berbasis State Machine tagihan (`unpaid` -> `paid`) dan audit log unik pada database PostgreSQL.
3. **Polling Transaksi Berkala (Cron Pull)** tanpa endpoint webhook.

Keputusan :
Memilih **Opsi 2: Verifikasi HMAC-SHA256 Secret Per Institusi dengan State Guard & Audit Log**.

Konsekuensi :
- **Positif**: Integritas data terjamin tinggi karena signature dihitung dari raw body payload menggunakan secret key unik per kampus (kebocoran secret satu kampus tidak berdampak pada kampus lain). Penanganan idempotensi yang melekat pada status tagihan dan transaksi database menjamin bahwa request duplikat tetap mengembalikan status sukses (HTTP 200) tanpa mengeksekusi mutasi ganda.
- **Tantangan / Mitigasi**: Perlu menyimpan secret key institusi dengan aman (kolom `hmac_secret`) dan memastikan raw request body tidak ter-consume sebelum verifikasi signature dilakukan di backend Gin.

---

## 3. Mekanisme Concurrency Control: PostgreSQL Row Locking vs Optimistic Locking

Konteks :
Pada puncak masa pembayaran UKT (misal hari terakhir registrasi semester), ribuan webhook dan request pengecekan status dapat terjadi secara simultan. Jika mahasiswa melakukan pembayaran hampir bersamaan atau gateway mengirim retry bersamaan, terdapat potensi race condition saat memperbarui status tagihan dan mencatat mutasi keuangan.

Opsi :
1. **Optimistic Locking**: Menggunakan kolom `version` (integer) pada tabel invoices (`UPDATE ... WHERE version = x`). Jika versi berubah, transaksi dibatalkan dan klien diminta retry.
2. **Pessimistic Locking (`SELECT FOR UPDATE`)**: Mengunci baris tagihan pada transaksi PostgreSQL hingga proses mutasi status dan pembuatan audit log selesai.
3. **No Locking (Last-Write-Wins)**: Melakukan update langsung tanpa mekanisme locking.

Keputusan :
Memilih **Opsi 2: Pessimistic Locking (`SELECT FOR UPDATE`) di dalam Database Transaction**.

Konsekuensi :
- **Positif**: Menjamin konsistensi serializable secara instan untuk baris tagihan yang sedang diproses. Request kedua yang tiba bersamaan akan antre secara atomik hingga transaksi pertama selesai dan langsung mendeteksi status baru (`paid`), sehingga langsung dialihkan ke cabang idempotensi aman tanpa kegagalan acak bagi gateway.
- **Tantangan / Mitigasi**: Lock baris dapat menimbulkan sedikit antrian latensi jika transaksi berjalan lama. Dimitigasi dengan menjaga ruang lingkup transaksi database sesingkat mungkin (hanya update status invoice dan insert transaksi dalam satu roundtrip cepat).

---

## 4. Pola Arsitektur Backend: Clean Architecture vs Flat MVC

Konteks :
Aplikasi mengelola logika finansial penting dengan kebutuhan audit trail, aturan multi-institusi yang ketat, dan integrasi eksternal. Struktur kode harus mudah diuji secara independen, tidak terikat mati pada web framework atau driver basis data, dan terstruktur rapi untuk kolaborasi tim.

Opsi :
1. **Flat MVC**: Menempatkan seluruh logika di dalam controller dan model aktif.
2. **Clean Architecture (Domain, Use Case, Adapter/Repository, Infrastructure)**: Memisahkan aturan bisnis murni ke dalam Domain & Use Case, sedangkan detail HTTP (Gin) dan database (pgx) berada di lapisan luar Adapter.
3. **Microservices Architecture**: Memisahkan modul auth, billing, dan webhook ke dalam service-service terpisah.

Keputusan :
Memilih **Opsi 2: Clean Architecture (Hexagonal / Onion Style)**.

Konsekuensi :
- **Positif**: Aturan bisnis isolasi tenant, validasi status, dan perhitungan kelayakan KRS dapat diuji melalui unit test tanpa memerlukan koneksi database riil. Setiap layer memiliki tanggung jawab tunggal (*Single Responsibility Principle*).
- **Tantangan / Mitigasi**: Memerlukan boilerplate kode awal (interface repository, DTO mapping, dependency injection manual di `cmd/api/main.go`). Hal ini sebanding dengan kemudahan pemeliharaan jangka panjang dan keandalan sistem.
