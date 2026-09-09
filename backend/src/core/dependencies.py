from typing import TypedDict, List, Optional
from fastapi import Depends
from fastapi.security import OAuth2PasswordBearer
from src.core.security import decode_token
from src.core.database import AsyncSession, get_db
from src.core.exceptions import PlatformException
# 定义当前用户返回结构的类型提示（便于后续迁移至 Pydantic Schema）
class UserPayload(TypedDict, total=False):
    id: int
    username: str
    permissions: List[str]
    # 后续可扩展更多字段（如 nickname, email, status 等）
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/v1/auth/login", auto_error=False)
async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: AsyncSession = Depends(get_db),
) -> UserPayload:
    """
    从 JWT Token 解析当前用户信息。
    
    **TODO（后续替换）**：
    当前实现仅从 Token 中解析用户 ID 和权限列表，返回模拟用户字典。
    后续认证模块完成数据库模型后，应替换为：
    1. 根据 token 中的 user_id 查询 auth_user 表
    2. 从数据库获取用户真实信息（用户名、昵称、状态等）
    3. 从角色-权限关联表获取完整的权限列表（与 Token 中比对或直接使用数据库结果）
    此设计可保持 Token 轻量，同时支持权限动态变更（但不影响已签发 Token）。
    
    当前模拟数据确保框架可立即运行，不影响前后端联调。
    """
    if not token:
        raise PlatformException(code=10001, message="未提供认证令牌", status_code=401)
    try:
        payload = decode_token(token)
        user_id = payload.get("sub")
        if not user_id:
            raise PlatformException(code=10001, message="无效的Token", status_code=401)
        # TODO: 替换为真实数据库查询
        # from src.modules.auth.models import User
        # user = await db.get(User, int(user_id))
        # if not user:
        #     raise PlatformException(code=10002, message="用户不存在", status_code=404)
        # return user
        # 临时返回模拟数据（符合后续替换的接口约定）
        return {
            "id": int(user_id),
            "username": payload.get("username", "unknown"),
            "permissions": payload.get("permissions", [])
        }
    except PlatformException:
        raise
    except Exception:
        raise PlatformException(code=10001, message="Token无效或已过期", status_code=401)
def require_permission(permission_code: str):
    """
    权限校验依赖工厂。
    
    使用方式：在路由中通过 Depends 调用，例如：
        @app.get("/admin", dependencies=[Depends(require_permission("admin:view"))])
    
    该函数会校验当前用户是否拥有指定权限，若无则抛出 PlatformException(code=20051)。
    """
    async def _check_permission(current_user: UserPayload = Depends(get_current_user)):
        # 启用权限检查
        if permission_code not in current_user.get("permissions", []):
            raise PlatformException(
                code=20051,
                message="无权限访问该资源",
                status_code=403
            )
        return current_user
    return _check_permission
