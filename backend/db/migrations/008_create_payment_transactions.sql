CREATE TABLE IF NOT EXISTS payment_transactions (
    id TEXT PRIMARY KEY,
    hold_id TEXT NOT NULL REFERENCES booking_holds(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(10,2) NOT NULL,
    payment_method TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('SUCCESS', 'FAILED', 'PENDING')),
    gateway_txn_id TEXT,
    failure_reason TEXT,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_txn_idempotency
ON payment_transactions(idempotency_key);
