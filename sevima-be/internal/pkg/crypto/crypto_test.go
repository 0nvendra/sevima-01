package crypto

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"sevima-be/internal/domain/entity"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "SecretPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Errorf("Expected password check to succeed")
	}

	if CheckPassword("WrongPassword", hash) {
		t.Errorf("Expected wrong password check to fail")
	}
}

func TestHMACGenerationAndVerification(t *testing.T) {
	secretKey := "test-secret-key-12345"
	payload := []byte(`{"institution_code":"UGM","amount":7500000,"invoice_number":"INV-001"}`)

	signature := GenerateHMAC(payload, secretKey)
	if signature == "" {
		t.Fatalf("Expected non-empty HMAC signature")
	}

	// Valid signature
	if !VerifyHMAC(payload, signature, secretKey) {
		t.Errorf("Expected HMAC verification to succeed for valid signature")
	}

	// Tampered payload
	tamperedPayload := []byte(`{"institution_code":"UGM","amount":1000000,"invoice_number":"INV-001"}`)
	if VerifyHMAC(tamperedPayload, signature, secretKey) {
		t.Errorf("Expected HMAC verification to fail for tampered payload")
	}

	// Wrong secret
	if VerifyHMAC(payload, signature, "different-secret-key") {
		t.Errorf("Expected HMAC verification to fail for wrong secret key")
	}
}

func TestJWTTokenGenerationAndParse(t *testing.T) {
	jwtSecret := "super-test-jwt-secret-xyz"
	instID := uuid.New()
	user := &entity.User{
		ID:          uuid.New(),
		Role:        entity.RoleAdminKeuangan,
		InstitusiID: &instID,
		Email:       "admin@ugm.ac.id",
	}

	tokenStr, err := GenerateAccessToken(user, jwtSecret, 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	claims, err := ParseAccessToken(tokenStr, jwtSecret)
	if err != nil {
		t.Fatalf("Failed to parse access token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user_id %s, got %s", user.ID, claims.UserID)
	}
	if claims.Role != entity.RoleAdminKeuangan {
		t.Errorf("Expected role %s, got %s", entity.RoleAdminKeuangan, claims.Role)
	}
	if claims.InstitusiID == nil || *claims.InstitusiID != instID {
		t.Errorf("Expected institusi_id %s, got %v", instID, claims.InstitusiID)
	}
}
