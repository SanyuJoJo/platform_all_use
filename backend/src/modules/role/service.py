from datetime import datetime, timezone
from typing import List, Optional, Dict
from src.core.exceptions import PlatformException
# 模拟数据
MOCK_ROLES: Dict[int, dict] = {
    1: {
        "id": 1,
        "name": "管理员",
        "code": "admin",
        "description": "拥有所有权限",
        "is_system": True,
        "permission_codes": [
            "dashboard:view",
            "auth:user:view", "auth:user:create", "auth:user:edit", "auth:user:delete",
            "auth:role:view", "auth:role:create", "auth:role:edit", "auth:role:delete",
            "module_manager:module:view", "module_manager:module:create",
            "module_manager:module:edit", "module_manager:module:delete",
            "audit_log:log:view", "audit_log:log:export",
            "license:license:view", "license:license:create"
        ],
        "created_at": datetime(2026, 1, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 28, tzinfo=timezone.utc),
    },
    2: {
        "id": 2,
        "name": "访客",
        "code": "guest",
        "description": "只读权限",
        "is_system": True,
        "permission_codes": ["dashboard:view", "auth:user:view", "module_manager:module:view"],
        "created_at": datetime(2026, 1, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 28, tzinfo=timezone.utc),
    },
}
_next_role_id = 3
def list_roles(page: int, page_size: int, keyword: str = None) -> dict:
    roles = list(MOCK_ROLES.values())
    if keyword:
        keyword = keyword.lower()
        roles = [r for r in roles if keyword in r["name"].lower() or keyword in r["code"].lower()]
    total = len(roles)
    start = (page - 1) * page_size
    end = start + page_size
    items = roles[start:end]
    return {
        "items": items,
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }
def create_role(data: dict) -> dict:
    global _next_role_id
    if any(r["code"] == data["code"] for r in MOCK_ROLES.values()):
        raise PlatformException(code=20001, message="角色编码已存在", status_code=400)
    if any(r["name"] == data["name"] for r in MOCK_ROLES.values()):
        raise PlatformException(code=20004, message="角色名称已存在", status_code=400)
    new_id = _next_role_id
    _next_role_id += 1
    now = datetime.now(timezone.utc)
    role = {
        "id": new_id,
        "name": data["name"],
        "code": data["code"],
        "description": data.get("description"),
        "is_system": False,
        "permission_codes": data.get("permission_codes", []),
        "created_at": now,
        "updated_at": now,
    }
    MOCK_ROLES[new_id] = role
    return role
def get_role_by_id(role_id: int) -> Optional[dict]:
    return MOCK_ROLES.get(role_id)
def update_role(role_id: int, data: dict) -> dict:
    role = MOCK_ROLES.get(role_id)
    if not role:
        raise PlatformException(code=20002, message="角色不存在", status_code=404)
    if role["is_system"]:
        raise PlatformException(code=20003, message="不能修改系统内置角色", status_code=403)
    if "name" in data and data["name"]:
        if any(r["name"] == data["name"] and r["id"] != role_id for r in MOCK_ROLES.values()):
            raise PlatformException(code=20004, message="角色名称已存在", status_code=400)
        role["name"] = data["name"]
    if "description" in data:
        role["description"] = data["description"]
    if "permission_codes" in data:
        role["permission_codes"] = data["permission_codes"]
    role["updated_at"] = datetime.now(timezone.utc)
    return role
def delete_role(role_id: int):
    role = MOCK_ROLES.get(role_id)
    if not role:
        raise PlatformException(code=20002, message="角色不存在", status_code=404)
    if role["is_system"]:
        raise PlatformException(code=20003, message="不能删除系统内置角色", status_code=403)
    # 检查是否被用户使用（模拟检查 admin 和 guest 使用了角色 1,2）
    if role_id in [1, 2]:
        # 简单模拟：如果有用户使用该角色则不允许删除
        from src.modules.user.service import MOCK_USERS
        for u in MOCK_USERS.values():
            if role_id in u["role_ids"]:
                raise PlatformException(code=20005, message="角色已被用户使用，请先解除关联", status_code=409)
    del MOCK_ROLES[role_id]
