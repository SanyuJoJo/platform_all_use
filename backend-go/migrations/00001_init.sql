-- +goose Up
-- P0 阶段只支持 SQLite 迁移。
-- PostgreSQL 方言的迁移在 P1 认证模块启动后补齐。
CREATE TABLE IF NOT EXISTS sys_config (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key VARCHAR(64) NOT NULL UNIQUE,
    value VARCHAR(512) NOT NULL,
    description VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose Down
DROP TABLE IF EXISTS sys_config;
