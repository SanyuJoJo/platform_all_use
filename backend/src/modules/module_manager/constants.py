"""模块管理模块常量与核心模块种子数据。"""
from datetime import datetime, timezone
from typing import Any, Dict, List

# 核心模块 ID，这些模块由主应用显式注册，不参与动态加载，也不允许停用/卸载
CORE_MODULE_IDS = {"auth", "platform", "module_manager", "audit_log", "license"}

# 允许变更核心模块的操作（用于白名单），未列入的操作将被拒绝
CORE_MODULE_PROTECTED_ACTIONS = {"enable", "disable", "uninstall", "upgrade"}

# ZIP 安全常量（默认值，可被 settings 覆盖）
DEFAULT_ZIP_MAX_SIZE = 50 * 1024 * 1024
DEFAULT_ZIP_MAX_TOTAL = 200 * 1024 * 1024
DEFAULT_ZIP_MAX_FILES = 2000

# 当前环境的前端 Host（部署到哪就改成哪）
_FRONTEND_HOST = "http://192.168.56.14"

# 复制源码时忽略的模式
COPY_IGNORE_PATTERNS = (
    ".git",
    ".gitignore",
    "__pycache__",
    "*.pyc",
    "*.pyo",
    ".mypy_cache",
    ".ruff_cache",
    ".pytest_cache",
    "node_modules",
    ".venv",
    "venv",
    "*.log",
    "*.db",
)


def _utcnow() -> datetime:
    return datetime.now(timezone.utc).replace(tzinfo=None)


CORE_MODULES: List[Dict[str, Any]] = [
    {
        "id": "platform",
        "name": "平台基础",
        "version": "1.0.0",
        "description": "平台核心功能",
        "author": "平台团队",
        "homepage": None,
        "status": "active",
        "entry_backend": "router:router",
        "entry_frontend": None,
        "manifest": {
            "id": "platform",
            "name": "平台基础",
            "version": "1.0.0",
            "description": "平台核心功能",
            "dependencies": [],
            "permissions": [],
            "menus": [
                {
                    "id": "platform:dashboard",
                    "parent_id": None,
                    "title": "仪表盘",
                    "icon": "Dashboard",
                    "path": "/dashboard",
                    "component": "views/dashboard/index.vue",
                    "permission": "platform:dashboard:view",
                    "order": 0,
                }
            ],
        },
        "config": {},
    },
    {
        "id": "auth",
        "name": "认证授权",
        "version": "1.0.0",
        "description": "用户管理、角色管理、权限管理",
        "author": "平台团队",
        "homepage": None,
        "status": "active",
        "entry_backend": "router:router",
        "entry_frontend": f"{_FRONTEND_HOST}:3001/",
        "manifest": {
            "id": "auth",
            "name": "认证授权",
            "version": "1.0.0",
            "description": "用户管理、角色管理、权限管理",
            "dependencies": [],
            "permissions": [],
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
        "config": {},
    },
    {
        "id": "module_manager",
        "name": "模块管理",
        "version": "1.0.0",
        "description": "管理已安装模块的生命周期",
        "author": "平台团队",
        "homepage": None,
        "status": "active",
        "entry_backend": "router:router",
        "entry_frontend": f"{_FRONTEND_HOST}:3002/",
        "manifest": {
            "id": "module_manager",
            "name": "模块管理",
            "version": "1.0.0",
            "description": "管理已安装模块的生命周期",
            "dependencies": [],
            "permissions": [],
            "menus": [
                {
                    "id": "module_manager:dashboard",
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
        "config": {},
    },
    {
        "id": "audit_log",
        "name": "日志审计",
        "version": "1.0.0",
        "description": "记录和查看系统操作日志",
        "author": "平台团队",
        "homepage": None,
        "status": "active",
        "entry_backend": "router:router",
        "entry_frontend": f"{_FRONTEND_HOST}:3003/",
        "manifest": {
            "id": "audit_log",
            "name": "日志审计",
            "version": "1.0.0",
            "description": "记录和查看系统操作日志",
            "dependencies": ["auth"],
            "permissions": [],
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
        "config": {},
    },
    {
        "id": "license",
        "name": "License 管理",
        "version": "1.0.0",
        "description": "License 导入、激活、状态查询",
        "author": "平台团队",
        "homepage": None,
        "status": "active",
        "entry_backend": "router:router",
        "entry_frontend": None,
        "manifest": {
            "id": "license",
            "name": "License 管理",
            "version": "1.0.0",
            "description": "License 导入、激活、状态查询",
            "dependencies": ["auth"],
            "permissions": [],
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
        "config": {},
    },
]
