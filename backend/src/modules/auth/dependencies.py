"""
认证模块 - 依赖注入
提供 get_current_user 与 require_permission。
"""
from typing import List, Optional, TypedDict
from fastapi import Depends
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy.orm import selectinload
from src.core.database import get_db
from src.core.exceptions import PlatformException
from src.core.security import decode_token
from src.modules.auth.models import Role, User
class CurrentUser(TypedDict):
    """当前用户上下文（依赖注入返回值）"""
    id: int
    username: str
    nickname: str
    email: Optional[str]
    avatar: Optional[str]
    status: int
    roles: List[str]
    permissions: List[str]
# 使用 HTTPBearer：Swagger UI Authorize 可直接填入 Bearer Token
# 通过 src.core.dependencies 暴露兼容别名 oauth2_scheme
bearer_scheme = HTTPBearer(auto_error=False)
async def get_current_user(
    credentials: Optional[HTTPAuthorizationCredentials] = Depends(bearer_scheme),
    db: AsyncSession = Depends(get_db),
) -> CurrentUser:
    """
    从 JWT 解析当前用户，并查询数据库获取角色与权限。
    抛出：
        PlatformException(10001)：未提供令牌 / Token 无效 / Token 类型错误
        PlatformException(10005)：用户不存在
        PlatformException(10002)：用户已被禁用
    """
    if not credentials or not credentials.credentials:
        raise PlatformException(code=10001, message="未提供认证令牌", status_code=401)
    token = credentials.credentials
    try:
        payload = decode_token(token)
    except Exception:
        raise PlatformException(code=10001, message="Token无效或已过期", status_code=401)
    token_type = payload.get("type")
    if token_type and token_type != "access":
        raise PlatformException(code=10001, message="无效的访问令牌", status_code=401)
    user_id = payload.get("sub")
    if not user_id:
        raise PlatformException(code=10001, message="无效的Token", status_code=401)
    try:
        uid = int(user_id)
    except (TypeError, ValueError):
        raise PlatformException(code=10001, message="无效的Token", status_code=401)
    result = await db.execute(
        select(User)
        .options(selectinload(User.roles).selectinload(Role.permissions))
        .where(User.id == uid)
    )
    user = result.scalar_one_or_none()
    if not user:
        raise PlatformException(code=10005, message="用户不存在", status_code=404)
    if user.status != 1:
        raise PlatformException(code=10002, message="用户已被禁用", status_code=403)
    roles = [role.code for role in user.roles]
    permissions = sorted(
        {perm.code for role in user.roles for perm in role.permissions}
    )
    return {
        "id": user.id,
        "username": user.username,
        "nickname": user.nickname,
        "email": user.email,
        "avatar": user.avatar,
        "status": user.status,
        "roles": roles,
        "permissions": permissions,
    }
def require_permission(permission_code: str):
    """
    权限校验依赖工厂。
    用法：
        @router.get(
            "/users",
            dependencies=[Depends(require_permission("auth:user:view"))],
        )
    """
    async def _check_permission(
        current_user: CurrentUser = Depends(get_current_user),
    ) -> CurrentUser:
        if permission_code not in current_user.get("permissions", []):
            raise PlatformException(code=20051, message="无权限访问", status_code=403)
        return current_user
    return _check_permission
