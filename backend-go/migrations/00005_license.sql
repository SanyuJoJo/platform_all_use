-- +goose Up
-- P5 阶段 License 管理表结构（SQLite 方言）
-- 与 Python SQLAlchemy 模型完全一致
CREATE TABLE IF NOT EXISTS license_license (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    license_key VARCHAR(255) NOT NULL,
    license_type VARCHAR(50) NOT NULL,
    max_users INTEGER,
    authorized_modules JSON,
    machine_code VARCHAR(255),
    issued_at DATETIME NOT NULL,
    expires_at DATETIME NOT NULL,
    is_active SMALLINT NOT NULL DEFAULT 1,
    activated_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_license_license_license_key
    ON license_license (license_key);
CREATE INDEX IF NOT EXISTS ix_license_license_expires_at
    ON license_license (expires_at);
-- ★ 部分唯一索引：保证同一时刻最多一条 is_active=1 记录（R-3）
-- SQLite（≥3.8.0）与 PostgreSQL（≥9.5）均支持部分索引。
CREATE UNIQUE INDEX IF NOT EXISTS uq_license_license_active
    ON license_license (is_active) WHERE is_active = 1;
-- +goose Down
DROP TABLE IF EXISTS license_license;
