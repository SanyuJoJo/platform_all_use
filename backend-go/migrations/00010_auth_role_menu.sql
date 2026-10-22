-- +goose Up
-- +goose StatementBegin

-- 角色-菜单关联表（方案 A）
--
-- 复合主键去重；两端外键 ON DELETE CASCADE：
--   - 角色删除时自动清关联
--   - 菜单删除时自动清关联
CREATE TABLE IF NOT EXISTS auth_role_menu (
    role_id    BIGINT    NOT NULL,
    menu_id    BIGINT    NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, menu_id),
    CONSTRAINT fk_rm_role FOREIGN KEY (role_id) REFERENCES auth_role(id)     ON DELETE CASCADE,
    CONSTRAINT fk_rm_menu FOREIGN KEY (menu_id) REFERENCES platform_menu(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_auth_role_menu_menu_id ON auth_role_menu(menu_id);

-- 旧数据迁移：auth_role_permission → auth_role_menu
--
-- 语义：旧模型里"角色拥有权限码 X" → "角色勾选拥有权限码 X 的 F 菜单"。
-- 只迁移 F（按钮）类型菜单；M/C 类型由管理员在新 UI 按层级勾选。
-- 幂等：NOT EXISTS 保证重复执行不插入重复行。
INSERT INTO auth_role_menu (role_id, menu_id)
SELECT DISTINCT rp.role_id, m.id
FROM auth_role_permission rp
JOIN auth_permission p ON p.id = rp.permission_id
JOIN platform_menu   m ON m.perms = p.code
WHERE m.menu_type = 'F'
  AND NOT EXISTS (
      SELECT 1 FROM auth_role_menu rm
      WHERE rm.role_id = rp.role_id AND rm.menu_id = m.id
  );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS auth_role_menu;
-- +goose StatementEnd
