---
name: tenant-isolation-verifier
description: Verify multi-tenant data segregation, RBAC role enforcement, and webhook HMAC signature validity in SEVIMA-PAY.
---

# Tenant Isolation & Webhook Security Verifier

Use this skill when auditing, testing, or reviewing code changes in SEVIMA-PAY to ensure multi-tenant security guarantees are not breached.

## Verification Checklist

1. **Handler & Route Check**:
   - Ensure routes under `/api/v1` that handle tenant resources have `middleware.TenantMiddleware()` attached.
   - Verify that role guards (`middleware.RequireRole(...)`) accurately restrict access (e.g., student cannot trigger administrative invoice generation).

2. **Repository & Query Check**:
   - Review SQL queries in `internal/adapter/repository` to confirm all queries include `WHERE institusi_id = $...`.
   - Ensure unique constraints (e.g. `(institusi_id, invoice_number)`) prevent cross-tenant collisions.

3. **Webhook Verification**:
   - Verify `crypto.VerifyHMAC(rawBody, signature, inst.HMACSecret)` is executed before database transaction begins.
   - Verify transaction locking (`SELECT ... FOR UPDATE`) prevents concurrent payment race conditions.
   - Verify that duplicate webhook notifications return idempotent no-op responses.

4. **Automated Test Execution**:
   - Run:
     ```bash
     cd sevima-be && go test -v ./...
     ```
   - Ensure all tenant isolation tests in `internal/adapter/middleware` and `internal/domain/entity` pass.
