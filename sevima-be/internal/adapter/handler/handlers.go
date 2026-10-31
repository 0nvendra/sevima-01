package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sevima-be/internal/adapter/middleware"
	"sevima-be/internal/domain/entity"
	domainrepo "sevima-be/internal/domain/repository"
	"sevima-be/internal/pkg/response"
	"sevima-be/internal/usecase"
)

type Handlers struct {
	uc *usecase.UseCases
}

func NewHandlers(uc *usecase.UseCases) *Handlers {
	return &Handlers{uc: uc}
}

// ---------------- Auth Handlers ----------------

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Email dan password wajib diisi", err.Error())
		return
	}

	res, err := h.uc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Login berhasil", res, nil)
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *Handlers) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Refresh token wajib disertakan", err.Error())
		return
	}

	newAccess, newRefresh, err := h.uc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "Refresh token tidak valid atau telah kedaluwarsa")
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Token berhasil diperbarui", gin.H{
		"access_token":  newAccess,
		"refresh_token": newRefresh,
		"expires_in":    900,
	}, nil)
}

func (h *Handlers) Logout(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.UserIDKey)
	userID := userIDVal.(uuid.UUID)

	_ = h.uc.Logout(c.Request.Context(), userID)
	response.JSONSuccess(c, http.StatusOK, "Logout berhasil", nil, nil)
}

func (h *Handlers) GetMe(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.UserIDKey)
	userID := userIDVal.(uuid.UUID)

	user, inst, student, err := h.uc.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "User profile not found")
		return
	}

	data := gin.H{
		"user":        user,
		"institution": inst,
		"student":     student,
	}
	response.JSONSuccess(c, http.StatusOK, "Profil pengguna", data, nil)
}

// ---------------- Institution Handlers ----------------

func (h *Handlers) ListInstitutions(c *gin.Context) {
	list, err := h.uc.ListInstitutions(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.JSONSuccess(c, http.StatusOK, "Daftar institusi", list, nil)
}

type CreateInstitutionRequest struct {
	Code    string `json:"code" binding:"required"`
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
	Email   string `json:"email" binding:"required,email"`
}

func (h *Handlers) CreateInstitution(c *gin.Context) {
	var req CreateInstitutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Data institusi tidak valid", err.Error())
		return
	}

	inst := &entity.Institution{
		Code:    req.Code,
		Name:    req.Name,
		Address: req.Address,
		Phone:   req.Phone,
		Email:   req.Email,
	}

	created, err := h.uc.CreateInstitution(c.Request.Context(), inst)
	if err != nil {
		response.BadRequest(c, "CREATION_FAILED", "Gagal menambahkan institusi (kode mungkin telah digunakan)", err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusCreated, "Institusi berhasil ditambahkan", created, nil)
}

// ---------------- Invoice Handlers ----------------

func (h *Handlers) ListInvoices(c *gin.Context) {
	roleVal, _ := c.Get(middleware.RoleKey)
	role := roleVal.(entity.UserRole)
	userIDVal, _ := c.Get(middleware.UserIDKey)
	userID := userIDVal.(uuid.UUID)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "10"))
	search := c.Query("search")
	academicYear := c.Query("academic_year")

	filter := domainrepo.InvoiceFilter{
		Page:    page,
		PerPage: perPage,
	}
	if search != "" {
		filter.Search = &search
	}
	if academicYear != "" {
		filter.AcademicYear = &academicYear
	}
	if status := c.Query("status"); status != "" {
		invStatus := entity.InvoiceStatus(status)
		filter.Status = &invStatus
	}
	if semester := c.Query("semester"); semester != "" {
		sem := entity.SemesterType(semester)
		filter.Semester = &sem
	}

	// Multi-tenant filtering rules
	if role == entity.RoleMahasiswa {
		// Only list invoices belonging to this student
		_, _, student, err := h.uc.GetUserProfile(c.Request.Context(), userID)
		if err != nil || student == nil {
			response.Forbidden(c, "Akun tidak terhubung dengan profil mahasiswa")
			return
		}
		filter.StudentID = &student.ID
		filter.InstitusiID = &student.InstitusiID
	} else if role == entity.RoleAdminKeuangan {
		tenantVal, exists := c.Get(middleware.TenantIDKey)
		if !exists {
			response.Forbidden(c, "No tenant bound")
			return
		}
		tenantID := tenantVal.(uuid.UUID)
		filter.InstitusiID = &tenantID
	} else if role == entity.RoleSuperAdmin {
		if instIDStr := c.Query("institusi_id"); instIDStr != "" {
			if id, err := uuid.Parse(instIDStr); err == nil {
				filter.InstitusiID = &id
			}
		}
	}

	invoices, total, err := h.uc.ListInvoices(c.Request.Context(), filter)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	meta := &response.Meta{
		Page:         page,
		PerPage:      perPage,
		TotalRecords: total,
		TotalPages:   totalPages,
	}

	response.JSONSuccess(c, http.StatusOK, "Daftar tagihan", invoices, meta)
}

func (h *Handlers) GetInvoiceDetail(c *gin.Context) {
	idStr := c.Param("id")
	invID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "INVALID_ID", "Format ID invoice tidak valid", nil)
		return
	}

	roleVal, _ := c.Get(middleware.RoleKey)
	role := roleVal.(entity.UserRole)
	userIDVal, _ := c.Get(middleware.UserIDKey)
	userID := userIDVal.(uuid.UUID)

	var requestingTenant *uuid.UUID
	if tenantVal, exists := c.Get(middleware.TenantIDKey); exists {
		tid := tenantVal.(uuid.UUID)
		requestingTenant = &tid
	}

	inv, history, err := h.uc.GetInvoiceDetail(c.Request.Context(), invID, requestingTenant, role, userID)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Detail tagihan", gin.H{
		"invoice": inv,
		"history": history,
	}, nil)
}

type CreateInvoiceRequest struct {
	StudentID    string `json:"student_id" binding:"required"`
	Amount       int64  `json:"amount" binding:"required,gt=0"`
	Semester     string `json:"semester" binding:"required"`
	AcademicYear string `json:"academic_year" binding:"required"`
	DueDate      string `json:"due_date" binding:"required"`
	Notes        string `json:"notes"`
}

func (h *Handlers) CreateInvoice(c *gin.Context) {
	tenantVal, exists := c.Get(middleware.TenantIDKey)
	if !exists {
		response.Forbidden(c, "No tenant bound to admin session")
		return
	}
	tenantID := tenantVal.(uuid.UUID)

	var req CreateInvoiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Data tagihan tidak lengkap", err.Error())
		return
	}

	studentUUID, err := uuid.Parse(req.StudentID)
	if err != nil {
		response.BadRequest(c, "INVALID_STUDENT_ID", "Format student_id tidak valid", nil)
		return
	}

	dueDate, err := time.Parse("2006-01-02", req.DueDate)
	if err != nil {
		dueDate, err = time.Parse(time.RFC3339, req.DueDate)
		if err != nil {
			response.BadRequest(c, "INVALID_DATE", "Format tanggal jatuh tempo harus YYYY-MM-DD", nil)
			return
		}
	}

	inv := &entity.Invoice{
		InstitusiID:  tenantID,
		StudentID:    studentUUID,
		Amount:       req.Amount,
		Semester:     entity.SemesterType(req.Semester),
		AcademicYear: req.AcademicYear,
		DueDate:      dueDate,
		Notes:        req.Notes,
	}

	if err := h.uc.CreateInvoice(c.Request.Context(), inv); err != nil {
		response.InternalServerError(c, "Gagal membuat tagihan: "+err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusCreated, "Tagihan berhasil diterbitkan", inv, nil)
}

// ---------------- Student Handlers ----------------

func (h *Handlers) GetKrsStatus(c *gin.Context) {
	userIDVal, _ := c.Get(middleware.UserIDKey)
	userID := userIDVal.(uuid.UUID)

	semester := entity.SemesterType(c.DefaultQuery("semester", "ganjil"))
	academicYear := c.DefaultQuery("academic_year", "2026/2027")

	krsStatus, err := h.uc.GetStudentKrsStatus(c.Request.Context(), userID, semester, academicYear)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Status KRS mahasiswa", krsStatus, nil)
}

// ---------------- Webhook Handler ----------------

func (h *Handlers) ProcessWebhook(c *gin.Context) {
	signature := c.GetHeader("X-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Hub-Signature")
	}

	if signature == "" {
		response.JSONError(c, http.StatusUnauthorized, "MISSING_SIGNATURE", "Header X-Signature wajib disertakan", nil)
		return
	}

	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "BODY_READ_ERROR", "Tidak dapat membaca request body", err.Error())
		return
	}

	var payload usecase.WebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		response.BadRequest(c, "MALFORMED_JSON", "Format JSON payload tidak valid", err.Error())
		return
	}

	result, err := h.uc.ProcessWebhook(c.Request.Context(), rawBody, signature, &payload)
	if err != nil {
		if err.Error() == "invalid signature" {
			response.JSONError(c, http.StatusUnauthorized, "INVALID_SIGNATURE", "Verifikasi HMAC signature gagal", nil)
			return
		}
		response.BadRequest(c, "WEBHOOK_FAILED", err.Error(), nil)
		return
	}

	response.JSONSuccess(c, http.StatusOK, result.Message, result, nil)
}

// ---------------- Mock Gateway Simulator Handler ----------------

func (h *Handlers) TriggerMockPayment(c *gin.Context) {
	var req usecase.MockPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "VALIDATION_ERROR", "Data simulator tidak lengkap", err.Error())
		return
	}

	res, err := h.uc.TriggerMockPayment(c.Request.Context(), &req)
	if err != nil {
		response.BadRequest(c, "SIMULATION_FAILED", err.Error(), res)
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Simulasi pembayaran berhasil dieksekusi", res, nil)
}
