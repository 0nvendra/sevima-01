package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	adapterrepo "sevima-be/internal/adapter/repository"
	"sevima-be/internal/domain/entity"
	domainrepo "sevima-be/internal/domain/repository"
	"sevima-be/internal/pkg/crypto"
)

type UseCases struct {
	repo      *adapterrepo.PostgresRepositories
	jwtSecret string
}

func NewUseCases(repo *adapterrepo.PostgresRepositories, jwtSecret string) *UseCases {
	return &UseCases{repo: repo, jwtSecret: jwtSecret}
}

// ---------------- Auth UseCase ----------------

type LoginResponse struct {
	AccessToken     string       `json:"access_token"`
	RefreshToken    string       `json:"refresh_token"`
	ExpiresIn       int64        `json:"expires_in"`
	User            *entity.User `json:"user"`
	InstitutionCode string       `json:"institution_code,omitempty"`
	InstitutionName string       `json:"institution_name,omitempty"`
}

func (u *UseCases) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is inactive")
	}

	if !crypto.CheckPassword(password, user.Password) {
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := crypto.GenerateAccessToken(user, u.jwtSecret, 15*time.Minute)
	if err != nil {
		return nil, err
	}

	refreshToken, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return nil, err
	}

	if err := u.repo.Store(ctx, user.ID, refreshToken); err != nil {
		return nil, err
	}

	_ = u.repo.UpdateLastLogin(ctx, user.ID)

	var instCode, instName string
	if user.InstitusiID != nil {
		if inst, err := u.repo.FindByID(ctx, *user.InstitusiID); err == nil {
			instCode = inst.Code
			instName = inst.Name
		}
	}

	return &LoginResponse{
		AccessToken:     accessToken,
		RefreshToken:    refreshToken,
		ExpiresIn:       900,
		User:            user,
		InstitutionCode: instCode,
		InstitutionName: instName,
	}, nil
}

func (u *UseCases) Refresh(ctx context.Context, refreshToken string) (string, string, error) {
	userID, err := u.repo.ValidateAndRevoke(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	user, err := u.repo.FindUserByID(ctx, *userID)
	if err != nil || !user.IsActive {
		return "", "", errors.New("user not found or inactive")
	}

	newAccessToken, err := crypto.GenerateAccessToken(user, u.jwtSecret, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := crypto.GenerateSecureToken(32)
	if err != nil {
		return "", "", err
	}

	if err := u.repo.Store(ctx, user.ID, newRefreshToken); err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *UseCases) Logout(ctx context.Context, userID uuid.UUID) error {
	return u.repo.RevokeAllForUser(ctx, userID)
}

func (u *UseCases) GetUserProfile(ctx context.Context, userID uuid.UUID) (*entity.User, *entity.Institution, *entity.Student, error) {
	user, err := u.repo.FindUserByID(ctx, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	var inst *entity.Institution
	if user.InstitusiID != nil {
		inst, _ = u.repo.FindByID(ctx, *user.InstitusiID)
	}

	var student *entity.Student
	if user.Role == entity.RoleMahasiswa {
		student, _ = u.repo.FindByUserID(ctx, user.ID)
	}

	return user, inst, student, nil
}

// ---------------- Institution UseCase ----------------

func (u *UseCases) ListInstitutions(ctx context.Context) ([]entity.Institution, error) {
	return u.repo.FindAll(ctx)
}

func (u *UseCases) CreateInstitution(ctx context.Context, inst *entity.Institution) (*entity.Institution, error) {
	if inst.HMACSecret == "" {
		secret, err := crypto.GenerateSecureToken(32)
		if err != nil {
			return nil, err
		}
		inst.HMACSecret = secret
	}
	inst.IsActive = true
	if err := u.repo.Create(ctx, inst); err != nil {
		return nil, err
	}
	return inst, nil
}

// ---------------- Invoice UseCase ----------------

func (u *UseCases) ListInvoices(ctx context.Context, filter domainrepo.InvoiceFilter) ([]entity.Invoice, int64, error) {
	return u.repo.List(ctx, filter)
}

func (u *UseCases) GetInvoiceDetail(ctx context.Context, id uuid.UUID, requestingTenant *uuid.UUID, role entity.UserRole, userID uuid.UUID) (*entity.Invoice, []entity.TransactionHistory, error) {
	inv, err := u.repo.FindInvoiceByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	// Verify Tenant Isolation
	if role != entity.RoleSuperAdmin {
		if requestingTenant != nil && inv.InstitusiID != *requestingTenant {
			return nil, nil, errors.New("access denied: cross-tenant access prohibited")
		}
		if role == entity.RoleMahasiswa {
			student, err := u.repo.FindByUserID(ctx, userID)
			if err != nil || student.ID != inv.StudentID {
				return nil, nil, errors.New("access denied: unauthorized invoice access")
			}
		}
	}

	history, _ := u.repo.ListByInvoiceID(ctx, id)
	return inv, history, nil
}

func (u *UseCases) CreateInvoice(ctx context.Context, inv *entity.Invoice) error {
	inv.Status = entity.InvoiceStatusUnpaid
	if inv.InvoiceNumber == "" {
		inv.InvoiceNumber = fmt.Sprintf("INV-%d-%s", time.Now().UnixNano()%1000000, inv.AcademicYear)
	}
	if err := u.repo.CreateInvoice(ctx, inv); err != nil {
		return err
	}

	// Record initial transaction state
	txn := &entity.TransactionHistory{
		InvoiceID:      inv.ID,
		InstitusiID:    inv.InstitusiID,
		PreviousStatus: entity.InvoiceStatusUnpaid,
		NewStatus:      entity.InvoiceStatusUnpaid,
		Source:         "invoice_created",
		Notes:          "Tagihan baru diterbitkan",
	}
	return u.repo.CreateTransaction(ctx, txn)
}

type KrsStatusResponse struct {
	NIM                string `json:"nim"`
	FullName           string `json:"full_name"`
	IsEligibleKRS      bool   `json:"is_eligible_krs"`
	UnpaidInvoices     int    `json:"unpaid_invoices_count"`
	ActiveSemester     string `json:"active_semester"`
	ActiveAcademicYear string `json:"active_academic_year"`
	Message            string `json:"message"`
}

func (u *UseCases) GetStudentKrsStatus(ctx context.Context, userID uuid.UUID, semester entity.SemesterType, year string) (*KrsStatusResponse, error) {
	student, err := u.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("student record not found")
	}

	unpaidCount, err := u.repo.CountUnpaidForStudent(ctx, student.ID, semester, year)
	if err != nil {
		return nil, err
	}

	isEligible := (unpaidCount == 0)
	var message string
	if isEligible {
		message = fmt.Sprintf("Tagihan semester %s tahun %s telah lunas. Anda dapat mengisi KRS.", semester, year)
	} else {
		message = fmt.Sprintf("Terdapat %d tagihan belum terselesaikan. Pelunasan tagihan wajib sebelum pengisian KRS.", unpaidCount)
	}

	return &KrsStatusResponse{
		NIM:                student.NIM,
		FullName:           student.User.FullName,
		IsEligibleKRS:      isEligible,
		UnpaidInvoices:     unpaidCount,
		ActiveSemester:     string(semester),
		ActiveAcademicYear: year,
		Message:            message,
	}, nil
}

// ---------------- Webhook & Payment Engine ----------------

type WebhookPayload struct {
	InstitutionCode string `json:"institution_code"`
	EventType       string `json:"event_type"` // payment.success, payment.expired
	InvoiceNumber   string `json:"invoice_number"`
	Amount          int64  `json:"amount"`
	PaymentMethod   string `json:"payment_method"`
	PaidAt          string `json:"paid_at"`
	TransactionID   string `json:"transaction_id"`
}

type WebhookResult struct {
	Success       bool   `json:"success"`
	Status        string `json:"status"` // processed, duplicate, rejected
	InvoiceNumber string `json:"invoice_number"`
	Message       string `json:"message"`
}

func (u *UseCases) ProcessWebhook(ctx context.Context, rawBody []byte, signature string, payload *WebhookPayload) (*WebhookResult, error) {
	// 1. Identify institution
	inst, err := u.repo.FindByCode(ctx, payload.InstitutionCode)
	if err != nil {
		_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
			PayloadRaw:     string(rawBody),
			SignatureRecv:  signature,
			SignatureValid: false,
			Status:         "invalid_institution",
			ErrorMessage:   "Institution code not found",
		})
		return nil, errors.New("institution not found")
	}

	// 2. Validate HMAC-SHA256 signature
	validSignature := crypto.VerifyHMAC(rawBody, signature, inst.HMACSecret)
	if !validSignature {
		_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
			InstitusiID:    &inst.ID,
			PayloadRaw:     string(rawBody),
			SignatureRecv:  signature,
			SignatureValid: false,
			Status:         "invalid_signature",
			ErrorMessage:   "HMAC verification failed",
		})
		return nil, errors.New("invalid signature")
	}

	// 3. Begin Atomic Transaction with Row Locking
	tx, err := u.repo.Pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// SELECT FOR UPDATE to prevent concurrent race condition
	inv, err := u.repo.FindByNumberForUpdateTx(ctx, tx, inst.ID, payload.InvoiceNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
				InstitusiID:    &inst.ID,
				PayloadRaw:     string(rawBody),
				SignatureRecv:  signature,
				SignatureValid: true,
				Status:         "invoice_not_found",
				ErrorMessage:   "Invoice not found",
			})
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	result := &WebhookResult{InvoiceNumber: inv.InvoiceNumber}

	switch payload.EventType {
	case "payment.success":
		// Idempotency: already paid -> return success without modifying state
		if inv.Status == entity.InvoiceStatusPaid {
			result.Success = true
			result.Status = "duplicate"
			result.Message = "Invoice already marked as paid (idempotent no-op)"
			_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
				InstitusiID:    &inst.ID,
				InvoiceID:      &inv.ID,
				PayloadRaw:     string(rawBody),
				SignatureRecv:  signature,
				SignatureValid: true,
				Status:         "duplicate",
			})
			return result, nil
		}

		if inv.Status == entity.InvoiceStatusExpired {
			_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
				InstitusiID:    &inst.ID,
				InvoiceID:      &inv.ID,
				PayloadRaw:     string(rawBody),
				SignatureRecv:  signature,
				SignatureValid: true,
				Status:         "invalid_transition",
				ErrorMessage:   "Cannot pay expired invoice",
			})
			return nil, errors.New("cannot pay expired invoice")
		}

		// Amount integrity validation
		if payload.Amount != inv.Amount {
			_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
				InstitusiID:    &inst.ID,
				InvoiceID:      &inv.ID,
				PayloadRaw:     string(rawBody),
				SignatureRecv:  signature,
				SignatureValid: true,
				Status:         "amount_mismatch",
				ErrorMessage:   fmt.Sprintf("Expected %d, got %d", inv.Amount, payload.Amount),
			})
			return nil, errors.New("paid amount does not match invoice amount")
		}

		paidAt, _ := time.Parse(time.RFC3339, payload.PaidAt)
		if paidAt.IsZero() {
			paidAt = time.Now()
		}

		prevStatus := inv.Status
		inv.Status = entity.InvoiceStatusPaid
		inv.PaidAt = &paidAt
		inv.PaidAmount = &payload.Amount
		inv.PaymentMethod = &payload.PaymentMethod

		if err := u.repo.UpdateInvoiceTx(ctx, tx, inv); err != nil {
			return nil, err
		}

		txn := &entity.TransactionHistory{
			InvoiceID:      inv.ID,
			InstitusiID:    inv.InstitusiID,
			PreviousStatus: prevStatus,
			NewStatus:      entity.InvoiceStatusPaid,
			Source:         "webhook_payment",
			Notes:          fmt.Sprintf("Method: %s, TrxID: %s", payload.PaymentMethod, payload.TransactionID),
		}
		if err := u.repo.CreateTransactionTx(ctx, tx, txn); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
			InstitusiID:    &inst.ID,
			InvoiceID:      &inv.ID,
			PayloadRaw:     string(rawBody),
			SignatureRecv:  signature,
			SignatureValid: true,
			Status:         "processed",
		})

		result.Success = true
		result.Status = "processed"
		result.Message = "Payment successfully verified and updated"
		return result, nil

	case "payment.expired":
		if inv.Status == entity.InvoiceStatusPaid {
			result.Success = true
			result.Status = "rejected"
			result.Message = "Cannot expire an already paid invoice"
			return result, nil
		}
		if inv.Status == entity.InvoiceStatusExpired {
			result.Success = true
			result.Status = "duplicate"
			result.Message = "Invoice already expired"
			return result, nil
		}

		prevStatus := inv.Status
		inv.Status = entity.InvoiceStatusExpired
		if err := u.repo.UpdateInvoiceTx(ctx, tx, inv); err != nil {
			return nil, err
		}

		txn := &entity.TransactionHistory{
			InvoiceID:      inv.ID,
			InstitusiID:    inv.InstitusiID,
			PreviousStatus: prevStatus,
			NewStatus:      entity.InvoiceStatusExpired,
			Source:         "webhook_expiry",
			Notes:          "Status marked expired via webhook notification",
		}
		if err := u.repo.CreateTransactionTx(ctx, tx, txn); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}

		_ = u.repo.CreateWebhookLog(ctx, &entity.WebhookLog{
			InstitusiID:    &inst.ID,
			InvoiceID:      &inv.ID,
			PayloadRaw:     string(rawBody),
			SignatureRecv:  signature,
			SignatureValid: true,
			Status:         "processed",
		})

		result.Success = true
		result.Status = "processed"
		result.Message = "Invoice marked expired"
		return result, nil

	default:
		return nil, errors.New("unsupported event_type")
	}
}

// ---------------- Mock Gateway Simulator ----------------

type MockPaymentRequest struct {
	InstitutionCode string `json:"institution_code"`
	InvoiceNumber   string `json:"invoice_number"`
	PaymentMethod   string `json:"payment_method"`
	SimulateTamper  bool   `json:"simulate_tamper"`
	EventType       string `json:"event_type"` // payment.success or payment.expired
}

type MockPaymentResponse struct {
	CalculatedSignature string         `json:"calculated_signature"`
	WebhookPayload      WebhookPayload `json:"webhook_payload"`
	WebhookResult       *WebhookResult `json:"webhook_result"`
}

func (u *UseCases) TriggerMockPayment(ctx context.Context, req *MockPaymentRequest) (*MockPaymentResponse, error) {
	inst, err := u.repo.FindByCode(ctx, req.InstitutionCode)
	if err != nil {
		return nil, fmt.Errorf("institution '%s' not found", req.InstitutionCode)
	}

	// Find invoice amount
	invFilter := domainrepo.InvoiceFilter{
		InstitusiID: &inst.ID,
		Search:      &req.InvoiceNumber,
		PerPage:     1,
		Page:        1,
	}
	invoices, _, err := u.repo.List(ctx, invFilter)
	if err != nil || len(invoices) == 0 {
		return nil, fmt.Errorf("invoice '%s' not found in institution %s", req.InvoiceNumber, req.InstitutionCode)
	}
	targetInv := invoices[0]

	eventType := req.EventType
	if eventType == "" {
		eventType = "payment.success"
	}

	method := req.PaymentMethod
	if method == "" {
		method = "BCA_VA"
	}

	payload := WebhookPayload{
		InstitutionCode: inst.Code,
		EventType:       eventType,
		InvoiceNumber:   targetInv.InvoiceNumber,
		Amount:          targetInv.Amount,
		PaymentMethod:   method,
		PaidAt:          time.Now().Format(time.RFC3339),
		TransactionID:   fmt.Sprintf("TRX-SIM-%d", time.Now().Unix()),
	}

	rawBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	signature := crypto.GenerateHMAC(rawBody, inst.HMACSecret)
	if req.SimulateTamper {
		signature = signature[:len(signature)-4] + "dead"
	}

	result, err := u.ProcessWebhook(ctx, rawBody, signature, &payload)
	if err != nil {
		return &MockPaymentResponse{
			CalculatedSignature: signature,
			WebhookPayload:      payload,
			WebhookResult:       nil,
		}, err
	}

	return &MockPaymentResponse{
		CalculatedSignature: signature,
		WebhookPayload:      payload,
		WebhookResult:       result,
	}, nil
}
