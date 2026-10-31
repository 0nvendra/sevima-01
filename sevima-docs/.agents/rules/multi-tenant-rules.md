# SEVIMA-PAY Multi-Tenancy & Security Rules for AI

When working on this repository, all AI agents and developers MUST enforce the following constraints:

## 1. Multi-Tenant Isolation
- Never query or mutate tables (`users`, `students`, `invoices`, `transaction_histories`, `webhook_logs`) without an explicit `institusi_id` filter, unless the requesting role is explicitly verified as `RoleSuperAdmin`.
- In HTTP handlers, always resolve the tenant ID from the authenticated JWT context (`middleware.TenantIDKey`), NEVER trust an unverified tenant ID from client query parameters or request body when the user is not a Super Admin.
- When handling student endpoints, verify that the logged-in user ID matches the student record (`student.UserID == claims.UserID`).

## 2. Webhook & Payment Security
- Always verify the `X-Signature` header using HMAC-SHA256 with the institution's specific secret key before processing any payment payload.
- Enforce idempotency: If an invoice is already in `paid` status, return an HTTP 200 success response with a duplicate indicator, and do not execute duplicate mutations.
- Invoices in `expired` status must NEVER transition to `paid`.

## 3. Secret Management & Repository Cleanliness
- Never hardcode or commit database credentials, JWT secrets, private keys, or HMAC secrets into Git.
- Always use environment variables defined via `.env` and documented in `.env.example`.
