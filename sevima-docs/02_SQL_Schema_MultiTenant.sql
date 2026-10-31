-- ============================================================================
-- SEVIMA-PAY: Multi-Tenant UKT Payment System
-- PostgreSQL Schema with Row-Level Security
-- Version: 1.0.0
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

CREATE TYPE user_role AS ENUM ('super_admin', 'admin_keuangan', 'mahasiswa');
CREATE TYPE invoice_status AS ENUM ('unpaid', 'paid', 'expired');
CREATE TYPE webhook_event_type AS ENUM ('payment.success', 'payment.expired');
CREATE TYPE webhook_log_status AS ENUM ('success', 'failed', 'invalid_signature', 'invoice_not_found', 'invalid_transition');
CREATE TYPE semester_type AS ENUM ('ganjil', 'genap');

-- ============================================================================
-- TABLE: institutions
-- Root entity for multi-tenant isolation
-- ============================================================================

CREATE TABLE institutions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            VARCHAR(20) NOT NULL,
    name            VARCHAR(255) NOT NULL,
    address         TEXT,
    phone           VARCHAR(20),
    email           VARCHAR(255),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    hmac_secret     VARCHAR(512) NOT NULL, -- encrypted at app level, stored hashed
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_institutions_code UNIQUE (code)
);

CREATE INDEX idx_institutions_code ON institutions (code);
CREATE INDEX idx_institutions_is_active ON institutions (is_active);

COMMENT ON TABLE institutions IS 'Root tenant entity. Each institution represents a campus/university.';
COMMENT ON COLUMN institutions.code IS 'Unique institution code used in webhook payload for identification.';
COMMENT ON COLUMN institutions.hmac_secret IS 'HMAC-SHA256 secret for webhook signature verification. Stored encrypted.';

-- ============================================================================
-- TABLE: users
-- Authentication and authorization for all roles
-- ============================================================================

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    institusi_id    UUID REFERENCES institutions(id) ON DELETE CASCADE,
    email           VARCHAR(255) NOT NULL,
    password_hash   VARCHAR(255) NOT NULL, -- bcrypt cost 12
    full_name       VARCHAR(255) NOT NULL,
    role            user_role NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_email UNIQUE (email),
    CONSTRAINT chk_users_super_admin_no_institusi
        CHECK (
            (role = 'super_admin' AND institusi_id IS NULL) OR
            (role != 'super_admin' AND institusi_id IS NOT NULL)
        )
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_institusi_id ON users (institusi_id);
CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_institusi_role ON users (institusi_id, role);

COMMENT ON TABLE users IS 'All system users. Super admin has NULL institusi_id; others are scoped to an institution.';
COMMENT ON COLUMN users.institusi_id IS 'NULL for super_admin, required for admin_keuangan and mahasiswa.';
COMMENT ON CONSTRAINT chk_users_super_admin_no_institusi ON users IS 'Super admins must not belong to any institution; other roles must.';

-- ============================================================================
-- TABLE: students
-- Student-specific data, linked to a user account
-- ============================================================================

CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institusi_id    UUID NOT NULL REFERENCES institutions(id) ON DELETE CASCADE,
    nim             VARCHAR(20) NOT NULL, -- Nomor Induk Mahasiswa
    study_program   VARCHAR(255) NOT NULL,
    enrollment_year INTEGER NOT NULL,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_students_nim_institusi UNIQUE (institusi_id, nim),
    CONSTRAINT uq_students_user_id UNIQUE (user_id)
);

CREATE INDEX idx_students_institusi_id ON students (institusi_id);
CREATE INDEX idx_students_nim ON students (nim);
CREATE INDEX idx_students_user_id ON students (user_id);

COMMENT ON TABLE students IS 'Student profile data. NIM is unique within an institution.';
COMMENT ON COLUMN students.nim IS 'Nomor Induk Mahasiswa - unique student ID within an institution.';

-- ============================================================================
-- TABLE: invoices
-- UKT payment invoices per student per semester
-- ============================================================================

CREATE TABLE invoices (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    institusi_id    UUID NOT NULL REFERENCES institutions(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    invoice_number  VARCHAR(50) NOT NULL,
    amount          BIGINT NOT NULL, -- stored in smallest currency unit (Rupiah, no decimals)
    status          invoice_status NOT NULL DEFAULT 'unpaid',
    semester        semester_type NOT NULL,
    academic_year   VARCHAR(9) NOT NULL, -- format: 2026/2027
    due_date        DATE NOT NULL,
    paid_at         TIMESTAMPTZ,
    paid_amount     BIGINT, -- actual amount paid (from gateway notification)
    payment_method  VARCHAR(100), -- from gateway notification
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_invoices_number_institusi UNIQUE (institusi_id, invoice_number),
    CONSTRAINT chk_invoices_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_invoices_paid_amount_positive CHECK (paid_amount IS NULL OR paid_amount > 0),
    CONSTRAINT chk_invoices_academic_year_format CHECK (academic_year ~ '^\d{4}/\d{4}$'),
    CONSTRAINT chk_invoices_paid_consistency
        CHECK (
            (status = 'paid' AND paid_at IS NOT NULL) OR
            (status != 'paid')
        )
);

CREATE INDEX idx_invoices_institusi_id ON invoices (institusi_id);
CREATE INDEX idx_invoices_student_id ON invoices (student_id);
CREATE INDEX idx_invoices_status ON invoices (status);
CREATE INDEX idx_invoices_due_date ON invoices (due_date);
CREATE INDEX idx_invoices_semester_year ON invoices (semester, academic_year);
CREATE INDEX idx_invoices_institusi_status ON invoices (institusi_id, status);
CREATE INDEX idx_invoices_institusi_student ON invoices (institusi_id, student_id);
CREATE INDEX idx_invoices_unpaid_due ON invoices (due_date) WHERE status = 'unpaid';

COMMENT ON TABLE invoices IS 'UKT payment invoices. Invoice number is unique within an institution.';
COMMENT ON COLUMN invoices.amount IS 'Invoice amount in Rupiah (integer, no decimals).';
COMMENT ON COLUMN invoices.paid_amount IS 'Actual amount paid as reported by payment gateway.';
COMMENT ON COLUMN invoices.invoice_number IS 'Unique within institusi_id, may repeat across institutions.';
COMMENT ON INDEX idx_invoices_unpaid_due IS 'Partial index for efficient expiry checking of unpaid invoices.';

-- ============================================================================
-- TABLE: transaction_histories
-- Append-only audit trail for invoice state changes
-- ============================================================================

CREATE TABLE transaction_histories (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id      UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    institusi_id    UUID NOT NULL REFERENCES institutions(id) ON DELETE CASCADE,
    previous_status invoice_status,
    new_status      invoice_status NOT NULL,
    amount          BIGINT, -- amount involved in this transaction
    source          VARCHAR(50) NOT NULL, -- 'webhook', 'manual', 'system_expiry'
    source_ref      VARCHAR(255), -- reference to webhook_log_id or admin user_id
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_txn_status_changed
        CHECK (previous_status IS NULL OR previous_status != new_status)
);

-- No UPDATE or DELETE should be allowed on this table (enforced at app level + trigger)
CREATE INDEX idx_txn_histories_invoice_id ON transaction_histories (invoice_id);
CREATE INDEX idx_txn_histories_institusi_id ON transaction_histories (institusi_id);
CREATE INDEX idx_txn_histories_created_at ON transaction_histories (created_at);

COMMENT ON TABLE transaction_histories IS 'Append-only audit trail. Records every invoice status change.';
COMMENT ON COLUMN transaction_histories.source IS 'Origin of the status change: webhook, manual, or system_expiry.';
COMMENT ON COLUMN transaction_histories.source_ref IS 'Reference ID: webhook_log.id for webhook events, user.id for manual changes.';

-- ============================================================================
-- TABLE: webhook_logs
-- Full audit log of all incoming webhook notifications
-- ============================================================================

CREATE TABLE webhook_logs (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    institusi_id    UUID REFERENCES institutions(id) ON DELETE SET NULL,
    event_type      webhook_event_type,
    invoice_number  VARCHAR(50),
    institution_code VARCHAR(20) NOT NULL,
    raw_payload     JSONB NOT NULL,
    raw_headers     JSONB NOT NULL,
    signature       VARCHAR(512),
    processing_status webhook_log_status NOT NULL,
    error_message   TEXT,
    ip_address      INET,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_webhook_signature_present
        CHECK (signature IS NOT NULL OR processing_status = 'invalid_signature')
);

CREATE INDEX idx_webhook_logs_institusi_id ON webhook_logs (institusi_id);
CREATE INDEX idx_webhook_logs_invoice_number ON webhook_logs (invoice_number);
CREATE INDEX idx_webhook_logs_institution_code ON webhook_logs (institution_code);
CREATE INDEX idx_webhook_logs_processing_status ON webhook_logs (processing_status);
CREATE INDEX idx_webhook_logs_created_at ON webhook_logs (created_at);
CREATE INDEX idx_webhook_logs_event_type ON webhook_logs (event_type);

COMMENT ON TABLE webhook_logs IS 'Complete audit log of all webhook notifications received, regardless of processing outcome.';
COMMENT ON COLUMN webhook_logs.raw_payload IS 'Original JSON payload from payment gateway, stored verbatim.';
COMMENT ON COLUMN webhook_logs.raw_headers IS 'Original HTTP headers from the webhook request.';
COMMENT ON COLUMN webhook_logs.processing_status IS 'Outcome of webhook processing: success, failed, invalid_signature, invoice_not_found, invalid_transition.';

-- ============================================================================
-- TABLE: refresh_tokens
-- JWT refresh token storage for token invalidation
-- ============================================================================

CREATE TABLE refresh_tokens (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash      VARCHAR(512) NOT NULL, -- SHA-256 hash of the actual token
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_refresh_tokens_hash UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);
CREATE INDEX idx_refresh_tokens_active ON refresh_tokens (user_id, expires_at) WHERE revoked_at IS NULL;

COMMENT ON TABLE refresh_tokens IS 'Stores hashed refresh tokens for JWT token rotation and invalidation.';

-- ============================================================================
-- ROW-LEVEL SECURITY (RLS) POLICIES
-- Defense-in-depth: even if app logic fails, DB enforces tenant isolation
-- ============================================================================

-- Enable RLS on tenant-scoped tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE students ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE transaction_histories ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_logs ENABLE ROW LEVEL SECURITY;

-- RLS policies use a session variable: SET app.current_institusi_id = '<uuid>'
-- This is set by the application on each database connection from the pool

-- Users: can only see users in their institution (super_admin bypasses)
CREATE POLICY users_tenant_isolation ON users
    USING (
        institusi_id = NULLIF(current_setting('app.current_institusi_id', true), '')::UUID
        OR current_setting('app.current_role', true) = 'super_admin'
        OR institusi_id IS NULL -- super_admin users visible to themselves
    );

-- Students: scoped by institution
CREATE POLICY students_tenant_isolation ON students
    USING (
        institusi_id = NULLIF(current_setting('app.current_institusi_id', true), '')::UUID
        OR current_setting('app.current_role', true) = 'super_admin'
    );

-- Invoices: scoped by institution
CREATE POLICY invoices_tenant_isolation ON invoices
    USING (
        institusi_id = NULLIF(current_setting('app.current_institusi_id', true), '')::UUID
        OR current_setting('app.current_role', true) = 'super_admin'
    );

-- Transaction histories: scoped by institution
CREATE POLICY txn_histories_tenant_isolation ON transaction_histories
    USING (
        institusi_id = NULLIF(current_setting('app.current_institusi_id', true), '')::UUID
        OR current_setting('app.current_role', true) = 'super_admin'
    );

-- Webhook logs: scoped by institution (null institusi_id visible to super_admin only)
CREATE POLICY webhook_logs_tenant_isolation ON webhook_logs
    USING (
        institusi_id = NULLIF(current_setting('app.current_institusi_id', true), '')::UUID
        OR current_setting('app.current_role', true) = 'super_admin'
    );

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_updated_at_institutions
    BEFORE UPDATE ON institutions
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_users
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_students
    BEFORE UPDATE ON students
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_updated_at_invoices
    BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- Prevent UPDATE/DELETE on transaction_histories (append-only)
CREATE OR REPLACE FUNCTION prevent_txn_history_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'transaction_histories is append-only. UPDATE and DELETE are not allowed.';
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER prevent_txn_history_update
    BEFORE UPDATE ON transaction_histories
    FOR EACH ROW EXECUTE FUNCTION prevent_txn_history_modification();

CREATE TRIGGER prevent_txn_history_delete
    BEFORE DELETE ON transaction_histories
    FOR EACH ROW EXECUTE FUNCTION prevent_txn_history_modification();

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Function to expire overdue unpaid invoices (called by cron/scheduled job)
CREATE OR REPLACE FUNCTION expire_overdue_invoices()
RETURNS INTEGER AS $$
DECLARE
    expired_count INTEGER;
BEGIN
    WITH expired AS (
        UPDATE invoices
        SET status = 'expired', updated_at = NOW()
        WHERE status = 'unpaid'
          AND due_date < CURRENT_DATE
        RETURNING id, institusi_id
    ),
    logged AS (
        INSERT INTO transaction_histories (invoice_id, institusi_id, previous_status, new_status, source, notes)
        SELECT id, institusi_id, 'unpaid', 'expired', 'system_expiry', 'Auto-expired: due date passed'
        FROM expired
    )
    SELECT COUNT(*) INTO expired_count FROM expired;

    RETURN expired_count;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION expire_overdue_invoices IS 'Expires all unpaid invoices past their due date and logs transitions. Call from scheduled job.';
