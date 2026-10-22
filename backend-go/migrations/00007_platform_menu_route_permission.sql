-- +goose Up
-- +goose StatementBegin
-- 菜单/路由元数据表（对标若依 sys_menu）
CREATE TABLE IF NOT EXISTS platform_menu (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    parent_id    INTEGER NOT NULL DEFAULT 0,
    menu_name    VARCHAR(50)  NOT NULL,
    order_num    INTEGER      NOT NULL DEFAULT 0,
    path         VARCHAR(200) NOT NULL DEFAULT '',
    component    VARCHAR(255) NOT NULL DEFAULT '',
    query        VARCHAR(255) NOT NULL DEFAULT '',
    is_frame     SMALLINT     NOT NULL DEFAULT 1,
    is_cache     SMALLINT     NOT NULL DEFAULT 0,
    menu_type    CHAR(1)      NOT NULL,
    visible      SMALLINT     NOT NULL DEFAULT 0,
    status       SMALLINT     NOT NULL DEFAULT 0,
    perms        VARCHAR(100) NOT NULL DEFAULT '',
    icon         VARCHAR(100) NOT NULL DEFAULT '#',
    module_id    VARCHAR(50)  NOT NULL DEFAULT '',
    create_by    VARCHAR(64)  NOT NULL DEFAULT '',
    create_time  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    update_by    VARCHAR(64)  NOT NULL DEFAULT '',
    update_time  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    remark       VARCHAR(500) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_platform_menu_parent ON platform_menu(parent_id);
CREATE INDEX IF NOT EXISTS idx_platform_menu_type   ON platform_menu(menu_type);
CREATE INDEX IF NOT EXISTS idx_platform_menu_status ON platform_menu(status);
CREATE INDEX IF NOT EXISTS idx_platform_menu_perms  ON platform_menu(perms);
CREATE INDEX IF NOT EXISTS idx_platform_menu_module ON platform_menu(module_id);
-- 路由-权限映射表（对标若依 Casbin 策略）
CREATE TABLE IF NOT EXISTS platform_route_permission (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    http_method     VARCHAR(10)  NOT NULL,
    route_pattern   VARCHAR(255) NOT NULL,
    permission_code VARCHAR(100) NOT NULL,
    module_id       VARCHAR(50)  NOT NULL DEFAULT '',
    status          SMALLINT     NOT NULL DEFAULT 0,
    description     VARCHAR(255) NOT NULL DEFAULT '',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_rp_method_pattern ON platform_route_permission(http_method, route_pattern);
CREATE INDEX IF NOT EXISTS idx_rp_perm_code      ON platform_route_permission(permission_code);
CREATE INDEX IF NOT EXISTS idx_rp_module         ON platform_route_permission(module_id);
CREATE INDEX IF NOT EXISTS idx_rp_status         ON platform_route_permission(status);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform_route_permission;
DROP TABLE IF EXISTS platform_menu;
-- +goose StatementEnd
