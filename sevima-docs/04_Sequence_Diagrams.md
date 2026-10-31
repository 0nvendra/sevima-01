# Sequence Diagrams
# SEVIMA-PAY: Multi-Tenant UKT Payment System

> **Version**: 1.0.0
> Semua diagram menggunakan Mermaid syntax

---

## 1. Authentication Flow

### 1.1 Login

```mermaid
sequenceDiagram
    actor User
    participant FE as React Frontend
    participant BE as Go Backend
    participant DB as PostgreSQL

    User->>FE: Submit email + password
    FE->>BE: POST /api/v1/auth/login
    BE->>DB: SELECT user by email
    DB-->>BE: User record (with password_hash)

    alt Invalid credentials
        BE-->>FE: 401 Unauthorized
        FE-->>User: Show error message
    else Valid credentials
        BE->>BE: Verify bcrypt hash
        BE->>BE: Generate JWT access token (15min)
        BE->>BE: Generate refresh token (7d)
        BE->>DB: INSERT refresh_token (hashed)
        DB-->>BE: OK
        BE-->>FE: 200 {access_token, refresh_token, user}
        FE->>FE: Store tokens, redirect by role
        FE-->>User: Redirect to dashboard
    end
```

### 1.2 Token Refresh

```mermaid
sequenceDiagram
    participant FE as React Frontend
    participant BE as Go Backend
    participant DB as PostgreSQL

    FE->>BE: POST /api/v1/auth/refresh {refresh_token}
    BE->>BE: Hash incoming refresh_token
    BE->>DB: SELECT refresh_token WHERE hash = ? AND NOT revoked AND NOT expired
    DB-->>BE: Token record

    alt Invalid/expired/revoked token
        BE-->>FE: 401 Unauthorized
        FE->>FE: Redirect to login
    else Valid token
        BE->>DB: UPDATE revoke old refresh_token
        BE->>BE: Generate new access_token + refresh_token
        BE->>DB: INSERT new refresh_token (hashed)
        BE-->>FE: 200 {access_token, refresh_token}
        FE->>FE: Update stored tokens
    end
```

### 1.3 Logout

```mermaid
sequenceDiagram
    participant FE as React Frontend
    participant BE as Go Backend
    participant DB as PostgreSQL

    FE->>BE: POST /api/v1/auth/logout {refresh_token}
    Note right of BE: Authorization: Bearer <access_token>
    BE->>BE: Hash incoming refresh_token
    BE->>DB: UPDATE refresh_tokens SET revoked_at = NOW() WHERE hash = ?
    DB-->>BE: OK
    BE-->>FE: 200 OK
    FE->>FE: Clear stored tokens
    FE->>FE: Redirect to login
```

---

## 2. Invoice Management Flow

### 2.1 Create Invoice (Admin Keuangan)

```mermaid
sequenceDiagram
    actor Admin as Admin Keuangan
    participant FE as React Frontend
    participant MW as RBAC Middleware
    participant BE as Go Backend
    participant DB as PostgreSQL

    Admin->>FE: Fill invoice form
    FE->>BE: POST /api/v1/invoices
    Note right of FE: Authorization: Bearer <access_token>

    BE->>MW: Validate JWT + Role check
    MW->>MW: Extract institusi_id from token
    MW->>MW: Verify role = admin_keuangan

    alt Unauthorized
        MW-->>FE: 403 Forbidden
    else Authorized
        MW->>BE: Pass request with institusi_id

        BE->>DB: Verify student belongs to institution
        DB-->>BE: Student record

        alt Student not in institution
            BE-->>FE: 404 Student not found
        else Student valid
            BE->>DB: BEGIN TRANSACTION
            BE->>DB: Check invoice_number unique in institution
            DB-->>BE: No conflict

            alt Duplicate invoice number
                BE->>DB: ROLLBACK
                BE-->>FE: 409 Conflict
            else Unique
                BE->>DB: INSERT invoice (institusi_id from token)
                DB-->>BE: Invoice created
                BE->>DB: INSERT transaction_history (created → unpaid)
                DB-->>BE: History recorded
                BE->>DB: COMMIT
                BE-->>FE: 201 Created {invoice}
                FE-->>Admin: Show success, update list
            end
        end
    end
```

### 2.2 List Invoices with Filters (Admin Keuangan)

```mermaid
sequenceDiagram
    actor Admin as Admin Keuangan
    participant FE as React Frontend
    participant MW as RBAC + Tenant MW
    participant BE as Go Backend
    participant DB as PostgreSQL

    Admin->>FE: Apply filters (status, semester, search)
    FE->>BE: GET /api/v1/invoices?status=unpaid&semester=ganjil&page=1
    BE->>MW: Validate JWT + Extract institusi_id
    MW->>BE: Request with institusi_id context

    BE->>DB: SELECT invoices WHERE institusi_id = $1 AND filters... LIMIT/OFFSET
    Note right of DB: RLS also enforces tenant isolation
    DB-->>BE: Paginated invoice list

    BE->>DB: SELECT COUNT(*) for pagination metadata
    DB-->>BE: Total count

    BE-->>FE: 200 {data: [...], meta: {page, total, ...}}
    FE-->>Admin: Render table with pagination
```

---

## 3. Webhook Payment Flow

### 3.1 Payment Success (Happy Path)

```mermaid
sequenceDiagram
    participant GW as Payment Gateway
    participant WH as Webhook Endpoint
    participant BE as Go Backend
    participant DB as PostgreSQL

    GW->>WH: POST /api/v1/webhook/payment
    Note right of GW: X-Signature: HMAC-SHA256(body, secret)
    Note right of GW: Body: {event, institution_code, invoice_number, amount, ...}

    WH->>WH: Read raw body for signature verification
    WH->>WH: Parse JSON payload → extract institution_code

    WH->>DB: SELECT institution WHERE code = institution_code
    DB-->>WH: Institution record (with hmac_secret)

    alt Institution not found
        WH->>DB: INSERT webhook_log (status: failed, error: institution_not_found)
        WH-->>GW: 200 OK {status: "error", message: "unknown institution"}
    else Institution found
        WH->>WH: Compute HMAC-SHA256(raw_body, secret)
        WH->>WH: Constant-time compare with X-Signature

        alt Signature mismatch
            WH->>DB: INSERT webhook_log (status: invalid_signature)
            WH-->>GW: 401 Unauthorized
        else Signature valid
            WH->>DB: BEGIN TRANSACTION

            WH->>DB: SELECT invoice WHERE institusi_id = ? AND invoice_number = ? FOR UPDATE
            DB-->>WH: Invoice record (with row lock)

            alt Invoice not found
                WH->>DB: INSERT webhook_log (status: invoice_not_found)
                WH->>DB: COMMIT
                WH-->>GW: 200 OK {status: "error", message: "invoice not found"}
            else Invoice found & status = unpaid
                WH->>DB: UPDATE invoice SET status = 'paid', paid_at = NOW(), paid_amount = ?
                WH->>DB: INSERT transaction_history (unpaid → paid, source: webhook)
                WH->>DB: INSERT webhook_log (status: success)
                WH->>DB: COMMIT
                WH-->>GW: 200 OK {status: "success"}
            else Invoice found & status = paid (idempotent)
                WH->>DB: INSERT webhook_log (status: success, notes: "duplicate - no-op")
                WH->>DB: COMMIT
                WH-->>GW: 200 OK {status: "success", message: "already paid"}
            end
        end
    end
```

### 3.2 Payment Expired

```mermaid
sequenceDiagram
    participant GW as Payment Gateway
    participant WH as Webhook Endpoint
    participant DB as PostgreSQL

    GW->>WH: POST /api/v1/webhook/payment
    Note right of GW: event: "payment.expired"

    WH->>WH: Verify HMAC signature (same as success flow)

    WH->>DB: BEGIN TRANSACTION
    WH->>DB: SELECT invoice FOR UPDATE
    DB-->>WH: Invoice record

    alt Status = unpaid
        WH->>DB: UPDATE invoice SET status = 'expired'
        WH->>DB: INSERT transaction_history (unpaid → expired)
        WH->>DB: INSERT webhook_log (status: success)
        WH->>DB: COMMIT
        WH-->>GW: 200 OK {status: "success"}
    else Status = paid
        WH->>DB: INSERT webhook_log (status: invalid_transition)
        Note right of WH: Cannot expire a paid invoice
        WH->>DB: COMMIT
        WH-->>GW: 200 OK {status: "error", message: "invoice already paid"}
    else Status = expired (idempotent)
        WH->>DB: INSERT webhook_log (status: success, notes: "already expired")
        WH->>DB: COMMIT
        WH-->>GW: 200 OK {status: "success"}
    end
```

### 3.3 Concurrent Webhook Processing

```mermaid
sequenceDiagram
    participant GW1 as Gateway (Request 1)
    participant GW2 as Gateway (Request 2)
    participant WH as Webhook Endpoint
    participant DB as PostgreSQL

    Note over GW1,DB: Race condition: two notifications for same invoice

    GW1->>WH: payment.success for INV-001
    GW2->>WH: payment.success for INV-001 (duplicate)

    WH->>DB: BEGIN TX1
    WH->>DB: SELECT invoice FOR UPDATE (TX1 acquires lock)
    DB-->>WH: Invoice (unpaid)

    WH->>DB: BEGIN TX2
    WH->>DB: SELECT invoice FOR UPDATE (TX2 waits for lock)
    Note right of DB: TX2 blocked until TX1 commits

    WH->>DB: UPDATE invoice → paid (TX1)
    WH->>DB: INSERT history + log (TX1)
    WH->>DB: COMMIT TX1 (releases lock)
    WH-->>GW1: 200 OK {status: "success"}

    DB-->>WH: Invoice (now paid, TX2 proceeds)
    Note right of WH: Status already paid → idempotent no-op
    WH->>DB: INSERT webhook_log (duplicate)
    WH->>DB: COMMIT TX2
    WH-->>GW2: 200 OK {status: "success", message: "already paid"}
```

---

## 4. Student View Flow

### 4.1 View Invoices & Transaction History

```mermaid
sequenceDiagram
    actor Student as Mahasiswa
    participant FE as React Frontend
    participant MW as RBAC + Tenant MW
    participant BE as Go Backend
    participant DB as PostgreSQL

    Student->>FE: Open dashboard
    FE->>BE: GET /api/v1/student/invoices
    Note right of FE: Authorization: Bearer <access_token>

    BE->>MW: Validate JWT
    MW->>MW: Extract user_id, institusi_id, role=mahasiswa
    MW->>BE: Pass context

    BE->>DB: SELECT student WHERE user_id = $1 AND institusi_id = $2
    DB-->>BE: Student record

    BE->>DB: SELECT invoices WHERE student_id = $1 AND institusi_id = $2
    DB-->>BE: Invoice list

    BE-->>FE: 200 {invoices: [...], summary: {total, paid, unpaid}}
    FE-->>Student: Render invoice list + summary cards

    Student->>FE: Click invoice detail
    FE->>BE: GET /api/v1/student/invoices/:id/transactions
    BE->>MW: Validate JWT + ownership check
    MW->>BE: Authorized

    BE->>DB: SELECT transaction_histories WHERE invoice_id = $1 AND institusi_id = $2
    Note right of DB: Also verify invoice belongs to this student
    DB-->>BE: Transaction history

    BE-->>FE: 200 {transactions: [...]}
    FE-->>Student: Render transaction timeline
```

---

## 5. Mock Gateway Flow

### 5.1 Trigger Webhook from Mock UI

```mermaid
sequenceDiagram
    actor Tester
    participant MockUI as Mock Gateway UI
    participant MockBE as Mock Gateway Service
    participant WH as SEVIMA-PAY Webhook

    Tester->>MockUI: Select institution, invoice, event type
    MockUI->>MockBE: POST /mock/trigger

    MockBE->>MockBE: Lookup institution HMAC secret
    MockBE->>MockBE: Build webhook payload
    MockBE->>MockBE: Compute HMAC-SHA256(payload, secret)

    MockBE->>WH: POST /api/v1/webhook/payment
    Note right of MockBE: X-Signature: computed_signature
    Note right of MockBE: Content-Type: application/json

    WH-->>MockBE: 200 OK {status: "success"}
    MockBE-->>MockUI: Webhook sent, response: success
    MockUI-->>Tester: Show result (status, response body)
```

---

## 6. Scheduled Invoice Expiry

```mermaid
sequenceDiagram
    participant Cron as Scheduled Job
    participant DB as PostgreSQL

    Note over Cron: Runs every hour (or configurable interval)

    Cron->>DB: SELECT expire_overdue_invoices()
    Note right of DB: Function atomically:<br/>1. Updates unpaid invoices past due_date to expired<br/>2. Inserts transaction_histories for each

    DB-->>Cron: expired_count (e.g., 5)

    alt expired_count > 0
        Cron->>Cron: Log "Expired {count} overdue invoices"
    else No invoices to expire
        Cron->>Cron: Log "No overdue invoices found"
    end
```

---

## 7. RBAC Middleware Pipeline

```mermaid
sequenceDiagram
    participant Client
    participant CORS as CORS Middleware
    participant Auth as Auth Middleware
    participant RBAC as RBAC Middleware
    participant Tenant as Tenant Middleware
    participant Handler as Route Handler
    participant DB as PostgreSQL

    Client->>CORS: HTTP Request
    CORS->>Auth: Pass (if allowed origin)

    Auth->>Auth: Extract Bearer token
    Auth->>Auth: Validate JWT signature + expiry

    alt Invalid/expired token
        Auth-->>Client: 401 Unauthorized
    else Valid token
        Auth->>Auth: Extract claims (user_id, role, institusi_id)
        Auth->>RBAC: Pass with claims in context

        RBAC->>RBAC: Check route requires role
        RBAC->>RBAC: Compare user role vs required roles

        alt Insufficient role
            RBAC-->>Client: 403 Forbidden
        else Role matches
            RBAC->>Tenant: Pass

            Tenant->>DB: SET app.current_institusi_id = claims.institusi_id
            Tenant->>DB: SET app.current_role = claims.role
            Note right of Tenant: RLS policies now active for this connection

            Tenant->>Handler: Pass with tenant context
            Handler->>DB: Execute queries (RLS enforced)
            DB-->>Handler: Results (tenant-filtered)
            Handler-->>Client: Response
        end
    end
```
