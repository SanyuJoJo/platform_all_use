-- +goose Up
-- P3 阶段模块管理表结构（SQLite 方言）
-- 与 Python SQLAlchemy 模型完全一致
CREATE TABLE IF NOT EXISTS module_manager_module (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    version VARCHAR(20) NOT NULL,
    description TEXT,
    author VARCHAR(100),
    homepage VARCHAR(255),
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    entry_backend VARCHAR(100) NOT NULL,
    entry_frontend VARCHAR(255),
    config JSON,
    manifest JSON,
    installed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS ix_module_manager_module_status ON module_manager_module (status);
CREATE TABLE IF NOT EXISTS module_manager_dependency (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    module_id VARCHAR(50) NOT NULL,
    dependency_id VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_module_manager_dependency_module_dep UNIQUE (module_id, dependency_id),
    FOREIGN KEY (module_id) REFERENCES module_manager_module (id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS ix_module_manager_dependency_module_id ON module_manager_dependency (module_id);
CREATE INDEX IF NOT EXISTS ix_module_manager_dependency_dependency_id ON module_manager_dependency (dependency_id);
-- +goose Down
DROP TABLE IF EXISTS module_manager_dependency;
DROP TABLE IF EXISTS module_manager_module;
