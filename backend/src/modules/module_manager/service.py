# backend/src/modules/module_manager/service.py

from datetime import datetime, timezone
from typing import List, Optional, Dict
from .schemas import ModuleOut, MenuItem
from src.core.exceptions import PlatformException

# ============================================================================
# 模拟模块数据（含菜单）
# ============================================================================

MOCK_MODULES = [
    # ----- 认证授权模块（核心） -----
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
                "order": 10
            },
            {
                "id": "auth:users",
                "parent_id": "auth:dashboard",
                "title": "用户管理",
                "icon": "Person",
                "path": "/users",
                "component": "views/users/index.vue",
                "permission": "auth:user:view",
                "order": 10
            },
            {
                "id": "auth:roles",
                "parent_id": "auth:dashboard",
                "title": "角色管理",
                "icon": "Shield",
                "path": "/roles",
                "component": "views/roles/index.vue",
                "permission": "auth:role:view",
                "order": 20
            }
        ]
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
                "permission": "dashboard:view",
                "order": 0
            }
        ]
    },
    # ----- 客户关系管理模块 -----
    {
        "id": "customer_relation",
        "name": "客户关系管理",
        "version": "1.0.0",
        "description": "管理客户信息、跟进记录及商机",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": None,
        "dependencies": ["auth", "audit_log"],
        "installed_at": datetime(2026, 8, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 28, 10, 30, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "customer_relation:dashboard",
                "parent_id": None,
                "title": "客户管理",
                "icon": "People",
                "path": "/customer/dashboard",
                "component": "views/dashboard/index.vue",
                "permission": "customer_relation:customer:view",
                "order": 10
            },
            {
                "id": "customer_relation:customer_list",
                "parent_id": "customer_relation:dashboard",
                "title": "客户列表",
                "icon": "List",
                "path": "/customer/list",
                "component": "views/customer/list.vue",
                "permission": "customer_relation:customer:view",
                "order": 10
            },
            {
                "id": "customer_relation:customer_create",
                "parent_id": "customer_relation:dashboard",
                "title": "创建客户",
                "icon": "Add",
                "path": "/customer/create",
                "component": "views/customer/create.vue",
                "permission": "customer_relation:customer:create",
                "order": 20
            }
        ]
    },
    # ----- 项目管理模块 -----
    {
        "id": "project_management",
        "name": "项目管理",
        "version": "1.0.0",
        "description": "管理项目计划、任务和进度",
        "author": "平台团队",
        "status": "active",
        "entry_frontend": None,
        "dependencies": ["auth"],
        "installed_at": datetime(2026, 8, 5, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 28, 11, 0, tzinfo=timezone.utc),
        "menus": [
            {
                "id": "project_management:dashboard",
                "parent_id": None,
                "title": "项目管理",
                "icon": "FolderOpen",
                "path": "/project/dashboard",
                "component": "views/dashboard/index.vue",
                "permission": "project_management:project:view",
                "order": 20
            },
            {
                "id": "project_management:project_list",
                "parent_id": "project_management:dashboard",
                "title": "项目列表",
                "icon": "List",
                "path": "/project/list",
                "component": "views/project/list.vue",
                "permission": "project_management:project:view",
                "order": 10
            }
        ]
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
                "order": 30
            }
        ]
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
                "order": 40
            }
        ]
    },
    # ----- License 管理模块（新增） -----
    {
        "id": "license",
        "name": "License管理",
        "version": "1.0.0",
        "description": "License状态查看、导入与激活",
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
                "title": "License管理",
                "icon": "Document",
                "path": "/status",
                "component": "views/license/index.vue",
                "permission": "license:license:view",
                "order": 50
            },
            {
                "id": "license:modules",
                "parent_id": "license:dashboard",
                "title": "模块授权",
                "icon": "List",
                "path": "/modules",
                "component": "views/modules/index.vue",
                "permission": "license:license:view",
                "order": 10
            }
        ]
    }
]

# 构建字典便于快速查找
MOCK_MODULES_DICT: Dict[str, dict] = {m["id"]: m for m in MOCK_MODULES}

# 模拟模块配置（独立存储）
MODULE_CONFIGS: Dict[str, dict] = {
    "customer_relation": {"enable_follow_up": True, "max_customers": 1000},
    "project_management": {"auto_assign": False},
    "module-manager": {"enable_auto_install": False},
    "audit_log": {"log_retention_days": 90, "enable_audit": True},
    "license": {"license_file": None, "auto_renew": False},  # 示例配置
}

# ============================================================================
# 核心业务函数
# ============================================================================

def filter_menus_by_permission(menus: List[dict], user_permissions: List[str]) -> List[dict]:
    """
    根据用户权限过滤菜单项（递归处理子菜单）。
    """
    filtered = []
    for menu in menus:
        # 如果有权限要求且用户不具备，则跳过
        if menu.get("permission") and menu["permission"] not in user_permissions:
            continue
        # 递归处理子菜单（如有）
        if "children" in menu:
            menu["children"] = filter_menus_by_permission(menu["children"], user_permissions)
        filtered.append(menu)
    return filtered


def get_modules(
    status: Optional[str] = None,
    keyword: Optional[str] = None,
    user_permissions: Optional[List[str]] = None
) -> List[dict]:
    """
    查询模块列表（支持状态和关键词过滤），并附加权限过滤后的菜单。
    """
    modules = MOCK_MODULES
    if status:
        modules = [m for m in modules if m["status"] == status]
    if keyword:
        keyword = keyword.lower()
        modules = [
            m for m in modules
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
        mod_copy["menus"] = [MenuItem(**item).dict() for item in filtered_menus]
        result.append(ModuleOut(**mod_copy).dict())
    return result


def install_module(install_type: str, file_path: str = None, source_path: str = None) -> dict:
    """
    安装模块（模拟）：生成一个新模块并加入内存列表。
    """
    import uuid
    new_id = "new_module_" + str(uuid.uuid4())[:8]
    now = datetime.now(timezone.utc)
    module = {
        "id": new_id,
        "name": "新模块",
        "version": "1.0.0",
        "description": "通过模拟安装添加的模块",
        "author": "未知",
        "status": "inactive",
        "entry_frontend": None,
        "dependencies": [],
        "installed_at": now,
        "updated_at": now,
        "menus": []
    }
    MOCK_MODULES_DICT[new_id] = module
    MOCK_MODULES.append(module)
    return module


def uninstall_module(module_id: str, force: bool = False):
    """
    卸载模块。若 force=False 且被其他模块依赖则报错。
    """
    if module_id not in MOCK_MODULES_DICT:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)

    if not force:
        # 检查是否有其他模块依赖此模块
        for m in MOCK_MODULES:
            if module_id in m.get("dependencies", []):
                raise PlatformException(
                    code=30006,
                    message="模块已被其他模块依赖，请使用强制卸载",
                    status_code=400
                )

    # 从字典和列表中移除
    del MOCK_MODULES_DICT[module_id]
    for i, m in enumerate(MOCK_MODULES):
        if m["id"] == module_id:
            del MOCK_MODULES[i]
            break


def enable_module(module_id: str) -> dict:
    """
    启用模块，会检查依赖模块是否已激活。
    """
    module = MOCK_MODULES_DICT.get(module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)
    if module["status"] == "active":
        return module

    # 检查依赖状态
    for dep in module.get("dependencies", []):
        dep_mod = MOCK_MODULES_DICT.get(dep)
        if not dep_mod or dep_mod["status"] != "active":
            raise PlatformException(
                code=30001,
                message=f"依赖模块 {dep} 未激活",
                status_code=400
            )

    module["status"] = "active"
    module["updated_at"] = datetime.now(timezone.utc)
    return module


def disable_module(module_id: str) -> dict:
    """
    停用模块。
    """
    module = MOCK_MODULES_DICT.get(module_id)
    if not module:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)
    if module["status"] == "inactive":
        return module

    module["status"] = "inactive"
    module["updated_at"] = datetime.now(timezone.utc)
    return module


def get_module_config(module_id: str) -> dict:
    """
    获取模块配置。
    """
    if module_id not in MOCK_MODULES_DICT:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)
    return MODULE_CONFIGS.get(module_id, {})


def update_module_config(module_id: str, config: dict) -> dict:
    """
    更新模块配置（全量覆盖）。
    """
    if module_id not in MOCK_MODULES_DICT:
        raise PlatformException(code=30003, message="模块不存在", status_code=404)
    # 实际应校验 config 是否符合 schema，此处仅简单存储
    MODULE_CONFIGS[module_id] = config
    return MODULE_CONFIGS[module_id]
