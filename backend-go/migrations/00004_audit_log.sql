-- +goose Up
-- P4 阶段日志审计表结构（SQLite 方言）
-- 与 Python SQLAlchemy 模型完全一致
CREATE TABLE IF NOT EXISTS audit_log_operation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    username VARCHAR(50),
    module_id VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50),
    resource_id VARCHAR(50),
    detail TEXT,
    ip VARCHAR(45),
    user_agent VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error_code INTEGER,
    request_id VARCHAR(36),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES auth_user (id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS ix_audit_log_operation_user_id ON audit_log_operation (user_id);
CREATE INDEX IF NOT EXISTS ix_audit_log_operation_module_id ON audit_log_operation (module_id);
CREATE INDEX IF NOT EXISTS ix_audit_log_operation_action ON audit_log_operation (action);
CREATE INDEX IF NOT EXISTS ix_audit_log_operation_created_at ON audit_log_operation (created_at);
-- +goose Down
DROP TABLE IF EXISTS audit_log_operation;
