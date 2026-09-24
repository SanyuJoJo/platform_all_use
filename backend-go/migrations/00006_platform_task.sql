-- +goose Up
-- 平台任务表（密码操作异步任务）
CREATE TABLE IF NOT EXISTS platform_task (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    task_id VARCHAR(36) NOT NULL,
    operation_id VARCHAR(50) NOT NULL,
    request_id VARCHAR(36) NOT NULL,
    actor_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    progress INTEGER,
    result_ref VARCHAR(255),
    error_code VARCHAR(50),
    error_message TEXT,
    timeout_ms INTEGER,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at DATETIME,
    finished_at DATETIME,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_platform_task_task_id ON platform_task (task_id);
CREATE INDEX IF NOT EXISTS ix_platform_task_operation_id ON platform_task (operation_id);
CREATE INDEX IF NOT EXISTS ix_platform_task_request_id ON platform_task (request_id);
CREATE INDEX IF NOT EXISTS ix_platform_task_actor_id ON platform_task (actor_id);
CREATE INDEX IF NOT EXISTS ix_platform_task_status ON platform_task (status);
-- +goose Down
DROP TABLE IF EXISTS platform_task;
