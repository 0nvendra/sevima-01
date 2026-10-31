package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sevima-be/internal/domain/entity"
	domainrepo "sevima-be/internal/domain/repository"
)

type PostgresRepositories struct {
	pool *pgxpool.Pool
}

func NewPostgresRepositories(pool *pgxpool.Pool) *PostgresRepositories {
	return &PostgresRepositories{pool: pool}
}

// ---------------- Institution Repository ----------------

func (r *PostgresRepositories) FindAll(ctx context.Context) ([]entity.Institution, error) {
	query := `SELECT id, code, name, address, phone, email, is_active, created_at, updated_at 
              FROM institutions ORDER BY name ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entity.Institution
	for rows.Next() {
		var i entity.Institution
		if err := rows.Scan(&i.ID, &i.Code, &i.Name, &i.Address, &i.Phone, &i.Email, &i.IsActive, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, i)
	}
	return list, nil
}

func (r *PostgresRepositories) FindByID(ctx context.Context, id uuid.UUID) (*entity.Institution, error) {
	query := `SELECT id, code, name, address, phone, email, is_active, hmac_secret, created_at, updated_at 
              FROM institutions WHERE id = $1`
	var i entity.Institution
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&i.ID, &i.Code, &i.Name, &i.Address, &i.Phone, &i.Email, &i.IsActive, &i.HMACSecret, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *PostgresRepositories) FindByCode(ctx context.Context, code string) (*entity.Institution, error) {
	query := `SELECT id, code, name, address, phone, email, is_active, hmac_secret, created_at, updated_at 
              FROM institutions WHERE code = $1 AND is_active = true`
	var i entity.Institution
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&i.ID, &i.Code, &i.Name, &i.Address, &i.Phone, &i.Email, &i.IsActive, &i.HMACSecret, &i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &i, nil
}

func (r *PostgresRepositories) Create(ctx context.Context, inst *entity.Institution) error {
	query := `INSERT INTO institutions (id, code, name, address, phone, email, is_active, hmac_secret, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())`
	if inst.ID == uuid.Nil {
		inst.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query, inst.ID, inst.Code, inst.Name, inst.Address, inst.Phone, inst.Email, inst.IsActive, inst.HMACSecret)
	return err
}

func (r *PostgresRepositories) Update(ctx context.Context, inst *entity.Institution) error {
	query := `UPDATE institutions SET name = $1, address = $2, phone = $3, email = $4, is_active = $5, updated_at = NOW() 
              WHERE id = $6`
	_, err := r.pool.Exec(ctx, query, inst.Name, inst.Address, inst.Phone, inst.Email, inst.IsActive, inst.ID)
	return err
}

// ---------------- User Repository ----------------

func (r *PostgresRepositories) FindUserByID(ctx context.Context, id uuid.UUID) (*entity.User, error) {
	query := `SELECT id, institusi_id, email, password_hash, full_name, role, is_active, last_login_at, created_at, updated_at 
              FROM users WHERE id = $1`
	var u entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.InstitusiID, &u.Email, &u.Password, &u.FullName, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepositories) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `SELECT id, institusi_id, email, password_hash, full_name, role, is_active, last_login_at, created_at, updated_at 
              FROM users WHERE email = $1`
	var u entity.User
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID, &u.InstitusiID, &u.Email, &u.Password, &u.FullName, &u.Role, &u.IsActive, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *PostgresRepositories) CreateUser(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (id, institusi_id, email, password_hash, full_name, role, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query, user.ID, user.InstitusiID, user.Email, user.Password, user.FullName, user.Role, user.IsActive)
	return err
}

func (r *PostgresRepositories) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// ---------------- Refresh Token Repository ----------------

func (r *PostgresRepositories) Store(ctx context.Context, userID uuid.UUID, token string) error {
	hash := sha256.Sum256([]byte(token))
	hashStr := hex.EncodeToString(hash[:])
	query := `INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at) 
              VALUES ($1, $2, $3, NOW() + INTERVAL '7 days', NOW())`
	_, err := r.pool.Exec(ctx, query, uuid.New(), userID, hashStr)
	return err
}

func (r *PostgresRepositories) ValidateAndRevoke(ctx context.Context, token string) (*uuid.UUID, error) {
	hash := sha256.Sum256([]byte(token))
	hashStr := hex.EncodeToString(hash[:])

	var userID uuid.UUID
	var revoked bool
	var expiresAt time.Time

	query := `SELECT user_id, revoked, expires_at FROM refresh_tokens WHERE token_hash = $1`
	err := r.pool.QueryRow(ctx, query, hashStr).Scan(&userID, &revoked, &expiresAt)
	if err != nil {
		return nil, errors.New("token not found")
	}

	if revoked || time.Now().After(expiresAt) {
		return nil, errors.New("token revoked or expired")
	}

	// Revoke old token
	_, err = r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE token_hash = $1`, hashStr)
	if err != nil {
		return nil, err
	}

	return &userID, nil
}

func (r *PostgresRepositories) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`, userID)
	return err
}

// ---------------- Student Repository ----------------

func (r *PostgresRepositories) FindStudentByID(ctx context.Context, id uuid.UUID) (*entity.Student, error) {
	query := `SELECT s.id, s.user_id, s.institusi_id, s.nim, s.study_program, s.enrollment_year, s.is_active, s.created_at, s.updated_at,
                     u.full_name, u.email
              FROM students s
              JOIN users u ON s.user_id = u.id
              WHERE s.id = $1`
	var s entity.Student
	var u entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.UserID, &s.InstitusiID, &s.NIM, &s.StudyProgram, &s.EnrollmentYear, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		&u.FullName, &u.Email,
	)
	if err != nil {
		return nil, err
	}
	u.ID = s.UserID
	s.User = &u
	return &s, nil
}

func (r *PostgresRepositories) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Student, error) {
	query := `SELECT s.id, s.user_id, s.institusi_id, s.nim, s.study_program, s.enrollment_year, s.is_active, s.created_at, s.updated_at,
                     u.full_name, u.email
              FROM students s
              JOIN users u ON s.user_id = u.id
              WHERE s.user_id = $1`
	var s entity.Student
	var u entity.User
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&s.ID, &s.UserID, &s.InstitusiID, &s.NIM, &s.StudyProgram, &s.EnrollmentYear, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		&u.FullName, &u.Email,
	)
	if err != nil {
		return nil, err
	}
	u.ID = s.UserID
	s.User = &u
	return &s, nil
}

func (r *PostgresRepositories) FindByNIM(ctx context.Context, institusiID uuid.UUID, nim string) (*entity.Student, error) {
	query := `SELECT s.id, s.user_id, s.institusi_id, s.nim, s.study_program, s.enrollment_year, s.is_active, s.created_at, s.updated_at,
                     u.full_name, u.email
              FROM students s
              JOIN users u ON s.user_id = u.id
              WHERE s.institusi_id = $1 AND s.nim = $2`
	var s entity.Student
	var u entity.User
	err := r.pool.QueryRow(ctx, query, institusiID, nim).Scan(
		&s.ID, &s.UserID, &s.InstitusiID, &s.NIM, &s.StudyProgram, &s.EnrollmentYear, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		&u.FullName, &u.Email,
	)
	if err != nil {
		return nil, err
	}
	u.ID = s.UserID
	s.User = &u
	return &s, nil
}

func (r *PostgresRepositories) ListByInstitution(ctx context.Context, institusiID uuid.UUID) ([]entity.Student, error) {
	query := `SELECT s.id, s.user_id, s.institusi_id, s.nim, s.study_program, s.enrollment_year, s.is_active, s.created_at, s.updated_at,
                     u.full_name, u.email
              FROM students s
              JOIN users u ON s.user_id = u.id
              WHERE s.institusi_id = $1 ORDER BY s.nim ASC`
	rows, err := r.pool.Query(ctx, query, institusiID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entity.Student
	for rows.Next() {
		var s entity.Student
		var u entity.User
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.InstitusiID, &s.NIM, &s.StudyProgram, &s.EnrollmentYear, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
			&u.FullName, &u.Email,
		); err != nil {
			return nil, err
		}
		u.ID = s.UserID
		s.User = &u
		list = append(list, s)
	}
	return list, nil
}

func (r *PostgresRepositories) CreateStudent(ctx context.Context, student *entity.Student) error {
	query := `INSERT INTO students (id, user_id, institusi_id, nim, study_program, enrollment_year, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())`
	if student.ID == uuid.Nil {
		student.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query, student.ID, student.UserID, student.InstitusiID, student.NIM, student.StudyProgram, student.EnrollmentYear, student.IsActive)
	return err
}

// ---------------- Invoice Repository ----------------

func (r *PostgresRepositories) FindInvoiceByID(ctx context.Context, id uuid.UUID) (*entity.Invoice, error) {
	query := `SELECT i.id, i.institusi_id, i.student_id, i.invoice_number, i.amount, i.status, i.semester, 
                     i.academic_year, i.due_date, i.paid_at, i.paid_amount, i.payment_method, i.notes, 
                     i.created_at, i.updated_at, s.nim, u.full_name, s.study_program
              FROM invoices i
              JOIN students s ON i.student_id = s.id
              JOIN users u ON s.user_id = u.id
              WHERE i.id = $1`
	var inv entity.Invoice
	var s entity.Student
	var u entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&inv.ID, &inv.InstitusiID, &inv.StudentID, &inv.InvoiceNumber, &inv.Amount, &inv.Status, &inv.Semester,
		&inv.AcademicYear, &inv.DueDate, &inv.PaidAt, &inv.PaidAmount, &inv.PaymentMethod, &inv.Notes,
		&inv.CreatedAt, &inv.UpdatedAt, &s.NIM, &u.FullName, &s.StudyProgram,
	)
	if err != nil {
		return nil, err
	}
	s.ID = inv.StudentID
	s.User = &u
	inv.Student = &s
	return &inv, nil
}

// FindByNumberForUpdate acquires an exclusive row lock (SELECT ... FOR UPDATE) inside a transaction.
func (r *PostgresRepositories) FindByNumberForUpdateTx(ctx context.Context, tx pgx.Tx, institusiID uuid.UUID, invoiceNumber string) (*entity.Invoice, error) {
	query := `SELECT id, institusi_id, student_id, invoice_number, amount, status, semester, 
                     academic_year, due_date, paid_at, paid_amount, payment_method, notes, created_at, updated_at
              FROM invoices 
              WHERE institusi_id = $1 AND invoice_number = $2 
              FOR UPDATE`
	var inv entity.Invoice
	err := tx.QueryRow(ctx, query, institusiID, invoiceNumber).Scan(
		&inv.ID, &inv.InstitusiID, &inv.StudentID, &inv.InvoiceNumber, &inv.Amount, &inv.Status, &inv.Semester,
		&inv.AcademicYear, &inv.DueDate, &inv.PaidAt, &inv.PaidAmount, &inv.PaymentMethod, &inv.Notes,
		&inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *PostgresRepositories) List(ctx context.Context, filter domainrepo.InvoiceFilter) ([]entity.Invoice, int64, error) {
	baseQuery := `FROM invoices i
                  JOIN students s ON i.student_id = s.id
                  JOIN users u ON s.user_id = u.id
                  WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if filter.InstitusiID != nil {
		baseQuery += fmt.Sprintf(" AND i.institusi_id = $%d", argIdx)
		args = append(args, *filter.InstitusiID)
		argIdx++
	}
	if filter.StudentID != nil {
		baseQuery += fmt.Sprintf(" AND i.student_id = $%d", argIdx)
		args = append(args, *filter.StudentID)
		argIdx++
	}
	if filter.Status != nil {
		baseQuery += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, *filter.Status)
		argIdx++
	}
	if filter.Semester != nil {
		baseQuery += fmt.Sprintf(" AND i.semester = $%d", argIdx)
		args = append(args, *filter.Semester)
		argIdx++
	}
	if filter.AcademicYear != nil && *filter.AcademicYear != "" {
		baseQuery += fmt.Sprintf(" AND i.academic_year = $%d", argIdx)
		args = append(args, *filter.AcademicYear)
		argIdx++
	}
	if filter.Search != nil && *filter.Search != "" {
		baseQuery += fmt.Sprintf(" AND (s.nim ILIKE $%d OR u.full_name ILIKE $%d OR i.invoice_number ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+*filter.Search+"%")
		argIdx++
	}

	// Count total
	countQuery := "SELECT COUNT(*) " + baseQuery
	var totalRecords int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, 0, err
	}

	// Pagination
	if filter.PerPage <= 0 {
		filter.PerPage = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PerPage

	selectQuery := fmt.Sprintf(`SELECT i.id, i.institusi_id, i.student_id, i.invoice_number, i.amount, i.status, i.semester, 
                                       i.academic_year, i.due_date, i.paid_at, i.paid_amount, i.payment_method, i.notes, 
                                       i.created_at, i.updated_at, s.nim, u.full_name, s.study_program
                                %s ORDER BY i.created_at DESC LIMIT %d OFFSET %d`, baseQuery, filter.PerPage, offset)

	rows, err := r.pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []entity.Invoice
	for rows.Next() {
		var inv entity.Invoice
		var s entity.Student
		var u entity.User
		if err := rows.Scan(
			&inv.ID, &inv.InstitusiID, &inv.StudentID, &inv.InvoiceNumber, &inv.Amount, &inv.Status, &inv.Semester,
			&inv.AcademicYear, &inv.DueDate, &inv.PaidAt, &inv.PaidAmount, &inv.PaymentMethod, &inv.Notes,
			&inv.CreatedAt, &inv.UpdatedAt, &s.NIM, &u.FullName, &s.StudyProgram,
		); err != nil {
			return nil, 0, err
		}
		s.ID = inv.StudentID
		s.User = &u
		inv.Student = &s
		list = append(list, inv)
	}

	return list, totalRecords, nil
}

func (r *PostgresRepositories) CreateInvoice(ctx context.Context, inv *entity.Invoice) error {
	query := `INSERT INTO invoices (id, institusi_id, student_id, invoice_number, amount, status, semester, academic_year, due_date, notes, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())`
	if inv.ID == uuid.Nil {
		inv.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query, inv.ID, inv.InstitusiID, inv.StudentID, inv.InvoiceNumber, inv.Amount, inv.Status, inv.Semester, inv.AcademicYear, inv.DueDate, inv.Notes)
	return err
}

func (r *PostgresRepositories) UpdateInvoice(ctx context.Context, inv *entity.Invoice) error {
	query := `UPDATE invoices SET status = $1, paid_at = $2, paid_amount = $3, payment_method = $4, notes = $5, updated_at = NOW() 
              WHERE id = $6`
	_, err := r.pool.Exec(ctx, query, inv.Status, inv.PaidAt, inv.PaidAmount, inv.PaymentMethod, inv.Notes, inv.ID)
	return err
}

func (r *PostgresRepositories) UpdateInvoiceTx(ctx context.Context, tx pgx.Tx, inv *entity.Invoice) error {
	query := `UPDATE invoices SET status = $1, paid_at = $2, paid_amount = $3, payment_method = $4, notes = $5, updated_at = NOW() 
              WHERE id = $6`
	_, err := tx.Exec(ctx, query, inv.Status, inv.PaidAt, inv.PaidAmount, inv.PaymentMethod, inv.Notes, inv.ID)
	return err
}

func (r *PostgresRepositories) CountUnpaidForStudent(ctx context.Context, studentID uuid.UUID, semester entity.SemesterType, academicYear string) (int, error) {
	query := `SELECT COUNT(*) FROM invoices 
              WHERE student_id = $1 AND semester = $2 AND academic_year = $3 AND status != 'paid'`
	var count int
	err := r.pool.QueryRow(ctx, query, studentID, semester, academicYear).Scan(&count)
	return count, err
}

func (r *PostgresRepositories) ExpireOverdueInvoices(ctx context.Context) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT expire_overdue_invoices()`).Scan(&count)
	return count, err
}

// ---------------- Transaction & Webhook Log Repositories ----------------

func (r *PostgresRepositories) CreateTransaction(ctx context.Context, txn *entity.TransactionHistory) error {
	query := `INSERT INTO transaction_histories (id, invoice_id, institusi_id, previous_status, new_status, source, notes, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	if txn.ID == uuid.Nil {
		txn.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query, txn.ID, txn.InvoiceID, txn.InstitusiID, txn.PreviousStatus, txn.NewStatus, txn.Source, txn.Notes)
	return err
}

func (r *PostgresRepositories) CreateTransactionTx(ctx context.Context, tx pgx.Tx, txn *entity.TransactionHistory) error {
	query := `INSERT INTO transaction_histories (id, invoice_id, institusi_id, previous_status, new_status, source, notes, created_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())`
	if txn.ID == uuid.Nil {
		txn.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, query, txn.ID, txn.InvoiceID, txn.InstitusiID, txn.PreviousStatus, txn.NewStatus, txn.Source, txn.Notes)
	return err
}

func (r *PostgresRepositories) ListByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]entity.TransactionHistory, error) {
	query := `SELECT id, invoice_id, institusi_id, previous_status, new_status, source, notes, created_at 
              FROM transaction_histories WHERE invoice_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entity.TransactionHistory
	for rows.Next() {
		var t entity.TransactionHistory
		if err := rows.Scan(&t.ID, &t.InvoiceID, &t.InstitusiID, &t.PreviousStatus, &t.NewStatus, &t.Source, &t.Notes, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (r *PostgresRepositories) CreateWebhookLog(ctx context.Context, log *entity.WebhookLog) error {
	query := `INSERT INTO webhook_logs (id, institusi_id, invoice_id, payload_raw, signature_recv, signature_valid, status, error_message, processed_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())`
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	_, err := r.pool.Exec(ctx, query, log.ID, log.InstitusiID, log.InvoiceID, log.PayloadRaw, log.SignatureRecv, log.SignatureValid, log.Status, log.ErrorMessage)
	return err
}

func (r *PostgresRepositories) Pool() *pgxpool.Pool {
	return r.pool
}
