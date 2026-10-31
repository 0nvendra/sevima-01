# Flow Diagrams
# SEVIMA-PAY: Multi-Tenant UKT Payment System

> **Version**: 1.0.0
> Semua diagram menggunakan Mermaid syntax

---

## 1. Invoice Lifecycle State Machine

```mermaid
stateDiagram-v2
    [*] --> unpaid : Invoice created

    unpaid --> paid : payment.success webhook
    unpaid --> expired : payment.expired webhook
    unpaid --> expired : Scheduled job (due_date passed)

    paid --> paid : Duplicate payment.success (no-op)

    expired --> expired : Duplicate payment.expired (no-op)
    expired --> paid : Late payment.success (configurable)

    paid --> [*] : Final state (KRS eligible)
    expired --> [*] : Final state (needs new invoice)

    note right of paid
        paid_at dan paid_amount
        diisi dari gateway notification
    end note

    note right of expired
        Bisa dari webhook
        atau scheduled job
    end note
```

---

## 2. Webhook Processing Flowchart

```mermaid
flowchart TD
    A[Receive POST /api/v1/webhook/payment] --> B{Parse JSON body}
    B -->|Parse error| B1[Log error, return 400]
    B -->|OK| C[Extract institution_code from body]

    C --> D{Lookup institution by code}
    D -->|Not found| D1[Log webhook as failed<br/>Return 200 with error message]
    D -->|Found| E[Get HMAC secret]

    E --> F{Verify HMAC-SHA256 signature}
    F -->|Invalid| F1[Log as invalid_signature<br/>Return 401 Unauthorized]
    F -->|Valid| G[Extract event_type & invoice_number]

    G --> H{Lookup invoice by<br/>institusi_id + invoice_number}
    H -->|Not found| H1[Log as invoice_not_found<br/>Return 200 with error message]
    H -->|Found| I{Check event_type}

    I -->|payment.success| J{Current status?}
    J -->|unpaid| J1[Update to paid<br/>Set paid_at, paid_amount]
    J -->|paid| J2[Idempotent: no-op<br/>Log as duplicate]
    J -->|expired| J3{Allow late payment?}
    J3 -->|Yes| J4[Update to paid<br/>Log as late payment]
    J3 -->|No| J5[Reject<br/>Log as invalid_transition]

    I -->|payment.expired| K{Current status?}
    K -->|unpaid| K1[Update to expired]
    K -->|paid| K2[Reject: cannot expire paid<br/>Log as invalid_transition]
    K -->|expired| K3[Idempotent: no-op<br/>Log as duplicate]

    J1 --> L[Insert transaction_history]
    J4 --> L
    K1 --> L
    L --> M[Insert webhook_log as success]
    J2 --> M
    K3 --> M
    M --> N[Return 200 OK success]

    J5 --> O[Insert webhook_log as invalid_transition]
    K2 --> O
    O --> P[Return 200 OK with error]

    style A fill:#4CAF50,color:#fff
    style F1 fill:#f44336,color:#fff
    style N fill:#4CAF50,color:#fff
    style B1 fill:#f44336,color:#fff
```

---

## 3. Authentication & Authorization Flow

```mermaid
flowchart TD
    A[HTTP Request] --> B{Has Authorization header?}
    B -->|No| B1{Is public route?}
    B1 -->|Yes| B2[Process request]
    B1 -->|No| B3[401 Unauthorized]

    B -->|Yes| C[Extract Bearer token]
    C --> D{Valid JWT?}
    D -->|No| D1[401 Invalid token]
    D -->|Expired| D2[401 Token expired]
    D -->|Valid| E[Extract claims]

    E --> F{Route requires specific role?}
    F -->|No| G[Process request with user context]
    F -->|Yes| H{User role in allowed roles?}
    H -->|No| H1[403 Forbidden]
    H -->|Yes| I{Route is tenant-scoped?}

    I -->|No| G
    I -->|Yes| J{User has institusi_id?}
    J -->|No, super_admin| K[Process with full access]
    J -->|Yes| L[Set tenant context in DB connection]
    L --> M[Process request with tenant isolation]

    style A fill:#2196F3,color:#fff
    style B2 fill:#4CAF50,color:#fff
    style G fill:#4CAF50,color:#fff
    style K fill:#4CAF50,color:#fff
    style M fill:#4CAF50,color:#fff
    style B3 fill:#f44336,color:#fff
    style D1 fill:#f44336,color:#fff
    style D2 fill:#FF9800,color:#fff
    style H1 fill:#f44336,color:#fff
```

---

## 4. Invoice Creation Flow

```mermaid
flowchart TD
    A[Admin submits invoice form] --> B{Validate input}
    B -->|Invalid| B1[Return 422 validation errors]
    B -->|Valid| C[Extract institusi_id from JWT]

    C --> D{Student exists in institution?}
    D -->|No| D1[Return 404 Student not found]
    D -->|Yes| E{Invoice number unique in institution?}

    E -->|No| E1[Return 409 Conflict: duplicate invoice number]
    E -->|Yes| F[BEGIN TRANSACTION]

    F --> G[INSERT invoice with status=unpaid]
    G --> H[INSERT transaction_history<br/>NULL → unpaid, source=manual]
    H --> I[COMMIT]
    I --> J[Return 201 Created]

    style A fill:#2196F3,color:#fff
    style J fill:#4CAF50,color:#fff
    style B1 fill:#f44336,color:#fff
    style D1 fill:#f44336,color:#fff
    style E1 fill:#FF9800,color:#fff
```

---

## 5. Tenant Data Access Control Flow

```mermaid
flowchart TD
    A[API Request with JWT] --> B[Middleware extracts institusi_id & role]

    B --> C{Role?}

    C -->|super_admin| D[Full access to all institutions]
    D --> D1[Set app.current_role = super_admin]
    D1 --> D2[RLS allows all rows]

    C -->|admin_keuangan| E[Scoped to own institution]
    E --> E1["Set app.current_institusi_id = user.institusi_id"]
    E1 --> E2[RLS filters by institusi_id]
    E2 --> E3[Can see all invoices in institution]

    C -->|mahasiswa| F[Scoped to own institution + own data]
    F --> F1["Set app.current_institusi_id = user.institusi_id"]
    F1 --> F2[RLS filters by institusi_id]
    F2 --> F3[App-level filter: student_id = own]
    F3 --> F4[Can only see own invoices]

    subgraph "Defense in Depth"
        G1[Layer 1: JWT claims validation]
        G2[Layer 2: RBAC middleware role check]
        G3[Layer 3: Application query filtering]
        G4[Layer 4: PostgreSQL RLS policies]
    end

    style D fill:#4CAF50,color:#fff
    style E3 fill:#2196F3,color:#fff
    style F4 fill:#FF9800,color:#fff
```

---

## 6. Mock Payment Gateway Flow

```mermaid
flowchart TD
    A[Tester opens Mock Gateway UI] --> B[Select institution]
    B --> C[Select invoice from dropdown]
    C --> D[Select event type]
    D --> E{Event type?}

    E -->|payment.success| F[Fill payment details<br/>amount, method]
    E -->|payment.expired| G[Auto-fill with invoice data]

    F --> H[Build webhook payload]
    G --> H

    H --> I[Lookup institution HMAC secret]
    I --> J["Compute signature = HMAC-SHA256(JSON payload, secret)"]
    J --> K[Send POST to SEVIMA-PAY webhook endpoint]

    K --> L{Response status?}
    L -->|200 Success| M[Show success result]
    L -->|401 Unauthorized| N[Show signature error]
    L -->|Other| O[Show error details]

    M --> P[Display response body]
    N --> P
    O --> P

    style A fill:#9C27B0,color:#fff
    style M fill:#4CAF50,color:#fff
    style N fill:#f44336,color:#fff
    style O fill:#FF9800,color:#fff
```

---

## 7. System Deployment Architecture

```mermaid
flowchart LR
    subgraph Docker Compose
        subgraph Frontend
            FE[React SPA<br/>Port 5173]
        end

        subgraph Backend
            BE[Go Gin API<br/>Port 8080]
        end

        subgraph MockGateway
            MG[Mock Gateway<br/>Port 8081]
        end

        subgraph Database
            PG[(PostgreSQL 16<br/>Port 5432)]
        end
    end

    Browser -->|HTTP| FE
    FE -->|API calls| BE
    MG -->|Webhook POST| BE
    BE -->|SQL + RLS| PG
    MG -->|Read config| PG

    style FE fill:#FF5722,color:#fff
    style BE fill:#2196F3,color:#fff
    style MG fill:#9C27B0,color:#fff
    style PG fill:#4CAF50,color:#fff
```

---

## 8. Scheduled Expiry Job Flow

```mermaid
flowchart TD
    A[Cron triggers every hour] --> B[Call expire_overdue_invoices function]
    B --> C[SELECT invoices WHERE status=unpaid AND due_date < NOW]

    C --> D{Any overdue invoices?}
    D -->|No| E[Log: no overdue invoices]
    D -->|Yes| F[BEGIN TRANSACTION]

    F --> G["UPDATE invoices SET status='expired'"]
    G --> H[INSERT transaction_histories for each]
    H --> I[COMMIT]
    I --> J["Log: expired {count} invoices"]

    style A fill:#9C27B0,color:#fff
    style E fill:#607D8B,color:#fff
    style J fill:#4CAF50,color:#fff
```

---

## 9. Data Flow Overview

```mermaid
flowchart LR
    subgraph External
        GW[Payment Gateway]
    end

    subgraph "SEVIMA-PAY Backend"
        WH[Webhook Handler]
        AUTH[Auth Service]
        INV[Invoice Service]
        TXN[Transaction Service]
        MW[RBAC + Tenant Middleware]
    end

    subgraph "SEVIMA-PAY Frontend"
        LOGIN[Login Page]
        ADMIN[Admin Dashboard]
        STUDENT[Student Dashboard]
    end

    subgraph Database
        PG[(PostgreSQL)]
    end

    GW -->|webhook notification| WH
    WH -->|verify & process| INV
    WH -->|log| TXN
    INV -->|read/write| PG
    TXN -->|append| PG
    AUTH -->|verify/issue JWT| PG

    LOGIN -->|credentials| AUTH
    ADMIN -->|CRUD invoices| MW --> INV
    STUDENT -->|view invoices| MW --> INV

    style GW fill:#FF9800,color:#fff
    style PG fill:#4CAF50,color:#fff
```
