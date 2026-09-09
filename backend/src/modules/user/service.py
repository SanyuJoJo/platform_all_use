import uuid
from datetime import datetime, timezone
from typing import List, Optional, Dict, Any
from src.core.exceptions import PlatformException
# ---- 模拟数据存储 ----
# 预设用户（与 auth 模块一致，但为了统一管理，此处单独维护）
MOCK_USERS: Dict[int, dict] = {
    1: {
        "id": 1,
        "username": "admin",
        "password": "123456",  # 明文仅用于演示
        "nickname": "管理员",
        "email": "admin@example.com",
        "avatar": "/uploads/avatar/admin.jpg",
        "status": 1,
        "role_ids": [1],  # admin 角色
        "created_at": datetime(2026, 1, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 28, tzinfo=timezone.utc),
    },
    2: {
        "id": 2,
        "username": "guest",
        "password": "123456",
        "nickname": "访客",
        "email": "guest@example.com",
        "avatar": None,
        "status": 1,
        "role_ids": [2],  # guest 角色
        "created_at": datetime(2026, 1, 1, tzinfo=timezone.utc),
        "updated_at": datetime(2026, 8, 28, tzinfo=timezone.utc),
    },
}
_next_user_id = 3
# 角色信息（用于关联查询）
ROLES = {
    1: {"id": 1, "name": "管理员", "code": "admin"},
    2: {"id": 2, "name": "访客", "code": "guest"},
}
def get_user_by_id(user_id: int) -> Optional[dict]:
    return MOCK_USERS.get(user_id)
def list_users(page: int, page_size: int, keyword: str = None, status: int = None, role_id: int = None) -> dict:
    users = list(MOCK_USERS.values())
    if keyword:
        keyword = keyword.lower()
        users = [u for u in users if keyword in u["username"].lower() or keyword in u["nickname"].lower() or (u.get("email") and keyword in u["email"].lower())]
    if status is not None:
        users = [u for u in users if u["status"] == status]
    if role_id is not None:
        users = [u for u in users if role_id in u["role_ids"]]
    total = len(users)
    start = (page - 1) * page_size
    end = start + page_size
    items = users[start:end]
    # 组装角色信息
    for u in items:
        u["roles"] = [ROLES[rid] for rid in u["role_ids"] if rid in ROLES]
    return {
        "items": items,
        "total": total,
        "page": page,
        "page_size": page_size,
        "pages": (total + page_size - 1) // page_size if page_size > 0 else 0,
    }
def create_user(data: dict) -> dict:
    global _next_user_id
    # 检查用户名唯一
    if any(u["username"] == data["username"] for u in MOCK_USERS.values()):
        raise PlatformException(code=10000, message="用户名已存在", status_code=400)
    # 检查邮箱唯一
    if data.get("email") and any(u.get("email") == data["email"] for u in MOCK_USERS.values()):
        raise PlatformException(code=10009, message="邮箱已被使用", status_code=409)
    new_id = _next_user_id
    _next_user_id += 1
    now = datetime.now(timezone.utc)
    user = {
        "id": new_id,
        "username": data["username"],
        "password": data["password"],  # 实际应加密
        "nickname": data["nickname"],
        "email": data.get("email"),
        "avatar": None,
        "status": data.get("status", 1),
        "role_ids": data.get("role_ids", []),
        "created_at": now,
        "updated_at": now,
    }
    MOCK_USERS[new_id] = user
    user["roles"] = [ROLES[rid] for rid in user["role_ids"] if rid in ROLES]
    return user
def update_user(user_id: int, data: dict) -> dict:
    user = MOCK_USERS.get(user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    # 如果更新邮箱，检查唯一
    if "email" in data and data["email"]:
        if any(u.get("email") == data["email"] and u["id"] != user_id for u in MOCK_USERS.values()):
            raise PlatformException(code=10009, message="邮箱已被使用", status_code=409)
    # 更新字段
    for field in ["nickname", "email", "status"]:
        if field in data and data[field] is not None:
            user[field] = data[field]
    if "role_ids" in data:
        user["role_ids"] = data["role_ids"]
    user["updated_at"] = datetime.now(timezone.utc)
    user["roles"] = [ROLES[rid] for rid in user["role_ids"] if rid in ROLES]
    return user
def delete_user(user_id: int):
    if user_id not in MOCK_USERS:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    if user_id == 1:  # 保护 admin
        raise PlatformException(code=10010, message="不能删除自己", status_code=403)
    del MOCK_USERS[user_id]
def patch_status(user_id: int, status: int) -> dict:
    user = MOCK_USERS.get(user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    if user_id == 1 and status == 0:
        raise PlatformException(code=10011, message="不能禁用自己", status_code=403)
    user["status"] = status
    user["updated_at"] = datetime.now(timezone.utc)
    user["roles"] = [ROLES[rid] for rid in user["role_ids"] if rid in ROLES]
    return user
def reset_password(user_id: int, new_password: str):
    user = MOCK_USERS.get(user_id)
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    user["password"] = new_password  # 实际应加密
    user["updated_at"] = datetime.now(timezone.utc)
