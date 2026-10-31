# Mitigation Plan
# SEVIMA-PAY: Multi-Tenant UKT Payment System

> **Version**: 1.0.0
> Risk assessment and mitigation strategies for production readiness

---

## 1. Risk Matrix

| # | Risk | Impact | Likelihood | Severity | Mitigation Status |
|---|------|--------|------------|----------|-------------------|
| R-01 | Cross-tenant data leakage | Critical | Low | Critical | Mitigated |
| R-02 | Webhook signature bypass | Critical | Low | Critical | Mitigated |
| R-03 | Duplicate webhook processing | High | Medium | High | Mitigated |
| R-04 | Race condition on invoice update | High | Medium | High | Mitigated |
| R-05 | Invoice status inconsistency | High | Medium | High | Mitigated |
| R-06 | Expired invoice receives late payment | Medium | Medium | Medium | Configurable |
| R-07 | Database connection pool exhaustion | High | Low | Medium | Mitigated |
| R-08 | JWT token theft | High | Low | Medium | Mitigated |
| R-09 | Webhook endpoint DDoS | Medium | Medium | Medium | Partially mitigated |
| R-10 | Secret exposure in repository | Critical | Low | Critical | Mitigated |
| R-11 | Audit log tampering | High | Low | High | Mitigated |
| R-12 | Stale invoice status (missed expiry) | Medium | Medium | Medium | Mitigated |

---

## 2. Detailed Mitigation Strategies

### R-01: Cross-Tenant Data Leakage

**Threat**: User from Institution A accessing data from Institution B.

**Impact**: Data breach, regulatory violation, loss of trust.

**4-Layer Defense:**

| Layer | Implementation | Bypass Difficulty |
|-------|---------------|-------------------|
| 1. JWT Claims | `institusi_id` embedded in token at login | Requires token forgery |
| 2. RBAC Middleware | Injects `institusi_id` into every request context | Requires middleware bypass |
| 3. Application Query | Every SQL query includes `WHERE institusi_id = $1` via sqlc | Requires code vulnerability |
| 4. PostgreSQL RLS | Row-Level Security policies on all tenant tables | Requires DBA access |

**Testing Strategy:**
```go
// Integration test: verify tenant isolation
func TestTenantIsolation_AdminCannotSeeOtherInstitution(t *testing.T) {
    // Login as admin of Institution A
    // Try to access invoices of Institution B
    // Assert: 404 Not Found or empty result
    // Assert: no Institution B data in response
}
```

**Monitoring:**
- Log every query that returns 0 results after tenant filtering (potential probing)
- Alert on any RLS violation attempt (PostgreSQL logs)

---

### R-02: Webhook Signature Bypass

**Threat**: Attacker sends forged webhook to change invoice status without valid HMAC.

**Impact**: Fraudulent payment confirmations, financial loss.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| HMAC-SHA256 verification | `crypto/hmac` + `hmac.Equal()` for constant-time comparison |
| Per-institution secret | Each institution has unique HMAC secret |
| Raw body preservation | Read body before JSON parsing, verify against raw bytes |
| Signature in header | `X-Signature` header, not in body (prevents tampering) |
| Full audit logging | All attempts (valid and invalid) logged to `webhook_logs` |

**Implementation:**
```go
func verifyHMACSignature(rawBody []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(rawBody)
    expectedMAC := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
```

**Never:**
- Use `==` for signature comparison (timing attack)
- Parse body before verification (TOCTOU)
- Log the actual secret or full signature

---

### R-03: Duplicate Webhook Processing

**Threat**: Gateway retries webhook, causing double state changes or duplicate records.

**Impact**: Incorrect invoice status, duplicate transaction records.

**Mitigations:**

| Strategy | Implementation |
|----------|---------------|
| Idempotent state transitions | Check current status before update; paid→paid = no-op |
| SELECT FOR UPDATE | Row-level lock prevents concurrent modifications |
| Webhook log deduplication | Log all webhooks but skip state change if already processed |
| Transaction wrapping | All state changes in single DB transaction |

**State Machine Idempotency Matrix:**

| Current State | Event | Action | Result |
|---------------|-------|--------|--------|
| unpaid | payment.success | Update to paid | ✅ State changed |
| paid | payment.success | No-op, log | ✅ Idempotent |
| expired | payment.success | Configurable | ⚠️ Needs decision |
| unpaid | payment.expired | Update to expired | ✅ State changed |
| paid | payment.expired | Reject, log | ✅ Protected |
| expired | payment.expired | No-op, log | ✅ Idempotent |

---

### R-04: Race Condition on Invoice Update

**Threat**: Two concurrent webhook notifications for the same invoice.

**Impact**: Lost update, inconsistent state, duplicate transaction records.

**Mitigation:**

```sql
-- Pessimistic locking: SELECT ... FOR UPDATE
-- Within a transaction, locks the row until commit/rollback

BEGIN;
SELECT * FROM invoices
WHERE institusi_id = $1 AND invoice_number = $2
FOR UPDATE;  -- blocks other transactions on this row

-- Process the event...
UPDATE invoices SET status = 'paid' WHERE id = $1;
INSERT INTO transaction_histories (...) VALUES (...);
COMMIT;
-- Lock released, next transaction sees updated state
```

**Why not optimistic locking?**
- Webhook processing is low-contention (same invoice rarely hit simultaneously)
- Pessimistic locking is simpler, guaranteed correct, and the retry overhead is lower than CAS loops

---

### R-05: Invoice Status Inconsistency

**Threat**: Status in `invoices` table doesn't match the actual payment state.

**Impact**: Student incorrectly blocked/allowed from KRS.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| DB constraints | `CHECK` constraint: `paid` → `paid_at IS NOT NULL` |
| Atomic operations | Status update + history insert in one transaction |
| State machine enforcement | Application layer validates state transitions before DB update |
| Audit trail | `transaction_histories` provides verifiable state change log |
| Reconciliation query | Compare invoice status vs. latest transaction_history entry |

**Reconciliation Query (for monitoring):**
```sql
SELECT i.id, i.status, th.new_status
FROM invoices i
JOIN LATERAL (
    SELECT new_status FROM transaction_histories
    WHERE invoice_id = i.id
    ORDER BY created_at DESC
    LIMIT 1
) th ON true
WHERE i.status != th.new_status;
-- Should return 0 rows if system is consistent
```

---

### R-06: Expired Invoice Receives Late Payment

**Threat**: Gateway sends `payment.success` for an invoice already marked `expired`.

**Impact**: Business decision needed — accept late payment or reject?

**Decision**: **Configurable per institution** (default: reject with logging).

| Option | Behavior | When to Use |
|--------|----------|-------------|
| `reject` (default) | Log as `invalid_transition`, keep expired | Strict policy |
| `accept` | Update to paid, log as late payment | Lenient policy |

**Implementation:**
```go
type InstitutionConfig struct {
    AllowLatePayment bool `json:"allow_late_payment"`
}
```

**Regardless of config:**
- All late payment attempts are logged
- Transaction history records the attempt
- Admin can review and manually adjust if needed

---

### R-07: Database Connection Pool Exhaustion

**Threat**: Too many concurrent requests exhaust the connection pool, causing timeouts.

**Impact**: Service degradation, webhook processing failures.

**Mitigations:**

| Control | Configuration |
|---------|--------------|
| Pool size | Max 25 connections (backend), 5 (mock gateway) |
| Idle timeout | 5 minutes |
| Connection lifetime | 1 hour max |
| Acquire timeout | 5 seconds (fail fast) |
| Health check | Ping every 30 seconds |

```go
// pgxpool configuration
poolConfig.MaxConns = 25
poolConfig.MinConns = 5
poolConfig.MaxConnLifetime = time.Hour
poolConfig.MaxConnIdleTime = 5 * time.Minute
poolConfig.HealthCheckPeriod = 30 * time.Second
```

---

### R-08: JWT Token Theft

**Threat**: Access token intercepted and used by attacker.

**Impact**: Unauthorized access to user's data.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| Short-lived access token | 15 minutes expiry |
| Refresh token rotation | New refresh token on each refresh; old one revoked |
| Secure storage | httpOnly cookie / secure memory storage for refresh token (React) |
| Token binding | Include user_id and institusi_id in claims; validate on every request |
| Logout invalidation | Revoke all refresh tokens on logout |

---

### R-09: Webhook Endpoint DDoS

**Threat**: Flood of requests to webhook endpoint.

**Impact**: Service unavailable, missed legitimate webhooks.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| Rate limiting | 100 requests/minute per IP on webhook endpoint |
| Request size limit | Max 1MB body size |
| Timeout | 30 second request timeout |
| IP allowlisting | Optional: restrict to known gateway IPs |
| Circuit breaker | If error rate > 50% for 30s, reject new requests for 60s |

---

### R-10: Secret Exposure in Repository

**Threat**: HMAC secrets, JWT signing keys, or database credentials committed to git.

**Impact**: Complete security compromise.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| `.gitignore` | Include `.env`, `*.key`, `secrets/` |
| `.env.example` | Template without real values |
| Environment variables | All secrets loaded from env vars |
| Docker secrets | Use `env_file` in Docker Compose |
| Pre-commit hook | Scan for patterns like `HMAC_SECRET=`, `password=` |
| HMAC secrets via email | Sent separately per assessment requirement |

---

### R-11: Audit Log Tampering

**Threat**: Modifying or deleting transaction history records.

**Impact**: Loss of audit trail, inability to reconcile disputes.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| DB trigger | `BEFORE UPDATE/DELETE` trigger raises exception on `transaction_histories` |
| App-level enforcement | No UPDATE/DELETE repository methods for transaction histories |
| RLS policy | No DELETE permission in RLS policy |
| Append-only design | Table has no `updated_at` column — immutable by design |

---

### R-12: Stale Invoice Status (Missed Expiry)

**Threat**: Unpaid invoices past due date remain as `unpaid` instead of `expired`.

**Impact**: Students appear to have valid (payable) invoices when they shouldn't.

**Mitigations:**

| Control | Implementation |
|---------|---------------|
| Scheduled job | `expire_overdue_invoices()` runs every hour |
| Database function | Atomic batch update with transaction history logging |
| Application-level check | API also checks `due_date < now()` when returning invoice status |
| Webhook fallback | `payment.expired` event from gateway also triggers expiry |

---

## 3. Incident Response Plan

### 3.1 Severity Levels

| Level | Definition | Response Time | Example |
|-------|-----------|---------------|---------|
| P0 | Data breach / cross-tenant leak | Immediate | R-01 violation |
| P1 | Payment status corruption | < 1 hour | R-05 inconsistency |
| P2 | Service degradation | < 4 hours | R-07 pool exhaustion |
| P3 | Minor issue | Next business day | UI rendering bug |

### 3.2 Response Procedures

**P0 - Cross-Tenant Data Leak:**
1. Immediately disable affected user accounts
2. Rotate all HMAC secrets for affected institutions
3. Audit `webhook_logs` and `transaction_histories` for scope
4. Notify affected institutions
5. Root cause analysis → fix → deploy → verify

**P1 - Payment Status Corruption:**
1. Run reconciliation query (R-05) to identify affected invoices
2. Cross-reference with `transaction_histories` for ground truth
3. Correct invoice statuses based on audit trail
4. Investigate root cause (webhook processing, race condition)

---

## 4. Monitoring Checklist

| Metric | Alert Threshold | Tool |
|--------|----------------|------|
| Webhook success rate | < 95% over 5 min | Structured logs |
| Invalid signature attempts | > 10 per minute | Log aggregation |
| API error rate (5xx) | > 1% over 5 min | Health check endpoint |
| DB connection pool utilization | > 80% | Connection pool metrics |
| Invoice reconciliation mismatch | > 0 rows | Scheduled query |
| Overdue unpaid invoices | Any after expiry job run | Scheduled check |
