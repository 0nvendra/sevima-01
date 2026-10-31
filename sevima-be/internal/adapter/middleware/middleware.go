package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sevima-be/internal/domain/entity"
	"sevima-be/internal/pkg/crypto"
	"sevima-be/internal/pkg/response"
)

const (
	UserIDKey   = "user_id"
	RoleKey     = "user_role"
	TenantIDKey = "tenant_id"
)

func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Unauthorized(c, "Authorization token is missing or malformed")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := crypto.ParseAccessToken(tokenString, jwtSecret)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired access token")
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(RoleKey, claims.Role)
		if claims.InstitusiID != nil {
			c.Set(TenantIDKey, *claims.InstitusiID)
		}

		c.Next()
	}
}

func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(RoleKey)
		if !exists {
			response.Unauthorized(c, "Unauthorized context")
			c.Abort()
			return
		}

		role := roleVal.(entity.UserRole)
		if role == entity.RoleSuperAdmin {
			instHeader := c.GetHeader("X-Institution-ID")
			if instHeader != "" {
				if instID, err := uuid.Parse(instHeader); err == nil {
					c.Set(TenantIDKey, instID)
					ctx := context.WithValue(c.Request.Context(), TenantIDKey, instID)
					c.Request = c.Request.WithContext(ctx)
				}
			}
			c.Next()
			return
		}

		tenantVal, hasTenant := c.Get(TenantIDKey)
		if !hasTenant {
			response.Forbidden(c, "User has no institution assigned")
			c.Abort()
			return
		}

		tenantID := tenantVal.(uuid.UUID)
		ctx := context.WithValue(c.Request.Context(), TenantIDKey, tenantID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func RequireRole(allowedRoles ...entity.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(RoleKey)
		if !exists {
			response.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		userRole := roleVal.(entity.UserRole)
		for _, allowed := range allowedRoles {
			if userRole == allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "Forbidden: insufficient permissions")
		c.Abort()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Signature, X-Institution-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
