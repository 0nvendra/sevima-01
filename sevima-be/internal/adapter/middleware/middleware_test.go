package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sevima-be/internal/domain/entity"
	"sevima-be/internal/pkg/crypto"
)

func init() {
	gin.SetMode(gin.TestMode)
}

const testJWTSecret = "test-secret-key-for-tenant-isolation-test-12345"

func createTestToken(role entity.UserRole, tenantID *uuid.UUID) string {
	user := &entity.User{
		ID:          uuid.New(),
		Role:        role,
		InstitusiID: tenantID,
		Email:       "test@institution.ac.id",
	}
	token, _ := crypto.GenerateAccessToken(user, testJWTSecret, 15*time.Minute)
	return token
}

func TestAuthMiddleware_Success(t *testing.T) {
	tenantID := uuid.New()
	token := createTestToken(entity.RoleAdminKeuangan, &tenantID)

	router := gin.New()
	router.Use(AuthMiddleware(testJWTSecret))
	router.GET("/test-auth", func(c *gin.Context) {
		role, _ := c.Get(RoleKey)
		tenant, _ := c.Get(TenantIDKey)
		c.JSON(http.StatusOK, gin.H{
			"role":      role,
			"tenant_id": tenant,
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test-auth", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware(testJWTSecret))
	router.GET("/test-auth", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test-auth", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 Unauthorized, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware(testJWTSecret))
	router.GET("/test-auth", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/test-auth", nil)
	req.Header.Set("Authorization", "Bearer invalid-garbage-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status 401 Unauthorized, got %d", w.Code)
	}
}

func TestTenantMiddleware_TenantIsolation(t *testing.T) {
	tenantUGM := uuid.New()
	tenantITS := uuid.New()

	router := gin.New()
	router.Use(AuthMiddleware(testJWTSecret))
	router.Use(TenantMiddleware())
	router.GET("/data", func(c *gin.Context) {
		tenantVal, _ := c.Get(TenantIDKey)
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantVal})
	})

	// Test 1: Admin UGM has tenant context bound to UGM
	tokenUGM := createTestToken(entity.RoleAdminKeuangan, &tenantUGM)
	reqUGM, _ := http.NewRequest(http.MethodGet, "/data", nil)
	reqUGM.Header.Set("Authorization", "Bearer "+tokenUGM)
	wUGM := httptest.NewRecorder()
	router.ServeHTTP(wUGM, reqUGM)

	if wUGM.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for Admin UGM, got %d", wUGM.Code)
	}

	// Test 2: Admin without institution assigned must be blocked with 403 Forbidden
	tokenNoTenant := createTestToken(entity.RoleAdminKeuangan, nil)
	reqNoTenant, _ := http.NewRequest(http.MethodGet, "/data", nil)
	reqNoTenant.Header.Set("Authorization", "Bearer "+tokenNoTenant)
	wNoTenant := httptest.NewRecorder()
	router.ServeHTTP(wNoTenant, reqNoTenant)

	if wNoTenant.Code != http.StatusForbidden {
		t.Fatalf("Expected status 403 Forbidden for admin without tenant, got %d", wNoTenant.Code)
	}

	// Test 3: Super Admin can access and optionally specify X-Institution-ID
	tokenSuper := createTestToken(entity.RoleSuperAdmin, nil)
	reqSuper, _ := http.NewRequest(http.MethodGet, "/data", nil)
	reqSuper.Header.Set("Authorization", "Bearer "+tokenSuper)
	reqSuper.Header.Set("X-Institution-ID", tenantITS.String())
	wSuper := httptest.NewRecorder()
	router.ServeHTTP(wSuper, reqSuper)

	if wSuper.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for Super Admin, got %d", wSuper.Code)
	}
}

func TestRequireRole_RBAC(t *testing.T) {
	tenantID := uuid.New()
	adminToken := createTestToken(entity.RoleAdminKeuangan, &tenantID)
	mhsToken := createTestToken(entity.RoleMahasiswa, &tenantID)

	router := gin.New()
	router.Use(AuthMiddleware(testJWTSecret))
	router.POST("/admin-only", RequireRole(entity.RoleAdminKeuangan), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// Test 1: Admin Keuangan accessing admin route -> allowed
	reqAdmin, _ := http.NewRequest(http.MethodPost, "/admin-only", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+adminToken)
	wAdmin := httptest.NewRecorder()
	router.ServeHTTP(wAdmin, reqAdmin)

	if wAdmin.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for Admin Keuangan, got %d", wAdmin.Code)
	}

	// Test 2: Mahasiswa trying to access admin route -> 403 Forbidden
	reqMhs, _ := http.NewRequest(http.MethodPost, "/admin-only", nil)
	reqMhs.Header.Set("Authorization", "Bearer "+mhsToken)
	wMhs := httptest.NewRecorder()
	router.ServeHTTP(wMhs, reqMhs)

	if wMhs.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for Mahasiswa on admin-only route, got %d", wMhs.Code)
	}
}
