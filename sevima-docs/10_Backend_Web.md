# Backend Web Specification
# SEVIMA-PAY: Go (Gin) Backend

> **Version**: 1.0.0
> **Language**: Go 1.22+
> **HTTP Framework**: Gin Web Framework (`github.com/gin-gonic/gin`)
> **Architecture**: Clean Architecture (Hexagonal / Ports & Adapters)
> **Data Access**: `sqlc` + `pgx/v5`
> **Database**: PostgreSQL 16+

---

## 1. Clean Architecture Structure

```
sevima-be/
├── cmd/
│   ├── api/
│   │   └── main.go                 # Entrypoint server REST API
│   ├── worker/
│   │   └── main.go                 # Entrypoint background cron worker (auto-expiry)
│   └── seed/
│       └── main.go                 # Database seeder (institusi, user, invoice)
├── internal/
│   ├── domain/                     # Layer 1: Enterprise Business Rules (Entities & Interfaces)
│   │   ├── entity/
│   │   │   ├── institution.go
│   │   │   ├── user.go
│   │   │   ├── student.go
│   │   │   ├── invoice.go
│   │   │   ├── transaction.go
│   │   │   └── audit_log.go
│   │   └── repository/             # Repository interfaces (Contracts)
│   │       ├── institution_repo.go
│   │       ├── user_repo.go
│   │       ├── student_repo.go
│   │       ├── invoice_repo.go
│   │       ├── transaction_repo.go
│   │       └── audit_log_repo.go
│   ├── usecase/                    # Layer 2: Application Business Rules (Use Cases)
│   │   ├── auth/
│   │   │   └── auth_usecase.go     # Login, Refresh, Logout, Profile
│   │   ├── institution/
│   │   │   └── institution_usecase.go # Super admin institusi management
│   │   ├── invoice/
│   │   │   ├── invoice_usecase.go  # Create, List, Detail, Cancel
│   │   │   └── expiry_worker.go    # Background job auto-expiry
│   │   ├── student/
│   │   │   └── student_usecase.go  # List student, detail student, import
│   │   ├── webhook/
│   │   │   └── webhook_usecase.go  # HMAC verification, Idempotent transition, Status update
│   │   └── mockgateway/
│   │       └── gateway_usecase.go  # Trigger mock payment, compute HMAC, forward webhook
│   ├── adapter/                    # Layer 3: Interface Adapters
│   │   ├── handler/                # HTTP Handlers (Gin)
│   │   │   ├── auth_handler.go
│   │   │   ├── institution_handler.go
│   │   │   ├── invoice_handler.go
│   │   │   ├── student_handler.go
│   │   │   ├── webhook_handler.go
│   │   │   └── mockgateway_handler.go
│   │   ├── middleware/             # HTTP Middlewares
│   │   │   ├── auth_middleware.go  # JWT verification & parsing
│   │   │   ├── rbac_middleware.go  # Role validation (super_admin, admin_keuangan, mahasiswa)
│   │   │   ├── tenant_middleware.go# Multi-tenant context injection & header validation
│   │   │   ├── logger_middleware.go# Structured JSON logging (Zap/Slog) with correlation ID
│   │   │   ├── recovery_middleware.go
│   │   │   └── cors_middleware.go
│   │   └── repository/             # Data Access Implementations (sqlc / pgx)
│   │       ├── pg_institution_repo.go
│   │       ├── pg_user_repo.go
│   │       ├── pg_student_repo.go
│   │       ├── pg_invoice_repo.go
│   │       ├── pg_transaction_repo.go
│   │       └── pg_audit_log_repo.go
│   └── pkg/                        # Cross-cutting packages
│       ├── crypto/
│       │   ├── hmac.go             # HMAC-SHA256 signature verification & generation
│       │   └── jwt.go              # Access token & Refresh token helper
│       ├── database/
│       │   ├── postgres.go         # pgxpool connection & health check
│       │   └── tx_manager.go       # Transaction runner abstraction
│       ├── logger/
│       │   └── zap.go              # Uber Zap logger configuration
│       └── response/
│           └── response.go         # Standardized JSON response envelope
├── db/
│   ├── migrations/                 # Golang-migrate / raw SQL files
│   │   ├── 000001_create_institutions_table.up.sql
│   │   ├── 000002_create_users_students_table.up.sql
│   │   ├── 000003_create_invoices_table.up.sql
│   │   └── 000004_create_audit_logs_table.up.sql
│   ├── queries/                    # sqlc input queries
│   │   ├── institutions.sql
│   │   ├── users.sql
│   │   ├── students.sql
│   │   ├── invoices.sql
│   │   └── audit_logs.sql
│   └── sqlc.yaml                   # sqlc code generation configuration
├── configs/
│   ├── config.go                   # Viper / Env config loader
│   └── config.yaml.example
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

---

## 2. Core Domain Entities & Interfaces

### 2.1 Domain Entities (`internal/domain/entity/`)

```go
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
	HMACSecret string    `json:"-"` // Never expose in JSON
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type User struct {
	ID            uuid.UUID  `json:"id"`
	InstitusiID   *uuid.UUID `json:"institusi_id,omitempty"` // nil for super_admin
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	FullName      string     `json:"full_name"`
	Role          UserRole   `json:"role"`
	IsActive      bool       `json:"is_active"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
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
	Amount        int64         `json:"amount"` // IDR in Integer
	Status        InvoiceStatus `json:"status"`
	Semester      SemesterType  `json:"semester"`
	AcademicYear  string        `json:"academic_year"` // e.g. "2026/2027"
	DueDate       time.Time     `json:"due_date"`
	PaidAt        *time.Time    `json:"paid_at,omitempty"`
	PaidAmount    *int64        `json:"paid_amount,omitempty"`
	PaymentMethod *string       `json:"payment_method,omitempty"`
	Notes         string        `json:"notes,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Student       *Student      `json:"student,omitempty"`
}

type WebhookLog struct {
	ID             uuid.UUID `json:"id"`
	InstitusiID    *uuid.UUID `json:"institusi_id,omitempty"`
	InvoiceID      *uuid.UUID `json:"invoice_id,omitempty"`
	PayloadRaw     string     `json:"payload_raw"`
	SignatureRecv  string     `json:"signature_recv"`
	SignatureValid bool       `json:"signature_valid"`
	Status         string     `json:"status"` // processed, duplicate, failed, invalid_signature
	ErrorMessage   string     `json:"error_message,omitempty"`
	ProcessedAt    time.Time  `json:"processed_at"`
}
```

---

## 3. Middleware Implementation

### 3.1 Tenant Context & Isolation (`internal/adapter/middleware/tenant_middleware.go`)

```go
package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	TenantIDContextKey contextKey = "tenant_id"
	UserIDContextKey   contextKey = "user_id"
	UserRoleContextKey contextKey = "user_role"
)

// TenantMiddleware extracts tenant_id from JWT claims or X-Institution-ID (for Super Admin)
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(string(UserRoleContextKey))
		
		// 1. If super_admin, they can specify X-Institution-ID to scope queries
		if role == "super_admin" {
			instHeader := c.GetHeader("X-Institution-ID")
			if instHeader != "" {
				if instUUID, err := uuid.Parse(instHeader); err == nil {
					c.Set(string(TenantIDContextKey), instUUID)
					ctx := context.WithValue(c.Request.Context(), TenantIDContextKey, instUUID)
					c.Request = c.Request.WithContext(ctx)
				}
			}
			c.Next()
			return
		}

		// 2. For admin_keuangan and mahasiswa, tenant_id MUST come from verified JWT claims
		claimTenant, exists := c.Get(string(TenantIDContextKey))
		if !exists || claimTenant == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: No institution bound to authenticated user",
			})
			return
		}

		instUUID, ok := claimTenant.(uuid.UUID)
		if !ok || instUUID == uuid.Nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden: Invalid institution ID",
			})
			return
		}

		// Inject into standard Go context for sqlc & use cases
		ctx := context.WithValue(c.Request.Context(), TenantIDContextKey, instUUID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
```

### 3.2 Role-Based Access Control (RBAC) (`internal/adapter/middleware/rbac_middleware.go`)

```go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"sevima-be/internal/domain/entity"
)

func RequireRole(roles ...entity.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(string(UserRoleContextKey))
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userRole := entity.UserRole(val.(string))
		allowed := false
		for _, r := range roles {
			if userRole == r {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Access denied: insufficient permissions for this resource",
			})
			return
		}

		c.Next()
	}
}
```

---

## 4. Webhook Engine & Signature Verification

### 4.1 HMAC-SHA256 Utility (`internal/pkg/crypto/hmac.go`)

```go
package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateHMAC computes HMAC-SHA256 of raw body using the given secret key
func GenerateHMAC(payload []byte, secretKey string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMAC compares received signature with expected HMAC using constant-time comparison
func VerifyHMAC(payload []byte, receivedSignature string, secretKey string) bool {
	expectedSignature := GenerateHMAC(payload, secretKey)
	return hmac.Equal([]byte(expectedSignature), []byte(receivedSignature))
}
```

### 4.2 Webhook Usecase & Idempotent Processing (`internal/usecase/webhook/webhook_usecase.go`)

```go
package webhook

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"sevima-be/internal/domain/entity"
	"sevima-be/internal/domain/repository"
	"sevima-be/internal/pkg/crypto"
)

type WebhookPayload struct {
	InstitutionCode string `json:"institution_code"`
	EventType       string `json:"event_type"` // payment.success, payment.expired
	InvoiceNumber   string `json:"invoice_number"`
	Amount          int64  `json:"amount"`
	PaymentMethod   string `json:"payment_method"`
	PaidAt          string `json:"paid_at"`
	TransactionID   string `json:"transaction_id"`
}

type WebhookUseCase struct {
	instRepo      repository.InstitutionRepository
	invoiceRepo   repository.InvoiceRepository
	webhookLogRepo repository.WebhookLogRepository
	txManager     repository.TransactionManager
}

func (u *WebhookUseCase) ProcessWebhook(ctx context.Context, rawBody []byte, signature string, payload *WebhookPayload) error {
	// 1. Look up institution by code
	inst, err := u.instRepo.FindByCode(ctx, payload.InstitutionCode)
	if err != nil {
		u.logFailure(ctx, nil, nil, rawBody, signature, false, "invalid_institution", "Institution code not recognized")
		return errors.New("institution not found")
	}

	// 2. Verify HMAC Signature
	isValid := crypto.VerifyHMAC(rawBody, signature, inst.HMACSecret)
	if !isValid {
		u.logFailure(ctx, &inst.ID, nil, rawBody, signature, false, "invalid_signature", "HMAC verification failed")
		return errors.New("invalid signature")
	}

	// 3. Begin Database Transaction for Atomic Invoice State Transition
	return u.txManager.WithTransaction(ctx, func(txCtx context.Context) error {
		// Acquire Row-Level Lock: SELECT * FROM invoices WHERE institusi_id = $1 AND invoice_number = $2 FOR UPDATE
		inv, err := u.invoiceRepo.FindByNumberForUpdate(txCtx, inst.ID, payload.InvoiceNumber)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				u.logFailure(txCtx, &inst.ID, nil, rawBody, signature, true, "invoice_not_found", "Invoice does not exist")
				return errors.New("invoice not found")
			}
			return err
		}

		// 4. State Machine Transition & Idempotency Evaluation
		switch payload.EventType {
		case "payment.success":
			if inv.Status == entity.InvoiceStatusPaid {
				// Idempotent: Already paid, return success without duplicate processing
				u.logSuccess(txCtx, &inst.ID, &inv.ID, rawBody, signature, "duplicate", "Already paid")
				return nil
			}
			if inv.Status == entity.InvoiceStatusExpired {
				u.logFailure(txCtx, &inst.ID, &inv.ID, rawBody, signature, true, "invalid_transition", "Cannot pay expired invoice")
				return errors.New("cannot pay expired invoice")
			}

			// Validate payment amount
			if payload.Amount != inv.Amount {
				u.logFailure(txCtx, &inst.ID, &inv.ID, rawBody, signature, true, "amount_mismatch", fmt.Sprintf("Expected %d, got %d", inv.Amount, payload.Amount))
				return errors.New("paid amount does not match invoice total")
			}

			paidAt, _ := time.Parse(time.RFC3339, payload.PaidAt)
			if paidAt.IsZero() {
				paidAt = time.Now()
			}

			// Update Invoice
			inv.Status = entity.InvoiceStatusPaid
			inv.PaidAt = &paidAt
			inv.PaidAmount = &payload.Amount
			inv.PaymentMethod = &payload.PaymentMethod
			if err := u.invoiceRepo.Update(txCtx, inv); err != nil {
				return err
			}

			// Record in Transaction History
			u.logSuccess(txCtx, &inst.ID, &inv.ID, rawBody, signature, "processed", "Payment successfully processed")

		case "payment.expired":
			if inv.Status == entity.InvoiceStatusPaid {
				// Cannot expire an already paid invoice
				return nil
			}
			if inv.Status == entity.InvoiceStatusExpired {
				return nil // Idempotent
			}

			inv.Status = entity.InvoiceStatusExpired
			if err := u.invoiceRepo.Update(txCtx, inv); err != nil {
				return err
			}
			u.logSuccess(txCtx, &inst.ID, &inv.ID, rawBody, signature, "processed", "Invoice marked expired")
		}

		return nil
	})
}
```

---

## 5. Mock Payment Gateway Service

Untuk memudahkan demo, pengujian lokal, dan live coding saat evaluasi teknis, backend menyediakan Mock Payment Gateway built-in:

1. **Endpoint Simulator**:
   - `POST /api/v1/mock-gateway/pay`: Menerima `invoice_number` & `institution_code`, lalu secara otomatis membuat payload webhook lengkap, menghitung HMAC-SHA256 signature menggunakan secret institusi terkait, dan menembakkannya ke endpoint webhook `/api/v1/webhook/payment`.
   - `POST /api/v1/mock-gateway/expire`: Mensimulasikan event invoice expired via webhook.
2. **Kesesuaian Pengujian**:
   - Memungkinkan mahasiswa atau admin mengklik tombol "Bayar via Simulator" pada UI.
   - Tidak memerlukan dependensi eksternal (ngrok atau third-party sandbox) untuk demo end-to-end.

---

## 6. Background Worker: Auto-Expiry Cron

- Dilengkapi background goroutine / ticker (`time.NewTicker(1 * time.Minute)`) untuk mencari semua invoice dengan status `unpaid` yang `due_date < NOW()`.
- Menggunakan query batch berukuran 100 dengan lock `FOR UPDATE SKIP LOCKED` untuk menghindari deadlocks atau race condition dengan worker concurrent lainnya.
- Memperbarui status menjadi `expired` dan mencatatnya ke `transaction_histories`.

---

## 7. Testing Strategy

1. **Unit Testing**:
   - `crypto/hmac_test.go`: Validasi integritas kalkulasi HMAC & ketahanan terhadap timing attack.
   - `usecase/webhook_test.go`: Mock repository untuk memvalidasi idempotensi (duplicate payload), transisi status tidak valid (`expired -> paid`), dan error handling.
2. **Integration Testing**:
   - Menggunakan test database PostgreSQL pada Docker untuk memverifikasi isolasi multi-tenant:
     - Mahasiswa Kampus A memanggil invoice Kampus B -> Pastikan mengembalikan 404/403.
     - Race condition test: Menjalankan 10 simultaneous webhook requests untuk invoice yang sama -> Hanya 1 yang sukses memproses, 9 lainnya direspons sebagai idempotent duplicate.
