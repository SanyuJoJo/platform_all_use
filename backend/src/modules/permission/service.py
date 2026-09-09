from typing import List, Optional, Dict
from datetime import datetime, timezone
# 模拟所有权限（从各模块 manifest 汇总）
MOCK_PERMISSIONS = [
    {"id": 1, "code": "dashboard:view", "name": "查看仪表盘", "module_id": "platform", "resource": "dashboard", "action": "view", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 2, "code": "auth:user:view", "name": "查看用户", "module_id": "auth", "resource": "user", "action": "view", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 3, "code": "auth:user:create", "name": "创建用户", "module_id": "auth", "resource": "user", "action": "create", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 4, "code": "auth:user:edit", "name": "编辑用户", "module_id": "auth", "resource": "user", "action": "edit", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 5, "code": "auth:user:delete", "name": "删除用户", "module_id": "auth", "resource": "user", "action": "delete", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 6, "code": "auth:role:view", "name": "查看角色", "module_id": "auth", "resource": "role", "action": "view", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 7, "code": "auth:role:create", "name": "创建角色", "module_id": "auth", "resource": "role", "action": "create", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 8, "code": "auth:role:edit", "name": "编辑角色", "module_id": "auth", "resource": "role", "action": "edit", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 9, "code": "auth:role:delete", "name": "删除角色", "module_id": "auth", "resource": "role", "action": "delete", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 10, "code": "module_manager:module:view", "name": "查看模块", "module_id": "module_manager", "resource": "module", "action": "view", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 11, "code": "module_manager:module:create", "name": "安装模块", "module_id": "module_manager", "resource": "module", "action": "create", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 12, "code": "module_manager:module:edit", "name": "编辑模块配置", "module_id": "module_manager", "resource": "module", "action": "edit", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 13, "code": "module_manager:module:delete", "name": "卸载模块", "module_id": "module_manager", "resource": "module", "action": "delete", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 14, "code": "audit_log:log:view", "name": "查看日志", "module_id": "audit_log", "resource": "log", "action": "view", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 15, "code": "audit_log:log:export", "name": "导出日志", "module_id": "audit_log", "resource": "log", "action": "export", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 16, "code": "license:license:view", "name": "查看License", "module_id": "license", "resource": "license", "action": "view", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
    {"id": 17, "code": "license:license:create", "name": "管理License", "module_id": "license", "resource": "license", "action": "create", "created_at": datetime(2026,1,1,tzinfo=timezone.utc)},
]
def list_permissions(module_id: Optional[str] = None, resource: Optional[str] = None) -> List[dict]:
    perms = MOCK_PERMISSIONS
    if module_id:
        perms = [p for p in perms if p["module_id"] == module_id]
    if resource:
        perms = [p for p in perms if p["resource"] == resource]
    return perms
def get_permissions_by_module(module_id: str) -> List[dict]:
    return [p for p in MOCK_PERMISSIONS if p["module_id"] == module_id]
