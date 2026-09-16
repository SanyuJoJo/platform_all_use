-- +goose Up
-- P1 阶段认证模块表结构（SQLite 方言）
-- 与 Python SQLAlchemy 模型完全一致
CREATE TABLE IF NOT EXISTS auth_user (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(50) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    email VARCHAR(100),
    avatar VARCHAR(255),
    status SMALLINT NOT NULL DEFAULT 1,
    last_login_at DATETIME,
    last_login_ip VARCHAR(45),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_auth_user_username ON auth_user (username);
CREATE INDEX IF NOT EXISTS ix_auth_user_email ON auth_user (email);
CREATE TABLE IF NOT EXISTS auth_role (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(50) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    is_system SMALLINT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_auth_role_code ON auth_role (code);
CREATE TABLE IF NOT EXISTS auth_permission (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code VARCHAR(100) NOT NULL,
    name VARCHAR(50) NOT NULL,
    module_id VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_auth_permission_code ON auth_permission (code);
CREATE INDEX IF NOT EXISTS ix_auth_permission_module_id ON auth_permission (module_id);
CREATE TABLE IF NOT EXISTS auth_user_role (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_auth_user_role_user_role UNIQUE (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES auth_user (id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES auth_role (id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS auth_role_permission (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_auth_role_permission_role_perm UNIQUE (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES auth_role (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES auth_permission (id) ON DELETE CASCADE
);
CREATE TABLE IF NOT EXISTS auth_refresh_token (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    token_hash VARCHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    revoked SMALLINT NOT NULL DEFAULT 0,
    revoked_at DATETIME,
    ip VARCHAR(45),
    user_agent VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES auth_user (id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS ix_auth_refresh_token_token_hash ON auth_refresh_token (token_hash);
CREATE INDEX IF NOT EXISTS ix_auth_refresh_token_user_id ON auth_refresh_token (user_id);
CREATE INDEX IF NOT EXISTS ix_auth_refresh_token_expires_at ON auth_refresh_token (expires_at);
-- +goose Down
DROP TABLE IF EXISTS auth_refresh_token;
DROP TABLE IF EXISTS auth_role_permission;
DROP TABLE IF EXISTS auth_user_role;
DROP TABLE IF EXISTS auth_permission;
DROP TABLE IF EXISTS auth_role;
DROP TABLE IF EXISTS auth_user;
