"""License 管理模块 - 依赖注入。
提供：
- CurrentUser / require_permission：复用认证模块
- check_license(module_id)：模块授权校验依赖工厂
"""
from typing import Callable
from fastapi import Depends
from sqlalchemy.ext.asyncio import AsyncSession
from src.core.database import get_db
from src.modules.auth.dependencies import (  # noqa: F401
    CurrentUser,
    require_permission,
)
from src.modules.license.service import check_module_authorized
__all__ = [
    "CurrentUser",
    "require_permission",
    "check_license",
]
def check_license(module_id: str) -> Callable:
    """
    模块授权校验依赖工厂。
    使用方式（在业务模块路由注册时挂载）：
        from src.modules.license.dependencies import check_license
        app.include_router(
            customer_relation_router,
            prefix="/api/v1/customer_relation",
            dependencies=[Depends(check_license("customer_relation"))],
        )
    核心模块（auth / platform / module_manager / audit_log / license）
    始终放行，不受 License 限制。
    """
    async def _check(db: AsyncSession = Depends(get_db)) -> None:
        await check_module_authorized(db, module_id)
    return _check
