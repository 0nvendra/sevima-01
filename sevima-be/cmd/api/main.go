package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"sevima-be/internal/adapter/handler"
	"sevima-be/internal/adapter/middleware"
	"sevima-be/internal/adapter/repository"
	"sevima-be/internal/domain/entity"
	"sevima-be/internal/pkg/database"
	"sevima-be/internal/usecase"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "sevima_pay"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-jwt-key-sevima-pay-2026-production"
	}

	ctx := context.Background()
	dbCfg := database.Config{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		DBName:   dbName,
		SSLMode:  "disable",
	}

	pool, err := database.NewPostgresPool(ctx, dbCfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	repos := repository.NewPostgresRepositories(pool)
	uc := usecase.NewUseCases(repos, jwtSecret)
	h := handler.NewHandlers(uc)

	// Background worker for invoice auto-expiry
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if count, err := repos.ExpireOverdueInvoices(context.Background()); err == nil && count > 0 {
				log.Printf("[Auto-Expiry] Marked %d overdue invoices as expired", count)
			}
		}
	}()

	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "UP",
			"timestamp": time.Now(),
			"service":   "sevima-pay-backend",
		})
	})

	api := router.Group("/api/v1")
	{
		// Public Auth
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/login", h.Login)
			authGroup.POST("/refresh", h.Refresh)

			authGroup.Use(middleware.AuthMiddleware(jwtSecret))
			authGroup.POST("/logout", h.Logout)
			authGroup.GET("/me", h.GetMe)
		}

		// Webhook (Public with HMAC-SHA256 verification)
		api.POST("/webhook/payment", h.ProcessWebhook)

		// Mock Payment Gateway Simulator
		api.POST("/mock-gateway/pay", h.TriggerMockPayment)

		// Authenticated Routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(jwtSecret))
		protected.Use(middleware.TenantMiddleware())
		{
			// Super Admin Only
			superAdmin := protected.Group("/institutions")
			superAdmin.Use(middleware.RequireRole(entity.RoleSuperAdmin))
			{
				superAdmin.GET("", h.ListInstitutions)
				superAdmin.POST("", h.CreateInstitution)
			}

			// Invoices (Admin Keuangan & Mahasiswa)
			invoices := protected.Group("/invoices")
			{
				invoices.GET("", middleware.RequireRole(entity.RoleSuperAdmin, entity.RoleAdminKeuangan, entity.RoleMahasiswa), h.ListInvoices)
				invoices.GET("/:id", middleware.RequireRole(entity.RoleSuperAdmin, entity.RoleAdminKeuangan, entity.RoleMahasiswa), h.GetInvoiceDetail)
				invoices.POST("", middleware.RequireRole(entity.RoleAdminKeuangan), h.CreateInvoice)
			}

			// Students & KRS
			students := protected.Group("/students")
			{
				students.GET("/krs-status", middleware.RequireRole(entity.RoleMahasiswa), h.GetKrsStatus)
			}
		}
	}

	fmt.Printf("\n🚀 SEVIMA-PAY Backend Server running on http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}
