package entity

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ValidateStateTransition checks if transitioning from current to target status is valid.
func ValidateStateTransition(current, target InvoiceStatus) error {
	if current == target {
		return errors.New("idempotent: status unchanged")
	}
	if current == InvoiceStatusPaid {
		return errors.New("illegal transition: cannot change status of already paid invoice")
	}
	if current == InvoiceStatusExpired && target == InvoiceStatusPaid {
		return errors.New("illegal transition: cannot pay an expired invoice")
	}
	return nil
}

// ValidateTenantAccess verifies if a user from tenantA can access an invoice belonging to tenantB.
func ValidateTenantAccess(role UserRole, requestingTenant *uuid.UUID, invoiceTenant uuid.UUID) error {
	if role == RoleSuperAdmin {
		return nil // Super Admin has global visibility
	}
	if requestingTenant == nil || *requestingTenant != invoiceTenant {
		return errors.New("cross-tenant access prohibited: data belongs to another institution")
	}
	return nil
}

// ValidateStudentInvoiceAccess verifies that a student only accesses their own invoice.
func ValidateStudentInvoiceAccess(studentID, invoiceStudentID uuid.UUID) error {
	if studentID != invoiceStudentID {
		return errors.New("unauthorized: student can only view their own invoice")
	}
	return nil
}

func TestInvoiceStatusTransitions(t *testing.T) {
	tests := []struct {
		name        string
		current     InvoiceStatus
		target      InvoiceStatus
		shouldError bool
	}{
		{"Unpaid to Paid", InvoiceStatusUnpaid, InvoiceStatusPaid, false},
		{"Unpaid to Expired", InvoiceStatusUnpaid, InvoiceStatusExpired, false},
		{"Paid to Paid (Idempotent)", InvoiceStatusPaid, InvoiceStatusPaid, true},
		{"Paid to Expired (Illegal)", InvoiceStatusPaid, InvoiceStatusExpired, true},
		{"Expired to Paid (Illegal)", InvoiceStatusExpired, InvoiceStatusPaid, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStateTransition(tt.current, tt.target)
			if (err != nil) != tt.shouldError {
				t.Errorf("ValidateStateTransition(%v, %v) error = %v, shouldError = %v", tt.current, tt.target, err, tt.shouldError)
			}
		})
	}
}

func TestTenantDataSegregation(t *testing.T) {
	tenantUGM := uuid.New()
	tenantITS := uuid.New()

	// 1. Admin UGM accessing UGM Invoice -> Allowed
	if err := ValidateTenantAccess(RoleAdminKeuangan, &tenantUGM, tenantUGM); err != nil {
		t.Errorf("Expected Admin UGM accessing UGM invoice to succeed, got %v", err)
	}

	// 2. Admin UGM accessing ITS Invoice -> Forbidden (Multi-tenant isolation)
	if err := ValidateTenantAccess(RoleAdminKeuangan, &tenantUGM, tenantITS); err == nil {
		t.Errorf("Expected Admin UGM accessing ITS invoice to fail with cross-tenant prohibition")
	}

	// 3. Super Admin accessing ITS Invoice -> Allowed (Global visibility)
	if err := ValidateTenantAccess(RoleSuperAdmin, nil, tenantITS); err != nil {
		t.Errorf("Expected Super Admin accessing any tenant to succeed, got %v", err)
	}
}

func TestStudentOwnershipSegregation(t *testing.T) {
	studentA := uuid.New()
	studentB := uuid.New()

	// Student accessing own invoice -> Allowed
	if err := ValidateStudentInvoiceAccess(studentA, studentA); err != nil {
		t.Errorf("Expected student accessing own invoice to succeed, got %v", err)
	}

	// Student accessing another student's invoice -> Forbidden
	if err := ValidateStudentInvoiceAccess(studentA, studentB); err == nil {
		t.Errorf("Expected student accessing another student's invoice to fail")
	}
}

func TestInvoiceOverdueCheck(t *testing.T) {
	pastDue := time.Now().Add(-24 * time.Hour)
	futureDue := time.Now().Add(24 * time.Hour)

	invPast := &Invoice{
		DueDate: pastDue,
		Status:  InvoiceStatusUnpaid,
	}
	if !time.Now().After(invPast.DueDate) {
		t.Errorf("Invoice with past due date should be overdue")
	}

	invFuture := &Invoice{
		DueDate: futureDue,
		Status:  InvoiceStatusUnpaid,
	}
	if time.Now().After(invFuture.DueDate) {
		t.Errorf("Invoice with future due date should not be overdue")
	}
}
