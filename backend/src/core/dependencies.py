from fastapi import Depends, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from src.core.security import decode_token
from src.core.database import AsyncSession, get_db
from src.core.exceptions import PlatformException
oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/v1/auth/login", auto_error=False)
async def get_current_user(
    token: str = Depends(oauth2_scheme),
    db: AsyncSession = Depends(get_db),
):
    """
    从 JWT Token 解析当前用户信息。
    此函数为依赖注入使用，后续认证模块需完善用户查询逻辑。
    当前实现返回模拟用户数据，以确保框架可运行。
    """
    if not token:
        raise PlatformException(code=10001, message="未提供认证令牌", status_code=401)
    try:
        payload = decode_token(token)
        user_id = payload.get("sub")
        if not user_id:
            raise PlatformException(code=10001, message="无效的Token", status_code=401)
        # TODO: 从数据库查询用户
        # user = await db.get(User, int(user_id))
        # if not user:
        #     raise PlatformException(code=10002, message="用户不存在", status_code=404)
        # return user
        # 临时返回模拟数据
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
    使用时：@app.get("/admin", dependencies=[Depends(require_permission("admin:view"))])
    """
    async def _check_permission(current_user: dict = Depends(get_current_user)):
        # TODO: 从数据库或缓存检查用户是否拥有 permission_code
        # if permission_code not in current_user.get("permissions", []):
        #     raise PlatformException(code=20051, message="无权限访问", status_code=403)
        return current_user
    return _check_permission
