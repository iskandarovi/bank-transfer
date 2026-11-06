-- 1. Database yaratish (psql ichida bir marta bajariladi):
-- CREATE DATABASE bank_transfer;
-- 2. bank_transfer bazasiga ulanib, quyidagi SQL buyruqlarini bajaring.

CREATE TABLE IF NOT EXISTS accounts (
    id BIGSERIAL PRIMARY KEY,
    owner_name VARCHAR(100) NOT NULL CHECK (length(trim(owner_name)) > 0),
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transfers (
    id BIGSERIAL PRIMARY KEY,
    sender_id BIGINT NOT NULL REFERENCES accounts(id),
    receiver_id BIGINT NOT NULL REFERENCES accounts(id),
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transfers_different_accounts CHECK (sender_id <> receiver_id)
);

CREATE INDEX IF NOT EXISTS idx_transfers_sender_id ON transfers(sender_id);
CREATE INDEX IF NOT EXISTS idx_transfers_receiver_id ON transfers(receiver_id);
CREATE INDEX IF NOT EXISTS idx_transfers_created_at ON transfers(created_at DESC);
