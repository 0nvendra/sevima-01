# Architecture Document
# SEVIMA-PAY: Multi-Tenant UKT Payment System

> **Version**: 1.0.0
> **Stack**: Go (Gin) + React (Vite) + PostgreSQL
> **Pattern**: Clean Architecture (both BE and FE)

---

## 1. System Overview

```mermaid
graph TB
    subgraph "Client Layer"
        Browser["Browser"]
    end

    subgraph "Frontend - React (Vite)"
        SK["SPA Router / React Router"]
        Pages["Views / Pages"]
        Stores["Zustand / TanStack Query"]
        API_Client["Axios API Client"]
    end

    subgraph "Backend - Go (Gin)"
        Router["Gin Router"]
        MW["Middleware Stack"]
        Handlers["HTTP Handlers"]
        UseCases["Use Cases (Business Logic)"]
        Repos["Repositories (Data Access)"]
        SQLC["sqlc Generated Code"]
    end

    subgraph "Mock Payment Gateway"
        MockUI["Mock Gateway UI"]
        MockBE["Mock Gateway Service"]
    end

    subgraph "Data Layer"
        PG[("PostgreSQL 16")]
    end

    Browser --> SK
    SK --> Pages
    Pages --> Stores
    Stores --> API_Client
    API_Client -->|"REST API"| Router

    Router --> MW
    MW --> Handlers
    Handlers --> UseCases
    UseCases --> Repos
    Repos --> SQLC
    SQLC --> PG

    MockUI --> MockBE
    MockBE -->|"Webhook POST"| Router

    style Browser fill:#FF5722,color:#fff
    style SK fill:#FF5722,color:#fff
    style Router fill:#2196F3,color:#fff
    style PG fill:#4CAF50,color:#fff
    style MockBE fill:#9C27B0,color:#fff
```

---

## 2. Clean Architecture — Backend (Go)

### 2.1 Layer Diagram

```mermaid
graph TB
    subgraph "Layer 4: Infrastructure (outermost)"
        Config["config/"]
        DB["database/"]
        Migration["migrations/"]
        Docker["Dockerfile"]
    end

    subgraph "Layer 3: Interface Adapters"
        Handlers["handler/ (HTTP handlers)"]
        Middleware["middleware/ (Auth, RBAC, Tenant)"]
        Router["router/ (Route definitions)"]
        DTO["dto/ (Request/Response structs)"]
    end

    subgraph "Layer 2: Application (Use Cases)"
        Services["usecase/ (Business logic)"]
    end

    subgraph "Layer 1: Domain (innermost)"
        Entities["domain/ (Entities, Enums)"]
        RepoInterfaces["domain/repository/ (Interfaces)"]
    end

    subgraph "Layer 3b: Data Access"
        RepoImpl["repository/ (Interface implementations)"]
        SQLCGen["internal/sqlc/ (Generated code)"]
    end

    Handlers --> Services
    Services --> RepoInterfaces
    RepoImpl --> RepoInterfaces
    RepoImpl --> SQLCGen
    SQLCGen --> DB
    Middleware --> Services
    Handlers --> DTO
    Services --> Entities

    style Entities fill:#E91E63,color:#fff
    style Services fill:#9C27B0,color:#fff
    style Handlers fill:#2196F3,color:#fff
    style Config fill:#607D8B,color:#fff
```

### 2.2 Directory Structure

```
sevima-be/
├── cmd/
│   └── server/
│       └── main.go                  # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go                # Environment config loading
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── institution.go       # Institution entity
│   │   │   ├── user.go              # User entity
│   │   │   ├── student.go           # Student entity
│   │   │   ├── invoice.go           # Invoice entity + status enum
│   │   │   ├── transaction.go       # Transaction history entity
│   │   │   └── webhook_log.go       # Webhook log entity
│   │   ├── repository/
│   │   │   ├── institution_repo.go  # Institution repository interface
│   │   │   ├── user_repo.go         # User repository interface
│   │   │   ├── student_repo.go      # Student repository interface
│   │   │   ├── invoice_repo.go      # Invoice repository interface
│   │   │   ├── transaction_repo.go  # Transaction repository interface
│   │   │   └── webhook_log_repo.go  # Webhook log repository interface
│   │   └── valueobject/
│   │       ├── role.go              # User role value object
│   │       ├── invoice_status.go    # Invoice status value object
│   │       └── semester.go          # Semester type value object
│   ├── usecase/
│   │   ├── auth_usecase.go          # Login, refresh, logout logic
│   │   ├── institution_usecase.go   # Institution CRUD logic
│   │   ├── invoice_usecase.go       # Invoice management logic
│   │   ├── webhook_usecase.go       # Webhook processing logic
│   │   ├── student_usecase.go       # Student data logic
│   │   └── transaction_usecase.go   # Transaction history logic
│   ├── handler/
│   │   ├── auth_handler.go          # Auth HTTP handlers
│   │   ├── institution_handler.go   # Institution HTTP handlers
│   │   ├── invoice_handler.go       # Invoice HTTP handlers
│   │   ├── webhook_handler.go       # Webhook HTTP handler
│   │   ├── student_handler.go       # Student HTTP handlers
│   │   └── health_handler.go        # Health check handler
│   ├── middleware/
│   │   ├── auth.go                  # JWT validation middleware
│   │   ├── rbac.go                  # Role-based access control
│   │   ├── tenant.go                # Tenant isolation middleware
│   │   ├── cors.go                  # CORS configuration
│   │   ├── logger.go                # Request logging middleware
│   │   └── ratelimit.go             # Rate limiting middleware
│   ├── dto/
│   │   ├── request/
│   │   │   ├── auth.go              # Login, refresh request DTOs
│   │   │   ├── invoice.go           # Create invoice, filter DTOs
│   │   │   └── webhook.go           # Webhook payload DTO
│   │   └── response/
│   │       ├── auth.go              # Token response DTOs
│   │       ├── invoice.go           # Invoice response DTOs
│   │       ├── common.go            # Pagination, error DTOs
│   │       └── student.go           # Student response DTOs
│   ├── repository/
│   │   └── postgres/
│   │       ├── institution_repo.go  # PostgreSQL institution repo
│   │       ├── user_repo.go         # PostgreSQL user repo
│   │       ├── student_repo.go      # PostgreSQL student repo
│   │       ├── invoice_repo.go      # PostgreSQL invoice repo
│   │       ├── transaction_repo.go  # PostgreSQL transaction repo
│   │       └── webhook_log_repo.go  # PostgreSQL webhook log repo
│   ├── router/
│   │   └── router.go                # All route definitions
│   └── sqlc/
│       ├── db.go                    # Generated: DB interface
│       ├── models.go                # Generated: Go structs
│       ├── querier.go               # Generated: Query interface
│       ├── institution.sql.go       # Generated: Institution queries
│       ├── user.sql.go              # Generated: User queries
│       ├── student.sql.go           # Generated: Student queries
│       ├── invoice.sql.go           # Generated: Invoice queries
│       ├── transaction.sql.go       # Generated: Transaction queries
│       └── webhook_log.sql.go       # Generated: Webhook log queries
├── pkg/
│   ├── hmac/
│   │   └── hmac.go                  # HMAC-SHA256 utilities
│   ├── jwt/
│   │   └── jwt.go                   # JWT token utilities
│   ├── password/
│   │   └── password.go              # Bcrypt password utilities
│   ├── validator/
│   │   └── validator.go             # Custom validation helpers
│   └── response/
│       └── response.go              # Standard API response helpers
├── migrations/
│   ├── 000001_create_institutions.up.sql
│   ├── 000001_create_institutions.down.sql
│   ├── 000002_create_users.up.sql
│   ├── 000002_create_users.down.sql
│   ├── 000003_create_students.up.sql
│   ├── 000003_create_students.down.sql
│   ├── 000004_create_invoices.up.sql
│   ├── 000004_create_invoices.down.sql
│   ├── 000005_create_transaction_histories.up.sql
│   ├── 000005_create_transaction_histories.down.sql
│   ├── 000006_create_webhook_logs.up.sql
│   ├── 000006_create_webhook_logs.down.sql
│   ├── 000007_create_refresh_tokens.up.sql
│   ├── 000007_create_refresh_tokens.down.sql
│   ├── 000008_enable_rls.up.sql
│   ├── 000008_enable_rls.down.sql
│   └── 000009_seed_data.up.sql
├── sql/
│   └── queries/
│       ├── institution.sql          # sqlc query definitions
│       ├── user.sql
│       ├── student.sql
│       ├── invoice.sql
│       ├── transaction.sql
│       └── webhook_log.sql
├── sqlc.yaml                        # sqlc configuration
├── Dockerfile
├── .env.example
├── go.mod
└── go.sum
```

### 2.3 Dependency Rule

```
Domain (entities, interfaces) → knows nothing about outer layers
    ↑
Use Cases → depends only on domain interfaces
    ↑
Handlers/Repos → depends on use cases and domain; implements interfaces
    ↑
Infrastructure → depends on everything; wires dependencies
```

**Key Rule**: Inner layers NEVER import outer layers. Dependencies always point inward.

### 2.4 Dependency Injection

```go
// cmd/server/main.go — composition root
func main() {
    cfg := config.Load()
    db := database.Connect(cfg.DatabaseURL)

    // Repositories (implement domain interfaces)
    institutionRepo := postgres.NewInstitutionRepo(db)
    userRepo := postgres.NewUserRepo(db)
    invoiceRepo := postgres.NewInvoiceRepo(db)
    // ...

    // Use cases (depend on repository interfaces)
    authUseCase := usecase.NewAuthUseCase(userRepo, cfg.JWTSecret)
    invoiceUseCase := usecase.NewInvoiceUseCase(invoiceRepo, studentRepo, transactionRepo)
    webhookUseCase := usecase.NewWebhookUseCase(invoiceRepo, institutionRepo, transactionRepo, webhookLogRepo)
    // ...

    // Handlers (depend on use cases)
    authHandler := handler.NewAuthHandler(authUseCase)
    invoiceHandler := handler.NewInvoiceHandler(invoiceUseCase)
    webhookHandler := handler.NewWebhookHandler(webhookUseCase)
    // ...

    // Router
    r := router.Setup(authHandler, invoiceHandler, webhookHandler, /* ... */)
    r.Run(":" + cfg.Port)
}
```

---

## 3. Clean Architecture — Frontend (React + Vite)

### 3.1 Layer Diagram

```mermaid
graph TB
    subgraph "Layer 4: Presentation / UI"
        Router["React Router v6/v7"]
        Pages["pages/ (Views)"]
        Components["components/ (shadcn/ui + Shared)"]
        Layouts["layouts/ (AuthLayout, DashboardLayout)"]
    end

    subgraph "Layer 3: Application / Use Cases"
        QueryHooks["TanStack Query Hooks"]
        ZustandStores["Zustand Stores (Auth, Tenant)"]
    end

    subgraph "Layer 2: Domain"
        Models["models/ (Entities & DTOs)"]
        RepoInterfaces["repositories/ (Interfaces)"]
    end

    subgraph "Layer 1: Infrastructure / Adapters"
        APIClient["http/ (Axios + JWT Interceptors)"]
        RepoImpls["infrastructure/repositories/"]
    end

    Router --> Layouts
    Layouts --> Pages
    Pages --> Components
    Pages --> QueryHooks
    QueryHooks --> ZustandStores
    QueryHooks --> RepoInterfaces
    RepoImpls --> RepoInterfaces
    RepoImpls --> APIClient
    QueryHooks --> Models

    style Models fill:#E91E63,color:#fff
    style QueryHooks fill:#9C27B0,color:#fff
    style Pages fill:#2196F3,color:#fff
    style APIClient fill:#607D8B,color:#fff
```

### 3.2 Directory Structure

```
sevima-fe/
├── src/
│   ├── core/
│   │   ├── config/                   # Environment & app config
│   │   ├── http/
│   │   │   ├── apiClient.ts          # Axios client with refresh interceptor
│   │   │   └── tokenManager.ts       # Secure local / session storage
│   │   └── router/
│   │       ├── AppRouter.tsx         # Route tree definition
│   │       └── ProtectedRoute.tsx    # Role & tenant authorization guards
│   ├── domain/
│   │   ├── models/                   # Pure entities & interfaces
│   │   │   ├── auth.model.ts
│   │   │   ├── institution.model.ts
│   │   │   ├── invoice.model.ts
│   │   │   ├── student.model.ts
│   │   │   └── transaction.model.ts
│   │   └── repositories/             # Repository interfaces
│   │       ├── IAuthRepository.ts
│   │       ├── IInstitutionRepository.ts
│   │       └── IInvoiceRepository.ts
│   ├── infrastructure/
│   │   └── repositories/             # Repository implementations (Axios)
│   │       ├── AuthRepository.ts
│   │       ├── InstitutionRepository.ts
│   │       └── InvoiceRepository.ts
│   ├── application/
│   │   ├── auth/
│   │   │   ├── useLogin.ts
│   │   │   └── useAuthStore.ts       # Zustand auth store
│   │   ├── institution/
│   │   │   ├── useInstitutions.ts
│   │   │   └── useCreateInstitution.ts
│   │   ├── invoice/
│   │   │   ├── useInvoices.ts
│   │   │   ├── useInvoiceDetail.ts
│   │   │   └── useCreateInvoice.ts
│   │   └── student/
│   │       └── useKrsEligibility.ts
│   └── presentation/
│       ├── components/
│       │   ├── ui/                   # shadcn/ui components (Button, Card, Table, etc.)
│       │   └── shared/
│       │       ├── Navbar.tsx
│       │       ├── Sidebar.tsx
│       │       ├── StatusBadge.tsx
│       │       ├── AmountDisplay.tsx
│       │       ├── KrsEligibilityBanner.tsx
│       │       └── EmptyState.tsx
│       ├── layouts/
│       │   ├── AuthLayout.tsx
│       │   ├── DashboardLayout.tsx
│       │   └── StudentLayout.tsx
│       └── pages/
│           ├── auth/
│           │   └── LoginPage.tsx
│           ├── superadmin/
│           │   └── InstitutionListPage.tsx
│           ├── admin/
│           │   ├── InvoiceListPage.tsx
│           │   ├── CreateInvoicePage.tsx
│           │   └── InvoiceDetailPage.tsx
│           ├── student/
│           │   ├── StudentInvoiceListPage.tsx
│           │   └── KrsStatusPage.tsx
│           └── mockgateway/
│               └── SimulatorPage.tsx # Mock gateway UI
├── public/
│   └── favicon.ico
├── index.html
├── tailwind.config.ts
├── vite.config.ts
├── tsconfig.json
├── Dockerfile
└── package.json
```

---

## 4. API Architecture

### 4.1 REST API Conventions

```
Base URL: /api/v1

Authentication:
  Authorization: Bearer <access_token>

Content-Type: application/json

Response format:
{
    "success": true,
    "data": { ... },
    "meta": { "page": 1, "per_page": 20, "total": 100 },
    "message": "..."
}

Error format:
{
    "success": false,
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "...",
        "details": [
            { "field": "amount", "message": "must be positive" }
        ]
    }
}
```

### 4.2 Route Groups

```
Public:
  POST   /api/v1/auth/login
  POST   /api/v1/auth/refresh
  POST   /api/v1/webhook/payment          # HMAC-verified, not JWT

Authenticated (any role):
  POST   /api/v1/auth/logout
  GET    /api/v1/auth/me

Super Admin:
  GET    /api/v1/institutions
  POST   /api/v1/institutions
  GET    /api/v1/institutions/:id
  PUT    /api/v1/institutions/:id
  DELETE /api/v1/institutions/:id
  POST   /api/v1/institutions/:id/regenerate-secret

Admin Keuangan:
  GET    /api/v1/invoices                  # Filtered by institusi_id
  POST   /api/v1/invoices
  GET    /api/v1/invoices/:id
  GET    /api/v1/invoices/:id/transactions
  GET    /api/v1/students                  # Students in institution
  GET    /api/v1/students/:id
  GET    /api/v1/dashboard/stats           # Institution stats

Mahasiswa:
  GET    /api/v1/student/invoices          # Own invoices only
  GET    /api/v1/student/invoices/:id
  GET    /api/v1/student/invoices/:id/transactions
  GET    /api/v1/student/dashboard         # Own summary

Health:
  GET    /api/v1/health
```

---

## 5. Database Connection Architecture

### 5.1 Connection Pool

```mermaid
graph LR
    subgraph "Go Application"
        H1["Handler 1"]
        H2["Handler 2"]
        H3["Handler 3"]
    end

    subgraph "pgxpool"
        P["Connection Pool<br/>Max: 25, Min: 5"]
        C1["Conn 1"]
        C2["Conn 2"]
        C3["Conn ..."]
    end

    subgraph "PostgreSQL"
        PG[("Database")]
    end

    H1 --> P
    H2 --> P
    H3 --> P
    P --> C1
    P --> C2
    P --> C3
    C1 --> PG
    C2 --> PG
    C3 --> PG
```

### 5.2 Tenant Context Injection

```go
// middleware/tenant.go
func TenantMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.MustGet("claims").(*jwt.Claims)

        // Set PostgreSQL session variables for RLS
        if claims.InstitusiID != uuid.Nil {
            db.Exec("SET app.current_institusi_id = $1", claims.InstitusiID)
        }
        db.Exec("SET app.current_role = $1", claims.Role)

        c.Next()

        // Reset session variables after request
        db.Exec("RESET app.current_institusi_id")
        db.Exec("RESET app.current_role")
    }
}
```

---

## 6. Docker Compose Architecture

```yaml
# docker-compose.yml
services:
  postgres:
    image: postgres:16-alpine
    ports: ["5432:5432"]
    volumes:
      - pgdata:/var/lib/postgresql/data
    environment:
      POSTGRES_DB: sevima_pay
      POSTGRES_USER: sevima
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U sevima"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build: ./sevima-be
    ports: ["8080:8080"]
    depends_on:
      postgres:
        condition: service_healthy
    env_file: .env
    environment:
      DATABASE_URL: postgres://sevima:${DB_PASSWORD}@postgres:5432/sevima_pay?sslmode=disable

  frontend:
    build: ./sevima-fe
    ports: ["5173:5173"]
    depends_on:
      - backend
    environment:
      API_URL: http://backend:8080

  mock-gateway:
    build: ./sevima-mock-gateway
    ports: ["8081:8081"]
    depends_on:
      - backend
    env_file: .env
    environment:
      WEBHOOK_URL: http://backend:8080/api/v1/webhook/payment

volumes:
  pgdata:
```

---

## 7. Security Architecture

### 7.1 Authentication Flow

```mermaid
graph LR
    A["Login"] -->|"email + password"| B["Backend"]
    B -->|"bcrypt verify"| C{"Valid?"}
    C -->|"Yes"| D["Generate JWT"]
    C -->|"No"| E["401"]
    D --> F["Access Token (15min)"]
    D --> G["Refresh Token (7d, hashed in DB)"]
    F --> H["Stored in memory (FE)"]
    G --> I["Stored in httpOnly cookie"]
```

### 7.2 Defense in Depth Layers

```mermaid
graph TB
    A["Request"] --> B["Layer 1: CORS"]
    B --> C["Layer 2: Rate Limiting"]
    C --> D["Layer 3: JWT Validation"]
    D --> E["Layer 4: RBAC Check"]
    E --> F["Layer 5: Tenant Middleware"]
    F --> G["Layer 6: Input Validation"]
    G --> H["Layer 7: Application Logic"]
    H --> I["Layer 8: PostgreSQL RLS"]
    I --> J["Layer 9: DB Constraints"]

    style A fill:#2196F3,color:#fff
    style I fill:#4CAF50,color:#fff
    style J fill:#4CAF50,color:#fff
```

---

## 8. Error Handling Strategy

### 8.1 Error Types

```go
// domain/error.go
type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

var (
    ErrUnauthorized       = &AppError{"UNAUTHORIZED", "Invalid credentials", 401}
    ErrForbidden          = &AppError{"FORBIDDEN", "Insufficient permissions", 403}
    ErrNotFound           = &AppError{"NOT_FOUND", "Resource not found", 404}
    ErrConflict           = &AppError{"CONFLICT", "Resource already exists", 409}
    ErrValidation         = &AppError{"VALIDATION_ERROR", "Invalid input", 422}
    ErrInvalidSignature   = &AppError{"INVALID_SIGNATURE", "Webhook signature mismatch", 401}
    ErrInvalidTransition  = &AppError{"INVALID_TRANSITION", "Invalid status transition", 422}
    ErrTenantViolation    = &AppError{"TENANT_VIOLATION", "Cross-tenant access denied", 403}
)
```

### 8.2 Error Flow

```
Handler → catches error → maps to AppError → JSON response
UseCase → returns domain error → Handler translates
Repository → returns wrapped DB error → UseCase handles
```

---

## 9. Logging Strategy

### 9.1 Structured Logging

```go
// Using slog (Go 1.21+)
slog.Info("webhook processed",
    "institusi_id", institusiID,
    "invoice_number", invoiceNumber,
    "event_type", eventType,
    "processing_status", "success",
    "duration_ms", duration.Milliseconds(),
)
```

### 9.2 Log Levels

| Level | Use Case |
|-------|----------|
| ERROR | Unrecoverable failures, DB errors, panic recovery |
| WARN | Invalid webhook signature, rejected state transition |
| INFO | Request completed, webhook processed, invoice created |
| DEBUG | SQL queries, full request/response bodies (dev only) |

---

## 10. Testing Architecture

### 10.1 Test Pyramid

```
                    ╱╲
                   ╱  ╲
                  ╱ E2E ╲          (Few — critical flows only)
                 ╱________╲
                ╱          ╲
               ╱ Integration ╲     (Medium — DB, API tests)
              ╱________________╲
             ╱                  ╲
            ╱    Unit Tests      ╲  (Many — use cases, domain)
           ╱______________________╲
```

### 10.2 Testing Strategy

| Layer | Type | Tool | What |
|-------|------|------|------|
| Domain | Unit | `testing` | Entity validation, value objects |
| Use Case | Unit | `testing` + `testify/mock` | Business logic with mocked repos |
| Repository | Integration | `testcontainers-go` | SQL queries against real PostgreSQL |
| Handler | Integration | `httptest` + `testcontainers-go` | Full HTTP request/response |
| Webhook | Integration | `httptest` | HMAC verification, state transitions |
| Tenant Isolation | Integration | `testcontainers-go` | Cross-tenant access prevention |
