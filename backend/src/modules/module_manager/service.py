"""模块管理模块 - 业务逻辑（mock 版）。

重要说明：
- 所有菜单的 permission 字段统一为三段格式 {module}:{resource}:{action}；
- platform 模块菜单权限为 platform:dashboard:view（历史 dashboard:view 已迁移）；
- 本 mock 版仅用于前端联调，正式版由任务 2-1 交付。

注意：
- MOCK_MODULES 中 entry_frontend 使用 //192.168.56.14:3001/ 等地址，
  请根据实际部署环境调整（或改为 None，让前端使用环境变量备选）。
"""

from datetime import datetime, timezone
from typing import Dict, List, Optional


# ============================================================================
# Mock 模块数据（含菜单）
# ============================================================================
MOCK_MODULES: List[dict] = [
    # ----- 认证授权模块 -----
    {
        "id": "auth",
        "name": "认证授权",
        "version": "1.0.0",
        "description": "用户管理、角色管理、权限管理",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": "//192.168.56.14:3001/",
        "dependencies": [],
        "installed_at": datetime(2026, 8, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 31, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "auth:dashboard",
                "parent_id": None,
                "title": "认证授权",
                "icon": "Lock",
                "path": "/dashboard",
                "component": "views/dashboard/index.vue",
                "permission": "auth:user:view",
                "order": 10,
            },
            {
                "id": "auth:users",
                "parent_id": "auth:dashboard",
                "title": "用户管理",
                "icon": "Person",
                "path": "/users",
                "component": "views/users/index.vue",
                "permission": "auth:user:view",
                "order": 10,
            },
            {
                "id": "auth:roles",
                "parent_id": "auth:dashboard",
                "title": "角色管理",
                "icon": "Shield",
                "path": "/roles",
                "component": "views/roles/index.vue",
                "permission": "auth:role:view",
                "order": 20,
            },
            {
                "id": "auth:permissions",
                "parent_id": "auth:dashboard",
                "title": "权限管理",
                "icon": "List",
                "path": "/permissions",
                "component": "views/permissions/index.vue",
                "permission": "auth:permission:view",
                "order": 30,
            },
        ],
    },
    # ----- 平台虚拟模块（提供仪表盘） -----
    {
        "id": "platform",
        "name": "平台基础",
        "version": "1.0.0",
        "description": "平台核心功能",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": None,
        "dependencies": [],
        "installed_at": datetime(2026, 8, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 31, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "platform:dashboard",
                "parent_id": None,
                "title": "仪表盘",
                "icon": "Dashboard",
                "path": "/dashboard",
                "component": "views/dashboard/index.vue",
                # ★ 关键：三段格式，与 admin 用户权限 platform:dashboard:view 匹配
                "permission": "platform:dashboard:view",
                "order": 0,
            }
        ],
    },
    # ----- 模块管理模块 -----
    {
        "id": "module-manager",
        "name": "模块管理",
        "version": "1.0.0",
        "description": "管理已安装模块的生命周期",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": "//192.168.56.14:3002/",
        "dependencies": [],
        "installed_at": datetime(2026, 9, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 9, 6, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "module-manager:dashboard",
                "parent_id": None,
                "title": "模块管理",
                "icon": "Apps",
                "path": "/modules",
                "component": "views/modules/index.vue",
                "permission": "module_manager:module:view",
                "order": 30,
            }
        ],
    },
    # ----- 日志审计模块 -----
    {
        "id": "audit_log",
        "name": "日志审计",
        "version": "1.0.0",
        "description": "记录和查看系统操作日志",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": "//192.168.56.14:3003/",
        "dependencies": ["auth"],
        "installed_at": datetime(2026, 9, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 9, 8, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "audit_log:dashboard",
                "parent_id": None,
                "title": "日志审计",
                "icon": "Document",
                "path": "/logs",
                "component": "views/logs/index.vue",
                "permission": "audit_log:log:view",
                "order": 40,
            }
        ],
    },
    # ----- License 管理模块 -----
    {
        "id": "license",
        "name": "License 管理",
        "version": "1.0.0",
        "description": "License 导入、激活、状态查询",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": "//192.168.56.14:3004/",
        "dependencies": ["auth"],
        "installed_at": datetime(2026, 9, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 9, 8, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "license:dashboard",
                "parent_id": None,
                "title": "License 管理",
                "icon": "Key",
                "path": "/license",
                "component": "views/license/index.vue",
                "permission": "license:license:view",
                "order": 50,
            }
        ],
    },
]

MOCK_MODULES_DICT: Dict[str, dict] = {m["id"]: m for m in MOCK_MODULES}


def filter_menus_by_permission(
    menus: List[dict], user_permissions: List[str]
) -> List[dict]:
    """根据用户权限过滤菜单项（递归处理子菜单）。"""
    filtered = []
    for menu in menus:
        if menu.get("permission") and menu["permission"] not in user_permissions:
            continue
        if "children" in menu:
            menu["children"] = filter_menus_by_permission(
                menu["children"], user_permissions
            )
        filtered.append(menu)
    return filtered


def get_modules(
    status: Optional[str] = None,
    keyword: Optional[str] = None,
    user_permissions: Optional[List[str]] = None,
) -> List[dict]:
    """查询模块列表（支持状态和关键词过滤），并附加权限过滤后的菜单。"""
    modules = MOCK_MODULES
    if status:
        modules = [m for m in modules if m["status"] == status]
    if keyword:
        keyword = keyword.lower()
        modules = [
            m
            for m in modules
            if keyword in m["id"].lower() or keyword in m["name"].lower()
        ]

    result = []
    for mod in modules:
        mod_copy = mod.copy()
        raw_menus = mod_copy.pop("menus", [])
        if user_permissions is not None:
            filtered_menus = filter_menus_by_permission(raw_menus, user_permissions)
        else:
            filtered_menus = raw_menus
        mod_copy["menus"] = filtered_menus
        result.append(mod_copy)
    return result
