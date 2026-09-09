import uuid
from datetime import datetime, timedelta, timezone
from src.core.security import create_access_token, verify_password
from src.core.exceptions import PlatformException

# 模拟用户数据（硬编码）
MOCK_USERS = {
    "admin": {
        "id": 1,
        "username": "admin",
        "password_hash": "$2b$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",  # 密码 "123456"
        "nickname": "管理员",
        "email": "admin@example.com",
        "avatar": "/uploads/avatar/admin.jpg",
        "status": 1,
        "roles": ["admin"],
        "permissions": [
            "dashboard:view",
            # 平台核心
            "auth:user:view", "auth:user:create", "auth:user:edit", "auth:user:delete",
            "auth:role:view", "auth:role:create", "auth:role:edit", "auth:role:delete",
            "module_manager:module:view", "module_manager:module:create",
            "module_manager:module:edit", "module_manager:module:delete",
            "audit_log:log:view", "audit_log:log:export",
            "license:license:view", "license:license:create",
            # 业务模块权限（新增）
            "customer_relation:customer:view",
            "customer_relation:customer:create",
            "project_management:project:view"
        ]
    },
    "guest": {
        "id": 2,
        "username": "guest",
        "password_hash": "$2b$12$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
        "nickname": "访客",
        "email": "guest@example.com",
        "avatar": None,
        "status": 1,
        "roles": ["guest"],
        "permissions": [
            "dashboard:view",
            "auth:user:view",
            "module_manager:module:view"
        ]
    }
}

# 模拟 refresh token 存储（内存）
REFRESH_TOKENS = {}

def authenticate_user(username: str, password: str):
    user = MOCK_USERS.get(username)
    if not user:
        raise PlatformException(code=10001, message="用户名或密码错误", status_code=401)
    if password != "123456":
        raise PlatformException(code=10001, message="用户名或密码错误", status_code=401)
    if user["status"] != 1:
        raise PlatformException(code=10002, message="用户已被禁用", status_code=403)
    return user

def create_tokens_for_user(user_data: dict):
    access_token = create_access_token(
        data={
            "sub": str(user_data["id"]),
            "username": user_data["username"],
            "permissions": user_data["permissions"]
        },
        expires_delta=timedelta(minutes=1440)
    )
    refresh_token = str(uuid.uuid4())
    REFRESH_TOKENS[refresh_token] = user_data["id"]
    return access_token, refresh_token

def refresh_access_token(refresh_token: str):
    if refresh_token not in REFRESH_TOKENS:
        raise PlatformException(code=10001, message="无效的refresh token", status_code=401)
    user_id = REFRESH_TOKENS[refresh_token]
    user_data = None
    for u in MOCK_USERS.values():
        if u["id"] == user_id:
            user_data = u
            break
    if not user_data:
        raise PlatformException(code=10001, message="用户不存在", status_code=404)
    new_token = create_access_token(
        data={
            "sub": str(user_data["id"]),
            "username": user_data["username"],
            "permissions": user_data["permissions"]
        },
        expires_delta=timedelta(minutes=1440)
    )
    return new_token

def get_user_info(user_id: int):
    for u in MOCK_USERS.values():
        if u["id"] == user_id:
            return u
    return None
