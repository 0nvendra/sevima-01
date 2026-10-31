package repository

import (
	"context"

	"github.com/google/uuid"
	"sevima-be/internal/domain/entity"
)

type InstitutionRepository interface {
	FindAll(ctx context.Context) ([]entity.Institution, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Institution, error)
	FindByCode(ctx context.Context, code string) (*entity.Institution, error)
	Create(ctx context.Context, inst *entity.Institution) error
	Update(ctx context.Context, inst *entity.Institution) error
}

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(ctx context.Context, user *entity.User) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	Revoked   bool
}

type RefreshTokenRepository interface {
	Store(ctx context.Context, userID uuid.UUID, tokenHash string) error
	ValidateAndRevoke(ctx context.Context, tokenHash string) (*uuid.UUID, error)
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

type StudentRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Student, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.Student, error)
	FindByNIM(ctx context.Context, institusiID uuid.UUID, nim string) (*entity.Student, error)
	ListByInstitution(ctx context.Context, institusiID uuid.UUID) ([]entity.Student, error)
	Create(ctx context.Context, student *entity.Student) error
}

type InvoiceFilter struct {
	InstitusiID  *uuid.UUID
	StudentID    *uuid.UUID
	Status       *entity.InvoiceStatus
	Semester     *entity.SemesterType
	AcademicYear *string
	Search       *string
	Page         int
	PerPage      int
}

type InvoiceRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Invoice, error)
	FindByNumberForUpdate(ctx context.Context, institusiID uuid.UUID, invoiceNumber string) (*entity.Invoice, error)
	List(ctx context.Context, filter InvoiceFilter) ([]entity.Invoice, int64, error)
	Create(ctx context.Context, inv *entity.Invoice) error
	Update(ctx context.Context, inv *entity.Invoice) error
	CountUnpaidForStudent(ctx context.Context, studentID uuid.UUID, semester entity.SemesterType, academicYear string) (int, error)
	ExpireOverdueInvoices(ctx context.Context) (int, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, txn *entity.TransactionHistory) error
	ListByInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]entity.TransactionHistory, error)
}

type WebhookLogRepository interface {
	Create(ctx context.Context, log *entity.WebhookLog) error
}
