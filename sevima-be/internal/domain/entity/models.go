package entity

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleSuperAdmin    UserRole = "super_admin"
	RoleAdminKeuangan UserRole = "admin_keuangan"
	RoleMahasiswa     UserRole = "mahasiswa"
)

type InvoiceStatus string

const (
	InvoiceStatusUnpaid  InvoiceStatus = "unpaid"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusExpired InvoiceStatus = "expired"
)

type SemesterType string

const (
	SemesterGanjil SemesterType = "ganjil"
	SemesterGenap  SemesterType = "genap"
)

type Institution struct {
	ID         uuid.UUID `json:"id"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	Phone      string    `json:"phone"`
	Email      string    `json:"email"`
	IsActive   bool      `json:"is_active"`
	HMACSecret string    `json:"-"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
	ID          uuid.UUID  `json:"id"`
	InstitusiID *uuid.UUID `json:"institusi_id,omitempty"`
	Email       string     `json:"email"`
	Password    string     `json:"-"`
	FullName    string     `json:"full_name"`
	Role        UserRole   `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Student struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	InstitusiID    uuid.UUID `json:"institusi_id"`
	NIM            string    `json:"nim"`
	StudyProgram   string    `json:"study_program"`
	EnrollmentYear int       `json:"enrollment_year"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	User           *User     `json:"user,omitempty"`
}

type Invoice struct {
	ID            uuid.UUID     `json:"id"`
	InstitusiID   uuid.UUID     `json:"institusi_id"`
	StudentID     uuid.UUID     `json:"student_id"`
	InvoiceNumber string        `json:"invoice_number"`
	Amount        int64         `json:"amount"`
	Status        InvoiceStatus `json:"status"`
	Semester      SemesterType  `json:"semester"`
	AcademicYear  string        `json:"academic_year"`
	DueDate       time.Time     `json:"due_date"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	PaidAmount    *int64        `json:"paid_amount,omitempty"`
	PaymentMethod *string       `json:"payment_method,omitempty"`
	Notes         string        `json:"notes,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Student       *Student      `json:"student,omitempty"`
}

type TransactionHistory struct {
	ID             uuid.UUID     `json:"id"`
	InvoiceID      uuid.UUID     `json:"invoice_id"`
	InstitusiID    uuid.UUID     `json:"institusi_id"`
	PreviousStatus InvoiceStatus `json:"previous_status"`
	NewStatus      InvoiceStatus `json:"new_status"`
	Source         string        `json:"source"`
	Notes          string        `json:"notes,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
}

type WebhookLog struct {
	ID             uuid.UUID  `json:"id"`
	InstitusiID    *uuid.UUID `json:"institusi_id,omitempty"`
	InvoiceID      *uuid.UUID `json:"invoice_id,omitempty"`
	PayloadRaw     string     `json:"payload_raw"`
	SignatureRecv  string     `json:"signature_recv"`
	SignatureValid bool       `json:"signature_valid"`
	Status         string     `json:"status"`
	ErrorMessage   string     `json:"error_message,omitempty"`
	ProcessedAt    time.Time  `json:"processed_at"`
}
