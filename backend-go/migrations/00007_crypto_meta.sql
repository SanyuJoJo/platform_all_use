-- +goose Up
CREATE TABLE IF NOT EXISTS platform_ca (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ca_id VARCHAR(64) NOT NULL,
    parent_ca_id VARCHAR(64),
    subject_cn VARCHAR(255) NOT NULL,
    subject_o VARCHAR(255),
    algorithm VARCHAR(32) NOT NULL,
    key_params JSON,
    validity_days INTEGER NOT NULL,
    cert_path VARCHAR(255) NOT NULL,
    key_ref VARCHAR(64) NOT NULL,
    chain_path VARCHAR(255),
    serial VARCHAR(128),
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_platform_ca_ca_id ON platform_ca (ca_id);
CREATE INDEX IF NOT EXISTS ix_platform_ca_parent_ca_id ON platform_ca (parent_ca_id);
CREATE INDEX IF NOT EXISTS ix_platform_ca_status ON platform_ca (status);

CREATE TABLE IF NOT EXISTS platform_certificate (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    cert_id VARCHAR(64) NOT NULL,
    cert_type VARCHAR(32) NOT NULL,
    serial VARCHAR(128) NOT NULL,
    subject_cn VARCHAR(255) NOT NULL,
    issuer_cn VARCHAR(255),
    ca_id VARCHAR(64) NOT NULL,
    algorithm VARCHAR(32) NOT NULL,
    not_before DATETIME NOT NULL,
    not_after DATETIME NOT NULL,
    cert_path VARCHAR(255) NOT NULL,
    chain_path VARCHAR(255),
    status VARCHAR(16) NOT NULL DEFAULT 'VALID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_platform_certificate_cert_id ON platform_certificate (cert_id);
CREATE INDEX IF NOT EXISTS ix_platform_certificate_serial ON platform_certificate (serial);
CREATE INDEX IF NOT EXISTS ix_platform_certificate_ca_id ON platform_certificate (ca_id);
CREATE INDEX IF NOT EXISTS ix_platform_certificate_not_after ON platform_certificate (not_after);
CREATE INDEX IF NOT EXISTS ix_platform_certificate_status ON platform_certificate (status);

CREATE TABLE IF NOT EXISTS platform_csr (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    csr_id VARCHAR(64) NOT NULL,
    subject_cn VARCHAR(255) NOT NULL,
    algorithm VARCHAR(32) NOT NULL,
    csr_path VARCHAR(255) NOT NULL,
    key_ref VARCHAR(64),
    status VARCHAR(16) NOT NULL DEFAULT 'NEW',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_platform_csr_csr_id ON platform_csr (csr_id);
CREATE INDEX IF NOT EXISTS ix_platform_csr_status ON platform_csr (status);

CREATE TABLE IF NOT EXISTS platform_crl (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    crl_id VARCHAR(64) NOT NULL,
    ca_id VARCHAR(64) NOT NULL,
    crl_path VARCHAR(255) NOT NULL,
    revoked_count INTEGER NOT NULL DEFAULT 0,
    digest_algorithm VARCHAR(16) NOT NULL,
    next_update DATETIME,
    status VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_platform_crl_crl_id ON platform_crl (crl_id);
CREATE INDEX IF NOT EXISTS ix_platform_crl_ca_id ON platform_crl (ca_id);
CREATE INDEX IF NOT EXISTS ix_platform_crl_status ON platform_crl (status);

CREATE TABLE IF NOT EXISTS platform_key_meta (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_id VARCHAR(64) NOT NULL,
    key_ref VARCHAR(64) NOT NULL,
    algorithm VARCHAR(32) NOT NULL,
    key_params JSON,
    encrypted_path VARCHAR(255) NOT NULL,
    permission VARCHAR(8) NOT NULL DEFAULT '0600',
    usage VARCHAR(64),
    state VARCHAR(16) NOT NULL DEFAULT 'ACTIVE',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_platform_key_meta_key_id ON platform_key_meta (key_id);
CREATE INDEX IF NOT EXISTS ix_platform_key_meta_key_ref ON platform_key_meta (key_ref);
CREATE INDEX IF NOT EXISTS ix_platform_key_meta_state ON platform_key_meta (state);

-- +goose Down
DROP TABLE IF EXISTS platform_key_meta;
DROP TABLE IF EXISTS platform_crl;
DROP TABLE IF EXISTS platform_csr;
DROP TABLE IF EXISTS platform_certificate;
DROP TABLE IF EXISTS platform_ca;
