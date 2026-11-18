-- +goose Up
-- +goose StatementBegin
-- TCPS 代理表（4 层 TCP/TLS 代理）
-- 生产环境跳过 AutoMigrate，需 goose 建表；与 GORM model 对齐。
CREATE TABLE IF NOT EXISTS platform_tcps_proxy (
    id                        INTEGER PRIMARY KEY AUTOINCREMENT,
    service_id                VARCHAR(64)  NOT NULL,
    service_name              VARCHAR(255) NOT NULL,
    mode                      VARCHAR(16)  NOT NULL,
    protocol                  VARCHAR(16)  NOT NULL,
    listen_ip                 VARCHAR(64)  NOT NULL,
    listen_port               INTEGER      NOT NULL,
    backend_ip                VARCHAR(64)  NOT NULL,
    backend_port              INTEGER      NOT NULL,
    backend_tls               BOOLEAN      NOT NULL DEFAULT false,
    ssl_server_name           VARCHAR(255) DEFAULT '',
    listen_ssl                BOOLEAN      NOT NULL DEFAULT false,
    cert_id                   VARCHAR(64)  DEFAULT '',
    enc_cert_id               VARCHAR(64)  DEFAULT '',
    intl_cert_id              VARCHAR(64)  DEFAULT '',
    ca_cert_id                VARCHAR(64)  DEFAULT '',
    ca_domain                 VARCHAR(64)  DEFAULT '',
    client_cert_id            VARCHAR(64)  DEFAULT '',
    client_enc_cert_id        VARCHAR(64)  DEFAULT '',
    intl_client_cert_id       VARCHAR(64)  DEFAULT '',
    client_key_ref            VARCHAR(64)  DEFAULT '',
    ssl_protocols             VARCHAR(128) DEFAULT '',
    ssl_ciphers               VARCHAR(1024) DEFAULT '',
    ssl_conf_command          VARCHAR(512) DEFAULT '',
    ssl_ciphers_mode          VARCHAR(16)  NOT NULL DEFAULT 'default',
    ssl_session_cache_enabled BOOLEAN      NOT NULL DEFAULT false,
    ssl_verify_client         BOOLEAN      NOT NULL DEFAULT false,
    ssl_session_tickets       BOOLEAN      NOT NULL DEFAULT false,
    status                    VARCHAR(16)  NOT NULL DEFAULT 'STOPPED',
    pid                       INTEGER      NOT NULL DEFAULT 0,
    config_kv_path            VARCHAR(255) DEFAULT '',
    config_nginx_path         VARCHAR(255) DEFAULT '',
    log_dir                   VARCHAR(255) DEFAULT '',
    created_at                DATETIME,
    updated_at                DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_tcps_proxy_service_id
    ON platform_tcps_proxy(service_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_tcps_proxy_mode_name
    ON platform_tcps_proxy(mode, service_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform_tcps_proxy;
-- +goose StatementEnd
