package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"sevima-be/internal/pkg/crypto"
	"sevima-be/internal/pkg/database"
)

func main() {
	ctx := context.Background()

	cfg := database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "postgres",
		Password: "",
		DBName:   "sevima_pay",
		SSLMode:  "disable",
	}

	pool, err := database.NewPostgresPool(ctx, cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	fmt.Println("🚀 Seeding Database 'sevima_pay'...")

	passwordHash, _ := crypto.HashPassword("Password123!")

	// 1. Seed Institutions
	ugmID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	itsID := uuid.MustParse("22222222-2222-2222-2222-222222222222")

	_, _ = pool.Exec(ctx, `
		INSERT INTO institutions (id, code, name, address, phone, email, is_active, hmac_secret)
		VALUES 
			($1, 'UGM', 'Universitas Gadjah Mada', 'Bulaksumur, Yogyakarta', '0274-588688', 'info@ugm.ac.id', true, 'secret_ugm_1234567890_abcdef'),
			($2, 'ITS', 'Institut Teknologi Sepuluh Nopember', 'Sukolilo, Surabaya', '031-5994251', 'humas@its.ac.id', true, 'secret_its_1234567890_abcdef')
		ON CONFLICT (code) DO NOTHING;
	`, ugmID, itsID)

	// 2. Seed Users
	superAdminID := uuid.MustParse("33333333-3333-3333-3333-333333333333")
	adminUgmID := uuid.MustParse("44444444-4444-4444-4444-444444444444")
	adminItsID := uuid.MustParse("55555555-5555-5555-5555-555555555555")
	mhsUgmUserID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	mhsItsUserID := uuid.MustParse("77777777-7777-7777-7777-777777777777")

	_, _ = pool.Exec(ctx, `
		INSERT INTO users (id, institusi_id, email, password_hash, full_name, role, is_active)
		VALUES 
			($1, NULL, 'superadmin@sevima.com', $6, 'Super Administrator SEVIMA', 'super_admin', true),
			($2, $7,   'admin@ugm.ac.id',       $6, 'Budi Santoso (Admin UGM)', 'admin_keuangan', true),
			($3, $8,   'admin@its.ac.id',       $6, 'Dewi Lestari (Admin ITS)', 'admin_keuangan', true),
			($4, $7,   'fauzi@mail.ugm.ac.id',  $6, 'Ahmad Fauzi (Mahasiswa UGM)', 'mahasiswa', true),
			($5, $8,   'bima@mhs.its.ac.id',    $6, 'Bima Sakti (Mahasiswa ITS)', 'mahasiswa', true)
		ON CONFLICT (email) DO NOTHING;
	`, superAdminID, adminUgmID, adminItsID, mhsUgmUserID, mhsItsUserID, passwordHash, ugmID, itsID)

	// 3. Seed Students
	studentUgmID := uuid.MustParse("88888888-8888-8888-8888-888888888888")
	studentItsID := uuid.MustParse("99999999-9999-9999-9999-999999999999")

	_, _ = pool.Exec(ctx, `
		INSERT INTO students (id, user_id, institusi_id, nim, study_program, enrollment_year, is_active)
		VALUES
			($1, $2, $3, '21/475123/TK/52111', 'Teknik Informatika', 2021, true),
			($4, $5, $6, '5025211001', 'Sistem Informasi', 2021, true)
		ON CONFLICT (institusi_id, nim) DO NOTHING;
	`, studentUgmID, mhsUgmUserID, ugmID, studentItsID, mhsItsUserID, itsID)

	// 4. Seed Invoices
	invUgmID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	invItsID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	dueDate := time.Now().Add(30 * 24 * time.Hour)

	_, _ = pool.Exec(ctx, `
		INSERT INTO invoices (id, institusi_id, student_id, invoice_number, amount, status, semester, academic_year, due_date, notes)
		VALUES
			($1, $2, $3, 'INV-2026-UGM-0001', 7500000, 'unpaid', 'ganjil', '2026/2027', $7, 'Tagihan UKT Semester Ganjil 2026/2027'),
			($4, $5, $6, 'INV-2026-ITS-0001', 6000000, 'unpaid', 'ganjil', '2026/2027', $7, 'Tagihan UKT Semester Ganjil 2026/2027')
		ON CONFLICT (institusi_id, invoice_number) DO NOTHING;
	`, invUgmID, ugmID, studentUgmID, invItsID, itsID, studentItsID, dueDate)

	// 5. Initial Transaction History
	_, _ = pool.Exec(ctx, `
		INSERT INTO transaction_histories (id, invoice_id, institusi_id, previous_status, new_status, source, notes)
		VALUES
			(gen_random_uuid(), $1, $2, 'unpaid', 'unpaid', 'invoice_created', 'Tagihan diterbitkan sistem'),
			(gen_random_uuid(), $3, $4, 'unpaid', 'unpaid', 'invoice_created', 'Tagihan diterbitkan sistem')
		ON CONFLICT DO NOTHING;
	`, invUgmID, ugmID, invItsID, itsID)

	fmt.Println("✅ Database Seeded Successfully!")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("Role              | Email                   | Password     | Tenant")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("Super Admin       | superadmin@sevima.com   | Password123! | Global")
	fmt.Println("Admin Keuangan    | admin@ugm.ac.id         | Password123! | UGM")
	fmt.Println("Admin Keuangan    | admin@its.ac.id         | Password123! | ITS")
	fmt.Println("Mahasiswa         | fauzi@mail.ugm.ac.id    | Password123! | UGM (NIM: 21/475123/TK/52111)")
	fmt.Println("Mahasiswa         | bima@mhs.its.ac.id      | Password123! | ITS (NIM: 5025211001)")
	fmt.Println("------------------------------------------------------------------------")
	fmt.Println("Sample Invoices:")
	fmt.Println("1. UGM: INV-2026-UGM-0001 | Rp 7.500.000 | Status: unpaid")
	fmt.Println("2. ITS: INV-2026-ITS-0001 | Rp 6.000.000 | Status: unpaid")
	fmt.Println("------------------------------------------------------------------------")
}
